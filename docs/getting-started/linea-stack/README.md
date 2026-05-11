# Linea Stack quickstart (Sepolia L1)

Local Linea stack for testing and evaluating a simple Linea app. It runs a local
L2 stack and uses your Sepolia RPC for L1 deployment and settlement.

The stack includes:

- Linea Besu sequencer and L2 RPC node
- Maru, Shomei, coordinator, postman, web3signer, and prover
- L2 Blockscout API and frontend UI
- Sepolia deployments for the rollup, message service, token bridge, and an
  ERC20 example token

This is a dev quickstart, not a production deployment. Use a disposable Sepolia
key only.

## Security Model

You provide one funded Sepolia key in `.env`:

```bash
L1_DEPLOYER_PRIVATE_KEY=0x...
```

`account-setup` generates fresh runtime keys on first boot and stores them in the
Docker shared volume:

- L1 blob/data-submission signer
- L1 finalization signer
- L1 postman signer
- L2 deployer
- L2 message anchorer
- L2 postman signer

The generated keys are not committed. `docker compose down -v` wipes them and
forces a fresh first boot.

## Prerequisites

| Requirement | Recommended |
|-------------|-------------|
| Docker | v24+ with Compose v2.19+ |
| CPU | 8 cores |
| RAM | 8 GB for dev proofs, 48 GB recommended |
| Disk | Around 80 GB free |
| Sepolia RPC | HTTPS endpoint with archive/debug-style support from your provider |
| Sepolia ETH | Around 1 ETH on the deployer |

The quickstart reserves `0.15 ETH` for each L1 coordinator signer and `0.05 ETH`
for the L1 postman signer. The rest is used for Sepolia deployment gas.

## Setup

```bash
cd docs/getting-started/linea-stack
cp .env.example .env
$EDITOR .env
```

Required values:

```bash
L1_RPC_URL=https://sepolia.infura.io/v3/<your-project-id>
L1_DEPLOYER_PRIVATE_KEY=0x<your-funded-sepolia-key>
```

Optional port and funding overrides are documented in `.env.example`.

## Boot

```bash
docker compose --env-file versions.env --env-file .env --profile stack-partial-prover up -d
```

A cold first boot can take 20-30 minutes because it pulls images, installs tools,
deploys contracts on Sepolia, funds generated signers, and waits for the
coordinator/prover pipeline.

First boot sequence, at a high level:

1. `account-setup` generates runtime keys and precomputes only the two addresses
   needed before boot: L1 `LineaRollupV8` and L2 `L2MessageService`.
2. `l2-genesis-init` renders genesis with only the generated L2 deployer funded,
   plus the precomputed `L2MessageService` funded with 1B ETH.
3. `config-render` renders service configs with the precomputed addresses.
4. `deploy-contracts` deploys Sepolia/L2 contracts, verifies the boot-critical
   addresses, funds generated runtime signers, writes `addresses.json`, and
   patches coordinator config with deploy-time values.
5. Coordinator, prover, postman, and Blockscout start from the rendered/shared
   artifacts.

## Success Checks

Use the helper scripts first:

```bash
./scripts/status.sh
./scripts/links.sh
```

A healthy dev-prover boot should show:

- `addresses.json` present
- coordinator ports listening on `9545` and `9546`
- prover request/response counts populated
- coordinator logs with L1 blob submissions and aggregation/finalization txs
- L2 Blockscout UI available locally

The local L2 may stop at the deployment/funding block height when there is no
new traffic. That is normal for this quickstart. Send an L2 transaction to create
fresh blocks and make Blockscout visibly move.

## Local Endpoints

| Service | URL |
|---------|-----|
| L2 RPC | http://localhost:8745 |
| L2 WebSocket | ws://localhost:8746 |
| L2 Blockscout UI | http://localhost:4001 |
| L2 Blockscout API | http://localhost:4000 |
| Coordinator observability | http://localhost:9545 |
| Postman API | http://localhost:9090 |
| Maru | http://localhost:8080 |
| Sepolia explorer | https://sepolia.etherscan.io |

Use `./scripts/links.sh` to print the exact Sepolia contract links and local L2
contract links for the current boot.

## Useful Commands

```bash
# Service status
docker compose --env-file versions.env --env-file .env --profile stack-partial-prover ps

# Redacted stack status
./scripts/status.sh

# Current L2 block
curl -s -X POST -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' \
  http://localhost:8745

# Stop containers but keep volumes and generated keys
docker compose --env-file versions.env --env-file .env --profile stack-partial-prover down

# Full reset: wipe DBs, generated runtime keys, rendered config, and deploy logs
docker compose --env-file versions.env --env-file .env --profile stack-partial-prover down -v
```

## Current Caveats

- Default mode uses dev/dummy proofs for quick feedback. Partial-prover validation
  is still a separate acceptance gate.
- A real bridge/message smoke test is still pending.
- `ForcedTransactionGateway` is deployed by the upstream rollup deploy script but
  is not used by this quickstart.
- Generated genesis/rendered files are ignored; only templates should be
  committed.
