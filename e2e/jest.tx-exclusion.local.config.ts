import type { Config } from "jest";

const config: Config = {
  preset: "ts-jest",
  testEnvironment: "node",
  rootDir: ".",
  testRegex: "transaction-exclusion.spec.ts$",
  verbose: true,
  setupFilesAfterEnv: ["./env-setup/setup-local-tx-exclusion.ts"],
  bail: 1,
  forceExit: true,
  detectOpenHandles: true,
};

export default config;
