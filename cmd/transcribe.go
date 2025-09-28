package cmd

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/sebnyberg/flagtags"
	"github.com/sebnyberg/sttrouter/audio"
	"github.com/sebnyberg/sttrouter/clipboard"
	"github.com/sebnyberg/sttrouter/openaix"
	"github.com/urfave/cli/v2"
)

// TranscribeConfig holds transcribe specific configuration flags.
type TranscribeConfig struct {
	// Model specifies the Whisper model to use
	Model string `name:"model" value:"gpt-4o-transcribe" usage:"Model to use for transcription"`
	// Language specifies the language code
	Language string `name:"language" value:"en" usage:"Language code (e.g., 'en', 'es')"`
	// ResponseFormat specifies the response format (json, text, srt, verbose_json, vtt)
	ResponseFormat string `name:"response-format" value:"text" usage:"Response format (json,text,srt,verbose_json,vtt)"`
	// Temperature specifies the sampling temperature (0.0 to 1.0)
	Temperature float64 `name:"temperature" value:"0" usage:"Sampling temperature (0.0 to 1.0)"`
	// OpenAI API Key
	APIKey string `name:"api-key"`
	// API Base URL
	BaseURL string `name:"base-url" value:"https://seblab-ai.openai.azure.com/openai/deployments/gpt-4o-transcribe"`
	// Additional query parameters for the API request
	AdditionalQueryParams string `name:"query-params" value:"api-version=2025-03-01-preview" usage:"Query params"`
	// Capture mode: enable audio capture from microphone
	Capture bool `name:"capture" usage:"Capture audio from microphone instead of using file/stdin"`
	// Duration specifies the capture duration (e.g., "10s", "1m")
	CaptureDuration string `name:"duration" value:"0s" usage:"Capture duration (e.g., 10s, 1m)"`
	// Device specifies the audio device name (defaults to system default)
	CaptureDevice string `name:"device" usage:"Audio device name (defaults to system default)"`
	// SampleRate specifies the sample rate in Hz (overrides device default)
	CaptureSampleRate string `name:"rate" usage:"Sample rate in Hz (overrides device default)"`
	// Channels specifies the number of audio channels
	CaptureChannels int `name:"channels" value:"2" usage:"Number of audio channels"`
	// BitDepth specifies the audio bit depth
	CaptureBitDepth int `name:"bit-depth" value:"16" usage:"Audio bit depth"`
	// EnableSilence enables silence-based auto-stop
	CaptureEnableSilence bool `name:"silence" usage:"Enable silence-based auto-stop"`
	// SilenceThreshold specifies the silence detection threshold (0.0-1.0)
	CaptureSilenceThreshold float64 `name:"silence-threshold" value:"0.01" usage:"Silence detection threshold (0.0-1.0)"`
	// SilenceMinDuration specifies the minimum silence duration to trigger stop (e.g., "1s")
	CaptureSilenceMinDuration string `name:"silence-min-duration" value:"1s" usage:"Min silence duration to stop"`
	// Clipboard enables copying transcription result to clipboard
	Clipboard bool `name:"clipboard" usage:"Copy transcription result to clipboard"`
	// OutputFormat specifies the output format (none, text, json)
	OutputFormat string `name:"output-format" value:"text" usage:"Output format (none, text, json)"`
}

// validate validates the TranscribeConfig and returns an error if required fields are missing.
func (c *TranscribeConfig) validate() error {
	if c.APIKey == "" {
		return fmt.Errorf("API key is required (use --api-key or set API_KEY environment variable)")
	}
	switch c.OutputFormat {
	case "none", "text", "json":
		// Valid output formats
	default:
		return fmt.Errorf("invalid output format: %s (valid values: none, text, json)", c.OutputFormat)
	}
	return nil
}

