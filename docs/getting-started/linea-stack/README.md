# Linea Stack — quickstart (Sepolia L1)

> v0 of the "Streamlined Linea Stack deployment" feature. Local L2 stack
> pointed at user-supplied **Sepolia** as L1.

> **Status:** Sepolia migration complete through Phase 4 (timing tunables +
> this README). Phase 5 — first-boot validation against a real funded Sepolia
> deployer key — has not yet run. If you boot this and hit something
> [`first-boot-fixes.md`](./first-boot-fixes.md) doesn't already cover, append
> what you learn there so the next person doesn't.

```
╔══════════════════════════════════════════════════════════════════════════════╗
║                                                                              ║
║   ⚠  DEV ONLY — DO NOT REUSE ON MAINNET ⚠                                    ║
║                                                                              ║
║   This stack ships pre-baked dev secrets for L2 (we own L2 genesis) and      ║
║   accepts your Sepolia-funded L1 deployer key from `.env`. The L2 secrets    ║
║   listed below are public knowledge — anyone reading this repo has them.     ║
║   Treat any L2 address derived from any L2 key here as already compromised.  ║
║                                                                              ║
║   What's committed to the repo (L2 only):                                    ║
║     • config/l2/sequencer/key                    (sequencer P2P identity)    ║
║     • config/l2/maru/private-key                 (Maru consensus signer)     ║
║     • L2_SIGNER_PRIVATE_KEY in config/l2/postman/env                         ║
║     • L2 deployer key + L2 dev EOAs in scripts + genesis-besu.json.template  ║
║     • web3signer mTLS material (TLS only — not Sepolia-funds-controlling)    ║
║                                                                              ║
║   What's user-supplied (your responsibility):                                ║
║     • L1_RPC_URL                — your Sepolia HTTPS RPC                     ║
║     • L1_DEPLOYER_PRIVATE_KEY   — your Sepolia-funded deployer key.          ║
║       This single key is used for ALL L1 roles (Option A): contract          ║
║       deployer, security council, rollup operators, blob/data submitter,     ║
║       aggregation/finalization submitter, message anchorer, postman L1       ║
║       signer. Address derived from this key MUST be Sepolia-funded.          ║
║                                                                              ║
║   At boot, account-setup writes web3signer keystores into a docker volume    ║
║   from your L1 key. They live in `/shared/web3signer-keys/` inside the       ║
║   stack and are wiped by `docker compose down -v`.                           ║
║                                                                              ║
║   Full inventory: see config/DEV-KEYS-INVENTORY.md.                          ║
║                                                                              ║
╚══════════════════════════════════════════════════════════════════════════════╝
```

> **Apple Silicon note.** The Linea prover image is `linux/amd64` only. On
> M-series Macs it runs under Rosetta and is 3–5× slower than native x86_64.
> First proof: 30+ min on M-series, 5–10 min on native x86_64. Everything
> else in the stack is multi-arch.

## 1. What this is

A self-contained Docker Compose stack: full Linea L2 (Linea Besu sequencer +
Maru consensus + L2 RPC follower + Shomei state manager), the coordinator,
postman, web3signer, the partial prover, an L2 Blockscout explorer, and a
Prometheus / Grafana / Loki observability bundle. All wired to use **your
Sepolia RPC** as L1.

What you get from `docker compose up`: a live L2 chain, fresh Linea contracts
deployed to Sepolia from your address, and a working L1↔L2 message round-trip
including the proving cycle.

What this is **NOT**:
- Not portable — bind-mounts `../../../contracts` from the linea-monorepo so
  Hardhat tasks run against the real contracts. v1 will package this.
- Not the full prover — partial prover only (full prover requires >700 GB
  RAM). Real proofs, but on a partial trace set.
