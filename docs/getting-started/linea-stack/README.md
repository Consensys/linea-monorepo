# Lineth Stack — quickstart (Sepolia L1)

Run a local Linea L2 stack that uses Sepolia as its L1. The quickstart deploys
the required L1 contracts to Sepolia from your funded deployer account, starts a
local L2 chain, runs coordinator/postman/prover services, and exposes a local L2
Blockscout explorer. By default, the stack runs in dev-proof mode: proofs are
fast and dummy. Partial proving is available for real validation, but requires
more RAM and time. This README covers both modes.

This quickstart is for local development and evaluation only. It is not a
production deployment model.

```
╔══════════════════════════════════════════════════════════════════════════════╗
║                                                                              ║
║   ⚠  DEV ONLY — DO NOT REUSE ON MAINNET ⚠                                    ║
║                                                                              ║
║   This stack keeps committed local-dev material for node identity, Maru      ║
║   consensus, and Web3signer mTLS. It also accepts your Sepolia-funded L1     ║
║   deployer key from `.env`. Treat committed local keys and generated volume  ║
║   keys as dev-only and already unsuitable for production.                    ║
║                                                                              ║
║   What's committed to the repo (local dev only):                             ║
║     • config/l2/sequencer/key                    (sequencer P2P identity)    ║
║     • config/l2/maru/private-key                 (Maru consensus signer)     ║
║     • web3signer/coordinator/sequencer mTLS material                         ║
║                                                                              ║
║   What's user-supplied (your responsibility):                                ║
║     • L1_RPC_URL: a Sepolia HTTPS RPC (local node or paid-tier provider)     ║
║     • L1_DEPLOYER_PRIVATE_KEY: your Sepolia-funded deployer key              ║
║       It deploys/admins L1 contracts and funds generated runtime signers.    ║
║       The deployer address MUST be Sepolia-funded.                           ║
║                                                                              ║
║   At boot, account-setup writes fresh runtime keys into a docker volume:     ║
║   L1 blob, L1 finalization, L1 postman, L2 deployer, L2 anchorer, and L2     ║
║   postman. They live in `/shared/runtime-keys.env` and                       ║
║   `/shared/web3signer-keys/`, wiped by `docker compose down -v`.             ║
║                                                                              ║
║   Full inventory: see config/DEV-KEYS-INVENTORY.md.                          ║
║                                                                              ║
╚══════════════════════════════════════════════════════════════════════════════╝
```

:::note Apple Silicon
The Linea prover image is `linux/amd64` only and runs under Rosetta on M-series,
3-5x slower than native x86_64. For day-to-day work, keep
`PROVER_DEV_OVERRIDE=true` in `.env` (see §8) so the prover serves dummy proofs
in seconds regardless of architecture. The slowdown only matters when you turn
the override off for real partial-mode validation: first proof there is roughly
30 minutes on M-series vs. 5-10 minutes on native x86_64. Everything else in the
stack is multi-arch.
:::

## 1. What this is

A Docker Compose stack with Linea Besu sequencer, Maru consensus, L2 RPC
follower, Shomei state manager, coordinator, postman, web3signer, prover, and an
L2 Blockscout API backend and frontend explorer UI. All L1 traffic goes through
**your Sepolia RPC**.

What you get from `docker compose up`: a live local L2 chain, fresh Linea
contracts deployed to Sepolia from your address, and a coordinator/prover path
that can create default dev proofs and submit L1 blob/finalization transactions.

Use the default dev-proof mode for the first evaluation. Partial proving is
available for heavier validation runs, but it needs more Docker memory and more
time. Known limitations are listed near the end of this page.

## 2. Prerequisites

| Requirement              | Minimum                                                  |
|--------------------------|----------------------------------------------------------|
| RAM (dev-proof mode)     | 8 GB Docker Desktop minimum                              |
| RAM (partial-proof mode) | 30-32 GB assigned to Docker; 48 GB recommended           |
| Disk                     | ~30 GB free; leave more headroom for long traffic/prover runs |
| CPU                      | 8 cores                                                  |
| Docker                   | v24+                                                     |
| Compose                  | v2.19+                                                   |
| Sepolia RPC              | HTTPS endpoint (Infura / Alchemy / your own node)        |
| Sepolia ETH              | ~1 ETH recommended on the deployer; quickstart reserves 0.15 ETH for L1 blob submission, 0.15 ETH for L1 finalization, 0.05 ETH for L1 postman, and uses the rest for L1 deploy gas |

