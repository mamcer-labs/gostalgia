package thumb

import (
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/mamcer/gostalgia/internal/domain"
)

type ThumbService struct {
	uow domain.UnitOfWork
	fs  domain.FileSystem
}

func NewThumbService(uow domain.UnitOfWork, fs domain.FileSystem) *ThumbService {
	return &ThumbService{
		uow: uow,
		fs:  fs,
	}
}

type ThumbOptions struct {
	Size          int
	NostalgiaPath string
	TargetPath    string
	NumWorkers    int
}

func (s *ThumbService) GenerateThumbnails(ctx context.Context, opts ThumbOptions) error {
	imageExtensions := []string{".jpg", ".jpeg", ".png", ".gif", ".bmp"}

	total, err := s.uow.Files().CountByExtensions(ctx, imageExtensions)
	if err != nil {
		return err
	}

	if total == 0 {
		return nil
	}

	slog.Info("Starting thumbnail generation", "total_files", total)

	imageCmd, err := s.findImageMagick()
	if err != nil {
		return fmt.Errorf("imagemagick not found: %w", err)
	}

	numWorkers := runtime.NumCPU()
	if opts.NumWorkers > 0 {
		numWorkers = opts.NumWorkers
	}
	if numWorkers < 2 {
		numWorkers = 2
	}

	pageSize := 100
	page := 1

	var wg sync.WaitGroup
	filesChan := make(chan *domain.NFile, pageSize)

	// Stats
	var successCount, skippedCount, errorCount int32
	var statsMutex sync.Mutex

	// Worker pool
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for f := range filesChan {
				select {
				case <-ctx.Done():
					return
				default:
					sourceFile := filepath.Join(opts.NostalgiaPath, f.Path)
					targetFile := filepath.Join(opts.TargetPath, f.Path)

					// Check if exists
					if s.fs.Exists(targetFile) {
						statsMutex.Lock()
						skippedCount++
						statsMutex.Unlock()
						continue
					}

					// Check source
					if !s.fs.Exists(sourceFile) {
						slog.Error("Source file not found", "path", sourceFile)
						statsMutex.Lock()
						errorCount++
						statsMutex.Unlock()
						continue
					}

					// Ensure target dir
					if err := s.fs.CreateDirectory(filepath.Dir(targetFile)); err != nil {
						slog.Error("Error creating target directory", "path", targetFile, "error", err)
						statsMutex.Lock()
						errorCount++
						statsMutex.Unlock()
						continue
					}

					// Generate
					sizeStr := fmt.Sprintf("%dx%d", opts.Size, opts.Size)
					cmd := exec.Command(imageCmd, sourceFile, "-resize", sizeStr, targetFile)
					if err := cmd.Run(); err != nil {
						slog.Error("Error processing image", "path", sourceFile, "error", err)
						statsMutex.Lock()
						errorCount++
						statsMutex.Unlock()
					} else {
						statsMutex.Lock()
						successCount++
						statsMutex.Unlock()
					}
				}
			}
		}()
	}

	// Producer
	for {
		files, err := s.uow.Files().GetByExtensionsPaged(ctx, imageExtensions, page, pageSize)
		if err != nil {
			close(filesChan)
			return err
		}
		if len(files) == 0 {
			break
		}
		for _, f := range files {
			filesChan <- f
		}
		page++
	}
	close(filesChan)
	wg.Wait()

	slog.Info("Thumbnail generation complete",
		"success", successCount,
		"skipped", skippedCount,
		"failed", errorCount)

	return nil
}

func (s *ThumbService) findImageMagick() (string, error) {
	for _, cmd := range []string{"magick", "convert"} {
		if _, err := exec.LookPath(cmd); err == nil {
			return cmd, nil
		}
	}
	return "", fmt.Errorf("imagemagick (magick or convert) not found in PATH")
}
