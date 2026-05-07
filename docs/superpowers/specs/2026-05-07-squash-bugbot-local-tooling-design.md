# Squash Bugbot Local Tooling Design

Date: 2026-05-07

## Context

The repository already exposes agent guidance through `AGENTS.md`, `CLAUDE.md`, `.cursor/rules/documentation.mdc`,
and repo-scoped Codex skills under `.agents/skills/`.

The maintainer has a personal `squash-bugbot` skill that triages unresolved bot comments on a GitHub pull request. The
goal is to make that workflow available to local contributors using Codex, Claude Code, and Cursor without adding a
GitHub Actions workflow or duplicating the workflow body across tools.

## Goals

- Add one canonical repo-scoped `squash-bugbot` skill under `.agents/skills/squash-bugbot/SKILL.md`.
- Preserve the existing personal skill behavior as closely as possible.
- Provide Claude Code local slash-command access through a symlink from `.claude/commands/squash-bugbot.md` to the
  canonical skill.
- Let Codex and Cursor use the canonical `.agents/skills/squash-bugbot/SKILL.md` directly.
- Update existing agent discoverability docs so contributors can find the local workflow.

## Non-Goals

- Do not add or modify GitHub Actions workflows.
- Do not add a `.cursor/skills` compatibility symlink.
- Do not update `README.md`.
- Do not turn the workflow into a custom script or package.
- Do not add new dependencies.

## Architecture

The canonical workflow lives in:

```text
.agents/skills/squash-bugbot/SKILL.md
```

Claude Code local invocation is exposed through:

```text
.claude/commands/squash-bugbot.md -> ../../.agents/skills/squash-bugbot/SKILL.md
```

Codex and Cursor rely on the canonical skill location:

```text
.agents/skills/squash-bugbot/SKILL.md
```

The skill content remains instruction-only. It will not include scripts, bundled assets, package metadata, or dependency
changes.

## Skill Behavior

The canonical skill keeps the same user-facing invocation:

```text
/squash-bugbot <PR_NUMBER>
```

The workflow:

1. Resolves the GitHub `owner/repo` from `git remote get-url origin`.
2. Fetches review comments, issue comments, and review thread status through `gh api`.
3. Joins REST review comments to GraphQL review threads by review comment database ID.
4. Uses REST metadata for bot detection. A comment is from a bot when `user.type == "Bot"` or the login matches a
   known bot: `github-actions[bot]`, `coderabbitai`, `cursor[bot]`, `copilot`, `sonarcloud[bot]`, `codecov[bot]`, or
   `dependabot[bot]`.
5. Keeps unresolved bot review comments and bot issue comments. Review comment resolution status comes from GraphQL
   review threads, matched by database ID. Issue comments have no thread resolution mechanism.
6. Uses the GraphQL review thread node ID, not the comment database ID, when resolving review threads.
7. Reads the referenced local source code.
8. Assesses each bot concern as valid, invalid, or uncertain.
9. Presents a consolidated report with file, line, comment link, concern, verdict, reasoning, and proposed fix.
10. Asks before applying a valid fix, dismissing an invalid comment, or deciding an uncertain comment.
11. For each applied fix, edits only the required file or files, stages only those files, and creates one scoped commit.
   Use a repository-compatible Conventional Commit message: `fix(squash-bugbot): address {bot-name} feedback in {file}`.
12. Replies to the relevant PR thread or issue comment with the decision and commit link when applicable.
13. Resolves review threads through the GraphQL `resolveReviewThread` mutation when applicable.
14. Pushes once after all selected actions.
15. Reports created commits, dismissed comments, skipped comments, and push status.

Where the personal skill names Claude-specific tools, the repo skill should use platform-neutral wording such as
"read the file" and "edit the file using the agent's normal edit tool" while preserving the required behavior.

## Safety

The workflow must keep the existing safety properties:

- Stop if no GitHub `origin` remote exists.
- Stop if `gh` is missing or unauthenticated.
- Stop with a clear message if the PR is not found.
- Skip fixes for referenced files that do not exist locally.
- Stage only explicitly modified files for each fix.
- Do not use `git add .`.
- Do not undo unrelated local changes.
- Do not log secrets, credentials, tokens, private keys, or sensitive RPC URLs.
- If `git push` fails, report the failure and leave local commits intact.

The workflow depends on contributor-local GitHub permissions. The skill should state that contributors need `gh`
authenticated with permission to read PR comments, write PR comments, resolve review threads, push to their branch, and
create commits locally.

## Documentation Updates

Update only the local agent discoverability files:

- `AGENTS.md`: add `squash-bugbot` to the agent skill index and local workflow notes.
- `CLAUDE.md`: mention the project slash command symlink and canonical skill source.
- `.cursor/rules/documentation.mdc`: mention the canonical `.agents/skills/squash-bugbot/SKILL.md` skill.

Do not update `README.md`.

## Validation

After implementation, verify:

- `.agents/skills/squash-bugbot/SKILL.md` exists and has valid frontmatter with `name` and `description`.
- `.claude/commands/squash-bugbot.md` is a symlink to `../../.agents/skills/squash-bugbot/SKILL.md`.
- The documentation updates point to the canonical skill and do not claim a `.cursor/skills` copy exists.
- `README.md` is unchanged.
- `git status --short` shows only the intended implementation files.

No package build or test command is required because the change is documentation and agent configuration only.

## External References

- Codex skills: https://developers.openai.com/codex/skills
- Claude Code skills and commands: https://code.claude.com/docs/en/slash-commands
- Cursor Agent Skills announcement: https://cursor.com/changelog/2-4
