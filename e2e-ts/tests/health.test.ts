import { describe, it, expect, inject } from "vitest";

const baseUrl = inject("baseUrl");

describe("health endpoints", () => {
  it("GET /healthz returns 200 with status field", async () => {
    const res = await fetch(`${baseUrl}/healthz`);
    expect(res.status).toBe(200);
    expect(res.headers.get("content-type")).toMatch(/application\/json/);
    const body = await res.json();
    expect(body).toHaveProperty("status");
  });

  it("GET /readyz returns 200 with status field", async () => {
    const res = await fetch(`${baseUrl}/readyz`);
    expect(res.status).toBe(200);
    expect(res.headers.get("content-type")).toMatch(/application\/json/);
    const body = await res.json();
    expect(body).toHaveProperty("status");
  });
});
