# AI-Powered E2E Failure Analysis - Design Spec

## Overview

Add AI-powered root cause analysis to E2E test failures in CI. When the main
workflow fails on push to `main`, Claude analyzes the failed test logs (and
Docker container logs when available), produces a structured failure analysis
markdown artifact, and enriches the existing Slack notification with a
truncated summary and link to the full analysis.

## Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Trigger | Same as `notify` job in `main.yml` - on `main`, any job failure | Couples naturally with existing Slack notification |
| Log sources | Jest logs always, Docker logs when available | Graceful degradation; `e2e-tests-logs-dump` is opt-in |
| Source access | Checkout repo for test source context | Enables correlation of stack traces with actual assertions |
| Output | GitHub Actions artifact (markdown) | Downloadable, archival, no issue noise |
| Architecture | Two jobs in `slack-notify-failed-jobs.yml` | Clean separation; `notify` always runs via `if: always()` |
| Fail-safety | `ai-analysis` job can fail without blocking `notify` | Structural isolation via separate job + `if: always()` |
| Model | Configurable, default Sonnet | Cost control on per-push trigger; callers can override |
| Secret passing | Explicit `anthropic_api_key` secret input | Matches existing explicit secret pattern in `main.yml` |

## Architecture

### Workflow Structure

Two jobs within `slack-notify-failed-jobs.yml`:

```
Job 1: ai-analysis (conditional on enable_ai_analysis, can fail)
  - Checkout repo
  - Fetch failed job logs via gh CLI
  - Download Docker logs artifact (best-effort)
  - Run claude-code-action with direct_prompt
  - Upload analysis artifact
  - Output: summary text, has_analysis flag

Job 2: notify (always runs)
  - needs: [ai-analysis]
  - if: always()
  - Existing Slack logic (unchanged)
  - + AI summary block and artifact link (when available)
```

The `notify` job waits for `ai-analysis` to complete (1-3 minutes typical for
Claude). This delay is accepted in exchange for the enriched Slack message
containing the AI summary inline.

### New Workflow Inputs

Added to `slack-notify-failed-jobs.yml`:

```yaml
inputs:
  enable_ai_analysis:
    description: "Run AI-powered failure analysis using Claude"
    required: false
    type: boolean
    default: false
  ai_analysis_model:
    description: "Claude model for AI analysis"
    required: false
    type: string
    default: "claude-sonnet-4-20250514"
  docker_logs_artifact_name:
    description: "Name of the artifact containing Docker container logs"
    required: false
    type: string
    default: "end-2-end-debug-logs"
secrets:
  anthropic_api_key:
    description: "Anthropic API key for Claude analysis"
    required: false
```

Workflow-level permissions are unchanged (`contents: read`, `actions: read`).
The `id-token: write` permission (required by `claude-code-action`) is scoped
to the `ai-analysis` job only - not the workflow level. This prevents
permission mismatches for the 7 other callers that do not grant `id-token`.

### AI Analysis Job

The `ai-analysis` job uses a **job-level** conditional:

```yaml
ai-analysis:
  if: ${{ inputs.enable_ai_analysis }}
  runs-on: ubuntu-latest
  timeout-minutes: 10
  permissions:
    contents: read
    actions: read
    id-token: write
```

When `enable_ai_analysis` is false (the default), the job is skipped entirely -
no runner is allocated, no permissions are evaluated. This keeps existing
callers completely unaffected.

**Steps:**

1. **Checkout repository** - `actions/checkout@v4`, `fetch-depth: 1`
2. **Fetch failed job logs** - `gh run view ${{ inputs.run_id }} --log-failed`
   saved to `/tmp/failed-job-logs.txt`. Requires `env: GH_TOKEN: ${{ secrets.GITHUB_TOKEN }}`.
   Truncate to last 500KB if oversized (`tail -c 500000`) to avoid exceeding
   Claude's context budget on multi-job failures.
3. **Download Docker logs artifact** - `actions/download-artifact@v4` with
   configurable artifact name to `/tmp/docker-logs/`, `continue-on-error: true`
   (artifact may not exist)
4. **Run Claude AI analysis** - `anthropics/claude-code-action` with
   `direct_prompt` and `claude_args: --model ${{ inputs.ai_analysis_model }}`.
   Writes `/tmp/e2e-failure-analysis.md` and `/tmp/e2e-failure-summary.txt`.
