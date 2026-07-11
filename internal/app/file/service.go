package file

import (
	"context"
	"time"

	"github.com/mamcer/gostalgia/internal/app/dto"
	"github.com/mamcer/gostalgia/internal/app/util"
	"github.com/mamcer/gostalgia/internal/domain"
	"github.com/patrickmn/go-cache"
)

type FileService struct {
	uow   domain.UnitOfWork
	cache *cache.Cache
}

const fileCountCacheKey = "file_count"

func NewFileService(uow domain.UnitOfWork, c *cache.Cache) *FileService {
	return &FileService{
		uow:   uow,
		cache: c,
	}
}

func (s *FileService) Count(ctx context.Context) (int64, error) {
	if s.cache != nil {
		if val, found := s.cache.Get(fileCountCacheKey); found {
			return val.(int64), nil
		}
	}

	count, err := s.uow.Files().Count(ctx)
	if err != nil {
		return 0, err
	}

	if s.cache != nil {
		s.cache.Set(fileCountCacheKey, count, 1*time.Minute)
	}

	return count, nil
}

func (s *FileService) GetByID(ctx context.Context, id int64) (*dto.NFileDto, error) {
	file, err := s.uow.Files().GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if file == nil {
		return nil, nil
	}

	return s.mapToDto(ctx, file), nil
}

func (s *FileService) Search(ctx context.Context, contains string, after, before *time.Time, fileType string, page, perPage int) (*dto.NFileSearchResultDto, error) {
	if after == nil {
		t := time.Date(1000, 1, 1, 0, 0, 0, 0, time.UTC)
		after = &t
	}
	if before == nil {
		t := time.Date(9999, 12, 31, 0, 0, 0, 0, time.UTC)
		before = &t
	}
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 50
	}

	var extensions []string
	if fileType != "any" {
		switch fileType {
		case "image":
			extensions = []string{".jpeg", ".png", ".jpg", ".bmp"}
		case "doc":
			extensions = []string{".doc", ".docx", ".odt", ".pdf"}
		case "sheet":
			extensions = []string{".xls", ".xlsx", ".ods"}
		case "audio":
			extensions = []string{".mp3", ".ogg", ".wma", ".arm", ".wav"}
		case "video":
			extensions = []string{".mp4", ".mkv", ".avi", ".wmv"}
		case "zip":
			extensions = []string{".zip", ".rar", ".7z", ".gz"}
		}
	}

	files, total, err := s.uow.Files().Search(ctx, contains, *after, *before, fileType, extensions, page, perPage)
	if err != nil {
		return nil, err
	}

	resultDto := make([]*dto.NFileDto, len(files))
	for i, f := range files {
		resultDto[i] = s.mapToDto(ctx, f)
	}

	return &dto.NFileSearchResultDto{
		Result:   resultDto,
		Page:     page,
		PerPage:  perPage,
		Contains: contains,
		Total:    total,
	}, nil
}

func (s *FileService) SearchByTag(ctx context.Context, tagName string, page, perPage int) (*dto.NFileSearchResultDto, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 50
	}

	files, total, err := s.uow.Files().SearchByTag(ctx, tagName, page, perPage)
	if err != nil {
		return nil, err
	}

	resultDto := make([]*dto.NFileDto, len(files))
	for i, f := range files {
		resultDto[i] = s.mapToDto(ctx, f)
	}

	return &dto.NFileSearchResultDto{
		Result:   resultDto,
		Page:     page,
		PerPage:  perPage,
		Contains: tagName,
		Total:    total,
	}, nil
}

func (s *FileService) mapToDto(ctx context.Context, file *domain.NFile) *dto.NFileDto {
	tags := make([]string, len(file.Tags))
	for i, t := range file.Tags {
		tags[i] = t.Name
	}

	d := &dto.NFileDto{
		ID:           file.ID,
		Name:         file.Name,
		Extension:    file.Extension,
		Path:         file.Path,
		DateModified: file.DateModified.Format("02-01-2006"),
		Size:         util.GetHumanReadableFileSize(file.Size),
		SizeRaw:      file.Size,
		Hash:         file.Hash,
		Tags:         tags,
	}

	parent, _ := s.uow.Directories().GetParentDirectory(ctx, file.ID)
	if parent != nil {
		d.ParentDirectory = &dto.NDirectoryDto{
			ID:   parent.ID,
			Name: parent.Name,
		}
	}

	return d
}
