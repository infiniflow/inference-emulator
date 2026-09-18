package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"ollama-mock/internal/httputil"
	"ollama-mock/internal/models"
)

// EmbeddingsLegacy handles the legacy POST /api/embeddings and returns a single fixed vector.
func EmbeddingsLegacy(c *gin.Context) {
	var req models.EmbeddingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.BadRequest(c, "invalid request body")
		return
	}
	if req.Model == "" {
		httputil.BadRequest(c, "model is required")
		return
	}

	c.JSON(http.StatusOK, models.EmbeddingsResponse{Embedding: models.FixedEmbedding()})
}
