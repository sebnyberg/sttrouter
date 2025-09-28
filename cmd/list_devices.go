package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"slices"
	"sort"
	"text/tabwriter"

	"github.com/sebnyberg/flagtags"
	"github.com/sebnyberg/sttrouter/audio"
	"github.com/urfave/cli/v2"
)

// runListDevices executes the device listing logic.
func runListDevices(config *Config) error {
	ctx := context.Background()

	logger := config.getLogger()
	slog.SetDefault(logger)

	ffmpeg, err := audio.NewFFmpeg(ctx)
	if err != nil {
		return fmt.Errorf("failed to list devices with ffmpeg: %w", err)
	}

	sp, err := audio.NewSystemProfiler(ctx)
	if err != nil {
		return fmt.Errorf("failed to list devices with system_profiler: %w", err)
	}

	ffmpegDevices := ffmpeg.ListDeviceNames()
	spDevices := sp.ListDeviceNames()

	// Sort both lists
	sort.Strings(ffmpegDevices)
	sort.Strings(spDevices)

	// Verify they match
	if !slices.Equal(ffmpegDevices, spDevices) {
		return fmt.Errorf("device lists do not match: ffmpeg=%v, system_profiler=%v", ffmpegDevices, spDevices)
	}

	defaultDev := sp.GetDefaultDeviceName()

	if config.Verbose {
		fmt.Printf("Found %d devices\n", len(ffmpegDevices))
	}

	// Output using tabwriter for aligned columns
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
	if _, err := fmt.Fprintln(w, "DEVICE_NAME|DEFAULT"); err != nil {
		return err
	}

	// Output devices
	for _, dev := range ffmpegDevices {
		isDefault := dev == defaultDev
		if _, err := fmt.Fprintf(w, "%s|%t\n", dev, isDefault); err != nil {
			return err
		}
	}

	if err := w.Flush(); err != nil {
		return err
	}

	return nil
}

func NewListDevicesCommand() *cli.Command {
	var config Config
	flags := flagtags.MustParseFlags(&config)

	return &cli.Command{
		Name:  "list-devices",
		Usage: "List available audio input devices",
		Description: `List available audio input devices for recording.

This command enumerates audio devices using both ffmpeg and system-profiler,
verifies they match, and outputs devices in lexicographical order with default marked.`,
		Flags: flags,
		Action: func(c *cli.Context) error {
			if err := config.validate(); err != nil {
				return err
			}
			return runListDevices(&config)
		},
	}
}
