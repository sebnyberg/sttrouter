package audio

import (
	"context"
	"fmt"
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

// CaptureAudio captures audio from the specified device to the output file or stdout
func (s *Sox) CaptureAudio(ctx context.Context, log *slog.Logger, device Device, duration time.Duration, targetFormat, output string) error {
	args := []string{"-t", "coreaudio", device.Name}

	// Set sample rate if provided
	if device.SampleRate > 0 {
		args = append(args, "-r", fmt.Sprintf("%d", device.SampleRate))
	}

	args = append(args, "-t", targetFormat)

	args = append(args, output)

	if duration > 0 {
		args = append(args, "trim", "0", fmt.Sprintf("%.1f", duration.Seconds()))
	}

	log.InfoContext(ctx, "Running sox",
		"args", args,
		"command", fmt.Sprintf("sox %s", strings.Join(args, " ")),
		"device", device)

	cmd := exec.CommandContext(ctx, "sox", args...)
	outputBytes, err := cmd.CombinedOutput()
	if err != nil {
		log.ErrorContext(ctx, "Sox execution failed",
			"error", err,
			"output", string(outputBytes),
			"device", device)
		return fmt.Errorf("sox capture failed: %s: %w", string(outputBytes), ErrAudioCaptureFailed)
	}

	return nil
}
