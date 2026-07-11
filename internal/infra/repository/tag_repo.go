package repository

import (
	"context"
	"errors"

	"github.com/mamcer/gostalgia/internal/domain"
	"gorm.io/gorm"
)

type GormNTagRepository struct {
	*GormRepository[domain.NTag, int64]
}

func NewGormNTagRepository(db *gorm.DB) *GormNTagRepository {
	return &GormNTagRepository{
		GormRepository: NewGormRepository[domain.NTag, int64](db),
	}
}

func (r *GormNTagRepository) GetAllNames(ctx context.Context) ([]string, error) {
	var names []string
	err := r.db.WithContext(ctx).Model(&domain.NTag{}).Pluck("name", &names).Error
	return names, err
}

func (r *GormNTagRepository) GetByName(ctx context.Context, name string) (*domain.NTag, error) {
	var tag domain.NTag
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&tag).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &tag, nil
}

func (r *GormNTagRepository) Search(ctx context.Context, query string) ([]string, error) {
	var names []string
	err := r.db.WithContext(ctx).Model(&domain.NTag{}).
		Where("name LIKE ?", "%"+query+"%").
		Pluck("name", &names).Error
	return names, err
}

func (r *GormNTagRepository) GetPopular(ctx context.Context, limit int) ([]string, error) {
	var names []string
	err := r.db.WithContext(ctx).Table("ntag").
		Select("ntag.name, COUNT(ntag_nfile.nfile_id) as count").
		Joins("LEFT JOIN ntag_nfile ON ntag_nfile.ntag_id = ntag.id").
		Group("ntag.name").
		Order("count DESC").
		Limit(limit).
		Pluck("name", &names).Error
	return names, err
}
