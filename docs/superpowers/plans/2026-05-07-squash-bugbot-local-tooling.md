# Squash Bugbot Local Tooling Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add one canonical local `squash-bugbot` workflow for Codex, Claude Code, and Cursor contributors.

**Architecture:** Store the workflow once as a repo-scoped skill in `.agents/skills/squash-bugbot/SKILL.md`. Expose Claude Code local slash-command access through a symlink in `.claude/commands/`, and update only existing agent discoverability docs.

**Tech Stack:** Markdown agent skills, Git symlink, GitHub CLI `gh`, GraphQL through `gh api`, repository documentation.

---

## File Structure

- Create `.agents/skills/squash-bugbot/SKILL.md`: canonical, instruction-only skill used by Codex and Cursor and targeted by the Claude Code command symlink.
- Create `.claude/commands/squash-bugbot.md`: symlink to `../../.agents/skills/squash-bugbot/SKILL.md`.
- Modify `AGENTS.md`: add local workflow discoverability and update the `.agents/skills/` repository map description.
- Modify `CLAUDE.md`: point Claude Code users to `/squash-bugbot <PR_NUMBER>` and the canonical skill source.
- Modify `.cursor/rules/documentation.mdc`: point Cursor users to the canonical `.agents` skill location.
- Do not modify `README.md`.
- Do not create `.cursor/skills`.

## Plan Review Gate

Before Task 1, satisfy the repository `AGENTS.md` plan-review rule: review this plan with an isolated reviewer that receives only the original request, the approved design spec, and this plan file. The review should check that the implementation keeps one canonical skill, does not update `README.md`, does not create `.cursor/skills`, and uses the allowed `misc` commit scope for generated commits.

### Task 1: Add Canonical Skill

**Files:**
- Create: `.agents/skills/squash-bugbot/SKILL.md`

- [ ] **Step 1: Create the skill directory**

Run:

```bash
mkdir -p .agents/skills/squash-bugbot
```

Expected: command exits 0.

- [ ] **Step 2: Add the canonical skill content**

Create `.agents/skills/squash-bugbot/SKILL.md` with this exact content:

````markdown
---
name: squash-bugbot
description: Triage unresolved bot review comments on a GitHub PR. Use when asked to run /squash-bugbot <PR_NUMBER> or to assess, fix, dismiss, reply to, or resolve bot review feedback.
---

# Squash Bugbot

This skill triages unresolved bot review comments on a GitHub PR by assessing validity, proposing fixes, and optionally applying fixes or dismissing comments.

## Usage

```text
/squash-bugbot <PR_NUMBER>
```

`PR_NUMBER` is required. If it is missing, stop with:

```text
Usage: /squash-bugbot <PR_NUMBER>
```

## Preconditions

- Run from a local checkout of the target GitHub repository.
- Use the current local source code for assessment.
- Require the `gh` CLI.
- Require `gh` authentication that can read PR comments, write PR comments, resolve review threads, and push to the current branch.
- Tell the user that assessment uses the current local source code, so their branch should be up to date.

## Step 1: Resolve Repository

Run:

```bash
git remote get-url origin
```

Parse `owner/repo` from the remote URL:

- HTTPS: extract the path after `github.com/`, removing a trailing `.git` when present.
- SSH: extract the path after `git@github.com:`, removing a trailing `.git` when present.
- Accept URLs with or without `.git`.

If no `origin` remote exists, stop with:

```text
Error: no git origin remote found. This command requires a GitHub remote.
```

## Step 2: Check GitHub CLI

Verify `gh` is installed and authenticated:

```bash
command -v gh
gh auth status
```

If either command fails, stop with:

```text
Error: `gh` CLI is not available or not authenticated. Run `gh auth login` first.
```

## Step 2.5: Check Local Git State and PR Head

Check the local git state before assessing comments or taking actions:

```bash
git status --short --branch
```

If unrelated staged changes already exist, stop and ask the user before applying fixes. Use the status output to record files with local changes. Before editing an approved target file, if that file already contains unrelated local changes, stop and ask whether to preserve, include, or skip those changes.

Fetch PR head metadata:

```bash
gh pr view {PR_NUMBER} --json headRefName,headRepositoryOwner,headRepository,headRefOid
```

Compare the PR head metadata with the current checkout using:

```bash
git branch --show-current
git rev-parse HEAD
git rev-parse --abbrev-ref --symbolic-full-name @{u}
git rev-parse @{u}
```

