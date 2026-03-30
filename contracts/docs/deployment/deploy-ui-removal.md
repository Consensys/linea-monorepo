# Removing the optional deploy UI (`DEPLOY_WITH_UI`)

[← Back to deployment index](README.md)

This document is for **maintainers** who want to delete the browser-signing stack entirely and return to **private-key / named-account deploys only**. It lists every integration point in this repository as of the introduction of `DEPLOY_WITH_UI`.

Do **not** follow this guide for day-to-day deployments. Operators should simply leave `DEPLOY_WITH_UI` unset.

## What gets removed

| Area | Path / symbol |
|------|----------------|
| Bridge + session logic | `contracts/scripts/hardhat/deployment-ui.ts` |
| Next.js app | `contracts/deploy-ui/` (package `@consensys/linea-contract-deploy-ui`) |
| Hardhat hook | `subtask(TASK_DEPLOY_RUN_DEPLOY)` in `contracts/hardhat.config.ts` |
| Config branch | `deployerAccounts()` UI branch in `contracts/hardhat.config.ts` |
| Deploy script wrappers | `withDeploymentUiSession`, `getDeploymentSigner`, `setDeploymentUiNextTransactionContext` across `contracts/deploy/*.ts` |
| Shared helpers | `resolveDeploymentRunner` / UI context in `contracts/scripts/hardhat/utils.ts` |
| ESLint ignore | `deploy-ui/**` in `contracts/eslint.config.mjs` |
| Documentation | Sections in `contracts/docs/deployment/README.md`, `chained-deployments.md`, `contracts/deploy-ui/README.md`, this file; cross-links from `contracts/AGENTS.md`, `docs/tech/components/contracts.md` |

CI workflows under `.github/` do **not** reference `deploy-ui` today; re-grep after removal.

## 1. `hardhat.config.ts`

1. Remove:

   - `import { createRequire } from "node:module";`
   - `import { subtask } from "hardhat/config"` (if nothing else uses `subtask`)
   - `import { TASK_DEPLOY_RUN_DEPLOY } from "hardhat-deploy"`
   - `const requireFromConfig = createRequire(__filename);`
   - The entire `subtask(TASK_DEPLOY_RUN_DEPLOY).setAction(...)` block that `require()`s `./scripts/hardhat/deployment-ui.ts`.

2. Simplify **`deployerAccounts()`**: delete the `if (process.env.DEPLOY_WITH_UI === "true") { return []; }` branch and always return `[process.env.DEPLOYER_PRIVATE_KEY || EMPTY_HASH]` (or your preferred account wiring).

## 2. Delete UI and bridge implementation

- Delete **`contracts/scripts/hardhat/deployment-ui.ts`**.
- Delete the directory **`contracts/deploy-ui/`** (including `package.json`, Next app, and lockfile references will update on the next `pnpm install` from the repo root).

## 3. `contracts/scripts/hardhat/utils.ts`

1. Remove imports from `./deployment-ui` (`resolveDeploymentRunner`, `setDeploymentUiNextTransactionContext`).

2. **`pushUiDeployContext`** / **`setDeploymentUiNextTransactionContext`**: remove the whole UI context helper and every call to `pushUiDeployContext` inside `deployFromFactory`, `deployFromFactoryWithOpts`, and the `deployUpgradable*` helpers — or inline constructor logging only if you still want console output without the UI.

3. **`resolveDeploymentRunner`**: reintroduce a **local** async helper that matches the **non-UI** branches previously in `deployment-ui.ts`, for example:

   - If `runnerOrProvider` is already an `AbstractSigner`, return it.
   - Else if it is a provider with `getSigner`, call `getSigner()`.
   - Else `return await ethers.provider.getSigner()` (or `hre.getNamedAccounts()` + `getSigner(deployer)` if you require named accounts everywhere).

   Each `deployFromFactory*` caller currently passes `null` or a signer; keep behaviour identical to “UI off” today.

4. Run **`pnpm -F contracts run lint:ts`** and fix any unused imports.

## 4. Deploy scripts under `contracts/deploy/`

List every consumer (the set changes over time):

```bash
rg 'from "../scripts/hardhat/deployment-ui"' contracts/deploy
rg "withDeploymentUiSession|getDeploymentSigner|setDeploymentUiNextTransactionContext" contracts/deploy
```

Scripts that **only** use fixed calldata / raw broadcast (e.g. some `EIP*SystemContract` predeploys) may never have imported `deployment-ui`; no change needed there.

### Pattern A — `withDeploymentUiSession` + `getDeploymentSigner(hre)`

**Example today:**

```typescript
const func: DeployFunction = withDeploymentUiSession("01_deploy_PlonkVerifier.ts", async function (hre) {
  const signer = await getDeploymentSigner(hre);
  // ...
});
```

**After removal:**

1. Drop the `withDeploymentUiSession` import and wrapper; export a plain `DeployFunction` async body.
2. Replace the signer with the standard Hardhat pattern:

   ```typescript
   const { deployer } = await hre.getNamedAccounts();
   const signer = await hre.ethers.getSigner(deployer);
   ```

3. Pass `signer` into your factories / `deployFromFactory` the same way as today.
4. Remove any **`setDeploymentUiNextTransactionContext`** calls (e.g. `11_deploy_TestERC20.ts`).

### Pattern B — `withDeploymentUiSession` only (uses module `ethers` / no `getDeploymentSigner`)

Remove the `withDeploymentUiSession` import and wrapper; keep the inner async function as the deploy function. Ensure the implicit or explicit signer still matches your network `namedAccounts` expectations.

### Final check

```bash
rg "deployment-ui|withDeploymentUiSession|getDeploymentSigner|setDeploymentUiNextTransactionContext" contracts/deploy
```

Expect **no matches**.

## 5. Tooling and workspace

1. **`contracts/eslint.config.mjs`** — remove `deploy-ui/**` from `ignores` (optional once the folder is gone).

2. **Root `pnpm install`** — refreshes the lockfile after removing `contracts/deploy-ui`.

3. **Search the monorepo** for stray references:

   ```bash
   rg "DEPLOY_WITH_UI|deployment-ui|linea-contract-deploy-ui|deploy-ui/" --glob '!**/node_modules/**' --glob '!**/pnpm-lock.yaml'
   ```

## 6. Documentation cleanup

Remove or rewrite:

- `contracts/docs/deployment/README.md` — “Browser wallet signing” section and links to `deploy-ui-removal.md`.
- `contracts/docs/deployment/chained-deployments.md` — “UI-backed chained deployments” section.
- `contracts/deploy-ui/README.md` — delete with the folder.
- `contracts/docs/deployment/deploy-ui-removal.md` — delete this file last.
- `contracts/AGENTS.md` — deploy UI bullet (if added).
- `docs/tech/components/contracts.md` — optional deploy UI paragraph (if added).

## 7. Verification

```bash
pnpm -F contracts run build
pnpm -F contracts run lint
pnpm -F contracts run test   # or a targeted deploy dry-run on hardhat network
```

Run a **non-UI** deploy smoke test (e.g. `TestERC20` on `hardhat`) with `DEPLOYER_PRIVATE_KEY` set to confirm parity with pre-UI behaviour.

## Rollback

Restore the deleted files and config from git history; re-run `pnpm install` at the repo root.
