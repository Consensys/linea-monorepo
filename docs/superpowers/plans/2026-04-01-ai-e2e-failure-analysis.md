# AI-Powered E2E Failure Analysis - Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add opt-in AI failure analysis to the Slack notification workflow, producing a Claude-generated root cause analysis artifact and enriching the Slack message with a summary.

**Architecture:** Two-job design in `slack-notify-failed-jobs.yml` - an `ai-analysis` job (conditional, can fail) runs Claude analysis, then the `notify` job (always runs) sends the existing Slack notification enriched with AI results when available. `main.yml` passes new inputs to enable AI analysis.

**Tech Stack:** GitHub Actions, `anthropics/claude-code-action@v1.0.71`, Slack Block Kit, `jq`, `gh` CLI

**Spec:** `docs/superpowers/specs/2026-04-01-ai-e2e-failure-analysis-design.md`

---

### Task 1: Add new inputs and secrets to `slack-notify-failed-jobs.yml`

**Files:**
- Modify: `.github/workflows/slack-notify-failed-jobs.yml:7-29`

- [ ] **Step 1: Add the three new inputs after `extra_message`**

Insert after line 22 (after the `extra_message` input block):

```yaml
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
```

- [ ] **Step 2: Add the `anthropic_api_key` secret after `slack_bot_token`**

Insert after line 29 (after the `slack_bot_token` secret):

```yaml
      anthropic_api_key:
        description: "Anthropic API key for Claude analysis"
        required: false
```

- [ ] **Step 3: Validate YAML syntax**

Run: `python3 -c "import yaml; yaml.safe_load(open('.github/workflows/slack-notify-failed-jobs.yml'))"`
Expected: No output (valid YAML)

- [ ] **Step 4: Commit**

```bash
git add .github/workflows/slack-notify-failed-jobs.yml
git commit -m "feat: add AI analysis inputs to slack-notify-failed-jobs"
```

---

### Task 2: Add the `ai-analysis` job to `slack-notify-failed-jobs.yml`

**Files:**
- Modify: `.github/workflows/slack-notify-failed-jobs.yml`

The new job goes between the `jobs:` key and the existing `notify:` job.

- [ ] **Step 1: Add the `ai-analysis` job definition**

Insert after line 31 (`jobs:`), before the existing `notify:` job. The full job:

```yaml
  ai-analysis:
    if: ${{ inputs.enable_ai_analysis }}
    runs-on: ubuntu-latest
    timeout-minutes: 10
    permissions:
      contents: read
      actions: read
      id-token: write
    outputs:
      summary: ${{ steps.extract_summary.outputs.summary }}
      has_analysis: ${{ steps.extract_summary.outputs.has_analysis }}

    steps:
      - name: Checkout repository
        uses: actions/checkout@v4
        with:
          fetch-depth: 1

      - name: Fetch failed job logs
        env:
          GH_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        run: |
          set -euo pipefail
          gh run view ${{ inputs.run_id }} --log-failed > /tmp/failed-job-logs-raw.txt 2>&1 || true
          tail -c 500000 /tmp/failed-job-logs-raw.txt > /tmp/failed-job-logs.txt

      - name: Download Docker logs artifact
        uses: actions/download-artifact@v4
        continue-on-error: true
        with:
          name: ${{ inputs.docker_logs_artifact_name }}
          path: /tmp/docker-logs

      - name: Run Claude AI analysis
        uses: anthropics/claude-code-action@5d0cc745cd0cce4c0e9e0b3511de26c3bc285eb5 #v1.0.71
        with:
          anthropic_api_key: ${{ secrets.anthropic_api_key }}
          claude_args: --model ${{ inputs.ai_analysis_model }}
          direct_prompt: |
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

            # E2E Failure Analysis
            **Run:** ${{ inputs.run_id }}
            **Date:** <current UTC timestamp>

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

            ### /tmp/e2e-failure-summary.txt
            A single line, max 300 characters, summarizing the failures for a Slack
            notification. Example:
            "2 tests failed: submission-finalization timed out (likely coordinator OOM -
            check memory limits), bridge-tokens assertion mismatch on L1->L2 claim
            (regression in fee calculation)"

      - name: Upload analysis artifact
        if: ${{ always() }}
        uses: actions/upload-artifact@v4
        continue-on-error: true
        with:
          name: ai-failure-analysis
          path: /tmp/e2e-failure-analysis.md
          if-no-files-found: ignore

      - name: Extract summary for Slack
        id: extract_summary
        if: ${{ always() }}
        shell: bash
        run: |
          set -euo pipefail
          if [[ -f /tmp/e2e-failure-summary.txt ]]; then
            # Use jq -Rs for JSON-safe escaping (handles quotes, $, backticks, newlines)
            summary=$(head -c 300 /tmp/e2e-failure-summary.txt | jq -Rs .)
            echo "summary=${summary}" >> "$GITHUB_OUTPUT"
            echo "has_analysis=true" >> "$GITHUB_OUTPUT"
          else
            echo 'summary=""' >> "$GITHUB_OUTPUT"
            echo "has_analysis=false" >> "$GITHUB_OUTPUT"
          fi
```

