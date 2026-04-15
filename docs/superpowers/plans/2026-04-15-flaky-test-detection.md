# Flaky Test Detection Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add failure counts and a flaky test leaderboard to the E2E CI runtime report so the team can spot flaky tests.

**Architecture:** Pure computation over in-memory `E2eRunResult[]` data - no new API calls. A new `computeFlakyTests()` function aggregates individual `TestResult` entries across runs. The summary table gets two new columns, a new HTML section renders the leaderboard, and the Slack workflow gets a conditional block for the top 3.

**Tech Stack:** TypeScript (tsx runner), Jest (unit tests), GitHub Actions YAML (Slack workflow)

---

## File Structure

| File | Responsibility |
|------|---------------|
| `e2e/scripts/generate-e2e-runtime-report.ts` | Types, flaky computation, HTML rendering, summary output |
| `.github/workflows/slack-notify-e2e-runtime-report.yml` | Slack message payload with conditional flaky tests block |
| `e2e/scripts/helpers/compute-flaky-tests.ts` | **New** - extracted `computeFlakyTests()` function and `FlakyTestEntry` type |
| `e2e/scripts/helpers/compute-flaky-tests.test.ts` | **New** - unit tests for flaky computation |

The flaky computation is extracted into its own helper (following the existing pattern of `parse-jest-log.ts` and `discover-timeout-eligible-spec-files.ts`) to keep `generate-e2e-runtime-report.ts` focused on orchestration/rendering and to make the logic independently testable.

---

### Task 1: Create `computeFlakyTests` helper with tests

**Files:**
- Create: `e2e/scripts/helpers/compute-flaky-tests.ts`
- Create: `e2e/scripts/helpers/compute-flaky-tests.test.ts`

- [ ] **Step 1: Write the failing tests**

Create `e2e/scripts/helpers/compute-flaky-tests.test.ts`:

