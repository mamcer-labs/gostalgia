package repository

import (
	"context"
	"errors"
	"time"

	"github.com/mamcer/gostalgia/internal/domain"
	"gorm.io/gorm"
)

type GormNDirectoryRepository struct {
	*GormRepository[domain.NDirectory, int64]
}

func NewGormNDirectoryRepository(db *gorm.DB) *GormNDirectoryRepository {
	return &GormNDirectoryRepository{
		GormRepository: NewGormRepository[domain.NDirectory, int64](db),
	}
}

func (r *GormNDirectoryRepository) GetByID(ctx context.Context, id int64) (*domain.NDirectory, error) {
	var directory domain.NDirectory
	err := r.db.WithContext(ctx).
		Preload("Tags").
		Preload("FileNodes.File.Tags").
		First(&directory, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &directory, nil
}

func (r *GormNDirectoryRepository) GetSourceDirectory(ctx context.Context, name string) (*domain.NDirectory, error) {
	var directory domain.NDirectory
	err := r.db.WithContext(ctx).Where("name = ? AND is_source = ?", name, true).First(&directory).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &directory, nil
}

func (r *GormNDirectoryRepository) GetByName(ctx context.Context, parentID int64, name string) (*domain.NDirectory, error) {
	var directory domain.NDirectory
	err := r.db.WithContext(ctx).Where("parent_directory_id = ? AND name = ?", parentID, name).First(&directory).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &directory, nil
}

func (r *GormNDirectoryRepository) FileNodeExists(ctx context.Context, directoryID int64, name string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.NFileNode{}).
		Where("ndirectory_id = ? AND name = ?", directoryID, name).
		Count(&count).Error
	return count > 0, err
}

func (r *GormNDirectoryRepository) GetFiles(ctx context.Context, id int64) ([]*domain.NFile, error) {
	var files []*domain.NFile
	err := r.db.WithContext(ctx).
		Joins("JOIN nfilenode ON nfilenode.nfile_id = nfile.id").
		Where("nfilenode.ndirectory_id = ?", id).
		Preload("Tags").
		Find(&files).Error
	return files, err
}

func (r *GormNDirectoryRepository) GetDirectories(ctx context.Context, id int64) ([]*domain.NDirectory, error) {
	var directories []*domain.NDirectory
	err := r.db.WithContext(ctx).Where("parent_directory_id = ?", id).Find(&directories).Error
	return directories, err
}

func (r *GormNDirectoryRepository) GetParentDirectory(ctx context.Context, id int64) (*domain.NDirectory, error) {
	var directory domain.NDirectory
	err := r.db.WithContext(ctx).
		Joins("JOIN nfilenode ON nfilenode.ndirectory_id = ndirectory.id").
		Where("nfilenode.nfile_id = ?", id).
		First(&directory).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &directory, nil
}

func (r *GormNDirectoryRepository) Search(ctx context.Context, contains string, after, before time.Time, page, perPage int) ([]*domain.NDirectory, int64, error) {
	var directories []*domain.NDirectory
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.NDirectory{})

	if contains != "" {
		query = query.Where("name LIKE ?", "%"+contains+"%")
	}
	if !after.IsZero() {
		query = query.Where("date_modified >= ?", after)
	}
	if !before.IsZero() {
		query = query.Where("date_modified <= ?", before)
	}

	query.Count(&total)

	offset := (page - 1) * perPage
	err := query.Offset(offset).Limit(perPage).Find(&directories).Error

	return directories, total, err
}
