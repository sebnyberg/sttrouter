package cmd

import (
	"fmt"
	"log/slog"
	"os"
)

// EnvPrefix is the prefix for environment variables
const EnvPrefix = "STTR"

// Config holds the global application configuration.
// This structure enables clean separation between CLI parsing and business logic.
type Config struct {
	// Log holds logging configuration
	Log LogConfig `name:"log"`

	// Verbose enables more detailed logging and output
	Verbose bool `name:"verbose" usage:"Enable verbose output"`
}

// LogConfig holds logging-related configuration
type LogConfig struct {
	// Level defines the minimum log level to output (debug, info, warn, error)
	Level string `value:"info" usage:"Set log level (debug, info, warn, error)"`

	// Format specifies the log output format (text, json)
	Format string `value:"text" usage:"Set log format (text, json)"`
}

// validate validates the configuration and sets defaults.
func (c *Config) validate() error {
	// Validate log level
	switch c.Log.Level {
	case "debug", "info", "warn", "error":
		// Valid log levels
	default:
		return fmt.Errorf("invalid log level: %s (valid values: debug, info, warn, error)", c.Log.Level)
	}

	// Validate log format
	switch c.Log.Format {
	case "text", "json":
		// Valid log formats
	default:
		return fmt.Errorf("invalid log format: %s (valid values: text, json)", c.Log.Format)
	}

	return nil
}

// getLogger returns a configured slog.Logger based on the config
func (c *Config) getLogger() *slog.Logger {
	// Map string levels to slog levels
	var level slog.Level
	switch c.Log.Level {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		if c.Verbose {
			level = slog.LevelDebug // show debug when verbose
		} else {
			level = slog.LevelWarn // hide info/debug by default
		}
	}

	// Create handler based on format
	var handler slog.Handler
	switch c.Log.Format {
	case "json":
		handler = slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: level})
	case "text":
		fallthrough
	default:
		handler = slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})
	}

	return slog.New(handler)
}
