import type { Config } from "jest";

// Define standard values used across the configuration
const MAX_TEST_TIMEOUT_MS = 3 * 60 * 1000; // 3 minutes for potentially slow integration/blockchain tests
const MAX_CONCURRENCY_WORKERS = 7;
const WORKERS_CPU_USAGE = "75%";


const config: Config = {
    // --- BASIC CONFIGURATION ---
    
    // Use ts-jest preset for TypeScript support
    preset: "ts-jest",
    // Test environment mimics a Node.js server environment (no browser DOM)
    testEnvironment: "node",
    // Base directory where Jest should look for files
    rootDir: "src",
    // Verbose logging of test results
    verbose: true,
    
    // --- TEST FILE LOCATION ---
    
    // Use the standard Jest testMatch pattern
    testMatch: ["**/?(*.)+(spec|test).[tj]s?(x)"], 
    // Removed testRegex: ".spec.ts$" to use the more flexible testMatch default.
    
    // --- SETUP & TEARDOWN ---

    // Path to the file run once before all tests (e.g., database connection setup)
    globalSetup: "./config/jest/global-setup.ts",
    // Path to the file run once after all tests (e.g., environment cleanup)
    globalTeardown: "./config/jest/global-teardown.ts",
    // Files executed once per test suite before tests start
    setupFilesAfterEnv: ["./config/jest/setup.ts"],

    // --- PERFORMANCE & RESOURCE MANAGEMENT ---
    
    // Set a long timeout for tests (critical for network/blockchain calls)
    testTimeout: MAX_TEST_TIMEOUT_MS,
    
    // Limit parallel test execution to prevent resource exhaustion
    maxConcurrency: MAX_CONCURRENCY_WORKERS,
    // Allocate maximum 75% of available CPU cores for workers
    maxWorkers: WORKERS_CPU_USAGE,
    // Use Node.js Worker Threads instead of child processes for worker execution
    workerThreads: true,
    
    // Forces Jest process to exit after test run (useful for preventing open handles/connections)
    forceExit: true,
};

export default config;
