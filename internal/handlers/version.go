package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"ollama-mock/internal/models"
)

// Version handles GET /api/version and returns the fixed version string.
func Version(c *gin.Context) {
	c.JSON(http.StatusOK, models.VersionResponse{Version: models.Version})
}
