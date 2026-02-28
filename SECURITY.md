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
