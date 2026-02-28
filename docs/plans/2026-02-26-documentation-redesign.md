# Documentation Redesign Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Slim the README from ~350 to ~130 lines, move reference content to `docs/`, fix stale links, create missing files.

**Architecture:** Hub-and-spoke — lean README links out to focused docs. Collapsible `<details>` for editor configs (Context7 pattern). Drop Fly.io (migrating to Cloudflare), Homebrew placeholder, and Mise.

**Tech Stack:** Markdown, GitHub-flavored `<details>` tags

---

### Task 1: Create docs/configuration.md

**Files:**
- Create: `docs/configuration.md`

**Step 1: Write configuration.md**

Migrate the Transport Modes table, all CLI Flags tables (Transport, Helm, Security, Server, Logging), and the env var convention from `internal/config/config.go`. Source of truth for flags is `config.go:63-87`. Include the `MCP_HELM_` env prefix convention documented at `config.go:47-53`.

```markdown
# Configuration

mcp-helm is configured via CLI flags or environment variables. CLI flags take precedence.

Environment variables use the `MCP_HELM_` prefix with uppercase flag names and hyphens replaced by underscores (e.g., `--helm-timeout` becomes `MCP_HELM_HELM_TIMEOUT`).

## Transport Modes

| Mode | Use case | Recommendation |
|------|----------|----------------|
| `stdio` | Local editor integration | **Use this for local setups** — faster, simpler, no ports to manage |
| `http` | Shared server, multiple clients | Use for self-hosted/production deployments |

## CLI Flags

### Transport

| Flag | Env | Default | Description |
|------|-----|---------|-------------|
| `--transport` | `MCP_HELM_TRANSPORT` | `stdio` | Transport mode: `stdio` or `http` |
| `--listen` | `MCP_HELM_LISTEN` | `:8012` | Listen address for HTTP mode |

### Helm

| Flag | Env | Default | Description |
|------|-----|---------|-------------|
| `--helm-timeout` | `MCP_HELM_HELM_TIMEOUT` | `30s` | Timeout for Helm operations |
| `--cache-size` | `MCP_HELM_CACHE_SIZE` | `50` | Maximum charts to cache in memory |
| `--index-ttl` | `MCP_HELM_INDEX_TTL` | `5m` | Repository index cache TTL |
| `--max-output-size` | `MCP_HELM_MAX_OUTPUT_SIZE` | `2097152` | Max tool output size in bytes |

### Security

| Flag | Env | Default | Description |
|------|-----|---------|-------------|
| `--allow-private-ips` | `MCP_HELM_ALLOW_PRIVATE_IPS` | `false` | Allow fetching from private/loopback IPs |
| `--allowed-hosts` | `MCP_HELM_ALLOWED_HOSTS` | | Hostname allowlist (comma-separated) |
| `--denied-hosts` | `MCP_HELM_DENIED_HOSTS` | | Hostname denylist (comma-separated) |

> **Note:** For HTTP deployments, use an API gateway (nginx, envoy, cloud load balancer) for rate limiting, authentication, and TLS termination.

### Server

| Flag | Env | Default | Description |
|------|-----|---------|-------------|
| `--read-timeout` | `MCP_HELM_READ_TIMEOUT` | `30s` | HTTP read timeout |
| `--write-timeout` | `MCP_HELM_WRITE_TIMEOUT` | `30s` | HTTP write timeout |

### Logging

| Flag | Env | Default | Description |
|------|-----|---------|-------------|
| `--log-level` | `MCP_HELM_LOG_LEVEL` | `info` | Log level: `debug`, `info`, `warn`, `error` |
| `--log-format` | `MCP_HELM_LOG_FORMAT` | `json` | Log format: `json` or `console` |
```

**Step 2: Commit**

```
git add docs/configuration.md
git commit -m "docs: add configuration reference"
```

---

### Task 2: Create docs/self-hosting.md

**Files:**
- Create: `docs/self-hosting.md`

**Step 1: Write self-hosting.md**

Migrate Docker HTTP mode and health endpoints from README. Drop Fly.io section entirely (migrating to Cloudflare). Keep Docker-generic.

```markdown
# Self-Hosting

Run your own mcp-helm instance for shared deployments or when you need an HTTP endpoint.

## Docker

```bash
docker run -p 8012:8012 ghcr.io/kubedoll-heavy-industries/mcp-helm:latest \
  --transport=http --listen=:8012
```

Connect your MCP client to `http://localhost:8012/mcp`.

## Health Endpoints

Available in HTTP mode:

| Endpoint | Purpose |
|----------|---------|
| `GET /healthz` | Liveness probe |
| `GET /readyz` | Readiness probe |

## Production Recommendations

