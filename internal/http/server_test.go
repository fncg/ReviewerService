package http

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewServer(t *testing.T) {
	server := NewServer(nil, nil)

	require.NotNil(t, server)
	require.NotNil(t, server.mux)
	require.Nil(t, server.db)
	require.Nil(t, server.bot)
}

func TestNewServer_NilParams(t *testing.T) {
	server := NewServer(nil, nil)

	require.NotNil(t, server)
	require.NotNil(t, server.mux)
	require.Nil(t, server.db)
	require.Nil(t, server.bot)
}

func TestHealthHandler(t *testing.T) {
	server := NewServer(nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	server.health(w, req)

	resp := w.Result()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	defer resp.Body.Close()

	body := w.Body.String()
	require.Equal(t, "ok", body)
}

func TestHealthHandler_DifferentMethods(t *testing.T) {
	server := NewServer(nil, nil)
	methods := []string{http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch}

	for _, method := range methods {
		req := httptest.NewRequest(method, "/health", nil)
		w := httptest.NewRecorder()

		server.health(w, req)

		require.Equal(t, http.StatusOK, w.Code, "Method %s should return OK", method)
	}
}

func TestWebhookHandler_WrongMethod(t *testing.T) {
	server := NewServer(nil, nil)
	methods := []string{http.MethodGet, http.MethodPut, http.MethodDelete, http.MethodPatch}

	for _, method := range methods {
		req := httptest.NewRequest(method, "/github/webhook", nil)
		w := httptest.NewRecorder()

		server.githubWebhook(w, req)

		require.Equal(t, http.StatusMethodNotAllowed, w.Code, "Method %s should return MethodNotAllowed", method)
	}
}

func TestWebhookHandler_WrongEvent(t *testing.T) {
	server := NewServer(nil, nil)
	events := []string{"push", "issues", "pull_request_review", "ping", ""}

	for _, event := range events {
		req := httptest.NewRequest(http.MethodPost, "/github/webhook", nil)
		req.Header.Set("X-GitHub-Event", event)
		w := httptest.NewRecorder()

		server.githubWebhook(w, req)

		require.Equal(t, http.StatusOK, w.Code, "Event %s should return OK", event)
	}
}

func TestWebhookHandler_InvalidJSON(t *testing.T) {
	server := NewServer(nil, nil)
	invalidJSONs := []string{
		"invalid json",
		"{",
		"{invalid}",
		`{"action":}`,
		`{"action": "opened"`,
	}

	for _, invalidJSON := range invalidJSONs {
		req := httptest.NewRequest(http.MethodPost, "/github/webhook", bytes.NewBuffer([]byte(invalidJSON)))
		req.Header.Set("X-GitHub-Event", "pull_request")
		w := httptest.NewRecorder()

		server.githubWebhook(w, req)

		require.Equal(t, http.StatusBadRequest, w.Code, "Invalid JSON should return BadRequest: %s", invalidJSON)
	}
}

func TestWebhookHandler_EmptyBody(t *testing.T) {
	server := NewServer(nil, nil)

	req := httptest.NewRequest(http.MethodPost, "/github/webhook", bytes.NewBuffer([]byte("")))
	req.Header.Set("X-GitHub-Event", "pull_request")
	w := httptest.NewRecorder()

	server.githubWebhook(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestWebhookHandler_NonOpenedAction(t *testing.T) {
	server := NewServer(nil, nil)
	actions := []string{"closed", "synchronize", "reopened", "assigned", "unassigned"}

	for _, action := range actions {
		jsonBody := []byte(`{
			"action": "` + action + `",
			"pull_request": {
				"id": 1,
				"title": "Test PR",
				"html_url": "http://github.com/test/pr",
				"user": {"login": "author"}
			},
			"repository": {"full_name": "test/repo"}
		}`)

		req := httptest.NewRequest(http.MethodPost, "/github/webhook", bytes.NewBuffer(jsonBody))
		req.Header.Set("X-GitHub-Event", "pull_request")
		w := httptest.NewRecorder()

		server.githubWebhook(w, req)

		require.Equal(t, http.StatusOK, w.Code, "Action %s should return OK", action)
	}
}

func TestWebhookHandler_ValidJSONParsing_NonOpenedActions(t *testing.T) {
	server := NewServer(nil, nil)
	
	testCases := []struct {
		name   string
		action string
	}{
		{"closed", "closed"},
		{"reopened", "reopened"},
		{"synchronize", "synchronize"},
		{"assigned", "assigned"},
		{"unassigned", "unassigned"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			jsonBody := []byte(`{
				"action": "` + tc.action + `",
				"pull_request": {"id": 1, "title": "Test", "html_url": "http://test", "user": {"login": "user"}},
				"repository": {"full_name": "org/repo"}
			}`)

			req := httptest.NewRequest(http.MethodPost, "/github/webhook", bytes.NewBuffer(jsonBody))
			req.Header.Set("X-GitHub-Event", "pull_request")
			w := httptest.NewRecorder()

			server.githubWebhook(w, req)
			require.Equal(t, http.StatusOK, w.Code, "Action %s should return OK", tc.action)
		})
	}
}

func TestWebhookHandler_JSONStructureValidation(t *testing.T) {
	server := NewServer(nil, nil)
	
	testCases := []struct {
		name        string
		jsonBody    string
		expectPanic bool
		expectCode  int
	}{
		{
			name: "valid opened action",
			jsonBody: `{
				"action": "opened",
				"pull_request": {"id": 123, "title": "Test PR", "html_url": "http://test", "user": {"login": "author"}},
				"repository": {"full_name": "org/repo"}
			}`,
			expectPanic: true,
		},
		{
			name: "valid closed action",
			jsonBody: `{
				"action": "closed",
				"pull_request": {"id": 123, "title": "Test PR", "html_url": "http://test", "user": {"login": "author"}},
				"repository": {"full_name": "org/repo"}
			}`,
			expectPanic: false,
			expectCode:  http.StatusOK,
		},
		{
			name: "minimal valid JSON",
			jsonBody: `{
				"action": "opened",
				"pull_request": {"id": 1},
				"repository": {}
			}`,
			expectPanic: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/github/webhook", bytes.NewBuffer([]byte(tc.jsonBody)))
			req.Header.Set("X-GitHub-Event", "pull_request")
			w := httptest.NewRecorder()

			if tc.expectPanic {
				require.Panics(t, func() {
					server.githubWebhook(w, req)
				}, "Should panic when db is nil and action is opened")
			} else {
				server.githubWebhook(w, req)
				require.Equal(t, tc.expectCode, w.Code)
			}
		})
	}
}
