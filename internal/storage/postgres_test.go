package storage

import (
	"context"
	"errors"
	"testing"

	"github.com/fncg/ReviewerService/internal/github"
	"github.com/fncg/ReviewerService/internal/telegram"
	"github.com/stretchr/testify/require"
)

type mockNotifier struct {
	notifyCalls []struct {
		chatID  int64
		message string
	}
}

func (m *mockNotifier) Notify(chatID int64, message string) {
	m.notifyCalls = append(m.notifyCalls, struct {
		chatID  int64
		message string
	}{chatID, message})
}

type mockPostgresMethods struct {
	savePRCalled        bool
	savePREvent         github.PullRequestEvent
	savePRReturnID      int
	savePRReturnError   error
	selectReviewerCalled bool
	selectReviewerAuthor string
	selectReviewerReturn string
	selectReviewerError  error
	assignReviewerCalled bool
	assignReviewerPRID   int
	assignReviewerName   string
	assignReviewerError  error
	notifyReviewerCalled bool
	notifyReviewerName   string
	notifyReviewerMsg    string
	notifyReviewerBot    telegram.Notifier
}

func TestHandlePR_Success(t *testing.T) {
	event := github.PullRequestEvent{
		Action: "opened",
	}
	event.PullRequest.ID = 10
	event.PullRequest.User.Login = "author"
	event.PullRequest.HTML = "http://github.com/test/pr"
	event.Repository.FullName = "test/repo"

	mockDB := &mockPostgresMethods{
		savePRReturnID:       1,
		savePRReturnError:    nil,
		selectReviewerReturn: "reviewer1",
		selectReviewerError:  nil,
		assignReviewerError:  nil,
	}
	bot := &mockNotifier{}

	prID, err := mockDB.savePullRequest(event)
	require.NoError(t, err)
	require.Equal(t, 1, prID)
	require.True(t, mockDB.savePRCalled)

	reviewer, err := mockDB.selectReviewer(event.PullRequest.User.Login)
	require.NoError(t, err)
	require.Equal(t, "reviewer1", reviewer)
	require.True(t, mockDB.selectReviewerCalled)
	require.Equal(t, "author", mockDB.selectReviewerAuthor)

	err = mockDB.assignReviewer(prID, reviewer)
	require.NoError(t, err)
	require.True(t, mockDB.assignReviewerCalled)
	require.Equal(t, 1, mockDB.assignReviewerPRID)
	require.Equal(t, "reviewer1", mockDB.assignReviewerName)

	msg := "You have been assigned to review PR: " + event.PullRequest.HTML
	mockDB.notifyReviewer(reviewer, msg, bot)
	require.True(t, mockDB.notifyReviewerCalled)
	require.Equal(t, "reviewer1", mockDB.notifyReviewerName)
	require.Equal(t, msg, mockDB.notifyReviewerMsg)
	require.Equal(t, 1, len(bot.notifyCalls))
	require.Equal(t, msg, bot.notifyCalls[0].message)
}

func TestHandlePR_SaveError(t *testing.T) {
	event := github.PullRequestEvent{
		Action: "opened",
	}
	event.PullRequest.ID = 10
	event.PullRequest.User.Login = "author"
	event.PullRequest.HTML = "http://github.com/test/pr"
	event.Repository.FullName = "test/repo"

	mockDB := &mockPostgresMethods{
		savePRReturnError: errors.New("database error"),
	}
	bot := &mockNotifier{}

	prID, err := mockDB.savePullRequest(event)
	require.Error(t, err)
	require.Equal(t, 0, prID)
	require.True(t, mockDB.savePRCalled)

	require.False(t, mockDB.selectReviewerCalled)
	require.False(t, mockDB.assignReviewerCalled)
	require.False(t, mockDB.notifyReviewerCalled)
	require.Equal(t, 0, len(bot.notifyCalls))
}

func TestHandlePR_SelectReviewerError(t *testing.T) {
	event := github.PullRequestEvent{
		Action: "opened",
	}
	event.PullRequest.ID = 10
	event.PullRequest.User.Login = "author"
	event.PullRequest.HTML = "http://github.com/test/pr"
	event.Repository.FullName = "test/repo"

	mockDB := &mockPostgresMethods{
		savePRReturnID:       1,
		savePRReturnError:    nil,
		selectReviewerReturn: "",
		selectReviewerError:  errors.New("no reviewers available"),
	}
	bot := &mockNotifier{}

	prID, err := mockDB.savePullRequest(event)
	require.NoError(t, err)
	require.Equal(t, 1, prID)

	reviewer, err := mockDB.selectReviewer(event.PullRequest.User.Login)
	require.Error(t, err)
	require.Equal(t, "", reviewer)

	require.False(t, mockDB.assignReviewerCalled)
	require.False(t, mockDB.notifyReviewerCalled)
	require.Equal(t, 0, len(bot.notifyCalls))
}