Confirm the local branch matches `headRefName` and the upstream or matching remote branch belongs to `headRepositoryOwner` and `headRepository`. Require local `HEAD` to match `headRefOid` before applying fixes. Do not treat an upstream match alone as sufficient. If local `HEAD` differs from `headRefOid`, stop before applying fixes and ask whether to switch or update the checkout, or continue report-only.

## Step 3: Fetch Comments

Fetch all three data sources. Run independent fetches in parallel when the agent environment supports parallel tool calls.

Review comments:

```bash
gh api repos/{owner}/{repo}/pulls/{PR_NUMBER}/comments --paginate
```

Issue comments:

```bash
gh api repos/{owner}/{repo}/issues/{PR_NUMBER}/comments --paginate
```

Review threads through GraphQL:

```graphql
query($owner: String!, $repo: String!, $number: Int!, $after: String) {
  repository(owner: $owner, name: $repo) {
    pullRequest(number: $number) {
      reviewThreads(first: 100, after: $after) {
        nodes {
          id
          isResolved
          comments(first: 1) {
            nodes {
              databaseId
            }
          }
        }
        pageInfo {
          hasNextPage
          endCursor
        }
      }
    }
  }
}
```

Paginate GraphQL review threads while `pageInfo.hasNextPage` is true.

If GitHub returns 404 for the PR, stop with:

```text
Error: PR #{PR_NUMBER} not found in {owner}/{repo}.
```

## Step 4: Detect Bots

A comment is from a bot if either condition is true:

- REST API field `user.type == "Bot"`.
- REST API field `user.login` is one of:
  - `github-actions[bot]`
  - `coderabbitai`
  - `cursor[bot]`
  - `copilot`
  - `sonarcloud[bot]`
  - `codecov[bot]`
  - `dependabot[bot]`

Use REST API metadata for bot detection. Do not use GraphQL author login for bot detection, because GraphQL and REST may represent bot logins differently.

## Step 5: Filter to Unresolved Bot Comments

For review comments:

1. Use GraphQL review threads as the source of truth for `isResolved`.
2. Match REST review comment `id` to the first GraphQL thread comment `databaseId`.
3. Keep the REST review comment when it is from a bot and the matching GraphQL thread has `isResolved == false`.
4. Preserve the GraphQL review thread node `id`; it is required for the `resolveReviewThread` mutation.

For issue comments:

1. Keep all bot issue comments.
2. Treat them as not resolvable because issue comments do not have review thread resolution.

If no comments remain, report:

```text
No unresolved bot comments found on PR #{PR_NUMBER}
```

Then stop.

## Step 6: Assess Each Comment

For each retained comment:

1. Read the comment body and summarize the concern in one sentence.
2. Identify the referenced file and line:
   - For review comments, use REST fields `path`, `line`, and `original_line`.
   - For issue comments, parse file references from the body when present.
3. Read the local source file around the referenced line.
4. If the file is missing locally, mark the verdict as `Uncertain` with proposed fix `File not found locally - skipping fix`.
5. Assess the concern against the current local source code.
6. Categorize the verdict as:
   - `Valid`: the bot identified a real problem.
   - `Invalid`: the concern is a false positive, already handled, or outdated.
   - `Uncertain`: more context is required.

## Step 7: Report

Present one consolidated report before taking action:

```markdown
### [BotName] - {verdict}
**File:** `path/to/file.ts` L{line}
**Comment:** [link to GitHub comment]
**Concern:** {one-line summary of what the bot flagged}
**Verdict:** {Valid/Invalid/Uncertain} - {reasoning}
**Proposed fix:** {concrete code change if valid or uncertain, or "N/A" if invalid}
```

## Step 8: Ask Before Acting

Ask the user what to do for each comment:

- For `Valid`, ask: `Apply this fix?`
- For `Invalid`, ask: `Dismiss this comment on the PR?`
- For `Uncertain`, ask whether to treat the comment as valid, treat it as invalid, or skip it.

Do not edit files, commit, push, reply, or resolve threads without the user's answer for that comment.

## Step 9: Apply Selected Fixes

For each approved fix:

1. Edit only the file or files required for that specific fix.
2. Before editing each target file, re-check whether it has unrelated local changes. If it does, stop and ask before continuing.
3. Stage only the approved files for that fix:

```bash
git add path/to/file1 path/to/file2
```

