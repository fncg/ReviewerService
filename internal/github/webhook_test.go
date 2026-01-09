package github

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPullRequestEvent_UnmarshalJSON_Success(t *testing.T) {
	jsonData := `{
		"action": "opened",
		"pull_request": {
			"id": 123,
			"title": "Test PR Title",
			"html_url": "https://github.com/org/repo/pull/123",
			"user": {
				"login": "testuser"
			}
		},
		"repository": {
			"full_name": "org/repo"
		}
	}`

	var event PullRequestEvent
	err := json.Unmarshal([]byte(jsonData), &event)

	require.NoError(t, err)
	require.Equal(t, "opened", event.Action)
	require.Equal(t, int64(123), event.PullRequest.ID)
	require.Equal(t, "Test PR Title", event.PullRequest.Title)
	require.Equal(t, "https://github.com/org/repo/pull/123", event.PullRequest.HTML)
	require.Equal(t, "testuser", event.PullRequest.User.Login)
	require.Equal(t, "org/repo", event.Repository.FullName)
}

func TestPullRequestEvent_UnmarshalJSON_AllActions(t *testing.T) {
	actions := []string{"opened", "closed", "reopened", "synchronize", "assigned", "unassigned", "labeled", "unlabeled"}

	for _, action := range actions {
		jsonData := `{
			"action": "` + action + `",
			"pull_request": {
				"id": 123,
				"title": "Test PR",
				"html_url": "https://github.com/org/repo/pull/123",
				"user": {"login": "testuser"}
			},
			"repository": {"full_name": "org/repo"}
		}`

		var event PullRequestEvent
		err := json.Unmarshal([]byte(jsonData), &event)

		require.NoError(t, err, "Action %s should unmarshal successfully", action)
		require.Equal(t, action, event.Action)
	}
}

func TestPullRequestEvent_UnmarshalJSON_MinimalFields(t *testing.T) {
	jsonData := `{
		"action": "opened",
		"pull_request": {
			"id": 456
		},
		"repository": {}
	}`

	var event PullRequestEvent
	err := json.Unmarshal([]byte(jsonData), &event)

	require.NoError(t, err)
	require.Equal(t, "opened", event.Action)
	require.Equal(t, int64(456), event.PullRequest.ID)
	require.Equal(t, "", event.PullRequest.Title)
	require.Equal(t, "", event.PullRequest.HTML)
	require.Equal(t, "", event.PullRequest.User.Login)
	require.Equal(t, "", event.Repository.FullName)
}

func TestPullRequestEvent_UnmarshalJSON_InvalidJSON(t *testing.T) {
	invalidJSONs := []string{
		"{",
		"{invalid}",
		`{"action":}`,
		`{"action": "opened"`,
		"not json",
		"",
	}

	for _, invalidJSON := range invalidJSONs {
		var event PullRequestEvent
		err := json.Unmarshal([]byte(invalidJSON), &event)
		require.Error(t, err, "Invalid JSON should return error: %s", invalidJSON)
	}
}

func TestPullRequestEvent_UnmarshalJSON_MissingFields(t *testing.T) {
	jsonData1 := `{
		"action": "opened",
		"repository": {"full_name": "org/repo"}
	}`

	var event1 PullRequestEvent
	err := json.Unmarshal([]byte(jsonData1), &event1)
	require.NoError(t, err)
	require.Equal(t, "opened", event1.Action)
	require.Equal(t, int64(0), event1.PullRequest.ID)

	jsonData2 := `{
		"action": "opened",
		"pull_request": {
			"id": 123,
			"title": "Test PR"
		}
	}`

	var event2 PullRequestEvent
	err = json.Unmarshal([]byte(jsonData2), &event2)
	require.NoError(t, err)
	require.Equal(t, "opened", event2.Action)
	require.Equal(t, int64(123), event2.PullRequest.ID)
	require.Equal(t, "", event2.Repository.FullName)
}

