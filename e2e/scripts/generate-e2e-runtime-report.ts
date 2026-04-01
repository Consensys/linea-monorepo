import { Octokit } from "@octokit/rest";

type JobConclusion = "success" | "failure";
type SpecStatus = "PASS" | "FAIL" | "TIMEOUT";
type TestStatus = "passed" | "failed" | "skipped";

interface E2eRun {
  runId: number;
  runAttempt: number;
  jobId: number;
  jobConclusion: JobConclusion;
  startedAt: string;
  commitSha: string;
}

interface SpecResult {
  specFile: string;
  status: SpecStatus;
  durationSeconds: number;
  tests: TestResult[];
}

interface TestResult {
  name: string;
  durationMs: number;
  status: TestStatus;
}

interface E2eRunResult {
  run: E2eRun;
  specResults: SpecResult[];
}

interface ReportSummary {
  totalRuns: number;
  passedRuns: number;
  failedRuns: number;
  dateRange: string;
  slowestSpec: string;
  slowestSpecMedianSeconds: number;
}

const DEFAULT_DAYS = 30;
const OWNER = "Consensys";
const REPO = "linea-monorepo";
const WORKFLOW_FILE = "main.yml";
const E2E_JOB_NAME = "run-e2e-tests / run-e2e-tests";

const SPEC_FILE_RE = /^(PASS|FAIL)\s+(src\/[^/]+\.spec\.ts)(?:\s+\(([0-9.]+)\s*s\))?/;
const TEST_CASE_RE = /^\s+(✓|✕|×)\s+(.+?)\s+\((\d+)\s*ms\)/;
const SKIPPED_TEST_RE = /^\s+○\s+skipped\s+(.+)/;
const GHA_TIMESTAMP_RE = /^\d{4}-\d{2}-\d{2}T[\d:.]+Z\s/;

function parseArgs(): { days: number } {
  const args = process.argv.slice(2);
  const daysIndex = args.indexOf("--days");
  const days = daysIndex !== -1 && args[daysIndex + 1] ? parseInt(args[daysIndex + 1], 10) : DEFAULT_DAYS;
  if (isNaN(days) || days < 1 || days > 365) {
    console.error("Invalid --days value. Must be between 1 and 365.");
    process.exit(1);
  }
  return { days };
}

function parseJestLog(rawLog: string, jobConclusion: JobConclusion): SpecResult[] {
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

  // Timeout detection: if the job failed, check for known spec files that
  // never got a PASS/FAIL line. Exclude liveness.spec.ts since it runs in
  // a second command that is skipped when the first command fails.
  if (jobConclusion === "failure") {
    const parsedFiles = new Set(specResults.map((s) => s.specFile));
    const knownSpecFiles = [
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
    ];

    for (const specFile of knownSpecFiles) {
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

  // Validate against Jest summary line
  const JEST_SUMMARY_RE = /^Test Suites:\s+(?:(\d+) failed,\s+)?(?:(\d+) passed,\s+)?(\d+) total/;
  for (const rawLine of lines) {
    const line = rawLine.replace(GHA_TIMESTAMP_RE, "");
    const summaryMatch = line.match(JEST_SUMMARY_RE);
    if (summaryMatch) {
      const expectedTotal = parseInt(summaryMatch[3], 10);
      const actualParsed = specResults.filter((s) => s.status !== "TIMEOUT").length;
      if (actualParsed !== expectedTotal) {
        console.warn(
          `Warning: Jest reported ${expectedTotal} test suites but parser found ${actualParsed} spec results (some may be helper tests excluded by filter).`,
        );
      }
      break;
    }
  }

  return specResults;
}

async function fetchE2eRuns(octokit: Octokit, days: number): Promise<{ run: E2eRun; rawLog: string }[]> {
  const since = new Date();
  since.setDate(since.getDate() - days);
  const sinceStr = since.toISOString().split("T")[0];

  console.log(`Fetching workflow runs since ${sinceStr}...`);

  const runs = await octokit.paginate(octokit.actions.listWorkflowRuns, {
    owner: OWNER,
    repo: REPO,
    workflow_id: WORKFLOW_FILE,
    branch: "main",
    event: "push",
    status: "completed",
    created: `>${sinceStr}`,
    per_page: 100,
  });

  console.log(`Found ${runs.length} completed workflow runs.`);

  const results: { run: E2eRun; rawLog: string }[] = [];

  for (const workflowRun of runs) {
    const { data: jobsData } = await octokit.actions.listJobsForWorkflowRun({
      owner: OWNER,
      repo: REPO,
      run_id: workflowRun.id,
      filter: "latest",
      per_page: 100,
    });

    const e2eJob = jobsData.jobs.find(
      (job) =>
        job.name === E2E_JOB_NAME &&
        (job.conclusion === "success" || job.conclusion === "failure"),
    );

    if (!e2eJob) continue;

    let rawLog: string;
    try {
      const { data } = await octokit.actions.downloadJobLogsForWorkflowRun({
        owner: OWNER,
        repo: REPO,
        job_id: e2eJob.id,
      });
      rawLog = data as unknown as string;
    } catch (err) {
      console.warn(`Warning: could not download logs for job ${e2eJob.id} (run ${workflowRun.id}), skipping.`);
      continue;
    }

    if (!rawLog) {
      console.warn(`Warning: empty log for job ${e2eJob.id} (run ${workflowRun.id}), skipping.`);
      continue;
    }

    results.push({
      run: {
        runId: workflowRun.id,
        runAttempt: workflowRun.run_attempt ?? 1,
        jobId: e2eJob.id,
        jobConclusion: e2eJob.conclusion as JobConclusion,
        startedAt: e2eJob.started_at ?? workflowRun.created_at,
        commitSha: workflowRun.head_sha,
      },
      rawLog,
    });

    console.log(
      `  Run ${workflowRun.id} (${e2eJob.conclusion}): fetched ${rawLog.length} bytes of logs`,
    );
  }

  console.log(`Fetched logs for ${results.length} qualifying E2E runs.`);
  return results;
}

async function main(): Promise<void> {
  const { days } = parseArgs();
  const token = process.env.GITHUB_TOKEN;
  if (!token) {
    console.error("GITHUB_TOKEN environment variable is required.");
    process.exit(1);
  }

  const octokit = new Octokit({ auth: token });

  console.log(`Generating E2E runtime report for the last ${days} days...`);

  // Phase 1: Fetch
  const runsWithLogs = await fetchE2eRuns(octokit, days);
  if (runsWithLogs.length === 0) {
    console.log("No qualifying E2E runs found. Exiting.");
    process.exit(0);
  }

  // Phase 2: Parse
  const runResults: E2eRunResult[] = runsWithLogs.map(({ run, rawLog }) => ({
    run,
    specResults: parseJestLog(rawLog, run.jobConclusion),
  }));

  // Sort by date ascending
  runResults.sort((a, b) => new Date(a.run.startedAt).getTime() - new Date(b.run.startedAt).getTime());

  console.log(`Parsed ${runResults.length} runs.`);

  for (const result of runResults) {
    const specCount = result.specResults.length;
    const date = result.run.startedAt.split("T")[0];
    console.log(`  ${date} run ${result.run.runId}: ${specCount} spec files, conclusion=${result.run.jobConclusion}`);
  }

  // Phase 3: Render

  console.log("Done.");
}

main().catch((err) => {
  console.error("Fatal error:", err);
  process.exit(1);
});
