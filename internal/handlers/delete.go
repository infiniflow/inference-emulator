package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"ollama-mock/internal/httputil"
	"ollama-mock/internal/models"
)

// Delete handles DELETE /api/delete; on success it returns 200 with an empty body.
func Delete(c *gin.Context) {
	var req models.DeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.BadRequest(c, "invalid request body")
		return
	}
	if req.Model == "" {
		httputil.BadRequest(c, "model is required")
		return
	}

	c.Status(http.StatusOK)
}
