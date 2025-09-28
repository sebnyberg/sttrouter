package audio

import "sort"

// ListDevices merges device lists from ffmpeg and system-profiler, validates current devices, and sorts by name.
func ListDevices(ffmpegDevices, spDevices []Device) ([]Device, error) {
	// Merge devices from both sources
	deviceModes := make(map[string]uint)
	for _, dev := range ffmpegDevices {
		deviceModes[dev.Name] |= dev.Mode
	}
	for _, dev := range spDevices {
		deviceModes[dev.Name] |= dev.Mode
	}

	// Check for more than one current device per direction
	currentSourceCount := 0
	currentSinkCount := 0
	for _, mode := range deviceModes {
		if mode&DeviceFlagCurrentSource != 0 {
			currentSourceCount++
		}
		if mode&DeviceFlagCurrentSink != 0 {
			currentSinkCount++
		}
	}
	if currentSourceCount > 1 {
		return nil, ErrMultipleCurrentDevices
	}
	if currentSinkCount > 1 {
		return nil, ErrMultipleCurrentDevices
	}

	// Create final device list
	devices := make([]Device, 0, len(deviceModes))
	for name, mode := range deviceModes {
		devices = append(devices, Device{Name: name, Mode: mode})
	}

	// Sort by name
	sort.Slice(devices, func(i, j int) bool {
		return devices[i].Name < devices[j].Name
	})

	return devices, nil
}

// ListSinks returns devices that can act as audio sinks (outputs)
func ListSinks(ffmpegDevices, spDevices []Device) ([]Device, error) {
	devices, err := ListDevices(ffmpegDevices, spDevices)
	if err != nil {
		return nil, err
	}
	var sinks []Device
	for _, dev := range devices {
		if dev.Mode&DeviceFlagSink != 0 {
			sinks = append(sinks, dev)
		}
	}
	return sinks, nil
}

// ListSources returns devices that can act as audio sources (inputs)
func ListSources(ffmpegDevices, spDevices []Device) ([]Device, error) {
	devices, err := ListDevices(ffmpegDevices, spDevices)
	if err != nil {
		return nil, err
	}
	var sources []Device
	for _, dev := range devices {
		if dev.Mode&DeviceFlagSource != 0 {
			sources = append(sources, dev)
		}
	}
	return sources, nil
}

// GetDefaultSink returns the current default sink device
func GetDefaultSink(ffmpegDevices, spDevices []Device) (Device, error) {
	devices, err := ListDevices(ffmpegDevices, spDevices)
	if err != nil {
		return Device{}, err
	}
	for _, dev := range devices {
		if dev.Mode&DeviceFlagCurrentSink != 0 {
			return dev, nil
		}
	}
	return Device{}, ErrNoDefaultDevice
}

// GetDefaultSource returns the current default source device
func GetDefaultSource(ffmpegDevices, spDevices []Device) (Device, error) {
	devices, err := ListDevices(ffmpegDevices, spDevices)
	if err != nil {
		return Device{}, err
	}
	for _, dev := range devices {
		if dev.Mode&DeviceFlagCurrentSource != 0 {
			return dev, nil
		}
	}
	return Device{}, ErrNoDefaultDevice
}
