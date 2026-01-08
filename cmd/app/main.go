package main

import (
	"log"

	"github.com/fncg/ReviewerService/internal/http"
	"github.com/fncg/ReviewerService/internal/storage"
)

func main() {
	dsn := "postgres://reviewer:reviewer@localhost:5432/reviewer"

	db, err := storage.NewPostgres(dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	log.Println("Connected to PostgreSQL successfully!")

	server := http.NewServer(db)

	log.Println("HTTP server starting on :8080")
	log.Fatal(server.Start(":8080"))
}
