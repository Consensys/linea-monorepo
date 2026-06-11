# Running two Linea L2s on one L1 (for the Bermuda cross-chain DvP demo)

This is the Linea Stack Quickstart, extended to run **two real Linea L2s sharing
one local L1** — the topology your DvP demo needs. It replaces the two Anvil L2s
+ Anvil L1 with a real Linea stack: real sequencer, prover, canonical bridge
(L2MessageService + postman), Shomei, and real L1 finality.

Branch: `feat/dual-l2-bermuda-interop`
Package: `docs/getting-started/linea-stack`

## Requirements

- Docker 24+ and Docker Compose v2.19+
- 12–16 GB free to Docker for the two-instance topology (~8 GB runs a single
  stack), ~30 GB disk
- Apple Silicon is fine in dev-proof mode (default here)

## Run it

```bash
cd docs/getting-started/linea-stack

# Instance 1 — owns the local L1 + L2-A. Plain local-L1 mode.
printf 'L1_MODE=local\nPROVER_DEV_OVERRIDE=true\n' > .env
./scripts/start.sh --tail
# wait for "first L1 finalization observed"

# Instance 2 — L2-only, attaches to instance 1's L1.
mkdir -p instances
cp profiles/instance-2.env.example instances/i2.env
LINETH_ENV_FILE=instances/i2.env ./scripts/start.sh --tail
```

Status / reset per instance:

```bash
./scripts/status.sh                                  # instance 1
LINETH_ENV_FILE=instances/i2.env ./scripts/status.sh # instance 2
LINETH_ENV_FILE=instances/i2.env ./scripts/reset.sh  # instance 2 FIRST
./scripts/reset.sh                                   # instance 1
```

Reset instance 2 before instance 1: resetting the owner tears down the shared
L1 under any attached instance.

The local L2s only build blocks when there are transactions, so with no traffic
the finalized block stops advancing — that's idle, not broken. Nudge a chain
with:

```bash
COUNT=6 ./scripts/traffic-generation/send-l2-test-tx.sh                                  # L2-A
LINETH_ENV_FILE=instances/i2.env COUNT=6 ./scripts/traffic-generation/send-l2-test-tx.sh # L2-B
```

One-shot check (boots both from clean, asserts one shared L1, no collisions,
two distinct rollups, both finalizing, then tears down):

```bash
./scripts/verify-dual-l2.sh
```

## Endpoints

| | Shared L1 | L2-A (instance 1) | L2-B (instance 2) |
|---|---|---|---|
| Chain ID | 31648428 | 1337 | 1338 |
| RPC (host) | http://localhost:8445 | http://localhost:8745 | http://localhost:9745 |
| Shomei (`linea_getProof`) | — | http://localhost:8998 | http://localhost:9998 |
| Blockscout | — | http://localhost:4001 | http://localhost:5101 |

Inside Docker, services reach the L1 at `http://l1-el-node:8545` on both
instances (instance 2 joins the owner's L1 network via the attach overlay).

## What you need for wiring the 7888 layer

Your `broadcaster` lib already ships the real Linea adapters — `LineaBuffer` /
`LineaPusher` and `provers/linea/*` (SMT + MiMC). They replace `DirectBuffer` +
the `anchor-daemon` and target real Linea contracts:

- Each L2 deploys its **own LineaRollup + L2MessageService** to the shared L1.
  Addresses are written per instance to:
  - `artifacts/deployments/addresses.json` (L2-A)
  - `artifacts-i2/deployments/addresses.json` (L2-B)
- `linea_getProof` is served by Shomei (ports above) — confirmed live for SMT
  account proofs at finalized blocks; storage proofs use the same call (not
  explicitly exercised yet).
- Relay rides the real L2MessageService + the postman service (already running),
  not a manual pusher.

## Two things to know

- **Finality, not instant.** A cross-chain proof can only be made after the
  source L2's state root is **finalized + anchored on L1** — order of minutes,
  vs Anvil's 2s. You prove against finalized state-root block numbers, not every
  block. This is the real cost of real interop.
- **Test with the dev prover (the default here).** The prover's job is proving
  the L2 state transition up to L1. For the demo we run it in dev mode, which
  still commits a **real finalized state root** to the rollup — and that
  finalized root is all your cross-chain prover needs. So you get real
  anchoring without paying for full validity proving, and the whole thing fits
  on a laptop. Partial-proving mode exists if you ever want real (much heavier)
  proving; not needed for the demo.

## Scope / status

- Phase 1 (this branch): two L2s on one L1, both finalizing. Done + verified.
- Built config-driven — a 3rd chain is just another `instances/*.env`. Only the
  two-chain pair is wired and tested.
- Next (joint): swap `DirectBuffer`/`anchor-daemon` for the Linea adapters and
  run your pools + keeper against the two real L2s.

Questions: ping Moris (Linea infra / quickstart).