func TestPullRequestEvent_UnmarshalJSON_RealWorldExample(t *testing.T) {
	jsonData := `{
		"action": "opened",
		"number": 42,
		"pull_request": {
			"id": 123456789,
			"number": 42,
			"title": "Add feature X",
			"html_url": "https://github.com/owner/repo/pull/42",
			"state": "open",
			"user": {
				"login": "developer",
				"id": 12345,
				"type": "User"
			},
			"body": "This PR adds feature X",
			"created_at": "2024-01-01T00:00:00Z",
			"updated_at": "2024-01-01T00:00:00Z"
		},
		"repository": {
			"id": 987654321,
			"name": "repo",
			"full_name": "owner/repo",
			"private": false
		},
		"sender": {
			"login": "developer",
			"id": 12345
		}
	}`

	var event PullRequestEvent
	err := json.Unmarshal([]byte(jsonData), &event)

	require.NoError(t, err)
	require.Equal(t, "opened", event.Action)
	require.Equal(t, int64(123456789), event.PullRequest.ID)
	require.Equal(t, "Add feature X", event.PullRequest.Title)
	require.Equal(t, "https://github.com/owner/repo/pull/42", event.PullRequest.HTML)
	require.Equal(t, "developer", event.PullRequest.User.Login)
	require.Equal(t, "owner/repo", event.Repository.FullName)
}

func TestPullRequestEvent_UnmarshalJSON_SpecialCharacters(t *testing.T) {
	jsonData := `{
		"action": "opened",
		"pull_request": {
			"id": 123,
			"title": "PR with special chars: !@#$%^&*()",
			"html_url": "https://github.com/org/repo/pull/123",
			"user": {"login": "user-with-dashes_123"}
		},
		"repository": {"full_name": "org/repo-name.with.dots"}
	}`

	var event PullRequestEvent
	err := json.Unmarshal([]byte(jsonData), &event)

	require.NoError(t, err)
	require.Equal(t, "PR with special chars: !@#$%^&*()", event.PullRequest.Title)
	require.Equal(t, "user-with-dashes_123", event.PullRequest.User.Login)
	require.Equal(t, "org/repo-name.with.dots", event.Repository.FullName)
}

func TestPullRequestEvent_UnmarshalJSON_UnicodeCharacters(t *testing.T) {
	jsonData := `{
		"action": "opened",
		"pull_request": {
			"id": 123,
			"title": "PR with unicode: Привет こんにちは 🚀",
			"html_url": "https://github.com/org/repo/pull/123",
			"user": {"login": "user_123"}
		},
		"repository": {"full_name": "org/repo"}
	}`

	var event PullRequestEvent
	err := json.Unmarshal([]byte(jsonData), &event)

	require.NoError(t, err)
	require.Equal(t, "PR with unicode: Привет こんにちは 🚀", event.PullRequest.Title)
}

func TestPullRequestEvent_MarshalJSON(t *testing.T) {
	event := PullRequestEvent{
		Action: "opened",
	}
	event.PullRequest.ID = 123
	event.PullRequest.Title = "Test PR"
	event.PullRequest.HTML = "https://github.com/org/repo/pull/123"
	event.PullRequest.User.Login = "testuser"
	event.Repository.FullName = "org/repo"

	data, err := json.Marshal(event)

	require.NoError(t, err)
	require.NotEmpty(t, data)

	var unmarshaled PullRequestEvent
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)
	require.Equal(t, event.Action, unmarshaled.Action)
	require.Equal(t, event.PullRequest.ID, unmarshaled.PullRequest.ID)
	require.Equal(t, event.PullRequest.Title, unmarshaled.PullRequest.Title)
	require.Equal(t, event.PullRequest.HTML, unmarshaled.PullRequest.HTML)
	require.Equal(t, event.PullRequest.User.Login, unmarshaled.PullRequest.User.Login)
	require.Equal(t, event.Repository.FullName, unmarshaled.Repository.FullName)
}

func TestPullRequestEvent_ZeroValues(t *testing.T) {
	var event PullRequestEvent

	require.Equal(t, "", event.Action)
	require.Equal(t, int64(0), event.PullRequest.ID)
	require.Equal(t, "", event.PullRequest.Title)
	require.Equal(t, "", event.PullRequest.HTML)
	require.Equal(t, "", event.PullRequest.User.Login)
	require.Equal(t, "", event.Repository.FullName)
}

func TestPullRequestEvent_LargeID(t *testing.T) {
	jsonData := `{
		"action": "opened",
		"pull_request": {
			"id": 9223372036854775807,
			"title": "Test PR",
			"html_url": "https://github.com/org/repo/pull/123",
			"user": {"login": "testuser"}
		},
		"repository": {"full_name": "org/repo"}
	}`

	var event PullRequestEvent
	err := json.Unmarshal([]byte(jsonData), &event)

	require.NoError(t, err)
	require.Equal(t, int64(9223372036854775807), event.PullRequest.ID)
}
