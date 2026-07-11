package repository

import (
	"context"
	"errors"

	"github.com/mamcer/gostalgia/internal/domain"
	"gorm.io/gorm"
)

type GormRepository[T any, K any] struct {
	db *gorm.DB
}

func NewGormRepository[T any, K any](db *gorm.DB) *GormRepository[T, K] {
	return &GormRepository[T, K]{db: db}
}

func (r *GormRepository[T, K]) GetByID(ctx context.Context, id K) (*T, error) {
	var entity T
	if err := r.db.WithContext(ctx).First(&entity, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &entity, nil
}

func (r *GormRepository[T, K]) Add(ctx context.Context, entity *T) error {
	return r.db.WithContext(ctx).Create(entity).Error
}

func (r *GormRepository[T, K]) Update(ctx context.Context, entity *T) error {
	return r.db.WithContext(ctx).Save(entity).Error
}

func (r *GormRepository[T, K]) Delete(ctx context.Context, entity *T) error {
	return r.db.WithContext(ctx).Delete(entity).Error
}

func (r *GormRepository[T, K]) AddRange(ctx context.Context, entities []*T) error {
	return r.db.WithContext(ctx).Create(entities).Error
}
