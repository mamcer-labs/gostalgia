package search

import (
	"context"

	"github.com/mamcer/gostalgia/internal/app/directory"
	"github.com/mamcer/gostalgia/internal/app/dto"
	"github.com/mamcer/gostalgia/internal/app/file"
	"github.com/mamcer/gostalgia/internal/app/tag"
	"github.com/mamcer/gostalgia/internal/domain"
)

type SearchService struct {
	uow              domain.UnitOfWork
	fileService      *file.FileService
	directoryService *directory.DirectoryService
	tagService       *tag.TagService
}

func NewSearchService(uow domain.UnitOfWork, fileService *file.FileService, directoryService *directory.DirectoryService, tagService *tag.TagService) *SearchService {
	return &SearchService{
		uow:              uow,
		fileService:      fileService,
		directoryService: directoryService,
		tagService:       tagService,
	}
}

func (s *SearchService) UnifiedSearch(ctx context.Context, query string, limit int) (*dto.UnifiedSearchResultDto, error) {
	if limit <= 0 {
		limit = 10
	}

	// 1. Search Tags
	tags, err := s.uow.Tags().Search(ctx, query)
	if err != nil {
		return nil, err
	}

	// 2. Search Files
	fileRes, err := s.fileService.Search(ctx, query, nil, nil, "any", 1, limit)
	if err != nil {
		return nil, err
	}

	// 3. Search Directories
	dirRes, err := s.directoryService.Search(ctx, query, nil, nil, 1, limit)
	if err != nil {
		return nil, err
	}

	return &dto.UnifiedSearchResultDto{
		Files:       fileRes.Result,
		Directories: dirRes.Result,
		Tags:        tags,
		Total:       int64(len(fileRes.Result) + len(dirRes.Result) + len(tags)),
	}, nil
}
