module.exports = {
  preset: "ts-jest",
  testEnvironment: "node",
  rootDir: ".",
  testRegex: "test.ts$",
  verbose: true,
  collectCoverage: true,
  collectCoverageFrom: ["src/**/*.ts"],
  coverageReporters: ["html", "lcov", "text"],
  testPathIgnorePatterns: ["src/common/index.ts", "src/ethTransfer/index.ts", "src/synctx/src/index.ts"],
  coveragePathIgnorePatterns: ["src/common/index.ts", "src/ethTransfer/index.ts", "src/synctx/src/index.ts"],
};
