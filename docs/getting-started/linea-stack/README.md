# Linea Stack — local quickstart

> _v0 of the "Streamlined Linea Stack deployment" feature. Stage 1 draft for review._

```
╔══════════════════════════════════════════════════════════════════════════════╗
║                                                                              ║
║   ⚠  DEV ONLY — DO NOT REUSE ON MAINNET ⚠                                    ║
║                                                                              ║
║   This stack ships pre-baked secrets and bootstrap state so it can boot      ║
║   on a fresh laptop in <10 minutes. EVERYTHING listed below is public        ║
║   knowledge — anyone reading this repository has it. Treat any address,      ║
║   key, or address-derived artefact in this stack as already compromised.    ║
║                                                                              ║
║   Compromised by virtue of being in the public repo:                         ║
║     • Private keys                                                           ║
║         - config/l1/el/besu.key                  (L1 Besu node identity)     ║
║         - config/l1/cl/teku.key                  (L1 Teku node identity)     ║
║         - config/l1/cl/teku-keys/0x*.json        (L1 validator keystores)    ║
║         - config/l2/sequencer/key                (sequencer identity)        ║
║         - config/l2/maru/private-key             (Maru consensus signer)     ║
║         - config/web3signer/key-files/*.yaml     (4 operational signers)     ║
║         - L1_SIGNER_PRIVATE_KEY, L2_SIGNER_PRIVATE_KEY in                    ║
║           config/l2/postman/env                  (postman EOAs)              ║
║         - hardcoded L1/L2 deployer keys in scripts/deploy-contracts.sh       ║
║         - hardcoded L1/L2 seed keys in scripts/seed-funds.sh                 ║
║     • Mnemonics                                                              ║
║         - config/l1/genesis-generator/mnemonics.yaml                         ║
║     • TLS certs and passwords                                                ║
║         - config/web3signer/tls-files/web3signer-keystore.p12                ║
║         - config/web3signer/tls-files/web3signer-keystore-password.txt       ║
║         - config/web3signer/tls-files/known-clients.txt                      ║
║         - config/l1/cl/teku-secrets/0x*.txt      (validator key passwords)   ║
║     • Funded accounts                                                        ║
║         - L1 deployer 0xf39Fd6e5… and L2 deployer 0x70997970…  (1000 ETH)    ║
║         - operational accounts seeded by scripts/seed-funds.sh   (100 ETH)   ║
║     • Genesis files (deterministic — anyone can regenerate identical ones)   ║
║         - config/l1/el/genesis.json                                          ║
║         - config/l2/genesis-init/genesis-besu.json.template                  ║
║         - config/l2/genesis-init/genesis-maru.json.template                  ║
║     • Contract addresses (deterministic from deployer key + nonce)           ║
║         - LineaRollup at 0xDc64a140…  (matches mainnet rollup address by     ║
║           virtue of the deterministic deploy — DO NOT confuse)               ║
║         - L2MessageService at 0xe537D669…                                    ║
║         - L1/L2 BridgedToken, L1/L2 TokenBridge (see addresses.json)         ║
║                                                                              ║
║   Full enumeration + 8-step regen checklist: config/DEV-KEYS-INVENTORY.md.   ║
║                                                                              ║
║   Never reuse, copy, or derive from any of the above for any non-local      ║
║   network. Never expose this stack on a public network without first         ║
║   regenerating EVERY item listed.                                            ║
║                                                                              ║
╚══════════════════════════════════════════════════════════════════════════════╝
```

> **Apple Silicon (M-series Mac) note — read first.** The Linea prover image
> is `linux/amd64` only. On M-series Macs it runs under Rosetta emulation and
> is **3–5× slower than native x86_64**. The first proof in
> `stack-partial-prover` mode can take 30+ minutes on M-series; expect 5–10
> min on native x86_64. Everything else in the stack is multi-arch and runs
> natively. If you're on Apple Silicon and only need L1↔L2 message flow
> (no zk proofs), use `stack-no-prover`.

## 1. What this is

A self-contained Docker Compose stack that brings up a complete local Linea zk-rollup — L1 (Besu + Teku), L2 (Linea Besu sequencer + Maru consensus + RPC follower + Shomei state manager), the coordinator, postman, web3signer, Blockscout explorers for both layers, and a Prometheus / Grafana / Loki observability bundle. From `docker compose up` to a working L1↔L2 message round-trip in under an hour, with no Linea team intervention.

Two profiles:

