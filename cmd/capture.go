package cmd

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/sebnyberg/flagtags"
	"github.com/sebnyberg/sttrouter/audio"
	"github.com/urfave/cli/v2"
)

// CaptureConfig holds capture specific configuration flags.
type CaptureConfig struct {
	// Duration specifies the capture duration (e.g., "10s", "1m")
	Duration string `name:"duration" value:"0" usage:"Capture duration (e.g., 10s, 1m)"`
	// Device specifies the audio device name (defaults to system default)
	Device string `name:"device" usage:"Audio device name (defaults to system default)"`
	// SampleRate specifies the sample rate in Hz (overrides device default)
	SampleRate int `name:"rate" usage:"Sample rate in Hz (overrides device default)"`
	// Channels specifies the number of audio channels
	Channels int `name:"channels" value:"2" usage:"Number of audio channels"`
	// BitDepth specifies the audio bit depth
	BitDepth int `name:"bit-depth" value:"16" usage:"Audio bit depth"`
	// EnableSilence enables silence-based auto-stop
	EnableSilence bool `name:"auto-stop" usage:"Enable auto-stop when silence is detected"`
	// SilenceThreshold specifies the silence detection threshold (0.0-1.0)
	SilenceThreshold float64 `name:"silence-threshold" value:"0.01" usage:"Silence detection threshold (0.0-1.0)"`
	// SilenceMinDuration specifies the minimum silence duration to trigger stop (e.g., "1s")
	SilenceMinDuration string `name:"silence-min-duration" value:"1s" usage:"Minimum silence duration to trigger stop"`
}

func (c *CaptureConfig) validate() error {
	if c.Duration != "" {
		if _, err := time.ParseDuration(c.Duration); err != nil {
			return fmt.Errorf("invalid capture duration '%v', %w", c.Duration, err)
		}
	}
	if c.SampleRate < 0 {
		return fmt.Errorf("sample rate must be >= 0, was '%v", c.SampleRate)
	}
	if c.Channels <= 0 || c.Channels > 2 {
		return fmt.Errorf("channels must be 1 or 2, was '%v'", c.Channels)
	}
	if c.BitDepth <= 0 {
		return fmt.Errorf("bit depth must be greater than 0, was '%v'", c.BitDepth)
	}
	const eps = 1e-5
	if c.SilenceThreshold <= 0 || c.SilenceThreshold > 1.0+eps {
		return fmt.Errorf("silence threshold must be in the interval (0,1.0], was '%v'", c.SilenceThreshold)
	}
	if _, err := time.ParseDuration(c.SilenceMinDuration); err != nil {
		return fmt.Errorf("invalid silence min duration '%v', %w", c.SilenceMinDuration, err)
	}
	return nil
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
	if config.SampleRate != 0 {
		selectedDevice.SampleRate = config.SampleRate
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
	slog.Debug("Starting audio capture",
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

	pipeReader, pipeWriter := io.Pipe()

	// Run LimitedCapture in a goroutine so it can write to the pipe concurrently with ConvertAudio reading
	go func() {
		err := audio.LimitedCapture(ctx, logger, selectedDevice, audio.LimitedCaptureArgs{
			EnableSilence:      config.EnableSilence,
			SilenceThreshold:   config.SilenceThreshold,
			SilenceMinDuration: minSilenceDuration,
			Duration:           duration,
			Channels:           config.Channels,
			BitDepth:           config.BitDepth,
			Writer:             pipeWriter,
		})
		if err != nil {
			slog.Error("Audio capture failed", "error", err, "device", selectedDevice)
			// Note: error handling in goroutine - in a real implementation you might need to signal this back
		}
	}()

	err = audio.ConvertAudio(ctx, logger, audio.ConvertAudioArgs{
		Reader:       pipeReader,
		Writer:       writer,
		SourceFormat: "raw",
		TargetFormat: "flac",
		SampleRate:   selectedDevice.SampleRate,
		Channels:     config.Channels,
		BitDepth:     config.BitDepth,
	})
	if err != nil {
		return fmt.Errorf("audio conversion failed: %w", err)
	}

	if baseConfig.Verbose {
		slog.Debug("Audio capture completed successfully")
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
Use "-" to output to stdout.
Output format is always FLAC.

Examples:
  # Output to file
  sttrouter capture recording.flac
  
  # Output to stdout
  sttrouter capture -`,
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
