package telegram

type Notifier interface {
	Notify(chatID int64, message string)
}
