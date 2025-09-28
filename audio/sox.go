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

// CaptureArgs holds the arguments for audio capture
type CaptureArgs struct {
	Device   Device
	Duration time.Duration
	Output   io.Writer
	Channels int
	BitDepth int
}

// CaptureAudio captures audio from the specified device to the output writer as raw PCM
func CaptureAudio(ctx context.Context, log *slog.Logger, args CaptureArgs) (retErr error) {
	if args.Duration > 0 {
		var cancel context.CancelCauseFunc
		ctx, cancel = context.WithCancelCause(ctx)
		defer func() { cancel(nil) }()
		time.AfterFunc(args.Duration, func() {
			cancel(errors.New("finished duration"))
		})
	}
	cmdArgs := []string{"-t", "coreaudio", args.Device.Name}

	// Set sample rate if provided
	if args.Device.SampleRate > 0 {
		cmdArgs = append(cmdArgs, "-r", fmt.Sprintf("%d", args.Device.SampleRate))
	}

	cmdArgs = append(
		cmdArgs,
		"-t", "raw",
		"-e", "signed-integer",
		"-b", strconv.Itoa(args.BitDepth),
		"-c", strconv.Itoa(args.Channels),
	)

	// Use "-" to output to stdout when writing to io.Writer
	cmdArgs = append(cmdArgs, "-")

	log.DebugContext(ctx, "Running sox",
		"args", cmdArgs,
		"command", fmt.Sprintf("sox %s", strings.Join(cmdArgs, " ")),
		"device", args.Device)

	cmd := exec.Command("sox", cmdArgs...)
	cmd.Stdout = args.Output

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		log.ErrorContext(ctx, "Sox start failed",
			"error", err,
			"device", args.Device)
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
				"device", args.Device)
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

// ConvertAudio converts audio from sourceFormat to targetFormat using sox
func ConvertAudio(ctx context.Context, log *slog.Logger, args ConvertAudioArgs) error {
	// Source format
	cmdArgs := []string{"-t", args.SourceFormat}
	if args.SourceFormat == "raw" {
		cmdArgs = append(
			cmdArgs,
			"-r", strconv.Itoa(args.SampleRate),
			"-c", strconv.Itoa(args.Channels),
			"-b", strconv.Itoa(args.BitDepth),
			"-e", "signed-integer",
		)
	}
	cmdArgs = append(cmdArgs, "-")

	// Target format
	cmdArgs = append(cmdArgs, "-t", args.TargetFormat)
	if args.TargetFormat == "raw" {
		cmdArgs = append(
			cmdArgs,
			"-r", strconv.Itoa(args.SampleRate),
			"-c", strconv.Itoa(args.Channels),
			"-b", strconv.Itoa(args.BitDepth),
			"-e", "signed-integer",
		)
	}
	cmdArgs = append(cmdArgs, "-")

	log.DebugContext(ctx, "Running sox convert",
		"args", cmdArgs,
		"command", fmt.Sprintf("sox %s", strings.Join(cmdArgs, " ")))

	cmd := exec.CommandContext(ctx, "sox", cmdArgs...)
	cmd.Stdin = args.Reader
	cmd.Stdout = args.Writer

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