- Use an API gateway (nginx, envoy, cloud load balancer) for TLS termination, authentication, and rate limiting
- Set `--allowed-hosts` to restrict which Helm repositories can be queried
- Set `--denied-hosts` to block specific repositories
- See [Configuration](configuration.md) for all available flags
```

**Step 2: Commit**

```
git add docs/self-hosting.md
git commit -m "docs: add self-hosting guide"
```

---

### Task 3: Create docs/contributing.md

**Files:**
- Create: `docs/contributing.md`

**Step 1: Write contributing.md**

Migrate Prerequisites, Setup, Dev Commands, Manual Build, MCP Inspector, Project Structure, and Submitting Changes from README. Fix stale `dev` branch reference to `main`.

```markdown
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
```

**Step 2: Commit**

```
git add docs/contributing.md
git commit -m "docs: add contributing guide"
```

---

### Task 4: Create docs/troubleshooting.md

**Files:**
- Create: `docs/troubleshooting.md`

**Step 1: Write troubleshooting.md**

Migrate all four troubleshooting sections from README verbatim.

```markdown
# Troubleshooting

## "Connection refused" or "Cannot connect to server"

**stdio mode:** Ensure the binary is in your PATH and executable:

```bash
which mcp-helm
mcp-helm --version
```

**HTTP mode:** Check the server is running and the port is correct:

```bash
curl http://localhost:8012/healthz
```

## "Tool not found" or tools not appearing

Your MCP client may need to be restarted after config changes. In Claude Desktop, fully quit and reopen the app.

## Rate limited on public instance

The public instance allows 60 requests/minute. For higher limits, [self-host](self-hosting.md) your own instance.

## Helm repository errors

Some repositories require authentication or have rate limits. Try a different repository or self-host with `--allowed-hosts` configured.
```

**Step 2: Commit**

```
git add docs/troubleshooting.md
git commit -m "docs: add troubleshooting guide"
```

---

### Task 5: Create SECURITY.md

**Files:**
- Create: `SECURITY.md`

**Step 1: Write SECURITY.md**

The README links to this but it doesn't exist. Create a standard vulnerability reporting policy.

```markdown
# Security Policy

## Reporting a Vulnerability

Please report security vulnerabilities by emailing **security@kubedoll.com**.

Do **not** open a public GitHub issue for security vulnerabilities.

We will acknowledge your report within 48 hours and provide a timeline for a fix.

## Supported Versions

| Version | Supported |
|---------|-----------|
| Latest release | Yes |
| Previous minor | Best effort |

## Security Features

mcp-helm includes several security measures:

- **SSRF protection** — private/loopback IP resolution is blocked by default (`--allow-private-ips=false`)
- **DNS timeout** — 5-second DNS resolution timeout prevents slow-resolution attacks
- **Host allowlist/denylist** — restrict which Helm repositories can be queried
- **Chart size limits** — prevents excessively large chart downloads
- **No shell execution** — mcp-helm never executes Helm CLI commands or shell processes
```

**Step 2: Commit**

```
git add SECURITY.md
git commit -m "docs: add security policy"
```

---

### Task 6: Rewrite README.md

**Files:**
- Modify: `README.md` (full rewrite)

**Step 1: Write the new README**

This is the core task. Target ~130 lines. Keep: badges, pitch, before/after, try-it-now, tool table. Add: collapsible editor configs, compact install. Replace: all detail sections with links to docs/.

The new README content (complete):

```markdown
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
```

**Step 2: Verify line count**

Run: `wc -l README.md`
Expected: ~130-145 lines

**Step 3: Commit**

```
git add README.md
git commit -m "docs: slim README, move reference content to docs/"
```

---

### Task 7: Clean up stale files

**Files:**
- Delete: `fly.toml` (migrating to Cloudflare)

**Step 1: Check if fly.toml is referenced anywhere besides README**

Run: `grep -r "fly.toml" --include="*.go" --include="*.yml" --include="*.yaml" .`
Expected: No matches (README reference already removed in Task 6)

**Step 2: Remove fly.toml**

```
git rm fly.toml
git commit -m "chore: remove fly.toml (migrating to Cloudflare)"
```

---

### Task 8: Final verification

**Step 1: Check for broken internal links**

Verify all links in README.md resolve:
- `docs/self-hosting.md` exists
- `docs/configuration.md` exists
- `docs/troubleshooting.md` exists
- `docs/contributing.md` exists
- `SECURITY.md` exists
- `LICENSE` exists

**Step 2: Check for stale references**

Run: `grep -r "CONTRIBUTING.md" . --include="*.md" --include="*.yml"`
Expected: No references to root `CONTRIBUTING.md` (now `docs/contributing.md`)

Run: `grep -r "fly" . --include="*.md"`
Expected: No Fly.io references

Run: `grep -r '"dev"' . --include="*.md" | grep -i branch`
Expected: No references to `dev` branch
