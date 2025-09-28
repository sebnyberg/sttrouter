package audio

import (
	"sort"
)

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
