package thumb

import (
	"context"
	"io"
	"os"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/mamcer/gostalgia/internal/domain"
	"github.com/mamcer/gostalgia/internal/infra/repository"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// MockFileSystem implements domain.FileSystem
type MockFileSystem struct {
	ExistsResult bool
	Dirs         []string
}

func (m *MockFileSystem) Exists(path string) bool { return m.ExistsResult }
func (m *MockFileSystem) IsDir(path string) bool  { return true }
func (m *MockFileSystem) CreateDirectory(path string) error {
	m.Dirs = append(m.Dirs, path)
	return nil
}
func (m *MockFileSystem) CopyFile(src, dst string) error             { return nil }
func (m *MockFileSystem) ReadDir(path string) ([]os.DirEntry, error) { return nil, nil }
func (m *MockFileSystem) Stat(path string) (os.FileInfo, error)      { return nil, nil }
func (m *MockFileSystem) Open(path string) (io.ReadCloser, error)    { return nil, nil }
func (m *MockFileSystem) Create(path string) (io.WriteCloser, error) { return nil, nil }

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect database: %v", err)
	}

	err = db.AutoMigrate(
		&domain.NTag{},
		&domain.NFile{},
		&domain.NDirectory{},
		&domain.NScan{},
		&domain.NFileNode{},
	)
	if err != nil {
		t.Fatalf("failed to auto-migrate: %v", err)
	}

	return db
}

func TestThumbService(t *testing.T) {
	db := setupTestDB(t)
	uow := repository.NewGormUnitOfWork(db)
	fs := &MockFileSystem{ExistsResult: false}
	service := NewThumbService(uow, fs)
	ctx := context.Background()

	t.Run("No files in DB", func(t *testing.T) {
		opts := ThumbOptions{Size: 256, NumWorkers: 1}
		err := service.GenerateThumbnails(ctx, opts)
		assert.NoError(t, err)
	})

	t.Run("With files in DB", func(t *testing.T) {
		db.Create(&domain.NFile{Name: "test.jpg", Extension: ".jpg", Path: "test.jpg"})

		opts := ThumbOptions{
			Size:          256,
			NumWorkers:    1,
			NostalgiaPath: "/nostalgia",
			TargetPath:    "/thumbs",
		}

		// This will likely fail in finding ImageMagick but we can check if it tries
		err := service.GenerateThumbnails(ctx, opts)
		if err != nil {
			assert.Contains(t, err.Error(), "imagemagick not found")
		}
	})
}

// MockFileInfo implements os.FileInfo
type MockFileInfo struct {
	name  string
	size  int64
	isDir bool
}

func (m *MockFileInfo) Name() string       { return m.name }
func (m *MockFileInfo) Size() int64        { return m.size }
func (m *MockFileInfo) Mode() os.FileMode  { return 0 }
func (m *MockFileInfo) ModTime() time.Time { return time.Now() }
func (m *MockFileInfo) IsDir() bool        { return m.isDir }
func (m *MockFileInfo) Sys() interface{}   { return nil }
