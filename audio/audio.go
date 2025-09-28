package audio

import (
	"encoding/json"
	"sort"
)

// DeviceFlag represents device capabilities and defaults as bit flags
const (
	DeviceFlagUnset         = 0
	DeviceFlagInput         = 1 << 0
	DeviceFlagOutput        = 1 << 1
	DeviceFlagCurrentInput  = 1 << 2
	DeviceFlagCurrentOutput = 1 << 3
)

// Device represents an audio device with its name and mode flags
type Device struct {
	Name string
	Mode uint
}

// MarshalJSON custom marshals Device to JSON with readable flags
func (d Device) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"name":           d.Name,
		"input":          d.Mode&DeviceFlagInput != 0,
		"output":         d.Mode&DeviceFlagOutput != 0,
		"current_input":  d.Mode&DeviceFlagCurrentInput != 0,
		"current_output": d.Mode&DeviceFlagCurrentOutput != 0,
	})
}

// UnmarshalJSON custom unmarshals JSON to Device
func (d *Device) UnmarshalJSON(data []byte) error {
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}
	d.Name = m["name"].(string)
	if input, ok := m["input"].(bool); ok && input {
		d.Mode |= DeviceFlagInput
	}
	if output, ok := m["output"].(bool); ok && output {
		d.Mode |= DeviceFlagOutput
	}
	if currentInput, ok := m["current_input"].(bool); ok && currentInput {
		d.Mode |= DeviceFlagCurrentInput
	}
	if currentOutput, ok := m["current_output"].(bool); ok && currentOutput {
		d.Mode |= DeviceFlagCurrentOutput
	}
	return nil
}

// GetDevices merges device lists from ffmpeg and system-profiler and sorts by name.
func GetDevices(ffmpegDevices, spDevices []Device) ([]Device, error) {
	// Merge devices from both sources
	deviceModes := make(map[string]uint)
	for _, dev := range ffmpegDevices {
		deviceModes[dev.Name] |= dev.Mode
	}
	for _, dev := range spDevices {
		deviceModes[dev.Name] |= dev.Mode
	}

	// Create final device list
	var devices []Device
	for name, mode := range deviceModes {
		devices = append(devices, Device{Name: name, Mode: mode})
	}

	// Sort by name
	sort.Slice(devices, func(i, j int) bool {
		return devices[i].Name < devices[j].Name
	})

	return devices, nil
}
