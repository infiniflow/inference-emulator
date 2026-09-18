// Package stream provides helpers for writing NDJSON (application/x-ndjson) streams.
package stream

import (
	"encoding/json"
	"time"

	"github.com/gin-gonic/gin"
)

// SetNDJSONHeader sets the streaming response headers.
// It must be called before the first line is written.
func SetNDJSONHeader(c *gin.Context) {
	c.Header("Content-Type", "application/x-ndjson")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no") // disable nginx buffering
}

// WriteNDJSON writes one JSON line and flushes it immediately.
func WriteNDJSON(c *gin.Context, v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	if _, err := c.Writer.Write(data); err != nil {
		return err
	}
	if _, err := c.Writer.Write([]byte("\n")); err != nil {
		return err
	}
	c.Writer.Flush()
	return nil
}

// WriteNDJSONWithDelay writes one JSON line, flushes it, and then waits for the
// given delay to mimic the pacing of a real inference.
func WriteNDJSONWithDelay(c *gin.Context, v interface{}, delay time.Duration) error {
	if err := WriteNDJSON(c, v); err != nil {
		return err
	}
	if delay > 0 {
		time.Sleep(delay)
	}
	return nil
}
