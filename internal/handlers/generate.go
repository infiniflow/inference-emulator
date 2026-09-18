package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"ollama-mock/internal/config"
	"ollama-mock/internal/httputil"
	"ollama-mock/internal/models"
	"ollama-mock/internal/stream"
)

// Generate handles POST /api/generate.
// When stream is true or omitted it emits NDJSON line by line, otherwise a single JSON object.
func Generate(c *gin.Context) {
	var req models.GenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.BadRequest(c, "invalid request body")
		return
	}
	if req.Model == "" {
		httputil.BadRequest(c, "model is required")
		return
	}

	// a nil stream field is treated as true
	if req.Stream != nil && !*req.Stream {
		c.JSON(http.StatusOK, nonStreamGenerate(req.Model))
		return
	}

	stream.SetNDJSONHeader(c)
	delay := config.StreamDelay()
	for _, token := range models.GenerateTokens {
		if err := stream.WriteNDJSONWithDelay(c, models.GenerateResponse{
			Model:     req.Model,
			CreatedAt: models.Now(),
			Response:  token,
			Done:      false,
		}, delay); err != nil {
			return // client is gone, stop writing
		}
	}

	total, load, promptEval, eval := models.GenerateStats()
	_ = stream.WriteNDJSON(c, models.GenerateResponse{
		Model:           req.Model,
		CreatedAt:       models.Now(),
		Response:        "",
		Done:            true,
		DoneReason:      "stop",
		TotalDuration:   total,
		LoadDuration:    load,
		PromptEvalCount: promptEval,
		EvalCount:       eval,
	})
}

// nonStreamGenerate builds the complete non-streaming response.
func nonStreamGenerate(model string) models.GenerateResponse {
	total, load, promptEval, eval := models.GenerateStats()
	return models.GenerateResponse{
		Model:           model,
		CreatedAt:       models.Now(),
		Response:        models.GenerateReply,
		Done:            true,
		DoneReason:      "stop",
		TotalDuration:   total,
		LoadDuration:    load,
		PromptEvalCount: promptEval,
		EvalCount:       eval,
	}
}
