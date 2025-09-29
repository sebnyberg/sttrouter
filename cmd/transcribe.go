package cmd

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/sebnyberg/flagtags"
	"github.com/sebnyberg/sttrouter/audio"
	"github.com/sebnyberg/sttrouter/clipboard"
	"github.com/sebnyberg/sttrouter/openaix"
	"github.com/urfave/cli/v2"
)

// TranscribeConfig holds transcribe specific configuration flags.
type TranscribeConfig struct {
	// Model specifies the GPT-4o model to use
	Model string `name:"model" value:"gpt-4o-transcribe" usage:"Model to use for transcription"`
	// Language specifies the language code
	Language string `name:"language" value:"en" usage:"Language code (e.g., 'en', 'es')"`
	// ResponseFormat specifies the response format (json, text, srt, verbose_json, vtt)
	ResponseFormat string `name:"response-format" value:"text" usage:"Response format (json,text,srt,verbose_json,vtt)"`
	// Temperature specifies the sampling temperature (0.0 to 1.0)
	Temperature float64 `name:"temperature" value:"0" usage:"Sampling temperature (0.0 to 1.0)"`
	// Azure OpenAI API Key
	APIKey string `name:"api-key"`
	// API Base URL
	BaseURL string `name:"base-url" value:"https://seblab-ai.openai.azure.com/openai/deployments/gpt-4o-transcribe"`
	// Additional query parameters for the API request
	AdditionalQueryParams string `name:"query-params" value:"api-version=2025-03-01-preview" usage:"Query params"`
	// Capture mode: enable audio capture from microphone
	DoCapture bool `name:"capture" usage:"Capture audio from microphone instead of using file/stdin"`
	// Configuration for audio capture
	Capture CaptureConfig
	// Whether to transcribe to clipboard
	Clipboard bool `name:"clipboard" usage:"Copy transcription result to clipboard"`
	// OutputFormat specifies the output format (none, text)
	OutputFormat string `name:"output-format" value:"text" usage:"Output format (none, text)"`
}

// validate validates the TranscribeConfig and returns an error if required fields are missing.
func (c *TranscribeConfig) validate() error {
	if err := c.Capture.validate(); err != nil {
		return fmt.Errorf("capture config validation err, %w", err)
	}
	if c.APIKey == "" {
		return fmt.Errorf("API key is required (use --api-key or set API_KEY environment variable)")
	}
	switch c.OutputFormat {
	case "none", "text":
		// Valid output formats
	default:
		return fmt.Errorf("invalid output format: %s (valid values: none, text)", c.OutputFormat)
	}
	return nil
}

