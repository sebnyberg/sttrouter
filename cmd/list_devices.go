package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sort"
	"text/tabwriter"

	"github.com/sebnyberg/flagtags"
	"github.com/sebnyberg/sttrouter/audio"
	"github.com/urfave/cli/v2"
)

// ListDevicesResult represents the result of the list-devices command
type ListDevicesResult struct {
	Devices []audio.Device `json:"devices"`
}

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

	ffmpegDevices := ffmpeg.ListDevices()
	spDevices := sp.ListDevices()

	// Merge devices from both sources
	deviceModes := make(map[string]uint)
	for _, dev := range ffmpegDevices {
		deviceModes[dev.Name] |= dev.Mode
	}
	for _, dev := range spDevices {
		deviceModes[dev.Name] |= dev.Mode
	}

	// Check for more than one default device
	defaultCount := 0
	for _, mode := range deviceModes {
		if mode&audio.DeviceFlagIsDefault != 0 {
			defaultCount++
		}
	}
	if defaultCount > 1 {
		return fmt.Errorf("more than one default device found")
	}

	// Create final device list
	var devices []audio.Device
	for name, mode := range deviceModes {
		devices = append(devices, audio.Device{Name: name, Mode: mode})
	}

	// Sort by name
	sort.Slice(devices, func(i, j int) bool {
		return devices[i].Name < devices[j].Name
	})

	if config.Verbose {
		fmt.Printf("Found %d devices\n", len(devices))
	}

	// Output using tabwriter for aligned columns
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
	if _, err := fmt.Fprintln(w, "DEVICE_NAME|INPUT|OUTPUT|DEFAULT"); err != nil {
		return err
	}

	// Output devices
	for _, dev := range devices {
		isInput := dev.Mode&audio.DeviceFlagInput != 0
		isOutput := dev.Mode&audio.DeviceFlagOutput != 0
		isDefault := dev.Mode&audio.DeviceFlagIsDefault != 0
		if _, err := fmt.Fprintf(w, "%s|%t|%t|%t\n", dev.Name, isInput, isOutput, isDefault); err != nil {
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
		Usage: "List available audio devices",
		Description: `List available audio devices for recording and playback.

This command enumerates audio devices using both ffmpeg and system-profiler,
merges the information, and outputs devices in lexicographical order with their capabilities.`,
		Flags: flags,
		Action: func(c *cli.Context) error {
			if err := config.validate(); err != nil {
				return err
			}
			return runListDevices(&config)
		},
	}
}
