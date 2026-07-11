package tag

import (
	"context"
	"testing"

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

func TestTagService(t *testing.T) {
	db := setupTestDB(t)
	uow := repository.NewGormUnitOfWork(db)
	service := NewTagService(uow, nil)
	ctx := context.Background()

	t.Run("Add and GetByName", func(t *testing.T) {
		tag := &domain.NTag{Name: "Nature"}
		err := service.Add(ctx, tag)
		assert.NoError(t, err)

		found, err := service.GetByName(ctx, "Nature")
		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, "Nature", found.Name)
	})

	t.Run("GetAllTags", func(t *testing.T) {
		service.Add(ctx, &domain.NTag{Name: "Urban"})
		service.Add(ctx, &domain.NTag{Name: "Abstract"})

		tags, err := service.GetAllTags(ctx)
		assert.NoError(t, err)
		assert.Len(t, tags, 3) // Nature, Urban, Abstract
		assert.Equal(t, "Abstract", tags[0])
		assert.Equal(t, "Nature", tags[1])
		assert.Equal(t, "Urban", tags[2])
	})
}
