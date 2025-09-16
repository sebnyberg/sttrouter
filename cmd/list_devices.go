package cmd

import (
	"fmt"
	"os"
	"reflect"
	"sort"
	"text/tabwriter"

	"github.com/sebnyberg/sttrouter/audio"
	"github.com/spf13/cobra"
)

// listDevicesCmd represents the list-devices command
var listDevicesCmd = &cobra.Command{
	Use:   "list-devices",
	Short: "List available audio devices",
	Long: `List available audio input devices for recording.

This command enumerates audio devices using both ffmpeg and system-profiler,
verifies they match, and outputs devices in lexicographical order with default marked.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

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
		if !reflect.DeepEqual(ffmpegDevices, spDevices) {
			return fmt.Errorf("device lists do not match: ffmpeg=%v, system_profiler=%v", ffmpegDevices, spDevices)
		}

		defaultDev := sp.GetDefaultDeviceName()

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
	},
}

func init() {
	rootCmd.AddCommand(listDevicesCmd)
}
