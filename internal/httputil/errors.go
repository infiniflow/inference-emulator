// Package httputil centralizes the unified error responses; every error uses the
// {"error": "..."} shape.
package httputil

import "github.com/gin-gonic/gin"

// Unauthorized responds with 401.
func Unauthorized(c *gin.Context, msg string) {
	c.JSON(401, gin.H{"error": msg})
}

// BadRequest responds with 400.
func BadRequest(c *gin.Context, msg string) {
	c.JSON(400, gin.H{"error": msg})
}

// NotFound responds with 404.
func NotFound(c *gin.Context, msg string) {
	c.JSON(404, gin.H{"error": msg})
}

// MethodNotAllowed responds with 405.
func MethodNotAllowed(c *gin.Context, msg string) {
	c.JSON(405, gin.H{"error": msg})
}

// InternalError responds with 500.
func InternalError(c *gin.Context, msg string) {
	c.JSON(500, gin.H{"error": msg})
}
