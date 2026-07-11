package filesystem

import (
	"io"
	"os"
)

// RealFileSystem implements the FileSystem interface using the standard os package.
type RealFileSystem struct{}

func NewRealFileSystem() *RealFileSystem {
	return &RealFileSystem{}
}

func (fs *RealFileSystem) Exists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func (fs *RealFileSystem) IsDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func (fs *RealFileSystem) CreateDirectory(path string) error {
	return os.MkdirAll(path, 0755)
}

func (fs *RealFileSystem) CopyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = sourceFile.Close() }()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() { _ = destFile.Close() }()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

func (fs *RealFileSystem) ReadDir(path string) ([]os.DirEntry, error) {
	return os.ReadDir(path)
}

func (fs *RealFileSystem) Stat(path string) (os.FileInfo, error) {
	return os.Stat(path)
}

func (fs *RealFileSystem) Open(path string) (io.ReadCloser, error) {
	return os.Open(path)
}

func (fs *RealFileSystem) Create(path string) (io.WriteCloser, error) {
	return os.Create(path)
}
