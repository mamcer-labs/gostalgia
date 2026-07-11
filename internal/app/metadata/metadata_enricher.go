package metadata

import (
	"archive/zip"
	"context"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/mamcer/gostalgia/internal/domain"

	"github.com/bogem/id3v2"
	"github.com/rwcarlsen/goexif/exif"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"sync"
)

var (
	// Patterns like 20031025 or 2003-10-25
	dateRegex = regexp.MustCompile(`(\d{4})[-_]?(\d{2})[-_]?(\d{2})`)
	// Patterns like /photos/2003/ or /2003-Vacation/
	yearRegex = regexp.MustCompile(`\b(19|20)\d{2}\b`)
)

type Enricher struct {
	db *gorm.DB
}

func NewEnricher(db *gorm.DB) *Enricher {
	return &Enricher{db: db}
}

// Run starts the enrichment process for a batch of files using a worker pool
func (e *Enricher) Run(ctx context.Context, batchSize int, workerCount int) (int, error) {
	var files []domain.NFile

	// Find files that haven't been enriched yet
	result := e.db.Where("metadata IS NULL").Limit(batchSize).Find(&files)
	if result.Error != nil {
		return 0, result.Error
	}

	count := len(files)
	if count == 0 {
		return 0, nil
	}

	// Parallel processing using a worker pool
	fileChan := make(chan domain.NFile, count)
	var wg sync.WaitGroup

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for f := range fileChan {
				select {
				case <-ctx.Done():
					return
				default:
					e.enrichFile(&f)
				}
			}
		}()
	}

	for _, f := range files {
		fileChan <- f
	}
	close(fileChan)
	wg.Wait()

	return count, nil
}

func (e *Enricher) enrichFile(file *domain.NFile) {
	fullPath := file.Path

	metadata := make(map[string]interface{})
	ext := strings.ToLower(filepath.Ext(file.Name))

	switch ext {
	case ".txt", ".md", ".log":
		e.extractTextMetadata(fullPath, metadata)
	case ".zip":
		e.extractZipMetadata(fullPath, metadata)
	case ".jpg", ".jpeg", ".tiff":
		e.extractExifMetadata(fullPath, metadata)
	case ".mp3":
		e.extractID3Metadata(fullPath, metadata)
	}

	// FALLBACK STRATEGY for captured_at
	var capturedAt *time.Time

	// 1. Try from metadata (EXIF/ID3)
	if capturedAtStr, ok := metadata["captured_at"].(string); ok {
		t, err := time.Parse(time.RFC3339, capturedAtStr)
		if err == nil {
			capturedAt = &t
		}
	}

	// 2. Fallback to Filename Regex (e.g., IMG_20031025_...)
	if capturedAt == nil {
		capturedAt = e.tryParseDate(file.Name)
		if capturedAt != nil {
			metadata["date_source"] = "filename"
		}
	}

	// 3. Fallback to Directory Name (e.g., /Photos/2003/...)
	if capturedAt == nil {
		capturedAt = e.tryParseYear(file.Path)
		if capturedAt != nil {
			metadata["date_source"] = "directory"
		}
	}

	metadata["enriched_at"] = time.Now().Format(time.RFC3339)

	jsonData, _ := json.Marshal(metadata)
	file.Metadata = datatypes.JSON(jsonData)
	file.CapturedAt = capturedAt

	if err := e.db.Save(file).Error; err != nil {
		log.Printf("Error saving metadata for file %d: %v", file.ID, err)
	}
}

func (e *Enricher) tryParseDate(input string) *time.Time {
	match := dateRegex.FindStringSubmatch(input)
	if len(match) == 4 {
		t, err := time.Parse("2006-01-02", match[1]+"-"+match[2]+"-"+match[3])
		if err == nil {
			return &t
		}
	}
	return nil
}

func (e *Enricher) tryParseYear(path string) *time.Time {
	matches := yearRegex.FindAllString(path, -1)
	if len(matches) > 0 {
		// Take the last year found in path (usually the most specific)
		year := matches[len(matches)-1]
		t, err := time.Parse("2006-01-02", year+"-01-01")
		if err == nil {
			return &t
		}
	}
	return nil
}

func (e *Enricher) extractID3Metadata(path string, meta map[string]interface{}) {
	tag, err := id3v2.Open(path, id3v2.Options{Parse: true})
	if err != nil {
		return
	}
	defer tag.Close()

	meta["type"] = "audio"
	meta["title"] = tag.Title()
	meta["artist"] = tag.Artist()
	meta["album"] = tag.Album()
	meta["genre"] = tag.Genre()
	meta["year"] = tag.Year()
}

func (e *Enricher) extractExifMetadata(path string, meta map[string]interface{}) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	x, err := exif.Decode(f)
	if err != nil {
		return
	}

	meta["type"] = "image"

	// Extract Date
	tm, err := x.DateTime()
	if err == nil {
		meta["captured_at"] = tm.Format(time.RFC3339)
	}

	// Extract Camera Model
	camModel, _ := x.Get(exif.Model)
	if camModel != nil {
		meta["camera_model"], _ = camModel.StringVal()
	}

	// Extract GPS
	lat, long, err := x.LatLong()
	if err == nil {
		meta["gps"] = map[string]float64{
			"lat":  lat,
			"long": long,
		}
	}
}

func (e *Enricher) extractTextMetadata(path string, meta map[string]interface{}) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return
	}

	// Only read if smaller than 100KB
	if fileInfo.Size() < 100*1024 {
		content, err := os.ReadFile(path)
		if err == nil {
			meta["content_snippet"] = string(content)
			meta["is_full_content"] = true
		}
	}
}

func (e *Enricher) extractZipMetadata(path string, meta map[string]interface{}) {
	r, err := zip.OpenReader(path)
	if err != nil {
		meta["error"] = "failed to open zip"
		return
	}
	defer r.Close()

	var files []string
	for _, f := range r.File {
		files = append(files, f.Name)
		// Limit to first 100 files to avoid massive metadata blobs
		if len(files) > 100 {
			meta["truncated"] = true
			break
		}
	}
	meta["internal_files"] = files
	meta["file_count"] = len(r.File)
}
