import type { Config } from "jest";

const config: Config = {
  preset: "ts-jest",
  testEnvironment: "node",
  rootDir: ".",
  testRegex: ".spec.ts$",
  verbose: true,
  globalSetup: "./config/global-setup.ts",
  maxWorkers: "50%",
  testTimeout: 3 * 60 * 1000,
};

export default config;
