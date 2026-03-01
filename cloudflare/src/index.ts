import { Container } from "cloudflare:workers";

interface Env {
  MCP_HELM: DurableObjectNamespace<McpHelmContainer>;
}

export class McpHelmContainer extends Container {
  defaultPort = 8012;
  sleepAfter = "5m";
}

export default {
  async fetch(request: Request, env: Env): Promise<Response> {
    const id = env.MCP_HELM.idFromName("default");
    const stub = env.MCP_HELM.get(id);
    return stub.fetch(request);
  },
};