5. **Upload analysis artifact** - `actions/upload-artifact@v4`, artifact name
   `ai-failure-analysis`
6. **Extract summary** - Shell step reads summary file, sanitizes for YAML
   safety (strip newlines, escape quotes via `jq -Rs`), truncates to 300 chars,
   and sets job outputs (`summary`, `has_analysis`)

### Claude Prompt

Uses `anthropics/claude-code-action` with `direct_prompt`. The following is
the verbatim prompt text passed to Claude:

````
You are analyzing a CI failure in the Linea monorepo's E2E test suite.

## Context
- Repository: Linea zkEVM monorepo (L2 zero-knowledge rollup)
- E2E tests: Jest + Viem, run against a local Docker stack (coordinator,
  prover, sequencer, shomei, postman, L1/L2 nodes)
- Tests cover: bridge tokens, messaging, submission/finalization, transaction
  exclusion, EIP-7702, opcodes, liveness, restarts

## Input Files
- `/tmp/failed-job-logs.txt` - Jest stdout/stderr from the failed CI run
- `/tmp/docker-logs/` - Docker container logs (may not exist)

## Instructions

1. Read `/tmp/failed-job-logs.txt` to identify which tests failed and how.
2. If `/tmp/docker-logs/` exists, read relevant container logs to find
   infrastructure-level root causes (coordinator crashes, prover failures,
   node sync issues, OOM, etc.).
3. For each failing test, read the test source file under `e2e/src/` to
   understand what the test was asserting.
4. Also read relevant helpers/utilities referenced in stack traces.

Parse logs for these failure patterns:

| Pattern Type | Indicators |
|---|---|
| Test failures | `FAIL`, `AssertionError`, `Expected`, stack traces |
| Timeouts | `Exceeded timeout`, jest timeout errors |
| Transaction reverts | `reverted`, `execution reverted`, revert reason strings |
| Build/setup errors | `error:`, `Error:`, `Cannot find module`, `ECONNREFUSED` |
| Infrastructure | OOM, container exit codes, connection refused in Docker logs |

Extract from errors:
- File paths and line numbers (e.g., `e2e/src/file.spec.ts:42:10`)
- Error messages and assertion details
- Stack traces (full chain from symptom to origin)

For each file path extracted, read the actual source file for context.

## Analysis Rules

- Analyze ALL failed tests before producing output.
- For test failures, prefer diagnosing the code/infrastructure under test
  over blaming the test itself, unless the test is clearly wrong.
- If a failure cannot be diagnosed, say "Manual investigation needed" with
  whatever partial context you have. Do not guess.
- Include the actual error snippet (max 10 lines) in each failure entry.

## Failure Types and Fix Strategies

| Type | Strategy |
|---|---|
| Assertion mismatch | Identify expected vs actual; check if test expectation is stale or if code regressed |
| Timeout | Check if underlying service (coordinator, prover, sequencer) is healthy in Docker logs; look for deadlocks or slow operations |
| Transaction revert | Decode revert reason; check contract state prerequisites and test setup |
| Setup/teardown failure | Check global-setup.ts, Docker stack health, port conflicts |
| Infrastructure crash | Identify which container failed and why from Docker logs |

## Confidence Levels
- **High**: Clear error message pointing to specific root cause
- **Medium**: Error is clear but root cause requires inference
- **Low**: Error is ambiguous or root cause is uncertain

## Output

Write TWO files:

### /tmp/e2e-failure-analysis.md
Full structured analysis:

```
# E2E Failure Analysis
**Run:** <run_id>
**Date:** <timestamp>

## Summary
<2-3 sentence executive summary>

## Failures

### 1. <Test Name>
- **File:** `e2e/src/<file>:<line>`
- **Type:** <Assertion mismatch | Timeout | Transaction revert | Setup failure | Infrastructure>
- **Error:**
  <error snippet, max 10 lines>
- **Root Cause:** <analysis>
- **Suggested Fix:** <concrete action> (Confidence: High/Medium/Low)
- **Category:** Flake | Regression | Infrastructure

### 2. <Test Name>
...

## Infrastructure Notes
<Container-level issues from Docker logs, or "No issues detected">
```

