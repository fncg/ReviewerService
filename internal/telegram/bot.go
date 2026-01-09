package telegram

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	api *tgbotapi.BotAPI
}

func NewBot(token string) (*Bot, error) {
	botAPI, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	log.Printf("Authorized on account %s", botAPI.Self.UserName)
	return &Bot{api: botAPI}, nil
}

func (b *Bot) Notify(chatID int64, message string) {
	msg := tgbotapi.NewMessage(chatID, message)
	if _, err := b.api.Send(msg); err != nil {
		log.Println("Failed to send Telegram message:", err)
	}
}
