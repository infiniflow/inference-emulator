// Package config reads the runtime configuration of the mock server (port, stream delay, ...).
package config

import (
	"os"
	"strconv"
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
