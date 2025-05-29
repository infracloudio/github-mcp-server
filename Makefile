.PHONY: setup build inspect test claude-test clean

# Default Go command
GO_CMD ?= go

# Output binary name
BINARY_NAME ?= github-mcp-server
BINARY_PATH ?= bin/$(BINARY_NAME)

# Main Go file
MAIN_GO_FILE ?= main.go

# Setup development environment (optional, good for first-time setup)
setup:
	@echo "Setting up GitHub MCP Server development environment..."
	$(GO_CMD) mod tidy
	@echo "Setup complete. Ensure GITHUB_TOKEN is set in your environment for runtime."

# Build the MCP server
build:
	@echo "Building GitHub MCP Server..."
	mkdir -p bin
	$(GO_CMD) build -o $(BINARY_PATH) $(MAIN_GO_FILE)
	@echo "Build complete: $(BINARY_PATH)"

# Inspect the MCP server (your existing target)
inspect:
	@echo "Inspecting GitHub MCP Server..."
	npx @modelcontextprotocol/inspector $(GO_CMD) run $(MAIN_GO_FILE)

# Run Go tests
test:
	@echo "Running Go tests..."
	$(GO_CMD) test ./... -v
	@echo "Go tests complete. Ensure GITHUB_TOKEN is set for integration tests."

# Build for Claude Desktop testing and provide guidance
claude-test: build
	@echo ""
	@echo "-------------------------------------------------------"
	@echo "  GitHub MCP Server - Claude Desktop Integration Test  "
	@echo "-------------------------------------------------------"
	@echo "Server binary built: $$(pwd)/$(BINARY_PATH)"
	@echo ""
	@echo "Next Steps for Claude Desktop:"
	@echo "1. Ensure Claude Desktop is installed."
	@echo "2. Open Claude Desktop configuration file:"
	@echo "   - macOS: ~/Library/Application Support/Claude/claude_desktop_config.json"
	@echo "   - Windows: %APPDATA%/Claude/claude_desktop_config.json"
	@echo "3. Add or update the server configuration (replace placeholders):"
	@echo "   {"
	@echo "     \"mcpServers\": {"
	@echo "       \"github-tools\": {  // Or your preferred name"
	@echo "         \"command\": \"$$(pwd)/$(BINARY_PATH)\","
	@echo "         \"env\": {"
	@echo "           \"GITHUB_TOKEN\": \"YOUR_GITHUB_PERSONAL_ACCESS_TOKEN\" // Or ensure it's in your system env"
	@echo "         }"
	@echo "       }"
	@echo "     }"
	@echo "   }"
	@echo "4. Save the config file and RESTART Claude Desktop."
	@echo "5. Test with natural language queries in Claude Desktop."
	@echo ""
	@echo "Make sure GITHUB_TOKEN is correctly set and accessible!"
	@echo "-------------------------------------------------------"

# Clean up build artifacts
clean:
	@echo "Cleaning up..."
	rm -f $(BINARY_PATH)
	# If you have other build artifacts, add them here e.g., rm -rf bin/
	@echo "Cleanup complete."