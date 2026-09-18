package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"ollama-mock/internal/config"
	"ollama-mock/internal/httputil"
	"ollama-mock/internal/models"
	"ollama-mock/internal/stream"
)

// Chat handles POST /api/chat.
// When stream is true or omitted it emits NDJSON line by line, otherwise a single JSON object.
func Chat(c *gin.Context) {
	var req models.ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.BadRequest(c, "invalid request body")
		return
	}
	if req.Model == "" {
		httputil.BadRequest(c, "model is required")
		return
	}

	if req.Stream != nil && !*req.Stream {
		c.JSON(http.StatusOK, nonStreamChat(req.Model))
		return
	}

	stream.SetNDJSONHeader(c)
	delay := config.StreamDelay()
	for _, token := range models.ChatTokens {
		if err := stream.WriteNDJSONWithDelay(c, models.ChatResponse{
			Model:     req.Model,
			CreatedAt: models.Now(),
			Message:   models.ChatMessage{Role: "assistant", Content: token},
			Done:      false,
		}, delay); err != nil {
			return
		}
	}

	total, promptEval, eval := models.ChatStats()
	_ = stream.WriteNDJSON(c, models.ChatResponse{
		Model:           req.Model,
		CreatedAt:       models.Now(),
		Message:         models.ChatMessage{Role: "assistant", Content: ""},
		Done:            true,
		DoneReason:      "stop",
		TotalDuration:   total,
		PromptEvalCount: promptEval,
		EvalCount:       eval,
	})
}

// nonStreamChat builds the complete non-streaming response.
func nonStreamChat(model string) models.ChatResponse {
	total, promptEval, eval := models.ChatStats()
	return models.ChatResponse{
		Model:           model,
		CreatedAt:       models.Now(),
		Message:         models.ChatMessage{Role: "assistant", Content: models.ChatReply},
		Done:            true,
		DoneReason:      "stop",
		TotalDuration:   total,
		PromptEvalCount: promptEval,
		EvalCount:       eval,
	}
}
