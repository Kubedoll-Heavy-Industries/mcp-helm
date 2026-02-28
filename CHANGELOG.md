# Changelog

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
