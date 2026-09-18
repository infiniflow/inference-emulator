package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"ollama-mock/internal/models"
	"ollama-mock/internal/router"
)

// TestMain disables the delay between streamed chunks to keep tests fast and deterministic.
func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	os.Setenv("OLLAMA_MOCK_STREAM_DELAY_MS", "0")
	os.Exit(m.Run())
}

func setupRouter() *gin.Engine {
	r := gin.New()
	router.Register(r)
	return r
}

// do performs a single request and returns the recorded response.
func do(t *testing.T, r *gin.Engine, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	var req *http.Request
	var err error
	if body == "" {
		req, err = http.NewRequest(method, path, nil)
	} else {
		req, err = http.NewRequest(method, path, strings.NewReader(body))
	}
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// ndjsonLines splits an NDJSON body into lines (trailing empty lines are ignored).
func ndjsonLines(t *testing.T, body string) []string {
	t.Helper()
	var lines []string
	for _, l := range strings.Split(body, "\n") {
		if strings.TrimSpace(l) != "" {
			lines = append(lines, l)
		}
	}
	return lines
}

func TestTags(t *testing.T) {
	w := do(t, setupRouter(), http.MethodGet, "/api/tags", "")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var resp models.TagsResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(resp.Models) != 3 {
		t.Fatalf("models = %d, want 3", len(resp.Models))
	}
	if resp.Models[0].Name != "llama3.2:latest" {
		t.Fatalf("first model = %q", resp.Models[0].Name)
	}
	if resp.Models[0].Details.Family != "llama" || resp.Models[0].Details.ParameterSize != "3.2B" {
		t.Fatalf("unexpected details: %+v", resp.Models[0].Details)
	}
	if ct := w.Header().Get("Content-Type"); !strings.Contains(ct, "application/json") {
		t.Fatalf("content-type = %q", ct)
	}
}

func TestVersion(t *testing.T) {
	w := do(t, setupRouter(), http.MethodGet, "/api/version", "")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var resp models.VersionResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.Version != models.Version {
		t.Fatalf("version = %q, want %q", resp.Version, models.Version)
	}
}

func TestPs(t *testing.T) {
	w := do(t, setupRouter(), http.MethodGet, "/api/ps", "")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var resp models.PsResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(resp.Models) != 3 {
		t.Fatalf("models = %d, want 3", len(resp.Models))
	}
}

func TestGenerateNonStream(t *testing.T) {
	w := do(t, setupRouter(), http.MethodPost, "/api/generate",
		`{"model":"llama3.2","prompt":"hi","stream":false}`)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var resp models.GenerateResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.Response != models.GenerateReply {
		t.Fatalf("response = %q, want %q", resp.Response, models.GenerateReply)
	}
	if !resp.Done || resp.DoneReason != "stop" {
		t.Fatalf("done = %v, done_reason = %q", resp.Done, resp.DoneReason)
	}
	if resp.Model != "llama3.2" || resp.EvalCount != 6 || resp.PromptEvalCount != 12 {
		t.Fatalf("unexpected response: %+v", resp)
	}
	if _, err := time.Parse(time.RFC3339Nano, resp.CreatedAt); err != nil {
		t.Fatalf("created_at = %q is not a valid RFC3339Nano timestamp", resp.CreatedAt)
	}
}

func TestGenerateStream(t *testing.T) {
	w := do(t, setupRouter(), http.MethodPost, "/api/generate", `{"model":"llama3.2","prompt":"hi"}`)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/x-ndjson" {
		t.Fatalf("content-type = %q, want application/x-ndjson", ct)
	}

	lines := ndjsonLines(t, w.Body.String())
	if len(lines) != len(models.GenerateTokens)+1 {
		t.Fatalf("lines = %d, want %d", len(lines), len(models.GenerateTokens)+1)
	}

	var sb strings.Builder
	for i, l := range lines {
		var chunk models.GenerateResponse
		if err := json.Unmarshal([]byte(l), &chunk); err != nil {
			t.Fatalf("line %d is not valid json: %v", i, err)
		}
		sb.WriteString(chunk.Response)
		if i < len(lines)-1 {
			if chunk.Done {
				t.Fatalf("line %d should not be done", i)
			}
			if chunk.Response != models.GenerateTokens[i] {
				t.Fatalf("line %d response = %q, want %q", i, chunk.Response, models.GenerateTokens[i])
			}
		}
	}

	if sb.String() != models.GenerateReply {
		t.Fatalf("concatenated = %q, want %q", sb.String(), models.GenerateReply)
	}

	var last models.GenerateResponse
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &last); err != nil {
		t.Fatalf("unmarshal last: %v", err)
	}
	if !last.Done {
		t.Fatalf("last chunk done = false")
	}
	if last.TotalDuration != 600000000 || last.EvalCount != 6 {
		t.Fatalf("unexpected last chunk: %+v", last)
	}
}

