package audio

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
)

// SystemProfiler handles device listing using macOS system_profiler
type SystemProfiler struct {
	devices []Device
}

// parseSystemProfilerDevices parses system_profiler JSON output
func parseSystemProfilerDevices(output []byte) ([]Device, error) {
	var data struct {
		SPAudioDataType []struct {
			Items []struct {
				Name                string `json:"_name"`
				DefaultInputDevice  string `json:"coreaudio_default_audio_input_device"`
				DefaultOutputDevice string `json:"coreaudio_default_audio_output_device"`
				DeviceInput         int    `json:"coreaudio_device_input"`
				DeviceOutput        int    `json:"coreaudio_device_output"`
			} `json:"_items"`
		} `json:"SPAudioDataType"`
	}

	if err := json.Unmarshal(output, &data); err != nil {
		return nil, fmt.Errorf("parsing: %w", ErrOutputParsingFailed)
	}

	var devices []Device
	for _, audioType := range data.SPAudioDataType {
		for _, item := range audioType.Items {
			mode := uint(0)
			if item.DeviceInput > 0 {
				mode |= DeviceFlagInput
			}
			if item.DeviceOutput > 0 {
				mode |= DeviceFlagOutput
			}
			if item.DefaultInputDevice == "spaudio_yes" {
				mode |= DeviceFlagCurrentInput
			}
			if item.DefaultOutputDevice == "spaudio_yes" {
				mode |= DeviceFlagCurrentOutput
			}
			devices = append(devices, Device{Name: item.Name, Mode: mode})
		}
	}

	return devices, nil
}

// NewSystemProfiler creates a new SystemProfiler by running system_profiler
func NewSystemProfiler(ctx context.Context) (*SystemProfiler, error) {
	cmd := exec.CommandContext(ctx, "system_profiler", "SPAudioDataType", "-json")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("system_profiler: %w", ErrCommandExecutionFailed)
	}

	devices, err := parseSystemProfilerDevices(output)
	if err != nil {
		return nil, err
	}

	return &SystemProfiler{devices: devices}, nil
}

// ListDevices returns the list of devices
func (s *SystemProfiler) ListDevices() []Device {
	return s.devices
}

// GetDefaultDeviceName returns the name of the current input device
func (s *SystemProfiler) GetDefaultDeviceName() string {
	for _, device := range s.devices {
		if device.Mode&DeviceFlagCurrentInput != 0 {
			return device.Name
		}
	}
	return ""
}