```typescript
import { computeFlakyTests } from "./compute-flaky-tests";
import type { SpecResult } from "./parse-jest-log";

// Minimal run shape needed by computeFlakyTests
interface MinimalRunResult {
  specResults: SpecResult[];
}

describe("computeFlakyTests", () => {
  it("returns empty array when no tests have failures", () => {
    const runs: MinimalRunResult[] = [
      { specResults: [{ specFile: "src/a.spec.ts", status: "PASS", durationSeconds: 10, tests: [{ name: "test1", durationMs: 100, status: "passed" }] }] },
      { specResults: [{ specFile: "src/a.spec.ts", status: "PASS", durationSeconds: 10, tests: [{ name: "test1", durationMs: 100, status: "passed" }] }] },
      { specResults: [{ specFile: "src/a.spec.ts", status: "PASS", durationSeconds: 10, tests: [{ name: "test1", durationMs: 100, status: "passed" }] }] },
    ];
    expect(computeFlakyTests(runs)).toEqual([]);
  });

  it("excludes tests with fewer than MIN_APPEARANCES (3)", () => {
    const runs: MinimalRunResult[] = [
      { specResults: [{ specFile: "src/a.spec.ts", status: "FAIL", durationSeconds: 10, tests: [{ name: "test1", durationMs: 100, status: "failed" }] }] },
      { specResults: [{ specFile: "src/a.spec.ts", status: "FAIL", durationSeconds: 10, tests: [{ name: "test1", durationMs: 100, status: "failed" }] }] },
    ];
    expect(computeFlakyTests(runs)).toEqual([]);
  });

  it("excludes skipped test appearances from counts", () => {
    const runs: MinimalRunResult[] = [
      { specResults: [{ specFile: "src/a.spec.ts", status: "PASS", durationSeconds: 10, tests: [{ name: "test1", durationMs: 0, status: "skipped" }] }] },
      { specResults: [{ specFile: "src/a.spec.ts", status: "PASS", durationSeconds: 10, tests: [{ name: "test1", durationMs: 0, status: "skipped" }] }] },
      { specResults: [{ specFile: "src/a.spec.ts", status: "PASS", durationSeconds: 10, tests: [{ name: "test1", durationMs: 0, status: "skipped" }] }] },
    ];
    expect(computeFlakyTests(runs)).toEqual([]);
  });

  it("ranks tests by failure rate descending, then by failures descending", () => {
    const runs: MinimalRunResult[] = Array.from({ length: 5 }, () => ({
      specResults: [
        {
          specFile: "src/a.spec.ts", status: "PASS" as const, durationSeconds: 10,
          tests: [
            { name: "stable", durationMs: 100, status: "passed" as const },
            { name: "flaky", durationMs: 100, status: "passed" as const },
          ],
        },
        {
          specFile: "src/b.spec.ts", status: "PASS" as const, durationSeconds: 10,
          tests: [{ name: "very-flaky", durationMs: 100, status: "passed" as const }],
        },
      ],
    }));
    // Make "flaky" fail 2/5 = 40%
    runs[0].specResults[0].tests[1] = { name: "flaky", durationMs: 100, status: "failed" };
    runs[2].specResults[0].tests[1] = { name: "flaky", durationMs: 100, status: "failed" };
    // Make "very-flaky" fail 3/5 = 60%
    runs[0].specResults[1].tests[0] = { name: "very-flaky", durationMs: 100, status: "failed" };
    runs[1].specResults[1].tests[0] = { name: "very-flaky", durationMs: 100, status: "failed" };
    runs[3].specResults[1].tests[0] = { name: "very-flaky", durationMs: 100, status: "failed" };

    const result = computeFlakyTests(runs);

    expect(result).toHaveLength(2);
    expect(result[0]).toEqual({
      testName: "very-flaky",
      specFile: "src/b.spec.ts",
      failures: 3,
      appearances: 5,
      failRate: 0.60,
    });
    expect(result[1]).toEqual({
      testName: "flaky",
      specFile: "src/a.spec.ts",
      failures: 2,
      appearances: 5,
      failRate: 0.40,
    });
  });

  it("groups by (specFile, testName) so same test name in different specs is separate", () => {
    const runs: MinimalRunResult[] = Array.from({ length: 3 }, () => ({
      specResults: [
        { specFile: "src/a.spec.ts", status: "FAIL" as const, durationSeconds: 10, tests: [{ name: "shared-name", durationMs: 100, status: "failed" as const }] },
        { specFile: "src/b.spec.ts", status: "PASS" as const, durationSeconds: 10, tests: [{ name: "shared-name", durationMs: 100, status: "passed" as const }] },
      ],
    }));

    const result = computeFlakyTests(runs);

    expect(result).toHaveLength(1);
    expect(result[0].specFile).toBe("src/a.spec.ts");
    expect(result[0].failures).toBe(3);
  });
});
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `pnpm -F e2e run test:script`
Expected: FAIL - cannot find module `./compute-flaky-tests`

- [ ] **Step 3: Implement `computeFlakyTests`**

Create `e2e/scripts/helpers/compute-flaky-tests.ts`:

```typescript
import type { SpecResult } from "./parse-jest-log";

export interface FlakyTestEntry {
  testName: string;
  specFile: string;
  failures: number;
  appearances: number;
  failRate: number; // 0.00 to 1.00
}

const MIN_APPEARANCES = 3;

