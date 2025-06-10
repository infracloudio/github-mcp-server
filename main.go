package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/himanshusharma89/github-mcp-server/auth"   // adjust this path if needed
	"github.com/himanshusharma89/github-mcp-server/config" // adjust this path if needed
	"github.com/himanshusharma89/github-mcp-server/tools"  // adjust this path if needed
)

func main() {
	// Initialize security config
	securityConfig := config.NewSecurityConfig()
	if err := securityConfig.Validate(); err != nil {
		log.Fatalf("Security configuration error: %v", err)
	}

	// Initialize OAuth config
	oauthConfig, err := auth.NewOAuthConfig(context.Background())
	if err != nil {
		log.Fatalf("Failed to initialize OAuth: %v", err)
	}

	// Initialize auth middleware
	authMiddleware := auth.NewAuthMiddleware(oauthConfig)

	s := server.NewMCPServer(
		"GitHub MCP Server",
		"0.1.0",
		server.WithToolCapabilities(false),
	)

	// Define the enhanced tools
	listPRsTool := mcp.NewTool("list_prs",
		mcp.WithDescription("List pull requests in a GitHub repository"),
		mcp.WithString("owner",
			mcp.Required(),
			mcp.Description("GitHub org or user"),
		),
		mcp.WithString("repo",
			mcp.Required(),
			mcp.Description("GitHub repository name"),
		),
		mcp.WithString("state",
			mcp.Description("State of PRs to list (open, closed, all). Defaults to open"),
		),
	)

	listIssuestool := mcp.NewTool("list_issues",
		mcp.WithDescription("List issues in a GitHub repository"),
		mcp.WithString("owner",
			mcp.Required(),
			mcp.Description("GitHub org or user"),
		),
		mcp.WithString("repo",
			mcp.Required(),
			mcp.Description("GitHub repository name"),
		),
		mcp.WithString("state",
			mcp.Description("State of issues to list (open, closed, all). Defaults to open"),
		),
	)

	searchIssuesTool := mcp.NewTool("search_issues",
		mcp.WithDescription("Search issues by keyword/topic and analyze priority"),
		mcp.WithString("owner",
			mcp.Required(),
			mcp.Description("GitHub org or user"),
		),
		mcp.WithString("repo",
			mcp.Required(),
			mcp.Description("GitHub repository name"),
		),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("Search query/topic to filter issues"),
		),
		mcp.WithString("state",
			mcp.Description("State of issues to search (open, closed, all). Defaults to open"),
		),
		mcp.WithBoolean("prioritize",
			mcp.Description("Whether to analyze and sort by priority. Defaults to false"),
		),
	)

	pendingReviewsTool := mcp.NewTool("get_pending_reviews",
		mcp.WithDescription("Get pull requests pending review"),
		mcp.WithString("owner",
			mcp.Required(),
			mcp.Description("GitHub org or user"),
		),
		mcp.WithString("repo",
			mcp.Required(),
			mcp.Description("GitHub repository name"),
		),
	)

	createIssueTool := mcp.NewTool("create_issue",
		mcp.WithDescription("Create a new GitHub issue (useful for K8s diagnostic integration)"),
		mcp.WithString("owner",
			mcp.Required(),
			mcp.Description("GitHub org or user"),
		),
		mcp.WithString("repo",
			mcp.Required(),
			mcp.Description("GitHub repository name"),
		),
		mcp.WithString("title",
			mcp.Required(),
			mcp.Description("Issue title"),
		),
		mcp.WithString("body",
			mcp.Description("Issue body/description"),
		),
		mcp.WithString("labels",
			mcp.Description("Comma-separated labels to apply to the issue"),
		),
		mcp.WithString("assignee",
			mcp.Description("Username to assign the issue to"),
		),
	)

	priorityTool := mcp.NewTool("analyze_issue_priority",
		mcp.WithDescription("Analyze and rank issues by priority based on comments, reactions, labels"),
		mcp.WithString("owner",
			mcp.Required(),
			mcp.Description("GitHub org or user"),
		),
		mcp.WithString("repo",
			mcp.Required(),
			mcp.Description("GitHub repository name"),
		),
		mcp.WithNumber("limit",
			mcp.Description("Maximum number of issues to analyze. Defaults to 20"),
		),
	)

	// Add new permissions tool
	permissionsTool := mcp.NewTool("get_my_permissions",
		mcp.WithDescription("Get current user's roles and permissions"),
	)

	// Register the tools with their handlers wrapped in authentication middleware
	s.AddTool(listPRsTool, wrapWithAuth(authMiddleware, auth.PermissionReadTools, listOpenPRsHandler))
	s.AddTool(listIssuestool, wrapWithAuth(authMiddleware, auth.PermissionReadTools, listOpenIssuesHandler))
	s.AddTool(searchIssuesTool, wrapWithAuth(authMiddleware, auth.PermissionReadTools, searchIssuesHandler))
	s.AddTool(pendingReviewsTool, wrapWithAuth(authMiddleware, auth.PermissionReadTools, getPendingReviewsHandler))
	s.AddTool(createIssueTool, wrapWithAuth(authMiddleware, auth.PermissionWriteTools, createIssueHandler))
	s.AddTool(priorityTool, wrapWithAuth(authMiddleware, auth.PermissionReadTools, analyzePriorityHandler))
	s.AddTool(permissionsTool, wrapWithAuth(authMiddleware, auth.PermissionReadTools, getMyPermissionsHandler))

	// Run the MCP server
	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}

