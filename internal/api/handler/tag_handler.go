package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mamcer/gostalgia/internal/app/directory"
	"github.com/mamcer/gostalgia/internal/app/file"
	"github.com/mamcer/gostalgia/internal/app/tag"
)

type TagHandler struct {
	tagService       *tag.TagService
	fileService      *file.FileService
	directoryService *directory.DirectoryService
}

func NewTagHandler(tagService *tag.TagService, fileService *file.FileService, directoryService *directory.DirectoryService) *TagHandler {
	return &TagHandler{
		tagService:       tagService,
		fileService:      fileService,
		directoryService: directoryService,
	}
}

// GetAllTags godoc
// @Summary      Obtener todas las etiquetas
// @Description  Devuelve la lista completa de etiquetas
// @Tags         tags
// @Produce      json
// @Success      200  {array}   string
// @Failure      500  {object}  map[string]string
// @Router       /tags [get]
func (h *TagHandler) GetAllTags(c *gin.Context) {
	tags, err := h.tagService.GetAllTags(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, tags)
}

// GetPopularTags godoc
// @Summary      Obtener etiquetas populares
// @Description  Devuelve las etiquetas más utilizadas
// @Tags         tags
// @Produce      json
// @Success      200  {array}   string
// @Failure      500  {object}  map[string]string
// @Router       /tags/popular [get]
func (h *TagHandler) GetPopularTags(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	tags, err := h.tagService.GetPopularTags(c.Request.Context(), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, tags)
}

// AddTagToDirectory godoc
// @Summary      Añadir etiqueta a directorio
// @Description  Asocia una etiqueta a un directorio específico y a todos sus archivos
// @Tags         tags
// @Produce      json
// @Param        tagName      path      string  true  "Nombre de la etiqueta"
// @Param        directoryId  path      int     true  "ID del directorio"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /tags/{tagName}/directories/{directoryId} [post]
func (h *TagHandler) AddTagToDirectory(c *gin.Context) {
	tagName := c.Param("tagName")
	directoryIDStr := c.Param("directoryId")

	if tagName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "A tag name must be provided."})
		return
	}

	directoryID, err := strconv.ParseInt(directoryIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid directory ID"})
		return
	}

	success, err := h.directoryService.AddTagToDirectory(c.Request.Context(), directoryID, tagName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	if !success {
		c.JSON(http.StatusNotFound, gin.H{"message": "Directory not found"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Tag added to directory"})
}

// SearchFilesByTag godoc
// @Summary      Buscar archivos por etiqueta
// @Description  Busca archivos asociados a una etiqueta específica
// @Tags         tags
// @Produce      json
// @Param        tag       query     string  true   "Nombre de la etiqueta"
// @Param        page      query     int     false  "Número de página" default(1)
// @Param        per_page  query     int     false  "Resultados por página" default(50)
// @Success      200  {object}  dto.NFileSearchResultDto
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /tags/search [get]
func (h *TagHandler) SearchFilesByTag(c *gin.Context) {
	tagName := c.Query("tag")
	if tagName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "tag parameter is required."})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "50"))

	result, err := h.fileService.SearchByTag(c.Request.Context(), tagName, page, perPage)
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
