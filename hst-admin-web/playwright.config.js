import { defineConfig } from "@playwright/test";

export default defineConfig({
  testDir: "./e2e",
  timeout: 120_000,
  expect: { timeout: 10_000 },
  use: {
    baseURL: process.env.PW_BASE_URL || "http://localhost:5175",
    viewport: { width: 1280, height: 800 },
    screenshot: "only-on-failure",
    trace: "off",
  },
  reporter: [["list"]],
});