export function computeFlakyTests(runResults: { specResults: SpecResult[] }[]): FlakyTestEntry[] {
  const counts = new Map<string, { testName: string; specFile: string; failures: number; appearances: number }>();

  for (const run of runResults) {
    for (const spec of run.specResults) {
      for (const test of spec.tests) {
        if (test.status === "skipped") continue;

        const key = `${spec.specFile}\0${test.name}`;
        let entry = counts.get(key);
        if (!entry) {
          entry = { testName: test.name, specFile: spec.specFile, failures: 0, appearances: 0 };
          counts.set(key, entry);
        }
        entry.appearances++;
        if (test.status === "failed") entry.failures++;
      }
    }
  }

  return [...counts.values()]
    .filter((e) => e.appearances >= MIN_APPEARANCES && e.failures >= 1)
    .map((e) => ({
      testName: e.testName,
      specFile: e.specFile,
      failures: e.failures,
      appearances: e.appearances,
      failRate: parseFloat((e.failures / e.appearances).toFixed(2)),
    }))
    .sort((a, b) => b.failRate - a.failRate || b.failures - a.failures);
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `pnpm -F e2e run test:script`
Expected: All tests PASS (both existing parse-jest-log tests and new compute-flaky-tests tests)

- [ ] **Step 5: Commit**

```bash
git add e2e/scripts/helpers/compute-flaky-tests.ts e2e/scripts/helpers/compute-flaky-tests.test.ts
git commit -m "feat(e2e): add computeFlakyTests helper with tests"
```

---

### Task 2: Extend `ReportSummary` and wire flaky data into summary output

**Files:**
- Modify: `e2e/scripts/generate-e2e-runtime-report.ts:24` (import)
- Modify: `e2e/scripts/generate-e2e-runtime-report.ts:41-48` (ReportSummary interface)
- Modify: `e2e/scripts/generate-e2e-runtime-report.ts:182` (renderHtmlReport - compute + pass to summary)
- Modify: `e2e/scripts/generate-e2e-runtime-report.ts:446-453` (empty summary sentinel)

- [ ] **Step 1: Add import for `computeFlakyTests` and `FlakyTestEntry`**

In `e2e/scripts/generate-e2e-runtime-report.ts`, after the existing imports (line 24), add:

```typescript
import { computeFlakyTests, type FlakyTestEntry } from "./helpers/compute-flaky-tests";
```

- [ ] **Step 2: Extend `ReportSummary` interface**

Add `topFlakyTests` field to the `ReportSummary` interface (line 41-48):

```typescript
interface ReportSummary {
  totalRuns: number;
  passedRuns: number;
  failedRuns: number;
  dateRange: string;
  medianRunDurationSeconds: number;
  maxRunDurationSeconds: number;
  topFlakyTests: FlakyTestEntry[];
}
```

- [ ] **Step 3: Compute flaky tests and add to summary in `renderHtmlReport`**

Inside `renderHtmlReport()`, compute `allFlakyTests` **before** the `const summary = { ... }` object
literal (around line 224). Insert before it:

```typescript
const allFlakyTests = computeFlakyTests(runResults);
```

Then include `topFlakyTests` in the `summary` literal:

```typescript
const summary: ReportSummary = {
  totalRuns,
  passedRuns,
  failedRuns,
  dateRange,
  medianRunDurationSeconds: Math.round(isNaN(medianRunDuration) ? 0 : medianRunDuration),
  maxRunDurationSeconds: Math.round(isNaN(maxRunDuration) ? 0 : maxRunDuration),
  topFlakyTests: allFlakyTests.slice(0, 3),
};
```

`allFlakyTests` stays in scope for Task 4's leaderboard HTML rendering later in the same function.

- [ ] **Step 4: Update empty summary sentinel**

In `main()` (around line 446-453), add `topFlakyTests: []` to the empty summary:

```typescript
const emptySummary: ReportSummary = {
  totalRuns: 0,
  passedRuns: 0,
  failedRuns: 0,
  dateRange: "N/A",
  medianRunDurationSeconds: 0,
  maxRunDurationSeconds: 0,
  topFlakyTests: [],
};
```

- [ ] **Step 5: Run tests to verify nothing broke**

Run: `pnpm -F e2e run test:script`
Expected: All tests PASS

- [ ] **Step 6: Commit**

```bash
git add e2e/scripts/generate-e2e-runtime-report.ts
git commit -m "feat(e2e): extend ReportSummary with topFlakyTests field"
```

---

### Task 3: Add Failures and Fail % columns to the Summary table

**Files:**
- Modify: `e2e/scripts/generate-e2e-runtime-report.ts:186-200` (specStats computation - add failureCount)
- Modify: `e2e/scripts/generate-e2e-runtime-report.ts:305-315` (totalRunSummaryRow HTML)
- Modify: `e2e/scripts/generate-e2e-runtime-report.ts:317-334` (summaryRows HTML)
- Modify: `e2e/scripts/generate-e2e-runtime-report.ts:399-406` (summary table thead)

- [ ] **Step 1: Add `failureCount` to `specStats`**

Extend the `specStats` map value type to include `failureCount`. After the existing `specStats.set()` call (line 193-199), compute failures for each spec:

```typescript
const specStats = new Map<string, { med: number; sd: number; min: number; max: number; count: number; failureCount: number }>();
for (const specFile of allSpecFiles) {
  const durations = runResults
    .flatMap((r) => r.specResults)
    .filter((s) => s.specFile === specFile && s.status === "PASS" && !isNaN(s.durationSeconds))
    .map((s) => s.durationSeconds);

  const failureCount = runResults.filter((r) =>
    r.specResults.some((s) => s.specFile === specFile && (s.status === "FAIL" || s.status === "TIMEOUT")),
  ).length;

  specStats.set(specFile, {
    med: median(durations),
    sd: stddev(durations),
    min: durations.length > 0 ? Math.min(...durations) : NaN,
    max: durations.length > 0 ? Math.max(...durations) : NaN,
    count: durations.length,
    failureCount,
  });
}
```

- [ ] **Step 2: Update summary table `<thead>`**

In the HTML template (line 401), change:

```html
<thead><tr><th>Spec File</th><th>Median (s)</th><th>Min (s)</th><th>Max (s)</th><th>Runs</th></tr></thead>
```

To:

```html
<thead><tr><th>Spec File</th><th>Median (s)</th><th>Min (s)</th><th>Max (s)</th><th>Runs</th><th>Failures</th><th>Fail %</th></tr></thead>
```

- [ ] **Step 3: Update `totalRunSummaryRow` with failure columns**

In the `totalRunSummaryRow` template (lines 309-315), add two cells after the `Runs` cell:

```typescript
const totalRunFailRate = totalRuns > 0 ? ((failedRuns / totalRuns) * 100).toFixed(1) : "0.0";
const totalRunSummaryRow = `<tr style="border-bottom:2px solid #aaa;">
        <td class="spec-name" style="font-weight:600;">total run</td>
        <td style="background:${COLOR_MAP.green};font-weight:600;">${isNaN(medianRunDuration) ? "-" : (medianRunDuration / 60).toFixed(1) + "m"}</td>
        <td>${isNaN(totalRunMin) ? "-" : (totalRunMin / 60).toFixed(1) + "m"}</td>
        <td>${isNaN(totalRunMax) ? "-" : (totalRunMax / 60).toFixed(1) + "m"}</td>
        <td>${totalRunDurations.length}</td>
        <td>${failedRuns}</td>
        <td>${totalRunFailRate}%</td>
      </tr>`;
```

- [ ] **Step 4: Update per-spec `summaryRows` with failure columns**

In the `summaryRows` template (lines 318-334), add failure cells. Replace the existing `hasFailures` boolean with the `failureCount` from specStats:

```typescript
const summaryRows = sortedSpecs
  .map((specFile) => {
    const stats = specStats.get(specFile)!;
    const shortName = specFile.replace("src/", "").replace(".spec.ts", "");
    const medColor: ColorLevel = isNaN(stats.med) ? "gray" : stats.failureCount > 0 ? "red" : "green";
    const total = stats.count + stats.failureCount;
    const failRate = total > 0 ? ((stats.failureCount / total) * 100).toFixed(1) : "0.0";
    return `<tr>
          <td class="spec-name">${shortName}</td>
          <td style="background:${COLOR_MAP[medColor]}">${isNaN(stats.med) ? "-" : stats.med.toFixed(1)}</td>
          <td>${isNaN(stats.min) ? "-" : stats.min.toFixed(1)}</td>
          <td>${isNaN(stats.max) ? "-" : stats.max.toFixed(1)}</td>
          <td>${stats.count}</td>
          <td>${stats.failureCount}</td>
          <td>${failRate}%</td>
        </tr>`;
  })
  .join("\n          ");
```

- [ ] **Step 5: Run tests and verify**

Run: `pnpm -F e2e run test:script`
Expected: All tests PASS

- [ ] **Step 6: Commit**

```bash
git add e2e/scripts/generate-e2e-runtime-report.ts
git commit -m "feat(e2e): add Failures and Fail % columns to summary table"
```

---

### Task 4: Add Top Flaky Tests leaderboard HTML section

**Files:**
- Modify: `e2e/scripts/generate-e2e-runtime-report.ts` (HTML template, between Summary and Timeline)

- [ ] **Step 1: Build the leaderboard HTML**

After the `summaryRows` computation and before the Timeline section in `renderHtmlReport()`, add the leaderboard section. Insert this code block that uses `allFlakyTests` (computed in Task 2):

```typescript
const LEADERBOARD_VISIBLE_ROWS = 10;
const FLAKY_RED_THRESHOLD = 0.30;
const FLAKY_YELLOW_THRESHOLD = 0.10;

let flakySection: string;
if (allFlakyTests.length === 0) {
  flakySection = `<p>No flaky tests detected in this period.</p>`;
} else {
  const flakyRowHtml = (entry: FlakyTestEntry, rank: number) => {
    const bg = entry.failRate >= FLAKY_RED_THRESHOLD
      ? `background:${COLOR_MAP.red}`
      : entry.failRate >= FLAKY_YELLOW_THRESHOLD
        ? `background:${COLOR_MAP.yellow}`
        : "";
    const shortSpec = entry.specFile.replace("src/", "").replace(".spec.ts", "");
    return `<tr style="${bg}">
            <td>${rank}</td>
            <td style="text-align:left">${entry.testName}</td>
            <td class="spec-name">${shortSpec}</td>
            <td>${entry.failures}</td>
            <td>${entry.appearances}</td>
            <td>${(entry.failRate * 100).toFixed(1)}%</td>
          </tr>`;
  };

  const visibleRows = allFlakyTests.slice(0, LEADERBOARD_VISIBLE_ROWS).map((e, i) => flakyRowHtml(e, i + 1)).join("\n          ");
  const hiddenEntries = allFlakyTests.slice(LEADERBOARD_VISIBLE_ROWS);
  const hiddenBlock = hiddenEntries.length > 0
    ? `<details>
        <summary>Show ${hiddenEntries.length} more...</summary>
        <table>
          <thead><tr><th>Rank</th><th>Test Name</th><th>Spec File</th><th>Failures</th><th>Appearances</th><th>Fail %</th></tr></thead>
          <tbody>
            ${hiddenEntries.map((e, i) => flakyRowHtml(e, LEADERBOARD_VISIBLE_ROWS + i + 1)).join("\n            ")}
          </tbody>
        </table>
      </details>`
    : "";

  flakySection = `<table>
      <thead><tr><th>Rank</th><th>Test Name</th><th>Spec File</th><th>Failures</th><th>Appearances</th><th>Fail %</th></tr></thead>
      <tbody>
        ${visibleRows}
      </tbody>
    </table>
    ${hiddenBlock}`;
}
```

- [ ] **Step 2: Insert the section in the HTML template**

In the HTML template string, between the Summary `</table>` and `<h2>Timeline</h2>`, insert:

```html
  <h2>Top Flaky Tests</h2>
  ${flakySection}
```

- [ ] **Step 3: Run tests and verify**

Run: `pnpm -F e2e run test:script`
Expected: All tests PASS

- [ ] **Step 4: Commit**

```bash
git add e2e/scripts/generate-e2e-runtime-report.ts
git commit -m "feat(e2e): add Top Flaky Tests leaderboard to HTML report"
```

---

### Task 5: Add top 3 flaky tests to Slack workflow

**Files:**
- Modify: `.github/workflows/slack-notify-e2e-runtime-report.yml:37-110` (add formatting step + Slack payload)

The Slack YAML needs care: inline JSON substitution via `${{ }}` inside `node -e "..."` will break
on unescaped quotes in the expanded JSON, and multiline `GITHUB_OUTPUT` values will corrupt the
`payload: |` YAML block when substituted. Solution: pass JSON via environment variable, parse with
`JSON.parse()`, and output the formatted text as a **single line** with literal `\n` characters
(Slack Block Kit interprets `\n` in `rich_text` text elements as line breaks). This keeps the
`${{ }}` substitution safe inside the YAML scalar.

- [ ] **Step 1: Add a "Format flaky tests" step**

After the "Read summary" step (line 43) and before "Upload HTML report", add a step that formats
the top flaky tests into a single-line Slack-safe string. Use `env:` to pass the JSON safely:

```yaml
      - name: Format flaky tests for Slack
        id: flaky
        env:
          SUMMARY_JSON: ${{ steps.summary.outputs.json }}
        run: |
          node -e '
            const summary = JSON.parse(process.env.SUMMARY_JSON);
            const tests = summary.topFlakyTests || [];
            if (tests.length === 0) {
              process.stdout.write("has_flaky=false\n");
            } else {
              const line = tests.map((t, i) => {
                const spec = t.specFile.replace("src/", "").replace(".spec.ts", "");
                return (i+1) + ". " + t.testName + " (" + spec + ") - " + (t.failRate*100).toFixed(1) + "% fail rate (" + t.failures + "/" + t.appearances + ")";
              }).join("\\n");
              process.stdout.write("has_flaky=true\ntext=" + line + "\n");
            }
          ' >> "$GITHUB_OUTPUT"
```

Key: `.join("\\n")` produces a single-line output with literal `\n` in the string.
`text=` assignment (single line) avoids the heredoc delimiter approach entirely.

- [ ] **Step 2: Add conditional Slack block for flaky tests**

In the Slack `blocks:` array, after the "E2E run duration" block (ending at line 98) and before
the "Full report" block (line 100), add a `rich_text` block. The block is always present to avoid
complex YAML conditional logic mid-array. When no flaky tests qualify, it shows
"None detected this period." which is harmless and lets readers know the feature exists:

```yaml
              - type: rich_text
                elements:
                  - type: rich_text_section
                    elements:
                      - type: text
                        text: "Top flaky tests:\n"
                        style:
                          bold: true
                      - type: text
                        text: "${{ steps.flaky.outputs.has_flaky == 'true' && steps.flaky.outputs.text || 'None detected this period.' }}"
```

Since `steps.flaky.outputs.text` is a single line with literal `\n`, the `${{ }}` substitution
is safe inside the YAML double-quoted string. Slack interprets `\n` as line breaks in
`rich_text` text elements.

**Note on special characters:** If test names ever contain `"` or `\`, they could break the YAML
double-quoted string. In practice, E2E test names in this repo are plain English descriptions
without special characters. If this becomes an issue in the future, the formatter step should
escape `"` -> `\"` and `\` -> `\\` before outputting.

- [ ] **Step 3: Commit**

```bash
git add .github/workflows/slack-notify-e2e-runtime-report.yml
git commit -m "feat(e2e): add top 3 flaky tests to Slack notification"
```

---

### Task 6: Manual verification

- [ ] **Step 1: Run the full test suite**

Run: `pnpm -F e2e run test:script`
Expected: All tests PASS

- [ ] **Step 2: Generate the report locally (if you have a GITHUB_TOKEN)**

Run: `GITHUB_TOKEN=<token> pnpm -F e2e run generate-runtime-report -- --days 30`
Expected: `e2e-runtime-report.html` generated, open in browser and verify:
- Summary table has "Failures" and "Fail %" columns
- "Top Flaky Tests" section appears between Summary and Timeline
- `e2e-runtime-report-summary.json` contains `topFlakyTests` array

- [ ] **Step 3: Verify the Slack YAML is valid**

Visually inspect `.github/workflows/slack-notify-e2e-runtime-report.yml` for correct YAML indentation and block structure. No automated validation available for Slack Block Kit payloads locally.
