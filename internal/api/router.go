package api

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/mamcer/gostalgia/internal/api/handler"
	"github.com/mamcer/gostalgia/internal/app/directory"
	"github.com/mamcer/gostalgia/internal/app/file"
	"github.com/mamcer/gostalgia/internal/app/scan"
	"github.com/mamcer/gostalgia/internal/app/search"
	"github.com/mamcer/gostalgia/internal/app/tag"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	_ "github.com/mamcer/gostalgia/docs"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type RouterConfig struct {
	FileService      *file.FileService
	TagService       *tag.TagService
	DirectoryService *directory.DirectoryService
	SearchService    *search.SearchService
	ScanService      *scan.ScanService
	Version          string
}

func NewRouter(cfg RouterConfig) *gin.Engine {
	r := gin.Default()

	// CORS
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	healthHandler := handler.NewHealthHandler(cfg.Version)
	fileHandler := handler.NewFileHandler(cfg.FileService)
	tagHandler := handler.NewTagHandler(cfg.TagService, cfg.FileService, cfg.DirectoryService)
	dirHandler := handler.NewDirectoryHandler(cfg.DirectoryService)
	searchHandler := handler.NewSearchHandler(cfg.SearchService)
	scanHandler := handler.NewScanHandler(cfg.ScanService)

	// Static Files
	r.Static("/media", os.Getenv("NOSTALGIA_HOME_PATH"))
	r.Static("/thumbs", os.Getenv("NOSTALGIA_THUMB_TARGET_PATH"))

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	v1 := r.Group("/v1")
	{
		v1.GET("/health", healthHandler.Health)
		v1.GET("/metrics", gin.WrapH(promhttp.Handler()))
		v1.GET("/search", searchHandler.UnifiedSearch)

		// Tags
		v1.GET("/tags", tagHandler.GetAllTags)
		v1.GET("/tags/popular", tagHandler.GetPopularTags)
		v1.POST("/tags/:tagName/directories/:directoryId", tagHandler.AddTagToDirectory)
		v1.GET("/tags/search", tagHandler.SearchFilesByTag)

		// Files
		v1.GET("/files/count", fileHandler.GetCount)
		v1.GET("/files/:id", fileHandler.GetByID)
		v1.GET("/files/search", fileHandler.Search)

		// Directories
		v1.GET("/directories/:id", dirHandler.GetByID)
		v1.GET("/directories/:id/files", dirHandler.GetFiles)
		v1.GET("/directories/:id/directories", dirHandler.GetDirectories)
		v1.GET("/directories/search", dirHandler.Search)

		// Scans
		v1.GET("/scans/recent", scanHandler.GetRecentScans)
	}

	return r
}
