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

describe("get_dependencies", () => {
  it("returns non-empty dependencies for kube-prometheus-stack", async () => {
    const result = await client.callTool({
      name: "get_dependencies",
      arguments: { repository_url: REPO_URL, chart_name: "kube-prometheus-stack" },
    });

    expect(result.isError).toBeFalsy();
    const content = result.content as Array<{ type: string; text: string }>;
    const parsed = JSON.parse(content[0].text);

    expect(parsed.version).toBeTruthy();
    expect(Array.isArray(parsed.dependencies)).toBe(true);
    expect(parsed.dependencies.length).toBeGreaterThan(0);

    const depNames = parsed.dependencies.map((d: { name: string }) => d.name);
    expect(depNames).toContain("grafana");
    expect(depNames).toContain("kube-state-metrics");
  });

  it("each dependency has name and version fields", async () => {
    const result = await client.callTool({
      name: "get_dependencies",
      arguments: { repository_url: REPO_URL, chart_name: "kube-prometheus-stack" },
    });

    expect(result.isError).toBeFalsy();
    const content = result.content as Array<{ type: string; text: string }>;
    const parsed = JSON.parse(content[0].text);

    for (const dep of parsed.dependencies) {
      expect(dep.name).toBeTruthy();
      expect(typeof dep.name).toBe("string");
      expect(dep.version).toBeTruthy();
      expect(typeof dep.version).toBe("string");
    }

    // At least some dependencies should have a repository field
    const depsWithRepo = parsed.dependencies.filter(
      (d: { repository?: string }) => d.repository,
    );
    expect(depsWithRepo.length).toBeGreaterThan(0);
  });

  it("returns empty dependencies array for a chart without deps", async () => {
    const result = await client.callTool({
      name: "get_dependencies",
      arguments: { repository_url: REPO_URL, chart_name: "alertmanager" },
    });

    expect(result.isError).toBeFalsy();
    const content = result.content as Array<{ type: string; text: string }>;
    const parsed = JSON.parse(content[0].text);

    expect(parsed.version).toBeTruthy();
    expect(Array.isArray(parsed.dependencies)).toBe(true);
    expect(parsed.dependencies.length).toBe(0);
  });
});
