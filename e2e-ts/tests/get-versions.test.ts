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

interface ChartVersion {
  version: string;
  app_version: string;
  created: string;
  deprecated: boolean;
}

interface GetVersionsResponse {
  versions: ChartVersion[];
  total: number;
}

const REPO_URL = "https://prometheus-community.github.io/helm-charts";

describe("get_versions", () => {
  it("returns versions with at least one entry", async () => {
    const result = await client.callTool({
      name: "get_versions",
      arguments: { repository_url: REPO_URL, chart_name: "alertmanager" },
    });

    expect(result.isError).toBeFalsy();
    const content = result.content as Array<{ type: string; text: string }>;
    const data = JSON.parse(content[0].text) as GetVersionsResponse;

    expect(Array.isArray(data.versions)).toBe(true);
    expect(data.versions.length).toBeGreaterThanOrEqual(1);
    expect(typeof data.total).toBe("number");
    expect(data.total).toBeGreaterThanOrEqual(1);
  });

  it("each version entry has version and app_version fields", async () => {
    const result = await client.callTool({
      name: "get_versions",
      arguments: { repository_url: REPO_URL, chart_name: "alertmanager" },
    });

    expect(result.isError).toBeFalsy();
    const content = result.content as Array<{ type: string; text: string }>;
    const data = JSON.parse(content[0].text) as GetVersionsResponse;

    for (const entry of data.versions) {
      expect(typeof entry.version).toBe("string");
      expect(entry.version.length).toBeGreaterThan(0);
      expect(typeof entry.app_version).toBe("string");
    }
  });

  it("versions are sorted in descending semver order", async () => {
    const result = await client.callTool({
      name: "get_versions",
      arguments: { repository_url: REPO_URL, chart_name: "alertmanager" },
    });

    expect(result.isError).toBeFalsy();
    const content = result.content as Array<{ type: string; text: string }>;
    const data = JSON.parse(content[0].text) as GetVersionsResponse;

    if (data.versions.length >= 2) {
      const parseParts = (v: string): number[] =>
        v.split(".").map((p) => parseInt(p.replace(/[^0-9]/g, ""), 10) || 0);

      const first = parseParts(data.versions[0].version);
      const second = parseParts(data.versions[1].version);

      for (let i = 0; i < Math.max(first.length, second.length); i++) {
        const a = first[i] ?? 0;
        const b = second[i] ?? 0;
        if (a > b) break;
        if (a < b) {
          expect(a).toBeGreaterThanOrEqual(b);
          break;
        }
      }
    }
  });

  it("respects the limit parameter", async () => {
    const result = await client.callTool({
      name: "get_versions",
      arguments: { repository_url: REPO_URL, chart_name: "alertmanager", limit: 3 },
    });

    expect(result.isError).toBeFalsy();
    const content = result.content as Array<{ type: string; text: string }>;
    const data = JSON.parse(content[0].text) as GetVersionsResponse;

    expect(data.versions.length).toBeLessThanOrEqual(3);
  });
});
