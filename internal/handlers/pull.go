package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"ollama-mock/internal/config"
	"ollama-mock/internal/httputil"
	"ollama-mock/internal/models"
	"ollama-mock/internal/stream"
)

// Pull handles POST /api/pull.
// When stream is true or omitted it emits progress as NDJSON, otherwise only the final success line.
func Pull(c *gin.Context) {
	var req models.PullRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.BadRequest(c, "invalid request body")
		return
	}
	if req.Model == "" {
		httputil.BadRequest(c, "model is required")
		return
	}

	digest, total := pullTarget(req.Model)

	if req.Stream != nil && !*req.Stream {
		c.JSON(http.StatusOK, models.ProgressResponse{Status: "success"})
		return
	}

	short := shortDigest(digest)
	delay := config.StreamDelay()

	steps := []models.ProgressResponse{
		{Status: "pulling manifest"},
		{Status: "pulling " + short, Digest: "sha256:" + digest, Total: total, Completed: total / 4},
		{Status: "pulling " + short, Digest: "sha256:" + digest, Total: total, Completed: total / 2},
		{Status: "pulling " + short, Digest: "sha256:" + digest, Total: total, Completed: total},
		{Status: "verifying sha256 digest"},
		{Status: "writing manifest"},
		{Status: "success"},
	}

	stream.SetNDJSONHeader(c)
	for _, s := range steps {
		if err := stream.WriteNDJSONWithDelay(c, s, delay); err != nil {
			return
		}
	}
}

// pullTarget returns the digest and total size of the model being pulled.
// Unknown models fall back to the first fixed model so the mock always succeeds.
func pullTarget(name string) (digest string, total int64) {
	if m := models.FindModel(name); m != nil {
		return m.Digest, m.Size
	}
	all := models.FixedModels()
	return all[0].Digest, all[0].Size
}

// shortDigest returns the short digest prefix used in Ollama progress lines.
func shortDigest(digest string) string {
	if len(digest) > 12 {
		return digest[:12]
	}
	return digest
}
