package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"ollama-mock/internal/models"
)

// Ps handles GET /api/ps and returns the fixed list of "running" models.
func Ps(c *gin.Context) {
	c.JSON(http.StatusOK, models.PsResponse{Models: models.ProcessModels()})
}
