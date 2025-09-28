package audio

import (
	"encoding/json"
)

// DeviceFlag represents device capabilities and defaults as bit flags
const (
	DeviceFlagUnset         = 0
	DeviceFlagSource        = 1 << 0
	DeviceFlagSink          = 1 << 1
	DeviceFlagCurrentSource = 1 << 2
	DeviceFlagCurrentSink   = 1 << 3
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
		"source":         d.Mode&DeviceFlagSource != 0,
		"sink":           d.Mode&DeviceFlagSink != 0,
		"current_source": d.Mode&DeviceFlagCurrentSource != 0,
		"current_sink":   d.Mode&DeviceFlagCurrentSink != 0,
	})
}

// UnmarshalJSON custom unmarshals JSON to Device
func (d *Device) UnmarshalJSON(data []byte) error {
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}
	d.Name = m["name"].(string)
	if source, ok := m["source"].(bool); ok && source {
		d.Mode |= DeviceFlagSource
	}
	if sink, ok := m["sink"].(bool); ok && sink {
		d.Mode |= DeviceFlagSink
	}
	if currentSource, ok := m["current_source"].(bool); ok && currentSource {
		d.Mode |= DeviceFlagCurrentSource
	}
	if currentSink, ok := m["current_sink"].(bool); ok && currentSink {
		d.Mode |= DeviceFlagCurrentSink
	}
	return nil
}
