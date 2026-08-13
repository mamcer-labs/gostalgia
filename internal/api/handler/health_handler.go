package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type HealthHandler struct {
	version string
}

func NewHealthHandler(version string) *HealthHandler {
	return &HealthHandler{version: version}
}

// Health godoc
// @Summary      Chequeo de salud
// @Description  Devuelve el estado de la API y su versión
// @Tags         health
// @Produce      json
// @Success      200  {object}  map[string]string
// @Router       /health [get]
func (h *HealthHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "OK",
		"version":   h.version,
		"timestamp": time.Now().UTC(),
		"message":   "API is ready",
	})
}

// Fail godoc
// @Summary      Fallo deliberado
// @Description  Endpoint de diagnóstico que siempre devuelve un error 500 — usado para probar la correlación trace<->logs en Grafana (Tempo/Loki)
// @Tags         health
// @Produce      json
// @Failure      500  {object}  map[string]string
// @Router       /debug/fail [get]
func (h *HealthHandler) Fail(c *gin.Context) {
	_ = c.Error(errors.New("deliberate failure for observability demo"))
	c.JSON(http.StatusInternalServerError, gin.H{"message": "deliberate failure for observability demo"})
}
