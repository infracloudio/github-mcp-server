# Secure GitHub MCP Server

A secure implementation of a GitHub MCP (Machine Control Protocol) server with Keycloak integration for authentication and RBAC (Role-Based Access Control).

## Features

- OAuth2 authentication with Keycloak
- Role-Based Access Control (RBAC)
- GitHub integration for repository management
- Secure configuration management
- Tool-level permission checks

## Prerequisites

- Go 1.19 or later
- Keycloak server (local or remote)
- GitHub account and personal access token
- Docker (optional, for running Keycloak locally)

## Setup

1. Clone the repository:
```bash
git clone https://github.com/yourusername/secure-github-mcp-server.git
cd secure-github-mcp-server
```

2. Set up Keycloak:
   - Install Keycloak locally or use a remote instance
   - Create a new realm called "mcp-realm"
   - Create a new client with the following settings:
     - Client ID: mcp-client
     - Client Protocol: openid-connect
     - Access Type: confidential
     - Valid Redirect URIs: http://localhost:8081/callback
   - Create roles: admin, user, viewer
   - Create test users and assign roles

3. Configure environment variables:
```bash
export KEYCLOAK_URL="http://localhost:8080"
export KEYCLOAK_REALM="mcp-realm"
export OAUTH_CLIENT_ID="your-client-id"
export OAUTH_CLIENT_SECRET="your-client-secret"
export MCP_AUTH_TOKEN="your-development-token"
```

4. Build and run the server:
```bash
make build
./bin/github-mcp-server
```

## Configuration

The server can be configured using environment variables or a configuration file. For Claude desktop integration, use the following configuration:

```json
{
  "mcpServers": {
    "secure-github": {
      "command": "/path/to/secure-github-mcp-server",
      "env": {
        "OAUTH_CLIENT_ID": "your_client_id",
        "OAUTH_CLIENT_SECRET": "your_client_secret",
        "KEYCLOAK_URL": "http://localhost:8080",
        "KEYCLOAK_REALM": "mcp-realm",
        "MCP_AUTH_TOKEN": "your-development-token"
      }
    }
  }
}
```

## Available Tools

1. `list_prs` - List pull requests (requires read:tools permission)
2. `list_issues` - List repository issues (requires read:tools permission)
3. `search_issues` - Search issues by keyword (requires read:tools permission)
4. `get_pending_reviews` - Get PRs pending review (requires read:tools permission)
5. `create_issue` - Create a new issue (requires write:tools permission)
6. `analyze_issue_priority` - Analyze issue priority (requires read:tools permission)

## Role Permissions

- Admin: All permissions
- User: read:tools, write:tools
- Viewer: read:tools

## Security Considerations

1. Always use HTTPS in production
2. Keep OAuth client secrets secure
3. Regularly rotate access tokens
4. Monitor authentication logs
5. Follow the principle of least privilege when assigning roles

## Development

To run the server in development mode:

```bash
make dev
```

For testing:

```bash
make test
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request

## License

MIT License - see LICENSE file for details