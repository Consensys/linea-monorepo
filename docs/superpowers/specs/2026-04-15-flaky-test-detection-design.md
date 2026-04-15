# Flaky Test Detection for E2E Runtime Report

## Context

PR2700 introduces a monthly E2E CI runtime report that shows per-spec pass/fail/timeout status and
runtime trends. A reviewer requested additional data to spot flaky tests - specifically failure
counts and a ranking of the most frequently failing tests.

## Scope

Three changes to the existing report:

1. Add failure columns to the Summary table
2. Add a "Top Flaky Tests" leaderboard section to the HTML report
3. Add top 3 flaky tests to the Slack notification

All data needed already exists in memory (`SpecResult[]` and `TestResult[]`). No new API calls or
data fetching required.

## Design

### 1. Summary Table - Failure Columns

Add two columns to the existing per-spec Summary table:

| Existing columns | New columns |
|---|---|
| Spec File, Median (s), Min (s), Max (s), Runs | **Failures**, **Fail %** |

- `Failures` = count of runs where the spec had status FAIL or TIMEOUT
- `Total` = `Failures + Runs` (total appearances of the spec across all runs, where `Runs` is the
  existing count from `specStats.count` - successful runs with valid durations)
- `Fail %` = `Failures / Total`
- The `total run` summary row: `Failures` = `failedRuns`, `Fail %` = `failedRuns / totalRuns`.
  The existing `Runs` column on this row already shows `totalRunDurations.length` (runs with valid
  duration). For consistency, the `total run` row uses `totalRuns` as the denominator rather than
  `Runs`, since `failedRuns + passedRuns = totalRuns` by definition in the existing code
- Sort order remains by median duration descending (unchanged)

### 2. Top Flaky Tests Leaderboard (HTML)

A new section placed between the Summary table and the Timeline table, titled "Top Flaky Tests".

**Table columns:** Rank, Test Name, Spec File, Failures, Appearances, Fail %

**Logic:**

- Iterate over all `E2eRunResult[]`, collect every individual `TestResult` across all runs
- Group by `(specFile, testName)` - a test appearing in multiple runs gets one row
- `Appearances` = number of runs where the test appeared (status `passed` or `failed`; `skipped`
  excluded)
- `Failures` = number of appearances where status is `failed`
- Filter: only include tests with `Appearances >= 3` AND `Failures >= 1`
- Sort by Fail % descending, then by Failures descending as tiebreaker
- Slack top 3 uses this same sorted list (first 3 entries)
- Top 10 rows visible in the main table
- Remaining rows (if any) wrapped in `<details><summary>Show N more...</summary>...</details>`
- Row background color: red for >= 30% fail rate, yellow for >= 10%, no highlight otherwise

**Edge case:** If no tests meet the threshold, show "No flaky tests detected in this period."
instead of an empty table.

### 3. Slack Message - Top 3 Flaky Tests

Add a new block to the Slack message after the "E2E run duration" line.

**Format:**

```
Top flaky tests:
  1. should bridge ERC20 tokens (token-bridge) - 40.0% fail rate (6/15)
  2. should handle message anchoring (message-anchoring) - 20.0% fail rate (3/15)
  3. should finalize bundles (finalization) - 13.3% fail rate (2/15)
```

**Conditional:** Only included when at least one test meets the flaky threshold. If zero qualify,
the block is omitted entirely.

### Data Model Changes

Extend `ReportSummary` with a new field:

```typescript
interface FlakyTestEntry {
  testName: string;
  specFile: string;
  failures: number;
  appearances: number;
  failRate: number; // 0.00 to 1.00
}

interface ReportSummary {
  // ...existing fields...
  topFlakyTests: FlakyTestEntry[]; // top 3 from the same sorted list as the HTML leaderboard
}

// The full leaderboard (all qualifying entries) is computed in-memory for HTML rendering.
// Only the top 3 are serialized into the JSON summary for Slack consumption.
```

### Workflow YAML Changes

The Slack payload in `slack-notify-e2e-runtime-report.yml` reads `topFlakyTests` from the JSON
summary and conditionally renders a "Top flaky tests" rich_text block when the array is non-empty.

## Constants

| Name | Value | Rationale |
|---|---|---|
| MIN_APPEARANCES | 3 | Filters noise from recently added/removed tests |
| FLAKY_RED_THRESHOLD | 0.30 | 30%+ fail rate = red highlight |
| FLAKY_YELLOW_THRESHOLD | 0.10 | 10%+ fail rate = yellow highlight |
| LEADERBOARD_VISIBLE_ROWS | 10 | Top 10 visible, rest in expandable |
| SLACK_TOP_N | 3 | Top 3 in Slack message |

## Files Changed

| File | Change |
|---|---|
| `e2e/scripts/generate-e2e-runtime-report.ts` | Add flaky test computation, extend Summary table HTML, add leaderboard section HTML, extend `ReportSummary` type |
| `.github/workflows/slack-notify-e2e-runtime-report.yml` | Add conditional Slack block for top flaky tests |

## Known Limitations

- **Test identity:** `TestResult.name` captures the leaf test name from Jest output, not the full
  describe path. If duplicate test names exist under different describe blocks within a spec file,
  they would be grouped together. This is a pre-existing parser limitation, not introduced by this
  feature.
- **Retries:** `fetchE2eRuns` already uses `filter: "latest"` when listing jobs, so only the last
  retry attempt per workflow run is included. No additional deduplication needed.

## Testing

Existing unit tests in `e2e/scripts/helpers/parse-jest-log.test.ts` cover log parsing (unchanged).
The flaky test aggregation logic is pure computation over in-memory data - no new I/O or API calls.
Manual verification by running the report locally with `GITHUB_TOKEN=<token> pnpm -F e2e run generate-runtime-report -- --days 30` and inspecting the HTML output.
