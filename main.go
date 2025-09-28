package main

import (
	"log/slog"
	"os"

	"github.com/sebnyberg/sttrouter/cmd"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "sttrouter",
		Usage: "speech-to-text transcription CLI for developers",
		Description: "Capture audio from microphone and transcribe speech to text using Whisper API.\n" +
			"Supports output to clipboard, stdout, or files for integration into development workflows.",
		Version: "dev",
		Authors: []*cli.Author{
			{
				Name: "Sebastian Nyberg",
			},
		},
		Commands: []*cli.Command{
			cmd.NewListDevicesCommand(),
			cmd.NewCaptureCommand(),
		},
	}

	if err := app.Run(os.Args); err != nil {
		logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{}))
		logger.Error("Application error", "error", err)
		os.Exit(1)
	}
}
