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
  // Phase 3: Render

  console.log("Done.");
}

main().catch((err) => {
  console.error("Fatal error:", err);
  process.exit(1);
});
