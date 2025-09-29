package audio

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"time"

	"golang.org/x/sync/errgroup"
)

func ListDevices() ([]Device, error) {
	sp, err := NewSystemProfiler(context.Background())
	if err != nil {
		return nil, err
	}

	return sp.ListDevices(), nil
}

func ListSinks() ([]Device, error) {
	devices, err := ListDevices()
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
func ListSources() ([]Device, error) {
	devices, err := ListDevices()
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
func GetDefaultSink() (Device, error) {
	devices, err := ListDevices()
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
func GetDefaultSource(spDevices []Device) (Device, error) {
	devices, err := ListDevices()
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
func GetDevice(name string, spDevices []Device) (Device, error) {
	devices, err := ListDevices()
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

// LimitedCaptureArgs holds the arguments for limited capture
type LimitedCaptureArgs struct {
	EnableSilence      bool
	SilenceThreshold   float64
	SilenceMinDuration time.Duration
	Duration           time.Duration
	Channels           int
	BitDepth           int
}

// LimitedCapture captures raw audio from the device until silence is detected or duration expires
func LimitedCapture(
	ctx context.Context,
	logger *slog.Logger,
	device Device,
	args LimitedCaptureArgs,
) (io.Reader, error) {
	// Buffer to collect raw audio
	buffer := &bytes.Buffer{}

	// Set up capture I/O
	g := new(errgroup.Group)
	captureReader, captureWriter := io.Pipe()

	captureCtx, captureCancel := context.WithCancel(ctx)
	defer captureCancel()
	if args.EnableSilence {
		// Set up silence splitter
		silenceReader, silenceWriter := io.Pipe()
		splitter := NewSilenceSplitter(
			ctx,
			args.Channels,
			args.BitDepth,
			args.SilenceThreshold,
			args.SilenceMinDuration,
			device.SampleRate,
			func(data []byte) {
				_, _ = silenceWriter.Write(data)
				captureCancel()
			},
		)

		g.Go(func() error {
			defer func() { _ = silenceWriter.Close() }()
			_, err := io.Copy(buffer, silenceReader)
			return err
		})

		g.Go(func() error {
			defer func() { _ = silenceWriter.Close() }()
			_, err := io.Copy(splitter, captureReader)
			splitter.Flush() // Flush any remaining data
			return err
		})
	} else {
		g.Go(func() error {
			_, err := io.Copy(buffer, captureReader)
			return err
		})
	}

	// Capture audio until duration is finished.
	logger.Info("audio capture started")
	err := CaptureAudio(captureCtx, logger, CaptureArgs{
		Device:   device,
		Duration: args.Duration,
		Output:   captureWriter,
		Channels: args.Channels,
		BitDepth: args.BitDepth,
	})
	_ = captureWriter.Close()
	if err != nil {
		return nil, err
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return buffer, nil
}
