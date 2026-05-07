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
2. Stage only those files:

```bash
git add path/to/file1 path/to/file2
```

3. Do not use `git add .`.
4. Commit with:

```bash
git commit -m "fix(misc): address {bot-name} feedback in {file}"
```

5. Capture the commit SHA:

```bash
git rev-parse HEAD
```

6. Build the commit URL:

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
