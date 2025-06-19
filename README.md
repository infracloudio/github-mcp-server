# GitHub MCP Server

This repository provides a production-ready [Model Context Protocol (MCP)](https://modelcontextprotocol.io/introduction) server for GitHub, maintained by InfraCloud and authored by [Himanshu Sharma](https://github.com/himanshusharma89).

This MCP server exposes a set of secure, stateless tools for listing, searching, and creating GitHub issues and pull requests, and more. It is designed for easy integration with LLM agents and automation frameworks.

> **For a detailed walkthrough and best practices, see my blog post:**
> [Build your own MCP server (by Himanshu Sharma)](https://www.infracloud.io/blogs/build-your-own-mcp-server/?utm_source=kcd-bangalore-2025&utm_medium=atul-talk&utm_campaign=kcd-blr)

---

## Quickstart

### Prerequisites
- Go 1.20+
- GitHub OAuth App credentials (for authentication)
- [mcp-go](https://github.com/mark3labs/mcp-go) library
- A valid GitHub token in your environment

### Installation & Run

```bash
git clone https://github.com/himanshusharma89/github-mcp-server.git
cd github-mcp-server
go build -o bin/github-mcp-server main.go
```

Set your environment variables:
```bash
export GITHUB_TOKEN=your_github_token
```

Start the server:
```bash
./bin/github-mcp-server
```

---

## Available Tools

| Tool Name                | Description                                      |
|--------------------------|--------------------------------------------------|
| `list_prs`               | List pull requests in a repository               |
| `list_issues`            | List issues in a repository                      |
| `search_issues`          | Search issues by keyword/topic                   |
| `get_pending_reviews`    | Get pull requests pending review                 |
| `create_issue`           | Create a new GitHub issue                        |
| `analyze_issue_priority` | Analyze and rank issues by priority              |

All tools require authentication and are protected by permission checks.

---

## About This Project & Blog

This repository is the official InfraCloud implementation for a GitHub MCP server, maintained by [Himanshu Sharma](https://github.com/himanshusharma89). For a deep dive into the architecture, security, and best practices, read the full blog post:

👉 [Build your own MCP server](https://www.infracloud.io/blogs/build-your-own-mcp-server/?utm_source=kcd-bangalore-2025&utm_medium=atul-talk&utm_campaign=kcd-blr)

The blog covers:
- How the MCP server is structured
- Tool registration and handler patterns
- Security and permissioning
- Real-world usage and integration tips

---

For questions, suggestions, or contributions, please open an issue or PR!