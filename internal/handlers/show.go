package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"ollama-mock/internal/httputil"
	"ollama-mock/internal/models"
)

// Show handles POST /api/show and returns the fixed details of the given model.
func Show(c *gin.Context) {
	var req models.ShowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.BadRequest(c, "invalid request body")
		return
	}
	if req.Model == "" {
		httputil.BadRequest(c, "model is required")
		return
	}

	info, ok := models.ShowInfo(req.Model)
	if !ok {
		httputil.NotFound(c, fmt.Sprintf("model %q not found", req.Model))
		return
	}

	c.JSON(http.StatusOK, info)
}
