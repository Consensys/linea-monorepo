module.exports = {
  collectCoverage: true,
  collectCoverageFrom: ["./src/**/*.ts"],
  coverageDirectory: "coverage",
  coverageProvider: "babel",
  coverageReporters: ["html", "json-summary", "text"],
  coverageThreshold: {
    global: {
      branches: 85.71,
      functions: 100,
      lines: 95.23,
      statements: 95.34,
    },
  },
  preset: "ts-jest",
  resetMocks: true,
  restoreMocks: true,
  testTimeout: 2500,
  testPathIgnorePatterns: ["src/scripts", "src/index.ts"],
  coveragePathIgnorePatterns: ["src/scripts", "src/index.ts"],
};
