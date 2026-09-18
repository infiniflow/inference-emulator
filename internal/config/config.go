// Package config reads the runtime configuration of the mock server (port, stream delay, ...).
package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	// DefaultPort is the default port of the native Ollama server.
	DefaultPort = "11434"
	// DefaultStreamDelay is the default delay between two streamed chunks.
	DefaultStreamDelay = 10 * time.Millisecond
)

// Port returns the listening port, overridable via the OLLAMA_MOCK_PORT env var.
func Port() string {
	if p := os.Getenv("OLLAMA_MOCK_PORT"); p != "" {
		return p
	}
	return DefaultPort
}

// StreamDelay returns the delay between two streamed chunks. It can be
// overridden with OLLAMA_MOCK_STREAM_DELAY_MS (milliseconds, 0 disables the wait).
func StreamDelay() time.Duration {
	raw := os.Getenv("OLLAMA_MOCK_STREAM_DELAY_MS")
	if raw == "" {
		return DefaultStreamDelay
	}
	ms, err := strconv.Atoi(raw)
	if err != nil || ms < 0 {
		return DefaultStreamDelay
	}
	return time.Duration(ms) * time.Millisecond
}

// APIKeys returns the accepted API keys, taken from OLLAMA_MOCK_API_KEY (single
// key) and OLLAMA_MOCK_API_KEYS (comma-separated list). An empty result means
// authentication is disabled, which matches the native Ollama behaviour.
func APIKeys() []string {
	var keys []string
	for _, raw := range []string{
		os.Getenv("OLLAMA_MOCK_API_KEY"),
		os.Getenv("OLLAMA_MOCK_API_KEYS"),
	} {
		for _, k := range strings.Split(raw, ",") {
			if k = strings.TrimSpace(k); k != "" {
				keys = append(keys, k)
			}
		}
	}
	return keys
}

// AuthEnabled reports whether API key authentication is turned on.
func AuthEnabled() bool {
	return len(APIKeys()) > 0
}

// AuthExemptPaths returns the route paths that stay open even when
// authentication is enabled, configured via OLLAMA_MOCK_AUTH_EXEMPT
// (comma-separated, e.g. "/api/version"). Empty by default: every /api route
// requires a key.
func AuthExemptPaths() []string {
	var paths []string
	for _, p := range strings.Split(os.Getenv("OLLAMA_MOCK_AUTH_EXEMPT"), ",") {
		if p = strings.TrimSpace(p); p != "" {
			paths = append(paths, p)
		}
	}
	return paths
}
