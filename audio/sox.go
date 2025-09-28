package audio

import (
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"strings"
)

// Sox handles audio capture using sox
type Sox struct {
}

// NewSox creates a new Sox instance
func NewSox(ctx context.Context) (*Sox, error) {
	return &Sox{}, nil
}

// CaptureAudio captures audio from the specified device to the output file or stdout
func (s *Sox) CaptureAudio(ctx context.Context, log *slog.Logger, device Device, duration, container, output string) error {
	args := []string{"-t", "coreaudio", device.Name}

	if container != "" {
		args = append(args, "-t", container)
	} else {
		args = append(args, "-t", "wav") // default to wav
	}

	// Set sample rate if provided
	if device.SampleRate > 0 {
		args = append(args, "-r", fmt.Sprintf("%d", device.SampleRate))
	}

	args = append(args, output)

	if duration != "" {
		args = append(args, "trim", "0", "10")
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