// runCaptureToReader captures audio and returns an io.Reader for the captured audio data.
func runCaptureToReader(baseConfig *Config, config *TranscribeConfig) io.Reader {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger := baseConfig.getLogger()
	slog.SetDefault(logger)

	sp, err := audio.NewSystemProfiler(ctx)
	if err != nil {
		panic(fmt.Errorf("failed to initialize system profiler: %w", err))
	}

	spDevices := sp.ListDevices()

	// Resolve device
	var selectedDevice audio.Device
	if config.CaptureDevice == "" {
		selectedDevice, err = audio.GetDefaultSource(spDevices)
		if err != nil {
			panic(fmt.Errorf("failed to get default source device: %w", err))
		}
	} else {
		selectedDevice, err = audio.GetDevice(config.CaptureDevice, spDevices)
		if err != nil {
			panic(fmt.Errorf("device '%s' not found", config.CaptureDevice))
		}
	}

	// Parse and set sample rate if provided
	if config.CaptureSampleRate != "" {
		rate, err := strconv.Atoi(config.CaptureSampleRate)
		if err != nil {
			panic(fmt.Errorf("invalid sample rate: %w", err))
		}
		selectedDevice.SampleRate = rate
	}

	// Parse silence min duration if silence is enabled
	var minSilenceDuration time.Duration
	if config.CaptureEnableSilence {
		minSilenceDuration, err = time.ParseDuration(config.CaptureSilenceMinDuration)
		if err != nil {
			panic(fmt.Errorf("invalid silence min duration: %w", err))
		}
	}

	// Parse capture duration
	duration, err := time.ParseDuration(config.CaptureDuration)
	if err != nil {
		panic(fmt.Errorf("invalid capture duration: %w", err))
	}

	// Print individual fields to avoid JSON serialization issues
	slog.Debug("Starting audio capture for transcription",
		"device_name", selectedDevice.Name,
		"duration", duration,
		"silence_enabled", config.CaptureEnableSilence)

	reader, err := audio.LimitedCapture(ctx, logger, selectedDevice, audio.LimitedCaptureArgs{
		EnableSilence:      config.CaptureEnableSilence,
		SilenceThreshold:   config.CaptureSilenceThreshold,
		SilenceMinDuration: minSilenceDuration,
		Duration:           duration,
		Channels:           config.CaptureChannels,
		BitDepth:           config.CaptureBitDepth,
	})
	if err != nil {
		panic(fmt.Errorf("audio capture failed: %w", err))
	}

	// Convert the raw audio to FLAC format in memory
	var buf bytes.Buffer
	err = audio.ConvertAudio(ctx, logger, audio.ConvertAudioArgs{
		Reader:       reader,
		Writer:       &buf,
		SourceFormat: "raw",
		TargetFormat: "flac",
		SampleRate:   selectedDevice.SampleRate,
		Channels:     config.CaptureChannels,
		BitDepth:     config.CaptureBitDepth,
	})
	if err != nil {
		panic(fmt.Errorf("audio conversion failed: %w", err))
	}

	if baseConfig.Verbose {
		slog.Debug("Audio capture completed successfully")
	}

	return &buf
}

// runTranscribe executes the audio transcription logic.
func runTranscribe(baseConfig *Config, config *TranscribeConfig, audioReader io.Reader) error {
	ctx := context.Background()
	_ = ctx

	logger := baseConfig.getLogger()
	slog.SetDefault(logger)

	// Create a temporary file for the audio data
	tempFile, err := os.CreateTemp("", "sttrouter-transcribe-*.flac")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer func() { _ = os.Remove(tempFile.Name()) }() // Clean up temp file
	defer func() { _ = tempFile.Close() }()

	// Copy audio data from reader to temp file
	_, err = io.Copy(tempFile, audioReader)
	if err != nil {
		return fmt.Errorf("failed to copy audio data to temp file: %w", err)
	}

	// Close the temp file so it can be opened by the transcriber
	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	client := openaix.NewClient(config.APIKey, config.BaseURL, config.AdditionalQueryParams)

	// Prepare transcription request
	req := openaix.TranscriptionRequest{
		File:           tempFile.Name(),
		Model:          config.Model,
		Language:       config.Language,
		ResponseFormat: config.ResponseFormat,
		Temperature:    config.Temperature,
	}

	t, err := client.Transcribe(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to transcribe audio: %w", err)
	}

	transcription := t.Text

	// Handle output based on format
	switch config.OutputFormat {
	case "none":
		// No output
	case "text":
		if config.Clipboard {
			reader := bytes.NewReader([]byte(transcription))
			if err := clipboard.CopyToClipboard(ctx, logger, reader); err != nil {
				return fmt.Errorf("failed to copy to clipboard: %w", err)
			}
			if baseConfig.Verbose {
				slog.Debug("Transcription copied to clipboard")
			}
		} else {
			// Print to stdout by default
			fmt.Println(transcription)
		}
	case "json":
		// Output full JSON response
		jsonOutput := fmt.Sprintf(`{"text": %q}`, transcription)
		if config.Clipboard {
			reader := bytes.NewReader([]byte(jsonOutput))
			if err := clipboard.CopyToClipboard(ctx, logger, reader); err != nil {
				return fmt.Errorf("failed to copy to clipboard: %w", err)
			}
			if baseConfig.Verbose {
				slog.Debug("Transcription JSON copied to clipboard")
			}
		} else {
			fmt.Println(jsonOutput)
		}
	}

	if baseConfig.Verbose {
		slog.Debug("Transcription completed successfully")
	}

	return nil
}

