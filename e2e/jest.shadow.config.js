module.exports = {
  preset: "ts-jest",
  testEnvironment: "node",
  rootDir: ".",
  testRegex: "(spec|test).ts$",
  verbose: true,
  setupFilesAfterEnv: ["./env-setup/setup-shadow.ts"],
};
