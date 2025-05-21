package tools

import (
	"context"
	"encoding/json"
	"os"

	"github.com/google/go-github/v56/github"
	"golang.org/x/oauth2"
)

func GitHubClient(ctx context.Context) *github.Client {
	token := os.Getenv("GITHUB_TOKEN")
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)
	return github.NewClient(tc)
}

type ToolInput struct {
	Owner string `json:"owner"`
	Repo  string `json:"repo"`
}

func GetOpenPRs(ctx context.Context, input json.RawMessage) (any, error) {
	var params ToolInput
	if err := json.Unmarshal(input, &params); err != nil {
		return nil, err
	}
	client := GitHubClient(ctx)
	prs, _, err := client.PullRequests.List(ctx, params.Owner, params.Repo, nil)
	if err != nil {
		return nil, err
	}
	return prs, nil
}
