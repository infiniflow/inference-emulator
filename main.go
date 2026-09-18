// Command ollama-mock starts an HTTP server that emulates the native Ollama API.
package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"ollama-mock/internal/config"
	"ollama-mock/internal/logging"
	"ollama-mock/internal/router"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	// One structured line per request; Recovery keeps the server alive on panics.
	r.Use(logging.Middleware(), gin.Recovery())

	router.Register(r)

	addr := ":" + config.Port()
	log.Printf("Ollama mock listening on %s", addr)
	if config.AuthEnabled() {
		log.Printf("API key authentication enabled with %d key(s)", len(config.APIKeys()))
	} else {
		log.Print("API key authentication disabled (no OLLAMA_MOCK_API_KEY configured)")
	}
	if err := r.Run(addr); err != nil {
		log.Fatal(err)
	}
}
