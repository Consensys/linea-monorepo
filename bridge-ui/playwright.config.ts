import { defineConfig, devices } from "@playwright/test";
import "dotenv/config";

const isUnit = process.env.UNIT === "true";

export default defineConfig({
  testDir: ".",
  testMatch: "**/*.spec.ts",
  // Timeout for tests that don't involve blockchain transactions
  timeout: 40_000,
  fullyParallel: true,
  maxFailures: process.env.CI ? 1 : 0,
  // To consider - cannot really run E2E tests involving blockchain tx in parallel. There is a high risk of reusing the same tx nonce -> leading to dropped transactions
  workers: process.env.CI ? 1 : undefined,
  reporter: isUnit
    ? undefined
    : process.env.CI
      ? [
          [
            "html",
            { open: "never", outputFolder: `playwright-report-${process.env.HEADLESS ? "headless" : "headful"}` },
          ],
          ["list"],
        ]
      : [["html"], ["list"]],
  use: {
    baseURL: "http://localhost:3000",
    trace: process.env.CI ? "on" : "retain-on-failure",
  },
  projects: [
    {
      name: "setup",
      testMatch: /global\.setup\.ts/,
    },
    {
      name: "chromium",
      testMatch: "test/**/*.spec.ts",
      use: { ...devices["Desktop Chrome"] },
      dependencies: ["setup"],
    },
    {
      name: "unit",
      testMatch: "src/**/*.spec.ts",
    },
  ],
  webServer: !isUnit
    ? {
        command: "pnpm run start",
        url: "http://localhost:3000",
        reuseExistingServer: true,
      }
    : undefined,
});