- **`stack-no-prover`** — full L1+L2 stack with explorers and dashboards. Coordinator runs without a paired prover (no zk proofs are produced; finalization is mocked). Lighter and fast to bring up.
- **`stack-partial-prover`** — same plus the Linea partial prover. Real zk proof generation, but with a partial trace set so it fits in ~32 GB RAM. Full prover requires >700 GB RAM and is out of scope for v0.

Two read-mes worth flagging up front:

- This v0 is **not portable** — it bind-mounts `../../../contracts` from the linea-monorepo so the deployment script can run Hardhat tasks against the real contracts. v1 will package contracts into a published image.
- L1 is local — Besu + Teku running locally. Pointing at Sepolia or another L1 is a v1 feature; the deploy-contracts script already takes the L1 RPC as a parameter so the wiring is forward-compatible.

## 2. Hardware requirements

|                | `stack-no-prover` | `stack-partial-prover` |
|----------------|-------------------|------------------------|
| **RAM (min)**  | 16 GB             | 48 GB                  |
| **RAM (rec.)** | 24 GB             | 64 GB                  |
| **CPU**        | 4 cores           | 8 cores                |
| **Disk**       | 50 GB free        | 80 GB free             |
| **Docker**     | v24+              | v24+                   |
| **Compose**    | v2.19+            | v2.19+                 |

Partial prover footprint is dominated by the prover process (`GOMEMLIMIT=32GiB`). Coordinator + sequencer + L1 nodes together fit comfortably in the no-prover budget.

> Apple Silicon: see the M-series note at the top of this README.

## 3. Quickstart — `stack-no-prover`

```bash
cd docs/getting-started/linea-stack
cp .env.example .env                # tune ports if you have collisions
docker compose --env-file versions.env --env-file .env --profile stack-no-prover up -d
```

The first boot takes ~5–10 minutes (image pulls + L1 genesis + L2 genesis + contract deployment). When it's ready, you'll see:

```
[STAGE-2 placeholder — fill with actual `docker compose ps` excerpt after first
 successful boot. Will list ~17 services in `running`/`healthy` state and the
 deploy-contracts + seed-funds + l1-genesis-generator + l2-genesis-init oneshots
 in `exited (0)`.]
```

**Endpoints:**

| What                | URL                            |
|---------------------|--------------------------------|
| L1 RPC              | http://localhost:8545          |
| L1 Beacon (Teku)    | http://localhost:4000          |
| **L2 RPC**          | **http://localhost:8745**      |
| L2 WebSocket        | ws://localhost:8746            |
| Coordinator API     | http://localhost:9545          |
| Web3signer          | http://localhost:9000          |
| Postman             | http://localhost:9090          |
| L1 Blockscout       | http://localhost:4001          |
| L2 Blockscout       | http://localhost:4000          |
| Grafana             | http://localhost:3001          |
| Prometheus          | http://localhost:9091          |

The sequencer also listens on `:8645` but **that port is internal-only** — connect your wallet/dapp/SDK to `:8745` (l2-node-besu).

To send a first L1↔L2 message:

```
[STAGE-2 placeholder — fill with the exact commands once seed-funds.sh is
 finalised. Will be: cast send to L1 LineaRollup.sendMessage(...), watch
 postman logs, see message anchored on L2.]
```

## 4. Quickstart — `stack-partial-prover`

Same as above with a different profile:

```bash
docker compose --env-file versions.env --env-file .env --profile stack-partial-prover up -d
```

First boot takes longer (~15–20 min) because the prover container needs to load assets and the first proof costs more wall-clock than steady state. On Apple Silicon, expect 30+ min on first conflation.

The prover is the **only** service that differs between profiles — everything else is identical.

## 5. What just happened — bring-up sequence

`docker compose up` walks through this dependency chain. Each service waits for its predecessors via `depends_on: condition: service_healthy` or `service_completed_successfully`.

