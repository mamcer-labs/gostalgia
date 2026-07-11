package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/mamcer/gostalgia/internal/app/directory"
	"github.com/mamcer/gostalgia/internal/app/file"
	"github.com/mamcer/gostalgia/internal/app/tag"
	"github.com/mamcer/gostalgia/internal/domain"
	"github.com/mamcer/gostalgia/internal/infra/repository"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestNewRouter(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect database: %v", err)
	}
	err = db.AutoMigrate(&domain.NTag{}, &domain.NFile{}, &domain.NDirectory{}, &domain.NScan{}, &domain.NFileNode{})
	if err != nil {
		t.Fatalf("failed to migrate database: %v", err)
	}
	uow := repository.NewGormUnitOfWork(db)

	tagSvc := tag.NewTagService(uow, nil)
	fileSvc := file.NewFileService(uow, nil)
	dirSvc := directory.NewDirectoryService(uow, tagSvc)

	r := NewRouter(RouterConfig{
		FileService:      fileSvc,
		TagService:       tagSvc,
		DirectoryService: dirSvc,
		Version:          "1.0.0",
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/health", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
