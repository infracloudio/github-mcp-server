package tools

import (
	"context"
	"encoding/json"
	"os"

	"github.com/google/go-github/v56/github"
	"golang.org/x/oauth2"
)

var GitHubClient = func(ctx context.Context) *github.Client {
	token := os.Getenv("GITHUB_TOKEN")
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)
	return github.NewClient(tc)
}

type ToolInput struct {
	Owner string `json:"owner"`
	Repo  string `json:"repo"`
	State string `json:"state"`
}

func GetOpenIssues(ctx context.Context, input json.RawMessage) ([]*github.Issue, error) {
	var params ToolInput
	if err := json.Unmarshal(input, &params); err != nil {
		return nil, err
	}

	client := GitHubClient(ctx)
	if params.State == "" {
		params.State = "open"
	}
	issues, _, err := client.Issues.ListByRepo(ctx, params.Owner, params.Repo, &github.IssueListByRepoOptions{
		State: params.State,
	})
	if err != nil {
		return nil, err
	}

	return issues, nil
}

func GetOpenPRs(ctx context.Context, input json.RawMessage) ([]*github.PullRequest, error) {
	var params ToolInput
	if err := json.Unmarshal(input, &params); err != nil {
		return nil, err
	}

	client := GitHubClient(ctx)
	if params.State == "" {
		params.State = "open"
	}
	prs, _, err := client.PullRequests.List(ctx, params.Owner, params.Repo, &github.PullRequestListOptions{
		State: params.State,
	})
	if err != nil {
		return nil, err
	}

	return prs, nil
}
