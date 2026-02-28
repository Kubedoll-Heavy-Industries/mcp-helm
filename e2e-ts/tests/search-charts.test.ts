import { describe, it, expect, inject, beforeAll, afterAll } from "vitest";
import { Client } from "@modelcontextprotocol/sdk/client/index.js";
import { StreamableHTTPClientTransport } from "@modelcontextprotocol/sdk/client/streamableHttp.js";

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

const REPO_URL = "https://prometheus-community.github.io/helm-charts";

describe("search_charts", () => {
  it("finds alertmanager in prometheus-community repo", async () => {
    const result = await client.callTool({
      name: "search_charts",
      arguments: { repository_url: REPO_URL, search: "alertmanager" },
    });

    expect(result.isError).toBeFalsy();
    const content = result.content as Array<{ type: string; text: string }>;
    const data = JSON.parse(content[0].text) as { charts: string[]; total: number };

    expect(data.charts).toContain("alertmanager");
    expect(data.total).toBeGreaterThanOrEqual(1);
  });

  it("respects the limit parameter", async () => {
    const result = await client.callTool({
      name: "search_charts",
      arguments: { repository_url: REPO_URL, limit: 2 },
    });

    expect(result.isError).toBeFalsy();
    const content = result.content as Array<{ type: string; text: string }>;
    const data = JSON.parse(content[0].text) as { charts: string[]; total: number };

    expect(data.charts.length).toBeLessThanOrEqual(2);
  });

  it("returns charts and total count when no filter is applied", async () => {
    const result = await client.callTool({
      name: "search_charts",
      arguments: { repository_url: REPO_URL },
    });

    expect(result.isError).toBeFalsy();
    const content = result.content as Array<{ type: string; text: string }>;
    const data = JSON.parse(content[0].text) as { charts: string[]; total: number };

    expect(Array.isArray(data.charts)).toBe(true);
    expect(data.charts.length).toBeGreaterThan(0);
    expect(typeof data.total).toBe("number");
    expect(data.total).toBeGreaterThan(0);
  });
});
