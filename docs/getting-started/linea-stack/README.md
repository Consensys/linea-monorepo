# Linea Stack — quickstart (Sepolia L1)

> v0 of the "Streamlined Linea Stack deployment" feature. Local L2 stack
> pointed at user-supplied **Sepolia** as L1.

> **Status:** Fresh Sepolia boot validated on 2026-05-11 with dev proofs.
> Contracts deploy on Sepolia, the L2 chain starts, coordinator binds `9545` /
> `9546`, the prover file pipeline writes requests and responses, Blockscout UI
> serves the local L2, and coordinator L1 blob/aggregation/finalization
> submissions have been observed on Sepolia.
>
> This is still a bring-up checkpoint, not a final "done" stamp: partial-prover
> validation and a real bridge/message smoke test remain the next gates.
>
> If you boot this and hit something
> [`bringup-notes.md`](./bringup-notes.md) doesn't already cover, append what you
> learn there so the next person doesn't have to rediscover it from logs.

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

> **Apple Silicon note.** The Linea prover image is `linux/amd64` only — runs
> under Rosetta on M-series, 3–5× slower than native x86_64. **Mitigation for
> day-to-day work**: keep `PROVER_DEV_OVERRIDE=true` in `.env` (see §8) so the
> prover serves dummy proofs in seconds regardless of arch. The 3–5× penalty
> only matters when you flip the override OFF for real partial-mode validation
> — first proof there is ~30 min on M-series vs ~5–10 min on native x86_64.
> Everything else in the stack is multi-arch.

## 1. What this is

A Docker Compose stack with Linea Besu sequencer, Maru consensus, L2 RPC
follower, Shomei state manager, coordinator, postman, web3signer, prover, and an
L2 Blockscout API backend and frontend explorer UI. All L1 traffic goes through **your Sepolia RPC**.

What you get from `docker compose up`: a live local L2 chain, fresh Linea
contracts deployed to Sepolia from your address, and a coordinator/prover path
that can create dev proofs and submit L1 blob/aggregation transactions.

What this is **NOT**:
- Not portable — bind-mounts `../../../contracts` from the linea-monorepo so
  Hardhat tasks run against the real contracts. v1 will package this.
- Not the full prover — full proving requires >700 GB RAM. Quickstart defaults
  to dev proofs; partial proving is opt-in for validation runs.
