package audio

import (
	"context"
	"fmt"
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
				name := strings.TrimSpace(matches[2])
				devices = append(devices, Device{Name: name, Mode: 0})
			}
		}
	}

	return devices
}
