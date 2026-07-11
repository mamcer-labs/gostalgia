package domain

import (
	"context"
	"time"
)

// Repository is a generic base repository interface
type Repository[T any, K any] interface {
	GetByID(ctx context.Context, id K) (*T, error)
	Add(ctx context.Context, entity *T) error
	Update(ctx context.Context, entity *T) error
	Delete(ctx context.Context, entity *T) error
	AddRange(ctx context.Context, entities []*T) error
}

// NFileRepository defines specific operations for NFile
type NFileRepository interface {
	Repository[NFile, int64]
	Exists(ctx context.Context, hash string) (bool, error)
	GetByHash(ctx context.Context, hash string) (*NFile, error)
	Count(ctx context.Context) (int64, error)
	Search(ctx context.Context, contains string, after, before time.Time, fileType string, extensions []string, page, perPage int) ([]*NFile, int, error)
	SearchByTag(ctx context.Context, tagName string, page, perPage int) ([]*NFile, int, error)
	CountByExtensions(ctx context.Context, extensions []string) (int64, error)
	GetByExtensionsPaged(ctx context.Context, extensions []string, page, perPage int) ([]*NFile, error)
}

// NDirectoryRepository defines specific operations for NDirectory
type NDirectoryRepository interface {
	Repository[NDirectory, int64]
	GetSourceDirectory(ctx context.Context, name string) (*NDirectory, error)
	GetByName(ctx context.Context, parentID int64, name string) (*NDirectory, error)
	FileNodeExists(ctx context.Context, directoryID int64, name string) (bool, error)
	GetFiles(ctx context.Context, id int64) ([]*NFile, error)
	GetDirectories(ctx context.Context, id int64) ([]*NDirectory, error)
	GetParentDirectory(ctx context.Context, id int64) (*NDirectory, error)
	Search(ctx context.Context, contains string, after, before time.Time, page, perPage int) ([]*NDirectory, int64, error)
}

// NScanRepository defines specific operations for NScan
type NScanRepository interface {
	Repository[NScan, int64]
	GetRecent(ctx context.Context, limit int) ([]*NScan, error)
}

// NTagRepository defines specific operations for NTag
type NTagRepository interface {
	Repository[NTag, int64]
	GetAllNames(ctx context.Context) ([]string, error)
	GetByName(ctx context.Context, name string) (*NTag, error)
	Search(ctx context.Context, query string) ([]string, error)
	GetPopular(ctx context.Context, limit int) ([]string, error)
}

// UnitOfWork defines the interface for managing transactions and multiple repositories
type UnitOfWork interface {
	Files() NFileRepository
	Directories() NDirectoryRepository
	Scans() NScanRepository
	Tags() NTagRepository
	Complete(ctx context.Context) error
}
