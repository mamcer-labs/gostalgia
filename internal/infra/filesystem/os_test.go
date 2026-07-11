package filesystem

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRealFileSystem(t *testing.T) {
	fs := NewRealFileSystem()
	tmpDir, err := os.MkdirTemp("", "nostalgia_test")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	t.Run("CreateDirectory and Exists", func(t *testing.T) {
		path := filepath.Join(tmpDir, "subdir")
		err := fs.CreateDirectory(path)
		assert.NoError(t, err)
		assert.True(t, fs.Exists(path))
		assert.True(t, fs.IsDir(path))
	})

	t.Run("CopyFile", func(t *testing.T) {
		src := filepath.Join(tmpDir, "src.txt")
		dst := filepath.Join(tmpDir, "dst.txt")
		err := os.WriteFile(src, []byte("hello"), 0644)
		assert.NoError(t, err)

		err = fs.CopyFile(src, dst)
		assert.NoError(t, err)
		assert.True(t, fs.Exists(dst))

		content, _ := os.ReadFile(dst)
		assert.Equal(t, "hello", string(content))
	})

	t.Run("ReadDir", func(t *testing.T) {
		entries, err := fs.ReadDir(tmpDir)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(entries), 2)
	})

	t.Run("Stat", func(t *testing.T) {
		info, err := fs.Stat(tmpDir)
		assert.NoError(t, err)
		assert.NotNil(t, info)
		assert.True(t, info.IsDir())
	})
}
