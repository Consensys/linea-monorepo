import type { Config } from "jest";

const config: Config = {
  preset: "ts-jest",
  testEnvironment: "node",
  rootDir: "src",
  testRegex: "gas-limit.spec.ts$",
  verbose: true,
  globalSetup: "./config/jest/global-setup.ts",
  globalTeardown: "./config/jest/global-teardown.ts",
  testTimeout: 3 * 60 * 1000,
  maxConcurrency: 7,
  maxWorkers: "75%",
  workerThreads: true,
  forceExit: true,
};

export default config;
