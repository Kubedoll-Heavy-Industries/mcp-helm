import path from "node:path";
import { GenericContainer, Wait } from "testcontainers";
import type { StartedTestContainer } from "testcontainers";
import type { GlobalSetupContext } from "vitest/node";

declare module "vitest" {
  export interface ProvidedContext {
    baseUrl: string;
  }
}

let container: StartedTestContainer;

export async function setup({ provide }: GlobalSetupContext) {
  const image = await GenericContainer.fromDockerfile(
    path.resolve(import.meta.dirname, ".."),
  )
    .withBuildkit()
    .build("mcp-helm-e2e:latest", { deleteOnExit: false });

  container = await image
    .withExposedPorts(8012)
    .withWaitStrategy(
      Wait.forHttp("/healthz", 8012).forStatusCode(200),
    )
    .withStartupTimeout(120_000)
    .start();

  // Use 127.0.0.1 explicitly to avoid IPv6 dual-stack issues with Node.js fetch
  const host = container.getHost().replace("localhost", "127.0.0.1");
  const baseUrl = `http://${host}:${container.getMappedPort(8012)}`;
  provide("baseUrl", baseUrl);
}

export async function teardown() {
  await container?.stop();
}
