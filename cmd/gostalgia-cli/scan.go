package main

import (
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/mamcer/gostalgia/internal/app/scan"
	"github.com/mamcer/gostalgia/internal/infra/database"
	"github.com/mamcer/gostalgia/internal/infra/filesystem"
	"github.com/mamcer/gostalgia/internal/infra/repository"
	"github.com/spf13/cobra"
)

var (
	tags   string
	source string
)

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan a directory and add files to Nostalgia",
	RunE: func(cmd *cobra.Command, args []string) error {
		if source == "" {
			return fmt.Errorf("source is required")
		}

		db := database.NewMySQLDB(cfg)
		uow := repository.NewGormUnitOfWork(db)
		fs := filesystem.NewRealFileSystem()
		
		scanService := scan.NewScanService(uow, fs)

		tagList := []string{}
		if tags != "" {
			tagList = strings.Split(tags, ";")
		}

		opts := scan.ScanOptions{
			Tags:          tagList,
			Source:        source,
			ScanPath:      cfg.NOSTALGIA_SCAN_PATH,
			NostalgiaPath: cfg.NOSTALGIA_HOME_PATH,
			ThumbPath:     cfg.NOSTALGIA_THUMB_TARGET_PATH,
		}

		slog.Info("Starting scan", "source", source, "tags", tagList, "scanPath", opts.ScanPath)

		start := time.Now()
		result, err := scanService.RunScan(cmd.Context(), opts)
		if err != nil {
			return err
		}

		duration := time.Since(start)
		slog.Info("Scan finished", 
			"duration", duration, 
			"directories", len(result.Directories), 
			"files", len(result.Files))

		return nil
	},
}

func init() {
	scanCmd.Flags().StringVarP(&tags, "tags", "t", "", "Semicolon separated list of tags")
	scanCmd.Flags().StringVarP(&source, "source", "s", "", "Source directory name")
	rootCmd.AddCommand(scanCmd)
}
