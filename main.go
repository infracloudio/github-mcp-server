package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/go-github/github"
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

	// Define the tool: list_open_prs
	listPRsTool := mcp.NewTool("list_open_prs",
		mcp.WithDescription("List open pull requests in a GitHub repository"),
		mcp.WithString("owner",
			mcp.Required(),
			mcp.Description("GitHub org or user"),
		),
		mcp.WithString("repo",
			mcp.Required(),
			mcp.Description("GitHub repository name"),
		),
	)

	// Register the tool with its handler
	s.AddTool(listPRsTool, listOpenPRsHandler)

	// Run the MCP server
	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}

// listOpenPRsHandler converts MCP input into raw JSON and delegates to tools.GetOpenPRs
func listOpenPRsHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	raw, err := json.Marshal(req.Params.Arguments)
	if err != nil {
		return nil, errors.New("failed to marshal arguments")
	}

	result, err := tools.GetOpenPRs(ctx, raw)
	if err != nil {
		return nil, err
	}

	// Convert result to simple string output (titles of PRs)
	prList, ok := result.([]*github.PullRequest)
	if !ok {
		return nil, errors.New("unexpected result format from GetOpenPRs")
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