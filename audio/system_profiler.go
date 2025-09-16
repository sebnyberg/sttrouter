package audio

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
)

// SystemProfiler handles device listing using macOS system_profiler
type SystemProfiler struct {
	devices       []string
	defaultDevice string
}

// parseSystemProfilerDevices parses system_profiler JSON output
func parseSystemProfilerDevices(output []byte) ([]string, string, error) {
	var data struct {
		SPAudioDataType []struct {
			Items []struct {
				Name               string `json:"_name"`
				DefaultInputDevice string `json:"coreaudio_default_audio_input_device"`
				DeviceInput        int    `json:"coreaudio_device_input"`
			} `json:"_items"`
		} `json:"SPAudioDataType"`
	}

	if err := json.Unmarshal(output, &data); err != nil {
		return nil, "", fmt.Errorf("parsing: %w", ErrOutputParsingFailed)
	}

	var devices []string
	var defaultDevice string
	for _, audioType := range data.SPAudioDataType {
		for _, item := range audioType.Items {
			if item.DeviceInput > 0 {
				devices = append(devices, item.Name)
			}
			if item.DefaultInputDevice == "spaudio_yes" {
				defaultDevice = item.Name
			}
		}
	}

	return devices, defaultDevice, nil
}

// NewSystemProfiler creates a new SystemProfiler by running system_profiler
func NewSystemProfiler(ctx context.Context) (*SystemProfiler, error) {
	cmd := exec.CommandContext(ctx, "system_profiler", "SPAudioDataType", "-json")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("system_profiler: %w", ErrCommandExecutionFailed)
	}

	devices, defaultDevice, err := parseSystemProfilerDevices(output)
	if err != nil {
		return nil, err
	}

	return &SystemProfiler{devices: devices, defaultDevice: defaultDevice}, nil
}

// ListDeviceNames returns the list of input device names
func (s *SystemProfiler) ListDeviceNames() []string {
	return s.devices
}

// GetDefaultDeviceName returns the name of the default input device
func (s *SystemProfiler) GetDefaultDeviceName() string {
	return s.defaultDevice
}
