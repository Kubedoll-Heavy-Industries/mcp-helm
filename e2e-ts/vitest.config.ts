import { defineConfig } from "vitest/config";

export default defineConfig({
  test: {
    globalSetup: "./global-setup.ts",
    testTimeout: 120_000,
    hookTimeout: 120_000,
    teardownTimeout: 30_000,
    fileParallelism: false,
  },
});
