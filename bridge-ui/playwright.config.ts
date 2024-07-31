import { defineConfig, devices } from "@playwright/test";

export default defineConfig({
  testDir: "./test/e2e",
  timeout: 60_000,
  fullyParallel: true,
  maxFailures: process.env.CI ? 1 : 0,
  workers: process.env.CI ? 1 : undefined,
  reporter: process.env.CI
    ? [['html', { open: 'never', outputFolder: `playwright-report-${process.env.HEADLESS ? 'headless' : 'headful'}` }]]
    : 'html',
  use: {
    baseURL: "http://localhost:3000",
    trace: process.env.CI ? 'on' : 'retain-on-failure'
  },
  projects: [
    {
      name: "chromium",
      use: { ...devices["Desktop Chrome"] },
    },
  ],
  webServer: {
    command: "pnpm run start",
    url: "http://localhost:3000",
    reuseExistingServer: true,
  },
});
