package main

import (
	"log/slog"

	"github.com/mamcer/gostalgia/internal/app/thumb"
	"github.com/mamcer/gostalgia/internal/infra/database"
	"github.com/mamcer/gostalgia/internal/infra/filesystem"
	"github.com/mamcer/gostalgia/internal/infra/repository"
	"github.com/spf13/cobra"
)

var (
	size       int
	numWorkers int
)

var thumbCmd = &cobra.Command{
	Use:   "thumb",
	Short: "Generate thumbnails for all images in the database",
	RunE: func(cmd *cobra.Command, args []string) error {
		db := database.NewMySQLDB(cfg)
		uow := repository.NewGormUnitOfWork(db)
		fs := filesystem.NewRealFileSystem()

		thumbService := thumb.NewThumbService(uow, fs)

		slog.Info("Starting thumbnail generation", "size", size, "workers", numWorkers)

		opts := thumb.ThumbOptions{
			Size:          size,
			NostalgiaPath: cfg.NOSTALGIA_HOME_PATH,
			TargetPath:    cfg.NOSTALGIA_THUMB_TARGET_PATH,
			NumWorkers:    numWorkers,
		}

		if err := thumbService.GenerateThumbnails(cmd.Context(), opts); err != nil {
			return err
		}

		slog.Info("Thumbnail generation finished")
		return nil
	},
}

func init() {
	thumbCmd.Flags().IntVarP(&size, "size", "s", 256, "Size of the thumbnails (width and height)")
	thumbCmd.Flags().IntVarP(&numWorkers, "workers", "w", 4, "Number of concurrent workers")
	rootCmd.AddCommand(thumbCmd)
}
