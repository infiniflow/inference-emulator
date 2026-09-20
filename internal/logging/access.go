// Package logging provides the per-request access log of the mock server.
//
// The middleware is enabled by default and can be switched off with
// OLLAMA_MOCK_ACCESS_LOG=off. It replaces gin.Logger() with a structured,
// greppable single-line entry per request.
package logging

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"ollama-mock/internal/auth"
	"ollama-mock/internal/config"

	"github.com/gin-gonic/gin"
)

// maxLoggedBody is the upper bound of the request body read to extract the model
// name. The body is restored afterward so handlers can still read it.
const maxLoggedBody = 4 << 20 // 4 MiB

var (
	mu     sync.Mutex
	out    io.Writer = os.Stdout
	logger           = log.New(os.Stdout, "", log.LstdFlags|log.Lmicroseconds)
)

// SetOutput redirects the access log (useful for tests or file output).
func SetOutput(w io.Writer) {
	mu.Lock()
	defer mu.Unlock()
	out = w
	logger.SetOutput(w)
}

// Output returns the current access log writer.
func Output() io.Writer {
	mu.Lock()
	defer mu.Unlock()
	return out
}

// Middleware logs one line per request. It is a no-op when the access log is
// disabled through OLLAMA_MOCK_ACCESS_LOG=off.
func Middleware() gin.HandlerFunc {
	if !config.AccessLogEnabled() {
		return func(c *gin.Context) { c.Next() }
	}

	return func(c *gin.Context) {
		start := time.Now()
		model := peekModel(c)

		c.Next()

		mu.Lock()
		defer mu.Unlock()
		logger.Printf("access method=%s path=%s status=%d latency=%s bytes=%d ip=%s key=%s model=%s",
			c.Request.Method,
			c.Request.URL.Path,
			c.Writer.Status(),
			time.Since(start).Truncate(time.Microsecond),
			c.Writer.Size(),
			c.ClientIP(),
			maskKey(keyOf(c)),
			model,
		)
	}
}

// peekModel reads the model field of a JSON request body without consuming it.
func peekModel(c *gin.Context) string {
	if c.Request.Body == nil {
		return ""
	}
	data, err := io.ReadAll(io.LimitReader(c.Request.Body, maxLoggedBody))
	// Restore the body so the handler can still bind it.
	c.Request.Body = io.NopCloser(bytes.NewReader(data))
	if err != nil {
		return ""
	}

	// Best effort: only try to parse bodies that look like a JSON object, so
	// form-encoded or empty bodies are left untouched.
	if trimmed := strings.TrimLeft(string(data), " \t\r\n"); !strings.HasPrefix(trimmed, "{") {
		return ""
	}

	var payload struct {
		Model string `json:"model"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return ""
	}
	return payload.Model
}

// keyOf returns the API key of the request, or "-" when none was supplied.
func keyOf(c *gin.Context) string {
	if k, ok := auth.ExtractKey(c); ok && k != "" {
		return k
	}
	return ""
}

// maskKey hides most of the key, e.g. "sk-mock-123" -> "sk-m...-123".
func maskKey(key string) string {
	if key == "" {
		return "-"
	}
	if len(key) <= 8 {
		return "***"
	}
	return key[:4] + "..." + key[len(key)-4:]
}
