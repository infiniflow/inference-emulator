package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"ollama-mock/internal/httputil"
	"ollama-mock/internal/models"
)

// Embed handles POST /api/embed; input accepts either a string or a []string.
// Every input gets the same fixed 768-dimension vector.
func Embed(c *gin.Context) {
	var req models.EmbedRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.BadRequest(c, "invalid request body")
		return
	}
	if req.Model == "" {
		httputil.BadRequest(c, "model is required")
		return
	}
	if len(req.Input) == 0 {
		httputil.BadRequest(c, "input is required")
		return
	}

	embeddings := make([][]float32, 0, len(req.Input))
	for range req.Input {
		embeddings = append(embeddings, models.FixedEmbedding())
	}

	total, load, promptEval := models.EmbedStats(len(req.Input))
	c.JSON(http.StatusOK, models.EmbedResponse{
		Model:           req.Model,
		Embeddings:      embeddings,
		TotalDuration:   total,
		LoadDuration:    load,
		PromptEvalCount: promptEval,
	})
}