4. Do not use `git add .`.
5. Verify the staged files:

```bash
git diff --cached --name-only
git diff --cached
```

If any staged file is not approved for that fix, stop and ask the user to clear or handle unrelated staged changes. Review the cached diff contents before committing. If the cached diff contains anything outside the approved fix, even within approved files, stop and ask the user how to handle it.

6. Commit with:

```bash
git commit -m "fix(misc): address {bot-name} feedback in {file}"
```

7. Capture the commit SHA:

```bash
git rev-parse HEAD
```

8. Build the commit URL:

```text
https://github.com/{owner}/{repo}/commit/{sha}
```

## Step 10: Reply and Resolve

For an approved fixed review comment, reply to the review thread with a concise explanation and commit link:

```bash
gh api repos/{owner}/{repo}/pulls/{PR_NUMBER}/comments/{comment_id}/replies -f body="{message}"
```

Then resolve the review thread with its GraphQL thread node ID:

```bash
gh api graphql -f query='mutation { resolveReviewThread(input: {threadId: "NODE_ID"}) { thread { isResolved } } }'
```

For an approved fixed issue comment, reply with:

```bash
gh api repos/{owner}/{repo}/issues/{PR_NUMBER}/comments -f body="{message}"
```

For an approved invalid review comment, reply with a concise dismissal explanation and resolve the review thread with the same `resolveReviewThread` mutation.

For an approved invalid issue comment, reply with a concise dismissal explanation. Do not attempt to resolve issue comments.

## Step 11: Push Once

If commits were created, push once after all selected actions:

```bash
git push
```

If push fails, report the error and tell the user to push manually. Do not undo local commits.

## Step 12: Final Summary

Report:

- Created commits: SHA and message for each fix.
- Dismissed comments.
- Skipped comments.
- Push status.

## Safety Rules

- Only process bot comments.
- Ignore human review comments.
- Do not undo unrelated local changes.
- Do not log secrets, credentials, private keys, tokens, or sensitive RPC URLs.
- Mask sensitive values if they must appear in output.
- Use current local source code for assessment, not only the PR diff.
````

- [ ] **Step 3: Verify the frontmatter**

Run:

```bash
sed -n '1,8p' .agents/skills/squash-bugbot/SKILL.md
```

Expected output includes:

```text
---
name: squash-bugbot
description: Triage unresolved bot review comments on a GitHub PR. Use when asked to run /squash-bugbot <PR_NUMBER> or to assess, fix, dismiss, reply to, or resolve bot review feedback.
---
```

- [ ] **Step 4: Commit the canonical skill**

Run:

```bash
git add .agents/skills/squash-bugbot/SKILL.md
git commit -m "docs(misc): add squash bugbot skill"
```

Expected: commit succeeds.

### Task 2: Add Claude Command Symlink

**Files:**
- Create: `.claude/commands/squash-bugbot.md` symlink

- [ ] **Step 1: Create the Claude commands directory**

Run:

```bash
mkdir -p .claude/commands
```

Expected: command exits 0.

- [ ] **Step 2: Create the symlink**

Run:

```bash
ln -s ../../.agents/skills/squash-bugbot/SKILL.md .claude/commands/squash-bugbot.md
```

Expected: command exits 0.

- [ ] **Step 3: Verify the symlink target**

Run:

```bash
test -L .claude/commands/squash-bugbot.md
test "$(readlink .claude/commands/squash-bugbot.md)" = "../../.agents/skills/squash-bugbot/SKILL.md"
```

Expected: both commands exit 0.

- [ ] **Step 4: Commit the symlink**

Run:

```bash
git add .claude/commands/squash-bugbot.md
git commit -m "docs(misc): expose squash bugbot claude command"
```

Expected: commit succeeds.

### Task 3: Update Agent Discoverability Docs

**Files:**
- Modify: `AGENTS.md:9-24`
- Modify: `AGENTS.md:383-384`
- Modify: `CLAUDE.md:12-18`
- Modify: `.cursor/rules/documentation.mdc:21-37`

- [ ] **Step 1: Add local workflow discoverability to `AGENTS.md`**

Insert this section immediately after the `## Agent Entry Points` list and before `## Discoverability Index`:

