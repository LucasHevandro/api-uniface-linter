package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"uniface-linter/internal/api"
)

const version = "1.1.0"

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := api.New(version)
	handler := api.Chain(server, api.Logger, api.CORS)

	addr := fmt.Sprintf(":%s", port)
	log.Printf("uniface-linter API v%s iniciando em http://0.0.0.0%s", version, addr)
	log.Printf("Endpoints disponíveis:")
	log.Printf("  GET  /health")
	log.Printf("  GET  /rules")
	log.Printf("  POST /analyze        (field: 'file'  | body XML direto)")
	log.Printf("  POST /analyze/batch  (field: 'files')")
	log.Printf("  Parâmetros de query: ?max_lines=100&disable=GLB002,DOC003")

	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("Erro ao iniciar servidor: %v", err)
	}
}
