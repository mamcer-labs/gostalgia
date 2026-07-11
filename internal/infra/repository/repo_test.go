package repository

import (
	"context"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/mamcer/gostalgia/internal/domain"
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

func TestNFileRepository(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormNFileRepository(db)
	ctx := context.Background()

	t.Run("Add and GetByID", func(t *testing.T) {
		file := &domain.NFile{
			Name:         "test.txt",
			Extension:    ".txt",
			Path:         "/tmp/test.txt",
			DateModified: time.Now(),
			Size:         100,
			Hash:         "hash123",
		}

		err := repo.Add(ctx, file)
		assert.NoError(t, err)
		assert.NotZero(t, file.ID)

		fetched, err := repo.GetByID(ctx, file.ID)
		assert.NoError(t, err)
		assert.Equal(t, file.Name, fetched.Name)
		assert.Equal(t, file.Hash, fetched.Hash)
	})

	t.Run("Exists and GetByHash", func(t *testing.T) {
		hash := "unique_hash"
		file := &domain.NFile{
			Name: "hash_file.txt",
			Hash: hash,
		}
		repo.Add(ctx, file)

		exists, err := repo.Exists(ctx, hash)
		assert.NoError(t, err)
		assert.True(t, exists)

		fetched, err := repo.GetByHash(ctx, hash)
		assert.NoError(t, err)
		assert.Equal(t, file.Name, fetched.Name)

		exists, err = repo.Exists(ctx, "non_existent")
		assert.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("SearchByTag", func(t *testing.T) {
		tag := &domain.NTag{Name: "Nature"}
		db.Create(tag)
		file := &domain.NFile{Name: "nature.jpg", Hash: "nature_hash", Tags: []*domain.NTag{tag}}
		repo.Add(ctx, file)

		files, total, err := repo.SearchByTag(ctx, "Nature", 1, 10)
		assert.NoError(t, err)
		assert.Equal(t, 1, total)
		assert.Equal(t, "nature.jpg", files[0].Name)
	})
}

func TestNDirectoryRepository(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormNDirectoryRepository(db)
	ctx := context.Background()

	t.Run("Add and GetByID with Preloads", func(t *testing.T) {
		tag := &domain.NTag{Name: "Tag1"}
		db.Create(tag)
		dir := &domain.NDirectory{Name: "Dir1", FullPath: "/Dir1", Tags: []*domain.NTag{tag}}
		repo.Add(ctx, dir)

		fetched, err := repo.GetByID(ctx, dir.ID)
		assert.NoError(t, err)
		assert.NotNil(t, fetched)
		assert.Equal(t, "Dir1", fetched.Name)
		assert.Len(t, fetched.Tags, 1)
	})

	t.Run("GetFiles and GetDirectories", func(t *testing.T) {
		root := &domain.NDirectory{Name: "root", FullPath: "root"}
		repo.Add(ctx, root)
		
		child := &domain.NDirectory{Name: "child", FullPath: "root/child", ParentDirectoryID: root.ID}
		repo.Add(ctx, child)

		file := &domain.NFile{Name: "f1.txt", Hash: "f1_hash"}
		db.Create(file)
		db.Create(&domain.NFileNode{NDirectoryID: root.ID, NFileID: file.ID, Name: "f1.txt"})

		files, err := repo.GetFiles(ctx, root.ID)
		assert.NoError(t, err)
		assert.Len(t, files, 1)
		assert.Equal(t, "f1.txt", files[0].Name)

		dirs, err := repo.GetDirectories(ctx, root.ID)
		assert.NoError(t, err)
		assert.Len(t, dirs, 1)
		assert.Equal(t, "child", dirs[0].Name)
	})

	t.Run("GetParentDirectory", func(t *testing.T) {
		root := &domain.NDirectory{Name: "root", FullPath: "root"}
		repo.Add(ctx, root)
		file := &domain.NFile{Name: "f2.txt", Hash: "f2_hash"}
		db.Create(file)
		db.Create(&domain.NFileNode{NDirectoryID: root.ID, NFileID: file.ID, Name: "f2.txt"})

		parent, err := repo.GetParentDirectory(ctx, file.ID)
		assert.NoError(t, err)
		assert.NotNil(t, parent)
		assert.Equal(t, "root", parent.Name)
	})
}

func TestNTagRepository(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormNTagRepository(db)
	ctx := context.Background()

	t.Run("Add and GetAllNames", func(t *testing.T) {
		repo.Add(ctx, &domain.NTag{Name: "A"})
		repo.Add(ctx, &domain.NTag{Name: "B"})

		names, err := repo.GetAllNames(ctx)
		assert.NoError(t, err)
		assert.Len(t, names, 2)
		assert.Contains(t, names, "A")
		assert.Contains(t, names, "B")
	})

	t.Run("GetByName", func(t *testing.T) {
		repo.Add(ctx, &domain.NTag{Name: "Unique"})
		tag, err := repo.GetByName(ctx, "Unique")
		assert.NoError(t, err)
		assert.NotNil(t, tag)
		assert.Equal(t, "Unique", tag.Name)
	})
}

func TestNScanRepository(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormNScanRepository(db)
	ctx := context.Background()

	t.Run("Add and GetByID", func(t *testing.T) {
		scan := &domain.NScan{FileCount: 10, Status: domain.NScanStatusCompleted}
		err := repo.Add(ctx, scan)
		assert.NoError(t, err)
		
		fetched, err := repo.GetByID(ctx, scan.ID)
		assert.NoError(t, err)
		assert.Equal(t, int64(10), fetched.FileCount)
	})
}

func TestNDirectoryRepositoryExtra(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormNDirectoryRepository(db)
	ctx := context.Background()

	t.Run("FileNodeExists and GetByName", func(t *testing.T) {
		dir := &domain.NDirectory{Name: "Parent", FullPath: "Parent"}
		repo.Add(ctx, dir)
		
		db.Create(&domain.NFileNode{NDirectoryID: dir.ID, Name: "node1"})
		
		exists, _ := repo.FileNodeExists(ctx, dir.ID, "node1")
		assert.True(t, exists)

		exists, _ = repo.FileNodeExists(ctx, dir.ID, "ghost")
		assert.False(t, exists)

		child := &domain.NDirectory{Name: "Child", ParentDirectoryID: dir.ID}
		repo.Add(ctx, child)
		
		fetched, _ := repo.GetByName(ctx, dir.ID, "Child")
		assert.NotNil(t, fetched)
		assert.Equal(t, child.ID, fetched.ID)
	})

	t.Run("GetSourceDirectory", func(t *testing.T) {
		db.Create(&domain.NDirectory{Name: "Source1", IsSource: true})
		src, err := repo.GetSourceDirectory(ctx, "Source1")
		assert.NoError(t, err)
		assert.NotNil(t, src)
		assert.True(t, src.IsSource)
	})

	t.Run("Search", func(t *testing.T) {
		now := time.Now()
		db.Create(&domain.NDirectory{Name: "SearchMe", DateModified: now})
		res, total, _ := repo.Search(ctx, "Search", now.Add(-1*time.Hour), now.Add(1*time.Hour), 1, 10)
		assert.Equal(t, int64(1), total)
		assert.Equal(t, "SearchMe", res[0].Name)
	})
}

func TestNFileRepositoryExtra(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormNFileRepository(db)
	ctx := context.Background()

	t.Run("CountByExtensions and GetPaged", func(t *testing.T) {
		db.Create(&domain.NFile{Name: "1.jpg", Extension: ".jpg", Hash: "h1"})
		db.Create(&domain.NFile{Name: "2.png", Extension: ".png", Hash: "h2"})

		count, _ := repo.CountByExtensions(ctx, []string{".jpg"})
		assert.Equal(t, int64(1), count)

		files, _ := repo.GetByExtensionsPaged(ctx, []string{".jpg", ".png"}, 1, 10)
		assert.Len(t, files, 2)
	})
}

func TestUnitOfWork(t *testing.T) {
	db := setupTestDB(t)
	uow := NewGormUnitOfWork(db)
	ctx := context.Background()

	t.Run("Transaction Rollback", func(t *testing.T) {
		err := uow.Transaction(ctx, func(txUow domain.UnitOfWork) error {
			txUow.Files().Add(ctx, &domain.NFile{Name: "tx_file.txt", Hash: "tx_hash"})
			return gorm.ErrInvalidData // force rollback
		})
		assert.Error(t, err)

		exists, _ := uow.Files().Exists(ctx, "tx_hash")
		assert.False(t, exists)
	})

	t.Run("Transaction Success", func(t *testing.T) {
		err := uow.Transaction(ctx, func(txUow domain.UnitOfWork) error {
			return txUow.Files().Add(ctx, &domain.NFile{Name: "success_file.txt", Hash: "success_hash"})
		})
		assert.NoError(t, err)

		exists, _ := uow.Files().Exists(ctx, "success_hash")
		assert.True(t, exists)
	})
}
