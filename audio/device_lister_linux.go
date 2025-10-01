package audio

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// DeviceLister handles device listing using pactl on Linux
type DeviceLister struct {
	devices []Device
}

// parsePactlDevices parses pactl list sources/sinks output
func parsePactlDevices(sourcesOutput, sinksOutput, defaultSource, defaultSink []byte) ([]Device, error) {
	var devices []Device
	index := 0

	defaultSourceStr := strings.TrimSpace(string(defaultSource))
	defaultSinkStr := strings.TrimSpace(string(defaultSink))

	// Parse sources (input devices)
	sourceRegex := regexp.MustCompile(`(?m)^Source #(\d+)\n(?:.*\n)*?\s+Name: (.*?)\n(?:.*\n)*?\s+Description: (.*?)\n(?:.*\n)*?\s+Sample Specification: s(\d+)[a-z]* (\d+)ch (\d+)Hz`)
	sourceMatches := sourceRegex.FindAllSubmatch(sourcesOutput, -1)

	for _, match := range sourceMatches {
		if len(match) < 7 {
			continue
		}

		deviceName := strings.TrimSpace(string(match[2]))
		sampleRate, _ := strconv.Atoi(string(match[6]))

		mode := uint(DeviceFlagSource)

		// Check if it's the default source
		if deviceName == defaultSourceStr {
			mode |= DeviceFlagCurrentSource
		}

		devices = append(devices, Device{
			Name:       deviceName, // Use device name (not description) for sox
			Mode:       mode,
			SampleRate: sampleRate,
			Index:      index,
		})
		index++
	}

	// Parse sinks (output devices)
	sinkRegex := regexp.MustCompile(`(?m)^Sink #(\d+)\n(?:.*\n)*?\s+Name: (.*?)\n(?:.*\n)*?\s+Description: (.*?)\n(?:.*\n)*?\s+Sample Specification: s(\d+)[a-z]* (\d+)ch (\d+)Hz`)
	sinkMatches := sinkRegex.FindAllSubmatch(sinksOutput, -1)

	for _, match := range sinkMatches {
		if len(match) < 7 {
			continue
		}

		deviceName := strings.TrimSpace(string(match[2]))
		sampleRate, _ := strconv.Atoi(string(match[6]))

		mode := uint(DeviceFlagSink)

		// Check if it's the default sink
		if deviceName == defaultSinkStr {
			mode |= DeviceFlagCurrentSink
		}

		devices = append(devices, Device{
			Name:       deviceName, // Use device name (not description) for sox
			Mode:       mode,
			SampleRate: sampleRate,
			Index:      index,
		})
		index++
	}

	return devices, nil
}

// NewDeviceLister creates a new DeviceLister by running pactl
func NewDeviceLister(ctx context.Context) (*DeviceLister, error) {
	// Get sources
	sourcesCmd := exec.CommandContext(ctx, "pactl", "list", "sources")
	sourcesOutput, err := sourcesCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("pactl list sources: %w", ErrCommandExecutionFailed)
	}

	// Get sinks
	sinksCmd := exec.CommandContext(ctx, "pactl", "list", "sinks")
	sinksOutput, err := sinksCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("pactl list sinks: %w", ErrCommandExecutionFailed)
	}

	// Get default source
	defaultSourceCmd := exec.CommandContext(ctx, "pactl", "get-default-source")
	defaultSource, err := defaultSourceCmd.Output()
	if err != nil {
		defaultSource = []byte("")
	}

	// Get default sink
	defaultSinkCmd := exec.CommandContext(ctx, "pactl", "get-default-sink")
	defaultSink, err := defaultSinkCmd.Output()
	if err != nil {
		defaultSink = []byte("")
	}

	devices, err := parsePactlDevices(sourcesOutput, sinksOutput, defaultSource, defaultSink)
	if err != nil {
		return nil, err
	}

	return &DeviceLister{devices: devices}, nil
}

// ListDevices returns the list of devices
func (s *DeviceLister) ListDevices() []Device {
	return s.devices
}

// GetDefaultDeviceName returns the name of the current source device
func (s *DeviceLister) GetDefaultDeviceName() string {
	for _, device := range s.devices {
		if device.Mode&DeviceFlagCurrentSource != 0 {
			return device.Name
		}
	}
	return ""
}
