package http

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/fncg/ReviewerService/internal/storage"
	"github.com/fncg/ReviewerService/internal/telegram"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/require"
)

type mockTelegramBot struct {
	notifications []struct {
		chatID  int64
		message string
	}
}

var _ telegram.Notifier = (*mockTelegramBot)(nil)

func (m *mockTelegramBot) Notify(chatID int64, message string) {
	m.notifications = append(m.notifications, struct {
		chatID  int64
		message string
	}{chatID, message})
}

func TestUserStory_PullRequestAutoAssignmentIntegration(t *testing.T) {

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://reviewer:reviewer@localhost:5432/reviewer?sslmode=disable"
	} else {
		if !strings.Contains(dsn, "sslmode") {
			if strings.Contains(dsn, "?") {
				dsn += "&sslmode=disable"
			} else {
				dsn += "?sslmode=disable"
			}
		}
	}
	db, err := storage.NewPostgres(dsn)
	require.NoError(t, err, "Не удалось подключиться к БД")
	defer db.Close()

	mockBot := &mockTelegramBot{
		notifications: make([]struct {
			chatID  int64
			message string
		}, 0),
	}

	server := NewServer(db, mockBot)

	ctx := context.Background()
	conn, err := sql.Open("pgx", dsn)
	require.NoError(t, err)
	defer conn.Close()

	type userData struct {
		githubLogin   string
		telegramChatID int64
	}
	var originalUsers []userData
	rows, err := conn.QueryContext(ctx, `SELECT github_login, telegram_chat_id FROM users`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var u userData
			if err := rows.Scan(&u.githubLogin, &u.telegramChatID); err == nil {
				originalUsers = append(originalUsers, u)
			}
		}
	}

	_, err = conn.ExecContext(ctx, `
		DELETE FROM review_assignments;
		DELETE FROM pull_requests;
		DELETE FROM repositories;
		DELETE FROM users;
	`)
	require.NoError(t, err, "Не удалось очистить тестовые данные")

	_, err = conn.ExecContext(ctx, `
		INSERT INTO users (github_login, telegram_chat_id) VALUES 
		('reviewer1', 123456),
		('reviewer2', 789012),
		('reviewer3', 345678)
	`)
	require.NoError(t, err, "Не удалось добавить тестовых пользователей")

	prEvent := map[string]interface{}{
		"action": "opened",
		"pull_request": map[string]interface{}{
			"id":       12345,
			"title":    "Test Pull Request",
			"html_url": "https://github.com/test/repo/pull/1",
			"user": map[string]interface{}{
				"login": "author",
			},
		},
		"repository": map[string]interface{}{
			"full_name": "test/repo",
		},
	}

	jsonBody, err := json.Marshal(prEvent)
	require.NoError(t, err, "Не удалось сериализовать JSON")

	req := httptest.NewRequest(http.MethodPost, "/github/webhook", bytes.NewBuffer(jsonBody))
	req.Header.Set("X-GitHub-Event", "pull_request")
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.githubWebhook(w, req)

	require.Equal(t, http.StatusOK, w.Code, "Webhook должен вернуть 200 OK")

	var prID int
	var githubPRID int64
	var authorLogin string
	var repoID int
	err = conn.QueryRowContext(ctx, `
		SELECT pr.id, pr.github_pr_id, pr.author_login, pr.repository_id 
		FROM pull_requests pr 
		WHERE pr.github_pr_id = $1
	`, 12345).Scan(&prID, &githubPRID, &authorLogin, &repoID)
	require.NoError(t, err, "PR должен быть сохранен в БД")
	require.Equal(t, int64(12345), githubPRID, "GitHub PR ID должен совпадать")
	require.Equal(t, "author", authorLogin, "Автор должен быть сохранен корректно")
	require.Greater(t, prID, 0, "PR ID должен быть больше 0")
	require.Greater(t, repoID, 0, "Repository ID должен быть больше 0")

	var repoFullName string
	err = conn.QueryRowContext(ctx, `
		SELECT full_name FROM repositories WHERE id = $1
	`, repoID).Scan(&repoFullName)
	require.NoError(t, err, "Репозиторий должен быть сохранен")
	require.Equal(t, "test/repo", repoFullName, "Имя репозитория должно совпадать")

	var reviewerLogin string
	var assignmentID int
	err = conn.QueryRowContext(ctx, `
		SELECT id, reviewer_login 
		FROM review_assignments 
		WHERE pr_id = $1
	`, prID).Scan(&assignmentID, &reviewerLogin)
	require.NoError(t, err, "Ревьюер должен быть назначен")
	require.NotEmpty(t, reviewerLogin, "Логин ревьюера должен быть заполнен")
	require.NotEqual(t, "author", reviewerLogin, "Ревьюер не должен быть автором PR")
	require.Contains(t, []string{"reviewer1", "reviewer2", "reviewer3"}, reviewerLogin,
		"Ревьюер должен быть одним из доступных ревьюеров")

	require.Equal(t, 1, len(mockBot.notifications), "Должно быть отправлено одно уведомление")

	notification := mockBot.notifications[0]
	require.Greater(t, notification.chatID, int64(0), "Chat ID должен быть задан")
	require.NotEmpty(t, notification.message, "Сообщение не должно быть пустым")
	require.Contains(t, notification.message, "You have been assigned to review PR:",
		"Сообщение должно содержать информацию о назначении")
	require.Contains(t, notification.message, "https://github.com/test/repo/pull/1",
		"Сообщение должно содержать ссылку на PR")

	var expectedChatID int64
	err = conn.QueryRowContext(ctx, `
		SELECT telegram_chat_id FROM users WHERE github_login = $1
	`, reviewerLogin).Scan(&expectedChatID)
	require.NoError(t, err, "Должен быть найден chat ID для ревьюера")
	require.Equal(t, expectedChatID, notification.chatID,
		"Chat ID в уведомлении должен соответствовать chat ID ревьюера")

	_, err = conn.ExecContext(ctx, `
		DELETE FROM review_assignments;
		DELETE FROM pull_requests;
		DELETE FROM repositories;
		DELETE FROM users;
	`)
	if err == nil {
		if len(originalUsers) > 0 {
			for _, u := range originalUsers {
				conn.ExecContext(ctx, `
					INSERT INTO users (github_login, telegram_chat_id) 
					VALUES ($1, $2)
					ON CONFLICT (github_login) DO NOTHING
				`, u.githubLogin, u.telegramChatID)
			}
		} else {
			conn.ExecContext(ctx, `
				INSERT INTO users (github_login, telegram_chat_id) VALUES 
				('reviewer1', 1009163017),
				('reviewer2', 1009163017),
				('reviewer3', 1009163017)
				ON CONFLICT (github_login) DO NOTHING
			`)
		}
	}
}