### /tmp/e2e-failure-summary.txt
A single line, max 300 characters, summarizing the failures for a Slack
notification. Example:
"2 tests failed: submission-finalization timed out (likely coordinator OOM -
check memory limits), bridge-tokens assertion mismatch on L1->L2 claim
(regression in fee calculation)"
````

### Slack Integration

The `notify` job checks `needs.ai-analysis.outputs.has_analysis`. When true,
appends these blocks after the existing failed/cancelled job sections:

```yaml
- type: divider

- type: rich_text
  elements:
    - type: rich_text_section
      elements:
        - type: text
          text: "AI Analysis: "
          style:
            bold: true
        - type: text
          text: "<summary from ai-analysis job output>"

- type: rich_text
  elements:
    - type: rich_text_section
      elements:
        - type: text
          text: "Full Analysis: "
          style:
            bold: true
        - type: link
          url: "https://github.com/<repo>/actions/runs/<run_id>"
          text: "View artifact"
```

When AI analysis failed or was skipped, these blocks are absent. The Slack
message is identical to today's.

### `main.yml` Changes

**Permissions:** Add `id-token: write` to the top-level permissions block.
Required because a reusable workflow's effective permissions cannot exceed the
caller's. Without this, `claude-code-action` in the called workflow would fail
with a permissions error.

```yaml
permissions:
  contents: read
  actions: read
  security-events: write
  packages: write
  id-token: write  # NEW - required for claude-code-action in notify workflow
```

**Notify job call:** Three lines added:

```yaml
notify:
  needs: [ ... ]  # unchanged
  if: ${{ always() && github.ref == 'refs/heads/main' && contains(needs.*.result, 'failure') }}
  uses: ./.github/workflows/slack-notify-failed-jobs.yml
  with:
    title: "Main workflow Failed"
    run_id: ${{ github.run_id }}
    enable_ai_analysis: true                    # NEW
    ai_analysis_model: "claude-sonnet-4-20250514"  # NEW
  secrets:
    channel_id: ${{ secrets.SLACK_ENGINEERING_ALERTS_CHANNEL_ID }}
    slack_bot_token: ${{ secrets.SLACK_GITHUB_ACTIONS_ALERTS_BOT_TOKEN }}
    anthropic_api_key: ${{ secrets.ANTHROPIC_API_KEY }}  # NEW
```

All other callers of `slack-notify-failed-jobs.yml` remain untouched.

## Files Changed

| File | Change |
|------|--------|
| `.github/workflows/slack-notify-failed-jobs.yml` | Add inputs, `ai-analysis` job (with job-level `id-token: write`), enrich `notify` Slack payload |
| `.github/workflows/main.yml` | Add `id-token: write` to top-level permissions; pass `enable_ai_analysis`, `ai_analysis_model`, `anthropic_api_key` to notify |

## Dependencies

| Dependency | Version | Already in repo |
|------------|---------|-----------------|
| `anthropics/claude-code-action` | `@5d0cc745cd0cce4c0e9e0b3511de26c3bc285eb5` (#v1.0.71) | Yes |
| `actions/checkout` | `@v4` | Yes |
| `actions/download-artifact` | `@v4` (v7 in some workflows) | Yes |
| `actions/upload-artifact` | `@v4` | Yes |

No new dependencies.

## Risks and Mitigations

| Risk | Mitigation |
|------|------------|
| AI analysis blocks Slack notification | Structural: separate job, `notify` uses `if: always()` |
| Claude API failure or timeout | `ai-analysis` job failure doesn't propagate to `notify` |
| Large logs exceed Claude context window | Log file truncated to last 500KB before Claude reads; Claude reads source files selectively |
| Cost of Sonnet on every main push failure | Sonnet default is cost-effective; configurable to cheaper model |
| Anthropic API key not set in repo secrets | `ai-analysis` job is conditional on `enable_ai_analysis`; will fail gracefully |
| Docker logs artifact doesn't exist | `continue-on-error: true` on download step; analysis proceeds with Jest logs only |
| AI summary text breaks Slack payload YAML | Extract step sanitizes via `jq -Rs` (JSON-safe string); Slack block built with `jq` matching existing `build_section` pattern |
| `ai-analysis` job delays Slack notification | Accepted trade-off; `timeout-minutes: 10` caps worst case |
