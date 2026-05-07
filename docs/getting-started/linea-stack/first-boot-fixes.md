# First-boot fixes — 2026-05-06

Issues hit during the first `docker compose up -d` pass and what we did about them.

| # | Fix | Why |
|---|---|---|
| 1 | POSTGRES_TAG 18.3 → 17.6 | PG 18 changed on-disk layout; Blockscout/coordinator schemas not compatible yet |
| 2 | traces-limits-v5.toml mounted into sequencer + l2-node-besu | Stage 2 missed the mount; both services need it |
| 3 | Sequencer TLS files mounted | Stage-2 gap; sequencer→web3signer needs mTLS material |
| 4 | Coordinator TLS files copied | Closes earlier 🔴 TODO from stage 2 |
| 5 | Sequencer pinned to 11.11.11.101 | Besu's enode parser rejects hostnames — stage-1/2 hostnames-only assumption was wrong |
| 6 | l2-node-besu bootnode → IP literal | Same root cause as #5 |
| 7 | Web3signer healthcheck removed | mTLS-only mode, no plain HTTP endpoint to probe |
| 8 | Dependents on web3signer → service_started | Follows from #7 |
| 9 | HOST_PORT_L1_BEACON 4000 → 4002 | Collided with l2-blockscout |
| 10 | deploy-contracts bind-mounts whole monorepo | pnpm workspace catalog resolution needs the full tree |
| 11 | --frozen-lockfile → --no-frozen-lockfile | Lockfile mismatch in container env |
| 12 | Foundry auto-installs in deploy-contracts container | Hardhat config requires `forge` |
| 16 | L2 Blockscout: dropped `BLOCK_TRANSFORMER=clique` from `config/explorer/l2-blockscout.env`; removed dead `clique` config block from L2 genesis; replaced clique-format extraData with `0x`; added `"ethash": {}` (Besu won't accept a chain config with no consensus mechanism — TTD=0 makes the merge happen at block 0, exactly as L1 does); commented out PoA-only `poa-block-txs-selection-max-time` in `sequencer.config.toml`; flipped `INDEXER_DISABLE_PENDING_TRANSACTIONS_FETCHER` to `true` (Linea Besu doesn't expose `txpool_content`). | The `BLOCK_TRANSFORMER=clique` env caused `Indexer.Transform.Blocks.Clique.recover_pub_key/2` to crash on every block: `LineaExtraDataPlugin` writes custom gas-pricing bytes, not a 65-byte clique seal. After fix: zero `recover_pub_key`/MatchError occurrences in logs; L2 Blockscout indexes blocks/txs/logs cleanly. **Note:** `localhost:4000/` returns 404 — Blockscout 7.x splits the UI into a separate frontend container that this scaffold does not deploy (same on L1 at `:4001`). API at `/api/v2/blocks` returns 200 and is queryable. Frontend container is a separate scaffold gap, not part of this fix. |

## Sepolia migration — Phase 1 (mechanical surgery), 2026-05-07

Per the migration plan: drop the local L1 stack, rewire L1 endpoints to a user-supplied Sepolia RPC. **Mechanical-only — no logic changes.** Stack does not yet boot end-to-end on Sepolia; that arrives in phases 2–4.

