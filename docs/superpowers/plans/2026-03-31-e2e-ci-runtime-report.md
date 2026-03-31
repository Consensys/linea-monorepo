# E2E CI Runtime Report - Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a GitHub Actions workflow that fetches E2E CI test runtimes from the last N days, parses Jest logs, generates a color-coded HTML report, and posts a summary to Slack.

**Architecture:** Single TypeScript script (`e2e/scripts/generate-e2e-runtime-report.ts`) with three phases - Fetch (Octokit API), Parse (regex state machine), Render (HTML template literals). Invoked by a `workflow_dispatch` GitHub Actions workflow that uploads the HTML as an artifact and posts summary stats to Slack.

**Tech Stack:** TypeScript, `@octokit/rest`, `tsx`, GitHub Actions, Slack Block Kit

**Spec:** `docs/superpowers/specs/2026-03-31-e2e-ci-runtime-report-design.md`

---

### Task 1: Add dependencies and script entry point

**Files:**
- Modify: `e2e/package.json`
- Create: `e2e/scripts/generate-e2e-runtime-report.ts` (skeleton only)

- [ ] **Step 1: Add `@octokit/rest` and `tsx` to `e2e/package.json`**

In `e2e/package.json`, add to `devDependencies`:

```json
"@octokit/rest": "22.0.1",
"tsx": "catalog:"
```

And add to `scripts`:

```json
"generate-runtime-report": "tsx scripts/generate-e2e-runtime-report.ts"
```

- [ ] **Step 2: Create skeleton script**

Create `e2e/scripts/generate-e2e-runtime-report.ts` with argument parsing and the main function shell:

```typescript
import { Octokit } from "@octokit/rest";

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
  // Phase 2: Parse
  // Phase 3: Render

  console.log("Done.");
}

main().catch((err) => {
  console.error("Fatal error:", err);
  process.exit(1);
});
```

- [ ] **Step 3: Run `pnpm install` to update lockfile**

Run: `pnpm install` (from workspace root)
Expected: lockfile updated, no errors

- [ ] **Step 4: Verify the skeleton runs**

Run: `cd e2e && GITHUB_TOKEN=fake npx tsx scripts/generate-e2e-runtime-report.ts --days 5`
Expected: prints "Generating E2E runtime report for the last 5 days..." then "Done."

- [ ] **Step 5: Commit**

```bash
git add e2e/package.json e2e/scripts/generate-e2e-runtime-report.ts pnpm-lock.yaml
git commit -m "feat(e2e): add skeleton for E2E runtime report script

Add @octokit/rest and tsx dependencies. Create script entry point
with argument parsing and main function shell."
```

---

### Task 2: Implement Phase 1 - Fetch E2E runs and logs

**Files:**
- Modify: `e2e/scripts/generate-e2e-runtime-report.ts`

- [ ] **Step 1: Define data model types**

Add types at the top of the file (after imports):

```typescript
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
```

- [ ] **Step 2: Implement `fetchE2eRuns()`**

```typescript
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
```

- [ ] **Step 3: Wire fetch into main()**

Replace the `// Phase 1: Fetch` comment in `main()` with:

```typescript
  const runsWithLogs = await fetchE2eRuns(octokit, days);
  if (runsWithLogs.length === 0) {
    console.log("No qualifying E2E runs found. Exiting.");
    process.exit(0);
  }
```

- [ ] **Step 4: Verify fetch works against the real API**

Run (requires a valid GitHub token with `actions:read` scope):

```bash
cd e2e && GITHUB_TOKEN=$GITHUB_TOKEN npx tsx scripts/generate-e2e-runtime-report.ts --days 3
```

Expected: prints "Fetching workflow runs since YYYY-MM-DD..." followed by lines like "Run XXXXX (success): fetched XXXXX bytes of logs" and "Fetched logs for N qualifying E2E runs."

- [ ] **Step 5: Commit**

