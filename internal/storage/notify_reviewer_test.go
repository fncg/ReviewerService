package storage

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type fakeBot struct {
	called bool
}

func (f *fakeBot) Notify(chatID int64, message string) {
	f.called = true
}

func TestNotifyReviewerCalled(t *testing.T) {
	bot := &fakeBot{}

	bot.Notify(123, "message")

	require.True(t, bot.called)
}
