package audio

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"time"
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
	Name       string
	Mode       uint
	SampleRate int
	Index      int // Index used by sox for coreaudio
}

// MarshalJSON custom marshals Device to JSON with readable flags
func (d Device) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"name":           d.Name,
		"source":         d.Mode&DeviceFlagSource != 0,
		"sink":           d.Mode&DeviceFlagSink != 0,
		"current_source": d.Mode&DeviceFlagCurrentSource != 0,
		"current_sink":   d.Mode&DeviceFlagCurrentSink != 0,
		"sample_rate":    d.SampleRate,
		"index":          d.Index,
	})
}

// CaptureAudio captures audio from this device using the provided Sox instance
func (d Device) CaptureAudio(sox *Sox, ctx context.Context, log *slog.Logger, duration time.Duration, output io.Writer) error {
	return sox.CaptureAudio(ctx, log, d, duration, output)
}
