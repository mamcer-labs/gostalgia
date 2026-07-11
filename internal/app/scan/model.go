package scan

import (
	"time"

	"github.com/mamcer/gostalgia/internal/domain"
)

// ScanResult holds the results of a directory scan.
type ScanResult struct {
	RootDirectory *DirectoryScan
	Directories   []*DirectoryScan
	Files         []*FileScan
}

// DirectoryScan represents a directory found during a scan.
type DirectoryScan struct {
	Name         string
	DateModified time.Time
	ScanPath     string
	FinalPath    string
	Size         int64
	Directories  []*DirectoryScan
	Files        []*FileScan
}

// FileScan represents a file found during a scan.
type FileScan struct {
	Name               string
	Extension          string
	ScanPath           string
	FinalPath          string
	DateModified       time.Time
	Size               int64
	Hash               string
	IsExistingUniverse bool
	IsExistingInternal bool
	ExistingNFile      *domain.NFile
}
