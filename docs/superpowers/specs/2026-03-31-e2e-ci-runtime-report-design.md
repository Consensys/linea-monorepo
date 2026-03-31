# E2E CI Runtime Report - Design Spec

## Overview

Automated workflow that fetches the last N days of E2E test CI runs from `main` branch pushes, parses Jest output logs to extract per-spec-file and per-test-case runtimes, generates an HTML report with color-coded trend tables, and posts a summary to Slack.

## Trigger

- GitHub Actions `workflow_dispatch` with configurable `days` input (default: 30)
- Triggered externally via Cursor Automation (out of scope for this spec)

## Architecture

```
workflow_dispatch (days=30)
  |
  v
e2e-runtime-report.yml
  |
  v
generate-e2e-runtime-report.ts
  |
  +-- Phase 1: Fetch (Octokit)
  |     List workflow runs -> List jobs -> Download logs
  |
  +-- Phase 2: Parse (regex)
  |     Strip GHA timestamps -> Match PASS/FAIL lines -> Match test cases
  |
  +-- Phase 3: Render (HTML template literals)
  |     Summary stats -> Color-coded tables -> Expandable detail sections
  |
  +-- Output:
        e2e-runtime-report.html
        e2e-runtime-report-summary.json
  |
  v
Upload HTML as workflow artifact
  |
  v
Post summary to Slack (Block Kit via slackapi/slack-github-action)
```

## Data Model

```typescript
// GitHub API job conclusion values
type JobConclusion = "success" | "failure";

// Jest spec-level result status
type SpecStatus = "PASS" | "FAIL" | "TIMEOUT";

// Jest test-case-level result status
type TestStatus = "passed" | "failed" | "skipped";

interface E2eRun {
  runId: number;
  runAttempt: number;
  jobId: number;
  jobConclusion: JobConclusion;
  startedAt: string; // ISO date from the job
  commitSha: string;
}

interface SpecResult {
  specFile: string; // e.g. "src/shomei-get-proof.spec.ts"
  status: SpecStatus;
  durationSeconds: number; // file-level duration, NaN if unavailable
  tests: TestResult[];
}

interface TestResult {
  name: string; // full test name
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
  dateRange: string; // "2026-03-01 to 2026-03-31"
  slowestSpec: string;
  slowestSpecMedianSeconds: number;
}
```

## Phase 1: Fetch

**Script:** `e2e/scripts/generate-e2e-runtime-report.ts`

**Dependencies:** `@octokit/rest` (added to `e2e/package.json` devDependencies, pinned exact version), `tsx` (via workspace catalog: `"tsx": "catalog:"`) for execution.

**Authentication:** `GITHUB_TOKEN` environment variable (provided by the workflow via `${{ secrets.GITHUB_TOKEN }}`).

**API calls:**

1. `octokit.actions.listWorkflowRuns()` - params:
   - `owner`: `Consensys`
   - `repo`: `linea-monorepo`
   - `workflow_id`: `main.yml`
   - `branch`: `main`
   - `event`: `push`
   - `created`: `>YYYY-MM-DD` (computed from `days` input)
   - `status`: `completed`
   - `per_page`: 100
   - Paginate through all results via `octokit.paginate()`

