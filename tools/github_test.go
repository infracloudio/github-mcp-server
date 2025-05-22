package tools

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetOpenPRs(t *testing.T) {
	// Skip if no GitHub token is set
	if os.Getenv("GITHUB_TOKEN") == "" {
		t.Skip("GITHUB_TOKEN not set")
	}

	// Create test input for a real repository
	input := ToolInput{
		Owner: "golang",
		Repo:  "go",
	}
	rawInput, _ := json.Marshal(input)

	// Test the function with real GitHub API
	ctx := context.Background()
	prs, err := GetOpenPRs(ctx, rawInput)

	// Verify results
	assert.NoError(t, err)
	assert.True(t, len(prs) > 0, "Expected at least one open PR")

	// Verify the structure of returned PRs
	for _, pr := range prs {
		assert.NotNil(t, pr.Number)
		assert.NotNil(t, pr.Title)
		assert.NotNil(t, pr.State)
		assert.Equal(t, "open", *pr.State)
	}
}
