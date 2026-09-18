// Package router registers every route in a single place.
package router

import (
	"github.com/gin-gonic/gin"
	"ollama-mock/internal/auth"
	"ollama-mock/internal/handlers"
	"ollama-mock/internal/httputil"
)

// Register mounts all mocked Ollama endpoints on the given gin.Engine.
func Register(r *gin.Engine) {
	// When enabled, a known path with an unmatched method hits NoMethod (405)
	// instead of gin's default 404.
	r.HandleMethodNotAllowed = true

	// Every /api route goes through the API key middleware; it is a no-op unless
	// at least one key is configured.
	api := r.Group("/api", auth.Middleware())
	{
		api.GET("/tags", handlers.Tags)
		api.GET("/version", handlers.Version)
		api.GET("/ps", handlers.Ps)

		api.POST("/generate", handlers.Generate)
		api.POST("/chat", handlers.Chat)
		api.POST("/embed", handlers.Embed)
		api.POST("/embeddings", handlers.EmbeddingsLegacy)
		api.POST("/show", handlers.Show)

		api.POST("/pull", handlers.Pull)
		api.POST("/push", handlers.Push)
		api.POST("/create", handlers.Create)
		api.POST("/copy", handlers.Copy)
		api.DELETE("/delete", handlers.Delete)
	}

	r.NoRoute(func(c *gin.Context) {
		httputil.NotFound(c, "not found")
	})
	r.NoMethod(func(c *gin.Context) {
		httputil.MethodNotAllowed(c, "method not allowed")
	})
}
