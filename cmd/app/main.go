package main

import (
	"log"
	"os"

	"github.com/fncg/ReviewerService/internal/http"
	"github.com/fncg/ReviewerService/internal/storage"
	"github.com/fncg/ReviewerService/internal/telegram"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://reviewer:reviewer@localhost:5432/reviewer"
	}

	db, err := storage.NewPostgres(dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	botToken := "8435677156:AAG98zFcGiAlIRxaC6Q-GL1X81h9So8qj9w"
	bot, err := telegram.NewBot(botToken)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Connected to PostgreSQL successfully!")

	server := http.NewServer(db, bot)

	log.Println("HTTP server starting on :8080")
	log.Fatal(server.Start(":8080"))
}
