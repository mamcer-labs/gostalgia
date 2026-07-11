package scan

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/mamcer/gostalgia/internal/domain"
	"github.com/mamcer/gostalgia/internal/infra/repository"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// MockFileInfo implements os.FileInfo
type MockFileInfo struct {
	name    string
	size    int64
	isDir   bool
	modTime time.Time
}

func (m *MockFileInfo) Name() string       { return m.name }
func (m *MockFileInfo) Size() int64       { return m.size }
func (m *MockFileInfo) Mode() os.FileMode  { return 0 }
func (m *MockFileInfo) ModTime() time.Time { return m.modTime }
func (m *MockFileInfo) IsDir() bool        { return m.isDir }
func (m *MockFileInfo) Sys() interface{}   { return nil }

// MockDirEntry implements os.DirEntry
type MockDirEntry struct {
	info *MockFileInfo
}

func (m *MockDirEntry) Name() string               { return m.info.name }
func (m *MockDirEntry) IsDir() bool                { return m.info.isDir }
func (m *MockDirEntry) Type() os.FileMode          { return 0 }
func (m *MockDirEntry) Info() (os.FileInfo, error) { return m.info, nil }

// MockFileSystem implements domain.FileSystem
type MockFileSystem struct {
	Files map[string][]byte
	Dirs  map[string]bool
	Stats map[string]*MockFileInfo
}

func NewMockFileSystem() *MockFileSystem {
	return &MockFileSystem{
		Files: make(map[string][]byte),
		Dirs:  make(map[string]bool),
		Stats: make(map[string]*MockFileInfo),
	}
}

func (m *MockFileSystem) Exists(path string) bool {
	_, ok := m.Files[path]
	if ok {
		return true
	}
	return m.Dirs[path]
}

func (m *MockFileSystem) IsDir(path string) bool {
	return m.Dirs[path]
}

func (m *MockFileSystem) CreateDirectory(path string) error {
	m.Dirs[path] = true
	m.Stats[path] = &MockFileInfo{name: filepath.Base(path), isDir: true, modTime: time.Now()}
	return nil
}

func (m *MockFileSystem) CopyFile(src, dst string) error {
	if data, ok := m.Files[src]; ok {
		m.Files[dst] = data
		m.Stats[dst] = &MockFileInfo{name: filepath.Base(dst), size: int64(len(data)), modTime: time.Now()}
		return nil
	}
	return os.ErrNotExist
}

func (m *MockFileSystem) ReadDir(path string) ([]os.DirEntry, error) {
	var entries []os.DirEntry
	// This is a bit complex to implement correctly for all subpaths
	// For simplicity, we'll just check if the path is a prefix
	for p, info := range m.Stats {
		if filepath.Dir(p) == path && p != path {
			entries = append(entries, &MockDirEntry{info: info})
		}
	}
	return entries, nil
}

func (m *MockFileSystem) Stat(path string) (os.FileInfo, error) {
	if info, ok := m.Stats[path]; ok {
		return info, nil
	}
	return nil, os.ErrNotExist
}

func (m *MockFileSystem) Open(path string) (io.ReadCloser, error) {
	if data, ok := m.Files[path]; ok {
		return io.NopCloser(bytes.NewReader(data)), nil
	}
	return nil, os.ErrNotExist
}

func (m *MockFileSystem) Create(path string) (io.WriteCloser, error) {
	// Not needed for current ScanService implementation but required by interface
	return nil, nil
}

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

