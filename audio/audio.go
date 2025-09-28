package audio

import "sort"

// ListDevices merges device lists from ffmpeg and system-profiler, validates current devices, and sorts by name.
func ListDevices(ffmpegDevices, spDevices []Device) ([]Device, error) {
	// Merge devices from both sources
	deviceModes := make(map[string]uint)
	deviceSampleRates := make(map[string]int)
	deviceIndices := make(map[string]string)

	for _, dev := range ffmpegDevices {
		deviceModes[dev.Name] |= dev.Mode
		deviceIndices[dev.Name] = dev.Index // Store the ffmpeg device index
	}

	for _, dev := range spDevices {
		deviceModes[dev.Name] |= dev.Mode
		deviceSampleRates[dev.Name] = dev.SampleRate // Safe: ffmpeg devices have no sample rate info, only spDevices provide it
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
		devices = append(devices, Device{
			Name:       name,
			Mode:       mode,
			SampleRate: deviceSampleRates[name], // Safe: deviceSampleRates[name] set from spDevices (non-zero) or 0 for ffmpeg-only devices
			Index:      deviceIndices[name],     // Include the ffmpeg device index
		})
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

// GetDevice returns the device with the specified name
func GetDevice(name string, ffmpegDevices, spDevices []Device) (Device, error) {
	devices, err := ListDevices(ffmpegDevices, spDevices)
	if err != nil {
		return Device{}, err
	}
	for _, dev := range devices {
		if dev.Name == name {
			return dev, nil
		}
	}
	return Device{}, ErrNoDefaultDevice
}
