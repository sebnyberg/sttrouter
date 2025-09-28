package audio

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// Sox handles audio capture using sox
type Sox struct {
}

// NewSox creates a new Sox instance
func NewSox(ctx context.Context) (*Sox, error) {
	return &Sox{}, nil
}

// CaptureAudio captures audio from the specified device to the output writer as raw PCM
func (s *Sox) CaptureAudio(ctx context.Context, log *slog.Logger, device Device, duration time.Duration, output io.Writer) (retErr error) {
	if duration > 0 {
		var cancel context.CancelCauseFunc
		ctx, cancel = context.WithCancelCause(ctx)
		defer func() { cancel(nil) }()
		time.AfterFunc(duration, func() {
			cancel(errors.New("finished duration"))
		})
	}
	args := []string{"-t", "coreaudio", device.Name}

	// Set sample rate if provided
	if device.SampleRate > 0 {
		args = append(args, "-r", fmt.Sprintf("%d", device.SampleRate))
	}

	args = append(args, "-t", "raw", "-e", "signed-integer", "-b", "16", "-c", "2")

	// Use "-" to output to stdout when writing to io.Writer
	args = append(args, "-")

	log.InfoContext(ctx, "Running sox",
		"args", args,
		"command", fmt.Sprintf("sox %s", strings.Join(args, " ")),
		"device", device)

	cmd := exec.Command("sox", args...)
	cmd.Stdout = output

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		log.ErrorContext(ctx, "Sox start failed",
			"error", err,
			"device", device)
		return fmt.Errorf("sox start failed: %w", err)
	}

	// Wait for context cancel or command finish
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case err := <-done:
		if err != nil {
			log.ErrorContext(ctx, "Sox execution failed",
				"error", err,
				"stderr", stderr.String(),
				"device", device)
			return fmt.Errorf("sox capture failed: %s: %w", stderr.String(), ErrAudioCaptureFailed)
		}
	case <-ctx.Done():
		// Send SIGINT for graceful shutdown
		if err := cmd.Process.Signal(syscall.SIGINT); err != nil {
			log.ErrorContext(ctx, "Failed to send SIGINT to sox", "error", err)
			// Fallback to kill
			_ = cmd.Process.Kill()
		}
		// Wait for process to exit
		<-done
	}

	return nil
}

// ConvertAudio converts audio from sourceFormat to targetFormat using sox
func (s *Sox) ConvertAudio(ctx context.Context, log *slog.Logger, reader io.Reader, writer io.Writer, sourceFormat, targetFormat string, sampleRate, channels, bitDepth int) error {
	// Source format
	args := []string{"-t", sourceFormat}
	if sourceFormat == "raw" {
		args = append(args, "-r", strconv.Itoa(sampleRate), "-c", strconv.Itoa(channels), "-b", strconv.Itoa(bitDepth), "-e", "signed-integer")
	}
	args = append(args, "-")

	// Target format
	args = append(args, "-t", targetFormat)
	if targetFormat == "raw" {
		args = append(args, "-r", strconv.Itoa(sampleRate), "-c", strconv.Itoa(channels), "-b", strconv.Itoa(bitDepth), "-e", "signed-integer")
	}
	args = append(args, "-")

	log.InfoContext(ctx, "Running sox convert",
		"args", args,
		"command", fmt.Sprintf("sox %s", strings.Join(args, " ")))

	cmd := exec.CommandContext(ctx, "sox", args...)
	cmd.Stdin = reader
	cmd.Stdout = writer

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		log.ErrorContext(ctx, "Sox convert failed",
			"error", err,
			"stderr", stderr.String())
		return fmt.Errorf("sox convert failed: %s: %w", stderr.String(), ErrAudioCaptureFailed)
	}

	return nil
}
