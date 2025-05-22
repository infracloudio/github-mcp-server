package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/himanshusharma89/github-mcp-server/tools" // adjust this path if needed
)

func main() {
	s := server.NewMCPServer(
		"GitHub MCP Server",
		"0.1.0",
		server.WithToolCapabilities(false),
	)

	// Define the tools
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

	// Register the tools with their handlers
	s.AddTool(listPRsTool, listOpenPRsHandler)
	s.AddTool(listIssuestool, listOpenIssuesHandler)

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
