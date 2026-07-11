package directory

import (
	"context"
	"fmt"
	"time"

	"github.com/mamcer/gostalgia/internal/app/dto"
	"github.com/mamcer/gostalgia/internal/app/tag"
	"github.com/mamcer/gostalgia/internal/app/util"
	"github.com/mamcer/gostalgia/internal/domain"
)

type DirectoryService struct {
	uow        domain.UnitOfWork
	tagService *tag.TagService
}

func NewDirectoryService(uow domain.UnitOfWork, tagService *tag.TagService) *DirectoryService {
	return &DirectoryService{
		uow:        uow,
		tagService: tagService,
	}
}

func (s *DirectoryService) GetByID(ctx context.Context, id int64) (*dto.NDirectoryDto, error) {
	dir, err := s.uow.Directories().GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if dir == nil {
		return nil, nil
	}

	d := s.mapToDto(dir)

	// Get child directories
	subDirs, err := s.uow.Directories().GetDirectories(ctx, id)
	if err == nil {
		d.Directories = make([]*dto.NDirectoryDto, len(subDirs))
		for i, sd := range subDirs {
			d.Directories[i] = &dto.NDirectoryDto{
				ID:             sd.ID,
				Name:           sd.Name,
				FullPath:       sd.FullPath,
				DateModified:   sd.DateModified.Format("02-01-2006"),
				Size:           util.GetHumanReadableFileSize(sd.Size),
				SizeRaw:        sd.Size,
				FileCount:      sd.FileCount,
				DirectoryCount: sd.DirectoryCount,
			}
		}
	}

	// Get files
	files, err := s.uow.Directories().GetFiles(ctx, id)
	if err == nil {
		d.Files = make([]*dto.NFileDto, len(files))
		for i, f := range files {
			tags := make([]string, len(f.Tags))
			for j, t := range f.Tags {
				tags[j] = t.Name
			}
			d.Files[i] = &dto.NFileDto{
				ID:           f.ID,
				Name:         f.Name,
				Extension:    f.Extension,
				Path:         f.Path,
				DateModified: f.DateModified.Format("02-01-2006"),
				Size:         util.GetHumanReadableFileSize(f.Size),
				SizeRaw:      f.Size,
				Hash:         f.Hash,
				Tags:         tags,
				ParentDirectory: &dto.NDirectoryDto{
					ID:   dir.ID,
					Name: dir.Name,
				},
			}
		}
	}

	return d, nil
}

func (s *DirectoryService) GetFiles(ctx context.Context, id int64) ([]*dto.NFileDto, error) {
	files, err := s.uow.Directories().GetFiles(ctx, id)
	if err != nil {
		return nil, err
	}

	result := make([]*dto.NFileDto, len(files))
	for i, f := range files {
		tags := make([]string, len(f.Tags))
		for j, t := range f.Tags {
			tags[j] = t.Name
		}

		d := &dto.NFileDto{
			ID:           f.ID,
			Name:         f.Name,
			Extension:    f.Extension,
			Path:         f.Path,
			DateModified: f.DateModified.Format("02-01-2006"),
			Size:         util.GetHumanReadableFileSize(f.Size),
			SizeRaw:      f.Size,
			Hash:         f.Hash,
			Tags:         tags,
		}

		parent, _ := s.uow.Directories().GetParentDirectory(ctx, f.ID)
		if parent != nil {
			d.ParentDirectory = &dto.NDirectoryDto{
				ID:   parent.ID,
				Name: parent.Name,
			}
		}
		result[i] = d
	}

	return result, nil
}

func (s *DirectoryService) GetDirectories(ctx context.Context, id int64) ([]*dto.NDirectoryDto, error) {
	dirs, err := s.uow.Directories().GetDirectories(ctx, id)
	if err != nil {
		return nil, err
	}

	result := make([]*dto.NDirectoryDto, len(dirs))
	for i, d := range dirs {
		result[i] = s.mapToDto(d)
	}

	return result, nil
}

func (s *DirectoryService) AddTagToDirectory(ctx context.Context, directoryID int64, tagName string) (bool, error) {
	dir, err := s.uow.Directories().GetByID(ctx, directoryID)
	if err != nil {
		return false, err
	}
	if dir == nil {
		return false, nil
	}

	if dir.IsSource {
		return false, fmt.Errorf("source directories cannot be tagged")
	}

	tagEntity, err := s.tagService.GetByName(ctx, tagName)
	if err != nil {
		return false, err
	}
	if tagEntity == nil {
		tagEntity = &domain.NTag{Name: tagName}
		err = s.tagService.Add(ctx, tagEntity)
		if err != nil {
			return false, err
		}
	}

	// Logic refactored to avoid saving the whole object tree with GORM
	// We should use more targeted updates if possible, but let's at least fix the nil pointers.

	hasTag := false
	for _, t := range dir.Tags {
		if t.Name == tagName {
			hasTag = true
			break
		}
	}
	if !hasTag {
		dir.Tags = append(dir.Tags, tagEntity)
	}

	for _, fn := range dir.FileNodes {
		if fn.File == nil {
			continue
		}
		fileHasTag := false
		for _, t := range fn.File.Tags {
			if t.Name == tagName {
				fileHasTag = true
				break
			}
		}
		if !fileHasTag {
			fn.File.Tags = append(fn.File.Tags, tagEntity)
		}
	}

	err = s.uow.Directories().Update(ctx, dir)
	if err != nil {
		return false, err
	}

	err = s.uow.Complete(ctx)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (s *DirectoryService) Search(ctx context.Context, contains string, after, before *time.Time, page, perPage int) (*dto.NDirectorySearchResultDto, error) {
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

	dirs, total, err := s.uow.Directories().Search(ctx, contains, *after, *before, page, perPage)
	if err != nil {
		return nil, err
	}

	resultDto := make([]*dto.NDirectoryDto, len(dirs))
	for i, d := range dirs {
		resultDto[i] = s.mapToDto(d)
	}

	return &dto.NDirectorySearchResultDto{
		Result:   resultDto,
		Page:     page,
		PerPage:  perPage,
		Contains: contains,
		Total:    total,
	}, nil
}

func (s *DirectoryService) mapToDto(dir *domain.NDirectory) *dto.NDirectoryDto {
	tags := make([]string, len(dir.Tags))
	for i, t := range dir.Tags {
		tags[i] = t.Name
	}

	return &dto.NDirectoryDto{
		ID:                dir.ID,
		ParentDirectoryID: dir.ParentDirectoryID,
		Name:              dir.Name,
		FullPath:          dir.FullPath,
		DateModified:      dir.DateModified.Format("02-01-2006"),
		Size:              util.GetHumanReadableFileSize(dir.Size),
		SizeRaw:           dir.Size,
		FileCount:         dir.FileCount,
		DirectoryCount:    dir.DirectoryCount,
		Tags:              tags,
	}
}