// runCaptureToWriter captures audio and writes it directly to the specified temp file
func runCaptureToWriter(baseConfig *Config, config *TranscribeConfig, resultsWriter io.Writer) error {
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
	if config.Capture.Device == "" {
		selectedDevice, err = audio.GetDefaultSource(spDevices)
		if err != nil {
			panic(fmt.Errorf("failed to get default source device: %w", err))
		}
	} else {
		selectedDevice, err = audio.GetDevice(config.Capture.Device, spDevices)
		if err != nil {
			panic(fmt.Errorf("device '%s' not found", config.Capture.Device))
		}
	}

	// Parse and set sample rate if provided
	if config.Capture.SampleRate != 0 {
		selectedDevice.SampleRate = config.Capture.SampleRate
	}

	// Parse silence min duration if silence is enabled
	var minSilenceDuration time.Duration
	if config.Capture.EnableSilence {
		minSilenceDuration, err = time.ParseDuration(config.Capture.SilenceMinDuration)
		if err != nil {
			panic(fmt.Errorf("invalid silence min duration: %w", err))
		}
	}

	// Parse capture duration
	duration, err := time.ParseDuration(config.Capture.Duration)
	if err != nil {
		panic(fmt.Errorf("invalid capture duration: %w", err))
	}

	// Print individual fields to avoid JSON serialization issues
	slog.Debug("Starting audio capture for transcription",
		"device_name", selectedDevice.Name,
		"duration", duration,
		"silence_enabled", config.Capture.EnableSilence)

	pipeReader, pipeWriter := io.Pipe()

	// Run LimitedCapture in a goroutine so it can write to the pipe concurrently with ConvertAudio reading
	go func() {
		err = audio.LimitedCapture(ctx, logger, selectedDevice, audio.LimitedCaptureArgs{
			EnableSilence:      config.Capture.EnableSilence,
			SilenceThreshold:   config.Capture.SilenceThreshold,
			SilenceMinDuration: minSilenceDuration,
			Duration:           duration,
			Channels:           config.Capture.Channels,
			BitDepth:           config.Capture.BitDepth,
			Writer:             pipeWriter,
		})
		if err != nil {
			panic(fmt.Errorf("audio capture failed: %w", err))
		}
	}()

	// Convert the raw audio to FLAC format and write directly to temp file
	err = audio.ConvertAudio(ctx, logger, audio.ConvertAudioArgs{
		Reader:       pipeReader,
		Writer:       resultsWriter,
		SourceFormat: "raw",
		TargetFormat: "flac",
		SampleRate:   selectedDevice.SampleRate,
		Channels:     config.Capture.Channels,
		BitDepth:     config.Capture.BitDepth,
	})
	if err != nil {
		return fmt.Errorf("audio conversion failed: %w", err)
	}

	return nil
}

// runTranscribe executes the audio transcription logic.
func runTranscribe(baseConfig *Config, config *TranscribeConfig) error {
	ctx := context.Background()

	logger := baseConfig.getLogger()
	slog.SetDefault(logger)

	// Capture mode: create temp file and capture audio to it
	tempFile, err := os.CreateTemp("", "sttrouter-capture-*.flac")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	slog.Debug("tempfile created", "path", tempFile.Name())
	if err := runCaptureToWriter(baseConfig, config, tempFile); err != nil {
		return err
	}
	if _, err := tempFile.Seek(0, 0); err != nil {
		return fmt.Errorf("failed to seek tempfile back to the beginning, %w", err)
	}
	slog.Info("capture completed")

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
	slog.Info("transcription completed")

	transcription := t.Text
	if config.Clipboard {
		slog.Info("copying transcription to clipboard")
		bs := bytes.NewBufferString(t.Text)
		if err := clipboard.CopyToClipboard(ctx, logger, bs); err != nil {
			return fmt.Errorf("failed to copy transcription output to clipboard, %w", err)
		}
	}

	// Handle output based on format
	switch config.OutputFormat {
	case "none":
		// No output
	case "text":
		fmt.Println(transcription)
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
		Usage:     "Capture audio from microphone and transcribe to text using Azure OpenAI GPT-4o",
		ArgsUsage: "",
		Description: `Capture audio from the microphone and transcribe it to text using Azure OpenAI's GPT-4o.

The --capture flag is required. Audio is captured from the microphone, converted to FLAC format,
and sent to GPT-4o for transcription.

Output format can be controlled with --output-format:
- text: Plain text output (default)
- none: No output

Use --clipboard to copy results to clipboard instead of stdout.

Examples:
  # Capture and transcribe from microphone
  sttrouter transcribe --capture --api-key YOUR_KEY

  # Capture and copy transcription to clipboard
  sttrouter transcribe --capture --api-key YOUR_KEY --clipboard

  # Capture with no output (silent)
  sttrouter transcribe --capture --api-key YOUR_KEY --output-format none`,
		Flags: flags,
		Action: func(c *cli.Context) error {
			// Capture mode: no arguments needed
			if c.NArg() > 0 {
				return fmt.Errorf("no arguments expected")
			}

			if err := baseConfig.validate(); err != nil {
				return err
			}

			if err := transcribeConfig.validate(); err != nil {
				return err
			}

			return runTranscribe(&baseConfig, &transcribeConfig)
		},
	}
}