func TestChatNonStream(t *testing.T) {
	w := do(t, setupRouter(), http.MethodPost, "/api/chat",
		`{"model":"llama3.2","messages":[{"role":"user","content":"hi"}],"stream":false}`)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var resp models.ChatResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.Message.Role != "assistant" || resp.Message.Content != models.ChatReply {
		t.Fatalf("unexpected message: %+v", resp.Message)
	}
	if !resp.Done || resp.DoneReason != "stop" || resp.EvalCount != 7 {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestChatStream(t *testing.T) {
	w := do(t, setupRouter(), http.MethodPost, "/api/chat",
		`{"model":"llama3.2","messages":[{"role":"user","content":"hi"}]}`)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/x-ndjson" {
		t.Fatalf("content-type = %q, want application/x-ndjson", ct)
	}

	lines := ndjsonLines(t, w.Body.String())
	if len(lines) != len(models.ChatTokens)+1 {
		t.Fatalf("lines = %d, want %d", len(lines), len(models.ChatTokens)+1)
	}

	var sb strings.Builder
	for i, l := range lines[:len(lines)-1] {
		var chunk models.ChatResponse
		if err := json.Unmarshal([]byte(l), &chunk); err != nil {
			t.Fatalf("line %d is not valid json: %v", i, err)
		}
		if chunk.Message.Role != "assistant" {
			t.Fatalf("line %d role = %q", i, chunk.Message.Role)
		}
		sb.WriteString(chunk.Message.Content)
	}
	if sb.String() != models.ChatReply {
		t.Fatalf("concatenated = %q, want %q", sb.String(), models.ChatReply)
	}

	var last models.ChatResponse
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &last); err != nil {
		t.Fatalf("unmarshal last: %v", err)
	}
	if !last.Done || last.DoneReason != "stop" || last.TotalDuration != 700000000 {
		t.Fatalf("unexpected last chunk: %+v", last)
	}
}

func TestEmbedBatch(t *testing.T) {
	w := do(t, setupRouter(), http.MethodPost, "/api/embed",
		`{"model":"nomic-embed-text","input":["a","b"]}`)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var resp models.EmbedResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(resp.Embeddings) != 2 {
		t.Fatalf("embeddings = %d, want 2", len(resp.Embeddings))
	}
	for i, e := range resp.Embeddings {
		if len(e) != models.EmbeddingDim {
			t.Fatalf("embedding[%d] dim = %d, want %d", i, len(e), models.EmbeddingDim)
		}
		for j, v := range e {
			if v != models.EmbeddingValue {
				t.Fatalf("embedding[%d][%d] = %v, want %v", i, j, v, models.EmbeddingValue)
			}
		}
	}
	if resp.PromptEvalCount != 2 || resp.TotalDuration != 1000000 || resp.LoadDuration != 500000 {
		t.Fatalf("unexpected stats: %+v", resp)
	}
}

func TestEmbedSingleString(t *testing.T) {
	w := do(t, setupRouter(), http.MethodPost, "/api/embed",
		`{"model":"nomic-embed-text","input":"hello"}`)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var resp models.EmbedResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(resp.Embeddings) != 1 || len(resp.Embeddings[0]) != models.EmbeddingDim {
		t.Fatalf("unexpected embeddings: %d", len(resp.Embeddings))
	}
	if resp.PromptEvalCount != 1 {
		t.Fatalf("prompt_eval_count = %d, want 1", resp.PromptEvalCount)
	}
}

func TestEmbeddingsLegacy(t *testing.T) {
	w := do(t, setupRouter(), http.MethodPost, "/api/embeddings",
		`{"model":"nomic-embed-text","prompt":"a"}`)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var resp models.EmbeddingsResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(resp.Embedding) != models.EmbeddingDim {
		t.Fatalf("embedding dim = %d, want %d", len(resp.Embedding), models.EmbeddingDim)
	}
}

func TestShow(t *testing.T) {
	w := do(t, setupRouter(), http.MethodPost, "/api/show", `{"model":"llama3.2:latest"}`)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var resp models.ShowResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.Details.Family != "llama" || resp.Details.ParameterSize != "3.2B" {
		t.Fatalf("unexpected details: %+v", resp.Details)
	}
	if resp.Template == "" || resp.Modelfile == "" {
		t.Fatalf("template/modelfile should not be empty")
	}
	if _, ok := resp.ModelInfo["llama.context_length"]; !ok {
		t.Fatalf("model_info missing llama.context_length: %+v", resp.ModelInfo)
	}
}

