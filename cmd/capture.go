package cmd

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/sebnyberg/flagtags"
	"github.com/sebnyberg/sttrouter/audio"
	"github.com/urfave/cli/v2"
)

// CaptureConfig holds capture specific configuration flags.
type CaptureConfig struct {
	// Duration specifies the capture duration (e.g., "10s", "1m")
	Duration string `name:"duration" value:"10s" usage:"Capture duration (e.g., 10s, 1m)"`
	// Device specifies the audio device name (defaults to system default)
	Device string `name:"device" usage:"Audio device name (defaults to system default)"`
	// SampleRate specifies the sample rate in Hz (overrides device default)
	SampleRate string `name:"rate" usage:"Sample rate in Hz (overrides device default)"`
	// EnableSilence enables silence-based auto-stop
	EnableSilence bool `name:"silence" usage:"Enable silence-based auto-stop"`
	// SilenceThreshold specifies the silence detection threshold (0.0-1.0)
	SilenceThreshold float64 `name:"silence-threshold" value:"0.01" usage:"Silence detection threshold (0.0-1.0)"`
	// SilenceMinDuration specifies the minimum silence duration to trigger stop (e.g., "1s")
	SilenceMinDuration string `name:"silence-min-duration" value:"1s" usage:"Minimum silence duration to trigger stop"`
}

// runCapture executes the audio capture logic.
func runCapture(baseConfig *Config, config *CaptureConfig, outputFile string, duration time.Duration) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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

	// Parse and set sample rate if provided
	if config.SampleRate != "" {
		rate, err := strconv.Atoi(config.SampleRate)
		if err != nil {
			return fmt.Errorf("invalid sample rate: %w", err)
		}
		selectedDevice.SampleRate = rate
	}

	// Parse silence min duration if silence is enabled
	var minSilenceDuration time.Duration
	if config.EnableSilence {
		minSilenceDuration, err = time.ParseDuration(config.SilenceMinDuration)
		if err != nil {
			return fmt.Errorf("invalid silence min duration: %w", err)
		}
	}

	// Print individual fields to avoid JSON serialization issues
	slog.Info("Starting audio capture",
		"device_name", selectedDevice.Name,
		"duration", duration,
		"output", outputFile,
		"silence_enabled", config.EnableSilence)

	// Asynchronously read from the capture inputs into the Converter, which in
	// turn writes to the FLAC output
	var writer io.Writer
	var file *os.File
	if outputFile == "-" {
		writer = os.Stdout
	} else {
		var err error
		file, err = os.Create(outputFile)
		if err != nil {
			return fmt.Errorf("failed to create file, %w", err)
		}
		writer = file
		defer func() { _ = file.Close() }()
	}

	reader, err := audio.LimitedCapture(ctx, logger, selectedDevice, config.EnableSilence, config.SilenceThreshold, minSilenceDuration, duration)
	if err != nil {
		slog.Error("Audio capture failed", "error", err, "device", selectedDevice)
		return fmt.Errorf("audio capture failed: %w", err)
	}

	err = sox.ConvertAudio(ctx, logger, reader, writer, "raw", "flac", selectedDevice.SampleRate, 2, 16)
	if err != nil {
		return fmt.Errorf("audio conversion failed: %w", err)
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
Output format is always FLAC.

Examples:
  # Output to file
  sttrouter capture recording.flac
  
  # Output to stdout
  sttrouter capture -
  
  # Pipe to another command
  sttrouter capture - | ffplay -`,
		Flags: flags,
		Action: func(c *cli.Context) error {
			if c.NArg() != 1 {
				return fmt.Errorf("exactly one argument (OUTPUT_FILE) is required")
			}

			outputFile := c.Args().Get(0)

			if err := baseConfig.validate(); err != nil {
				return err
			}

			duration, err := time.ParseDuration(captureConfig.Duration)
			if err != nil {
				return fmt.Errorf("invalid duration: %w", err)
			}

			return runCapture(&baseConfig, &captureConfig, outputFile, duration)
		},
	}
}
