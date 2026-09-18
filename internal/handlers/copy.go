package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"ollama-mock/internal/httputil"
	"ollama-mock/internal/models"
)

// Copy handles POST /api/copy; on success it returns 200 with an empty body.
func Copy(c *gin.Context) {
	var req models.CopyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.BadRequest(c, "invalid request body")
		return
	}
	if req.Source == "" || req.Destination == "" {
		httputil.BadRequest(c, "source and destination are required")
		return
	}

	c.Status(http.StatusOK)
}
