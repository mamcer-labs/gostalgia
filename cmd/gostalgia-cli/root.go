package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/mamcer/gostalgia/internal/config"
	"github.com/mamcer/gostalgia/internal/infra/database"
	"log/slog"
)

var cfg *config.Config

var rootCmd = &cobra.Command{
	Use:   "nostalgia-cli",
	Short: "Nostalgia CLI tool for managing scans and thumbnails",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
		slog.SetDefault(logger)

		cfg = config.LoadConfig()
		database.RunMigrations(cfg)
	},
}

func Execute() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
}

func initConfig() {
	viper.AutomaticEnv()
}
