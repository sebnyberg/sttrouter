package audio

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
)

// DeviceLister handles device listing using macOS system_profiler
type DeviceLister struct {
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
				SampleRate          int    `json:"coreaudio_device_srate"`
			} `json:"_items"`
		} `json:"SPAudioDataType"`
	}

	if err := json.Unmarshal(output, &data); err != nil {
		return nil, fmt.Errorf("parsing: %w", ErrOutputParsingFailed)
	}

	var devices []Device
	index := 0
	for _, audioType := range data.SPAudioDataType {
		for _, item := range audioType.Items {
			mode := uint(0)
			if item.DeviceInput > 0 {
				mode |= DeviceFlagSource
			}
			if item.DeviceOutput > 0 {
				mode |= DeviceFlagSink
			}
			if item.DefaultInputDevice == "spaudio_yes" {
				mode |= DeviceFlagCurrentSource
			}
			if item.DefaultOutputDevice == "spaudio_yes" {
				mode |= DeviceFlagCurrentSink
			}
			devices = append(devices, Device{Name: item.Name, Mode: mode, SampleRate: item.SampleRate, Index: index})
			index++
		}
	}

	return devices, nil
}

// NewDeviceLister creates a new DeviceLister by running system_profiler
func NewDeviceLister(ctx context.Context) (*DeviceLister, error) {
	cmd := exec.CommandContext(ctx, "system_profiler", "SPAudioDataType", "-json")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("system_profiler: %w", ErrCommandExecutionFailed)
	}

	devices, err := parseSystemProfilerDevices(output)
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