1. **L1 genesis** — `l1-genesis-generator` (one-shot, ethpandaops/ethereum-genesis-generator) generates a fresh L1 genesis SSZ + JSON into the shared volume.
2. **L1 boot** — `l1-el-node` (Besu) boots from the generated genesis, then `l1-cl-node` (Teku) attaches to it via the engine API. Healthchecks gate further steps.
3. **L2 genesis** — `l2-genesis-init` (one-shot, busybox) renders the L2 Besu and Maru genesis files from templates. The L2MessageService's deterministic address is funded in genesis.
4. **L2 boot** — `sequencer` boots, then `maru` (consensus) and `l2-node-besu` (RPC follower) attach. `shomei` (state manager) attaches to the follower.
5. **Web3signer** — boots in parallel; serves the four pre-baked dev keys (anchoring, data-submission, finalization, liveness) over mTLS on `:9000`.
6. **Postgres** — four isolated instances (`coordinator-pg`, `postman-pg`, `blockscout-l1-pg`, `blockscout-l2-pg`) come up in parallel.
7. **Contract deployment** — `deploy-contracts` (one-shot, node) runs the Hardhat tasks in this order: L1 Verifier → L1 LineaRollup → L1 BridgedToken → L2 L2MessageService → L2 BridgedToken → compute L2 Token Bridge address → L1 Token Bridge → L2 Token Bridge. Writes `addresses.json` to a shared volume. **No Timelock, no Security Council in v0.**
8. **Fund seeding** — `seed-funds` (one-shot) reads the genesis seed account and dispatches ETH to the operational accounts on both L1 and L2 (deployer, shnarf submitter, finalisation submitter, postman, anchorer). Idempotent — re-running is safe.
9. **L2 service activation** — `coordinator` and `postman` start once `deploy-contracts` has completed. Both read contract addresses from `/shared/addresses.json` (mounted read-only).
10. **Prover (partial profile only)** — joins after the coordinator is up.
11. **Explorers + observability** — Blockscout L1 + L2 attach to their respective nodes; Prometheus scrapes; Loki ingests via Promtail; Grafana provisions datasources and the Linea L2 health dashboard.

When step 11 finishes, the stack is ready.

## 6. Step-by-step (manual flow, for understanding)

This section describes what `deploy-contracts.sh` and `seed-funds.sh` do under the hood. You don't run any of these yourself — but reading this lets you trust the automation, and gives you a hook for debugging if a step fails.

```
[STAGE-2 placeholder — distilled walkthrough of the internal "Deploy a new
 Linea Network" guide, restricted to the v0 scope (no Timelock, no Security
 Council, no upgrades). Will cover:
   - Verifier (Plonk) deployment
   - LineaRollup (the validium-capable rollup contract) deployment with
     initial operators, rate limits, and the LineaShnarf seed
   - L1 BridgedToken deployment
   - L2MessageService deployment at the deterministic address; checking the
     genesis pre-fund matches
   - L2 BridgedToken deployment
   - Computing the deterministic L2 Token Bridge address (deployer + nonce)
   - L1 Token Bridge deployment with the precomputed L2 address baked in
   - L2 Token Bridge deployment
   - addresses.json shape and how each service consumes it
   - Operational account derivation and funding amounts]
```

## 7. Customisation

### Changing host ports

All host port mappings are env-driven. Copy `.env.example` to `.env` and uncomment the ports you want to change. Container ports are fixed.

### Pointing at a different L1

`scripts/deploy-contracts.sh` takes the L1 RPC URL as its first positional argument. Today, `docker-compose.yml` calls it with `http://l1-el-node:8545`. To target a remote L1 (Sepolia, a private testnet, etc.) in v1, the v1 release will add an `external-l1` profile that disables the local L1 services and accepts a `L1_EXTERNAL_RPC_URL` env var. Until then, the script is forward-compatible — only the compose wiring needs to change.

### Overriding configs

Every config file is bind-mounted from `config/<group>/...`. To override one, edit the file in place and restart the relevant service. If you change a genesis-related file, you must `docker compose down -v` to wipe the volumes; otherwise services reuse the previous genesis.

### Switching data availability mode (Rollup ↔ Validium)

```bash
LINEA_COORDINATOR_DATA_AVAILABILITY=VALIDIUM docker compose --profile stack-no-prover up -d
```

The deploy-contracts script reads the same env var and configures the LineaRollup constructor accordingly.

### Increasing prover memory

The prover's `GOMEMLIMIT` is set in `docker-compose.yml`. v0 uses 32 GiB for partial mode. If you have a larger machine and want longer trace runs, raise it before the first prover boot. It's a Docker container memory ceiling — Docker Desktop's overall memory limit must be raised first.

## 8. Troubleshooting

### "Just start over"

```bash
docker compose down -v --remove-orphans
docker volume prune -f
docker compose --env-file versions.env --env-file .env --profile stack-no-prover up -d
```

`down -v` is critical — without `-v`, the genesis volumes survive and the next boot will mismatch the sequencer key against a stale chain.

