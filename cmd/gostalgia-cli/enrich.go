package main

import (
	"fmt"
	"github.com/mamcer/gostalgia/internal/app/metadata"
	"github.com/mamcer/gostalgia/internal/infra/database"
	"github.com/spf13/cobra"
	"log/slog"
)

var (
	batchSize   int
	totalLimit  int
	workerCount int
)

var enrichCmd = &cobra.Command{
	Use:   "enrich",
	Short: "Enrich files with metadata (EXIF, ZIP contents, Text snippets, ID3)",
	RunE: func(cmd *cobra.Command, args []string) error {
		db := database.NewMySQLDB(cfg)
		enricher := metadata.NewEnricher(db)

		slog.Info("Starting metadata enrichment",
			"batchSize", batchSize,
			"totalLimit", totalLimit,
			"workers", workerCount)

		processed := 0
		for {
			// Check if we reached the total limit
			currentBatch := batchSize
			if totalLimit > 0 && processed+batchSize > totalLimit {
				currentBatch = totalLimit - processed
			}

			if currentBatch <= 0 {
				break
			}

			slog.Info("Processing batch", "offset", processed, "size", currentBatch)

			count, err := enricher.Run(cmd.Context(), currentBatch, workerCount)
			if err != nil {
				return fmt.Errorf("enrichment failed: %w", err)
			}

			if count == 0 {
				slog.Info("No more files to enrich")
				break
			}

			processed += count

			if totalLimit > 0 && processed >= totalLimit {
				break
			}
		}

		slog.Info("Enrichment finished", "totalProcessed", processed)
		return nil
	},
}

func init() {
	enrichCmd.Flags().IntVarP(&batchSize, "batch", "b", 100, "Size of each processing batch")
	enrichCmd.Flags().IntVarP(&totalLimit, "limit", "l", 0, "Total number of files to process (0 for unlimited)")
	enrichCmd.Flags().IntVarP(&workerCount, "workers", "w", 4, "Number of parallel workers")
	rootCmd.AddCommand(enrichCmd)
}
