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

// OpenAIConfig holds OpenAI-specific configuration.
type OpenAIConfig struct {
	// APIKey is the Azure OpenAI API key
	APIKey string `name:"api-key"`
	// BaseURL is the API base URL
	BaseURL string `name:"base-url"`
}

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
	// OpenAI configuration
	OpenAI OpenAIConfig `name:"openai"`
	// Additional query parameters for the API request
	AdditionalQueryParams string `name:"query-params" value:"api-version=2025-03-01-preview" usage:"Query params"`
	// Configuration for audio capture
	Capture CaptureConfig
	// NoClipboard disables copying transcription result to clipboard
	NoClipboard bool `name:"no-clipboard" usage:"Disable copying transcription result to clipboard"`
	// NoCapture disables audio capture and uses a provided file instead
	NoCapture bool `name:"no-capture" usage:"Disable audio capture and use provided file for transcription"`
	// OutputFormat specifies the output format (none, text)
	OutputFormat string `name:"output-format" value:"text" usage:"Output format (none, text)"`
	// Debug enables debug mode, keeping temp files and printing their locations
	Debug bool `name:"debug" usage:"Enable debug mode (keeps temp files and prints locations)"`
}

// validate validates the TranscribeConfig and returns an error if required fields are missing.
func (c *TranscribeConfig) validate() error {
	if !c.NoCapture {
		if err := c.Capture.validate(); err != nil {
			return fmt.Errorf("capture config validation err, %w", err)
		}
	}
	if c.OpenAI.APIKey == "" {
		return fmt.Errorf("API key is required (use --openai-api-key or set OPENAI_API_KEY environment variable)")
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

	lister, err := audio.NewDeviceLister(ctx)
	if err != nil {
		panic(fmt.Errorf("failed to initialize device lister: %w", err))
	}

	devices := lister.ListDevices()

	// Resolve device
	var selectedDevice audio.Device
	if config.Capture.Device == "" {
		selectedDevice, err = audio.GetDefaultSource(devices)
		if err != nil {
			panic(fmt.Errorf("failed to get default source device: %w", err))
		}
	} else {
		selectedDevice, err = audio.GetDevice(config.Capture.Device, devices)
		if err != nil {
			panic(fmt.Errorf("device '%s' not found", config.Capture.Device))
		}
	}

	// Parse and set sample rate if provided
	if config.Capture.SampleRate != 0 {
		selectedDevice.SampleRate = config.Capture.SampleRate
	}

	// Parse auto-stop min duration if auto-stop is enabled
	var minSilenceDuration time.Duration
	if !config.Capture.NoAutoStop {
		minSilenceDuration, err = time.ParseDuration(config.Capture.AutoStopMinDuration)
		if err != nil {
			panic(fmt.Errorf("invalid auto-stop min duration: %w", err))
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
		"auto_stop_enabled", !config.Capture.NoAutoStop)

	pipeReader, pipeWriter := io.Pipe()

	// Run LimitedCapture in a goroutine so it can write to the pipe concurrently with ConvertAudio reading
	go func() {
		err = audio.LimitedCapture(ctx, logger, selectedDevice, audio.LimitedCaptureArgs{
			EnableAutoStop:      !config.Capture.NoAutoStop,
			AutoStopThreshold:   config.Capture.AutoStopThreshold,
			AutoStopMinDuration: minSilenceDuration,
			Duration:            duration,
			Channels:            config.Capture.Channels,
			BitDepth:            config.Capture.BitDepth,
			Writer:              pipeWriter,
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
func runTranscribe(baseConfig *Config, config *TranscribeConfig, inputFile string) error {
	ctx := context.Background()

	logger := baseConfig.getLogger()
	slog.SetDefault(logger)

	var audioFilePath string
	if config.NoCapture {
		// Use the provided file directly
		audioFilePath = inputFile
		fmt.Println("Using provided audio file for transcription")
	} else {
		// Create temp file and capture audio to it
		tempFile, err := os.CreateTemp("", "sttrouter-capture-*.flac")
		if err != nil {
			return fmt.Errorf("failed to create temp file: %w", err)
		}
		if !config.Debug {
			defer func() { _ = os.Remove(tempFile.Name()) }()
		}
		slog.Debug("tempfile created", "path", tempFile.Name())
		if config.Debug {
			fmt.Printf("Debug: temp file created at %s\n", tempFile.Name())
		}
		fmt.Println("Audio capture started")
		if err := runCaptureToWriter(baseConfig, config, tempFile); err != nil {
			return err
		}
		// Flush by closing the file
		if err := tempFile.Close(); err != nil {
			return fmt.Errorf("failed to close the temporary audio file, %w", err)
		}
		fmt.Println("Audio capture completed")
		slog.Info("capture completed")
		audioFilePath = tempFile.Name()
	}

	client := openaix.NewClient(config.OpenAI.APIKey, config.OpenAI.BaseURL, config.AdditionalQueryParams)

	// Prepare transcription request
	req := openaix.TranscriptionRequest{
		File:           audioFilePath,
		Model:          config.Model,
		Language:       config.Language,
		ResponseFormat: config.ResponseFormat,
		Temperature:    config.Temperature,
	}

	fmt.Println("Transcription started")
	t, err := client.Transcribe(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to transcribe audio: %w", err)
	}
	fmt.Println("Transcription completed")
	slog.Info("transcription completed")

	transcription := t.Text
	if !config.NoClipboard {
		bs := bytes.NewBufferString(t.Text)
		if err := clipboard.CopyToClipboard(ctx, logger, bs); err != nil {
			return fmt.Errorf("failed to copy transcription output to clipboard, %w", err)
		}
		fmt.Println("Transcription copied to clipboard")
		slog.Info("transcription copied to clipboard")
	}

	// Handle output based on format
	switch config.OutputFormat {
	case "none":
		// No output
	case "text":
		fmt.Println()
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
		ArgsUsage: "[FILE]",
		Description: `Capture audio from the microphone and transcribe it to text using Azure OpenAI's GPT-4o.

Audio is captured from the microphone, converted to FLAC format,
and sent to GPT-4o for transcription.

Use --no-capture to skip audio capture and transcribe an existing audio file instead.
When --no-capture is used, FILE is a required positional argument.

By default, transcription results are copied to the clipboard. Use --no-clipboard to disable this.

Output format can be controlled with --output-format:
- none: No stdout output
- text: Plain text output to stdout (default)

Examples:
  # Capture and transcribe from microphone (clipboard default)
  sttrouter transcribe --api-key YOUR_KEY

  # Capture and transcribe without copying to clipboard
  sttrouter transcribe --api-key YOUR_KEY --no-clipboard

  # Capture and output to stdout in addition to clipboard
  sttrouter transcribe --api-key YOUR_KEY --output-format text

  # Transcribe an existing audio file
  sttrouter transcribe --no-capture --api-key YOUR_KEY recording.flac`,
		Flags: flags,
		Action: func(c *cli.Context) error {
			if transcribeConfig.NoCapture {
				if c.NArg() != 1 {
					return fmt.Errorf("exactly one argument (FILE) is required when using --no-capture")
				}
			} else {
				if c.NArg() > 0 {
					return fmt.Errorf("no arguments expected when capturing from microphone")
				}
			}

			if err := baseConfig.validate(); err != nil {
				return err
			}

			if err := transcribeConfig.validate(); err != nil {
				return err
			}

			return runTranscribe(&baseConfig, &transcribeConfig, c.Args().Get(0))
		},
	}
}
