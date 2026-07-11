package repository

import (
	"context"

	"github.com/mamcer/gostalgia/internal/domain"
	"gorm.io/gorm"
)

type GormUnitOfWork struct {
	db          *gorm.DB
	fileRepo    domain.NFileRepository
	dirRepo     domain.NDirectoryRepository
	scanRepo    domain.NScanRepository
	tagRepo     domain.NTagRepository
}

func NewGormUnitOfWork(db *gorm.DB) *GormUnitOfWork {
	return &GormUnitOfWork{
		db:       db,
		fileRepo: NewGormNFileRepository(db),
		dirRepo:  NewGormNDirectoryRepository(db),
		scanRepo: NewGormNScanRepository(db),
		tagRepo:  NewGormNTagRepository(db),
	}
}

func (u *GormUnitOfWork) Files() domain.NFileRepository {
	return u.fileRepo
}

func (u *GormUnitOfWork) Directories() domain.NDirectoryRepository {
	return u.dirRepo
}

func (u *GormUnitOfWork) Scans() domain.NScanRepository {
	return u.scanRepo
}

func (u *GormUnitOfWork) Tags() domain.NTagRepository {
	return u.tagRepo
}

func (u *GormUnitOfWork) Complete(ctx context.Context) error {
	// In GORM, if we are already in a transaction, this might be handled differently.
	// But for a simple implementation, if we are not manually starting transactions,
	// Complete might just be a no-op or a commit if we implemented manual TX.
	
	// If we want to support transactions properly in UoW:
	// We should probably have a Begin method that returns a new UoW with a TX db.
	return nil
}

// Transaction helper for UoW
func (u *GormUnitOfWork) Transaction(ctx context.Context, fn func(domain.UnitOfWork) error) error {
	return u.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txUow := &GormUnitOfWork{
			db:       tx,
			fileRepo: NewGormNFileRepository(tx),
			dirRepo:  NewGormNDirectoryRepository(tx),
			scanRepo: NewGormNScanRepository(tx),
			tagRepo:  NewGormNTagRepository(tx),
		}
		return fn(txUow)
	})
}