| # | Change | Why |
|---|---|---|
| 17a | Dropped 5 services from `docker-compose.yml`: `l1-genesis-generator`, `l1-el-node`, `l1-cl-node`, `l1-blockscout`, `blockscout-l1-pg`. Dropped `seed-funds` service entirely. | Sepolia is the v0 L1; local Besu+Teku was dev-loop scaffolding. seed-funds dispatched ETH from a local genesis seed — irrelevant for Sepolia (Option A: single L1 deployer key for all roles). |
| 17b | Deleted `config/l1/`, `config/explorer/l1-blockscout.env`, `config/postgres/blockscout-l1-init.sql`, `scripts/seed-funds.sh`. | Source-of-truth files for the dropped services. No L1 explorer in this scaffold; users get pointed at sepolia.etherscan.io. |
| 17c | Renamed `coordinator-config.toml` and `maru/config.toml` to `.template`; replaced literal `http://l1-el-node:8545` with placeholder `__L1_RPC_URL__`. Maru's hardcoded `contract-address = "0xDc64..."` (LineaRollup proxy) became `__LINEA_ROLLUP_ADDRESS__` — to be patched at deploy time in Phase 2. | Coordinator + Maru read TOML at JVM start; neither natively supports env interpolation in TOML. Templates+render is more robust than betting on `config__override__*` env paths working for both. |
| 17d | Added `config-render` busybox init service. Mounts both `.template` files read-only at `/templates`, `sed`-substitutes `__VAR__` placeholders from compose env, writes rendered files to `linea-rendered-config` volume. coordinator + maru mount the rendered files read-only and target them via `--config=/rendered/...`. | Single chokepoint for runtime config substitution. NOT idempotent — re-renders every boot so an .env change is always picked up. Phase 2 will add more placeholders (LINEA_ROLLUP_ADDRESS, L1_CHAIN_ID, security council, etc). |
| 17e | `deploy-contracts` + `postman` services: `L1_RPC_URL` now passes through from compose env (`${L1_RPC_URL:?...}`). For postman, compose `environment:` overrides the dead line in `config/l2/postman/env` — that line gets cleaned up in Phase 4. | Both services consume L1_RPC_URL at runtime; passing through env beats template rendering for non-TOML consumers. |
| 17f | `versions.env`: dropped `TEKU_TAG`, `ETH_GENESIS_GENERATOR_TAG`. Updated comments to reflect Sepolia-only L1. `LINEA_BESU_PACKAGE_TAG` stays — sequencer + l2-node-besu still use it. | Image tags for services that no longer exist would just produce confusing "tag not pinned" warnings. |
| 17g | New `.env.example` with two REQUIRED variables: `L1_RPC_URL` (Sepolia HTTPS RPC) and `L1_DEPLOYER_PRIVATE_KEY` (Sepolia-funded). Rest of the file is optional knobs. | These two are the only user-supplied secrets; everything else is L2 dev keys checked into the repo. The `${VAR:?msg}` pattern in compose surfaces a clear error if either is missing. |

**Phase 1 validation:** `docker compose config` parses cleanly; service list shows 18 services with `config-render` added and all 6 dropped services absent. The stack does not boot end-to-end yet because `deploy-contracts.sh` still has three references to the local L1 (default `http://l1-el-node:8545`, hardcoded `L1_CHAIN_ID="31648428"`, comment) — that's intentional Phase 2 scope.

**Known carry-overs into later phases:**
- Phase 2: `deploy-contracts.sh` chain-ID detection, security council/operator key derivation from `L1_DEPLOYER_PRIVATE_KEY`, runtime L2 genesis state-root + shnarf computation, Sepolia readiness check. Bump L2MessageService genesis balance to 1B ETH.
- Phase 3: replace 3 web3signer keystore entries (blob, aggregation, anchoring) with the user's L1 key. Keep web3signer in the loop (Victorien required it).
- Phase 4: Sepolia timing tunables (`block-time = "PT12S"`, `consistent-number-of-blocks-on-l1-to-wait`, postman polling intervals). README rewrite.

## Sepolia migration — Phase 2 (deploy-contracts + config patching), 2026-05-07

Wires the deployment script to user-supplied Sepolia config, makes runtime values discoverable instead of hardcoded, patches the rendered configs after deploy. Stack still doesn't run end-to-end (Phase 3 owns the web3signer keystore + signer-key alignment).

