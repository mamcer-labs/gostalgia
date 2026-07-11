package handler

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect database: %v", err)
	}
	err = db.AutoMigrate(&domain.NTag{}, &domain.NFile{}, &domain.NDirectory{}, &domain.NScan{}, &domain.NFileNode{})
	if err != nil {
		t.Fatalf("failed to migrate database: %v", err)
	}
	return db
}

func TestHandlers(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestDB(t)
	uow := repository.NewGormUnitOfWork(db)

	tagSvc := tag.NewTagService(uow, nil)
	fileSvc := file.NewFileService(uow, nil)
	dirSvc := directory.NewDirectoryService(uow, tagSvc)

	healthHandler := NewHealthHandler("1.0.0")
	fileHandler := NewFileHandler(fileSvc)
	tagHandler := NewTagHandler(tagSvc, fileSvc, dirSvc)
	directoryHandler := NewDirectoryHandler(dirSvc)

	t.Run("Health", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/health", nil)
		healthHandler.Health(c)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("File Handlers", func(t *testing.T) {
		f := &domain.NFile{Name: "handler_test.jpg", Hash: "h_handler", Extension: ".jpg", Path: "p1", DateModified: time.Now()}
		db.Create(f)

		t.Run("GetCount", func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/count", nil)
			fileHandler.GetCount(c)
			assert.Equal(t, http.StatusOK, w.Code)
		})

		t.Run("GetByID Success", func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/", nil)
			c.Params = []gin.Param{{Key: "id", Value: fmt.Sprint(f.ID)}}
			fileHandler.GetByID(c)
			assert.Equal(t, http.StatusOK, w.Code)
		})

		t.Run("GetByID Invalid ID", func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Params = []gin.Param{{Key: "id", Value: "abc"}}
			fileHandler.GetByID(c)
			assert.Equal(t, http.StatusBadRequest, w.Code)
		})

		t.Run("GetByID Not Found", func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/", nil)
			c.Params = []gin.Param{{Key: "id", Value: "9999"}}
			fileHandler.GetByID(c)
			assert.Equal(t, http.StatusNotFound, w.Code)
		})

		t.Run("Search Success", func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/search?contains=handler", nil)
			fileHandler.Search(c)
			assert.Equal(t, http.StatusOK, w.Code)
		})

		t.Run("Search No Contains", func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/search", nil)
			fileHandler.Search(c)
			assert.Equal(t, http.StatusBadRequest, w.Code)
		})

		t.Run("Search Success with Dates", func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/search?contains=handler&after=2020-01-01&before=2030-01-01", nil)
			fileHandler.Search(c)
			assert.Equal(t, http.StatusOK, w.Code)
		})
	})

	t.Run("Tag Handlers", func(t *testing.T) {
		t.Run("GetAllTags", func(t *testing.T) {
			db.Create(&domain.NTag{Name: "TagH1"})
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/tags", nil)
			tagHandler.GetAllTags(c)
			assert.Equal(t, http.StatusOK, w.Code)
		})

		t.Run("AddTagToDirectory No Name", func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Params = []gin.Param{{Key: "tagName", Value: ""}, {Key: "directoryId", Value: "1"}}
			tagHandler.AddTagToDirectory(c)
			assert.Equal(t, http.StatusBadRequest, w.Code)
		})

		t.Run("SearchFilesByTag No Tag", func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/search", nil)
			tagHandler.SearchFilesByTag(c)
			assert.Equal(t, http.StatusBadRequest, w.Code)
		})

		t.Run("SearchFilesByTag Not Found", func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/search?tag=NonExistent", nil)
			tagHandler.SearchFilesByTag(c)
			assert.Equal(t, http.StatusNotFound, w.Code)
		})
	})

	t.Run("Directory Handlers", func(t *testing.T) {
		d := &domain.NDirectory{Name: "DirH1", FullPath: "DirH1", DateModified: time.Now()}
		db.Create(d)

		t.Run("GetByID Success", func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/", nil)
			c.Params = []gin.Param{{Key: "id", Value: fmt.Sprint(d.ID)}}
			directoryHandler.GetByID(c)
			assert.Equal(t, http.StatusOK, w.Code)
		})

		t.Run("GetFiles Success", func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/files", nil)
			c.Params = []gin.Param{{Key: "id", Value: fmt.Sprint(d.ID)}}
			directoryHandler.GetFiles(c)
			assert.Equal(t, http.StatusOK, w.Code)
		})

		t.Run("GetDirectories Success", func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/directories", nil)
			c.Params = []gin.Param{{Key: "id", Value: fmt.Sprint(d.ID)}}
			directoryHandler.GetDirectories(c)
			assert.Equal(t, http.StatusOK, w.Code)
		})

		t.Run("Search Success", func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/search?contains=DirH", nil)
			directoryHandler.Search(c)
			assert.Equal(t, http.StatusOK, w.Code)
		})

		t.Run("Search No Contains", func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/search", nil)
			directoryHandler.Search(c)
			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	})
}
