package storage

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSelectReviewerLogic(t *testing.T) {
	author := "author"
	users := []string{"author", "reviewer1", "reviewer2"}

	var selected string
	for _, u := range users {
		if u != author {
			selected = u
			break
		}
	}

	require.NotEmpty(t, selected)
	require.NotEqual(t, author, selected)
}