// listOpenIssuesHandler converts MCP input into raw JSON and delegates to tools.GetOpenIssues
func listOpenIssuesHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	raw, err := json.Marshal(req.Params.Arguments)
	if err != nil {
		return nil, errors.New("failed to marshal arguments")
	}

	issues, err := tools.GetOpenIssues(ctx, raw)
	if err != nil {
		return nil, err
	}

	if len(issues) == 0 {
		return mcp.NewToolResultText("No open issues found."), nil
	}

	var output string
	for _, issue := range issues {
		output += fmt.Sprintf("- #%d: %s\n", issue.GetNumber(), issue.GetTitle())
	}

	return mcp.NewToolResultText(output), nil
}

// listOpenPRsHandler converts MCP input into raw JSON and delegates to tools.GetOpenPRs
func listOpenPRsHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	raw, err := json.Marshal(req.Params.Arguments)
	if err != nil {
		return nil, errors.New("failed to marshal arguments")
	}

	prList, err := tools.GetOpenPRs(ctx, raw)
	if err != nil {
		return nil, err
	}

	if len(prList) == 0 {
		return mcp.NewToolResultText("No open pull requests found."), nil
	}

	var output string
	for _, pr := range prList {
		output += fmt.Sprintf("- #%d: %s\n", pr.GetNumber(), pr.GetTitle())
	}

	return mcp.NewToolResultText(output), nil
}