- No L1 explorer — use [sepolia.etherscan.io](https://sepolia.etherscan.io)
  with the LineaRollup address from `addresses.json`.
- No Timelock, no Security Council. The L1 deployer key owns everything.
  Governance/upgrade flows are out of scope.

## 2. Prerequisites

| Requirement     | Minimum                                                  |
|-----------------|----------------------------------------------------------|
| RAM             | 8 GB Docker Desktop works for default dev proofs; partial mode needs much more |
| RAM (recommended) | 48 GB                                                  |
| Disk            | ~80 GB free (image pulls + chaindata + service DBs)      |
| CPU             | 8 cores                                                  |
| Docker          | v24+                                                     |
| Compose         | v2.19+                                                   |
| Sepolia RPC     | HTTPS endpoint (Infura / Alchemy / your own node)        |
| Sepolia ETH     | ~1 ETH recommended on the deployer; quickstart reserves 0.15 ETH for L1 blob submission, 0.15 ETH for L1 finalization, 0.05 ETH for L1 postman, and uses the rest for L1 deploy gas |

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
# optional: tune generated L1 runtime signer top-ups if Sepolia gas spikes
# L1_ROLE_MIN_BALANCE_WEI=100000000000000000
# L1_ROLE_TOP_UP_WEI=150000000000000000
# L1_POSTMAN_MIN_BALANCE_WEI=20000000000000000
# L1_POSTMAN_TOP_UP_WEI=50000000000000000
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
T+0:00  account-setup       → /shared/runtime-keys.env, /shared/web3signer-keys,
                              and /shared/addresses-precomputed.json
                              (queries Sepolia for chain ID + deployer nonce;
                              generates fresh runtime keys; precomputes only
                              boot-critical LineaRollupV8 + L2MessageService)
T+0:00  l2-genesis-init     → renders L2 genesis with only the generated L2
                              deployer funded, plus precomputed L2MessageService
                              pre-funded at 1B ETH
T+0:00  config-render       → writes /rendered/{coordinator, maru, sequencer,
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
T+8m+   coordinator writes prover requests; dev prover writes responses
T+10m+  coordinator starts L1 blob/aggregation submissions once enough L2
        blocks have been conflated/proven
```

The dominating variables on real Sepolia are:
- Sepolia RPC latency / rate limits
- Sepolia gas-price spikes (deploys can take much longer when basefee is high)
- Image pull on first boot (~6 GB total)

Plan ~20-30 min for first-boot evidence on a cold Docker host. Do not call a
run successful just because containers are up: wait for `./scripts/status.sh` to
show deployed addresses, coordinator ports, prover responses, and coordinator L1
submission logs. Use `./scripts/links.sh` to print the current Sepolia and local
L2 explorer links.

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

Start with the helper scripts:

```bash
# Redacted milestone view across Docker, deploy logs, coordinator, and prover
./scripts/status.sh

# Current Sepolia contract links and local L2 explorer links
./scripts/links.sh
```

A useful first-boot checkpoint has all of these signals:

- `addresses.json` exists and includes LineaRollupV8, L2MessageService,
  TokenBridge, and ERC20Example addresses.
- coordinator ports `9545` and `9546` are listening.
- prover request/response counts are non-zero.
- coordinator logs show L1 blob submission and aggregation/finalization activity.
- L2 Blockscout UI responds on `http://localhost:4001`.

Lower-level checks, when you need them:

```bash
# Service status
docker compose --env-file versions.env --env-file .env --profile stack-partial-prover ps

# Final contract addresses (post-deploy)
docker run --rm -v linea-stack-shared-config:/shared:ro busybox \
  cat /shared/addresses.json | head -60

# Boot-critical precomputed addresses (account-setup output)
docker run --rm -v linea-stack-shared-config:/shared:ro busybox \
  cat /shared/addresses-precomputed.json

# Per-step deploy logs (one file per step, persists across container restarts)
docker run --rm -v linea-stack-shared-config:/shared:ro busybox \
  ls /shared/deploy-logs/

# Current L2 block
curl -s -X POST -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' \
  http://localhost:8745

# L2 Blockscout indexing?
curl -s http://localhost:4000/api/v2/main-page/blocks | head -c 200

# L2 Blockscout frontend responding?
curl -fsS http://localhost:4001 >/dev/null
```

The local L2 can look idle in Blockscout after deployment if nobody sends L2
transactions. Seeing only the first deployment/funding blocks is not, by itself,
a failure; send any L2 transaction when you want the explorer UI to visibly move.

### Bridge/message smoke test

A scripted bridge/message smoke test is still a follow-up gate. Do not treat the
quickstart as final until that test proves the user-facing message path, not only
contract deployment and coordinator L1 submissions.

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
shape, but this knob is not part of the currently validated v0 first-boot path.
Use `ROLLUP` unless you are intentionally debugging the Validium path.

### Switching prover mode (dev ↔ partial)

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
# Expected in default quickstart mode:
#   prover_mode = "dev" × 4
#   is_allowed_circuit_id = 963
```

To run partial mode, set `PROVER_DEV_OVERRIDE=false`, raise Docker Desktop
memory substantially, optionally set `PROVER_GOMEMLIMIT`, and re-run the same
`up -d --force-recreate config-render prover` command.

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
docker compose --env-file versions.env --env-file .env --profile stack-partial-prover down -v --remove-orphans
docker compose --env-file versions.env --env-file .env --profile stack-partial-prover up -d
```

`down -v` is critical — without `-v`, the genesis volumes survive and the next
boot will reuse stale precomputed addresses against a deployer whose Sepolia
nonce has advanced.

### Common failures

| Symptom | Likely cause | Fix |
|---------|--------------|-----|
| `account-setup` exits with "L1 RPC not reachable" | `L1_RPC_URL` rate-limited or wrong | Try a different Sepolia RPC; check the URL with `cast chain-id --rpc-url $L1_RPC_URL` |
| `account-setup` exits with "could not extract deployers.l1" | malformed JSON output (Foundry version mismatch?) | See [bringup-notes.md](./bringup-notes.md) — `cast wallet address` flag drift is one possibility |
| deploy-contracts step 1 fails with "insufficient funds" | deployer's Sepolia balance too low | Top up via faucet; Sepolia gas spikes can blow through 0.5 ETH |
| deploy-contracts dies with `ADDRESS MISMATCH` | LineaRollupV8 or L2MessageService no longer lands at the boot-critical precomputed address | Usually means the deploy script changed and `account-setup.sh` needs a corresponding nonce/precompute update |
| Coordinator retries `linea_generateConflatedTracesToFileV2` with `Conflation not finished` on old block ranges | L2 ran far ahead while coordinator was down, beyond Besu's retained Bonsai history | Start over with `docker compose --env-file versions.env --env-file .env --profile stack-partial-prover down -v --remove-orphans`; this scaffold keeps a larger `bonsai-historical-block-limit` to make delayed first boots recoverable |
| Coordinator logs `address already reserved` during L1 blob/finalization submission | stale boot volume or old image/config still maps multiple jobs to the same signer | Fresh boot after the runtime-key cleanup should use distinct generated signer addresses; run `docker compose --env-file versions.env --env-file .env --profile stack-partial-prover down -v` before retesting |
| Aggregation/finalization occasionally reverts with a starting-root mismatch while catching up | coordinator is retrying aggregation windows while the local L2 catches up to the L1 contract state | Known caveat; watch whether `lastFinalizedBlockNumber` advances before calling it stuck |
| Coordinator restarts on first boot | usually a race against shomei's first-block trace; self-heals within ~30s | If it persists past 1 min, check `docker logs coordinator`; see fix log entries on web3signer mTLS |
| Web3signer mTLS handshake errors | known-clients fingerprint out of sync (only after regenerating one side) | Regenerate both sides or restore from git |
| Prover execution proofs exit `137` or files get `.large.failure.code_137` | partial prover ran under too little Docker memory | Use default `PROVER_DEV_OVERRIDE=true`, or raise Docker memory substantially before partial-mode validation |
| Port collision (5432, 8745, 4000, 4001, 9000, 3001, 9091) | Another service uses the port | Override via `HOST_PORT_*` in `.env` |

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

## 10. What's NOT yet in v0

- **No Timelock, no Security Council.** L1 contracts are owned by your deployer
  EOA. Governance/upgrade flows = v1.
- **No full prover.** >700 GB RAM is out of scope.
- **No L1 Blockscout.** Use Sepolia Etherscan.
- **Partial-prover validation is still pending.** The current validated path uses
  dev proofs for quick feedback; partial proving remains a separate acceptance
  gate.
- **No scripted bridge smoke test yet.** A real message/bridge path test is still
  needed before calling the quickstart final.
- **No CI smoke test.** Validation today is compose-config + manual
  first-boot-against-Sepolia. A scripted end-to-end smoke test remains follow-up
  work.
- **Partial proving is not the default.** The default is dev proofs for
  quickstart usability; partial proving remains opt-in and resource-heavy.

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

> Observability stack (prometheus / loki / promtail / grafana) was dropped in
> Phase-7 — out of scope for v0 deliverable. Use `docker logs <service>` for
> per-container output.

### File structure

```
docs/getting-started/linea-stack/
├── README.md                    ← you are here
├── SCAFFOLD-PLAN.md             ← historical planning doc (pre-Sepolia)
├── bringup-notes.md            ← fix history, current caveats, validation notes
├── docker-compose.yml
├── versions.env                 ← pinned image tags
├── .env.example                 ← copy to .env, fill in L1 RPC + key
├── scripts/
│   ├── account-setup.sh         ← generates runtime keys + boot addresses
│   ├── deploy-contracts.sh      ← 6-step deploy + address capture
│   ├── aggregate-addresses.ts   ← writes addresses.json from deploy logs
│   ├── links.sh                 ← prints useful Sepolia + local explorer links
│   ├── status.sh                ← redacted boot status summary
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

- `linea-shared-config` — written by `account-setup` (runtime keys,
  boot-critical precomputed addresses, web3signer keystores) and
  `deploy-contracts` (addresses.json, deploy logs); read by config-render,
  coordinator, postman, prover, web3signer
- `linea-rendered-config` — written by `config-render` (full render of 5
  templates) and `deploy-contracts` (in-place patch of coord-config); read by
  sequencer, maru, l2-node-besu, coordinator, prover
- `linea-local-dev` — chaindata + prover state
- `linea-logs` — kept as a shared log-output volume; was previously tailed
  by promtail (dropped Phase-7 with the rest of the observability stack).
- per-service postgres volumes
