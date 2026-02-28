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

describe("get_values", () => {
  it("returns YAML values and version for alertmanager", async () => {
    const result = await client.callTool({
      name: "get_values",
      arguments: { repository_url: REPO_URL, chart_name: "alertmanager" },
    });

    expect(result.isError).toBeFalsy();
    const content = result.content as Array<{ type: string; text: string }>;
    const parsed = JSON.parse(content[0].text);

    expect(parsed.version).toBeTruthy();
    expect(typeof parsed.version).toBe("string");
    expect(parsed.values).toBeTruthy();
    expect(typeof parsed.values).toBe("string");
  });

  it("accepts include_schema parameter", async () => {
    const result = await client.callTool({
      name: "get_values",
      arguments: {
        repository_url: REPO_URL,
        chart_name: "prometheus-pushgateway",
        include_schema: true,
      },
    });

    expect(result.isError).toBeFalsy();
    const content = result.content as Array<{ type: string; text: string }>;
    const parsed = JSON.parse(content[0].text);

    expect(parsed.version).toBeTruthy();
    expect(parsed.values).toBeTruthy();
  });

  it("narrows to subtree when path is specified", async () => {
    const result = await client.callTool({
      name: "get_values",
      arguments: {
        repository_url: REPO_URL,
        chart_name: "alertmanager",
        path: ".image",
      },
    });

    expect(result.isError).toBeFalsy();
    const content = result.content as Array<{ type: string; text: string }>;
    const parsed = JSON.parse(content[0].text);

    expect(parsed.values).toBeTruthy();
    expect(parsed.path).toBe(".image");
    expect(parsed.values).toMatch(/repository|tag|pullPolicy/);
  });
});