Faucets: [sepoliafaucet.com](https://sepoliafaucet.com) ·
[Alchemy Sepolia faucet](https://www.alchemy.com/faucets/ethereum-sepolia).

## 3. Setup

```bash
cd docs/getting-started/linea-stack
cp .env.example .env
$EDITOR .env
```

Required values in `.env`:

```bash
L1_RPC_URL=https://sepolia.infura.io/v3/<your-project-id>
L1_DEPLOYER_PRIVATE_KEY=0x<your-sepolia-funded-key>
```

Optional tuning knobs:

```bash
# tune generated L1 runtime signer top-ups if Sepolia gas spikes
# L1_ROLE_MIN_BALANCE_WEI=100000000000000000
# L1_ROLE_TOP_UP_WEI=150000000000000000
# L1_POSTMAN_MIN_BALANCE_WEI=20000000000000000
# L1_POSTMAN_TOP_UP_WEI=50000000000000000

# keep 1337 unless intentionally validating a different local L2 chain ID
# L2_CHAIN_ID=1337
```

Everything else has sensible defaults. Optional knobs (port collisions, DA
mode, postgres credentials) are commented in `.env.example`.

Check host-port collisions before boot:

```bash
./scripts/check-ports.sh
```

If a port is busy, either stop the conflicting local service or override the
matching `HOST_PORT_*` value in `.env`.

### Accounts and funding model

The only user-funded account is `L1_DEPLOYER_PRIVATE_KEY` on Sepolia. During
boot, `account-setup` generates fresh runtime keys for L1 blob submission, L1
finalization, L1 postman, L2 deployer, L2 message anchorer, and L2 postman.

L2 ETH is local genesis ETH, not bridged Sepolia ETH. The generated L2 deployer
is injected into the L2 genesis and funded with 90,000 ETH so it can deploy L2
contracts and pay local gas. The precomputed `L2MessageService` address is also
funded in genesis with 1,000,000,000 ETH so L1-to-L2 message claims can pay out
value on the local L2. After L2 boot, `deploy-contracts` tops up the generated
L2 anchorer and L2 postman from the generated L2 deployer.

`ERC20Example` is deployed on both Sepolia and the local L2 during boot. The L2
copy mints its initial supply to the generated L2 deployer, because that key is
the bootstrap/admin deployer for local L2 contracts. For demo activity,
`send-l2-erc20-transfer.sh` and `generate-l2-erc20-traffic.sh` create or reuse a
disposable traffic account in `/shared/demo-traffic.env`. The L2 deployer funds
that account once with small local L2 ETH and `ERC20Example`, then the traffic
account sends the visible ERC20 transfers.

## 4. Boot

```bash
docker compose --env-file versions.env --env-file .env --profile stack-partial-prover up -d
```

`stack-partial-prover` is the only profile. The profile name means "start the
full stack including the prover service"; whether proofs are dev or partial is
controlled by `PROVER_DEV_OVERRIDE`.

### Boot timeline

```
T+0:00  account-setup       → /shared/runtime-keys.env, /shared/web3signer-keys,
                              and /shared/addresses-precomputed.json
                              (queries Sepolia for chain ID + deployer nonce;
                              generates fresh runtime keys; precomputes only
                              boot-critical LineaRollupV8 + L2MessageService)
T+0:05  l2-genesis-init     → starts after account-setup completes; renders L2
                              genesis with only the generated L2
                              deployer funded, plus precomputed L2MessageService
                              pre-funded at 1B ETH; output lives in the
                              linea-stack-l2-genesis Docker volume
T+0:05  config-render       → starts after account-setup completes; writes
                              /rendered/{coordinator, maru, sequencer,
                              l2-node-besu, prover}-config with precomputed addrs
T+0:30  web3signer ready    (3 generated signer keys: L1 blob submission,
                              L1 finalization, L2 message anchoring)
T+1:00  sequencer healthy
T+1:30  maru + l2-node-besu + shomei healthy
T+2:00  deploy-contracts begins
        ├─ pre-flight (waits for Sepolia + Shomei reachable)
        ├─ pnpm install + Foundry install (cold ~3 min, warm ~30s)
        ├─ Step 1: deploy L1 LineaRollupV8 (~30s on Sepolia)
        ├─ Step 2: deploy L2 MessageService
        ├─ Step 3-4: deploy TokenBridge L1 + L2
        ├─ Step 5-6: deploy ERC20Example L1 + L2
        ├─ verify-or-die: LineaRollupV8 + L2MessageService match precompute
        ├─ capture all other deployed addresses into /shared/addresses.json
        ├─ fund generated L1/L2 runtime signers
        └─ patch /rendered/coordinator-config (state root + shnarf + deploy block)
T+5-8m  coordinator + postman + prover start
        postman reads generated signer keys and deployed addresses at startup
T+8m+   coordinator writes prover requests; dev prover writes fast responses
        (partial mode can spend many minutes on the first execution proof)
T+10m+  coordinator starts L1 blob/aggregation submissions once enough L2
        blocks have been conflated/proven
```

The dominating variables on real Sepolia are:
- Sepolia RPC latency / rate limits
- Sepolia gas-price spikes (deploys can take much longer when basefee is high)
- Image pull on first boot (~6 GB total)

Plan ~20-30 min for first-boot evidence on a cold Docker host in default dev
mode. In partial mode, the 2026-05-12 M-series validation with 30 GiB assigned
to Docker and `PROVER_GOMEMLIMIT=24GiB` reached first L1 finalization about 25
min after coordinator/prover startup. Do not call a run successful just because
containers are up: wait for `./scripts/status.sh` to show deployed addresses,
coordinator ports, prover responses, L1 blob submission, and a separate
`finalizeBlocks` tx that advances the rollup's `currentL2BlockNumber`. Use
`./scripts/links.sh` to print the current Sepolia and local L2 explorer links.

### Prover timing expectations

Default mode is dev proving, so proofs are dummy/fast and the slow parts are
mostly Docker image pulls, contract deploys, Sepolia RPC latency, and L1
inclusion. Partial mode is different: execution proofs are real partial proofs,
while compression and aggregation remain dev-mode proofs in this quickstart.

Observed on 2026-05-12 with Docker Desktop set to 30 GiB and
`PROVER_GOMEMLIMIT=24GiB`:

| Milestone | Observed timing | What to watch |
|-----------|-----------------|---------------|
| First execution proof request | a few seconds after coordinator/prover start | `execution proof request generated` |
| Each 2-block execution proof | about 11-16 min after the prover picks it up | prover log: `The executor is about to run ... execution ...`, then `processing of file ... took N seconds` |
| Compression proof | usually sub-second to a few seconds after execution proof | `blob compression proof generated` |
| Aggregation proof | usually sub-second to a few seconds, but can sit behind a running execution proof | `aggregation proof generated` |
| First Sepolia blob tx | about 24 min after the first proof request in the observed run | `blobs submitted` |
| First Sepolia finalization tx | about 25 min after the first proof request in the observed run | `submitted aggregation`, then `finalization update` |
| Later finalized ranges | roughly one execution-proof duration per new 2-block range, plus Sepolia inclusion/polling | `rollup currentL2BlockNumber` in `./scripts/status.sh` |

Do not use local L2 block height as proof of L1 finality. Blockscout may show
new local L2 blocks immediately, while the rollup's Sepolia
`currentL2BlockNumber` only advances after a successful `finalizeBlocks` tx.

### L1 gas caps and Sepolia congestion

The coordinator uses dynamic gas-price caps for L1 submissions. The defaults
are intentionally bounded for a quickstart:

| Path | Default cap |
|------|-------------|
| Blob transaction max fee per gas | 100 gwei |
| Blob transaction max fee per blob gas | 100 gwei |
| Blob transaction priority fee | 20 gwei |
| Finalization transaction max fee per gas | 200 gwei |
| Finalization transaction priority fee | 40 gwei |

If Sepolia execution gas or blob gas spikes above those caps, the prover can be
done while the coordinator waits or retries L1 submission. `status.sh` will show
proof/compression/aggregation responses, but no new blob tx or no new
`finalizeBlocks` tx. Coordinator logs may repeat `blobs to submit`, or show
messages such as `Estimated miner tip ... exceeds configured max fee`,
`replacement transaction underpriced`, or gas-price cap details.

Inspect the L1 submission path with:

```bash
docker logs --since 15m coordinator | \
  grep -Ei 'blobs to submit|blobs submitted|submitted aggregation|Estimated miner tip|underpriced|gasPriceCaps|error'
```

Adjust before boot in `.env` when Sepolia is congested:

```bash
L1_BLOB_MAX_FEE_PER_GAS_CAP_WEI=150000000000
L1_BLOB_MAX_FEE_PER_BLOB_GAS_CAP_WEI=150000000000
L1_BLOB_MAX_PRIORITY_FEE_PER_GAS_CAP_WEI=30000000000
L1_FINALIZATION_MAX_FEE_PER_GAS_CAP_WEI=300000000000
L1_FINALIZATION_MAX_PRIORITY_FEE_PER_GAS_CAP_WEI=60000000000
```

Values are wei. Raising caps can spend more Sepolia ETH from the generated L1
runtime signers; the deployer tops those signers up during boot.

## 5. Endpoints

| Service | URL | Note |
|---------|-----|------|
| L2 RPC (HTTP) | http://localhost:8745 | end-user RPC (l2-node-besu); use this from wallets/SDKs |
| L2 RPC (WS)   | ws://localhost:8746  | |
| L2 Blockscout UI | http://localhost:4001 | Frontend explorer for the local L2 chain. |
| L2 Blockscout API | http://localhost:4000/api/v2/blocks | Backend API used by the frontend. |
| Coordinator   | http://localhost:9545 | observability + JSON-RPC; use `./scripts/status.sh` to check whether `9545`/`9546` are listening. |
| Postman       | http://localhost:9090 | |
| Web3signer    | http://localhost:9000 | mTLS only — won't respond to plain HTTP |
| Maru          | http://localhost:8080 | observability/health |
| L1 explorer   | https://sepolia.etherscan.io | `https://sepolia.etherscan.io/address/<LineaRollupAddress>` |

Sequencer also runs RPC on `:8645` but that port is **internal-only by
convention** — connect dapps/wallets to `:8745`.

## 6. Verifying it works

You have a successful boot when `addresses.json` exists, the local L2 and
Blockscout are live, and the Sepolia rollup has both a blob submission and a
successful `finalizeBlocks` transaction that advances `currentL2BlockNumber`.

First boot starts the stack and deploys contracts. It does **not** send demo
traffic or bridge messages for you.

1. Print status and links:

```bash
./scripts/status.sh
./scripts/links.sh
```

`status.sh` should show:

- `addresses.json` present;
- coordinator ports `9545` and `9546` listening;
- prover request/response counts above zero;
- a blob tx under `DA only`;
- a separate finalization tx that advances `rollup currentL2BlockNumber`.

On Sepolia Etherscan, finalization may appear as method selector `0x755bc62f`
instead of a decoded name. That selector is `finalizeBlocks(bytes,uint256,tuple)`.

:::warning
Do not treat `Submit Blobs` as finalization. Blob submission publishes L2 batch
data/commitments for data availability; only a successful `finalizeBlocks` tx
advances the rollup's finalized L2 block/state on Sepolia.
:::

2. Open the local L2 explorer:

```text
http://localhost:4001
```

If the L2 block height stops advancing when there is no traffic, that is
expected. Maru is configured with `allow-empty-blocks = false`, so the sequencer
does not produce empty blocks just to keep the explorer moving. Send demo
traffic when you want visible Blockscout activity.

3. Start continuous ERC20 traffic for demos/evaluation:

```bash
./scripts/generate-l2-erc20-traffic.sh start
./scripts/generate-l2-erc20-traffic.sh logs
```

Default: one `ERC20Example.transfer(...)` every 2 seconds until stopped. The
first run creates/reuses the disposable demo traffic account, tops it up from
the L2 deployer if needed, and then sends transfers from that traffic account.

```bash
./scripts/generate-l2-erc20-traffic.sh stop
```

Useful knobs:

```bash
INTERVAL_SECONDS=1 ./scripts/generate-l2-erc20-traffic.sh start
MAX_TXS=100 ./scripts/generate-l2-erc20-traffic.sh start
```

4. Run the L1-to-L2 message smoke when you want to prove inbound message relay:

```bash
./scripts/smoke-bridge-message.sh
```

The smoke sends `sendMessage` on Sepolia, waits for Postman to claim on local
L2, verifies `MessageClaimed`, checks the recipient balance delta, and prints
both explorer links.

5. Run the L1-to-L2 ERC20 TokenBridge smoke when you want to prove a real
TokenBridge transfer:

```bash
./scripts/smoke-bridge-erc20-l1-to-l2.sh
```

The smoke mints `ERC20Example` to the configured Sepolia deployer, approves the
L1 TokenBridge, calls `bridgeToken(...)`, waits for Postman to claim on local
L2, and verifies the recipient's balance on the L2 bridged token contract. By
default it sends `1 ERC20` to the disposable demo traffic account. Run this
again whenever the disposable account needs more bridged ERC20 for the
withdrawal smoke.

6. Run the L2-to-L1 ERC20 TokenBridge smoke when you want to prove withdrawal
back to Sepolia:

```bash
./scripts/smoke-bridge-erc20-l2-to-l1.sh
```

This smoke expects the L1-to-L2 ERC20 smoke to have created and funded the L2
bridged token first. It approves the L2 TokenBridge, calls `bridgeToken(...)`
from the disposable demo traffic account, waits for coordinator/prover
finalization, claims on Sepolia if Postman has not already claimed, and
verifies the L1 `ERC20Example` recipient balance delta. It sends L2 bridged
ERC20 back to Sepolia; it is not an ETH withdrawal test.

7. Run the L2-to-L1 message smoke when you want to prove outbound message
finality and claiming:

```bash
./scripts/smoke-bridge-message-l2-to-l1.sh
```

The smoke sends a zero-value, zero-postman-fee `sendMessage` from the
disposable L2 traffic account, pays the L2 `minimumFeeInWei` required by
`L2MessageService`, waits until the coordinator finalizes/anchors that L2
message on Sepolia, claims it on L1 with the SDK
`getMessageProof`/`claimOnL1` path, and verifies the Sepolia receipt emitted
`MessageClaimed`.

Together, these smoke scripts cover generic message relay in both directions and
TokenBridge ERC20 deposit plus withdrawal. They are still **not** ETH withdrawal
tests.

The L2 traffic helpers mutate only the local quickstart L2 chain. The bridge
smokes spend real Sepolia gas. The L1-to-L2 smoke also sends Sepolia value into
the quickstart LineaRollup contract, then credits the local L2 recipient through
Postman. The generic L2-to-L1 message smoke sends a zero-value message, pays
only the L2 minimum message-service fee on send, and uses the generated L1
postman key for the manual L1 claim transaction. The ERC20 L2-to-L1 smoke burns
the disposable account's bridged ERC20 on L2 and releases `ERC20Example` on
Sepolia.

Optional one-shot local checks:

```bash
./scripts/send-l2-test-tx.sh
./scripts/send-l2-erc20-transfer.sh
```

## 7. Tearing down

```bash
# Stop, keep volumes — chaindata + addresses persist; next `up` is faster
docker compose --env-file versions.env --env-file .env --profile stack-partial-prover down

# Wipe everything — REQUIRED if you change L1_RPC_URL or L1_DEPLOYER_PRIVATE_KEY
# between runs (precomputed addresses + L2 genesis depend on these).
docker compose --env-file versions.env --env-file .env --profile stack-partial-prover down -v
```

## 8. Customisation

### Changing host ports

All host ports are env-driven. Set them in `.env`:

```bash
HOST_PORT_L2_RPC=8745
HOST_PORT_L2_BLOCKSCOUT=4000
HOST_PORT_L2_BLOCKSCOUT_FRONTEND=4001
# ... see .env.example for the full list
```

### Switching DA mode (Rollup ↔ Validium)

```bash
LINEA_COORDINATOR_DATA_AVAILABILITY=VALIDIUM \
  docker compose --env-file versions.env --env-file .env \
  --profile stack-partial-prover up -d
```

`deploy-contracts.sh` and the coordinator both respect the env var. Validium
deploys a different rollup contract (`ValidiumV2`) with a separate constructor
shape, but this knob is not part of the currently validated quickstart path.
Use `ROLLUP` unless you are intentionally debugging the Validium path. This is
also listed in §10 because the validated boot path is Rollup-only today.

### Switching prover mode (dev ↔ partial)

Default mode is **dev proofs everywhere**. Keep this for first boot, demos,
Blockscout, message smoke tests, and normal evaluation:

```bash
PROVER_DEV_OVERRIDE=true
```

Partial validation mode runs execution + invalidity in `partial` mode while
compression + aggregation stay in `dev` mode. Use a clean boot and set an
explicit memory limit:

```bash
PROVER_DEV_OVERRIDE=false
PROVER_GOMEMLIMIT=24GiB

docker compose --env-file versions.env --env-file .env \
  --profile stack-partial-prover down -v --remove-orphans
docker compose --env-file versions.env --env-file .env \
  --profile stack-partial-prover up -d
```

For a 30-32 GB Docker allocation, `PROVER_GOMEMLIMIT=24GiB` leaves room for the
rest of the stack. If Docker has 48 GB or more, `32GiB` is a better validation
setting.

Check which mode rendered:

```bash
docker run --rm -v linea-stack-rendered-config:/rendered:ro busybox \
  grep -E "^chain_id|^prover_mode|^is_allowed_circuit_id" /rendered/prover-config-partial.toml
```

If `PROVER_DEV_OVERRIDE=false` and `PROVER_GOMEMLIMIT` is missing,
`config-render` exits before the prover starts. That is intentional: a silent
fallback to `4GiB` wastes time and usually fails later with OOM-style proof
failure files.

### Pointing at a different L1

This quickstart is intentionally Sepolia-shaped. To target a different L1
(another testnet, your own devnet) you'd need to:
- Update timing tunables that assume 12s blocks (`coordinator-config.toml.template`'s
  `block-time`, `consistent-number-of-blocks-on-l1-to-wait`; postman's
  `L1_LISTENER_INTERVAL`)
- Verify the LineaRollup contract's expected verifier shape works on that L1
- Make sure `account-setup`'s nonce-offset assumptions hold (they're tied to
  the deploy script's tx sequence — see `scripts/account-setup.sh`)
- Keep `L2_CHAIN_ID` consistent everywhere. The quickstart templates derive
  Besu genesis, Maru genesis, prover public inputs, deploy metadata, and
  Blockscout config from the same value.

## 9. Troubleshooting

### "Just start over"

```bash
docker compose --env-file versions.env --env-file .env --profile stack-partial-prover down -v --remove-orphans
docker compose --env-file versions.env --env-file .env --profile stack-partial-prover up -d
```

`down -v` is critical — without `-v`, the shared config, rendered config,
chaindata, and generated L2 genesis volumes survive. The repo keeps only
`genesis-*.json.template`; rendered `genesis-besu.json`, `genesis-maru.json`,
and `fork-timestamp.txt` live in Docker volume `linea-stack-l2-genesis`.

### Common failures

| Symptom | Likely cause | Fix |
|---------|--------------|-----|
| `account-setup` exits with "L1 RPC not reachable" | `L1_RPC_URL` rate-limited or wrong | Try a different Sepolia RPC; check the URL with `cast chain-id --rpc-url $L1_RPC_URL` |
| `account-setup` exits with "could not extract deployers.l1" | malformed JSON output (Foundry version mismatch?) | See [bringup-notes.md](./bringup-notes.md) — `cast wallet address` flag drift is one possibility |
| deploy-contracts step 1 fails with "insufficient funds" | deployer's Sepolia balance too low | Top up via faucet; Sepolia gas spikes can blow through 0.5 ETH |
| deploy-contracts dies with `ADDRESS MISMATCH` | LineaRollupV8 or L2MessageService no longer lands at the boot-critical precomputed address | Usually means the deploy script changed and `account-setup.sh` needs a corresponding nonce/precompute update |
| Coordinator retries `linea_generateConflatedTracesToFileV2` with `Conflation not finished` on old block ranges | L2 ran far ahead while coordinator was down, beyond Besu's retained Bonsai history | Start over with `docker compose --env-file versions.env --env-file .env --profile stack-partial-prover down -v --remove-orphans`; this quickstart keeps a larger `bonsai-historical-block-limit` to make delayed first boots recoverable |
| Coordinator logs `address already reserved` during L1 blob/finalization submission | stale boot volume or old image/config still maps multiple jobs to the same signer | Fresh boot after the runtime-key cleanup should use distinct generated signer addresses; run `docker compose --env-file versions.env --env-file .env --profile stack-partial-prover down -v` before retesting |
| Aggregation/finalization occasionally reverts with a starting-root mismatch while catching up | coordinator retried an aggregation window after an earlier finalization tx succeeded | Known caveat; watch whether `finalization update` advances before calling it stuck |
| Coordinator restarts on first boot | usually a race against shomei's first-block trace; self-heals within ~30s | If it persists past 1 min, check `docker logs coordinator`; see fix log entries on web3signer mTLS |
| Web3signer mTLS handshake errors | known-clients fingerprint out of sync (only after regenerating one side) | Regenerate both sides or restore from git |
| `config-render` exits with `PROVER_GOMEMLIMIT must be set explicitly` | `PROVER_DEV_OVERRIDE=false` selected without a partial-prover memory limit | Set `PROVER_GOMEMLIMIT=24GiB` for a 30-32 GB Docker allocation, or `32GiB` for a larger validation machine, then clean boot |
| Prover execution proofs exit `137` or files get `.large.failure.code_137` | partial prover ran under too little Docker memory | Use default `PROVER_DEV_OVERRIDE=true`, or raise Docker memory substantially before partial-mode validation |
| Port collision | Another local service uses a required host port | Run `./scripts/check-ports.sh`, then stop the conflicting service or override the matching `HOST_PORT_*` in `.env` |

### Inspecting

```bash
# All services + health
docker compose --env-file versions.env --env-file .env --profile stack-partial-prover ps

# Logs (multi-tail)
docker compose --env-file versions.env --env-file .env --profile stack-partial-prover logs -f coordinator sequencer postman

# Inspect addresses.json (the contract handoff). deploy-contracts is a one-shot
# init container — `docker exec` won't work after it exits. Use a throwaway
# busybox bound to the shared volume:
docker run --rm -v linea-stack-shared-config:/shared:ro busybox \
  cat /shared/addresses.json

# Inspect precomputed addresses (account-setup output)
docker run --rm -v linea-stack-shared-config:/shared:ro busybox \
  cat /shared/addresses-precomputed.json

# Coordinator's rendered config (post-deploy patches). Coordinator IS long-running.
docker exec coordinator cat /rendered/coordinator-config.toml | head -40

# All deploy-step logs (one file per step)
docker run --rm -v linea-stack-shared-config:/shared:ro busybox \
  sh -c 'for f in /shared/deploy-logs/*.log; do echo "=== $f ==="; cat "$f"; done' | less

# Open a debug shell on the linea network
docker run --rm -it --network linea-stack_linea \
  weibeld/ubuntu-networking bash
```

### Known-issue catalog

[`bringup-notes.md`](./bringup-notes.md) tracks the fix history plus the
current caveats. Entries #1-#15 cover the original local-L1 boot, #16-#27 the
Sepolia migration phases, #28-#34 the first real-Sepolia boot fixes, and #35+
the coordinator/prover bring-up, security cleanups, and remaining validation
notes.

## 10. Known limitations

- **Monorepo-bound.** The deploy container bind-mounts `../../../contracts` so
  Hardhat tasks run against the real Linea contracts in this repository.
- **No Timelock, no Security Council.** L1 contracts are owned by your deployer
  EOA. Governance/upgrade flows are out of scope for this quickstart.
- **No full prover.** >700 GB RAM is out of scope.
- **No L1 Blockscout.** Use Sepolia Etherscan.
- **Partial proving is opt-in and resource-heavy.** It has been validated
  through L1 finalization, but the default path uses dev proofs for quick
  feedback.
- **Dev-proof mode is for fast evaluation only.** It does not produce real ZK
  proofs. Use partial mode when you need production-shaped proving behavior.
- **Validium mode is not part of the validated quickstart path.** The env knob
  exists for debugging, but the validated first boot uses `ROLLUP`.
- **Verifier setup is quickstart-only.** The current deploy path uses
  `IntegrationTestTrueVerifier`, not a production verifier configuration.
- **ETH withdrawal smoke is still missing.** The included TokenBridge smokes
  prove ERC20 L1-to-L2 deposit and ERC20 L2-to-L1 withdrawal. They do not yet
  prove ETH withdrawal behavior.
- **No one-shot verification orchestrator or CI smoke test.** Validation today
  is compose-config + manual first-boot-against-Sepolia. The local helper
  scripts are intentionally staged; a future `quickstart-verify.sh` can chain
  read-only checks, optional demo traffic, and the bridge smoke.
- **Most orchestration is still shell.** Bash remains the wrapper/orchestrator
  for boot and deploy scripts. Broader TypeScript migration is a follow-up once
  the team aligns on the final boot flow.

## 11. Reference

### Service list (`stack-partial-prover`)

```
account-setup            (init, foundry, runs first)
l2-genesis-init          (init, busybox)
config-render            (init, busybox)
web3signer               (consensys/web3signer)
sequencer                (consensys/linea-besu-package, role=SEQUENCER)
maru                     (consensys/maru, L2 consensus)
l2-node-besu             (consensys/linea-besu-package, role=RPC)
shomei                   (consensys/linea-shomei, ZK state manager)
coordinator              (consensys/linea-coordinator)
postman                  (consensys/linea-postman, L1↔L2 message relay)
prover                   (consensys/linea-prover, dev mode by default)
deploy-contracts         (init, node:24, runs Hardhat)
coordinator-pg           (postgres)
postman-pg               (postgres)
blockscout-l2-pg         (postgres)
l2-blockscout            (blockscout/blockscout, API backend)
l2-blockscout-frontend   (ghcr.io/blockscout/frontend, explorer UI)
```

> Prometheus/Grafana/Loki are not included in this quickstart. Use
> `docker logs <service>` and `./scripts/status.sh` for per-container output and
> milestone checks.

### File structure

```
docs/getting-started/linea-stack/
├── README.md                    ← you are here
├── bringup-notes.md            ← fix history, current caveats, validation notes
├── docker-compose.yml
├── versions.env                 ← pinned image tags
├── .env.example                 ← copy to .env, fill in L1 RPC + key
├── scripts/
│   ├── account-setup.sh         ← generates runtime keys + boot addresses
│   ├── deploy-contracts.sh      ← 6-step deploy + address capture
│   ├── aggregate-addresses.ts   ← writes addresses.json from deploy logs
│   ├── deployBridgedTokenAndTokenBridgeV1_1.ts ← TokenBridge deploy helper
│   ├── check-ports.sh           ← preflights host port collisions
│   ├── links.sh                 ← prints useful Sepolia + local explorer links
│   ├── status.sh                ← redacted boot status summary
│   ├── send-l2-test-tx.sh       ← sends tiny L2 ETH txs for Blockscout demos
│   ├── send-l2-erc20-transfer.sh ← sends a tiny L2 ERC20Example transfer
│   ├── generate-l2-erc20-traffic.sh ← runs continuous L2 ERC20Example traffic
│   ├── smoke-bridge-message.sh  ← sends and verifies L1-to-L2 message delivery
│   ├── smoke-bridge-erc20-l1-to-l2.sh ← sends and verifies an ERC20 TokenBridge deposit
│   ├── smoke-bridge-erc20-l2-to-l1.sh ← sends, claims, and verifies an ERC20 TokenBridge withdrawal
│   ├── smoke-bridge-message-l2-to-l1.sh ← sends, claims, and verifies L2-to-L1 message delivery
│   └── DEPLOY-ENV-CONTRACT.md   ← env vars per deploy step
└── config/
    ├── DEV-KEYS-INVENTORY.md    ← what's checked in vs runtime
    ├── explorer/
    ├── postgres/
    ├── web3signer/
    │   └── tls-files/           ← mTLS keystore + password + known-clients
    └── l2/
        ├── coordinator/
        ├── genesis-init/        ← genesis-besu.json.template + init.sh
        ├── l2-node-besu/
        ├── maru/
        ├── postman/
        ├── prover/
        ├── sequencer/
        └── shomei/
```

### Volumes

- `linea-stack-shared-config` — written by `account-setup` (runtime keys,
  boot-critical precomputed addresses, web3signer keystores) and
  `deploy-contracts` (addresses.json, deploy logs); read by config-render,
  coordinator, postman, prover, web3signer
- `linea-stack-l2-genesis` — written by `l2-genesis-init` (rendered Besu/Maru
  genesis plus fork timestamp); read by L2 services and deploy-contracts
- `linea-stack-rendered-config` — written by `config-render` (full render of 5
  templates) and `deploy-contracts` (in-place patch of coord-config); read by
  sequencer, maru, l2-node-besu, coordinator, prover
- `linea-stack-local-dev` — chaindata + prover state
- `linea-stack-logs` — shared log-output volume
- per-service postgres volumes
