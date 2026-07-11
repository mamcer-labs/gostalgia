package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mamcer/gostalgia/internal/app/search"
)

type SearchHandler struct {
	searchService *search.SearchService
}

func NewSearchHandler(searchService *search.SearchService) *SearchHandler {
	return &SearchHandler{searchService: searchService}
}

// UnifiedSearch godoc
// @Summary      Búsqueda unificada
// @Description  Busca archivos, directorios y etiquetas que coincidan con el texto
// @Tags         search
// @Produce      json
// @Param        q     query     string  true  "Texto a buscar"
// @Success      200  {object}  dto.UnifiedSearchResultDto
// @Failure      400  {object}  map[string]string
// @Router       /search [get]
func (h *SearchHandler) UnifiedSearch(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Query parameter 'q' is required"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	result, err := h.searchService.UnifiedSearch(c.Request.Context(), query, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}
