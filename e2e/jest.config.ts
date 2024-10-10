import type { Config } from "jest";

const config: Config = {
  preset: "ts-jest",
  testEnvironment: "node",
  rootDir: "src",
  testRegex: ".spec.ts$",
  verbose: true,
  globalSetup: "./config/jest/global-setup.ts",
  maxWorkers: "50%",
  testTimeout: 3 * 60 * 1000,
  workerThreads: true,
};

export default config;