| # | Change | Why |
|---|---|---|
| 18a | `genesis-besu.json.template`: L2MessageService genesis balance 9e23 wei → 1e27 wei (≈900K → 1B ETH). Stripped the misleading "@WARNING / account 21" comment; replaced with "L2MessageService — pre-funded for L1->L2 message payouts". | Per Victorien's review for Sepolia migration. |
| 18b | `coordinator-config.toml.template`: hardcoded `genesis-state-root-hash`, `genesis-shnarf`, `[protocol.l1].contract-address`, `[protocol.l2].contract-address`, `contract-deployment-block-number` replaced with `__PLACEHOLDER__` tokens. `block-time` bumped from `PT1S` (local devnet) to `PT12S` (Sepolia). | Each value is either deploy-time discoverable or known only after L2 genesis state settles; hardcoded values were wrong on any chain other than the local dev L1. |
| 18c | `config-render` service extended: substitutes 6 placeholders. `__L1_RPC_URL__` from `.env`. The other five (`__LINEA_ROLLUP_ADDRESS__`, `__L2_MESSAGE_SERVICE_ADDRESS__`, `__GENESIS_STATE_ROOT_HASH__`, `__GENESIS_SHNARF__`, `__LINEA_ROLLUP_DEPLOY_BLOCK__`) get safe defaults: zero-address, zero-hash, zero-block, plus the deterministic L2MessageService address. | Two-phase model: defaults at boot so maru can come up before deploy-contracts; deploy-contracts re-patches the rendered files with real values; `post-deploy-restart` cycles maru. |
| 18d | `deploy-contracts.sh` rewrite — runtime detection block added between pre-flight and pnpm install: <br/>• `cast chain-id --rpc-url $L1_RPC_URL` → `L1_CHAIN_ID` (drops the hardcoded `31648428`). <br/>• `cast wallet address --private-key $L1_DEPLOYER_PRIVATE_KEY` → `L1_DEPLOYER_ADDRESS`. <br/>• `cast block 0 --rpc-url $L2_RPC_URL --field stateRoot` → `L2_GENESIS_STATE_ROOT`. <br/>• `cast keccak <5×32-byte concat>` → `L2_GENESIS_SHNARF`. | Foundry is already installed at the top of the script; cast runs before any deploy. Values flow into step 1 (`INITIAL_L2_STATE_ROOT_HASH`, security council, operators, security-council-private-key) and into the post-deploy patch block. |
| 18e | `deploy-contracts.sh` step 1 + step 3: dropped 4 hardcoded Hardhat-style addresses (`L1_SECURITY_COUNCIL`, `LINEA_ROLLUP_OPERATORS`, `SECURITY_COUNCIL_PRIVATE_KEY`, `INITIAL_L2_STATE_ROOT_HASH`). Now all derive from `L1_DEPLOYER_PRIVATE_KEY` (Option A). The L2 deployer key + L2 security council + L1L2_MESSAGE_SETTER stay pre-baked — L2 is dev. | Single user-supplied L1 key drives every L1 role. Phase-3 will align the web3signer keystore + coordinator inline signer keys to the same address. |
| 18f | `deploy-contracts.sh` end-of-script: new "Patch rendered coordinator + maru configs" section. Maru patched via line-anchored sed (only one `contract-address` line). Coordinator patched via section-aware awk (two `contract-address` lines must be discriminated by `[protocol.l1]` vs `[protocol.l2]`). State root, shnarf, and deploy block also patched into coordinator config. | The Spring `config__override__` env mechanism couldn't cleanly express section-specific overrides; awk on the rendered file is more direct. |
| 18g | `deploy-contracts.sh`: pre-flight L1_RPC_URL fallback `http://l1-el-node:8545` removed. `${L1_RPC_URL:?...}` makes it required. `L1_DEPLOYER_PRIVATE_KEY` similarly required (no Hardhat default). | Local-L1 fallbacks no longer make sense — that path is dead. |
| 18h | New `post-deploy-restart` compose service (image: `docker:cli`, mounts `/var/run/docker.sock`). Depends on `deploy-contracts:service_completed_successfully`; runs `docker restart maru`. | Maru reads its config at boot; the deploy-time patch happens after maru starts. Restart cycles it onto the patched config. Coordinator + postman do NOT need restart — they depend on `deploy-contracts:completed_successfully` and start *after* the patch lands. |
| 18i | `deploy-contracts` service: added `linea-rendered-config:/rendered:rw` volume mount + `config-render:service_completed_successfully` depends_on. | Script needs RW access to /rendered to patch the configs; depends_on ensures the rendered files exist before patching. |

**Phase 2 validation:**
- `bash -n scripts/deploy-contracts.sh` → OK
- `docker compose config` → no errors, 19 services in profile (was 18; added `post-deploy-restart`)
- All 6 `__PLACEHOLDER__` tokens in templates have a matching substitution in `config-render`'s sed pipeline ✅

**Verified at first-boot (TODOs for Phase 5):**
- `cast block 0 --field stateRoot` syntax works against Linea Besu's block-0 response. (Some Foundry versions name the flag differently; if so, fall back to `cast block 0 --json | jq -r .stateRoot`.)
- `cast keccak` 5×32-byte concat shnarf formula matches what V8 LineaRollup expects. (The formula in the original coordinator-config comment was annotated "shnarf for contract V6"; verify V8 didn't change it.)
- L1_CHAIN_ID is captured as a decimal integer (Foundry default) — the lane-assignment code in addresses.json comparison is string-equal so it has to match exactly.

