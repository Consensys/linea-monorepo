import type { Config } from "jest";

const config: Config = {
  preset: "ts-jest",
  testEnvironment: "node",
  rootDir: ".",
  testRegex: "ordered-test.ts$",
  verbose: true,
  setupFilesAfterEnv: ["./env-setup/setup-local.ts"],
  bail: 1,
  forceExit: true,
  detectOpenHandles: true,
};

export default config;
