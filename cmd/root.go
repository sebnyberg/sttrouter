package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

// Config holds the global application configuration.
// This structure enables clean separation between CLI parsing and business logic.
type Config struct {
	// LogLevel defines the minimum log level to output (debug, info, warn, error)
	LogLevel string `flag:"log-level" env:"LOG_LEVEL" default:"info"`

	// LogFormat specifies the log output format (text, json)
	LogFormat string `flag:"log-format" env:"LOG_FORMAT" default:"text"`

	// Verbose enables more detailed logging and output
	Verbose bool `flag:"verbose" env:"VERBOSE" default:"false"`
}

// Validate validates the configuration and sets defaults.
func (c *Config) Validate() error {
	// Validate log level
	switch c.LogLevel {
	case "debug", "info", "warn", "error":
		// Valid log levels
	default:
		return fmt.Errorf("invalid log level: %s (valid values: debug, info, warn, error)", c.LogLevel)
	}

	// Validate log format
	switch c.LogFormat {
	case "text", "json":
		// Valid log formats
	default:
		return fmt.Errorf("invalid log format: %s (valid values: text, json)", c.LogFormat)
	}

	return nil
}

// SetupLogger configures the global logger based on configuration.
func (c *Config) SetupLogger() {
	// Determine log level
	var level slog.Level
	switch c.LogLevel {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}

	// Configure handler based on format
	var handler slog.Handler
	opts := &slog.HandlerOptions{Level: level}

	if c.LogFormat == "json" {
		handler = slog.NewJSONHandler(os.Stderr, opts)
	} else {
		handler = slog.NewTextHandler(os.Stderr, opts)
	}

	// Set the default logger
	slog.SetDefault(slog.New(handler))
}

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "sttrouter",
	Short: "A CLI tool for audio device management and transcription",
	Long: `sttrouter is a command line tool that enables users to enumerate audio devices,
capture audio, and transcribe it using various transcription backends.

This application supports both macOS-native and cross-platform methods for device discovery,
audio capture, and output management.

Usage:
  sttrouter [command]

Available Commands:
  help        Help about any command

Flags:
  --device-source string   Device enumeration source (auto, ffmpeg, system-profiler) (default "auto")
  -h, --help               help for sttrouter
  --log-format string      Set log format (text, json) (default "text")
  --log-level string       Set log level (debug, info, warn, error) (default "info")
  -v, --verbose            Enable verbose output`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Access the bound configuration
		config, err := getConfig(cmd)
		if err != nil {
			return err
		}

		// Validate the configuration
		if err := config.Validate(); err != nil {
			return err
		}

		// Setup logger based on configuration
		config.SetupLogger()

		return nil
	},
}

// getConfig retrieves the bound configuration from the cobra command.
func getConfig(cmd *cobra.Command) (*Config, error) {
	config := &Config{}

	// Bind flags to config struct
	config.Verbose, _ = cmd.Flags().GetBool("verbose")
	config.LogLevel, _ = cmd.Flags().GetString("log-level")
	config.LogFormat, _ = cmd.Flags().GetString("log-format")

	return config, nil
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.ExecuteContext(context.Background())
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Define persistent flags for the root command
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().String("log-level", "info", "Set log level (debug, info, warn, error)")
	rootCmd.PersistentFlags().String("log-format", "text", "Set log format (text, json)")
	rootCmd.PersistentFlags().String("device-source", "auto", "Device enumeration source (auto, ffmpeg, system-profiler)")
}