- [ ] **Step 2: Validate YAML syntax**

Run: `python3 -c "import yaml; yaml.safe_load(open('.github/workflows/slack-notify-failed-jobs.yml'))"`
Expected: No output (valid YAML)

- [ ] **Step 3: Commit**

```bash
git add .github/workflows/slack-notify-failed-jobs.yml
git commit -m "feat: add ai-analysis job to slack-notify-failed-jobs"
```

---

### Task 3: Modify the `notify` job to consume AI analysis outputs

**Files:**
- Modify: `.github/workflows/slack-notify-failed-jobs.yml`

The existing `notify` job needs: `needs`, `if: always()`, a new step to build the AI block, and the Slack payload updated to include it.

- [ ] **Step 1: Add `needs` and `if` to the `notify` job**

Change the `notify:` job definition from:

```yaml
  notify:
    runs-on: gha-runner-scale-set-ubuntu-22.04-amd64-small
```

to:

```yaml
  notify:
    needs: [ai-analysis]
    if: ${{ always() }}
    runs-on: gha-runner-scale-set-ubuntu-22.04-amd64-small
```

- [ ] **Step 2: Add the "Build AI analysis block" step**

Insert after the "Build extra block" step (after the existing `extra_block` step), before "Send Slack Notification":

```yaml
      - name: Build AI analysis block
        id: ai_block
        if: ${{ steps.extract_jobs.outputs.has_results == 'true' && needs.ai-analysis.outputs.has_analysis == 'true' }}
        shell: bash
        env:
          AI_SUMMARY: ${{ needs.ai-analysis.outputs.summary }}
          RUN_URL: ${{ github.server_url }}/${{ github.repository }}/actions/runs/${{ inputs.run_id }}
        run: |
          set -euo pipefail

          divider='{"type":"divider"}'

          analysis_block=$(jq -n -c --arg summary "$AI_SUMMARY" '{
            type: "rich_text",
            elements: [{
              type: "rich_text_section",
              elements: [
                { type: "text", text: "AI Analysis: ", style: { bold: true } },
                { type: "text", text: $summary }
              ]
            }]
          }')

          link_block=$(jq -n -c --arg url "$RUN_URL" '{
            type: "rich_text",
            elements: [{
              type: "rich_text_section",
              elements: [
                { type: "text", text: "Full Analysis: ", style: { bold: true } },
                { type: "link", url: $url, text: "View artifact" }
              ]
            }]
          }')

          echo "divider_block=- ${divider}" >> "$GITHUB_OUTPUT"
          echo "analysis_block=- ${analysis_block}" >> "$GITHUB_OUTPUT"
          echo "link_block=- ${link_block}" >> "$GITHUB_OUTPUT"
```

Each output is a single-line value, matching the existing `extra_block` and `failed_section` patterns. This avoids the multiline heredoc indentation issue where GitHub Actions only indents the first line of a multiline substitution - which would break the Slack payload YAML parsing.

- [ ] **Step 3: Add the AI block outputs to the Slack payload**

In the "Send Slack Notification" step, append the three AI block outputs after the last existing block interpolation. Change the end of the payload from:

```yaml
              ${{ steps.job_sections.outputs.failed_section }}
              ${{ steps.job_sections.outputs.cancelled_section }}
              ${{ steps.extra_block.outputs.block }}
```

to:

```yaml
              ${{ steps.job_sections.outputs.failed_section }}
              ${{ steps.job_sections.outputs.cancelled_section }}
              ${{ steps.extra_block.outputs.block }}
              ${{ steps.ai_block.outputs.divider_block }}
              ${{ steps.ai_block.outputs.analysis_block }}
              ${{ steps.ai_block.outputs.link_block }}
```

- [ ] **Step 4: Validate YAML syntax**

Run: `python3 -c "import yaml; yaml.safe_load(open('.github/workflows/slack-notify-failed-jobs.yml'))"`
Expected: No output (valid YAML)

- [ ] **Step 5: Commit**

```bash
git add .github/workflows/slack-notify-failed-jobs.yml
git commit -m "feat: enrich Slack notification with AI analysis summary"
```

