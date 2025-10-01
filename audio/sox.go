package audio

import (
	"io"
	"time"
)

// CaptureArgs holds the arguments for audio capture
type CaptureArgs struct {
	Device   Device
	Duration time.Duration
	Output   io.Writer
	Channels int
	BitDepth int
}

// ConvertAudioArgs holds the arguments for audio conversion
type ConvertAudioArgs struct {
	Reader       io.Reader
	Writer       io.Writer
	SourceFormat string
	TargetFormat string
	SampleRate   int
	Channels     int
	BitDepth     int
}
