// Package auth provides optional API key authentication for the mocked endpoints.
//
// Authentication is enabled only when at least one key is configured through
// OLLAMA_MOCK_API_KEY / OLLAMA_MOCK_API_KEYS. Otherwise every request is
// allowed, which mirrors the native Ollama server.
package auth

import (
	"strings"

	"github.com/gin-gonic/gin"
	"ollama-mock/internal/config"
	"ollama-mock/internal/httputil"
)

// Middleware builds the API key middleware. It must be attached to the route
// group that should be protected.
func Middleware() gin.HandlerFunc {
	keys := config.APIKeys()
	if len(keys) == 0 {
		// No key configured: authentication stays disabled.
		return func(c *gin.Context) { c.Next() }
	}

	accepted := make(map[string]struct{}, len(keys))
	for _, k := range keys {
		accepted[k] = struct{}{}
	}

	exempt := make(map[string]struct{}, len(config.AuthExemptPaths()))
	for _, p := range config.AuthExemptPaths() {
		exempt[p] = struct{}{}
	}

	return func(c *gin.Context) {
		// Routes explicitly excluded from authentication stay open.
		if _, ok := exempt[c.FullPath()]; ok {
			c.Next()
			return
		}

		if key, ok := extractKey(c); ok {
			if _, ok := accepted[key]; ok {
				c.Next()
				return
			}
		}

		// Rejected: 401 with the unified {"error": "..."} payload.
		c.Header("WWW-Authenticate", `Bearer realm="ollama-mock"`)
		httputil.Unauthorized(c, "unauthorized")
		c.Abort()
	}
}

// extractKey reads the key from X-API-Key or Authorization ("Bearer <key>" or a
// bare key). It reports false when no usable header is present.
func extractKey(c *gin.Context) (string, bool) {
	if v := strings.TrimSpace(c.GetHeader("X-API-Key")); v != "" {
		return v, true
	}

	auth := strings.TrimSpace(c.GetHeader("Authorization"))
	if auth == "" {
		return "", false
	}
	const prefix = "bearer "
	if len(auth) > len(prefix) && strings.EqualFold(auth[:len(prefix)], prefix) {
		return strings.TrimSpace(auth[len(prefix):]), true
	}
	return auth, true
}
