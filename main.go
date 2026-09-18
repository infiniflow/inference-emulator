// Command ollama-mock starts an HTTP server that emulates the native Ollama API.
package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"ollama-mock/internal/config"
	"ollama-mock/internal/router"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	router.Register(r)

	addr := ":" + config.Port()
	log.Printf("Ollama mock listening on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatal(err)
	}
}
