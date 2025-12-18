# mcp-helm

[![CI](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/actions/workflows/ci.yml/badge.svg)](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/actions/workflows/ci.yml)
[![OpenSSF Scorecard](https://api.scorecard.dev/projects/github.com/Kubedoll-Heavy-Industries/mcp-helm/badge)](https://scorecard.dev/viewer/?uri=github.com/Kubedoll-Heavy-Industries/mcp-helm)
[![codecov](https://codecov.io/gh/Kubedoll-Heavy-Industries/mcp-helm/branch/main/graph/badge.svg)](https://codecov.io/gh/Kubedoll-Heavy-Industries/mcp-helm)
[![Go Report Card](https://goreportcard.com/badge/github.com/Kubedoll-Heavy-Industries/mcp-helm)](https://goreportcard.com/report/github.com/Kubedoll-Heavy-Industries/mcp-helm)
[![Go Version](https://img.shields.io/github/go-mod/go-version/Kubedoll-Heavy-Industries/mcp-helm)](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/blob/main/go.mod)
[![Release](https://img.shields.io/github/v/release/Kubedoll-Heavy-Industries/mcp-helm)](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/releases/latest)
[![License: MIT](https://img.shields.io/github/license/Kubedoll-Heavy-Industries/mcp-helm)](LICENSE)

Give your AI assistant access to real Helm chart data. No more hallucinated `values.yaml` files.

## What is this?

When you ask Claude, Cursor, or other AI assistants to help with Kubernetes deployments, they don't have access to Helm chart schemas. So they guess — and the guesses look plausible but don't match reality.

**Without mcp-helm:**
- :x: Makes multiple web requests searching for chart documentation
- :x: Tries to download values.yaml from GitHub (often wrong branch or version)
- :x: Uses `grep -A 50` to extract config sections, missing context or grabbing irrelevant lines
- :x: Hallucinates field names that look right but don't exist
- :x: Suggests stale or deprecated chart versions
- :x: Wastes tokens and your time on approaches that can't work

**With mcp-helm:**
- :white_check_mark: Queries the actual Helm repository for real chart data
- :white_check_mark: Extracts exact subkeys and values the agent needs
- :white_check_mark: Gets the latest chart version automatically
- :white_check_mark: No wasted tokens on web fetches or guesswork
- :white_check_mark: Correct configurations the first time

mcp-helm implements the [Model Context Protocol (MCP)](https://modelcontextprotocol.io/) — a standard way for AI assistants to access external data sources.

## Try It Now

Add this to your editor's MCP config to use our public instance (rate limited, no install required):

```json
{
  "mcpServers": {
    "helm": {
      "type": "streamable-http",
      "url": "https://mcp-helm.fly.dev/mcp"
    }
  }
}
```

Then ask your AI: *"What values can I configure for the bitnami/postgresql chart?"*

## Editor Setup

### Claude Desktop

Edit `~/Library/Application Support/Claude/claude_desktop_config.json` (macOS) or `%APPDATA%\Claude\claude_desktop_config.json` (Windows):

```json
{
  "mcpServers": {
    "helm": {
      "command": "mcp-helm"
    }
  }
}
```

### Cursor

Edit MCP settings in Cursor's configuration:

```json
{
  "mcpServers": {
    "helm": {
      "command": "mcp-helm"
    }
  }
}
```

### VS Code + Continue

Add to your Continue config (`~/.continue/config.json`):

```json
{
  "experimental": {
    "modelContextProtocolServers": [
      {
        "transport": {
          "type": "stdio",
          "command": "mcp-helm"
        }
      }
    ]
  }
}
```

### Other Editors

Any MCP-compatible client can connect. For local use, **stdio is recommended** — it's faster and requires no port management:

```bash
mcp-helm
```

For shared or remote setups, use HTTP:

```bash
mcp-helm --transport=http --listen=:8012
# Connect to http://localhost:8012/mcp
```

## Available Tools

Once connected, your AI assistant can use these tools:

| Tool | What it does |
|------|--------------|
| `list_repository_charts` | List all charts in a Helm repo (e.g., `https://charts.bitnami.com/bitnami`) |
| `list_chart_versions` | List all versions of a chart with metadata |
| `get_latest_version_of_chart` | Get the current version of a chart |
| `get_chart_values` | Fetch the actual `values.yaml` for any chart version |
| `get_values_schema` | Get the `values.schema.json` if the chart provides one |
| `get_chart_contents` | Get full chart contents (templates, helpers, NOTES.txt) |
| `get_chart_dependencies` | List chart dependencies from Chart.yaml |
| `refresh_repository_index` | Force refresh the cached repository index |

## Installation

You only need to install mcp-helm if you want to run it locally instead of using the public instance.

### Homebrew (macOS/Linux)

```bash
# Coming soon
```

### Binary Download

```bash
# Linux/macOS - downloads the right binary for your system
curl -fsSL https://github.com/Kubedoll-Heavy-Industries/mcp-helm/releases/latest/download/mcp-helm_$(uname -s)_$(uname -m).tar.gz | tar xz
sudo mv mcp-helm /usr/local/bin/
```

### Go Install

```bash
go install github.com/Kubedoll-Heavy-Industries/mcp-helm/cmd/mcp-helm@latest
```

### Mise

```bash
mise install ubi:Kubedoll-Heavy-Industries/mcp-helm@latest
```

## Self-Hosting

If you need a private instance or want lower latency:

### Docker

```bash
docker run -p 8012:8012 ghcr.io/kubedoll-heavy-industries/mcp-helm:latest \
  --transport=http --listen=:8012
```

### Fly.io

Deploy your own instance using our [fly.toml](fly.toml):

```bash
fly launch --copy-config
fly deploy
```

## Configuration Reference

### Transport Modes

| Mode | Use case | Recommendation |
|------|----------|----------------|
| `stdio` | Local editor integration | **Use this for local setups** — faster, simpler, no ports to manage |
| `http` | Shared server, multiple clients | Use for self-hosted/production deployments |

### All CLI Flags

**Transport:**

| Flag | Default | Description |
|------|---------|-------------|
| `--transport` | `stdio` | Transport mode: `stdio` or `http` |
| `--listen` | `:8012` | Listen address for HTTP mode |

**Helm:**

| Flag | Default | Description |
|------|---------|-------------|
| `--helm-timeout` | `30s` | Timeout for Helm operations |
| `--cache-size` | `50` | Maximum number of charts to cache |
| `--index-ttl` | `5m` | Repository index cache TTL |

**Security:**

| Flag | Default | Description |
|------|---------|-------------|
| `--allow-private-ips` | `false` | Allow fetching from private/loopback IPs |
| `--allowed-hosts` | | Hostname allowlist (comma-separated) |
| `--denied-hosts` | | Hostname denylist (comma-separated) |
| `--trusted-proxies` | | CIDR ranges of trusted proxies for X-Forwarded-For |

**Rate Limiting (HTTP mode):**

| Flag | Default | Description |
|------|---------|-------------|
| `--rate-limit` | `false` | Enable rate limiting |
| `--rate-limit-rps` | `10` | Requests per second per client |
| `--rate-limit-burst` | `20` | Burst capacity |

**Server:**

| Flag | Default | Description |
|------|---------|-------------|
| `--read-timeout` | `30s` | HTTP read timeout |
| `--write-timeout` | `30s` | HTTP write timeout |
| `--max-output-size` | `2097152` | Max tool output size in bytes |

**Logging:**

| Flag | Default | Description |
|------|---------|-------------|
| `--log-level` | `info` | Log level: `debug`, `info`, `warn`, `error` |
| `--log-format` | `json` | Log format: `json` or `console` |

### Health Endpoints

Available in `http` mode:

- `GET /healthz` — Liveness probe
- `GET /readyz` — Readiness probe

## Development

### Prerequisites

- Go 1.24+
- [Mise](https://mise.jdx.dev/) (recommended) or manual tool installation

### Setup

```bash
git clone https://github.com/Kubedoll-Heavy-Industries/mcp-helm.git
cd mcp-helm
mise install          # Install Go and dev tools
```

### Development Commands

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

### Project Structure

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

### Submitting Changes

1. Fork the repository
2. Create a feature branch from `dev`
3. Make your changes with [conventional commits](https://www.conventionalcommits.org/)
4. Run `mise run test && mise run lint`
5. Open a PR to `dev`

See [CONTRIBUTING.md](CONTRIBUTING.md) for detailed guidelines.

## Troubleshooting

### "Connection refused" or "Cannot connect to server"

**stdio mode**: Ensure the binary is in your PATH and executable:
```bash
which mcp-helm
mcp-helm --version
```

**http mode**: Check the server is running and the port is correct:
```bash
curl http://localhost:8012/healthz
```

### "Tool not found" or tools not appearing

Your MCP client may need to be restarted after config changes. In Claude Desktop, fully quit and reopen the app.

### Rate limited on public instance

The public instance allows 60 requests/minute. For higher limits, [self-host](#self-hosting) your own instance.

### Helm repository errors

Some repositories require authentication or have rate limits. Try a different repository or self-host with `--allowed-hosts` configured.

## Security

See [SECURITY.md](SECURITY.md) for reporting vulnerabilities.

## License

MIT — see [LICENSE](LICENSE).
