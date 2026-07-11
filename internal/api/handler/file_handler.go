package handler

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mamcer/gostalgia/internal/app/file"
	"github.com/mamcer/gostalgia/internal/domain"
)

type FileHandler struct {
	fileService *file.FileService
}

func NewFileHandler(fileService *file.FileService) *FileHandler {
	return &FileHandler{fileService: fileService}
}

// GetCount godoc
// @Summary      Contar archivos
// @Description  Devuelve el número total de archivos en la base de datos
// @Tags         files
// @Produce      json
// @Success      200  {object}  map[string]int64
// @Router       /files/count [get]
func (h *FileHandler) GetCount(c *gin.Context) {
	count, err := h.fileService.Count(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"count": count})
}

// GetByID godoc
// @Summary      Obtener archivo por ID
// @Description  Devuelve los detalles de un archivo específico
// @Tags         files
// @Produce      json
// @Param        id   path      int  true  "File ID"
// @Success      200  {object}  dto.NFileDto
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /files/{id} [get]
func (h *FileHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid file ID"})
		return
	}

	file, err := h.fileService.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"message": "File not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	if file == nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "File not found"})
		return
	}

	c.JSON(http.StatusOK, file)
}

// Search godoc
// @Summary      Buscar archivos
// @Description  Busca archivos por nombre, fecha y tipo
// @Tags         files
// @Produce      json
// @Param        contains  query     string  true   "Texto a buscar en el nombre"
// @Param        after     query     string  false  "Fecha mínima (YYYY-MM-DD)"
// @Param        before    query     string  false  "Fecha máxima (YYYY-MM-DD)"
// @Param        type      query     string  false  "Tipo de archivo (any, image, video)" default(any)
// @Param        page      query     int     false  "Número de página" default(1)
// @Param        per_page  query     int     false  "Resultados por página" default(50)
// @Success      200  {object}  dto.NFileSearchResultDto
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /files/search [get]
func (h *FileHandler) Search(c *gin.Context) {
	contains := c.Query("contains")
	if contains == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Contains parameter is required."})
		return
	}

	afterStr := c.Query("after")
	beforeStr := c.Query("before")
	fileType := c.DefaultQuery("type", "any")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "50"))

	var after, before *time.Time
	if afterStr != "" {
		if t, err := time.Parse("2006-01-02", afterStr); err == nil {
			after = &t
		}
	}
	if beforeStr != "" {
		if t, err := time.Parse("2006-01-02", beforeStr); err == nil {
			before = &t
		}
	}

	result, err := h.fileService.Search(c.Request.Context(), contains, after, before, fileType, page, perPage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	if result.Total == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "No results found."})
		return
	}

	c.JSON(http.StatusOK, result)
}
