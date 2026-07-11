package directory

import (
	"context"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/mamcer/gostalgia/internal/app/tag"
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

func TestDirectoryService(t *testing.T) {
	db := setupTestDB(t)
	uow := repository.NewGormUnitOfWork(db)
	tagService := tag.NewTagService(uow, nil)
	service := NewDirectoryService(uow, tagService)
	ctx := context.Background()

	// Seed data
	root := &domain.NDirectory{
		Name:         "root",
		FullPath:     "root",
		DateModified: time.Now(),
		Size:         2048,
		FileCount:    2,
		IsSource:     false,
	}
	db.Create(root)

	file1 := &domain.NFile{
		Name:      "f1.jpg",
		Extension: ".jpg",
		Size:      1024,
	}
	db.Create(file1)

	db.Create(&domain.NFileNode{
		NDirectoryID: root.ID,
		NFileID:      file1.ID,
	})

	child := &domain.NDirectory{
		Name:              "child",
		ParentDirectoryID: root.ID,
		FullPath:          "root/child",
	}
	db.Create(child)

	t.Run("GetByID", func(t *testing.T) {
		dto, err := service.GetByID(ctx, root.ID)
		assert.NoError(t, err)
		assert.NotNil(t, dto)
		assert.Equal(t, "root", dto.Name)
		assert.Len(t, dto.Files, 1)
		assert.Len(t, dto.Directories, 1)
	})

	t.Run("AddTagToDirectory", func(t *testing.T) {
		success, err := service.AddTagToDirectory(ctx, root.ID, "Vacation")
		assert.NoError(t, err)
		assert.True(t, success)

		// Verify tag added to dir
		dir, _ := uow.Directories().GetByID(ctx, root.ID)
		assert.Len(t, dir.Tags, 1)
		assert.Equal(t, "Vacation", dir.Tags[0].Name)

		// Verify tag added to file via filenode
		f, _ := uow.Files().GetByID(ctx, file1.ID)
		assert.Len(t, f.Tags, 1)
		assert.Equal(t, "Vacation", f.Tags[0].Name)
	})

	t.Run("GetFiles", func(t *testing.T) {
		files, err := service.GetFiles(ctx, root.ID)
		assert.NoError(t, err)
		assert.Len(t, files, 1)
		assert.Equal(t, "f1.jpg", files[0].Name)
	})

	t.Run("GetDirectories", func(t *testing.T) {
		dirs, err := service.GetDirectories(ctx, root.ID)
		assert.NoError(t, err)
		assert.Len(t, dirs, 1)
		assert.Equal(t, "child", dirs[0].Name)
	})

	t.Run("Search", func(t *testing.T) {
		res, err := service.Search(ctx, "root", nil, nil, 1, 10)
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, int64(1), res.Total)
		assert.Equal(t, "root", res.Result[0].Name)
	})
}
