package main

import (
	"log"

	"github.com/fncg/ReviewerService/internal/http"
)

func main() {
	server := http.NewServer()

	log.Println("HTTP server starting on :8080")
	log.Fatal(server.Start(":8080"))
}
