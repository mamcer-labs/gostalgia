package repository

import (
	"context"
	"errors"

	"github.com/mamcer/gostalgia/internal/domain"
	"gorm.io/gorm"
)

type GormNScanRepository struct {
	*GormRepository[domain.NScan, int64]
}

func NewGormNScanRepository(db *gorm.DB) *GormNScanRepository {
	return &GormNScanRepository{
		GormRepository: NewGormRepository[domain.NScan, int64](db),
	}
}

func (r *GormNScanRepository) GetByID(ctx context.Context, id int64) (*domain.NScan, error) {
	var scan domain.NScan
	if err := r.db.WithContext(ctx).First(&scan, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &scan, nil
}

func (r *GormNScanRepository) GetRecent(ctx context.Context, limit int) ([]*domain.NScan, error) {
	var scans []*domain.NScan
	err := r.db.WithContext(ctx).Order("date_created DESC").Limit(limit).Find(&scans).Error
	return scans, err
}