2. `octokit.actions.listJobsForWorkflowRun()` - for each run:
   - Params: `owner`, `repo`, `run_id`, `filter: "latest"` (returns only the latest attempt's jobs, avoiding the need to specify `attempt_number`)
   - Filter results for job name `run-e2e-tests / run-e2e-tests`
   - Filter for conclusion `success` or `failure` (excludes `skipped` runs where E2E was not needed)

3. `octokit.actions.downloadJobLogsForWorkflowRun()` - for each matching job:
   - Access raw log text via `response.data` (string). The API issues a 302 redirect to a blob URL; Octokit follows it automatically.
   - Handle empty responses (e.g., job deleted before log retrieval) by skipping with a warning.

**Rate limiting:** ~3 API calls per qualifying run. For 30 days with ~30-60 runs, that's ~90-180 calls. Well within the 1000/hour authenticated limit.

## Phase 2: Parse

**Input:** Raw CI log text (may include GHA timestamp prefixes).

**Scope:** Only `src/*.spec.ts` files (top-level). Excludes helper tests like `src/common/test-helpers/*.spec.ts`.

**Parsing rules:**

1. **Strip GHA timestamps** - optional prefix: `/^(\d{4}-\d{2}-\d{2}T[\d:.]+Z\s)?/`

2. **Match spec file results** - regex: `/^(PASS|FAIL)\s+(src\/[^/]+\.spec\.ts)(?:\s+\(([0-9.]+)\s*s\))?/`
   - The path pattern `src/[^/]+\.spec\.ts` ensures only top-level spec files match (no nested paths like `src/common/test-helpers/...`)
   - Duration group is optional (Jest omits it for suites that complete in <1s)
   - Captures: (1) status, (2) file path, (3) duration in seconds or undefined

3. **Match test case results** - regex: `/^\s+(\u2713|\u2715|\u00d7)\s+(.+?)\s+\((\d+)\s*ms\)/`
   - `✓` (U+2713) = passed, `✕` (U+2715) or `×` (U+00D7) = failed
   - Captures: (1) status symbol, (2) test name, (3) duration in ms. Map `\u2713` to `passed`, `\u2715`/`\u00D7` to `failed`.

4. **Match skipped tests** - regex: `/^\s+○\s+skipped\s+(.+)/`
   - Captured with `status: "skipped"`, `durationMs: 0`

5. **State machine:** Track "current spec file" when a PASS/FAIL line is matched. Subsequent test case lines are associated with it until the next PASS/FAIL line.

6. **Ignore lines:**
   - `timestamp=...` application log lines
   - `●` error detail blocks and stack traces
   - `ELIFECYCLE`, `ERR_PNPM` error lines
   - Jest summary lines (`Test Suites:`, `Tests:`, `Snapshots:`, `Time:`, `Ran all test suites`)
   - `A worker process has failed to exit gracefully` warnings

7. **Timeout detection:** If `jobConclusion` is `failure`, check which of the known top-level spec files did not get a PASS/FAIL line. Mark those as `TIMEOUT` with `durationSeconds: NaN`. (GitHub Actions step timeouts from `timeout-minutes` result in a `failure` conclusion, not a separate `timed_out` status at the job level.)

8. **Spec files with no duration suffix** (ran in <1s) - set `durationSeconds: 0`.

9. **Validation:** Parse the Jest summary line (`Test Suites: X failed, Y passed, Z total`) and cross-check against the number of parsed spec results. Log a warning if mismatched.

10. **Two-phase test run:** `test:local` runs the main suite first, then `test:liveness:local` separately. If the first command fails, `liveness.spec.ts` output is absent from the log. The parser handles this naturally - it only captures what's present. The absence of `liveness.spec.ts` on a failed run is expected (not a timeout) and should not be marked as `TIMEOUT`.

## Phase 3: Render

**Output:** Self-contained HTML file with all CSS inline (no external dependencies).

### Report structure

1. **Header**
   - Title: "E2E CI Runtime Report"
   - Date range and generation timestamp
   - Total runs, passed runs, failed runs

2. **Summary table** - one row per spec file, sorted by median duration descending
   - Columns: Spec File | Median (s) | Min (s) | Max (s) | Runs with data
   - Color-coded median cell: green (normal), yellow (>1 stddev), red (>2 stddev or has failures)

3. **Timeline table** - one row per spec file, one column per CI run (sorted by date)
   - Each cell shows duration in seconds
   - Cell color coding:
     - Green: within 1 stddev of that spec's median
     - Yellow: 1-2 stddev above median
     - Red: >2 stddev above median, or FAIL/TIMEOUT
     - Gray: no data (spec didn't run or skipped)
   - Column headers: date + commit SHA (truncated to 7 chars)
   - Failed/timed-out cells show "FAIL" or "TIMEOUT" text

4. **Detail sections** - one `<details>` element per spec file
   - When expanded, shows per-test-case timings for each run
   - Format: table with columns Run Date | Test Name | Duration (ms) | Status
   - Skipped tests shown in gray italic

### Color thresholds

Computed per spec file from all successful runs in the dataset:
- `median`: median duration of successful runs
- `stddev`: standard deviation of successful runs
- Green: `duration <= median + 1 * stddev`
- Yellow: `median + 1 * stddev < duration <= median + 2 * stddev`
- Red: `duration > median + 2 * stddev` or status is FAIL/TIMEOUT

## Workflow

**File:** `.github/workflows/e2e-runtime-report.yml`

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

## File Layout

```
e2e/
  scripts/
    generate-e2e-runtime-report.ts   # main script (fetch + parse + render)
  package.json                        # add @octokit/rest, tsx to devDependencies

.github/
  workflows/
    e2e-runtime-report.yml            # workflow definition
```

The script is a single file. If it exceeds ~400 lines during implementation, the three phases (fetch, parse, render) can be split into separate modules under `e2e/scripts/e2e-runtime-report/`.

## Dependencies

Added to `e2e/package.json` devDependencies:
- `@octokit/rest` - GitHub API client (pinned exact version)
- `tsx` - TypeScript execution (via `"catalog:"` to use workspace-pinned version)

Added to `e2e/package.json` scripts:
- `"generate-runtime-report": "tsx scripts/generate-e2e-runtime-report.ts"`

Output files are written to `process.cwd()` (the `e2e/` directory when invoked via `pnpm -F e2e run`).

## Secrets Required

- `GITHUB_TOKEN` - provided automatically by GitHub Actions
- `SLACK_BOT_TOKEN` - existing secret (used by `slack-notify-failed-jobs.yml`)
- `SLACK_ENGINEERING_ALERTS_CHANNEL_ID` - existing secret (used by `main.yml` notify job)

No new secrets need to be created.

## Out of Scope

- Cursor Automation trigger setup
- Interactive charts or graphs (future enhancement)
- Historical data persistence across runs (each run generates a fresh report)
- Alerting on regressions (future enhancement)
- `linea-besu-fleet.spec.ts` runtimes (never runs in `test:local`)
