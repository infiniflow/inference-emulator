// Package handlers implements the handlers of every native Ollama API endpoint.
package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"ollama-mock/internal/models"
)

// Tags handles GET /api/tags and returns the fixed model list.
func Tags(c *gin.Context) {
	c.JSON(http.StatusOK, models.TagsResponse{Models: models.FixedModels()})
}
