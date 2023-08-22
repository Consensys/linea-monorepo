module.exports = {
  preset: "ts-jest",
  testEnvironment: "node",
  rootDir: ".",
  testRegex: "test.ts$",
  verbose: true,
  collectCoverage: true,
  collectCoverageFrom: ["src/**/*.ts"],
  coverageReporters: ["html", "lcov", "text"],
  testPathIgnorePatterns: [
    "src/typechain/",
    "src/lib/postman/migrations/",
    "src/lib/postman/repositories/",
    "src/lib/index.ts",
  ],
  coveragePathIgnorePatterns: [
    "src/typechain/",
    "src/lib/postman/migrations/",
    "src/lib/postman/repositories/",
    "src/lib/index.ts",
  ],
};
