package audio

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
