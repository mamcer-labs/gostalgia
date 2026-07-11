package handler

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mamcer/gostalgia/internal/app/directory"
	"github.com/mamcer/gostalgia/internal/domain"
)

type DirectoryHandler struct {
	directoryService *directory.DirectoryService
}

func NewDirectoryHandler(directoryService *directory.DirectoryService) *DirectoryHandler {
	return &DirectoryHandler{directoryService: directoryService}
}

// GetByID godoc
// @Summary      Obtener directorio por ID
// @Description  Devuelve los detalles de un directorio específico
// @Tags         directories
// @Produce      json
// @Param        id   path      int  true  "Directory ID"
// @Success      200  {object}  dto.NDirectoryDto
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /directories/{id} [get]
func (h *DirectoryHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid directory ID"})
		return
	}

	dir, err := h.directoryService.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"message": "Directory not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	if dir == nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Directory not found"})
		return
	}

	c.JSON(http.StatusOK, dir)
}

// GetFiles godoc
// @Summary      Obtener archivos de un directorio
// @Description  Devuelve la lista de archivos dentro de un directorio
// @Tags         directories
// @Produce      json
// @Param        id   path      int  true  "Directory ID"
// @Success      200  {array}   dto.NFileDto
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /directories/{id}/files [get]
func (h *DirectoryHandler) GetFiles(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid directory ID"})
		return
	}

	files, err := h.directoryService.GetFiles(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	if files == nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Directory not found"})
		return
	}

	c.JSON(http.StatusOK, files)
}

// GetDirectories godoc
// @Summary      Obtener subdirectorios de un directorio
// @Description  Devuelve la lista de subdirectorios dentro de un directorio
// @Tags         directories
// @Produce      json
// @Param        id   path      int  true  "Directory ID"
// @Success      200  {array}   dto.NDirectoryDto
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /directories/{id}/directories [get]
func (h *DirectoryHandler) GetDirectories(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid directory ID"})
		return
	}

	dirs, err := h.directoryService.GetDirectories(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	if dirs == nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Directory not found"})
		return
	}

	c.JSON(http.StatusOK, dirs)
}

// Search godoc
// @Summary      Buscar directorios
// @Description  Busca directorios por nombre y fecha
// @Tags         directories
// @Produce      json
// @Param        contains  query     string  true   "Texto a buscar en el nombre"
// @Param        after     query     string  false  "Fecha mínima (YYYY-MM-DD)"
// @Param        before    query     string  false  "Fecha máxima (YYYY-MM-DD)"
// @Param        page      query     int     false  "Número de página" default(1)
// @Param        per_page  query     int     false  "Resultados por página" default(50)
// @Success      200  {object}  dto.NDirectorySearchResultDto
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /directories/search [get]
func (h *DirectoryHandler) Search(c *gin.Context) {
	contains := c.Query("contains")
	if contains == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Contains parameter is required."})
		return
	}

	afterStr := c.Query("after")
	beforeStr := c.Query("before")
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

	result, err := h.directoryService.Search(c.Request.Context(), contains, after, before, page, perPage)
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
