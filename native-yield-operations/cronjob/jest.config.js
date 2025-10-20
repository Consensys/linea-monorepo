module.exports = {
  preset: "ts-jest",
  testEnvironment: "node",
  rootDir: ".",
  testRegex: "test.ts$",
  verbose: true,
  collectCoverage: true,
  collectCoverageFrom: ["src/**/*.ts"],
  coverageReporters: ["html", "lcov", "text"],
  testPathIgnorePatterns: ["src/run.ts"],
  coveragePathIgnorePatterns: ["src/run.ts"],
};