```markdown
## Local Agent Workflows

- `squash-bugbot`: canonical skill at `.agents/skills/squash-bugbot/SKILL.md`.
- Use `/squash-bugbot <PR_NUMBER>` locally in Codex, Cursor, or Claude Code to triage unresolved bot PR comments.
- Claude Code exposes the same workflow through `.claude/commands/squash-bugbot.md`, a symlink to the canonical skill.
- Requires an authenticated `gh` CLI with access to read PR comments, write PR comments, resolve review threads, and push the current branch.
```

Replace the `.agents/skills/` row in the repository map with:

```text
.agents/skills/          Agent skills (smart contract, dependency maintenance, local PR workflows)
```

- [ ] **Step 2: Update `CLAUDE.md` discoverability**

Add this bullet under `## Discoverability` after the `Global agent index` bullet:

```markdown
- Local slash commands: `.claude/commands/` (`/squash-bugbot <PR_NUMBER>` links to `.agents/skills/squash-bugbot/SKILL.md`)
```

- [ ] **Step 3: Update Cursor documentation index**

Add this section in `.cursor/rules/documentation.mdc` after `## Rules` and before `## Package-Specific Agent Docs`:

```markdown
## Local Skills

- Squash Bugbot skill: `/.agents/skills/squash-bugbot/SKILL.md`
- Invocation: `/squash-bugbot <PR_NUMBER>` in agents that support project skills.
- This repository does not maintain a separate `/.cursor/skills` copy; use the canonical `/.agents/skills` skill.
```

- [ ] **Step 4: Verify discoverability references**

Run:

```bash
rg -n "squash-bugbot|squash bugbot|Local Agent Workflows|Local Skills" AGENTS.md CLAUDE.md .cursor/rules/documentation.mdc
```

Expected: output includes references in all three files.

- [ ] **Step 5: Commit documentation updates**

Run:

```bash
git add AGENTS.md CLAUDE.md .cursor/rules/documentation.mdc
git commit -m "docs(misc): document squash bugbot local workflow"
```

Expected: commit succeeds.

### Task 4: Final Validation

**Files:**
- Validate: `.agents/skills/squash-bugbot/SKILL.md`
- Validate: `.claude/commands/squash-bugbot.md`
- Validate: `AGENTS.md`
- Validate: `CLAUDE.md`
- Validate: `.cursor/rules/documentation.mdc`
- Validate unchanged: `README.md`

- [ ] **Step 1: Verify canonical skill and symlink**

Run:

```bash
test -f .agents/skills/squash-bugbot/SKILL.md
test -L .claude/commands/squash-bugbot.md
test "$(readlink .claude/commands/squash-bugbot.md)" = "../../.agents/skills/squash-bugbot/SKILL.md"
```

Expected: all commands exit 0.

- [ ] **Step 2: Verify there is no Cursor skill copy**

Run:

```bash
test ! -e .cursor/skills
```

Expected: command exits 0.

- [ ] **Step 3: Verify `README.md` is unchanged**

Run:

```bash
git diff --exit-code -- README.md
```

Expected: command exits 0 with no output.

- [ ] **Step 4: Verify safety hardening references**

Run:

```bash
rg -n "gh pr view|headRefOid|git diff --cached|git status --short --branch" .agents/skills/squash-bugbot/SKILL.md docs/superpowers/specs/2026-05-07-squash-bugbot-local-tooling-design.md docs/superpowers/plans/2026-05-07-squash-bugbot-local-tooling.md
```

Expected: output includes references in the canonical skill, design spec, and implementation plan.

- [ ] **Step 5: Check for placeholders and disallowed em dashes in changed Markdown**

Run:

```bash
em_dash="$(printf '\342\200\224')"
git diff --unified=0 origin/main -- .agents/skills/squash-bugbot/SKILL.md AGENTS.md CLAUDE.md .cursor/rules/documentation.mdc docs/superpowers/specs/2026-05-07-squash-bugbot-local-tooling-design.md docs/superpowers/plans/2026-05-07-squash-bugbot-local-tooling.md | rg -n "^\+.*(T[B]D|T[O]DO|N[E]EDS VERIFICATION|${em_dash})"
```

Expected: command exits 1 with no matches. This checks newly added lines only, so pre-existing text in touched files does not fail validation.

- [ ] **Step 6: Review final status**

Run:

```bash
git status --short --branch
```

Expected: branch is ahead of `origin/main`; no unstaged or staged changes remain after all task commits.

- [ ] **Step 7: Report validation**

Report the validation commands and results to the user. State that no package build or test command was run because the implementation changes only Markdown agent configuration and documentation.
