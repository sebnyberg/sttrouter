package cmd

import (
	"context"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// ConfigKey is the context key for the config
type ConfigKey struct{}

// EnvPrefix is the prefix for environment variables
const EnvPrefix = "STTR"

// getConfigFromContext retrieves the config from the command context
func getConfigFromContext(ctx context.Context) *Config {
	if config, ok := ctx.Value(ConfigKey{}).(*Config); ok {
		return config
	}
	return nil
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
   -h, --help               help for sttrouter
   --log-format string      Set log format (text, json) (default "text")
   --log-level string       Set log level (debug, info, warn, error) (default "info")
   -v, --verbose            Enable verbose output`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Define flags
	rootCmd.PersistentFlags().String("log-level", "info", "Set log level (debug, info, warn, error)")
	rootCmd.PersistentFlags().String("log-format", "text", "Set log format (text, json)")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose output")

	// Setup config in PersistentPreRunE
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		// Initialize viper
		viper.SetEnvPrefix(EnvPrefix)
		viper.AutomaticEnv()

		// Bind flags (map hyphenated flag names to nested viper keys)
		flagMappings := map[string]string{
			"log.level":  "log-level",
			"log.format": "log-format",
			"verbose":    "verbose",
		}
		for viperKey, flagName := range flagMappings {
			_ = viper.BindPFlag(viperKey, rootCmd.PersistentFlags().Lookup(flagName))
		}

		// Parse config
		config := &Config{}
		if err := viper.Unmarshal(config); err != nil {
			return err
		}

		// Validate config
		if err := config.validate(); err != nil {
			return err
		}

		// Set config in context
		ctx := context.WithValue(cmd.Context(), ConfigKey{}, config)
		cmd.SetContext(ctx)

		return nil
	}
}
