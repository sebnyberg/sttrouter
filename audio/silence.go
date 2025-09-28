package audio

import (
	"bytes"
	"context"
	"math"
	"time"
)

// SilenceSplitter implements io.Writer and splits the audio stream into segments based on silence detection
type SilenceSplitter struct {
	ctx              context.Context
	channels         int
	bitDepth         int
	threshold        int
	minSilentSamples int // silent samples prior to flushing
	buffer           *bytes.Buffer
	silentCount      int
	callback         func([]byte)
}

// NewSilenceSplitter creates a new SilenceSplitter
func NewSilenceSplitter(ctx context.Context, channels, bitDepth int, threshold float64, minDuration time.Duration, sampleRate int, callback func([]byte)) *SilenceSplitter {
	maxAmp := 1 << uint(bitDepth-1)
	thresh := int(threshold * float64(maxAmp))
	minSamples := int(minDuration.Seconds() * float64(sampleRate) * float64(channels))
	return &SilenceSplitter{
		ctx:              ctx,
		channels:         channels,
		bitDepth:         bitDepth,
		threshold:        thresh,
		minSilentSamples: minSamples,
		buffer:           &bytes.Buffer{},
		silentCount:      0,
		callback:         callback,
	}
}

// Write implements io.Writer
func (s *SilenceSplitter) Write(p []byte) (n int, err error) {
	s.buffer.Write(p)

	// Process the new data for silence detection
	bytesPerSample := s.bitDepth / 8
	samples := len(p) / bytesPerSample / s.channels

	allSilent := true
	for i := 0; i < samples; i++ {
		offset := i * bytesPerSample * s.channels
		for ch := 0; ch < s.channels; ch++ {
			sampleOffset := offset + ch*bytesPerSample
			sampleBytes := p[sampleOffset : sampleOffset+bytesPerSample]
			var val int
			if s.bitDepth == 16 {
				val = int(int16(sampleBytes[0]) | int16(sampleBytes[1])<<8)
			}
			if math.Abs(float64(val)) >= float64(s.threshold) {
				allSilent = false
				break
			}
		}
		if !allSilent {
			break
		}
	}

	if allSilent {
		s.silentCount += samples
		if s.silentCount >= s.minSilentSamples && s.buffer.Len() > 0 {
			s.flush()
		}
	} else {
		s.silentCount = 0
	}

	return len(p), nil
}

// flush calls the callback with the buffered data
func (s *SilenceSplitter) flush() {
	if s.buffer.Len() == 0 {
		return
	}
	s.callback(s.buffer.Bytes())
	s.buffer.Reset()
	s.silentCount = 0
}

// Flush forces flushing of any remaining buffered data
func (s *SilenceSplitter) Flush() {
	s.flush()
}
