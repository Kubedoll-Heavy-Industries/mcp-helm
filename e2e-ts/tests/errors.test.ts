import { describe, it, expect, inject, beforeAll, afterAll } from "vitest";
import { Client } from "@modelcontextprotocol/sdk/client/index.js";
import { StreamableHTTPClientTransport } from "@modelcontextprotocol/sdk/client/streamableHttp.js";
import { McpError } from "@modelcontextprotocol/sdk/types.js";

const baseUrl = inject("baseUrl");

let client: Client;
let transport: StreamableHTTPClientTransport;

beforeAll(async () => {
  transport = new StreamableHTTPClientTransport(new URL(`${baseUrl}/mcp`));
  client = new Client({ name: "e2e-test", version: "1.0.0" });
  await client.connect(transport);
});

afterAll(async () => {
  await client?.close();
  await transport?.close();
});

describe("error handling", () => {
  it("rejects search_charts when repository_url is missing", async () => {
    await expect(
      client.callTool({ name: "search_charts", arguments: {} }),
    ).rejects.toThrow(McpError);
  });

  it("rejects get_versions when chart_name is missing", async () => {
    await expect(
      client.callTool({
        name: "get_versions",
        arguments: { repository_url: "https://prometheus-community.github.io/helm-charts" },
      }),
    ).rejects.toThrow(McpError);
  });

  it("returns isError for chart not found", async () => {
    const result = await client.callTool({
      name: "get_versions",
      arguments: {
        repository_url: "https://prometheus-community.github.io/helm-charts",
        chart_name: "nonexistent-chart-that-does-not-exist-xyz",
      },
    });
    expect(result.isError).toBe(true);
  });

  it("returns isError for invalid repository URL", async () => {
    const result = await client.callTool({
      name: "search_charts",
      arguments: { repository_url: "https://not-a-real-helm-repo.example.com/charts" },
    });
    expect(result.isError).toBe(true);
  });
});