```bash
git add e2e/scripts/generate-e2e-runtime-report.ts
git commit -m "feat(e2e): implement Phase 1 - fetch E2E CI runs and logs

Add data model types and fetchE2eRuns() using Octokit to list
workflow runs, find E2E jobs, and download their logs."
```

---

### Task 3: Implement Phase 2 - Parse Jest logs

**Files:**
- Modify: `e2e/scripts/generate-e2e-runtime-report.ts`

- [ ] **Step 1: Implement `parseJestLog()`**

```typescript
const SPEC_FILE_RE = /^(PASS|FAIL)\s+(src\/[^/]+\.spec\.ts)(?:\s+\(([0-9.]+)\s*s\))?/;
const TEST_CASE_RE = /^\s+(✓|✕|×)\s+(.+?)\s+\((\d+)\s*ms\)/;
const SKIPPED_TEST_RE = /^\s+○\s+skipped\s+(.+)/;
const GHA_TIMESTAMP_RE = /^\d{4}-\d{2}-\d{2}T[\d:.]+Z\s/;

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
```

- [ ] **Step 2: Wire parse into main() and build E2eRunResult array**

Replace the `// Phase 2: Parse` comment in `main()` with:

```typescript
  const runResults: E2eRunResult[] = runsWithLogs.map(({ run, rawLog }) => ({
    run,
    specResults: parseJestLog(rawLog, run.jobConclusion),
  }));

  // Sort by date ascending
  runResults.sort((a, b) => new Date(a.run.startedAt).getTime() - new Date(b.run.startedAt).getTime());

  console.log(`Parsed ${runResults.length} runs.`);
```

- [ ] **Step 3: Add a quick validation log**

Add after the parse step:

```typescript
  for (const result of runResults) {
    const specCount = result.specResults.length;
    const date = result.run.startedAt.split("T")[0];
    console.log(`  ${date} run ${result.run.runId}: ${specCount} spec files, conclusion=${result.run.jobConclusion}`);
  }
```

- [ ] **Step 4: Verify parsing against real data**

Run:

```bash
cd e2e && GITHUB_TOKEN=$GITHUB_TOKEN npx tsx scripts/generate-e2e-runtime-report.ts --days 3
```

Expected: lines like "2026-03-30 run 12345: 11 spec files, conclusion=success" showing correct spec file counts. Failed runs should show TIMEOUT entries for missing specs.

- [ ] **Step 5: Commit**

```bash
git add e2e/scripts/generate-e2e-runtime-report.ts
git commit -m "feat(e2e): implement Phase 2 - parse Jest logs

Add parseJestLog() with regex state machine to extract spec-level and
test-case-level results. Handle PASS/FAIL/TIMEOUT, skipped tests, and
optional GHA timestamp prefixes."
```

---

### Task 4: Implement Phase 3 - Render HTML report

**Files:**
- Modify: `e2e/scripts/generate-e2e-runtime-report.ts`

- [ ] **Step 1: Implement stats helper functions**

```typescript
function median(values: number[]): number {
  if (values.length === 0) return 0;
  const sorted = [...values].sort((a, b) => a - b);
  const mid = Math.floor(sorted.length / 2);
  return sorted.length % 2 !== 0 ? sorted[mid] : (sorted[mid - 1] + sorted[mid]) / 2;
}

function stddev(values: number[]): number {
  if (values.length < 2) return 0;
  const mean = values.reduce((a, b) => a + b, 0) / values.length;
  const variance = values.reduce((sum, v) => sum + (v - mean) ** 2, 0) / values.length;
  return Math.sqrt(variance);
}

type ColorLevel = "green" | "yellow" | "red" | "gray";

function getCellColor(duration: number, med: number, sd: number, status: SpecStatus): ColorLevel {
  if (status === "FAIL" || status === "TIMEOUT") return "red";
  if (isNaN(duration)) return "gray";
  if (sd === 0) return "green";
  if (duration > med + 2 * sd) return "red";
  if (duration > med + 1 * sd) return "yellow";
  return "green";
}

const COLOR_MAP: Record<ColorLevel, string> = {
  green: "#d4edda",
  yellow: "#fff3cd",
  red: "#f8d7da",
  gray: "#e2e3e5",
};
```