// searchIssuesHandler handles searching issues by topic/keyword with optional priority analysis
func searchIssuesHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	raw, err := json.Marshal(req.Params.Arguments)
	if err != nil {
		return nil, errors.New("failed to marshal arguments")
	}

	issues, err := tools.SearchIssues(ctx, raw)
	if err != nil {
		return nil, err
	}

	if len(issues) == 0 {
		return mcp.NewToolResultText("No issues found matching the search criteria."), nil
	}

	// Check if prioritization was requested
	var input struct {
		Query      string `json:"query"`
		State      string `json:"state"`
		Prioritize bool   `json:"prioritize"`
	}
	json.Unmarshal(raw, &input)

	var output strings.Builder
	output.WriteString(fmt.Sprintf("Found %d issues related to '%s':\n\n", len(issues), input.Query))

	if input.Prioritize {
		// Sort by priority score (comments + reactions)
		sort.Slice(issues, func(i, j int) bool {
			scoreI := issues[i].GetComments() + issues[i].GetReactions().GetTotalCount()
			scoreJ := issues[j].GetComments() + issues[j].GetReactions().GetTotalCount()
			return scoreI > scoreJ
		})

		// Group by priority levels
		var high, medium, low []string
		for _, issue := range issues {
			score := issue.GetComments() + issue.GetReactions().GetTotalCount()
			labels := ""
			for _, label := range issue.Labels {
				labels += fmt.Sprintf("[%s] ", label.GetName())
			}

			line := fmt.Sprintf("- #%d: %s %s(Score: %d - %d comments, %d reactions)",
				issue.GetNumber(), issue.GetTitle(), labels, score, issue.GetComments(), issue.GetReactions().GetTotalCount())

			if score >= 10 {
				high = append(high, line)
			} else if score >= 3 {
				medium = append(medium, line)
			} else {
				low = append(low, line)
			}
		}

		if len(high) > 0 {
			output.WriteString("🔴 HIGH PRIORITY:\n")
			for _, item := range high {
				output.WriteString(item + "\n")
			}
			output.WriteString("\n")
		}
		if len(medium) > 0 {
			output.WriteString("🟡 MEDIUM PRIORITY:\n")
			for _, item := range medium {
				output.WriteString(item + "\n")
			}
			output.WriteString("\n")
		}
		if len(low) > 0 {
			output.WriteString("🟢 LOW PRIORITY:\n")
			for _, item := range low {
				output.WriteString(item + "\n")
			}
		}
	} else {
		// Simple list format
		for _, issue := range issues {
			output.WriteString(fmt.Sprintf("- #%d: %s\n", issue.GetNumber(), issue.GetTitle()))
		}
	}

	return mcp.NewToolResultText(output.String()), nil
}

// getPendingReviewsHandler gets PRs that are pending review
func getPendingReviewsHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	raw, err := json.Marshal(req.Params.Arguments)
	if err != nil {
		return nil, errors.New("failed to marshal arguments")
	}

	prs, err := tools.GetPendingReviews(ctx, raw)
	if err != nil {
		return nil, err
	}

	if len(prs) == 0 {
		return mcp.NewToolResultText("No pull requests pending review found."), nil
	}

	var output strings.Builder
	output.WriteString(fmt.Sprintf("Found %d PRs pending review:\n\n", len(prs)))

	for _, pr := range prs {
		age := ""
		if pr.CreatedAt != nil {
			age = fmt.Sprintf("(opened %s)", pr.CreatedAt.Format("2006-01-02"))
		}
		output.WriteString(fmt.Sprintf("- #%d: %s %s\n", pr.GetNumber(), pr.GetTitle(), age))
		if pr.GetDraft() {
			output.WriteString("  ⚠️  DRAFT PR\n")
		}
	}

	return mcp.NewToolResultText(output.String()), nil
}

// createIssueHandler creates a new GitHub issue
func createIssueHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	raw, err := json.Marshal(req.Params.Arguments)
	if err != nil {
		return nil, errors.New("failed to marshal arguments")
	}

	issue, err := tools.CreateIssue(ctx, raw)
	if err != nil {
		return nil, err
	}

	output := fmt.Sprintf("✅ Issue created successfully!\n\n"+
		"- Number: #%d\n"+
		"- Title: %s\n"+
		"- URL: %s\n"+
		"- State: %s",
		issue.GetNumber(),
		issue.GetTitle(),
		issue.GetHTMLURL(),
		issue.GetState())

	return mcp.NewToolResultText(output), nil
}

// analyzePriorityHandler analyzes issue priority based on engagement metrics
func analyzePriorityHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	raw, err := json.Marshal(req.Params.Arguments)
	if err != nil {
		return nil, errors.New("failed to marshal arguments")
	}

	analysis, err := tools.AnalyzeIssuePriority(ctx, raw)
	if err != nil {
		return nil, err
	}

	if len(analysis) == 0 {
		return mcp.NewToolResultText("No issues found for priority analysis."), nil
	}

	var output strings.Builder
	output.WriteString("📊 ISSUE PRIORITY ANALYSIS\n\n")

	for category, issues := range analysis {
		output.WriteString(fmt.Sprintf("%s (%d issues):\n", strings.ToUpper(category), len(issues)))
		for _, issue := range issues {
			output.WriteString(fmt.Sprintf("- #%d: %s (Score: %d)\n",
				issue["number"], issue["title"], issue["priority_score"]))
		}
		output.WriteString("\n")
	}

	return mcp.NewToolResultText(output.String()), nil
}

