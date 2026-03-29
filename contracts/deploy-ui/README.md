# Contract deploy UI (browser signing)

When `DEPLOY_WITH_UI=true`, Hardhat deploy scripts can route transactions through this small Next.js app so you approve them in a browser wallet (MetaMask, Rabby, etc.) instead of putting `DEPLOYER_PRIVATE_KEY` in the environment.

The Hardhat side (`contracts/scripts/hardhat/deployment-ui.ts`) starts a local HTTP bridge, spawns `next dev` for this package on a free port, opens the UI, and waits for connect → chain switch → per-tx approval.

## Prerequisites

1. **Monorepo install** (from repo root):

   ```bash
   pnpm install
   ```

2. **Local Linea stack** running so the RPC you deploy against is up. For example:

   ```bash
   make start-env
   ```

   or the tracing variant from [Local development guide](../../docs/local-development-guide.md). The stack exposes **L1 JSON-RPC on port `8445`** and **L2 on `8545`** (see `makefile-contracts.mk`).

3. **Wallet** with ETH on the chain you deploy to (funded dev account on local L1/L2 as appropriate).

## Network: `zkevm_dev` (local L1)

Hardhat network `zkevm_dev` uses `L1_RPC_URL` when set, otherwise defaults to `http://127.0.0.1:8545`. For the docker stack, point L1 at **8445**:

| Item | Value |
|------|--------|
| Hardhat `--network` | `zkevm_dev` |
| Env | `L1_RPC_URL=http://127.0.0.1:8445` |
| Chain ID | **`31648428`** for the docker local L1 (`docker/config/l1-node/el/genesis.json`). Use **`59139`** only if your RPC is the hosted Linea devnet (e.g. `https://rpc.devnet.linea.build`). |

Add a custom network in your wallet matching the RPC you use (same URL and chain ID the node reports).

## Example: deploy `TestERC20` on local L1

From the **`contracts`** package directory:

```bash
cd contracts

L1_RPC_URL=http://127.0.0.1:8445 \
TEST_ERC20_NAME="LocalUI" \
TEST_ERC20_SYMBOL="LUI" \
TEST_ERC20_INITIAL_SUPPLY="1000000" \
DEPLOY_WITH_UI=true \
npx hardhat deploy --network zkevm_dev --tags TestERC20
```

What happens:

1. A browser tab opens to the local deploy UI (unless you set `DEPLOY_WITH_UI_OPEN_BROWSER=false`).
2. Connect the wallet and switch to the chain your RPC uses (`31648428` for local docker L1) if prompted.
3. Approve the deployment transaction in the wallet; the CLI continues when the tx is broadcast.

Artifacts and `deployments/zkevm_dev/` behave the same as a private-key deploy.

## Network: `l2` (local L2)

Use Hardhat network `l2` and set **`L2_RPC_URL`** (for the stack, typically `http://127.0.0.1:8545`). The UI will use the chain reported by that RPC—add the same RPC/chain ID in your wallet if needed.

```bash
cd contracts

L2_RPC_URL=http://127.0.0.1:8545 \
DEPLOY_WITH_UI=true \
npx hardhat deploy --network l2 --tags <YourL2Tag>
```

## Optional environment variables

| Variable | Effect |
|----------|--------|
| `DEPLOY_WITH_UI=true` | Enable browser signing (default is off; scripts use `DEPLOYER_PRIVATE_KEY`). You do **not** need `DEPLOYER_PRIVATE_KEY` when this is set—Hardhat uses an RPC-only provider so the placeholder key is not loaded. |
| `DEPLOY_WITH_UI_OPEN_BROWSER=false` | Do not auto-open a tab; open the URL printed in the terminal manually. |
| `DEPLOY_WITH_UI_DEBUG=true` | Forward `next dev` stdout/stderr to the terminal (noisier). |

## Manual UI dev (optional)

You normally do **not** run the UI yourself—Hardhat starts it. For UI-only debugging:

```bash
cd contracts/deploy-ui
pnpm dev
```

You still need the bridge server from a deploy run for full signing flow.

## Further reading

- [Linea deployment scripts overview](../docs/deployment/README.md) — env vars, tags, verification.
- [Chained deployments](../docs/deployment/chained-deployments.md) — UI sessions are **per deploy file**, so chained `--tags` get separate connect/sign phases.

## Limitations

- Scripts that broadcast **fixed pre-signed** transactions (e.g. some system-contract deploys) do not use the browser signer.
- One in-flight transaction per deploy file at a time; the UI is for interactive signing, not batch automation.