- [ ] **Step 2: Implement `renderHtmlReport()`**

This is the largest function. It builds the HTML string with template literals. The structure is:
1. HTML head with inline CSS
2. Header section (title, date range, run counts)
3. Summary table (one row per spec, median/min/max)
4. Timeline table (one row per spec, one column per run, color-coded cells)
5. Detail sections (expandable per-test-case tables)

```typescript
function renderHtmlReport(runResults: E2eRunResult[], days: number): { html: string; summary: ReportSummary } {
  const allSpecFiles = [...new Set(runResults.flatMap((r) => r.specResults.map((s) => s.specFile)))].sort();

  // Compute per-spec stats from successful runs
  const specStats = new Map<string, { med: number; sd: number; min: number; max: number; count: number }>();
  for (const specFile of allSpecFiles) {
    const durations = runResults
      .flatMap((r) => r.specResults)
      .filter((s) => s.specFile === specFile && s.status === "PASS" && !isNaN(s.durationSeconds))
      .map((s) => s.durationSeconds);

    specStats.set(specFile, {
      med: median(durations),
      sd: stddev(durations),
      min: durations.length > 0 ? Math.min(...durations) : NaN,
      max: durations.length > 0 ? Math.max(...durations) : NaN,
      count: durations.length,
    });
  }

  // Sort spec files by median duration descending
  const sortedSpecs = [...allSpecFiles].sort((a, b) => (specStats.get(b)?.med ?? 0) - (specStats.get(a)?.med ?? 0));

  const totalRuns = runResults.length;
  const passedRuns = runResults.filter((r) => r.run.jobConclusion === "success").length;
  const failedRuns = totalRuns - passedRuns;

  const startDate = runResults[0].run.startedAt.split("T")[0];
  const endDate = runResults[runResults.length - 1].run.startedAt.split("T")[0];
  const dateRange = `${startDate} to ${endDate}`;

  const slowestEntry = sortedSpecs[0];
  const slowestMedian = specStats.get(slowestEntry)?.med ?? 0;

  const summary: ReportSummary = {
    totalRuns,
    passedRuns,
    failedRuns,
    dateRange,
    slowestSpec: slowestEntry,
    slowestSpecMedianSeconds: Math.round(slowestMedian * 10) / 10,
  };

  // Build HTML
  const css = `
    body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; margin: 20px; color: #333; }
    h1, h2, h3 { color: #1a1a1a; }
    table { border-collapse: collapse; width: 100%; margin-bottom: 20px; }
    th, td { border: 1px solid #ddd; padding: 6px 10px; text-align: center; font-size: 13px; }
    th { background-color: #f5f5f5; font-weight: 600; position: sticky; top: 0; }
    td.spec-name { text-align: left; font-family: monospace; font-size: 12px; white-space: nowrap; }
    .header-stats { display: flex; gap: 24px; margin-bottom: 16px; }
    .stat { font-size: 14px; }
    .stat strong { font-size: 20px; }
    details { margin-bottom: 12px; }
    summary { cursor: pointer; font-weight: 600; font-family: monospace; padding: 4px 0; }
    .detail-table td { font-size: 12px; }
    .skipped { color: #888; font-style: italic; }
    .timeline-container { overflow-x: auto; }
  `;

  // Timeline table header
  const timelineHeaders = runResults
    .map((r) => {
      const date = r.run.startedAt.split("T")[0].slice(5); // MM-DD
      const sha = r.run.commitSha.slice(0, 7);
      const icon = r.run.jobConclusion === "failure" ? " ✗" : "";
      return `<th title="${r.run.startedAt}\n${r.run.commitSha}">${date}<br>${sha}${icon}</th>`;
    })
    .join("\n            ");

  // Timeline table rows
  const timelineRows = sortedSpecs
    .map((specFile) => {
      const stats = specStats.get(specFile)!;
      const cells = runResults
        .map((r) => {
          const spec = r.specResults.find((s) => s.specFile === specFile);
          if (!spec) return `<td style="background:${COLOR_MAP.gray}">-</td>`;

          const color = getCellColor(spec.durationSeconds, stats.med, stats.sd, spec.status);
          if (spec.status === "TIMEOUT") return `<td style="background:${COLOR_MAP.red}">TIMEOUT</td>`;
          if (spec.status === "FAIL") return `<td style="background:${COLOR_MAP.red}">FAIL ${spec.durationSeconds.toFixed(1)}s</td>`;
          return `<td style="background:${COLOR_MAP[color]}">${spec.durationSeconds.toFixed(1)}s</td>`;
        })
        .join("");

      const shortName = specFile.replace("src/", "").replace(".spec.ts", "");
      return `<tr><td class="spec-name">${shortName}</td>${cells}</tr>`;
    })
    .join("\n          ");

  // Summary table rows (color-code median cell: red if any run had FAIL/TIMEOUT)
  const summaryRows = sortedSpecs
    .map((specFile) => {
      const stats = specStats.get(specFile)!;
      const shortName = specFile.replace("src/", "").replace(".spec.ts", "");
      const hasFailures = runResults.some((r) =>
        r.specResults.some((s) => s.specFile === specFile && (s.status === "FAIL" || s.status === "TIMEOUT")),
      );
      const medColor: ColorLevel = isNaN(stats.med) ? "gray" : hasFailures ? "red" : "green";
      return `<tr>
            <td class="spec-name">${shortName}</td>
            <td style="background:${COLOR_MAP[medColor]}">${isNaN(stats.med) ? "-" : stats.med.toFixed(1)}</td>
            <td>${isNaN(stats.min) ? "-" : stats.min.toFixed(1)}</td>
            <td>${isNaN(stats.max) ? "-" : stats.max.toFixed(1)}</td>
            <td>${stats.count}</td>
          </tr>`;
    })
    .join("\n          ");

  // Detail sections
  const detailSections = sortedSpecs
    .map((specFile) => {
      const shortName = specFile.replace("src/", "").replace(".spec.ts", "");
      const rows = runResults
        .flatMap((r) => {
          const spec = r.specResults.find((s) => s.specFile === specFile);
          if (!spec || spec.tests.length === 0) return [];
          const date = r.run.startedAt.split("T")[0];
          return spec.tests.map(
            (t) =>
              `<tr class="${t.status === "skipped" ? "skipped" : ""}">
                <td>${date}</td>
                <td style="text-align:left">${t.name}</td>
                <td>${t.status === "skipped" ? "-" : t.durationMs}</td>
                <td>${t.status}</td>
              </tr>`,
          );
        })
        .join("\n            ");

      return `<details>
        <summary>${shortName}</summary>
        <table class="detail-table">
          <thead><tr><th>Date</th><th>Test Name</th><th>Duration (ms)</th><th>Status</th></tr></thead>
          <tbody>
            ${rows}
          </tbody>
        </table>
      </details>`;
    })
    .join("\n      ");

  const html = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>E2E CI Runtime Report</title>
  <style>${css}</style>
</head>
<body>
  <h1>E2E CI Runtime Report</h1>
  <div class="header-stats">
    <div class="stat"><strong>${totalRuns}</strong> runs</div>
    <div class="stat"><strong>${passedRuns}</strong> passed</div>
    <div class="stat"><strong>${failedRuns}</strong> failed</div>
    <div class="stat">Period: ${dateRange}</div>
    <div class="stat">Generated: ${new Date().toISOString()}</div>
  </div>

  <h2>Summary</h2>
  <table>
    <thead><tr><th>Spec File</th><th>Median (s)</th><th>Min (s)</th><th>Max (s)</th><th>Runs</th></tr></thead>
    <tbody>
      ${summaryRows}
    </tbody>
  </table>

  <h2>Timeline</h2>
  <div class="timeline-container">
    <table>
      <thead><tr><th>Spec File</th>${timelineHeaders}</tr></thead>
      <tbody>
        ${timelineRows}
      </tbody>
    </table>
  </div>

  <h2>Test Case Details</h2>
  ${detailSections}
</body>
</html>`;

  return { html, summary };
}
```

- [ ] **Step 3: Wire render into main() and write output files**

At the top of the file, add after the Octokit import:

```typescript
import { writeFileSync } from "node:fs";
```

Replace the `// Phase 3: Render` comment in `main()` with:

```typescript
  const { html, summary } = renderHtmlReport(runResults, days);

  writeFileSync("e2e-runtime-report.html", html);
  writeFileSync("e2e-runtime-report-summary.json", JSON.stringify(summary));

  console.log(`Report written to e2e-runtime-report.html`);
  console.log(`Summary: ${JSON.stringify(summary)}`);
```

- [ ] **Step 4: Test the full pipeline end-to-end**

Run:

```bash
cd e2e && GITHUB_TOKEN=$GITHUB_TOKEN npx tsx scripts/generate-e2e-runtime-report.ts --days 3
```

Expected: creates `e2e/e2e-runtime-report.html` and `e2e/e2e-runtime-report-summary.json`. Open the HTML file in a browser to visually inspect:
- Header shows correct run counts
- Summary table has all spec files with median/min/max
- Timeline table has color-coded cells
- Detail sections expand to show per-test-case timings

- [ ] **Step 5: Add output files to `.gitignore`**

Check if `e2e/.gitignore` exists. If so, add:

```
e2e-runtime-report.html
e2e-runtime-report-summary.json
```

If no `e2e/.gitignore` exists, create it with those entries. Alternatively, add to the root `.gitignore`.

- [ ] **Step 6: Commit**

```bash
git add e2e/scripts/generate-e2e-runtime-report.ts e2e/.gitignore
git commit -m "feat(e2e): implement Phase 3 - render HTML report

Add stats helpers (median, stddev, color thresholds) and
renderHtmlReport() generating a self-contained HTML file with
summary table, color-coded timeline, and expandable test details."
```

---

### Task 5: Create the GitHub Actions workflow

**Files:**
- Create: `.github/workflows/e2e-runtime-report.yml`

- [ ] **Step 1: Create the workflow file**

Create `.github/workflows/e2e-runtime-report.yml` with the full content from the spec. Reference: `docs/superpowers/specs/2026-03-31-e2e-ci-runtime-report-design.md` - Workflow section.

```yaml
name: e2e-runtime-report

on:
  workflow_dispatch:
    inputs:
      days:
        description: "Number of days to look back"
        required: false
        type: number
        default: 30

permissions:
  contents: read
  actions: read

jobs:
  generate-report:
    runs-on: gha-runner-scale-set-ubuntu-22.04-amd64-small
    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Setup nodejs environment
        uses: ./.github/actions/setup-nodejs
        with:
          pnpm-install-options: '-F "e2e..." --frozen-lockfile --prefer-offline'

      - name: Generate E2E runtime report
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        run: pnpm -F e2e run generate-runtime-report -- --days ${{ inputs.days }}

      - name: Read summary
        id: summary
        run: |
          echo "json<<EOF" >> $GITHUB_OUTPUT
          cat e2e/e2e-runtime-report-summary.json >> $GITHUB_OUTPUT
          echo "EOF" >> $GITHUB_OUTPUT

      - name: Upload HTML report
        uses: actions/upload-artifact@v4
        with:
          name: e2e-runtime-report
          path: e2e/e2e-runtime-report.html

      - name: Post to Slack
        uses: slackapi/slack-github-action@91efab103c0de0a537f72a35f6b8cda0ee76bf0a #v2.1.1
        with:
          method: chat.postMessage
          token: ${{ secrets.SLACK_BOT_TOKEN }}
          payload: |
            channel: ${{ secrets.SLACK_ENGINEERING_ALERTS_CHANNEL_ID }}
            text: "E2E Runtime Report - ${{ fromJSON(steps.summary.outputs.json).totalRuns }} runs (${{ fromJSON(steps.summary.outputs.json).passedRuns }} passed, ${{ fromJSON(steps.summary.outputs.json).failedRuns }} failed)"
            blocks:
              - type: header
                text:
                  type: plain_text
                  text: "E2E Runtime Report"
                  emoji: true

              - type: rich_text
                elements:
                  - type: rich_text_section
                    elements:
                      - type: text
                        text: "Date range: "
                        style:
                          bold: true
                      - type: text
                        text: "${{ fromJSON(steps.summary.outputs.json).dateRange }}"

              - type: rich_text
                elements:
                  - type: rich_text_section
                    elements:
                      - type: text
                        text: "Runs: "
                        style:
                          bold: true
                      - type: text
                        text: "${{ fromJSON(steps.summary.outputs.json).totalRuns }} total, ${{ fromJSON(steps.summary.outputs.json).passedRuns }} passed, ${{ fromJSON(steps.summary.outputs.json).failedRuns }} failed"

              - type: rich_text
                elements:
                  - type: rich_text_section
                    elements:
                      - type: text
                        text: "Slowest spec: "
                        style:
                          bold: true
                      - type: text
                        text: "${{ fromJSON(steps.summary.outputs.json).slowestSpec }} (median ${{ fromJSON(steps.summary.outputs.json).slowestSpecMedianSeconds }}s)"

              - type: rich_text
                elements:
                  - type: rich_text_section
                    elements:
                      - type: text
                        text: "Full report: "
                        style:
                          bold: true
                      - type: link
                        url: "${{ github.server_url }}/${{ github.repository }}/actions/runs/${{ github.run_id }}"
                        text: "View workflow run artifacts"
```

- [ ] **Step 2: Validate YAML syntax**

Run: `npx yaml-lint .github/workflows/e2e-runtime-report.yml` (or skip if yaml-lint is not available - GitHub Actions will validate on push)
Expected: no errors

- [ ] **Step 3: Commit**

```bash
git add .github/workflows/e2e-runtime-report.yml
git commit -m "feat(ci): add E2E runtime report workflow

workflow_dispatch triggered, generates HTML report from last N days
of E2E CI runs and posts summary to Slack."
```

---

### Task 6: Lint and final verification

**Files:**
- Modify: `e2e/scripts/generate-e2e-runtime-report.ts` (if lint fixes needed)

- [ ] **Step 1: Run ESLint on the script**

Run: `pnpm -F e2e run lint`
Expected: no errors on the new script file. If there are lint errors, fix them.

- [ ] **Step 2: Run a full end-to-end test with 30 days**

Run:

```bash
cd e2e && GITHUB_TOKEN=$GITHUB_TOKEN npx tsx scripts/generate-e2e-runtime-report.ts --days 30
```

Expected: completes without errors, generates both output files, HTML report looks correct when opened in browser.

- [ ] **Step 3: Verify output file sizes are reasonable**

Run: `ls -lh e2e/e2e-runtime-report.html e2e/e2e-runtime-report-summary.json`
Expected: HTML < 1MB, JSON < 1KB

- [ ] **Step 4: Commit any lint fixes**

```bash
git add -A
git commit -m "chore(e2e): lint fixes for runtime report script"
```

(Skip this commit if no lint fixes were needed.)
