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
