package cmd

import "fmt"

// Config holds the global application configuration.
// This structure enables clean separation between CLI parsing and business logic.
type Config struct {
	// Log holds logging configuration
	Log LogConfig `mapstructure:"log"`

	// Verbose enables more detailed logging and output
	Verbose bool `mapstructure:"verbose" default:"false"`
}

// LogConfig holds logging-related configuration
type LogConfig struct {
	// Level defines the minimum log level to output (debug, info, warn, error)
	Level string `mapstructure:"level" default:"info"`

	// Format specifies the log output format (text, json)
	Format string `mapstructure:"format" default:"text"`
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
