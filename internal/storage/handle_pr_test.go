package storage

import (
	"testing"

	"github.com/fncg/ReviewerService/internal/github"
	"github.com/stretchr/testify/require"
)

type mockBot struct {
	called bool
}

func (m *mockBot) Notify(chatID int64, message string) {
	m.called = true
}

type mockPostgres struct {
	saveCalled   bool
	assignCalled bool
}

func (m *mockPostgres) SavePullRequest(event github.PullRequestEvent) (int, error) {
	m.saveCalled = true
	return 1, nil
}

func (m *mockPostgres) SelectReviewer(author string) (string, error) {
	return "reviewer1", nil
}

func (m *mockPostgres) AssignReviewer(prID int, reviewer string) error {
	m.assignCalled = true
	return nil
}

func (m *mockPostgres) NotifyReviewer(reviewer string, message string, bot interface{}) {}

func TestHandlePR(t *testing.T) {
	event := github.PullRequestEvent{
		Action: "opened",
	}
	event.PullRequest.ID = 10
	event.PullRequest.User.Login = "author"
	event.PullRequest.HTML = "http://github/pr"
	event.Repository.FullName = "test/repo"

	mockDB := &mockPostgres{}
	bot := &mockBot{}

	prID, err := mockDB.SavePullRequest(event)
	require.NoError(t, err)
	require.Equal(t, 1, prID)

	reviewer, err := mockDB.SelectReviewer(event.PullRequest.User.Login)
	require.NoError(t, err)
	require.NotEqual(t, "author", reviewer)

	err = mockDB.AssignReviewer(prID, reviewer)
	require.NoError(t, err)

	bot.Notify(1, "test")

	require.True(t, mockDB.saveCalled)
	require.True(t, mockDB.assignCalled)
	require.True(t, bot.called)
}
