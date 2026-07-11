package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mamcer/gostalgia/internal/app/scan"
)

type ScanHandler struct {
	scanService *scan.ScanService
}

func NewScanHandler(scanService *scan.ScanService) *ScanHandler {
	return &ScanHandler{scanService: scanService}
}

// GetRecentScans godoc
// @Summary      Obtener escaneos recientes
// @Description  Devuelve una lista de los escaneos realizados recientemente
// @Tags         scans
// @Produce      json
// @Param        limit  query     int  false  "Cantidad máxima de resultados" default(5)
// @Success      200    {array}   domain.NScan
// @Failure      500    {object}  map[string]string
// @Router       /scans/recent [get]
func (h *ScanHandler) GetRecentScans(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "5"))
	scans, err := h.scanService.GetRecentScans(c.Request.Context(), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, scans)
}
