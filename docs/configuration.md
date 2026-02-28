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
