module.exports = {
  preset: "ts-jest",
  testEnvironment: "node",
  rootDir: ".",
  testRegex: "(spec|test).ts$",
  verbose: true,
  globalSetup: "",
  setupFilesAfterEnv: ["./env-setup/setup-local.ts"],
  globalTeardown: "",
};
