export type JobConclusion = "success" | "failure";
export type SpecStatus = "PASS" | "FAIL" | "TIMEOUT";
export type TestStatus = "passed" | "failed" | "skipped";

export interface TestResult {
  name: string;
  durationMs: number;
  status: TestStatus;
}

export interface SpecResult {
  specFile: string;
  status: SpecStatus;
  durationSeconds: number;
  tests: TestResult[];
}

const SPEC_FILE_RE = /^(PASS|FAIL)\s+(src\/[^/]+\.spec\.ts)(?:\s+\(([0-9.]+)\s*s\))?/;
const TEST_CASE_RE = /^\s+(✓|✕|×)\s+(.+?)\s+\((\d+)\s*ms\)/;
const SKIPPED_TEST_RE = /^\s+○\s+skipped\s+(.+)/;
const GHA_TIMESTAMP_RE = /^\d{4}-\d{2}-\d{2}T[\d:.]+Z\s/;
const JEST_SUMMARY_RE = /^Test Suites:\s+(?:(\d+) failed,\s+)?(?:(\d+) passed,\s+)?(\d+) total/;
const KNOWN_SPEC_FILES = [
  "src/bridge-tokens.spec.ts",
  "src/eip7702.spec.ts",
  "src/l2.spec.ts",
  "src/messaging.spec.ts",
  "src/opcodes.spec.ts",
  "src/restart.spec.ts",
  "src/send-bundle.spec.ts",
  "src/shomei-get-proof.spec.ts",
  "src/submission-finalization.spec.ts",
  "src/transaction-exclusion.spec.ts",
] as const;

export function parseJestLog(rawLog: string, jobConclusion: JobConclusion): SpecResult[] {
  const lines = rawLog.split("\n");
  const specResults: SpecResult[] = [];
  let currentSpec: SpecResult | null = null;

  for (const rawLine of lines) {
    const line = rawLine.replace(GHA_TIMESTAMP_RE, "");

    const specMatch = line.match(SPEC_FILE_RE);
    if (specMatch) {
      currentSpec = {
        specFile: specMatch[2],
        status: specMatch[1] as "PASS" | "FAIL",
        durationSeconds: specMatch[3] !== undefined ? parseFloat(specMatch[3]) : 0,
        tests: [],
      };
      specResults.push(currentSpec);
      continue;
    }

    if (!currentSpec) continue;

    const testMatch = line.match(TEST_CASE_RE);
    if (testMatch) {
      currentSpec.tests.push({
        name: testMatch[2],
        durationMs: parseInt(testMatch[3], 10),
        status: testMatch[1] === "✓" ? "passed" : "failed",
      });
      continue;
    }

    const skippedMatch = line.match(SKIPPED_TEST_RE);
    if (skippedMatch) {
      currentSpec.tests.push({
        name: skippedMatch[1],
        durationMs: 0,
        status: "skipped",
      });
    }
  }

  if (jobConclusion === "failure") {
    const parsedFiles = new Set(specResults.map((result) => result.specFile));

    for (const specFile of KNOWN_SPEC_FILES) {
      if (!parsedFiles.has(specFile)) {
        specResults.push({
          specFile,
          status: "TIMEOUT",
          durationSeconds: NaN,
          tests: [],
        });
      }
    }
  }

  for (const rawLine of lines) {
    const line = rawLine.replace(GHA_TIMESTAMP_RE, "");
    const summaryMatch = line.match(JEST_SUMMARY_RE);
    if (!summaryMatch) continue;

    const expectedTotal = parseInt(summaryMatch[3], 10);
    const actualParsed = specResults.filter((result) => result.status !== "TIMEOUT").length;

    if (actualParsed !== expectedTotal) {
      console.warn(
        `Warning: Jest reported ${expectedTotal} test suites but parser found ${actualParsed} spec results (some may be helper tests excluded by filter).`,
      );
    }

    break;
  }

  return specResults;
}
