package repository

import (
	"context"
	"errors"
	"time"

	"github.com/mamcer/gostalgia/internal/domain"
	"gorm.io/gorm"
)

type GormNFileRepository struct {
	*GormRepository[domain.NFile, int64]
}

func NewGormNFileRepository(db *gorm.DB) *GormNFileRepository {
	return &GormNFileRepository{
		GormRepository: NewGormRepository[domain.NFile, int64](db),
	}
}

func (r *GormNFileRepository) GetByID(ctx context.Context, id int64) (*domain.NFile, error) {
	var file domain.NFile
	if err := r.db.WithContext(ctx).Preload("Tags").First(&file, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &file, nil
}

func (r *GormNFileRepository) Exists(ctx context.Context, hash string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.NFile{}).Where("hash = ?", hash).Count(&count).Error
	return count > 0, err
}

func (r *GormNFileRepository) GetByHash(ctx context.Context, hash string) (*domain.NFile, error) {
	var file domain.NFile
	if err := r.db.WithContext(ctx).Preload("Tags").Where("hash = ?", hash).First(&file).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &file, nil
}

func (r *GormNFileRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.NFile{}).Count(&count).Error
	return count, err
}

func (r *GormNFileRepository) Search(ctx context.Context, contains string, after, before time.Time, fileType string, extensions []string, page, perPage int) ([]*domain.NFile, int, error) {
	var files []*domain.NFile
	query := r.db.WithContext(ctx).Model(&domain.NFile{})

	if contains != "" {
		query = query.Where("nfile.name LIKE ?", "%"+contains+"%")
	}
	if !after.IsZero() {
		query = query.Where("nfile.date_modified >= ?", after)
	}
	if !before.IsZero() {
		query = query.Where("nfile.date_modified <= ?", before)
	}
	if len(extensions) > 0 {
		query = query.Where("nfile.extension IN ?", extensions)
	}

	var total int64
	query.Count(&total)

	offset := (page - 1) * perPage
	err := query.Preload("Tags").Offset(offset).Limit(perPage).Find(&files).Error

	return files, int(total), err
}

func (r *GormNFileRepository) SearchByTag(ctx context.Context, tagName string, page, perPage int) ([]*domain.NFile, int, error) {
	var files []*domain.NFile
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.NFile{}).
		Joins("JOIN ntag_nfile ON ntag_nfile.nfile_id = nfile.id").
		Joins("JOIN ntag ON ntag.id = ntag_nfile.ntag_id").
		Where("ntag.name = ?", tagName)

	query.Count(&total)

	offset := (page - 1) * perPage
	err := query.Preload("Tags").Offset(offset).Limit(perPage).Find(&files).Error

	return files, int(total), err
}

func (r *GormNFileRepository) CountByExtensions(ctx context.Context, extensions []string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.NFile{}).Where("extension IN ?", extensions).Count(&count).Error
	return count, err
}

func (r *GormNFileRepository) GetByExtensionsPaged(ctx context.Context, extensions []string, page, perPage int) ([]*domain.NFile, error) {
	var files []*domain.NFile
	offset := (page - 1) * perPage
	err := r.db.WithContext(ctx).Where("extension IN ?", extensions).Offset(offset).Limit(perPage).Find(&files).Error
	return files, err
}
