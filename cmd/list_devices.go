package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"text/tabwriter"

	"github.com/sebnyberg/flagtags"
	"github.com/sebnyberg/sttrouter/audio"
	"github.com/urfave/cli/v2"
)

// ListDevicesConfig holds list-devices specific configuration flags.
type ListDevicesConfig struct {
	// OutputFormat specifies the output format for the device list
	OutputFormat string `name:"o" env:"OUTPUT_FORMAT" value:"json" usage:"Output format (json, table, csv)"`
}

// listDevicesResult is used for output formatting
type listDevicesResult struct {
	Devices []audio.Device `json:"devices"`
}

// ToJSON returns the result as JSON bytes
func (r *listDevicesResult) ToJSON() ([]byte, error) {
	return json.MarshalIndent(r, "", "  ")
}

// ToTable returns the result as a table string
func (r *listDevicesResult) ToTable() string {
	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, 1, ' ', 0)
	if _, err := fmt.Fprintln(w, "DEVICE_NAME\tSOURCE\tSINK\tCURRENT_SOURCE\tCURRENT_SINK"); err != nil {
		return ""
	}
	for _, dev := range r.Devices {
		isSource := dev.Mode&audio.DeviceFlagSource != 0
		isSink := dev.Mode&audio.DeviceFlagSink != 0
		isCurrentSource := dev.Mode&audio.DeviceFlagCurrentSource != 0
		isCurrentSink := dev.Mode&audio.DeviceFlagCurrentSink != 0
		if _, err := fmt.Fprintf(w, "%s\t%t\t%t\t%t\t%t\n", dev.Name, isSource, isSink, isCurrentSource, isCurrentSink); err != nil {
			return ""
		}
	}
	if err := w.Flush(); err != nil {
		return ""
	}
	return buf.String()
}

// ToCSV returns the result as CSV string
func (r *listDevicesResult) ToCSV() string {
	var buf bytes.Buffer
	fmt.Fprintln(&buf, "DEVICE_NAME,SOURCE,SINK,CURRENT_SOURCE,CURRENT_SINK")
	for _, dev := range r.Devices {
		isSource := dev.Mode&audio.DeviceFlagSource != 0
		isSink := dev.Mode&audio.DeviceFlagSink != 0
		isCurrentSource := dev.Mode&audio.DeviceFlagCurrentSource != 0
		isCurrentSink := dev.Mode&audio.DeviceFlagCurrentSink != 0
		fmt.Fprintf(&buf, "%s,%t,%t,%t,%t\n", dev.Name, isSource, isSink, isCurrentSource, isCurrentSink)
	}
	return buf.String()
}

// validate validates the list-devices configuration.
func (c *ListDevicesConfig) validate() error {
	switch c.OutputFormat {
	case textOutputFormatJSON, textOutputFormatTable, textOutputFormatCSV:
		return nil
	default:
		return fmt.Errorf("invalid output format: %s (valid values: %s, %s, %s)", c.OutputFormat, textOutputFormatJSON, textOutputFormatTable, textOutputFormatCSV)
	}
}

// runListDevices executes the device listing logic.
func runListDevices(baseConfig *Config, config *ListDevicesConfig) error {
	ctx := context.Background()

	logger := baseConfig.getLogger()
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
	devices, err := audio.ListDevices(ffmpegDevices, spDevices)
	if err != nil {
		return err
	}

	result := &listDevicesResult{Devices: devices}

	output, err := formatOutput(config.OutputFormat, result)
	if err != nil {
		return err
	}

	if baseConfig.Verbose {
		slog.Info("Found devices", "count", len(devices))
	}

	fmt.Print(output)

	return nil
}

func NewListDevicesCommand() *cli.Command {
	var baseConfig Config
	var listDevicesConfig ListDevicesConfig
	baseFlags := flagtags.MustParseFlags(&baseConfig)
	listDevicesFlags := flagtags.MustParseFlags(&listDevicesConfig)
	flags := append(baseFlags, listDevicesFlags...)

	return &cli.Command{
		Name:  "list-devices",
		Usage: "List available audio devices",
		Description: `List available audio devices for recording and playback.

This command enumerates audio devices using both ffmpeg and system-profiler,
merges the information, and outputs devices in lexicographical order with their capabilities.`,
		Flags: flags,
		Action: func(c *cli.Context) error {
			if err := baseConfig.validate(); err != nil {
				return err
			}
			if err := listDevicesConfig.validate(); err != nil {
				return err
			}
			return runListDevices(&baseConfig, &listDevicesConfig)
		},
	}
}
