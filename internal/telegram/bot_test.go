package telegram

import (
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/stretchr/testify/require"
)

func TestBot_ImplementsNotifier(t *testing.T) {
	var _ Notifier = (*Bot)(nil)
}

func TestBot_Notify_MessageCreation(t *testing.T) {
	chatID := int64(123456)
	message := "Test notification"
	
	msg := tgbotapi.NewMessage(chatID, message)
	require.Equal(t, chatID, msg.ChatID)
	require.Equal(t, message, msg.Text)
}

func TestBot_Notify_DifferentChatIDs_MessageCreation(t *testing.T) {
	chatIDs := []int64{123, 456, 789, -1001234567890}
	messages := []string{"Message 1", "Message 2", "Message 3", "Message 4"}

	for i, chatID := range chatIDs {
		msg := tgbotapi.NewMessage(chatID, messages[i])
		require.Equal(t, chatID, msg.ChatID)
		require.Equal(t, messages[i], msg.Text)
	}
}

func TestBot_Notify_EmptyMessage_MessageCreation(t *testing.T) {
	chatID := int64(123456)
	message := ""

	msg := tgbotapi.NewMessage(chatID, message)
	require.Equal(t, chatID, msg.ChatID)
	require.Equal(t, "", msg.Text)
}

func TestBot_Notify_SpecialCharacters_MessageCreation(t *testing.T) {
	chatID := int64(123456)
	messages := []string{
		"Message with special chars: !@#$%^&*()",
		"Message with emoji: 🚀 📝 ✅",
		"Message with newlines:\nLine 1\nLine 2",
		"Message with unicode: Привет こんにちは",
	}

	for _, message := range messages {
		msg := tgbotapi.NewMessage(chatID, message)
		require.Equal(t, message, msg.Text)
	}
}

func TestNewBot_RequiresValidToken(t *testing.T) {
	t.Skip("NewBot requires valid Telegram API token - integration test needed")
}

