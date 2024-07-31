import type { Config } from "jest";

const config: Config = {
  preset: "ts-jest",
  testEnvironment: "node",
  rootDir: ".",
  testRegex: "(spec|test).ts$",
  verbose: true,
  setupFilesAfterEnv: ["./env-setup/setup-uat.ts"],
};

export default config;
