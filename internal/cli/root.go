// Package cli builds and executes the rmqctl command tree.
//
// It defines all user-facing commands and their flags, handles config loading
// and validation, and invokes the underlying api package to interact with
// RabbitMQ.
package cli

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/nepec/rmqctl/internal/api"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type rootCommand struct {
	cmd *cobra.Command

	cfgFile  string // Flag
	logLevel string // Flag
}

func newRootCommand(info BuildInfo) *rootCommand {
	c := &rootCommand{}

	rootCmd := &cobra.Command{
		Use:          "rmqctl",
		Short:        "rmqctl - a tool to remotely admin RabbitMQ resources",
		SilenceUsage: false,
		Args:         cobra.NoArgs,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if err := c.initConfig(); err != nil {
				return err
			}
			setupLogger(c.logLevel)
			return nil
		},
	}

	rootCmd.PersistentFlags().SortFlags = false
	rootCmd.Flags().SortFlags = false

	// Local flags
	rootCmd.PersistentFlags().StringVar(&c.cfgFile, "config", "", "Config file (default is $HOME/.rmqctl.yaml")
	rootCmd.PersistentFlags().StringVar(&c.logLevel, "log-level", "info", "Set log verbosity. (Default is 'info')")

	// Viper Flags
	rootCmd.PersistentFlags().StringP("hostname", "H", "localhost", "Management API hostname")
	rootCmd.PersistentFlags().IntP("port", "P", 15672, "Management API port")
	rootCmd.PersistentFlags().StringP("username", "u", "guest", "Username for API authentication")
	rootCmd.PersistentFlags().StringP("password", "p", "guest", "Password for API authentication")

	_ = viper.BindPFlag("hostname", rootCmd.PersistentFlags().Lookup("hostname"))
	_ = viper.BindPFlag("port", rootCmd.PersistentFlags().Lookup("port"))
	_ = viper.BindPFlag("username", rootCmd.PersistentFlags().Lookup("username"))
	_ = viper.BindPFlag("password", rootCmd.PersistentFlags().Lookup("password"))

	replacer := strings.NewReplacer("-", "_")
	viper.SetEnvKeyReplacer(replacer)
	viper.SetEnvPrefix("RMQ")
	viper.AutomaticEnv()

	clientFactory := func() (api.RabbitClient, error) {
		return ClientFromConfig()
	}

	// Commands
	rootCmd.AddCommand(NewVersionCommand(info))
	rootCmd.AddCommand(NewListCommand(clientFactory))
	rootCmd.AddCommand(NewMergeCommand(clientFactory))
	rootCmd.AddCommand(NewApplyCommand(clientFactory))

	c.cmd = rootCmd

	return c
}

func (c rootCommand) initConfig() error {
	if c.cfgFile != "" {
		// Use config file from the flag
		viper.SetConfigFile(c.cfgFile)
	} else {
		// Find home directory
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("resolving home dir: %w", err)
		}

		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("resolving current working dir: %w", err)
		}

		configName := ".rmqctl"

		viper.AddConfigPath(home)
		viper.AddConfigPath(cwd)
		viper.SetConfigType("yaml")
		viper.SetConfigName(configName)
	}

	if err := viper.ReadInConfig(); err != nil {
		var fileNotFound viper.ConfigFileNotFoundError
		if !errors.As(err, &fileNotFound) {
			return fmt.Errorf("reading config file: %w", err)
		}
	} else {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}

	return nil
}

func setupLogger(chosenLevel string) {
	var level slog.Level

	switch strings.ToLower(chosenLevel) {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: level,
	}
	handler := slog.NewTextHandler(os.Stderr, opts)

	logger := slog.New(handler)
	slog.SetDefault(logger)
}

func (c rootCommand) Execute() error {
	return c.cmd.Execute()
}

// Execute builds the rmqctl command tree and runs it against os.Args,
// embedding info (e.g. the build version) into the "version" subcommand.
func Execute(info BuildInfo) error {
	r := newRootCommand(info)
	return r.Execute()
}
