package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

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
	Owner    string   `json:"owner"`
	Repo     string   `json:"repo"`
	State    string   `json:"state"`
	Query    string   `json:"query"`
	Title    string   `json:"title"`
	Body     string   `json:"body"`
	Labels   []string `json:"labels"`
	Assignee string   `json:"assignee"`
	Limit    int      `json:"limit"`
	Prioritize bool   `json:"prioritize"`
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
		ListOptions: github.ListOptions{PerPage: 100},
	})
	if err != nil {
		return nil, err
	}

	// Filter out pull requests (GitHub API returns PRs as issues)
	var actualIssues []*github.Issue
	for _, issue := range issues {
		if !issue.IsPullRequest() {
			actualIssues = append(actualIssues, issue)
		}
	}

	return actualIssues, nil
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
		ListOptions: github.ListOptions{PerPage: 100},
	})
	if err != nil {
		return nil, err
	}

	return prs, nil
}

// SearchIssues searches for issues by keyword/topic in title and body
func SearchIssues(ctx context.Context, input json.RawMessage) ([]*github.Issue, error) {
	var params ToolInput
	if err := json.Unmarshal(input, &params); err != nil {
		return nil, err
	}

	client := GitHubClient(ctx)
	if params.State == "" {
		params.State = "open"
	}

	// Build search query for GitHub's search API
	query := fmt.Sprintf("%s repo:%s/%s type:issue state:%s",
		params.Query, params.Owner, params.Repo, params.State)

	searchResult, _, err := client.Search.Issues(ctx, query, &github.SearchOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	})
	if err != nil {
		return nil, err
	}

	return searchResult.Issues, nil
}

// GetPendingReviews returns PRs that are open and potentially need review
func GetPendingReviews(ctx context.Context, input json.RawMessage) ([]*github.PullRequest, error) {
	var params ToolInput
	if err := json.Unmarshal(input, &params); err != nil {
		return nil, err
	}

	client := GitHubClient(ctx)
	
	prs, _, err := client.PullRequests.List(ctx, params.Owner, params.Repo, &github.PullRequestListOptions{
		State: "open",
		ListOptions: github.ListOptions{PerPage: 100},
	})
	if err != nil {
		return nil, err
	}

	// Filter for PRs that might need review
	var pendingReviews []*github.PullRequest
	for _, pr := range prs {
		// Get review status for each PR
		reviews, _, err := client.PullRequests.ListReviews(ctx, params.Owner, params.Repo, pr.GetNumber(), nil)
		if err != nil {
			// If we can't get reviews, assume it needs review
			pendingReviews = append(pendingReviews, pr)
			continue
		}

		// Check if PR has been approved or if it's still pending
		hasApproval := false
		for _, review := range reviews {
			if review.GetState() == "APPROVED" {
				hasApproval = true
				break
			}
		}

		// Include if no approval yet or if it's a draft
		if !hasApproval || pr.GetDraft() {
			pendingReviews = append(pendingReviews, pr)
		}
	}

	return pendingReviews, nil
}

// CreateIssue creates a new GitHub issue
func CreateIssue(ctx context.Context, input json.RawMessage) (*github.Issue, error) {
	var params ToolInput
	if err := json.Unmarshal(input, &params); err != nil {
		return nil, err
	}

	client := GitHubClient(ctx)

	issueRequest := &github.IssueRequest{
		Title: &params.Title,
		Body:  &params.Body,
	}

	if len(params.Labels) > 0 {
		issueRequest.Labels = &params.Labels
	}

	if params.Assignee != "" {
		issueRequest.Assignee = &params.Assignee
	}

	issue, _, err := client.Issues.Create(ctx, params.Owner, params.Repo, issueRequest)
	if err != nil {
		return nil, err
	}

	return issue, nil
}

