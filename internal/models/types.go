// Package models defines request/response structs whose JSON field names are
// identical to the native Ollama API, together with the fixed mock data
// (model list, fixed replies, fixed embeddings).
package models

import (
	"encoding/json"
	"time"
)

// Now returns an RFC3339Nano timestamp, e.g. 2025-01-15T10:30:00.000000000Z.
func Now() string {
	return time.Now().UTC().Format(time.RFC3339Nano)
}

// ---------------------------------------------------------------------------
// Common
// ---------------------------------------------------------------------------

// ModelDetails describes model metadata; field names match the native Ollama API.
type ModelDetails struct {
	Format            string   `json:"format"`
	Family            string   `json:"family"`
	Families          []string `json:"families"`
	ParameterSize     string   `json:"parameter_size"`
	QuantizationLevel string   `json:"quantization_level"`
}

// ModelInfo is a single model entry of /api/tags and /api/ps.
type ModelInfo struct {
	Name       string       `json:"name"`
	Model      string       `json:"model"`
	ModifiedAt string       `json:"modified_at"`
	Size       int64        `json:"size"`
	Digest     string       `json:"digest"`
	Details    ModelDetails `json:"details"`
}

// TagsResponse is the response of GET /api/tags.
type TagsResponse struct {
	Models []ModelInfo `json:"models"`
}

// ProcessModelInfo is a single running-model entry of GET /api/ps;
// it carries two extra fields compared to ModelInfo.
type ProcessModelInfo struct {
	Name       string       `json:"name"`
	Model      string       `json:"model"`
	ModifiedAt string       `json:"modified_at"`
	Size       int64        `json:"size"`
	Digest     string       `json:"digest"`
	Details    ModelDetails `json:"details"`
	ExpiresAt  string       `json:"expires_at"`
	SizeVram   int64        `json:"size_vram"`
}

// PsResponse is the response of GET /api/ps.
type PsResponse struct {
	Models []ProcessModelInfo `json:"models"`
}

// ErrorResponse is the unified error payload.
type ErrorResponse struct {
	Error string `json:"error"`
}

// VersionResponse is the response of GET /api/version.
type VersionResponse struct {
	Version string `json:"version"`
}

// ---------------------------------------------------------------------------
// /api/generate
// ---------------------------------------------------------------------------

// GenerateRequest is the body of POST /api/generate.
type GenerateRequest struct {
	Model   string                 `json:"model"`
	Prompt  string                 `json:"prompt"`
	Stream  *bool                  `json:"stream,omitempty"` // nil means true
	Options map[string]interface{} `json:"options,omitempty"`
}

// GenerateResponse is the body of POST /api/generate
// (shared by the streaming and non-streaming modes).
type GenerateResponse struct {
	Model           string `json:"model"`
	CreatedAt       string `json:"created_at"`
	Response        string `json:"response"`
	Done            bool   `json:"done"`
	DoneReason      string `json:"done_reason,omitempty"`
	TotalDuration   int64  `json:"total_duration,omitempty"`
	LoadDuration    int64  `json:"load_duration,omitempty"`
	PromptEvalCount int    `json:"prompt_eval_count,omitempty"`
	EvalCount       int    `json:"eval_count,omitempty"`
}

// ---------------------------------------------------------------------------
// /api/chat
// ---------------------------------------------------------------------------

// ChatMessage is a single chat message.
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatRequest is the body of POST /api/chat.
type ChatRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
	Stream   *bool         `json:"stream,omitempty"`
}

// ChatResponse is the body of POST /api/chat
// (shared by the streaming and non-streaming modes).
type ChatResponse struct {
	Model           string      `json:"model"`
	CreatedAt       string      `json:"created_at"`
	Message         ChatMessage `json:"message"`
	Done            bool        `json:"done"`
	DoneReason      string      `json:"done_reason,omitempty"`
	TotalDuration   int64       `json:"total_duration,omitempty"`
	LoadDuration    int64       `json:"load_duration,omitempty"`
	PromptEvalCount int         `json:"prompt_eval_count,omitempty"`
	EvalCount       int         `json:"eval_count,omitempty"`
}

