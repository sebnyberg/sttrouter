package audio

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os/exec"
	"strings"
	"time"
)

// Sox handles audio capture using sox
type Sox struct {
}

// NewSox creates a new Sox instance
func NewSox(ctx context.Context) (*Sox, error) {
	return &Sox{}, nil
}

// CaptureAudio captures audio from the specified device to the output writer
func (s *Sox) CaptureAudio(ctx context.Context, log *slog.Logger, device Device, duration time.Duration, targetFormat string, output io.Writer) error {
	args := []string{"-t", "coreaudio", device.Name}

	// Set sample rate if provided
	if device.SampleRate > 0 {
		args = append(args, "-r", fmt.Sprintf("%d", device.SampleRate))
	}

	args = append(args, "-t", targetFormat)

	// Use "-" to output to stdout when writing to io.Writer
	args = append(args, "-")

	if duration > 0 {
		args = append(args, "trim", "0", fmt.Sprintf("%.1f", duration.Seconds()))
	}

	log.InfoContext(ctx, "Running sox",
		"args", args,
		"command", fmt.Sprintf("sox %s", strings.Join(args, " ")),
		"device", device)

	cmd := exec.CommandContext(ctx, "sox", args...)
	cmd.Stdout = output

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		log.ErrorContext(ctx, "Sox execution failed",
			"error", err,
			"stderr", stderr.String(),
			"device", device)
		return fmt.Errorf("sox capture failed: %s: %w", stderr.String(), ErrAudioCaptureFailed)
	}

	return nil
}
