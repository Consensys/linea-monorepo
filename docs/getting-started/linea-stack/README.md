# Linea Stack — quickstart (Sepolia L1)

> v0 of the "Streamlined Linea Stack deployment" feature. Local L2 stack
> pointed at user-supplied **Sepolia** as L1.

> **Status:** Sepolia migration validated end-to-end through Phase 5 against a
> real funded deployer (2026-05-07). All 6 contract steps deploy and verify on
> Sepolia; sequencer + maru + l2-node-besu + shomei + postman + prover all
> healthy. **One open issue from Phase 6**: the coordinator boots cleanly with
> the user's web3signer pubkey wired into all 3 signer slots, but its pipeline
> goes silent after `Coordinator app instantiated` — no conflation, no blob
> submissions to LineaRollup. Triage notes in
> [`first-boot-fixes.md`](./first-boot-fixes.md) under "Phase 6 (open)". User
> deploys (deployer-driven Sepolia txs) work; coordinator-driven L1 txs (blob
> submission, finalization, message anchoring) don't yet.
>
> If you boot this and hit something
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

> **Apple Silicon note.** The Linea prover image is `linux/amd64` only — runs
> under Rosetta on M-series, 3–5× slower than native x86_64. **Mitigation for
> day-to-day work**: keep `PROVER_DEV_OVERRIDE=true` in `.env` (see §8) so the
> prover serves dummy proofs in seconds regardless of arch. The 3–5× penalty
> only matters when you flip the override OFF for real partial-mode validation
> — first proof there is ~30 min on M-series vs ~5–10 min on native x86_64.
> Everything else in the stack is multi-arch.

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
| Coordinator   | http://localhost:9545 | observability + JSON-RPC. NOTE: container's healthcheck shells `curl` which isn't installed in the linea-coordinator image, so `docker compose ps` always shows "unhealthy" — this is cosmetic. The port may also not bind reliably from host while the Phase-6 coordinator-stuck issue is open. |
| Postman       | http://localhost:9090 | |
| Web3signer    | http://localhost:9000 | mTLS only — won't respond to plain HTTP |
| Maru          | http://localhost:8080 | observability/health |
| L1 explorer   | https://sepolia.etherscan.io | `https://sepolia.etherscan.io/address/<LineaRollupAddress>` |

Sequencer also runs RPC on `:8645` but that port is **internal-only by
convention** — connect dapps/wallets to `:8745`.

## 6. Verifying it works

`deploy-contracts` exits cleanly when its 6 steps finish — `docker exec` won't
work against an exited container. Use a throwaway busybox bound to the shared
volume to read its outputs:

```bash
# Service status (ignore "unhealthy" on coordinator — see endpoint table note)
docker compose --profile stack-partial-prover ps

# Final contract addresses (post-deploy)
docker run --rm -v linea-stack-shared-config:/shared:ro busybox \
  cat /shared/addresses.json | head -40

# Pre-computed addresses (account-setup output, written before any deploy)
docker run --rm -v linea-stack-shared-config:/shared:ro busybox \
  cat /shared/addresses-precomputed.json

# Per-step deploy logs (one file per step, persists across container restarts)
docker run --rm -v linea-stack-shared-config:/shared:ro busybox \
  ls /shared/deploy-logs/

# L2 producing blocks?
curl -s -X POST -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' \
  http://localhost:8745

# L2 Blockscout indexing?
curl -s http://localhost:4000/api/v2/main-page/blocks | head -c 200

# LineaRollup deployed on Sepolia?
docker run --rm -v linea-stack-shared-config:/shared:ro busybox \
  sh -c 'sed -nE "s/.*\"LineaRollupV8\":[[:space:]]*\"(0x[a-fA-F0-9]{40})\".*/\1/p" /shared/addresses.json | head -1'
# Then open: https://sepolia.etherscan.io/address/<that-address>
```

### First L1→L2 message

```bash
# 1. Read the LineaRollup address from the shared volume
LINEA_ROLLUP=$(docker run --rm -v linea-stack-shared-config:/shared:ro busybox \
  sh -c 'sed -nE "s/.*\"LineaRollupV8\":[[:space:]]*\"(0x[a-fA-F0-9]{40})\".*/\1/p" /shared/addresses.json | head -1')

# 2. Send a message via Sepolia (replace recipient + load .env first)
source .env
cast send "$LINEA_ROLLUP" "sendMessage(address,uint256,bytes)" \
  0xRecipientOnL2 0 0x \
  --value 0.01ether \
  --rpc-url "$L1_RPC_URL" \
  --private-key "$L1_DEPLOYER_PRIVATE_KEY"

# 3. Watch postman pick it up
docker logs -f postman | grep -i "MessageSent\|claimed"
```

