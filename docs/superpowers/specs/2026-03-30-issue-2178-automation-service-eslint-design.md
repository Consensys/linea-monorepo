# Issue 2178 - Automation Service ESLint 9 Follow-up Design

## Summary

This spec covers a narrow follow-up for GitHub issue `#2178` in `native-yield-operations/automation-service`.

The package currently disables `import/order` in its local flat ESLint config, which leaves it partially out of alignment with the shared ESLint 9 configuration introduced in PR `#2068`.

The goal is to remove that local override, run package-local autofixes, and leave the package in a state where its lint command passes.

## Context

- PR `#2068` migrated repository ESLint configuration to ESLint 9 flat config.
- `native-yield-operations/automation-service` still contains a local override in `eslint.config.mjs`:
  - `"import/order": "off"`
- `native-yield-operations/lido-governance-monitor` is not part of this change because it does not have the same package-local override.
- The shared rule definition already exists in `ts-libs/eslint-config/base.js`.

## Scope

In scope:

- Remove the `import/order` disable from `native-yield-operations/automation-service/eslint.config.mjs`.
- Run package-local ESLint autofixes for `@consensys/linea-native-yield-automation-service`.
- Keep the resulting package diff limited to:
  - the config edit
  - deterministic ESLint autofixes required for lint to pass
- Re-run package-local lint and confirm it passes.

Out of scope:

- Changes to `native-yield-operations/lido-governance-monitor`
- Changes to shared ESLint config under `ts-libs/eslint-config`
- Manual refactors unrelated to lint compliance
- Behavioral code changes
- Test changes, unless a lint-driven file edit unexpectedly requires one

## Recommended Approach

Use the existing shared ESLint 9 configuration as-is and remove the package-local exception.

Then run `lint:fix` for `@consensys/linea-native-yield-automation-service`, review the resulting diff, and keep only safe autofix output inside the package. This is the smallest approach that fully closes the known gap without expanding the issue into a broader cleanup effort.

## Alternatives Considered

### Option 1 - Config change plus targeted manual import ordering

Remove the override and manually reorder only the imports that violate `import/order`.

Why not chosen:

- More manual effort for no real benefit if ESLint autofix can apply the same changes safely.

### Option 2 - Package-local autofix cleanup after enabling the rule

Remove the override, run `lint:fix`, keep the resulting safe package-local autofixes, and verify `lint` passes.

Why chosen:

- Matches requested scope
- Keeps the change inside one package
- Produces a clean post-migration state

### Option 3 - Re-enable the rule only for part of the package

Keep a narrower override for selected paths such as scripts.

Why not chosen:

- Preserves technical debt
- Not justified by the current issue scope

## Execution Plan

Prerequisite:

- Use the repository-supported Node.js runtime (`>= 22.22.2`) before running lint commands.

1. Edit `native-yield-operations/automation-service/eslint.config.mjs` to remove the `import/order` override.
2. Run `pnpm --filter @consensys/linea-native-yield-automation-service lint:fix`.
3. Review the resulting diff and confirm the changes are limited to the package and to safe ESLint autofix output:
   - import reordering
   - related whitespace and newline normalization
4. Run `pnpm --filter @consensys/linea-native-yield-automation-service lint`.

## Verification

Primary verification:

- `pnpm --filter @consensys/linea-native-yield-automation-service lint`

Secondary verification:

- Review the final diff and confirm it contains only:
  - the config change
  - package-local lint autofixes

## Risks And Response

### Risk: Autofix touches more files than expected

Response:

- Review the diff before accepting it.
- If the diff expands into unrelated churn, stop and surface the exact files and rule categories before proceeding further.

### Risk: Enabling `import/order` exposes non-autofixable lint issues

Response:

- Stop before broad manual cleanup.
- Report the exact blocking files and rules so the scope can be re-decided explicitly.

## Success Criteria

- `native-yield-operations/automation-service` no longer disables `import/order`
- Package-local `lint:fix` has been applied as needed
- `pnpm --filter @consensys/linea-native-yield-automation-service lint` passes
- The diff remains limited to the automation service package and this specific ESLint 9 follow-up
