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

describe("get_notes", () => {
  it("returns non-empty notes and version for alertmanager", async () => {
    const result = await client.callTool({
      name: "get_notes",
      arguments: { repository_url: REPO_URL, chart_name: "alertmanager" },
    });

    expect(result.isError).toBeFalsy();
    const content = result.content as Array<{ type: string; text: string }>;
    const parsed = JSON.parse(content[0].text);

    expect(parsed.version).toBeTruthy();
    expect(typeof parsed.version).toBe("string");
    expect(parsed.notes).toBeTruthy();
    expect(typeof parsed.notes).toBe("string");
    expect(parsed.notes.length).toBeGreaterThan(0);
  });

  it("notes content contains deployment-related template text", async () => {
    const result = await client.callTool({
      name: "get_notes",
      arguments: { repository_url: REPO_URL, chart_name: "alertmanager" },
    });

    expect(result.isError).toBeFalsy();
    const content = result.content as Array<{ type: string; text: string }>;
    const parsed = JSON.parse(content[0].text);

    expect(parsed.notes).toMatch(/URL|port|service|running|commands/i);
  });
});
