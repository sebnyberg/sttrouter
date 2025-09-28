package audio

import (
	"encoding/json"
	"fmt"
	"sort"
)

// DeviceFlag represents device capabilities and defaults as bit flags
const (
	DeviceFlagUnset     = 0
	DeviceFlagInput     = 1 << 0
	DeviceFlagOutput    = 1 << 1
	DeviceFlagIsDefault = 1 << 2
)

// Device represents an audio device with its name and mode flags
type Device struct {
	Name string
	Mode uint
}

// MarshalJSON custom marshals Device to JSON with readable flags
func (d Device) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"name":    d.Name,
		"input":   d.Mode&DeviceFlagInput != 0,
		"output":  d.Mode&DeviceFlagOutput != 0,
		"default": d.Mode&DeviceFlagIsDefault != 0,
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
	if def, ok := m["default"].(bool); ok && def {
		d.Mode |= DeviceFlagIsDefault
	}
	return nil
}

// GetDevices merges device lists from ffmpeg and system-profiler, validates defaults, and sorts by name.
func GetDevices(ffmpegDevices, spDevices []Device) ([]Device, error) {
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
		if mode&DeviceFlagIsDefault != 0 {
			defaultCount++
		}
	}
	if defaultCount > 1 {
		return nil, fmt.Errorf("more than one default device found")
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