// ---------------------------------------------------------------------------
// /api/embed and /api/embeddings (legacy)
// ---------------------------------------------------------------------------

// EmbedInput accepts either a single string or an array of strings.
type EmbedInput []string

// UnmarshalJSON first tries to parse a plain string, then falls back to a string array.
func (e *EmbedInput) UnmarshalJSON(data []byte) error {
	var single string
	if err := json.Unmarshal(data, &single); err == nil {
		*e = []string{single}
		return nil
	}
	var multi []string
	if err := json.Unmarshal(data, &multi); err != nil {
		return err
	}
	*e = multi
	return nil
}

// EmbedRequest is the body of POST /api/embed.
type EmbedRequest struct {
	Model string     `json:"model"`
	Input EmbedInput `json:"input"`
}

// EmbedResponse is the body of POST /api/embed.
type EmbedResponse struct {
	Model           string      `json:"model"`
	Embeddings      [][]float32 `json:"embeddings"`
	TotalDuration   int64       `json:"total_duration"`
	LoadDuration    int64       `json:"load_duration"`
	PromptEvalCount int         `json:"prompt_eval_count"`
}

// EmbeddingsRequest is the body of the legacy POST /api/embeddings.
type EmbeddingsRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

// EmbeddingsResponse is the body of the legacy POST /api/embeddings.
type EmbeddingsResponse struct {
	Embedding []float32 `json:"embedding"`
}

// ---------------------------------------------------------------------------
// /api/show
// ---------------------------------------------------------------------------

// ShowRequest is the body of POST /api/show.
type ShowRequest struct {
	Model string `json:"model"`
}

// ShowResponse is the body of POST /api/show.
type ShowResponse struct {
	License    string                 `json:"license"`
	Modelfile  string                 `json:"modelfile"`
	Parameters string                 `json:"parameters"`
	Template   string                 `json:"template"`
	Details    ModelDetails           `json:"details"`
	ModelInfo  map[string]interface{} `json:"model_info"`
}

// ---------------------------------------------------------------------------
// /api/pull, /api/push, /api/create, /api/copy, /api/delete
// ---------------------------------------------------------------------------

// ProgressResponse is the shared payload of the progress endpoints
// (pull / push / create).
type ProgressResponse struct {
	Status    string `json:"status"`
	Digest    string `json:"digest,omitempty"`
	Total     int64  `json:"total,omitempty"`
	Completed int64  `json:"completed,omitempty"`
}

// PullRequest is the body of POST /api/pull.
type PullRequest struct {
	Model    string `json:"model"`
	Insecure bool   `json:"insecure,omitempty"`
	Stream   *bool  `json:"stream,omitempty"`
}

// PushRequest is the body of POST /api/push.
type PushRequest struct {
	Model    string `json:"model"`
	Insecure bool   `json:"insecure,omitempty"`
	Stream   *bool  `json:"stream,omitempty"`
}

// CreateRequest is the body of POST /api/create.
type CreateRequest struct {
	Model      string                 `json:"model"`
	From       string                 `json:"from,omitempty"`
	Files      map[string]string      `json:"files,omitempty"`
	Adapters   map[string]string      `json:"adapters,omitempty"`
	Template   string                 `json:"template,omitempty"`
	Stream     *bool                  `json:"stream,omitempty"`
	Quantize   string                 `json:"quantize,omitempty"`
	License    interface{}            `json:"license,omitempty"`
	System     string                 `json:"system,omitempty"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
}

// CopyRequest is the body of POST /api/copy.
type CopyRequest struct {
	Source      string `json:"source"`
	Destination string `json:"destination"`
}

// DeleteRequest is the body of DELETE /api/delete.
type DeleteRequest struct {
	Model string `json:"model"`
}
