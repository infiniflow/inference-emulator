package tests

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"ollama-mock/internal/logging"
	"ollama-mock/internal/router"
)

// setupRouterWithAccessLog builds a router whose access log goes into a buffer,
// and returns the router together with that buffer.
func setupRouterWithAccessLog(t *testing.T) (*gin.Engine, *bytes.Buffer) {
	t.Helper()
	var buf bytes.Buffer
	logging.SetOutput(&buf)
	t.Cleanup(func() { logging.SetOutput(os.Stdout) })

	r := gin.New()
	r.Use(logging.Middleware())
	router.Register(r)
	return r, &buf
}

func TestAccessLogEnabled(t *testing.T) {
	r, buf := setupRouterWithAccessLog(t)

	req, _ := http.NewRequest(http.MethodGet, "/api/tags", nil)
	r.ServeHTTP(httptest.NewRecorder(), req)

	line := buf.String()
	for _, want := range []string{"access", "GET", "/api/tags", "status=200", "key=-"} {
		if !strings.Contains(line, want) {
			t.Fatalf("log line %q is missing %q", line, want)
		}
	}
	if strings.Count(strings.TrimSpace(line), "\n") != 0 {
		t.Fatalf("expected exactly one log line, got %q", line)
	}
}

func TestAccessLogContainsModelAndMaskedKey(t *testing.T) {
	r, buf := setupRouterWithAccessLog(t)
	t.Setenv("OLLAMA_MOCK_API_KEY", "sk-mock-123456")

	req, _ := http.NewRequest(http.MethodPost, "/api/generate",
		strings.NewReader(`{"model":"llama3.2:latest","prompt":"hi","stream":false}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer sk-mock-123456")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body was restored correctly?)", w.Code)
	}

	line := buf.String()
	for _, want := range []string{"method=POST", "/api/generate", "status=200", "model=llama3.2:latest", "key=sk-m...3456"} {
		if !strings.Contains(line, want) {
			t.Fatalf("log line %q is missing %q", line, want)
		}
	}
}

func TestAccessLogDisabled(t *testing.T) {
	t.Setenv("OLLAMA_MOCK_ACCESS_LOG", "off")
	r, buf := setupRouterWithAccessLog(t)

	req, _ := http.NewRequest(http.MethodGet, "/api/tags", nil)
	r.ServeHTTP(httptest.NewRecorder(), req)

	if buf.Len() != 0 {
		t.Fatalf("expected no access log, got %q", buf.String())
	}
}