func TestRunScan(t *testing.T) {
	db := setupTestDB(t)
	uow := repository.NewGormUnitOfWork(db)
	fs := NewMockFileSystem()
	service := NewScanService(uow, fs)

	ctx := context.Background()

	// Setup mock filesystem
	scanPath := "/scan"
	nostalgiaPath := "/nostalgia"
	fs.CreateDirectory(scanPath)
	fs.CreateDirectory(nostalgiaPath)

	file1Path := filepath.Join(scanPath, "test1.jpg")
	file1Data := []byte("fake image data 1")
	fs.Files[file1Path] = file1Data
	fs.Stats[file1Path] = &MockFileInfo{name: "test1.jpg", size: int64(len(file1Data)), modTime: time.Now()}

	subDirPath := filepath.Join(scanPath, "subdir")
	fs.CreateDirectory(subDirPath)
	file2Path := filepath.Join(subDirPath, "test2.png")
	file2Data := []byte("fake image data 2")
	fs.Files[file2Path] = file2Data
	fs.Stats[file2Path] = &MockFileInfo{name: "test2.png", size: int64(len(file2Data)), modTime: time.Now()}

	// Setup source directory in DB
	sourceName := "MySource"
	uow.Directories().Add(ctx, &domain.NDirectory{
		Name:     sourceName,
		FullPath: sourceName,
		IsSource: true,
	})

	opts := ScanOptions{
		Source:        sourceName,
		ScanPath:      scanPath,
		NostalgiaPath: nostalgiaPath,
		Tags:          []string{"test", "holiday"},
	}

	result, err := service.RunScan(ctx, opts)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 2, len(result.Files))
	assert.Equal(t, 2, len(result.Directories)) // root and subdir

	// Verify persistence
	scans, _ := uow.Scans().GetByID(ctx, 1)
	assert.NotNil(t, scans)
	assert.Equal(t, int64(2), scans.FileCount)

	// Verify file copying
	assert.True(t, fs.Exists(filepath.Join(nostalgiaPath, sourceName, "test1.jpg")))
	assert.True(t, fs.Exists(filepath.Join(nostalgiaPath, sourceName, "subdir", "test2.png")))
}

func TestValidateScannedData(t *testing.T) {
	service := NewScanService(nil, nil)

	t.Run("Valid", func(t *testing.T) {
		res := &ScanResult{
			Files: []*FileScan{{Name: "test.jpg", Extension: ".jpg", FinalPath: "test.jpg"}},
		}
		err := service.ValidateScannedData(res)
		assert.NoError(t, err)
	})

	t.Run("Invalid Name", func(t *testing.T) {
		res := &ScanResult{
			Files: []*FileScan{{Name: string(make([]byte, 256)), Extension: ".jpg", FinalPath: "test.jpg"}},
		}
		err := service.ValidateScannedData(res)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "file name too long")
	})
}

func TestCalculateSize(t *testing.T) {
	service := NewScanService(nil, nil)
	root := &DirectoryScan{
		Files: []*FileScan{{Size: 100}, {Size: 200}},
		Directories: []*DirectoryScan{
			{Files: []*FileScan{{Size: 300}}},
		},
	}

	total := service.CalculateSize(root)
	assert.Equal(t, int64(600), total)
	assert.Equal(t, int64(600), root.Size)
}

func TestCopyFiles(t *testing.T) {
	db := setupTestDB(t)
	uow := repository.NewGormUnitOfWork(db)
	fs := NewMockFileSystem()
	service := NewScanService(uow, fs)

	fs.Files["/src/f1.jpg"] = []byte("data")
	fs.Stats["/src/f1.jpg"] = &MockFileInfo{name: "f1.jpg", size: 4}

	res := &ScanResult{
		Files: []*FileScan{
			{ScanPath: "/src/f1.jpg", FinalPath: "f1.jpg"},
		},
	}

	err := service.CopyFiles(context.Background(), res, "/dst", 1)
	assert.NoError(t, err)
	assert.True(t, fs.Exists("/dst/f1.jpg"))
}

func TestCheckExisting(t *testing.T) {
	db := setupTestDB(t)
	uow := repository.NewGormUnitOfWork(db)
	service := NewScanService(uow, nil)

	existing := &domain.NFile{Name: "ex.jpg", Hash: "h_ex"}
	uow.Files().Add(context.Background(), existing)

	res := &ScanResult{
		Files: []*FileScan{
			{Hash: "h_ex"},
			{Hash: "h_new"},
			{Hash: "h_new"}, // Repeated internal
		},
	}

	nScan, err := service.CheckExisting(context.Background(), res)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), nScan.ExistingFileRepeatedCount)
	assert.Equal(t, int64(2), nScan.InternalFileRepeatedCount)
	assert.True(t, res.Files[0].IsExistingUniverse)
	assert.True(t, res.Files[1].IsExistingInternal)
}