- No L1 explorer — use [sepolia.etherscan.io](https://sepolia.etherscan.io)
  with the LineaRollup address from `addresses.json`.
- No Timelock, no Security Council. The L1 deployer key owns everything.
  Governance/upgrade flows are out of scope.

## 2. Prerequisites

| Requirement     | Minimum                                                  |
|-----------------|----------------------------------------------------------|
| RAM             | 32 GB (prover container has `GOMEMLIMIT=32GiB`)          |
| RAM (recommended) | 48 GB                                                  |
| Disk            | ~80 GB free (image pulls + chaindata + observability)    |
| CPU             | 8 cores                                                  |
| Docker          | v24+                                                     |
| Compose         | v2.19+                                                   |
| Sepolia RPC     | HTTPS endpoint (Infura / Alchemy / your own node)        |
| Sepolia ETH     | ~0.5 ETH on the deployer for the L1 deploys + first batch submissions |

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

Everything else has sensible defaults. Optional knobs (port collisions, DA
mode, postgres credentials) are commented in `.env.example`.

## 4. Boot

```bash
docker compose --env-file versions.env --env-file .env --profile stack-partial-prover up -d
```

`stack-partial-prover` is the only profile.

### Boot timeline

```
T+0:00  account-setup       → /shared/addresses-precomputed.json + web3signer keys
                              (queries Sepolia for your deployer's nonce; computes
                              all 6 L1 + 4 L2 contract addresses deterministically)
T+0:00  l2-genesis-init     → renders L2 genesis with the precomputed
                              L2MessageService address pre-funded at 1B ETH
T+0:00  config-render       → writes /rendered/{coordinator, maru, sequencer,
                              l2-node-besu, prover}-config with precomputed addrs
T+0:30  web3signer ready    (3 L1 signers + 1 L2 liveness)
T+1:00  sequencer healthy
T+1:30  maru + l2-node-besu + shomei healthy
T+2:00  deploy-contracts begins
        ├─ pre-flight (waits for Sepolia + Shomei reachable)
        ├─ pnpm install + Foundry install (cold ~3 min, warm ~30s)
        ├─ Step 1: deploy L1 LineaRollupV8 (~30s on Sepolia)
        ├─ Step 2: deploy L2 MessageService
        ├─ Step 3-4: deploy TokenBridge L1 + L2
        ├─ Step 5-6: deploy TestERC20 L1 + L2
        ├─ verify-or-die: each deployed addr must equal precomputed
        ├─ patch /rendered/coordinator-config (state root + shnarf + deploy block)
        └─ patch postman/env (contract addresses + L1 deploy block)
T+5-8m  coordinator + postman + prover start
        post-deploy-restart cycles postman (re-reads patched env)
T+15m+  first proof submitted to Sepolia (varies; longer on Apple Silicon)
```

The dominating variables on real Sepolia are:
- Sepolia RPC latency / rate limits
- Sepolia gas-price spikes (deploys can take much longer when basefee is high)
- Image pull on first boot (~6 GB total)

Plan ~30 min for the first full cycle including a finalised proof.

## 5. Endpoints

| Service | URL | Note |
|---------|-----|------|
| L2 RPC (HTTP) | http://localhost:8745 | end-user RPC (l2-node-besu); use this from wallets/SDKs |
| L2 RPC (WS)   | ws://localhost:8746  | |
| L2 Blockscout API | http://localhost:4000/api/v2/blocks | Frontend `/` returns 404 — Blockscout 7.x splits the UI into a separate frontend container that this scaffold doesn't deploy. JSON API works. |
| Coordinator   | http://localhost:9545 | observability + JSON-RPC; `/health` returns 200 once healthy |
| Postman       | http://localhost:9090 | |
| Web3signer    | http://localhost:9000 | mTLS only — won't respond to plain HTTP |
| Maru          | http://localhost:8080 | observability/health |
| Grafana       | http://localhost:3001 | dashboards: "Linea / L2 Health" |
| Prometheus    | http://localhost:9091 | |
| L1 explorer   | https://sepolia.etherscan.io | `https://sepolia.etherscan.io/address/<LineaRollupAddress>` |

Sequencer also runs RPC on `:8645` but that port is **internal-only by
convention** — connect dapps/wallets to `:8745`.

## 6. Verifying it works

```bash
# Service health
docker compose --profile stack-partial-prover ps

# Contract addresses (post-deploy)
docker exec deploy-contracts cat /shared/addresses.json | head -40

# Coordinator alive?
curl -fsS http://localhost:9545/health && echo OK

# L2 producing blocks?
curl -s -X POST -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' \
  http://localhost:8745

# L2 Blockscout indexing?
curl -s http://localhost:4000/api/v2/main-page/blocks | head -c 200

# LineaRollup deployed on Sepolia?
docker exec deploy-contracts node -e \
  'const j=require("/shared/addresses.json");console.log(j.l1.LineaRollupV8)'
# Then open: https://sepolia.etherscan.io/address/<that-address>
```

### First L1→L2 message

```bash
# 1. Read the LineaRollup address
LINEA_ROLLUP=$(docker exec deploy-contracts node -e \
  'console.log(require("/shared/addresses.json").l1.LineaRollupV8)')

# 2. Send a message via Sepolia (replace KEY + recipient)
cast send "$LINEA_ROLLUP" "sendMessage(address,uint256,bytes)" \
  0xRecipientOnL2 0 0x \
  --value 0.01ether \
  --rpc-url "$L1_RPC_URL" \
  --private-key "$L1_DEPLOYER_PRIVATE_KEY"

# 3. Watch postman pick it up
docker logs -f postman | grep -i "MessageSent\|claimed"
```

End-to-end finalisation (proof submitted to L1, message marked claimable on L2)
takes one full coordinator conflation + proof cycle. Typical: 5–15 min steady
state on x86_64; longer on Apple Silicon for the first proof.

## 7. Tearing down

```bash
# Stop, keep volumes — chaindata + addresses persist; next `up` is faster
docker compose --profile stack-partial-prover down

# Wipe everything — REQUIRED if you change L1_RPC_URL or L1_DEPLOYER_PRIVATE_KEY
# between runs (precomputed addresses + L2 genesis depend on these).
docker compose --profile stack-partial-prover down -v
```

## 8. Customisation

### Changing host ports

All host ports are env-driven. Set them in `.env`:

```bash
HOST_PORT_L2_RPC=8745
HOST_PORT_L2_BLOCKSCOUT=4000
HOST_PORT_GRAFANA=3001
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
shape. The verify-or-die check in deploy-contracts skips for the validium
variant (precomputed JSON tracks the rollup variant).

### Pointing at a different L1

This scaffold is intentionally Sepolia-shaped. To target a different L1
(another testnet, your own devnet) you'd need to:
- Update timing tunables that assume 12s blocks (`coordinator-config.toml.template`'s
  `block-time`, `consistent-number-of-blocks-on-l1-to-wait`; postman's
  `L1_LISTENER_INTERVAL`)
- Verify the LineaRollup contract's expected verifier shape works on that L1
- Make sure `account-setup`'s nonce-offset assumptions hold (they're tied to
  the deploy script's tx sequence — see `scripts/account-setup.sh`)

## 9. Troubleshooting

### "Just start over"

```bash
docker compose down -v --remove-orphans
docker volume prune -f
docker compose --env-file versions.env --env-file .env --profile stack-partial-prover up -d
```

`down -v` is critical — without `-v`, the genesis volumes survive and the next
boot will reuse stale precomputed addresses against a deployer whose Sepolia
nonce has advanced.

### Common failures

| Symptom | Likely cause | Fix |
|---------|--------------|-----|
| `account-setup` exits with "L1 RPC not reachable" | `L1_RPC_URL` rate-limited or wrong | Try a different Sepolia RPC; check the URL with `cast chain-id --rpc-url $L1_RPC_URL` |
| `account-setup` exits with "could not extract deployers.l1" | malformed JSON output (Foundry version mismatch?) | See [first-boot-fixes.md](./first-boot-fixes.md) — `cast wallet address` flag drift is one possibility |
| deploy-contracts step 1 fails with "insufficient funds" | deployer's Sepolia balance too low | Top up via faucet; Sepolia gas spikes can blow through 0.5 ETH |
| deploy-contracts dies with `ADDRESS MISMATCH` | deploy script's contract sequence drifted from `account-setup.sh`'s nonce offsets | See fix log #22; usually means the deploy script changed and `account-setup.sh` needs a corresponding update |
| Coordinator restarts on first boot | usually a race against shomei's first-block trace; self-heals within ~30s | If it persists past 1 min, check `docker logs coordinator`; see fix log entries on web3signer mTLS |
| Web3signer mTLS handshake errors | known-clients fingerprint out of sync (only after regenerating one side) | Regenerate both sides or restore from git |
| Apple Silicon prover OOM | Docker Desktop memory cap < 48 GB | Raise the cap in Docker Desktop settings |
| Port collision (5432, 8745, 4000, 9000, 3001, 9091) | Another service uses the port | Override via `HOST_PORT_*` in `.env` |

### Inspecting

```bash
# All services + health
docker compose --profile stack-partial-prover ps

# Logs (multi-tail)
docker compose logs -f coordinator sequencer postman

# Inspect addresses.json (the contract handoff)
docker exec deploy-contracts cat /shared/addresses.json

# Inspect precomputed addresses (account-setup output)
docker exec deploy-contracts cat /shared/addresses-precomputed.json

# Coordinator's rendered config (post-deploy patches)
docker exec coordinator cat /rendered/coordinator-config.toml | head -40

# Open a debug shell on the linea network
docker run --rm -it --network linea-stack_linea \
  weibeld/ubuntu-networking bash
```

### Known-issue catalog

[`first-boot-fixes.md`](./first-boot-fixes.md) tracks every fix applied during
scaffold development plus carry-overs into Phase 5 (real Sepolia validation).
Read entries #19–#26 for the Sepolia migration phases.

## 10. What's NOT in v0

- **No Timelock, no Security Council.** L1 contracts are owned by your
  deployer EOA. Governance/upgrade flows = v1.
- **No full prover.** >700 GB RAM is out of scope.
- **No L1 Blockscout.** Use Sepolia Etherscan.
- **No Blockscout L2 frontend.** API only on `:4000` until a `blockscout-frontend`
  container is added.
- **No CI smoke test.** The validation has been compose-config + render-pipeline
  level; first real-Sepolia validation is Phase 5 of the migration.

## 11. Reference

### Service list (21 services on `stack-partial-prover`)

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
prover                   (consensys/linea-prover, partial mode)
deploy-contracts         (init, node:24, runs Hardhat)
post-deploy-restart      (init, docker:cli, restarts postman)
coordinator-pg           (postgres)
postman-pg               (postgres)
blockscout-l2-pg         (postgres)
l2-blockscout            (blockscout/blockscout, API only)
prometheus               (prom/prometheus)
loki                     (grafana/loki)
promtail                 (grafana/promtail)
grafana                  (grafana/grafana)
```

### File structure

```
docs/getting-started/linea-stack/
├── README.md                    ← you are here
├── SCAFFOLD-PLAN.md             ← historical planning doc (pre-Sepolia)
├── first-boot-fixes.md          ← every fix + Phase carry-overs
├── docker-compose.yml
├── versions.env                 ← pinned image tags
├── .env.example                 ← copy to .env, fill in L1 RPC + key
├── scripts/
│   ├── account-setup.sh         ← derives addresses + writes keystores
│   ├── deploy-contracts.sh      ← 6-step deploy + verify-or-die
│   └── DEPLOY-ENV-CONTRACT.md   ← env vars per deploy step
└── config/
    ├── DEV-KEYS-INVENTORY.md    ← what's checked in vs runtime
    ├── explorer/
    ├── observability/
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

- `linea-shared-config` — written by `account-setup` (precomputed addresses,
  web3signer keystores) and `deploy-contracts` (addresses.json, deploy logs);
  read by config-render, coordinator, postman, prover, web3signer
- `linea-rendered-config` — written by `config-render` (full render of 5
  templates) and `deploy-contracts` (in-place patch of coord-config); read by
  sequencer, maru, l2-node-besu, coordinator, prover
- `linea-local-dev` — chaindata + prover state
- `linea-logs` — promtail tails this for Loki ingestion
- per-service postgres volumes
