// Package cli builds and executes the rmqctl command tree.
//
// It defines all user-facing commands and their flags, handles config loading
// and validation, and invokes the underlying api package to interact with
// RabbitMQ.
package cli

import (
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

var RootCmd = &cobra.Command{
	Use:          "rmqctl",
	Short:        "rmqctl - a tool to remotely admin RabbitMQ nodes",
	Aliases:      []string{"rctl"},
	SilenceUsage: false,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		loglevel := viper.GetString("log-level")
		setupLogger(loglevel)

		return nil
	},
}

func Execute() {
	err := RootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	RootCmd.PersistentFlags().SortFlags = false
	RootCmd.Flags().SortFlags = false

	// Local flags
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "Config file (default is $HOME/.rmqctl.yaml")

	// Viper Flags
	RootCmd.PersistentFlags().StringP("hostname", "H", "localhost", "Management API hostname")
	RootCmd.PersistentFlags().IntP("port", "P", 15672, "Management API port")
	RootCmd.PersistentFlags().StringP("username", "u", "guest", "Username for API authentication")
	RootCmd.PersistentFlags().StringP("password", "p", "guest", "Password for API authentication")
	RootCmd.PersistentFlags().String("log-level", "info", "Log level (debug, info, warn, error")

	// Flag Bindings
	_ = viper.BindPFlag("log-level", RootCmd.PersistentFlags().Lookup("log-level"))
	_ = viper.BindPFlag("hostname", RootCmd.PersistentFlags().Lookup("hostname"))
	_ = viper.BindPFlag("port", RootCmd.PersistentFlags().Lookup("port"))
	_ = viper.BindPFlag("username", RootCmd.PersistentFlags().Lookup("username"))
	_ = viper.BindPFlag("password", RootCmd.PersistentFlags().Lookup("password"))

	replacer := strings.NewReplacer("-", "_")
	viper.SetEnvKeyReplacer(replacer)
	viper.SetEnvPrefix("RMQ")
}

func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		cwd, err := os.Getwd()
		cobra.CheckErr(err)

		configName := ".rmqctl"
		if runtime.GOOS == "linux" {
			configName = ".rmqctl.yaml"
		}

		viper.AddConfigPath(home)
		viper.AddConfigPath(cwd)
		viper.SetConfigType("yaml")
		viper.SetConfigFile(configName)
	}

	viper.AutomaticEnv() // read environment variables that match

	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
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
