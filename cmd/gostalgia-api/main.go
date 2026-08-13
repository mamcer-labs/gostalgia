package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mamcer/gostalgia/internal/api"
	"github.com/mamcer/gostalgia/internal/app/directory"
	"github.com/mamcer/gostalgia/internal/app/file"
	"github.com/mamcer/gostalgia/internal/app/scan"
	"github.com/mamcer/gostalgia/internal/app/search"
	"github.com/mamcer/gostalgia/internal/app/tag"
	"github.com/mamcer/gostalgia/internal/config"
	"github.com/mamcer/gostalgia/internal/infra/database"
	"github.com/mamcer/gostalgia/internal/infra/filesystem"
	"github.com/mamcer/gostalgia/internal/infra/logging"
	"github.com/mamcer/gostalgia/internal/infra/repository"
	"github.com/mamcer/gostalgia/internal/infra/tracing"
	"github.com/patrickmn/go-cache"
)

type container struct {
	cfg              *config.Config
	uow              *repository.GormUnitOfWork
	cache            *cache.Cache
	tagService       *tag.TagService
	fileService      *file.FileService
	directoryService *directory.DirectoryService
	searchService    *search.SearchService
	scanService      *scan.ScanService
}

func (c *container) init() {
	// Initialize structured logger
	logger := slog.New(logging.NewTraceHandler(slog.NewJSONHandler(os.Stdout, nil)))
	slog.SetDefault(logger)

	c.cfg = config.LoadConfig()
	database.RunMigrations(c.cfg)
	db := database.NewMySQLDB(c.cfg)
	c.uow = repository.NewGormUnitOfWork(db)
	c.cache = cache.New(5*time.Minute, 10*time.Minute)

	fs := filesystem.NewRealFileSystem()

	c.tagService = tag.NewTagService(c.uow, c.cache)
	c.fileService = file.NewFileService(c.uow, c.cache)
	c.directoryService = directory.NewDirectoryService(c.uow, c.tagService)
	c.searchService = search.NewSearchService(c.uow, c.fileService, c.directoryService, c.tagService)
	c.scanService = scan.NewScanService(c.uow, fs)
}

// @title           Nostalgia API
// @version         1.0
// @description     API para gestionar archivos, directorios y etiquetas en Nostalgia.
// @host            localhost:8080
// @BasePath        /v1

func main() {
	c := &container{}
	c.init()

	shutdownTracing, err := tracing.Init(context.Background(), c.cfg.OTEL_EXPORTER_OTLP_ENDPOINT)
	if err != nil {
		slog.Error("Failed to initialize tracing, continuing without it", "error", err)
	} else {
		defer func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := shutdownTracing(ctx); err != nil {
				slog.Error("Failed to shut down tracing", "error", err)
			}
		}()
	}

	routerConfig := api.RouterConfig{
		FileService:      c.fileService,
		TagService:       c.tagService,
		DirectoryService: c.directoryService,
		SearchService:    c.searchService,
		ScanService:      c.scanService,
		Version:          "1.0.0",
	}

	r := api.NewRouter(routerConfig)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	go func() {
		slog.Info("Nostalgia API starting", "port", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	slog.Info("Server exiting")
}
