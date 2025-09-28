package cmd

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/sebnyberg/flagtags"
	"github.com/sebnyberg/sttrouter/audio"
	"github.com/urfave/cli/v2"
)

// CaptureConfig holds capture specific configuration flags.
type CaptureConfig struct {
	// Format specifies the output format/container (wav or flac)
	Format string `name:"f" usage:"Output format/container: wav or flac (defaults to file extension)"`
	// Duration specifies the capture duration (e.g., "10s", "1m")
	Duration string `name:"duration" value:"10s" usage:"Capture duration (e.g., 10s, 1m)"`
	// Device specifies the audio device name (defaults to system default)
	Device string `name:"device" usage:"Audio device name (defaults to system default)"`
}

// parse parses and validates the format configuration.
func parseOutputFormat(c *CaptureConfig, outputFile string) (format string, err error) {
	// Try to parse the format from the outputfile
	extIdx := strings.LastIndexByte(outputFile, '.')
	var fileFormat string
	if extIdx != -1 {
		ext := outputFile[extIdx+1:]
		extLower := strings.ToLower(ext)
		fileFormat = extLower
	}

	// If format is unspecified, use the file format
	if c.Format == "" {
		format = fileFormat
	}

	// Validate format
	if format != "flac" && format != "wav" {
		return "", errors.New("file format (--format or file extension) must be 'wav' or 'flac'")
	}

	return format, nil
}

// runCapture executes the audio capture logic.
func runCapture(baseConfig *Config, config *CaptureConfig, outputFile string) error {
	ctx := context.Background()

	logger := baseConfig.getLogger()
	slog.SetDefault(logger)

	sp, err := audio.NewSystemProfiler(ctx)
	if err != nil {
		return fmt.Errorf("failed to initialize system profiler: %w", err)
	}

	sox, err := audio.NewSox(ctx)
	if err != nil {
		return fmt.Errorf("failed to initialize sox: %w", err)
	}

	spDevices := sp.ListDevices()

	// Resolve device
	var selectedDevice audio.Device
	if config.Device == "" {
		selectedDevice, err = audio.GetDefaultSource(spDevices)
		if err != nil {
			return fmt.Errorf("failed to get default source device: %w", err)
		}
	} else {
		selectedDevice, err = audio.GetDevice(config.Device, spDevices)
		if err != nil {
			return fmt.Errorf("device '%s' not found", config.Device)
		}
	}

	// Parse format
	format, err := parseOutputFormat(config, outputFile)
	if err != nil {
		return fmt.Errorf("failed to parse format: %w", err)
	}

	// Sentinel: set config.Format to "invalid"
	config.Format = "invalid"

	// Print individual fields to avoid JSON serialization issues
	slog.Info("Starting audio capture",
		"device_name", selectedDevice.Name,
		"duration", config.Duration,
		"format", format,
		"output", outputFile)

	err = selectedDevice.CaptureAudio(sox, ctx, logger, config.Duration, format, outputFile)
	if err != nil {
		slog.Error("Audio capture failed", "error", err, "device", selectedDevice)
		return fmt.Errorf("audio capture failed: %w", err)
	}

	if baseConfig.Verbose {
		slog.Info("Audio capture completed successfully")
	}

	return nil
}

func NewCaptureCommand() *cli.Command {
	var baseConfig Config
	var captureConfig CaptureConfig
	baseFlags := flagtags.MustParseFlags(&baseConfig)
	captureFlags := flagtags.MustParseFlags(&captureConfig)
	flags := append(baseFlags, captureFlags...)

	return &cli.Command{
		Name:      "capture",
		Usage:     "Capture audio from microphone to a file or stdout",
		ArgsUsage: "OUTPUT_FILE",
		Description: `Capture audio from the specified microphone device using Sox.

The OUTPUT_FILE is a required positional argument that specifies where to send the audio output.
Use "-" to output to stdout (similar to how ffmpeg works).
Format is inferred from file extension when a regular file is specified,
or must be specified with -f/--format when outputting to stdout.

Examples:
  # Output to file (format inferred from extension)
  sttrouter capture recording.flac
  
  # Output to stdout (format must be specified)
  sttrouter capture -f flac -
  
  # Pipe to another command (format must be specified)
  sttrouter capture -f wav - | ffplay -`,
		Flags: flags,
		Action: func(c *cli.Context) error {
			if c.NArg() != 1 {
				return fmt.Errorf("exactly one argument (OUTPUT_FILE) is required")
			}

			outputFile := c.Args().Get(0)

			if err := baseConfig.validate(); err != nil {
				return err
			}

			return runCapture(&baseConfig, &captureConfig, outputFile)
		},
	}
}
