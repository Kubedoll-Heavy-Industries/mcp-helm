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