// getMyPermissionsHandler gets current user's roles and permissions
func getMyPermissionsHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Get user info from context
	userInfo, ok := ctx.Value(auth.UserContextKey).(*auth.UserInfo)
	if !ok || userInfo == nil {
		return mcp.NewToolResultError("failed to get user information"), nil
	}

	// Get all permissions for the user's roles
	permissions := auth.GetUserPermissions(userInfo.Roles)

	var output strings.Builder
	output.WriteString("🔑 Your Access Details:\n\n")

	// Show roles
	output.WriteString("Roles:\n")
	for _, role := range userInfo.Roles {
		output.WriteString(fmt.Sprintf("- %s\n", role))
	}
	output.WriteString("\n")

	// Show permissions
	output.WriteString("Permissions:\n")
	for _, perm := range permissions {
		output.WriteString(fmt.Sprintf("- %s\n", perm))
	}
	output.WriteString("\n")

	// Show available tools based on permissions
	output.WriteString("Available Tools:\n")
	if auth.HasPermission(userInfo.Roles, auth.PermissionReadTools) {
		output.WriteString("- list_prs (List pull requests)\n")
		output.WriteString("- list_issues (List repository issues)\n")
		output.WriteString("- search_issues (Search issues by keyword)\n")
		output.WriteString("- get_pending_reviews (Get PRs pending review)\n")
		output.WriteString("- analyze_issue_priority (Analyze issue priority)\n")
		output.WriteString("- get_my_permissions (Show your access details)\n")
	}
	if auth.HasPermission(userInfo.Roles, auth.PermissionWriteTools) {
		output.WriteString("- create_issue (Create new issues)\n")
	}

	return mcp.NewToolResultText(output.String()), nil
}

// wrapWithAuth wraps a tool handler with authentication and permission checks
func wrapWithAuth(am *auth.AuthMiddleware, permission auth.Permission, handler func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error)) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// For development, use environment variable for authentication
		token := os.Getenv("MCP_AUTH_TOKEN")
		if token == "" {
			return mcp.NewToolResultError("authentication token not found in environment"), nil
		}

		// Create a mock request context with the token
		httpReq, _ := http.NewRequest("POST", "/", nil)
		httpReq.Header.Set("Authorization", "Bearer "+token)
		httpReq = httpReq.WithContext(ctx)

		mockWriter := &mockResponseWriter{}
		authenticated := false
		var userInfo *auth.UserInfo

		// Use the auth middleware
		am.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ui, ok := r.Context().Value(auth.UserContextKey).(*auth.UserInfo)
			if !ok {
				return
			}
			authenticated = true
			userInfo = ui
		})).ServeHTTP(mockWriter, httpReq)

		if !authenticated || userInfo == nil {
			return mcp.NewToolResultError("authentication failed"), nil
		}

		if !auth.HasPermission(userInfo.Roles, permission) {
			return mcp.NewToolResultError("permission denied"), nil
		}

		// Add user info to the context before calling the handler
		ctxWithUser := context.WithValue(ctx, auth.UserContextKey, userInfo)
		return handler(ctxWithUser, req)
	}
}

// mockResponseWriter implements http.ResponseWriter for testing
type mockResponseWriter struct {
	headers http.Header
	status  int
	body    []byte
}

func (w *mockResponseWriter) Header() http.Header {
	if w.headers == nil {
		w.headers = make(http.Header)
	}
	return w.headers
}

func (w *mockResponseWriter) Write(b []byte) (int, error) {
	w.body = b
	return len(b), nil
}

func (w *mockResponseWriter) WriteHeader(statusCode int) {
	w.status = statusCode
}