func NewTranscribeCommand() *cli.Command {
	var baseConfig Config
	var transcribeConfig TranscribeConfig
	baseFlags := flagtags.MustParseFlags(&baseConfig)
	transcribeFlags := flagtags.MustParseFlags(&transcribeConfig)
	flags := append(baseFlags, transcribeFlags...)

	return &cli.Command{
		Name:      "transcribe",
		Usage:     "Transcribe audio from microphone or file to text using OpenAI Whisper API",
		ArgsUsage: "[AUDIO_FILE]",
		Description: `Transcribe audio to text using OpenAI's Whisper API.

When --capture is specified, audio is captured from the microphone and transcribed.
Otherwise, an AUDIO_FILE argument specifies the audio file to transcribe.
Use "-" to read audio data from stdin when not using --capture.
Supported formats include FLAC, MP3, MP4, MPEG, MPGA, M4A, OGG, WAV, and WEBM.

Output format can be controlled with --output-format:
- text: Plain text output (default)
- json: JSON formatted output
- none: No output

Use --clipboard to copy results to clipboard instead of stdout.

Examples:
  # Capture and transcribe from microphone
  sttrouter transcribe --capture --api-key YOUR_KEY

  # Transcribe a FLAC file
  sttrouter transcribe --api-key YOUR_KEY recording.flac

  # Transcribe and copy to clipboard
  sttrouter transcribe --api-key YOUR_KEY --clipboard recording.flac

  # Transcribe with JSON output
  sttrouter transcribe --api-key YOUR_KEY --output-format json recording.flac

  # Transcribe with no output (silent)
  sttrouter transcribe --api-key YOUR_KEY --output-format none recording.flac

  # Transcribe from stdin
  sttrouter transcribe --api-key YOUR_KEY -

  # Transcribe to JSON format (legacy flag still works)
  sttrouter transcribe --api-key YOUR_KEY --response-format json recording.flac`,
		Flags: flags,
		Action: func(c *cli.Context) error {
			var audioReader io.Reader

			if transcribeConfig.Capture {
				// Capture mode: no arguments needed
				if c.NArg() > 0 {
					return fmt.Errorf("no arguments expected when using --capture")
				}
				audioReader = runCaptureToReader(&baseConfig, &transcribeConfig)
			} else {
				// File/stdin mode: require exactly one argument
				if c.NArg() != 1 {
					return fmt.Errorf("exactly one argument (AUDIO_FILE or - for stdin) is required")
				}

				audioArg := c.Args().Get(0)
				if audioArg == "-" {
					audioReader = os.Stdin
				} else {
					file, err := os.Open(audioArg)
					if err != nil {
						return fmt.Errorf("failed to open audio file: %w", err)
					}
					defer func() { _ = file.Close() }()
					audioReader = file
				}
			}

			if err := baseConfig.validate(); err != nil {
				return err
			}

			if err := transcribeConfig.validate(); err != nil {
				return err
			}

			return runTranscribe(&baseConfig, &transcribeConfig, audioReader)
		},
	}
}
