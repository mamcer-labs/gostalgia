package tag

import (
	"context"
	"errors"
	"sort"
	"time"

	"github.com/mamcer/gostalgia/internal/domain"
	"github.com/patrickmn/go-cache"
)

type TagService struct {
	uow   domain.UnitOfWork
	cache *cache.Cache
}

const tagsCacheKey = "all_tags"

func NewTagService(uow domain.UnitOfWork, c *cache.Cache) *TagService {
	return &TagService{
		uow:   uow,
		cache: c,
	}
}

func (s *TagService) GetAllTags(ctx context.Context) ([]string, error) {
	if s.cache != nil {
		if val, found := s.cache.Get(tagsCacheKey); found {
			return val.([]string), nil
		}
	}

	names, err := s.uow.Tags().GetAllNames(ctx)
	if err != nil {
		return nil, err
	}
	sort.Strings(names)

	if s.cache != nil {
		s.cache.Set(tagsCacheKey, names, 5*time.Minute)
	}

	return names, nil
}

func (s *TagService) GetPopularTags(ctx context.Context, limit int) ([]string, error) {
	return s.uow.Tags().GetPopular(ctx, limit)
}

func (s *TagService) GetByName(ctx context.Context, name string) (*domain.NTag, error) {
	tag, err := s.uow.Tags().GetByName(ctx, name)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return tag, nil
}

func (s *TagService) Add(ctx context.Context, tag *domain.NTag) error {
	err := s.uow.Tags().Add(ctx, tag)
	if err != nil {
		return err
	}
	err = s.uow.Complete(ctx)
	if err == nil && s.cache != nil {
		s.cache.Delete(tagsCacheKey)
	}
	return err
}
