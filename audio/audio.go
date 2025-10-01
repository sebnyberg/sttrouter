package audio

import (
	"context"
	"io"
	"log/slog"
	"time"

	"golang.org/x/sync/errgroup"
)

func ListDevices() ([]Device, error) {
	lister, err := NewDeviceLister(context.Background())
	if err != nil {
		return nil, err
	}

	return lister.ListDevices(), nil
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
	EnableAutoStop      bool
	AutoStopThreshold   float64
	AutoStopMinDuration time.Duration
	Duration            time.Duration
	Channels            int
	BitDepth            int
	Writer              io.WriteCloser
}

// LimitedCapture captures raw audio from the device until auto-stop is detected or duration expires
func LimitedCapture(
	ctx context.Context,
	logger *slog.Logger,
	device Device,
	args LimitedCaptureArgs,
) error {
	defer func() { _ = args.Writer.Close() }()

	// Set up capture I/O
	g := new(errgroup.Group)
	captureReader, captureWriter := io.Pipe()

	captureCtx, captureCancel := context.WithCancel(ctx)
	defer captureCancel()
	if args.EnableAutoStop {
		// Set up auto-stop splitter
		autoStopReader, autoStopWriter := io.Pipe()
		splitter := NewSilenceSplitter(
			ctx,
			args.Channels,
			args.BitDepth,
			args.AutoStopThreshold,
			args.AutoStopMinDuration,
			device.SampleRate,
			func(data []byte) {
				_, _ = autoStopWriter.Write(data)
				captureCancel()
			},
		)

		g.Go(func() error {
			defer func() { _ = autoStopWriter.Close() }()
			_, err := io.Copy(args.Writer, autoStopReader)
			return err
		})

		g.Go(func() error {
			defer func() { _ = autoStopWriter.Close() }()
			_, err := io.Copy(splitter, captureReader)
			splitter.Flush() // Flush any remaining data
			return err
		})
	} else {
		g.Go(func() error {
			_, err := io.Copy(args.Writer, captureReader)
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
		return err
	}

	if err := g.Wait(); err != nil {
		return err
	}

	return nil
}
