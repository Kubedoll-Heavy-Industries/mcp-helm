# Contributing

## Prerequisites

- Go 1.24+
- [Mise](https://mise.jdx.dev/) (recommended) or manual tool installation

## Setup

```bash
git clone https://github.com/Kubedoll-Heavy-Industries/mcp-helm.git
cd mcp-helm
mise install          # Install Go and dev tools
```

## Development Commands

```bash
mise run dev          # Start server with hot reload (uses Air)
mise run test         # Run tests
mise run lint         # Run golangci-lint
mise run build        # Build binary to ./bin/mcp-helm
```

### Manual Build

```bash
go build -o mcp-helm ./cmd/mcp-helm
./mcp-helm --help
```

### Testing with MCP Inspector

```bash
mise run inspector    # Opens MCP Inspector UI
```

## Project Structure

```
├── cmd/mcp-helm/       # CLI entrypoint
├── internal/
│   ├── config/         # Configuration loading
│   ├── helm/           # Helm client and domain types
│   ├── server/         # HTTP server and middleware
│   ├── handler/        # MCP tool handlers
│   └── mcputil/        # MCP helper utilities
└── docker/             # Container configurations
```

## Submitting Changes

1. Fork the repository
2. Create a feature branch from `main`
3. Make your changes with [conventional commits](https://www.conventionalcommits.org/)
4. Run `mise run test && mise run lint`
5. Open a PR to `main`