func TestShowUnknownModel(t *testing.T) {
	w := do(t, setupRouter(), http.MethodPost, "/api/show", `{"model":"does-not-exist:latest"}`)
	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", w.Code)
	}
	assertErrorBody(t, w.Body.String())
}

func TestPullStream(t *testing.T) {
	w := do(t, setupRouter(), http.MethodPost, "/api/pull", `{"model":"llama3.2:latest"}`)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/x-ndjson" {
		t.Fatalf("content-type = %q", ct)
	}
	lines := ndjsonLines(t, w.Body.String())
	if len(lines) != 7 {
		t.Fatalf("lines = %d, want 7", len(lines))
	}
	var last models.ProgressResponse
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &last); err != nil {
		t.Fatalf("unmarshal last: %v", err)
	}
	if last.Status != "success" {
		t.Fatalf("last status = %q, want success", last.Status)
	}
}

func TestPullNonStream(t *testing.T) {
	w := do(t, setupRouter(), http.MethodPost, "/api/pull", `{"model":"llama3.2:latest","stream":false}`)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var resp models.ProgressResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.Status != "success" {
		t.Fatalf("status = %q, want success", resp.Status)
	}
}

func TestPushStream(t *testing.T) {
	w := do(t, setupRouter(), http.MethodPost, "/api/push", `{"model":"llama3.2:latest"}`)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	lines := ndjsonLines(t, w.Body.String())
	if len(lines) != 3 {
		t.Fatalf("lines = %d, want 3", len(lines))
	}
	var last models.ProgressResponse
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &last); err != nil {
		t.Fatalf("unmarshal last: %v", err)
	}
	if last.Status != "success" {
		t.Fatalf("last status = %q", last.Status)
	}
}

func TestCreateStream(t *testing.T) {
	w := do(t, setupRouter(), http.MethodPost, "/api/create",
		`{"model":"mario","from":"llama3.2:latest","system":"you are mario"}`)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	lines := ndjsonLines(t, w.Body.String())
	if len(lines) != 4 {
		t.Fatalf("lines = %d, want 4", len(lines))
	}
	var last models.ProgressResponse
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &last); err != nil {
		t.Fatalf("unmarshal last: %v", err)
	}
	if last.Status != "success" {
		t.Fatalf("last status = %q", last.Status)
	}
}

func TestCopy(t *testing.T) {
	w := do(t, setupRouter(), http.MethodPost, "/api/copy",
		`{"source":"llama3.2:latest","destination":"llama3.2-backup"}`)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if w.Body.Len() != 0 {
		t.Fatalf("body = %q, want empty", w.Body.String())
	}
}

func TestDelete(t *testing.T) {
	w := do(t, setupRouter(), http.MethodDelete, "/api/delete", `{"model":"llama3.2:latest"}`)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if w.Body.Len() != 0 {
		t.Fatalf("body = %q, want empty", w.Body.String())
	}
}

func TestInvalidRequestBody(t *testing.T) {
	w := do(t, setupRouter(), http.MethodPost, "/api/generate", `{"model":`)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", w.Code)
	}
	assertErrorBody(t, w.Body.String())
}

func TestMissingModel(t *testing.T) {
	for _, ep := range []string{"/api/generate", "/api/chat", "/api/embed", "/api/pull"} {
		w := do(t, setupRouter(), http.MethodPost, ep, `{}`)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("%s status = %d, want 400", ep, w.Code)
		}
		assertErrorBody(t, w.Body.String())
	}
}

func TestUnknownRoute(t *testing.T) {
	w := do(t, setupRouter(), http.MethodGet, "/api/unknown", "")
	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", w.Code)
	}
	assertErrorBody(t, w.Body.String())
}

func TestMethodNotAllowed(t *testing.T) {
	w := do(t, setupRouter(), http.MethodGet, "/api/generate", "")
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want 405", w.Code)
	}
	assertErrorBody(t, w.Body.String())
}

// assertErrorBody validates the unified error payload {"error": "..."}.
func assertErrorBody(t *testing.T, body string) {
	t.Helper()
	var resp models.ErrorResponse
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		t.Fatalf("error body is not json: %q", body)
	}
	if resp.Error == "" {
		t.Fatalf("error field is empty: %q", body)
	}
}
