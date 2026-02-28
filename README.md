# mcp-helm

[![CI](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/actions/workflows/ci.yml/badge.svg)](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/Kubedoll-Heavy-Industries/mcp-helm/branch/main/graph/badge.svg)](https://codecov.io/gh/Kubedoll-Heavy-Industries/mcp-helm)
[![Go Report Card](https://goreportcard.com/badge/github.com/Kubedoll-Heavy-Industries/mcp-helm)](https://goreportcard.com/report/github.com/Kubedoll-Heavy-Industries/mcp-helm)
[![Release](https://img.shields.io/github/v/release/Kubedoll-Heavy-Industries/mcp-helm)](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/releases/latest)
[![License: MIT](https://img.shields.io/github/license/Kubedoll-Heavy-Industries/mcp-helm)](LICENSE)

Give your AI assistant access to real Helm chart data. No more hallucinated `values.yaml` files.

## What is this?

When you ask Claude, Cursor, or other AI assistants to help with Kubernetes deployments, they don't have access to Helm chart schemas. So they guess — and the guesses look plausible but don't match reality.

**Without mcp-helm:**
- :x: Hallucinates field names that look right but don't exist
- :x: Suggests stale or deprecated chart versions
- :x: Wastes tokens on web fetches and guesswork

**With mcp-helm:**
- :white_check_mark: Queries actual Helm repositories for real chart data
- :white_check_mark: Gets the latest chart version automatically
- :white_check_mark: Correct configurations the first time

mcp-helm implements the [Model Context Protocol (MCP)](https://modelcontextprotocol.io/) — a standard way for AI assistants to access external data sources.

## Try It Now

Add this to your editor's MCP config to use our public instance (rate limited, no install required):

```json
{
  "mcpServers": {
    "helm": {
      "type": "streamable-http",
      "url": "https://helm-mcp.kubedoll.com/mcp"
    }
  }
}
```

Then ask your AI: *"What values can I configure for the bitnami/postgresql chart?"*

## Editor Setup

<details>
<summary>Claude Code</summary>

Edit `~/.claude/mcp.json`:

```json
{
  "mcpServers": {
    "helm": {
      "command": "docker",
      "args": ["run", "--rm", "-i", "ghcr.io/kubedoll-heavy-industries/mcp-helm", "--transport=stdio"]
    }
  }
}
```

</details>

<details>
<summary>Claude Desktop</summary>

Edit `~/Library/Application Support/Claude/claude_desktop_config.json` (macOS) or `%APPDATA%\Claude\claude_desktop_config.json` (Windows):

```json
{
  "mcpServers": {
    "helm": {
      "command": "docker",
      "args": ["run", "--rm", "-i", "ghcr.io/kubedoll-heavy-industries/mcp-helm", "--transport=stdio"]
    }
  }
}
```

</details>

<details>
<summary>Cursor</summary>

Edit MCP settings in Cursor's configuration:

```json
{
  "mcpServers": {
    "helm": {
      "command": "docker",
      "args": ["run", "--rm", "-i", "ghcr.io/kubedoll-heavy-industries/mcp-helm", "--transport=stdio"]
    }
  }
}
```

</details>

<details>
<summary>VS Code + Continue</summary>

Add to your Continue config (`~/.continue/config.json`):

```json
{
  "experimental": {
    "modelContextProtocolServers": [
      {
        "transport": {
          "type": "stdio",
          "command": "docker",
          "args": ["run", "--rm", "-i", "ghcr.io/kubedoll-heavy-industries/mcp-helm", "--transport=stdio"]
        }
      }
    ]
  }
}
```

</details>

<details>
<summary>Without Docker</summary>

If you prefer to run the binary directly, [install mcp-helm](#install) and replace the Docker config with:

```json
{
  "mcpServers": {
    "helm": {
      "command": "mcp-helm"
    }
  }
}
```

</details>

## Available Tools

| Tool | What it does |
|------|--------------|
| `search_charts` | Search for charts in a Helm repository |
| `get_versions` | Get available versions of a chart (newest first, use `limit=1` for latest) |
| `get_values` | Get chart `values.yaml` with optional JSON schema (`include_schema=true`) |
| `get_dependencies` | Get chart dependencies from Chart.yaml |
| `get_notes` | Get chart NOTES.txt (post-install instructions) |

## Install

**Docker** (recommended — no install required, used in Editor Setup above):

```bash
docker pull ghcr.io/kubedoll-heavy-industries/mcp-helm:latest
```

**Binary:**

```bash
curl -fsSL https://github.com/Kubedoll-Heavy-Industries/mcp-helm/releases/latest/download/mcp-helm_$(uname -s)_$(uname -m).tar.gz | tar xz
sudo mv mcp-helm /usr/local/bin/
```

**Go:**

```bash
go install github.com/Kubedoll-Heavy-Industries/mcp-helm/cmd/mcp-helm@latest
```

## Self-Hosting

For shared deployments or when you need an HTTP endpoint:

```bash
docker run -p 8012:8012 ghcr.io/kubedoll-heavy-industries/mcp-helm:latest \
  --transport=http --listen=:8012
# Connect to http://localhost:8012/mcp
```

See [docs/self-hosting.md](docs/self-hosting.md) for health endpoints and production recommendations.

## Documentation

- [Configuration Reference](docs/configuration.md) — CLI flags, env vars, transport modes
- [Self-Hosting Guide](docs/self-hosting.md) — Docker HTTP, health endpoints, production tips
- [Troubleshooting](docs/troubleshooting.md) — common issues and fixes
- [Contributing](docs/contributing.md) — development setup, testing, PR guidelines
- [Security Policy](SECURITY.md) — reporting vulnerabilities

## License

MIT — see [LICENSE](LICENSE).
