import type { Config } from "jest";

const config: Config = {
  preset: "ts-jest",
  testEnvironment: "node",
  rootDir: "src",
  testRegex: ".spec.ts$",
  verbose: true,
  maxWorkers: "50%",
  testTimeout: 3 * 60 * 1000,
};

export default config;