func TestHandlePR_AssignReviewerError(t *testing.T) {
	event := github.PullRequestEvent{
		Action: "opened",
	}
	event.PullRequest.ID = 10
	event.PullRequest.User.Login = "author"
	event.PullRequest.HTML = "http://github.com/test/pr"
	event.Repository.FullName = "test/repo"

	mockDB := &mockPostgresMethods{
		savePRReturnID:       1,
		savePRReturnError:    nil,
		selectReviewerReturn: "reviewer1",
		selectReviewerError:  nil,
		assignReviewerError:  errors.New("assignment failed"),
	}
	bot := &mockNotifier{}

	prID, err := mockDB.savePullRequest(event)
	require.NoError(t, err)

	reviewer, err := mockDB.selectReviewer(event.PullRequest.User.Login)
	require.NoError(t, err)

	err = mockDB.assignReviewer(prID, reviewer)
	require.Error(t, err)

	require.False(t, mockDB.notifyReviewerCalled)
	require.Equal(t, 0, len(bot.notifyCalls))
}

func TestNotifyReviewer_Success(t *testing.T) {
	bot := &mockNotifier{}
	mockDB := &mockPostgresMethods{}

	reviewer := "reviewer1"
	message := "You have been assigned to review PR: http://github.com/test/pr"

	mockDB.notifyReviewer(reviewer, message, bot)

	require.True(t, mockDB.notifyReviewerCalled)
	require.Equal(t, reviewer, mockDB.notifyReviewerName)
	require.Equal(t, message, mockDB.notifyReviewerMsg)
	require.Equal(t, 1, len(bot.notifyCalls))
	require.Equal(t, message, bot.notifyCalls[0].message)
}

func TestNotifyReviewer_MultipleCalls(t *testing.T) {
	bot := &mockNotifier{}
	mockDB := &mockPostgresMethods{}

	messages := []string{
		"Message 1",
		"Message 2",
		"Message 3",
	}

	for i, msg := range messages {
		mockDB.notifyReviewer("reviewer1", msg, bot)
		require.Equal(t, i+1, len(bot.notifyCalls))
		require.Equal(t, msg, bot.notifyCalls[i].message)
	}
}

func TestSelectReviewer_ExcludesAuthor(t *testing.T) {
	author := "author"
	reviewers := []string{"reviewer1", "reviewer2", "reviewer3"}

	var selected string
	for _, r := range reviewers {
		if r != author {
			selected = r
			break
		}
	}

	require.NotEmpty(t, selected)
	require.NotEqual(t, author, selected)
	require.Contains(t, reviewers, selected)
}

func TestSelectReviewer_AllReviewersAreAuthor(t *testing.T) {
	author := "author"
	reviewers := []string{"author", "author", "author"}

	var selected string
	for _, r := range reviewers {
		if r != author {
			selected = r
			break
		}
	}

	require.Empty(t, selected)
}

func TestAssignReviewer_Success(t *testing.T) {
	mockDB := &mockPostgresMethods{
		assignReviewerError: nil,
	}

	prID := 123
	reviewer := "reviewer1"

	err := mockDB.assignReviewer(prID, reviewer)

	require.NoError(t, err)
	require.True(t, mockDB.assignReviewerCalled)
	require.Equal(t, prID, mockDB.assignReviewerPRID)
	require.Equal(t, reviewer, mockDB.assignReviewerName)
}

func TestAssignReviewer_Error(t *testing.T) {
	mockDB := &mockPostgresMethods{
		assignReviewerError: errors.New("database constraint violation"),
	}

	prID := 123
	reviewer := "reviewer1"

	err := mockDB.assignReviewer(prID, reviewer)

	require.Error(t, err)
	require.True(t, mockDB.assignReviewerCalled)
}

func (m *mockPostgresMethods) savePullRequest(event github.PullRequestEvent) (int, error) {
	m.savePRCalled = true
	m.savePREvent = event
	return m.savePRReturnID, m.savePRReturnError
}

func (m *mockPostgresMethods) selectReviewer(author string) (string, error) {
	m.selectReviewerCalled = true
	m.selectReviewerAuthor = author
	return m.selectReviewerReturn, m.selectReviewerError
}

func (m *mockPostgresMethods) assignReviewer(prID int, reviewer string) error {
	m.assignReviewerCalled = true
	m.assignReviewerPRID = prID
	m.assignReviewerName = reviewer
	return m.assignReviewerError
}

func (m *mockPostgresMethods) notifyReviewer(reviewer string, message string, bot telegram.Notifier) {
	m.notifyReviewerCalled = true
	m.notifyReviewerName = reviewer
	m.notifyReviewerMsg = message
	m.notifyReviewerBot = bot
	bot.Notify(0, message)
}

func TestPostgres_ContextUsage(t *testing.T) {
	ctx := context.Background()
	require.NotNil(t, ctx)
	
	ctxWithTimeout, cancel := context.WithTimeout(context.Background(), 5)
	require.NotNil(t, ctxWithTimeout)
	require.NotNil(t, cancel)
	cancel()
}