### Specific failures

```
[STAGE-2 placeholder — fill from first-boot validation. Expected entries:
   - Port collisions (5432, 8545, 4000, 4001, 9000) → override via .env
   - deploy-contracts fails because contracts/ build artifacts missing →
     run `pnpm install && pnpm exec hardhat compile` once at the repo root
   - Apple Silicon prover OOM → raise Docker Desktop memory cap to 48GB+
   - Coordinator restarts on first boot → race against shomei first-block
     trace; usually self-heals within 30s
   - Web3signer mTLS handshake errors → known-clients.txt out of sync with
     coordinator's TLS cert; only happens if you regenerate one side
   - Promtail can't read /var/log → SELinux/AppArmor; disable for the
     `linea-stack-logs` volume on Linux hosts]
```

### Inspecting

```bash
# All services and their health
docker compose --profile stack-no-prover ps

# Tail coordinator + sequencer logs
docker compose logs -f coordinator sequencer

# Inspect addresses.json (the contract handoff)
docker compose exec coordinator cat /shared/addresses.json

# Open a debug shell on the linea network
docker run --rm -it --network linea-stack_linea weibeld/ubuntu-networking bash
```

## 9. What's NOT in v0

To set expectations clearly:

- **No Timelock, no Security Council.** Contracts are owned by the deployer EOA. Any governance/upgrade flow is a v1 concern.
- **Full prover.** Out of scope (>700 GB RAM). Use partial prover or no-prover.
- **External L1 (Sepolia/mainnet).** The script is parameterised for this but the compose wiring and verifier choice still hard-assume local L1.
- **Cross-host deployment / Kubernetes.** v0 is single-host Docker Compose only.
- **Backups, snapshots, restore from L1.** State recovery exists in the internal stack (`compose-tracing-v2-staterecovery-extension.yml`) but is not part of v0.
- **CI smoke tests.** Coming in a separate PR after v0 is reviewed.

## 10. Once more, with feeling

```
╔══════════════════════════════════════════════════════════════════════════════╗
║                                                                              ║
║   ⚠  DEV ONLY — DO NOT REUSE ON MAINNET ⚠                                    ║
║                                                                              ║
║   Every artefact listed below is public knowledge by virtue of being in      ║
║   this repository. None of it will protect any funds. Treat every address    ║
║   derived from any of it as compromised by default.                          ║
║                                                                              ║
║   What is compromised:                                                       ║
║     • Private keys      — L1 Besu node, L1 Teku node, L1 validator keys      ║
║                           (4 keystores), L2 sequencer node, L2 Maru          ║
║                           consensus signer, 4 web3signer operational keys    ║
║                           (anchoring/data-submission/finalization/liveness), ║
║                           postman L1+L2 signer EOAs, deploy-contracts.sh     ║
║                           hardcoded deployer keys, seed-funds.sh hardcoded   ║
║                           seed keys.                                         ║
║     • Mnemonics         — L1 genesis-generator mnemonic (seeds funded        ║
║                           accounts including the deployer at 0xf39Fd6e5…).   ║
║     • TLS certs         — web3signer mTLS keystore and its password,         ║
║                           known-clients fingerprints.                        ║
║     • Passwords         — postgres credentials (postgres/postgres),          ║
║                           validator key password files, TLS keystore         ║
║                           password.                                          ║
║     • Funded accounts   — every account seeded by genesis or by              ║
║                           scripts/seed-funds.sh, including the L1 + L2       ║
║                           deployers (1000 ETH each) and 5 operational        ║
║                           accounts (100 ETH each).                           ║
║     • Genesis files     — L1 EL genesis.json, L2 Besu genesis,               ║
║                           L2 Maru genesis. Deterministic from the            ║
║                           checked-in mnemonic + node keys.                   ║
║     • Contract addresses — LineaRollup, L2MessageService, both               ║
║                           BridgedTokens, both TokenBridges. All              ║
║                           deterministic from deployer key + nonce — your    ║
║                           local stack will produce the SAME addresses as     ║
║                           anyone else's local stack.                         ║
║                                                                              ║
║   Before exposing this stack on any non-loopback interface, follow the       ║
║   8-step regen checklist in config/DEV-KEYS-INVENTORY.md. If you can't       ║
║   tick all eight, keep it on localhost. There is no middle ground.           ║
║                                                                              ║
╚══════════════════════════════════════════════════════════════════════════════╝
```
