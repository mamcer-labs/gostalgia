package file

import (
	"context"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/mamcer/gostalgia/internal/domain"
	"github.com/mamcer/gostalgia/internal/infra/repository"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

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

func TestFileService(t *testing.T) {
	db := setupTestDB(t)
	uow := repository.NewGormUnitOfWork(db)
	service := NewFileService(uow, nil)
	ctx := context.Background()

	// Seed data
	tag1 := &domain.NTag{Name: "Nature"}
	db.Create(tag1)

	file1 := &domain.NFile{
		Name:         "test.jpg",
		Extension:    ".jpg",
		Path:         "nature/test.jpg",
		DateModified: time.Now(),
		Size:         1024,
		Hash:         "hash1",
		Tags:         []*domain.NTag{tag1},
	}
	db.Create(file1)

	parentDir := &domain.NDirectory{
		Name: "nature",
		FileNodes: []*domain.NFileNode{
			{NFileID: file1.ID},
		},
	}
	db.Create(parentDir)

	t.Run("Count", func(t *testing.T) {
		count, err := service.Count(ctx)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), count)
	})

	t.Run("GetByID", func(t *testing.T) {
		dto, err := service.GetByID(ctx, file1.ID)
		assert.NoError(t, err)
		assert.NotNil(t, dto)
		assert.Equal(t, "test.jpg", dto.Name)
		assert.Equal(t, "1.02 KB", dto.Size)
		assert.Len(t, dto.Tags, 1)
		assert.Equal(t, "nature", dto.ParentDirectory.Name)
	})

	t.Run("Search", func(t *testing.T) {
		res, err := service.Search(ctx, "test", nil, nil, "image", 1, 10)
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, 1, res.Total)
		assert.Equal(t, "test.jpg", res.Result[0].Name)
	})

	t.Run("SearchByTag", func(t *testing.T) {
		res, err := service.SearchByTag(ctx, "Nature", 1, 10)
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, 1, res.Total)
		assert.Equal(t, "test.jpg", res.Result[0].Name)
	})
}
