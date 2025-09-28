package cmd

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/sebnyberg/flagtags"
	"github.com/sebnyberg/sttrouter/openaix"
	"github.com/urfave/cli/v2"
)

// TranscribeConfig holds transcribe specific configuration flags.
type TranscribeConfig struct {
	// Model specifies the Whisper model to use
	Model string `name:"model" value:"gpt-4o-transcribe" usage:"Model to use for transcription"`
	// Language specifies the language code (required)
	Language string `name:"language" usage:"Language code (e.g., 'en', 'es')"`
	// ResponseFormat specifies the response format (json, text, srt, verbose_json, vtt)
	ResponseFormat string `name:"response-format" value:"text" usage:"Response format (json,text,srt,verbose_json,vtt)"`
	// Temperature specifies the sampling temperature (0.0 to 1.0)
	Temperature float64 `name:"temperature" value:"0" usage:"Sampling temperature (0.0 to 1.0)"`
	// OpenAI API Key
	APIKey string `name:"api-key"`
	// API Base URL
	BaseURL string `name:"base-url" value:"https://seblab-ai.openai.azure.com/openai/deployments/gpt-4o-transcribe"`
	// Additional query parameters for the API request
	AdditionalQueryParams string `name:"query-params" value:"api-version=2025-03-01-preview" usage:"Additional query parameters for the API request"`
	// EnhanceWithGPT4o enables post-processing with GPT-4o for improved transcription
	EnhanceWithGPT4o bool `name:"enhance" usage:"Enhance transcription with GPT-4o for better accuracy"`
}

// validate validates the TranscribeConfig and returns an error if required fields are missing.
func (c *TranscribeConfig) validate() error {
	if c.APIKey == "" {
		return fmt.Errorf("API key is required (use --api-key or set API_KEY environment variable)")
	}
	if c.Language == "" {
		return fmt.Errorf("language is required (use --language or set LANGUAGE environment variable)")
	}
	return nil
}

// runTranscribe executes the audio transcription logic.
func runTranscribe(baseConfig *Config, config *TranscribeConfig, audioFile string) error {
	ctx := context.Background()
	_ = ctx

	logger := baseConfig.getLogger()
	slog.SetDefault(logger)

	client := openaix.NewClient(config.APIKey, config.BaseURL, config.AdditionalQueryParams)

	// Prepare transcription request
	req := openaix.TranscriptionRequest{
		File:           audioFile,
		Model:          config.Model,
		Language:       config.Language,
		ResponseFormat: config.ResponseFormat,
		Temperature:    config.Temperature,
	}

	t, err := client.Transcribe(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to transcribe audio: %w", err)
	}

	transcription := t.Text
	fmt.Println(transcription)

	if baseConfig.Verbose {
		slog.Info("Transcription completed successfully")
	}

	return nil
}

func NewTranscribeCommand() *cli.Command {
	var baseConfig Config
	var transcribeConfig TranscribeConfig
	baseFlags := flagtags.MustParseFlags(&baseConfig)
	transcribeFlags := flagtags.MustParseFlags(&transcribeConfig)
	flags := append(baseFlags, transcribeFlags...)

	return &cli.Command{
		Name:      "transcribe",
		Usage:     "Transcribe audio file to text using OpenAI Whisper API",
		ArgsUsage: "AUDIO_FILE",
		Description: `Transcribe an audio file to text using OpenAI's Whisper API.

The AUDIO_FILE is a required positional argument that specifies the audio file to transcribe.
Supported formats include FLAC, MP3, MP4, MPEG, MPGA, M4A, OGG, WAV, and WEBM.

The transcribed text is output to stdout by default.

Examples:
  # Transcribe a FLAC file
  sttrouter transcribe --api-key YOUR_KEY --language en recording.flac

  # Transcribe with GPT-4o enhancement
  sttrouter transcribe --api-key YOUR_KEY --language en --enhance recording.flac

  # Transcribe to JSON format
  sttrouter transcribe --api-key YOUR_KEY --language en --response-format json recording.flac`,
		Flags: flags,
		Action: func(c *cli.Context) error {
			if c.NArg() != 1 {
				return fmt.Errorf("exactly one argument (AUDIO_FILE) is required")
			}

			audioFile := c.Args().Get(0)

			if err := baseConfig.validate(); err != nil {
				return err
			}

			if err := transcribeConfig.validate(); err != nil {
				return err
			}

			return runTranscribe(&baseConfig, &transcribeConfig, audioFile)
		},
	}
}
