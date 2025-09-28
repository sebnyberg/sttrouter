package audio

import (
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"regexp"
	"strings"
)

// FFmpeg handles device listing using ffmpeg
type FFmpeg struct {
	output string
}

// NewFFmpeg creates a new FFmpeg by running ffmpeg list_devices
func NewFFmpeg(ctx context.Context) (*FFmpeg, error) {
	cmd := exec.CommandContext(ctx, "ffmpeg", "-f", "avfoundation", "-list_devices", "true", "-i", "")
	output, err := cmd.CombinedOutput()
	if err != nil && !strings.Contains(string(output), "AVFoundation") {
		return nil, fmt.Errorf("ffmpeg: %w", ErrCommandExecutionFailed)
	}

	return &FFmpeg{output: string(output)}, nil
}

// ListDevices returns the list of input devices
func (f *FFmpeg) ListDevices() []Device {
	return parseFFmpegDevices(f.output)
}

// CaptureAudio captures audio from the specified device to the output file or stdout
func (f *FFmpeg) CaptureAudio(ctx context.Context, log *slog.Logger, device Device, duration, container, codec, output string) error {
	// Use device.Index if available, otherwise try with device.Name
	deviceIdentifier := device.Name
	if device.Index != "" {
		// For audio-only capture in avfoundation, we need to prefix the index with a colon
		deviceIdentifier = ":" + device.Index
	}

	args := []string{"-f", "avfoundation", "-i", deviceIdentifier}

	if duration != "" {
		args = append(args, "-t", duration)
	}

	if codec != "" {
		args = append(args, "-acodec", codec)
	}

	// Set sample rate if provided
	if device.SampleRate > 0 {
		args = append(args, "-ar", fmt.Sprintf("%d", device.SampleRate))
	}

	if container != "" {
		args = append(args, "-f", container)
	}

	args = append(args, output)

	log.InfoContext(ctx, "Running ffmpeg",
		"args", args,
		"command", fmt.Sprintf("ffmpeg %s", strings.Join(args, " ")),
		"device", device)

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	outputBytes, err := cmd.CombinedOutput()
	if err != nil {
		log.ErrorContext(ctx, "FFmpeg execution failed",
			"error", err,
			"output", string(outputBytes),
			"device", device)
		return fmt.Errorf("ffmpeg capture failed: %s: %w", string(outputBytes), ErrAudioCaptureFailed)
	}

	return nil
}

// parseFFmpegDevices parses ffmpeg output for devices
func parseFFmpegDevices(output string) []Device {
	var devices []Device
	lines := strings.Split(output, "\n")

	inAudioSection := false
	// Regex to match: [AVFoundation indev @ 0x...] [0] Device Name
	re := regexp.MustCompile(`\[AVFoundation indev @ [^\]]+\]\s*\[(\d+)\]\s*(.+)`)
	for _, line := range lines {
		if strings.Contains(line, "AVFoundation audio devices:") {
			inAudioSection = true
			continue
		}
		if inAudioSection {
			if strings.Contains(line, "Error") || strings.Contains(line, "Error opening") {
				break
			}
			matches := re.FindStringSubmatch(line)
			if len(matches) == 3 {
				index := matches[1]
				name := strings.TrimSpace(matches[2])
				devices = append(devices, Device{Name: name, Mode: 0, Index: index})
			}
		}
	}

	return devices
}
