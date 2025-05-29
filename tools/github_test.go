package tools

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetOpenIssues(t *testing.T) {
	// Skip if no GitHub token is set
	if os.Getenv("GITHUB_TOKEN") == "" {
		t.Skip("GITHUB_TOKEN not set")
	}

	// Create test input for a real repository
	input := ToolInput{
		Owner: "golang",
		Repo:  "go",
		State: "open",
	}
	rawInput, _ := json.Marshal(input)

	// Test the function with real GitHub API
	ctx := context.Background()
	issues, err := GetOpenIssues(ctx, rawInput)

	// Verify results
	assert.NoError(t, err)
	assert.True(t, len(issues) > 0, "Expected at least one open issue")

	// Verify the structure of returned issues
	for _, issue := range issues {
		assert.NotNil(t, issue.Number)
		assert.NotNil(t, issue.Title)
		assert.NotNil(t, issue.State)
		assert.Equal(t, "open", *issue.State)
	}
}

func TestGetOpenPRs(t *testing.T) {
	// Skip if no GitHub token is set
	if os.Getenv("GITHUB_TOKEN") == "" {
		t.Skip("GITHUB_TOKEN not set")
	}

	// Create test input for a real repository
	input := ToolInput{
		Owner: "golang",
		Repo:  "go", 
		State: "open",
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

func TestGetOpenIssuesInvalidJSON(t *testing.T) {
    ctx := context.Background()
    invalidJSON := json.RawMessage(`{"invalid": json}`)
    
    _, err := GetOpenIssues(ctx, invalidJSON)
    assert.Error(t, err)
}

func TestGetOpenIssuesInvalidRepo(t *testing.T) {
    if os.Getenv("GITHUB_TOKEN") == "" {
        t.Skip("GITHUB_TOKEN not set")
    }
    
    input := ToolInput{
        Owner: "nonexistent",
        Repo:  "nonexistent-repo-12345",
        State: "open",
    }
    rawInput, _ := json.Marshal(input)
    
    ctx := context.Background()
    _, err := GetOpenIssues(ctx, rawInput)
    assert.Error(t, err)
}

func TestGetOpenPRsDefaultState(t *testing.T) {
    if os.Getenv("GITHUB_TOKEN") == "" {
        t.Skip("GITHUB_TOKEN not set")
    }
    
    input := ToolInput{
        Owner: "golang",
        Repo:  "go",
        // State is empty, should default to "open"
    }
    rawInput, _ := json.Marshal(input)
    
    ctx := context.Background()
    prs, err := GetOpenPRs(ctx, rawInput)
    
    assert.NoError(t, err)
    for _, pr := range prs {
        assert.Equal(t, "open", *pr.State)
    }
}