---

### Task 4: Update `main.yml` to enable AI analysis

**Files:**
- Modify: `.github/workflows/main.yml:9-13` (permissions)
- Modify: `.github/workflows/main.yml:310-331` (notify job)

- [ ] **Step 1: Add AI analysis inputs, secret, and job-level permissions to notify job**

Change the notify job call from:

```yaml
    uses: ./.github/workflows/slack-notify-failed-jobs.yml
    with:
      title: "Main workflow Failed"
      run_id: ${{ github.run_id }}
    secrets:
      channel_id: ${{ secrets.SLACK_ENGINEERING_ALERTS_CHANNEL_ID }}
      slack_bot_token: ${{ secrets.SLACK_GITHUB_ACTIONS_ALERTS_BOT_TOKEN }}
```

to:

```yaml
    permissions:
      contents: read
      actions: read
      id-token: write
    uses: ./.github/workflows/slack-notify-failed-jobs.yml
    with:
      title: "Main workflow Failed"
      run_id: ${{ github.run_id }}
      enable_ai_analysis: true
      ai_analysis_model: "claude-sonnet-4-20250514"
    secrets:
      channel_id: ${{ secrets.SLACK_ENGINEERING_ALERTS_CHANNEL_ID }}
      slack_bot_token: ${{ secrets.SLACK_GITHUB_ACTIONS_ALERTS_BOT_TOKEN }}
      anthropic_api_key: ${{ secrets.ANTHROPIC_API_KEY }}
```

- [ ] **Step 2: Validate YAML syntax**

Run: `python3 -c "import yaml; yaml.safe_load(open('.github/workflows/main.yml'))"`
Expected: No output (valid YAML)

- [ ] **Step 3: Commit**

```bash
git add .github/workflows/main.yml
git commit -m "feat: enable AI failure analysis in main workflow notify"
```

---

### Task 5: Final validation and review

**Files:**
- Review: `.github/workflows/slack-notify-failed-jobs.yml` (full file)
- Review: `.github/workflows/main.yml` (permissions + notify job)

- [ ] **Step 1: Validate both workflow files parse as valid YAML**

```bash
python3 -c "
import yaml
for f in ['.github/workflows/slack-notify-failed-jobs.yml', '.github/workflows/main.yml']:
    yaml.safe_load(open(f))
    print(f'{f}: OK')
"
```

Expected:
```
.github/workflows/slack-notify-failed-jobs.yml: OK
.github/workflows/main.yml: OK
```

- [ ] **Step 2: Verify `ai-analysis` job structure**

Check that the `ai-analysis` job has:
- `if: ${{ inputs.enable_ai_analysis }}` (job-level conditional)
- `timeout-minutes: 10`
- `permissions` with `id-token: write` at job level
- `outputs` declaring `summary` and `has_analysis`
- 6 steps: Checkout, Fetch logs, Download artifact, Claude analysis, Upload artifact, Extract summary

Run: `grep -c 'ai-analysis\|enable_ai_analysis\|claude-code-action\|ai-failure-analysis\|extract_summary' .github/workflows/slack-notify-failed-jobs.yml`
Expected: 7+ matches

- [ ] **Step 3: Verify `notify` job changes**

Check that the `notify` job has:
- `needs: [ai-analysis]`
- `if: ${{ always() }}`
- "Build AI analysis block" step referencing `needs.ai-analysis.outputs`
- Slack payload including `${{ steps.ai_block.outputs.divider_block }}`, `${{ steps.ai_block.outputs.analysis_block }}`, `${{ steps.ai_block.outputs.link_block }}`

- [ ] **Step 4: Verify `main.yml` changes**

Check that `main.yml` has:
- `id-token: write` in permissions
- `enable_ai_analysis: true` in notify inputs
- `anthropic_api_key` in notify secrets

- [ ] **Step 5: Verify no changes to other callers**

Confirm no other workflow files were modified:

```bash
git diff --name-only HEAD~4
```

Expected: Only these two files:
```
.github/workflows/slack-notify-failed-jobs.yml
.github/workflows/main.yml
```

- [ ] **Step 6: Review complete workflow file**

Read the full `slack-notify-failed-jobs.yml` end-to-end. Verify:
- Inputs section has 6 inputs (3 existing + 3 new)
- Secrets section has 3 secrets (2 existing + 1 new)
- Two jobs: `ai-analysis` then `notify`
- `ai-analysis` has correct permissions, conditional, timeout
- `notify` uses `needs: [ai-analysis]` and `if: always()`
- Slack payload includes AI block at the end