// AnalyzeIssuePriority analyzes issues and categorizes them by priority
func AnalyzeIssuePriority(ctx context.Context, input json.RawMessage) (map[string][]map[string]interface{}, error) {
	var params ToolInput
	if err := json.Unmarshal(input, &params); err != nil {
		return nil, err
	}

	if params.Limit == 0 {
		params.Limit = 20
	}

	client := GitHubClient(ctx)
	
	issues, _, err := client.Issues.ListByRepo(ctx, params.Owner, params.Repo, &github.IssueListByRepoOptions{
		State: "open",
		ListOptions: github.ListOptions{PerPage: params.Limit},
	})
	if err != nil {
		return nil, err
	}

	// Filter out pull requests
	var actualIssues []*github.Issue
	for _, issue := range issues {
		if !issue.IsPullRequest() {
			actualIssues = append(actualIssues, issue)
		}
	}

	// Calculate priority scores and categorize
	type issueWithScore struct {
		issue *github.Issue
		score int
	}

	var issuesWithScores []issueWithScore
	for _, issue := range actualIssues {
		score := calculatePriorityScore(issue)
		issuesWithScores = append(issuesWithScores, issueWithScore{issue, score})
	}

	// Sort by priority score
	sort.Slice(issuesWithScores, func(i, j int) bool {
		return issuesWithScores[i].score > issuesWithScores[j].score
	})

	// Categorize by priority
	result := make(map[string][]map[string]interface{})
	result["🔴 critical"] = []map[string]interface{}{}
	result["🟡 high"] = []map[string]interface{}{}
	result["🟢 medium"] = []map[string]interface{}{}
	result["⚪ low"] = []map[string]interface{}{}

	for _, item := range issuesWithScores {
		issueInfo := map[string]interface{}{
			"number":         item.issue.GetNumber(),
			"title":          item.issue.GetTitle(),
			"priority_score": item.score,
			"comments":       item.issue.GetComments(),
			"reactions":      item.issue.GetReactions().GetTotalCount(),
			"url":           item.issue.GetHTMLURL(),
		}

		// Categorize based on score and labels
		if item.score >= 20 || hasLabel(item.issue, []string{"critical", "urgent", "p0"}) {
			result["🔴 critical"] = append(result["🔴 critical"], issueInfo)
		} else if item.score >= 10 || hasLabel(item.issue, []string{"high", "important", "p1"}) {
			result["🟡 high"] = append(result["🟡 high"], issueInfo)
		} else if item.score >= 5 || hasLabel(item.issue, []string{"medium", "p2"}) {
			result["🟢 medium"] = append(result["🟢 medium"], issueInfo)
		} else {
			result["⚪ low"] = append(result["⚪ low"], issueInfo)
		}
	}

	return result, nil
}

func calculatePriorityScore(issue *github.Issue) int {
	score := 0

	// Comments weight (more discussion usually means more urgency or complexity)
	score += issue.GetComments() * 2

	// Reactions weight (e.g., 👍 or 👎)
	if issue.GetReactions() != nil {
		score += issue.GetReactions().GetTotalCount()
	}

	// Age weight (older unresolved issues might be more urgent to address)
	created := issue.GetCreatedAt().Time
	ageDays := int(time.Since(created).Hours() / 24)

	if ageDays > 30 {
		score += 5
	} else if ageDays > 14 {
		score += 3
	} else if ageDays > 7 {
		score += 1
	}

	// Label-based score boosts
	if hasLabel(issue, []string{"critical", "urgent", "p0"}) {
		score += 10
	} else if hasLabel(issue, []string{"high", "important", "p1"}) {
		score += 5
	} else if hasLabel(issue, []string{"medium", "p2"}) {
		score += 3
	} else if hasLabel(issue, []string{"low", "p3"}) {
		score += 1
	}

	// Title/body heuristics (quick signal from keywords)
	title := strings.ToLower(issue.GetTitle())
	body := strings.ToLower(issue.GetBody())
	if strings.Contains(title, "crash") || strings.Contains(body, "crash") {
		score += 5
	}
	if strings.Contains(title, "security") || strings.Contains(body, "security") {
		score += 5
	}
	if strings.Contains(title, "blocker") || strings.Contains(body, "blocker") {
		score += 5
	}

	return score
}

func hasLabel(issue *github.Issue, keywords []string) bool {
	for _, label := range issue.Labels {
		for _, keyword := range keywords {
			if strings.EqualFold(label.GetName(), keyword) {
				return true
			}
		}
	}
	return false
}