> ⚠ **Phase-6 caveat (2026-05-07).** End-to-end finalisation requires the
> coordinator to conflate L2 blocks → submit blobs to LineaRollup → finalise.
> The coordinator currently goes silent after startup (see status banner at
> top); blob submission is **not** firing. Postman will still observe the L1
> `MessageSent` event and attempt the L2-side anchor, but the corresponding
> L1-side finalisation won't happen until Phase 6 is closed out. Until then
> the L1→L2 walkthrough above succeeds at step 2 (the deployer-driven L1 tx
> lands on Sepolia) but step 3 won't show a "claimed" log.
>
> When Phase 6 closes: end-to-end finalisation takes one full coordinator
> conflation + dummy-proof cycle. Typical (target): 5–15 min steady state on
> x86_64; longer on Apple Silicon for the first proof.

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

### Switching prover mode (partial ↔ all-dev for fast iteration)

The deliverable shape is **partial prover** — execution + invalidity in `partial`
mode, data_availability + aggregation in `dev` mode, aggregation gate at
`is_allowed_circuit_id = 483`. That's what `config/l2/prover/prover-config-partial.toml.template`
ships and what gets rendered into the live volume on a fresh `up -d` after
`down -v`.

For fast iteration (dummy proofs everywhere, ~minutes per cycle vs ~tens-of-minutes
for partial proving), set one env var in `.env`:

```bash
PROVER_DEV_OVERRIDE=true
```

Then recreate the rendered config and bounce the prover:

```bash
docker compose --env-file versions.env --env-file .env \
  --profile stack-partial-prover up -d --force-recreate config-render prover
```

`config-render` post-patches the rendered prover config in the volume — flips
the 4 prover_mode lines to `dev` and the aggregation bitmask to `963`. The
**template on disk stays partial** (deliverable correct), only the rendered
file in the volume is patched. Verify:

```bash
docker run --rm -v linea-stack-rendered-config:/rendered:ro busybox \
  grep -E "^prover_mode|^is_allowed_circuit_id" /rendered/prover-config-partial.toml
# Expected with PROVER_DEV_OVERRIDE=true:
#   prover_mode = "dev" × 4
#   is_allowed_circuit_id = 963
```

To go back to partial, unset `PROVER_DEV_OVERRIDE` (or set `=false`) in `.env`
and re-run the same `up -d --force-recreate config-render prover` command.

> ⚠ Don't commit `PROVER_DEV_OVERRIDE=true` to your `.env.example` or any
> shared config. The profile is named `stack-partial-prover` because the
> deliverable is partial — the dev override is a local-iteration knob only.

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

[`first-boot-fixes.md`](./first-boot-fixes.md) tracks every fix applied during
scaffold development plus carry-overs into Phase 5 (real Sepolia validation)
and the open Phase 6 work. Entries #1–#15 cover the original local-L1 boot,
#16–#27 the Sepolia migration phases (incl. prover all-dev pre-flight), #28–#34
the Phase 5 first-boot fixes against real Sepolia (nonce-offset bugs in the
upstream deploy scripts, idempotency for retry cycles, prover kzgsrs mount),
and #35 the Phase 6 web3signer pubkey alignment. The "Phase 6 (open)" section
documents the coordinator-stuck-after-startup symptom + hypotheses for next
session.

## 10. What's NOT yet in v0

- **No Timelock, no Security Council.** L1 contracts are owned by your
  deployer EOA. Governance/upgrade flows = v1.
- **No full prover.** >700 GB RAM is out of scope.
- **No L1 Blockscout.** Use Sepolia Etherscan.
- **No CI smoke test.** Validation today is compose-config + render-pipeline
  level + manual first-boot-against-Sepolia (Phase 5, 2026-05-07). A scripted
  end-to-end smoke test is Phase-7+ once the coordinator pipeline is closed out.
- **No working coordinator-driven L1 finalisation yet.** See status banner.
  Deploys land on Sepolia; coordinator-driven txs (blob submission, finalisation,
  message anchoring) are blocked on the Phase-6 coordinator-stuck investigation.

## 11. Reference

### Service list (17 services on `stack-partial-prover`)

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
```

> Observability stack (prometheus / loki / promtail / grafana) was dropped in
> Phase-7 — out of scope for v0 deliverable. Use `docker logs <service>` for
> per-container output.

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
- `linea-logs` — kept as a shared log-output volume; was previously tailed
  by promtail (dropped Phase-7 with the rest of the observability stack).
- per-service postgres volumes
