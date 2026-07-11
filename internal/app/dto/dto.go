package dto

type NFileDto struct {
	ID              int64          `json:"id"`
	Name            string         `json:"name" binding:"required"`
	Extension       string         `json:"extension"`
	Path            string         `json:"path" binding:"required"`
	DateModified    string         `json:"date_modified"`
	Size            string         `json:"size"`
	SizeRaw         int64          `json:"size_raw"`
	Hash            string         `json:"hash" binding:"required"`
	Tags            []string       `json:"tags"`
	ParentDirectory *NDirectoryDto `json:"parent_directory,omitempty"`
}

type NDirectoryDto struct {
	ID                int64            `json:"id"`
	ParentDirectoryID int64            `json:"parent_directory_id,omitempty"`
	Name              string           `json:"name" binding:"required"`
	FullPath          string           `json:"full_path" binding:"required"`
	DateModified      string           `json:"date_modified"`
	Size              string           `json:"size"`
	SizeRaw           int64            `json:"size_raw"`
	FileCount         int64            `json:"file_count"`
	DirectoryCount    int64            `json:"directory_count"`
	Tags              []string         `json:"tags"`
	Files             []*NFileDto      `json:"files,omitempty"`
	Directories       []*NDirectoryDto `json:"directories,omitempty"`
}

type NFileSearchResultDto struct {
	Result   []*NFileDto `json:"result"`
	Page     int         `json:"page"`
	PerPage  int         `json:"per_page"`
	Contains string      `json:"contains"`
	Total    int         `json:"total"`
}

type NDirectorySearchResultDto struct {
	Result   []*NDirectoryDto `json:"result"`
	Page     int              `json:"page"`
	PerPage  int              `json:"per_page"`
	Contains string           `json:"contains"`
	Total    int64            `json:"total"`
}

type UnifiedSearchResultDto struct {
	Files       []*NFileDto      `json:"files"`
	Directories []*NDirectoryDto `json:"directories"`
	Tags        []string         `json:"tags"`
	Total       int64            `json:"total"`
}
