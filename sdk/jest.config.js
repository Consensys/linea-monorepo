module.exports = {
  preset: "ts-jest",
  testEnvironment: "node",
  rootDir: ".",
  testRegex: "test.ts$",
  verbose: true,
  collectCoverage: true,
  collectCoverageFrom: ["src/**/*.ts"],
  coverageReporters: ["html", "lcov", "text"],
  testPathIgnorePatterns: ["src/clients/typechain", "src/index.ts"],
  coveragePathIgnorePatterns: ["src/clients/typechain", "src/index.ts", "src/utils/testing/"],
};
