package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"ollama-mock/internal/router"
)

// setupRouterWithAuth builds a router with a single accepted API key.
func setupRouterWithAuth(t *testing.T, key string, exempt ...string) *gin.Engine {
	t.Helper()
	t.Setenv("OLLAMA_MOCK_API_KEY", key)
	if len(exempt) > 0 {
		t.Setenv("OLLAMA_MOCK_AUTH_EXEMPT", exempt[0])
	}
	r := gin.New()
	router.Register(r)
	return r
}

// getWithHeaders performs a GET request with the given headers.
func getWithHeaders(t *testing.T, r *gin.Engine, path string, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestAuthDisabledByDefault(t *testing.T) {
	r := setupRouter()
	w := getWithHeaders(t, r, "/api/tags", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
}

func TestAuthBearerHeaderAccepted(t *testing.T) {
	r := setupRouterWithAuth(t, "sk-test-key")
	w := getWithHeaders(t, r, "/api/tags", map[string]string{
		"Authorization": "Bearer sk-test-key",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
}

func TestAuthBareHeaderAccepted(t *testing.T) {
	r := setupRouterWithAuth(t, "sk-test-key")
	w := getWithHeaders(t, r, "/api/tags", map[string]string{
		"Authorization": "sk-test-key",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
}

func TestAuthXAPIKeyHeaderAccepted(t *testing.T) {
	r := setupRouterWithAuth(t, "sk-test-key")
	w := getWithHeaders(t, r, "/api/tags", map[string]string{
		"X-API-Key": "sk-test-key",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
}

func TestAuthMissingKey(t *testing.T) {
	r := setupRouterWithAuth(t, "sk-test-key")
	w := getWithHeaders(t, r, "/api/tags", nil)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", w.Code)
	}
	assertErrorBody(t, w.Body.String())
	if ww := w.Header().Get("WWW-Authenticate"); ww == "" {
		t.Fatal("missing WWW-Authenticate header")
	}
}

func TestAuthWrongKey(t *testing.T) {
	r := setupRouterWithAuth(t, "sk-test-key")
	w := getWithHeaders(t, r, "/api/tags", map[string]string{
		"Authorization": "Bearer wrong-key",
	})
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", w.Code)
	}
	assertErrorBody(t, w.Body.String())
}

func TestAuthAppliesToAllEndpoints(t *testing.T) {
	r := setupRouterWithAuth(t, "sk-test-key")
	for _, p := range []string{"/api/tags", "/api/version", "/api/ps"} {
		w := getWithHeaders(t, r, p, nil)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("%s status = %d, want 401", p, w.Code)
		}
	}
	w := do(t, r, http.MethodPost, "/api/generate", `{"model":"llama3.2","prompt":"hi"}`)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("/api/generate status = %d, want 401", w.Code)
	}
}

func TestAuthExemptPath(t *testing.T) {
	r := setupRouterWithAuth(t, "sk-test-key", "/api/version")

	w := getWithHeaders(t, r, "/api/version", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("exempt path status = %d, want 200", w.Code)
	}

	w = getWithHeaders(t, r, "/api/tags", nil)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("protected path status = %d, want 401", w.Code)
	}
}
