# Changelog

## [0.1.2](https://github.com/Kubedoll-Heavy-Industries/helm-mcp/compare/v0.1.1...v0.1.2) (2026-03-01)


### Bug Fixes

* **deploy:** fix Container import and update cloudflare deps ([#9](https://github.com/Kubedoll-Heavy-Industries/helm-mcp/issues/9)) ([9bb9590](https://github.com/Kubedoll-Heavy-Industries/helm-mcp/commit/9bb95901443ef3dac7da5c7f573b87636919fd31))
* **server:** return JSON 404 for unmatched routes ([#11](https://github.com/Kubedoll-Heavy-Industries/helm-mcp/issues/11)) ([b4cddb7](https://github.com/Kubedoll-Heavy-Industries/helm-mcp/commit/b4cddb7284fc763dd0a01ba5d013688f57abe930))

## [0.1.1](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/compare/v0.1.0...v0.1.1) (2026-02-28)


### Bug Fixes

* **ci:** reset release-please state to v0.1.0 ([#5](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/5)) ([198775d](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/198775d177a925bb81631f5035fc15442d3ceadc))
* **deps:** upgrade go.opentelemetry.io/otel/sdk to v1.40.0 ([5387da2](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/5387da2eacdcd3e14e01b9012ddc56a9cf5ef22b))
* **deps:** upgrade otel SDK to fix GHSA-9h8m-3fm2-qjrq ([258b27a](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/258b27a3158d76e3bafdb659e6f89dac8b431f6f))
* **docker:** bump alpine from 3.23.0 to 3.23.3 ([#1](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/issues/1)) ([b2a84fe](https://github.com/Kubedoll-Heavy-Industries/mcp-helm/commit/b2a84fe3dc91afe442ed5cc1146abbc1a6632331))

## 0.1.0 (2026-02-28)

Initial release as a clean-history fork.

### Features

* MCP server providing read-only Helm chart tools for LLM agents
* `search_charts` — search for charts in HTTP/HTTPS repositories
* `get_versions` — list available chart versions (HTTP/HTTPS and OCI registries)
* `get_values` — retrieve values.yaml with depth-limited collapsing, path extraction, and optional JSON schema
* `get_dependencies` — list chart dependencies
* `get_notes` — retrieve NOTES.txt post-install instructions
* OCI registry support (`oci://`) for all chart-specific tools
* SSRF protection and URL validation for both HTTP and OCI endpoints
* Streamable HTTP and stdio transports
* Docker image published to GHCR
