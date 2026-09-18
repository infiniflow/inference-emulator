package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"ollama-mock/internal/config"
	"ollama-mock/internal/httputil"
	"ollama-mock/internal/models"
	"ollama-mock/internal/stream"
)

// Create handles POST /api/create and returns the fixed progress statuses.
func Create(c *gin.Context) {
	var req models.CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.BadRequest(c, "invalid request body")
		return
	}
	if req.Model == "" {
		httputil.BadRequest(c, "model is required")
		return
	}

	steps := []models.ProgressResponse{
		{Status: "reading model metadata"},
		{Status: "creating system layer"},
		{Status: "writing manifest"},
		{Status: "success"},
	}

	if req.Stream != nil && !*req.Stream {
		c.JSON(http.StatusOK, steps[len(steps)-1])
		return
	}

	stream.SetNDJSONHeader(c)
	delay := config.StreamDelay()
	for _, s := range steps {
		if err := stream.WriteNDJSONWithDelay(c, s, delay); err != nil {
			return
		}
	}
}
