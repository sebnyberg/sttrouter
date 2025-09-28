package main

import (
	"fmt"
	"os"

	"github.com/sebnyberg/sttrouter/cmd"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:        "sttrouter",
		Usage:       "speech-to-text transcription CLI for developers",
		Description: "Capture audio from microphone and transcribe speech to text using Whisper API.\nSupports output to clipboard, stdout, or files for integration into development workflows.",
		Version:     "dev",
		Authors: []*cli.Author{
			{
				Name: "Sebastian Nyberg",
			},
		},
		Commands: []*cli.Command{
			cmd.NewListDevicesCommand(),
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