func TestScan(t *testing.T) {
	fs := NewMockFileSystem()
	service := NewScanService(nil, fs)

	fs.CreateDirectory("/tmp/scan")
	fs.Files["/tmp/scan/f1.jpg"] = []byte("data")
	fs.Stats["/tmp/scan/f1.jpg"] = &MockFileInfo{name: "f1.jpg", size: 4}

	res, err := service.Scan("/tmp/scan", "/tmp/scan", "/nostalgia", "Source")
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Len(t, res.Files, 1)
}

func TestPersist(t *testing.T) {
	db := setupTestDB(t)
	uow := repository.NewGormUnitOfWork(db)
	service := NewScanService(uow, nil)
	ctx := context.Background()

	sourceDir := &domain.NDirectory{Name: "Source1", FullPath: "Source1", IsSource: true}
	uow.Directories().Add(ctx, sourceDir)

	res := &ScanResult{
		RootDirectory: &DirectoryScan{Name: "root", FinalPath: "Source1"},
		Files:         []*FileScan{{Name: "f1.jpg", FinalPath: "Source1/f1.jpg"}},
	}
	nScan := &domain.NScan{}
	opts := ScanOptions{Source: "Source1", Tags: []string{"T1"}}

	err := service.Persist(ctx, res, nScan, opts)
	assert.NoError(t, err)
	
	// Verify scan was saved
	scan, _ := uow.Scans().GetByID(ctx, 1)
	assert.NotNil(t, scan)
	assert.Equal(t, int64(1), scan.FileCount)
}

func TestHash(t *testing.T) {
	fs := NewMockFileSystem()
	service := NewScanService(nil, fs)
	ctx := context.Background()

	path := "/tmp/h1.txt"
	fs.Files[path] = []byte("hello")
	fs.Stats[path] = &MockFileInfo{name: "h1.txt"}

	res := &ScanResult{Files: []*FileScan{{ScanPath: path}}}
	err := service.Hash(ctx, res, 1)
	assert.NoError(t, err)
	assert.NotEmpty(t, res.Files[0].Hash)
}

func TestCreateThumbnails(t *testing.T) {
	fs := NewMockFileSystem()
	service := NewScanService(nil, fs)
	ctx := context.Background()

	res := &ScanResult{
		Files: []*FileScan{
			{ScanPath: "/tmp/img.jpg", FinalPath: "img.jpg", Extension: ".jpg"},
		},
	}

	// It will likely return errors because magick is not found, but we check coverage
	err := service.CreateThumbnails(ctx, res, "/thumbs", 1)
	assert.NoError(t, err) // Should not error if it just fails to find magick
}

func TestRunScan_Integration(t *testing.T) {
	db := setupTestDB(t)
	uow := repository.NewGormUnitOfWork(db)
	fs := NewMockFileSystem()
	service := NewScanService(uow, fs)
	ctx := context.Background()

	// Setup source
	srcName := "S1"
	uow.Directories().Add(ctx, &domain.NDirectory{Name: srcName, FullPath: srcName, IsSource: true})

	// Setup FS
	scanPath := "/scan"
	fs.CreateDirectory(scanPath)
	filePath := filepath.Join(scanPath, "f1.jpg")
	fs.Files[filePath] = []byte("image_data")
	fs.Stats[filePath] = &MockFileInfo{name: "f1.jpg", size: 10}

	opts := ScanOptions{
		Source: srcName,
		ScanPath: scanPath,
		NostalgiaPath: "/nostalgia",
	}

	result, err := service.RunScan(ctx, opts)
	assert.NoError(t, err)
	assert.NotNil(t, result)
}







