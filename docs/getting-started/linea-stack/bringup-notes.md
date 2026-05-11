# Linea Stack bring-up notes

Running log for the Sepolia quickstart: fixes applied, current status, and caveats that still block calling the project finished.

## Current snapshot - 2026-05-11

Fresh Sepolia boot validates the cleaned key model and the current dev-prover path:

- user still supplies one funded `L1_DEPLOYER_PRIVATE_KEY`;
- `account-setup` generates fresh runtime keys for L1 blob submission, L1 finalization, L1 postman, L2 deployer, L2 anchorer, and L2 postman;
- L2 genesis only funds the generated L2 deployer plus the precomputed `L2MessageService` address;
- only `LineaRollupV8` and `L2MessageService` are precomputed before boot;
- generated genesis/rendered artifacts are ignored, while templates stay committed;
- liveness is disabled for the v0 quickstart;
- `deploy-contracts` writes `addresses.json`, funds generated runtime signers, and is retry-safe from persisted deploy logs;
- Web3Signer loads 3 generated signer key files, postman starts with generated L1/L2 postman keys, and coordinator ports bind;
- coordinator submitted L1 blob transactions and aggregations on Sepolia, and finalization advanced to L2 block 10;
- L2 Blockscout frontend works locally at `http://localhost:4001`.

Fixes found during this clean boot:

- L1 blob/finalization signer top-up default reduced from 0.25 ETH to 0.15 ETH each;
- `/shared/runtime-keys.env` must be readable by non-root service containers, otherwise postman cannot start;
- `/shared/web3signer-keys/*.yaml` must be readable by the Web3Signer container, otherwise Web3Signer starts with zero signers and coordinator gets signer `404 Not Found` errors.

Current caveats:

- default proving is still dev/dummy proof mode; partial-prover validation remains a separate gate;
- a documented bridge/message smoke test is still needed;
- transient nonce/replacement retries can appear during catch-up; judge progress by blob/aggregation txs and finalized block advancing;
- the local L2 does not necessarily keep producing visible user blocks when idle. Send an L2 transaction to create fresh blocks for Blockscout demos.

Next work should focus on a repeatable L2 traffic command, bridge/message smoke test, and partial-prover validation.

## Historical fix log

Issues hit during the first `docker compose up -d` passes and what we did about them.

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
| 16 | L2 Blockscout: dropped `BLOCK_TRANSFORMER=clique` from `config/explorer/l2-blockscout.env`; removed dead `clique` config block from L2 genesis; replaced clique-format extraData with `0x`; added `"ethash": {}` (Besu won't accept a chain config with no consensus mechanism — TTD=0 makes the merge happen at block 0, exactly as L1 does); commented out PoA-only `poa-block-txs-selection-max-time` in `sequencer.config.toml`; flipped `INDEXER_DISABLE_PENDING_TRANSACTIONS_FETCHER` to `true` (Linea Besu doesn't expose `txpool_content`). | The `BLOCK_TRANSFORMER=clique` env caused `Indexer.Transform.Blocks.Clique.recover_pub_key/2` to crash on every block: `LineaExtraDataPlugin` writes custom gas-pricing bytes, not a 65-byte clique seal. After fix: zero `recover_pub_key`/MatchError occurrences in logs; L2 Blockscout indexes blocks/txs/logs cleanly. At this point only the API was wired; the frontend was added later in #47. |

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

## Sepolia migration — Phase 2 redo (pre-compute → inject → boot model)

Phase 2 (deploy-then-patch) was the wrong shape. The internal Linea engineering playbook computes contract addresses deterministically *before* boot and injects them into genesis + every service config. Phase 2.1 lays the foundation for that model.

### Phase 2.1 (pre-boot account derivation + address pre-computation), 2026-05-07

| # | Change | Why |
|---|---|---|
| 19a | New `scripts/account-setup.sh` (POSIX sh — runs in `ghcr.io/foundry-rs/foundry` alpine image, no bash). Inputs: `L1_RPC_URL`, `L1_DEPLOYER_PRIVATE_KEY`, optional `L2_DEPLOYER_PRIVATE_KEY` and `L2_LIVENESS_SIGNER_PRIVATE_KEY`. Outputs: `/shared/addresses-precomputed.json` + 4 web3signer YAMLs. | The first init in the new boot order. Everything downstream (genesis pre-fund, config-render, deploy-contracts verify-or-die) reads from this JSON. |
| 19b | Pre-computes 6 L1 + 4 L2 contract addresses via `cast compute-address` against the existing `deploy-contracts.sh` nonce sequence. L1 LineaRollupV8 (proxy) at `L1_NONCE+5`, L1 TokenBridge at `+10`. L2MessageService at L2 nonce 2. | Per the user's Option A clarification: pre-compute against the *existing* deploy script, not refactor to the playbook's Timelock-first sequence. If the deploy script changes contract order, both this script and Phase 2.4's verify-or-die must be updated. |
| 19c | Generates 4 web3signer keystore YAMLs (`anchoring`, `data-submission`, `finalization`, `liveness`) into `/shared/web3signer-keys/`. The 3 L1 signers all use `L1_DEPLOYER_PRIVATE_KEY` (Option A — single key). The liveness signer keeps the pre-baked dev L2 key (it signs L2 sequencer-liveness txs only). Files are `file-raw` YAMLs, not PKCS#12. | Web3signer stays in the loop per Victorien's requirement; we just align its key material with the actual user-supplied L1 key. The misleading "anchoring on L2" comment in the original keystore was wrong about which chain it signs — anchoring submits L1 txs to LineaRollup. |
| 19d | New compose service `account-setup` (image `ghcr.io/foundry-rs/foundry:${FOUNDRY_TAG}`). Mounts `linea-shared-config:/shared:rw` + `./scripts:/scripts:ro`. Runs first; no `depends_on`. | Foundry image gives us `cast` for free (deriving addresses, querying nonces, computing CREATE addresses) without installing anything at boot. |
| 19e | `web3signer` service rewired: `--key-store-path=/shared/web3signer-keys/` (was `/key-files/`). Volumes: `linea-shared-config:/shared:ro` (was bind-mount of `config/web3signer/key-files/`). Added `depends_on: account-setup:service_completed_successfully`. | Web3signer now reads keystores generated at boot from the user's `.env` instead of from static dev keys checked into the repo. |
| 19f | `versions.env`: added `FOUNDRY_TAG=latest` with a Phase-5 TODO to pin to a specific build. | Image tag pinning convention; `latest` is acceptable for v0 but not for any production-ish use. |

**Phase 2.1 validation:**
- `sh -n` on `account-setup.sh` → POSIX clean, no bashisms
- `docker compose config -q` on `--profile stack-partial-prover` → no errors
- Resolved config shows `web3signer` correctly depending on `account-setup` and mounting the shared volume at `/shared` read-only

**Static `config/web3signer/key-files/*.yaml`** are now vestigial (still in the repo as a safety net during the migration). Removal scheduled for Phase 4 cleanup.

**Outstanding for next phases:**
- 2.2: update `genesis-besu.json.template` to use `__L2_MESSAGE_SERVICE_ADDRESS__` placeholder; render genesis from precomputed JSON before `l2-genesis-init`
- 2.3: rewrite `config-render` to read `/shared/addresses-precomputed.json` and substitute every address-related placeholder (coordinator, maru, sequencer, l2-node-besu, postman)
- 2.4: rewrite `deploy-contracts.sh` for verify-or-die: read precomputed JSON, query Shomei (not eth_getBlockByNumber) for state root, deploy, verify each address matches
- 2.5: second-pass render after Shomei is up — fills coordinator's `genesis-state-root-hash` + shnarf so coordinator never sees a placeholder
- 2.6: drop `stack-no-prover` profile

### Phase 2.2 (genesis pre-funds the precomputed L2MessageService address), 2026-05-07

| # | Change | Why |
|---|---|---|
| 20a | `genesis-besu.json.template`: the alloc-map KEY for L2MessageService changed from the literal `0xe537D669CA013d86EBeF1D64e40fC74CADC91987` to the placeholder `__L2_MESSAGE_SERVICE_ADDRESS__`. Comment updated: nonce 0 → nonce 2 (impl + ProxyAdmin + proxy = 3 contracts; the proxy is the third), and references the precomputed JSON as the source. | Address is no longer hardcoded — it comes from `addresses-precomputed.json` which uses the actual L2 deployer (currently the pre-baked dev key, but future-proof if it ever changes). The "nonce 0" was a Phase-2 typo; the L2MessageService proxy is at nonce 2 because two contracts deploy before it. |
| 20b | `config/l2/genesis-init/init.sh`: rewritten to (i) require `/shared/addresses-precomputed.json` (FATAL if missing), (ii) extract `l2.L2MessageService` via POSIX `sed -nE 's/.*"L2MessageService":[[:space:]]*"(0x[a-fA-F0-9]{40})".*/\1/p'`, (iii) validate the extracted value with `grep -qE '^0x[a-fA-F0-9]{40}$'`, (iv) sed-substitute `__L2_MESSAGE_SERVICE_ADDRESS__` into `genesis-besu.json` after the `cp` step, (v) sanity-grep the rendered file for any leftover placeholders. Maru genesis is unchanged (no address placeholders). | init.sh runs in busybox; pure POSIX sed/grep is enough — no need for `jq` or another runtime. Multiple guard clauses surface failures as clear errors instead of producing a malformed genesis that breaks Besu silently. |
| 20c | `l2-genesis-init` compose service: added `depends_on: account-setup:service_completed_successfully` and a `linea-shared-config:/shared:ro` volume mount. | init.sh now reads the precomputed JSON; account-setup must complete first to write it. |

**Phase 2.2 validation** (smoke test, simulated render in `/tmp`):
- The sed extraction pulls `L2MessageService` correctly from a JSON shaped like account-setup's output.
- Substituting both `__L2_MESSAGE_SERVICE_ADDRESS__` and `%FORK_TIME%` produces a syntactically valid JSON.
- Resulting genesis has the L2MessageService address as an alloc key with a `1000000000000000000000000000` wei balance (1B ETH).
- 30 alloc entries total — same as before, just with the L2MessageService key now coming from the precomputed JSON.
- `docker compose config -q --profile stack-partial-prover` → clean.

**Idempotency note:** The l2-genesis-init compose entrypoint still has the "skip if already rendered" guard. If `addresses-precomputed.json` changes between runs (e.g., L1 deployer nonce advanced on Sepolia), the user MUST `docker compose down -v` to force a re-render — otherwise the stale rendered genesis stays in place. Documented in init.sh's preamble.

### Phase 2.3 (config-render consumes precomputed addresses; sequencer/l2-node-besu/prover templatized), 2026-05-07

| # | Change | Why |
|---|---|---|
| 21a | Renamed three configs to `.template`: `sequencer.config.toml`, `l2-node-besu.config.toml`, `prover-config-partial.toml`. Replaced their hardcoded `0xe537...` (L2MessageService) with `__L2_MESSAGE_SERVICE_ADDRESS__`. | Per the playbook: "All Besu nodes (sequencer, archive, full): set plugin-linea-l1l2-bridge-contract = L2MessageService address." Same for prover's `message_service_contract`. |
| 21b | Rewrote `config-render`'s entrypoint. New behaviour: read `/shared/addresses-precomputed.json` (FATAL if missing) → POSIX-sed-extract `LineaRollupV8` + `L2MessageService` → substitute into 5 templates (coord, maru, sequencer, l2-node-besu, prover) → sanity-grep each rendered file for any leftover `__PLACEHOLDER__` (FATAL if found). State root, shnarf, and deploy-block remain seeded with zero defaults; Phase 2.5 will fill the first two and deploy-contracts will fill the third. | Replaces Phase-2's "boot with zeros, patch after" model with the playbook's "pre-compute → inject → boot" model for the address-related placeholders. State root + shnarf are genuinely post-L2-boot (need Shomei) so they remain in a deferred-render slot. |
| 21c | `config-render` mounts the 3 new templates + `linea-shared-config:/shared:ro` (to read precomputed JSON). Now has `depends_on: account-setup:service_completed_successfully`. | Reading from /shared requires account-setup to have run first. |
| 21d | `sequencer`, `l2-node-besu`, `prover` services rewired: bind-mount of `*.toml` swapped for `linea-rendered-config:/rendered:ro` mount; `--config-file=` (or prover's `CONFIG_FILE` env) now points at `/rendered/...`; new `depends_on: config-render:service_completed_successfully`. | Each service now boots reading the rendered TOML containing the actual precomputed L2MessageService address — no more boot-with-zeros for address-related fields. |
| 21e | `post-deploy-restart` now restarts `maru` AND `postman`. | Maru reads `/rendered/maru-config.toml` which got Phase 2.3's address substitutions at boot, BUT Phase 2.5 will add a second-pass render for genesis state root + shnarf, and deploy-contracts will patch the deploy-block — so maru still needs a restart to pick up tail patches. Postman keeps the static env_file (compose env_file reads from the host, not from a docker volume — can't render via config-render); deploy-contracts.sh sed-patches `config/l2/postman/env` and the restart cycles postman to re-read it. |
| 21f | Coordinator already templatized in Phase 2; no template change needed in 2.3, but it now reads address values from precomputed JSON instead of zero-defaults. coordinator-config.toml.template's `block-time` already bumped to `PT12S` in Phase 2 (Sepolia block time). | No churn — the existing template already had the right placeholders; only config-render's substitution logic changed. |

**Phase 2.3 validation**:
- `docker compose config -q --profile stack-partial-prover` → clean, 20 services
- All 6 placeholders defined across the 5 templates have a matching substitution in `config-render`'s sed pipeline
- Smoke-test of the substitution pipeline against a fake `addresses-precomputed.json`:
  - sequencer.config.toml renders with `plugin-linea-l1l2-bridge-contract="<real L2MS addr>"`
  - l2-node-besu.config.toml: same
  - prover-config-partial.toml: `message_service_contract = "<real L2MS addr>"`
  - coordinator-config.toml: section-correct addresses for `[protocol.l1]` (LineaRollup) and `[protocol.l2]` (L2MessageService)
  - maru-config.toml: `contract-address = "<real LineaRollup addr>"` + `l1-eth-api endpoint` set to the .env Sepolia RPC URL
  - Zero leftover `__PLACEHOLDER__` tokens in any rendered file

**Postman is intentionally not rendered via config-render** — compose's `env_file:` directive reads from the host filesystem at container-create, not from a docker volume. The deploy-contracts-sed-then-restart pattern gives postman the right values via the existing static env_file path. Documented in `config-render`'s service comment in compose.

### Phase 2.4 (deploy-contracts: pre-compute consumer + verify-or-die + Shomei state root), 2026-05-07

| # | Change | Why |
|---|---|---|
| 22a | `deploy-contracts.sh`: replaced the Phase-2 cast-based detection block (chain ID + deployer address + state root via `cast block 0 --field stateRoot` + shnarf via `cast keccak`) with two separate sections: <br/>(1) Read `/shared/addresses-precomputed.json` — extract `l1ChainId`, `deployers.l1`, all six L1 + four L2 contract addresses via POSIX sed/awk. <br/>(2) Query Shomei's `rollup_getZkEVMStateMerkleProofV0` for the L2 genesis ZK state root, then compute the genesis shnarf via `cast keccak`. | The script no longer derives anything that account-setup already computed — single source of truth. Critical correction: the L1 LineaRollup contract verifies against the **ZK state root** (Shomei) not the **MPT state root** (`eth_getBlockByNumber`); the Phase 2 implementation was wrong and would have caused proof submission failures in production. |
| 22b | New `verify_address()` helper. Case-insensitive comparison; FATAL on mismatch with a diagnostic that points users to `account-setup.sh` nonce-offset adjustment. Empty expected = skip (for intermediate / library contracts not tracked in precomputed JSON). | Per the playbook: "Verify each deployed address matches pre-computed. If mismatch → FATAL ERROR, stop." Catches deploy-script drift early — if someone adds a contract to step 1 and bumps the nonce sequence, the verify-or-die fires at step 2 instead of producing silent address mismatches downstream. |
| 22c | Inserted `verify_address` call after each of the 6 deploy steps. Step 1 verifies LineaRollupV8 + ForcedTransactionGateway. Step 2: L2MessageService. Step 3: L1 TokenBridge. Step 4: L2 TokenBridge. Step 5: L1 TestERC20. Step 6: L2 TestERC20. Validium variant skips verify (precomputed targets the ROLLUP variant). | Tight feedback loop: ANY drift between deploy-script behaviour and account-setup's pre-computation surfaces as a clean error at the exact step. |
| 22d | Simplified the post-deploy patch block. Removed the redundant maru `contract-address` patch and the coordinator `[protocol.l1] / [protocol.l2] contract-address` patches (config-render set those at boot from precomputed JSON in Phase 2.3). Kept only the genuinely-post-deploy values: coord `genesis-state-root-hash`, `genesis-shnarf`, `contract-deployment-block-number`. Postman env patch unchanged (compose env_file constraint). | Boot-time and deploy-time patches now have clean, non-overlapping responsibilities. coordinator boots once with the right values for first time. |
| 22e | `deploy-contracts` compose service: added `depends_on: shomei:service_healthy`. | Script queries Shomei before deploying L1 contracts; needs Shomei reachable. |

**Phase 2.4 validation**:
- `bash -n scripts/deploy-contracts.sh` → OK
- `docker compose config -q --profile stack-partial-prover` → no errors
- JSON extraction smoke test against a synthetic `addresses-precomputed.json`: all 9 fields (l1ChainId, l1 deployer, 6 L1 contracts, 4 L2 contracts) extracted correctly via the section-aware awk pattern (correctly disambiguates `_meta`, `deployers`, `l1`, `l2` blocks)
- Shomei state-root parse test: extracts `0x07977874...` from a sample `rollup_getZkEVMStateMerkleProofV0` response
- `verify_address` test matrix: matching (case-insensitive) → OK, mismatch → exits 1 with FATAL, empty expected → skipped silently

**Two important observations carried forward to Phase 2.5 / first-boot:**
- The Shomei `zkStateManagerVersion: "2.3.0"` parameter in the request is the protocol version, NOT the Shomei container image tag (which is `3.2.3`). If the request fails with a "version mismatch" error at first-boot, this is the place to look.
- `verify_address` for ForcedTransactionGateway is opportunistic — `extract_address` returns empty if step 1 didn't emit that contract (e.g., in the Validium variant). The verify call is wrapped in a guard so the script doesn't die in cases where the contract simply wasn't deployed.

### Phase 2.5 (verification: boot-order audit + render-pipeline smoke test + maru-restart correction), 2026-05-07

Phase 2.5 was scoped as either a separate `render-2` init container OR a verification pass. Decision: collapse to verification — Phase 2.4's in-place patch of coord-config (state root + shnarf + deploy-block) is sufficient because coordinator depends on `deploy-contracts:completed`, so it never sees pre-patch values.

**What was verified**

| Audit | Result |
|---|---|
| Boot-order DAG | Acyclic. 21 services on `stack-partial-prover`. account-setup runs first; init layer (config-render, l2-genesis-init, web3signer) reads its outputs; service layer (sequencer, maru, l2-node-besu, shomei) reads /rendered/; deploy-contracts runs after sequencer healthy + shomei healthy + config-render done; coordinator/postman/prover/post-deploy-restart wait for deploy-contracts:completed. |
| Placeholder coverage | All 6 templated `__PLACEHOLDER__` tokens (across 5 TOML templates + the genesis JSON template) have a matching substitution. config-render handles `__L1_RPC_URL__`, `__LINEA_ROLLUP_ADDRESS__`, `__L2_MESSAGE_SERVICE_ADDRESS__`, `__GENESIS_STATE_ROOT_HASH__`, `__GENESIS_SHNARF__`, `__LINEA_ROLLUP_DEPLOY_BLOCK__`. init.sh handles `__L2_MESSAGE_SERVICE_ADDRESS__` + `%FORK_TIME%` for genesis. |
| Volume produce/consume | `linea-shared-config`: producers = account-setup (rw), deploy-contracts (rw); consumers = l2-genesis-init, config-render, web3signer, coordinator, postman, prover. `linea-rendered-config`: producers = config-render (rw, full render), deploy-contracts (rw, in-place patch of coord-config only); consumers = sequencer, maru, l2-node-besu, coordinator, prover. No write conflicts (config-render writes 5 distinct files; deploy-contracts only patches coord-config). |
| End-to-end render pipeline | Single-shot smoke test simulating account-setup → config-render → l2-genesis-init → deploy-contracts post-patch. 12/12 assertions pass. Zero leftover placeholders in all 6 rendered output files. Genesis L2MessageService allocated at 1B ETH; coord state-root + shnarf + deploy-block all patched correctly; sequencer, l2-node-besu, prover all get the precomputed L2MS address; maru gets LineaRollup address + Sepolia L1 RPC URL from `.env`. |
| Validation commands | `docker compose config -q --profile stack-partial-prover`: clean. `bash -n` on `account-setup.sh` + `deploy-contracts.sh`: OK. `sh -n` on `init.sh`: OK. |

**One correction surfaced by the audit**

| # | Change | Why |
|---|---|---|
| 23a | `post-deploy-restart`: dropped `maru` from the restart list. Now restarts only `postman`. | Phase 2.3 added maru on the assumption that deploy-contracts would patch /rendered/maru-config.toml. Phase 2.4 simplified the post-deploy patch — deploy-contracts no longer touches maru-config (config-render set everything correctly at boot from precomputed JSON). Maru's first-boot config is already correct, so restart is wasted work. Postman stays in the restart list because compose's `env_file` directive may cache file content at parse time and the deploy-contracts sed-patch happens later. |

**Open carry-overs into later phases**

- **Phase 2.6** — drop `stack-no-prover` profile (user constraint: "stack-no-prover is dead, not a deliverable"). Single profile = `stack-partial-prover`.
- **Phase 3** — finalize web3signer alignment. Already mostly done in Phase 2.1 (account-setup writes 3 keystores from `L1_DEPLOYER_PRIVATE_KEY`); this phase removes the now-vestigial static keystore files in `config/web3signer/key-files/` and updates the postman env's `L1_SIGNER_PRIVATE_KEY` to also derive from `L1_DEPLOYER_PRIVATE_KEY` (currently still a Hardhat dev key in the static env file — deploy-contracts.sh would need to patch this).
- **Phase 4** — Sepolia timing tunables: postman `L1_LISTENER_INTERVAL` 500 → 5000ms; `L1_LISTENER_INITIAL_FROM_BLOCK` 0 → deploy block (must come from deploy-contracts post-deploy patch); coord `consistent-number-of-blocks-on-l1-to-wait` 1 → 5. Plus README rewrite for the Sepolia flow.
- **Phase 5** — first-boot validation against a real Sepolia RPC + funded deployer key. Catches the things only a real run can: Foundry image version compatibility (cast flag names), Shomei `zkStateManagerVersion` value, V8 LineaRollup shnarf formula match, prover happy with `verifier_id = 0` against the IntegrationTestTrueVerifier on Sepolia.

### Phase 2.6 (drop `stack-no-prover` profile), 2026-05-07

| # | Change | Why |
|---|---|---|
| 24a | `docker-compose.yml`: redefined the YAML anchor `*default-profiles` from `["stack-no-prover", "stack-partial-prover"]` to `["stack-partial-prover"]`. Every service that uses the anchor inherits the change. | Single source of truth for the profile name; no per-service edits needed. |
| 24b | `docker-compose.yml`: prover service's explicit `profiles: ["stack-partial-prover"]` swapped to `profiles: *default-profiles`. The "differs between profiles" comment is gone — there's only one profile now. | Cosmetic — the explicit single-element list and the anchor now resolve to the same thing. Anchor wins for consistency. |
| 24c | Header banner in compose: explicitly states `stack-partial-prover` is the only profile, that the partial prover requires ~32 GB host RAM (GOMEMLIMIT 32GiB), and that running without the prover is no longer supported. | Surfaces the resource requirement at the top of the file so anyone reading the compose source sees it before trying to bring the stack up. |
| 24d | `config/observability/prometheus.yml`: comment for the prover scrape job updated. The "missing target acceptable when running stack-no-prover" caveat is gone — prover is always present. | Avoid confusing "target down" alarms in the future from an outdated comment. |
| 24e | `README.md`: replaced every command-line `--profile stack-no-prover` with `--profile stack-partial-prover` (4 occurrences). Added a top-of-file banner flagging the README narrative as out-of-date pending Phase 4 rewrite — the descriptive references to `stack-no-prover` (profile bullet, hardware table column, section title, Apple Silicon recommendation) are intentionally left in place under the banner. | Anyone running through the README hits a working command. The descriptive content needs a full rewrite for the Sepolia flow (Phase 4) — surgical fixes to half of it would produce a worse document. |
| 24f | `SCAFFOLD-PLAN.md` left unchanged — historical planning artifact predating the Sepolia migration; references to `stack-no-prover` there are correct in their original temporal context. | Don't rewrite history in planning docs. |

**Phase 2.6 validation**:
- `docker compose --profile stack-partial-prover config --services | wc -l` → 21 (full stack)
- `docker compose --profile stack-no-prover config --services | wc -l` → 0 (dead profile)
- `docker compose config --services | wc -l` (no profile) → 0 (everything gated)
- `docker compose --profile stack-partial-prover config -q` → no errors
- prover service still resolves to the partial-prover profile

**Single-profile invocation now is**:
```bash
docker compose --env-file versions.env --env-file .env --profile stack-partial-prover up -d
```

### Phase 3 (web3signer keystore + postman L1 signer alignment cleanup), 2026-05-07

| # | Change | Why |
|---|---|---|
| 25a | Trashed `config/web3signer/key-files/` (4 vestigial YAMLs from before account-setup wrote them at boot) and `config/web3signer/conf/config.yaml` (unused — web3signer reads its config purely from CLI args, not from a file). | Phase 2.1 made these files dead code. Leaving committed-but-unread Hardhat keys in the repo is a security paper-cut, not actively dangerous, but cleanly removable. |
| 25b | Postman compose `environment:` block: added `L1_SIGNER_PRIVATE_KEY: ${L1_DEPLOYER_PRIVATE_KEY:?...}`. Postman now signs L1 transactions with the user's deployer key (Option A — same address as the L1 contract deployer, security council, rollup operators, web3signer-backed signers). | Closes the last hole where a Hardhat dev key was still being used for an L1 signing role. Single key for everything on L1. |
| 25c | `config/l2/postman/env`: commented out the `L1_RPC_URL` and `L1_SIGNER_PRIVATE_KEY` literals (they're shadowed by compose env). Replaced `L1_CONTRACT_ADDRESS=0xDc64...` with `0x0000…0000` placeholder (deploy-contracts.sh's sed-patch pattern still matches). Added a header comment explaining what each commented-out line maps to in the runtime. | Static file no longer contains misleading dev-key values. Anyone reading the file sees what's static-vs-runtime at a glance. |
| 25d | `config/DEV-KEYS-INVENTORY.md`: rewritten. Drops rows for the deleted L1 + web3signer files. Adds a top section explaining the Sepolia split: L2 dev keys (committed) vs L1 keys (user-supplied) vs runtime-rendered web3signer keystores (in /shared, ephemeral). The "non-loopback deployment" replacement checklist is collapsed from 8 steps to 4 since most of the L1-side regeneration is no longer applicable (user supplies their own L1 key already). | The document was the canonical inventory of "what's checked into the repo as a dev key"; it had to track the actual repo state. |
| 25e | README.md: trimmed the DEV-ONLY warning banner (lines pointing at deleted `config/l1/*`, `config/web3signer/key-files/*`, `seed-funds.sh`, mnemonic file). Added a line about the runtime-rendered web3signer keystores (still security-relevant — they hold the user's L1 key inside the docker volume). The rest of the README narrative remains under the Phase-2.6 "rewrite pending" banner. | Pointing security warnings at non-existent files is worse than removing them; the actual sensitive surface area is now smaller and clearer. |
| 25f | `docker-compose.yml` (web3signer service comment) + `scripts/account-setup.sh` (comments): removed references to the now-deleted `config/web3signer/key-files/` directory. | Stale references in comments produce confusion. |

**Phase 3 validation**:
- `docker compose config -q --profile stack-partial-prover` → no errors
- Resolved postman environment shows `L1_SIGNER_PRIVATE_KEY: 0xac0974…ff80` (the test L1 deployer key passed in `.env`); `L1_RPC_URL: https://example.com` (from `.env`); `L1_CONTRACT_ADDRESS: 0x0000…0000` (placeholder, will be sed-patched by deploy-contracts)
- `bash -n` on both scripts: OK
- `config/web3signer/key-files/` and `config/web3signer/conf/` no longer exist
- `config/web3signer/tls-files/` (mTLS keystore + password + known-clients) preserved — still actively mounted by web3signer

**No L1 Hardhat keys remain anywhere on the L1 path.** Every L1 transaction (contract deploy, blob submission, finalization, anchoring, postman messages) is signed by the user's `L1_DEPLOYER_PRIVATE_KEY` from `.env`, either directly (postman, deploy-contracts) or via web3signer (coordinator). L2 dev keys remain committed by design — we own L2 genesis.

### Phase 4 (Sepolia timing tunables + README rewrite), 2026-05-07

| # | Change | Why |
|---|---|---|
| 26a | `coordinator-config.toml.template`: `consistent-number-of-blocks-on-l1-to-wait` 1 → 5. | Sepolia has occasional shallow reorgs; 5 blocks (~60s) gives enough confidence the L1 head won't reorg under us. The local devnet's 1-block setting (12s on Sepolia) was effectively no protection. |
| 26b | `config/l2/postman/env`: `L1_LISTENER_INTERVAL` 500 → 5000ms. | Sepolia produces a block every ~12s; 500ms polling burns RPC quota for no benefit. 5s polling is 2.4 polls per block — still plenty of headroom. |
| 26c | `deploy-contracts.sh`: extracts L1 deploy block from `step1-linea-rollup.log` (same regex as the existing L2 extraction). sed-patches `L1_LISTENER_INITIAL_FROM_BLOCK` in `config/l2/postman/env` alongside the existing `L1_CONTRACT_ADDRESS` + `L2_CONTRACT_ADDRESS` patches. | Without this patch, postman's `L1_LISTENER_INITIAL_FROM_BLOCK=0` would scan all of Sepolia from genesis (~5M blocks) on first boot. With the patch, it starts at the LineaRollup deploy block — same idea as the coord's `contract-deployment-block-number`. |
| 26d | `README.md`: full rewrite. Old narrative (local Besu+Teku L1, dual profiles, "v1 will switch to Sepolia" forward-looking notes, `[STAGE-2 placeholder]` blocks, seed-funds references) replaced with the actual Sepolia flow: Setup → Boot → Boot-timeline → Endpoints → Verifying → Tearing down → Customisation → Troubleshooting → "What's not in v0" → Reference. ~400 lines (was 313). | Anyone reading the README now sees the actual current behaviour, not the pre-migration narrative. The "Status" banner up-front makes clear Phase 5 (real Sepolia first-boot validation) hasn't yet happened, so users running through the README are partly the validators. |

**Phase 4 validation**:
- `docker compose config -q --profile stack-partial-prover` → no errors
- Resolved postman env: `L1_LISTENER_INTERVAL=5000`, `L1_LISTENER_INITIAL_FROM_BLOCK=0` (will be patched at deploy time), `L1_RPC_URL` and `L1_SIGNER_PRIVATE_KEY` both correctly overridden from `.env`
- Coordinator template: `consistent-number-of-blocks-on-l1-to-wait = 5`
- `bash -n` on deploy-contracts.sh: OK
- README: zero remaining `stack-no-prover` references

**Sepolia migration is now complete through Phase 4.** Phase 5 — first-boot validation against a real funded Sepolia deployer key + RPC URL — is the only thing left. Pre-flight checklist for Phase 5:

1. User provides `L1_RPC_URL` (Sepolia HTTPS) + `L1_DEPLOYER_PRIVATE_KEY` (Sepolia-funded, ~0.5 ETH)
2. `docker compose --env-file versions.env --env-file .env --profile stack-partial-prover up -d`
3. Watch each phase: account-setup → genesis + config-render → service health → deploy-contracts → coord/postman/prover up → first proof submission
4. Append findings to this fix log as fix #27+ (Phase-5 validation findings)

Likely first-boot surprises (in priority order — see prior fix entries for context):
- Foundry image's `cast` flag-name drift (compute-address, wallet address)
- Shomei `zkStateManagerVersion: "2.3.0"` rejection or unexpected response shape
- V8 LineaRollup shnarf formula divergence (current formula matches V6 comment in coord template)
- Hardhat-foundry plugin failure if `forge --version` errors during deploy-contracts pnpm install
- `cp -T` semantics on busybox vs GNU (used in `init.sh`)
- Sepolia gas-price spikes blowing through 0.5 ETH on the deployer
- Apple Silicon prover OOM on first proof if Docker Desktop memory cap < 48 GB

### Phase 5 pre-flight: prover all-dev mode for IntegrationTestTrueVerifier flow, 2026-05-07 (PARTIALLY REVERTED 2026-05-07 — see #36)

| # | Change | Why |
|---|---|---|
| 27 | `config/l2/prover/prover-config-partial.toml.template`: contents replaced verbatim with the canonical `docker/config/prover/v3/prover-config.toml` (all four `prover_mode` set to `"dev"`; `is_allowed_circuit_id` 483 → 963; partial-mode tuning fields `conflated_traces_dir`, `ignore_compatibility_check`, `serialization` dropped). The `__L2_MESSAGE_SERVICE_ADDRESS__` placeholder is preserved. | At the time, the goal was to exercise the FULL pipeline (conflation → compression → execution → aggregation → on-chain submission → on-chain verify) end-to-end without burning hours of real ZK math per proof. The intent was correct (use dev for fast iteration); the implementation was wrong (silently overwrote the partial template, leaving the `--profile stack-partial-prover` name lying about what it does). |

**Schema correction surfaced before applying**: the user's instruction named TOML fields (`[prover] type = "development"`, `dev_mode = true`) that don't exist in the prover schema. A grep across the entire monorepo's `*.toml`/`*.yaml`/`*.json` returned **zero** matches for those exact names. The actual schema uses `prover_mode = "<value>"` per section, with valid values `bench | dev | full | limitless | partial`.

**What was wrong with #27** (caught 2026-05-07 evening when the user asked what prover version was running and noticed the dev mode under a "partial" profile name):
- The justification "Victorien said you can't mix dev and real prover modes" was **wrong as stated**. Upstream's `docker/config/prover/v3/prover-config-partial.toml` literally ships with execution+invalidity in `partial` and data_availability+aggregation in `dev`, and the L1 `IntegrationTestTrueVerifier` accepts those proofs. The "no mixing" rule was either misremembered or applied out of context.
- The deliverable for v0 is **partial-prover** behaviour — that's why the profile is named `stack-partial-prover`. Replacing the template with all-dev as the *default* shipped a config that contradicts the profile name.

**Correction applied in #36** (Phase 6 add-on, see below): template restored to upstream partial verbatim. The all-dev config is documented as a *fast-iterate override* in the template's header + the README "Switching prover mode" section, not the default.

**Boot sequence (final shape, post-Phase-2.5)**

```
[Init layer]
account-setup ─────────────► /shared/{addresses-precomputed.json, web3signer-keys/}
    ├── config-render ─────► /rendered/{coordinator,maru,sequencer,l2-node-besu,prover}-config.toml
    ├── l2-genesis-init ──► /initialization/{genesis-besu.json,genesis-maru.json,fork-timestamp.txt}
    └── web3signer (mTLS, 3 L1 + 1 L2 signers) ───► serves signing requests

[Service layer]
sequencer (l2-genesis-init + config-render) ► l2-node-besu + maru ► shomei
                                                                       │
[Deploy + post-deploy]                                                 ▼
deploy-contracts (queries Shomei for state root → deploys 6 steps → verify-or-die per step
                  → patches /rendered/coordinator-config.toml + config/l2/postman/env)
    ├── coordinator (reads patched /rendered/coordinator-config.toml)
    ├── postman (reads patched static config/l2/postman/env; restarted by post-deploy-restart)
    ├── prover (depends on coordinator)
    └── post-deploy-restart (restarts postman)
```

### Phase 5 boot fixes against real Sepolia, 2026-05-07

| # | Change | Why |
|---|---|---|
| 28 | `account-setup.sh`: `L1_LINEA_ROLLUP` offset +5 → +4. | Empirical Sepolia test: Hardhat's `deployPlonkVerifierAndLineaRollupV8.ts` assigns nonce 4 to the LineaRollup proxy and nonce 5 to Mimc (despite emitting Mimc 5th in stdout). Verified via `cast compute-address` against the actually-deployed proxy. |
| 29 | Forked `deployBridgedTokenAndTokenBridgeV1_1.ts` into `scripts/`, bind-mounted over `/workspace/contracts/local-deployments-artifacts/` in `deploy-contracts` compose service. (a) drops the stale `ORDERED_NONCE_POST_LINEAROLLUP = 7` arithmetic and uses `await wallet.getNonce()` directly; (b) serialises the 3 implementation deploys (no `Promise.all`, no explicit `nonce:` overrides); (c) takes `REMOTE_TOKEN_BRIDGE_ADDRESS` from env instead of deriving it via `remoteDeployerNonce + 4` (the same stale-offset arithmetic). `deploy-contracts.sh` step 3 supplies the precomputed L2 TokenBridge; step 4 supplies the L1 TokenBridge just deployed. | Step 1 actually consumes 8 nonces (7 deploys + 1 `grantRole(FORCED_TRANSACTION_SENDER_ROLE)`), not the 7 the upstream constant assumes. With the +7 offset, step 3's BridgedToken deploy gets `nonce too low: next nonce N+8, tx nonce N+7` against any non-fresh Sepolia deployer. The remote-sender derivation hit the same bug, silently producing the wrong cross-chain TokenBridge `remoteSender` for L2's init. |
| 30 | `account-setup.sh`: shifted post-step-1 L1 offsets — `L1_BRIDGED_TOKEN` +7 → +8, `L1_TOKEN_BRIDGE` +10 → +12, `L1_TEST_ERC20` +11 → +13. Also shifted step-3-relative L2 offsets — `L2_TOKEN_BRIDGE` +6 → +7, `L2_TEST_ERC20` +7 → +8. Updated the file's nonce-sequence comment to call out that step 3 (and step 4) deploys 5 contracts, not 4 (the previous comment forgot the `ProxyAdmin`). | Compounding the fix above: account-setup's pre-computed addresses depended on the same stale "step 1 = 7 nonces" assumption AND missed the `ProxyAdmin` in step 3/4, leaving `L1_TOKEN_BRIDGE` precomputed at the wrong CREATE address. Verify-or-die would have caught it eventually but only after burning Sepolia ETH on the deploy. |
| 31 | `deploy-contracts.sh`: added `step_already_done` helper + idempotency guards at the top of all six steps (skip the deploy if `/shared/deploy-logs/stepN-*.log` exists with the expected `contract=NAME deployed:` line; just re-extract the address). Also ran `verify_address` from inside the skip-branch so a stale precomputed JSON still surfaces. | Without this, restarting `deploy-contracts` after a partial failure (e.g. step 5 nonce error) would re-fire steps 1-4 against the now-advanced wallet nonce, deploying NEW contracts at NEW addresses and breaking the precomputed-address verify chain. Costs ~0.10 ETH per redo on Sepolia. The shared volume preserves logs, so step state survives container restart. |
| 32 | `account-setup.sh`: added an idempotency exit at the top — `if [ -f /shared/addresses-precomputed.json ]; then exit 0; fi`. To force fresh derivation: `down -v`. | When `deploy-contracts` is restarted, compose's dependency chain re-fires `account-setup` too. Without idempotency, `account-setup` queries the wallet's CURRENT nonce (now advanced by prior partial deploys), regenerates precomputed addresses against the new baseline, and breaks verify against contracts already deployed at the OLD baseline. Hit this once on the Phase-5 retry; restored the original JSON manually before re-running. |
| 33 | `deploy-contracts.sh` steps 5 + 6: prefixed the `pnpm exec ts-node deployTestERC20.ts` invocation with `env -u L1_NONCE` (step 5) / `env -u L2_NONCE` (step 6). | `deployTestERC20.ts` carries the same stale `ORDERED_NONCE_POST_LINEAROLLUP = 7` constant. Unsetting the env var falls through to `wallet.getNonce()`. Cheaper than another fork. |
| 34 | `docker-compose.yml` `prover` service: bind-mounted `../../../prover/prover-assets:/opt/linea/prover/prover-assets:ro`. | The prover binary checks for `kzgsrs/` at startup even in `prover_mode = "dev"`. Without the mount it crashes immediately with `kzgsrs directory does not exist`. The upstream `compose-spec-l2-services.yml` has the identical mount; the scaffold dropped it accidentally during Phase 1 service trim. |

**Phase 5 end state (2026-05-07T16:18 UTC)**:

```
account-setup    ✓ derived L1+L2 deployers, web3signer keystores, precomputed addresses
config-render    ✓ rendered 5 templates with substituted addresses
l2-genesis-init  ✓ wrote genesis-besu.json + genesis-maru.json
sequencer        ✓ healthy
l2-node-besu     ✓ healthy
maru             ✓ running
shomei           ✓ healthy (returned ZK genesis state root for deploy-contracts)
web3signer       ✓ serving 4 keystores
deploy-contracts ✓ all 6 steps verified
                   - LineaRollupV8 (L1)
                   - ForcedTransactionGateway (L1)
                   - L2MessageService (L2)
                   - L1 TokenBridge
                   - L2 TokenBridge (via forked sequential script)
                   - L1 + L2 TestERC20
postman          ✓ polling for L2 messages
coordinator      ✓ DB migrated, conflation resuming from block 1
prover           ✓ polling for jobs (kzgsrs mounted)
```

Sepolia ETH spent across 3 attempts: ~0.30 ETH (one full success after two partial failures).

### Phase 6 (open): coordinator/web3signer key alignment + post-startup silence, 2026-05-07

After all 6 deploys landed and verified on Sepolia, the user asked whether any L1↔L2 user txs round-tripped. They had not — only the 14 deploy txs were on chain. Investigation surfaced two distinct issues; the first is **fixed**, the second is **open**.

**Issue A — fixed: coordinator/web3signer key mismatch** (#35)

| # | Change | Why |
|---|---|---|
| 35 | (a) `account-setup.sh`: derive uncompressed secp256k1 pubkey from `L1_DEPLOYER_PRIVATE_KEY` via `cast wallet public-key`; emit it as `signers.l1Pubkey` in `addresses-precomputed.json`. (b) `coordinator-config.toml.template`: replace the 3 hardcoded upstream dev pubkeys (data-submission `0x9d9031…`, finalization `0xba5734…`, anchoring `0x4a788a…`) with `__L1_SIGNER_PUBKEY_NO_PREFIX__`. (c) `docker-compose.yml` config-render: read `signers.l1Pubkey`, strip `0x`, substitute into all rendered configs. | The coordinator-config carried over upstream's hardcoded web3signer pubkeys, but Option A wires all 3 keystores to the user's `L1_DEPLOYER_PRIVATE_KEY`. When coordinator asked web3signer to sign with pubkey `0x9d9031…`, web3signer would return "key not found" and silently never submit blob/finalization/anchoring txs. With this patch, all 3 slots resolve to the same single user-funded address. |

Verified via the coordinator's `App configs` dump at startup — all three signer slots show the same `publicKey=0x...` value, matching the pubkey derived from the user's `L1_DEPLOYER_PRIVATE_KEY` via `cast wallet public-key`. (Actual pubkey value redacted from this doc — it's specific to whichever wallet the running operator funded; reproduce by running `cast wallet public-key --private-key $L1_DEPLOYER_PRIVATE_KEY`.)

The fix landed against the existing volume without needing a `down -v`: manually patched `signers.l1Pubkey` into the persisted `addresses-precomputed.json`, then `up -d --force-recreate config-render` re-rendered with the new substitution, then `up -d --force-recreate --no-deps deploy-contracts` re-applied the deploy-time patches (state_root, shnarf, deploy_block) that re-render zeroed out, then `up -d --force-recreate --no-deps coordinator postman prover` to pick up the new TOML. Saved another ~0.10 ETH that a full re-deploy would have burned.

**Issue B — historical open item: coordinator silent after startup**

> Historical note: this was the live state on 2026-05-07 and is fixed later by #37.

After the pubkey fix landed and coordinator restarted cleanly, it logged `Coordinator app instantiated` and then went **completely silent**. 7+ min of zero output. JVM alive (PID 1, ~11s CPU, idle). Symptoms:
- L2 chain advancing (sequencer at block 398; multiple test txs sent via `cast send` to generate traffic).
- Coordinator NOT polling sequencer for new blocks (no log even at INFO level).
- No prover requests written (`/data/prover/v3/{execution,compression,aggregation}/requests/` all empty).
- Coordinator's `:9545` (observability) and `:9546` (json-rpc) ports don't bind from inside the linea network (`curl http://coordinator:9545` → connection refused). Either not yet started, or bound to localhost only.
- Container marked "unhealthy" — but this is a **separate red herring**: the healthcheck uses `curl` which doesn't exist in the linea-coordinator image. Healthcheck failure ≠ app failure.

Hypotheses to check next session:
1. **`log4j2-dev.xml` suppresses INFO post-startup**. The container CMD passes `-Dlog4j2.configurationFile=/var/lib/coordinator/log4j2-dev.xml`. If that config sets specific logger thresholds higher than INFO, periodic activity (block polling, conflation ticks) would be silent. Check the upstream `log4j2-dev.xml` content vs the prod log4j2.
2. **`L1FinalizationMonitor` is blocked**. It logged `Rollup finalized block updated from null to 0, waiting 5 blocks for confirmation` and then nothing. With `consistentNumberOfBlocksOnL1ToWait=5`, the monitor needs 5 Sepolia blocks of confirmation depth on the LineaRollup state read. Sepolia blocks are 12s each, so 60s — should have completed long ago. But if the L1FinalizationMonitor is gating downstream pipelines, a stuck monitor would explain the silence.
3. **Traces API mismatch**. `expectedTracesApiVersion=beta-v5.0-rc3` is hardcoded in our coordinator config. The `linea-besu-package` we run for the sequencer might serve a different traces API version. A version mismatch on first poll could throw at INFO without a stack-trace, then... no wait, that should at least log an error.
4. **The `--no-deps` recreate cycle left coordinator in a half-initialised state**. Try a clean `down`/`up` cycle (preserve volumes — only stop+restart containers) and re-observe.

For session resumption: leave the stack running, log file persists, can re-check `docker logs --since 2026-05-07T17:00 coordinator` against any reference build.

**Status as of 2026-05-07T17:05Z**:
- ✅ All 6 deploys on Sepolia, all verified.
- ✅ Pubkey alignment landed cleanly across template + config-render.
- ❌ Zero L1↔L2 user txs (no LineaRollup blob submissions, no message anchoring, no L1↔L2 message-passing demos).
- ❌ Coordinator pipeline silent after startup — root cause TBD, see hypotheses above.
- ⏸  L2->L1 / L1->L2 message round-trip: blocked on coordinator pipeline working.

**Issue C — fixed: prover template lied about being partial** (#36)

| # | Change | Why |
|---|---|---|
| 36 | `config/l2/prover/prover-config-partial.toml.template`: contents restored to the upstream `docker/config/prover/v3/prover-config-partial.toml` verbatim — execution + invalidity in `prover_mode = "partial"` (with `conflated_traces_dir = "/"` + `ignore_compatibility_check = true` + `serialization = false` re-added), data_availability + aggregation in `prover_mode = "dev"`, `is_allowed_circuit_id = 483`. The `__L2_MESSAGE_SERVICE_ADDRESS__` placeholder is preserved. The all-dev variant from #27 is now documented as a *fast-iterate override* in the template's header comment + the README "Switching prover mode" section. | The `--profile stack-partial-prover` name says "partial prover" but #27 silently swapped the contents to all-dev. Caught when the user noticed the dev mode running under a partial-named profile. The correct shape: deliverable = partial (matches profile name), dev mode = clearly-documented local override only. |

**Live runtime impact (intentional)**: the live `/rendered/prover-config-partial.toml` in the docker volume was NOT updated — it stays on the all-dev rendering from #27 because the user is still iterating on the Phase-6 coordinator-stuck issue and wants the faster cycle. A future `down -v` (or `up -d --force-recreate config-render` after editing nothing) will render the partial-mode template into `/rendered/`. To stay on dev after such a re-render, apply the override per the README.

**Verification**:
- Template now has 4 prover_mode lines: `partial`, `dev`, `partial`, `dev` (matching upstream partial).
- Live rendered file in volume still has 4× `dev` + `is_allowed_circuit_id = 963` (untouched).
- Prover container kept running through this change with no restart.

**Issue B follow-up — fixed: coordinator startup blocked by L1 RPC `eth_getLogs` limits** (#37)

The "silent after startup" symptom was not log4j. A thread dump showed the app blocked in `CoordinatorApp.start()` before the API ports were opened. The blocking call was the L1 finalization monitor's startup read: for a V8 rollup it calls `CONTRACT_VERSION`, `currentL2BlockNumber`, then a broad `eth_getLogs` from `earliest` to `latest` for `FinalizedStateUpdated(0)`.

The original free-tier L1 RPC rejected that broad log query with a 10-block-range limit. Coordinator retries this path indefinitely, so it never reaches `Conflation started` or binds the observability/API ports. Replacing the L1 RPC with a provider that supports the required `eth_getLogs` call unblocked startup.

**Issue D — fixed for fresh boots: coordinator catch-up can outrun Besu history** (#38)

After the RPC fix, coordinator started polling and requested `linea_generateConflatedTracesToFileV2` from block `1..2` upward. Because the stack had been left running while coordinator was blocked, L2 had advanced thousands of blocks. `l2-node-besu` was configured with `bonsai-historical-block-limit=1024`, so old block ranges could no longer be traced and Besu returned `Plugin internal error: Conflation not finished`.

Direct probes confirmed recent ranges traced successfully while old ranges failed. The quickstart now raises `bonsai-historical-block-limit` to `100000`, giving first-boot recovery and coordinator catch-up far more room before old Bonsai state is unavailable.

**Security cleanup — fixed: config-render leaked RPC URLs** (#39)

`config-render` used to print the full `L1_RPC_URL`, which can include provider API keys. It now logs only that the URL is set.

**Security/retry cleanup — fixed: deploy preflight leaked RPC URLs and timed out too aggressively** (#40)

The deploy-contracts service used to pass `L1_RPC_URL` as a positional command argument, making provider URLs visible in container process listings. Its pre-flight RPC loop also logged the full URL on failure and used a 2s timeout. The service now reads all deploy inputs from environment variables, and `wait_rpc` logs only the logical RPC name while requiring a valid `eth_blockNumber` JSON-RPC result with a longer timeout window.


**Observability/security cleanup — fixed: deploy/runtime logs expose key milestones without secrets** (#41)

Fresh boot retry on 2026-05-08 confirmed Sepolia contracts were created and coordinator progressed past the prior silent startup point: `deploy-contracts` exited `0`, `/shared/addresses.json` was written, coordinator bound `9545`/`9546`, and prover request files appeared for L2 block ranges starting at `1-2`.

The quickstart now includes `scripts/status.sh`, a redacted milestone view over container status, deploy markers, final contract addresses, coordinator ports, and prover request counts. This gives users a single command for the critical boot boundary instead of requiring manual Docker-volume inspection.

Security cleanup in the same pass:

- `contracts/common/helpers/environment.ts` now redacts required env var values whose names look secret-bearing (`PRIVATE_KEY`, `SECRET`, `TOKEN`, `PASSWORD`, `RPC_URL`) before logging.
- `account-setup.sh` no longer logs the full L1 RPC URL and writes `<redacted>` in `addresses-precomputed.json` metadata.
- `deploy-contracts.sh` no longer passes the L1 RPC URL into the address-aggregation Node process and writes `<redacted>` in `addresses.json` metadata.


**Issue E — fixed for quickstart defaults: partial execution prover OOM under laptop Docker limits** (#42)

The 2026-05-08 fresh boot proved coordinator was no longer stuck: it bound `9545`/`9546`, generated execution/compression requests, and compression proofs completed. The next blocker was prover execution requests exiting `137`, leaving files with `.large.failure.code_137` and zero execution responses. Docker Desktop was capped around 8 GiB, while partial execution proving needs substantially more memory.

For quickstart usability, `PROVER_DEV_OVERRIDE` now defaults to `true`, making config-render patch the rendered prover config to all-dev mode. The partial template remains on disk and can be selected with `PROVER_DEV_OVERRIDE=false` plus a larger Docker memory cap. `scripts/status.sh` now reports failed/in-progress prover request counts so this failure mode is visible immediately.


**Runtime state cleanup — fixed: postman env no longer gets sed-patched in the repo** (#43)

`deploy-contracts.sh` previously patched `config/l2/postman/env` on the host with live contract addresses and the L1 deploy block. That made every boot dirty the git worktree with run-specific Sepolia addresses. Postman now reads `/shared/addresses.json` and the step-1 deploy log at container startup, exports `L1_CONTRACT_ADDRESS`, `L2_CONTRACT_ADDRESS`, and `L1_LISTENER_INITIAL_FROM_BLOCK`, then starts the Node app. The `post-deploy-restart` docker-socket helper is no longer needed and was removed.


**Healthcheck cleanup — fixed: coordinator no longer reports false unhealthy** (#44)

The coordinator image does not include `curl`, so the compose healthcheck reported `unhealthy` even when ports `9545` and `9546` were listening and the prover pipeline was active. No service depended on `coordinator:service_healthy`, so the quickstart now removes that healthcheck and relies on `scripts/status.sh` for the meaningful coordinator readiness checks.


**Current checkpoint — coordinator/prover active, L1 submissions observed** (#45)

Fresh Sepolia boot on 2026-05-08 reached the first real bring-up milestone: all contract steps completed, `addresses.json` was written, coordinator opened its observability/JSON-RPC ports, the dev prover produced responses, and coordinator emitted L1 submission transaction hashes for blob and aggregation paths.

This should be documented as progress, not final success. The acceptance bar is now higher: a repeatable run should show non-zero prover responses, L1 submissions without persistent nonce contention, and a documented user-facing bridge/message smoke test.

**Historical caveat — single L1 key caused noisy coordinator submission retries** (#46)

The first working Sepolia boot used one funded key for every L1 role. That kept setup simple, but coordinator could attempt blob/data-submission and aggregation/finalization work concurrently through the same account. The live run still progressed, but logs showed repeated `address already reserved` errors during L1 submission.

This is superseded by #46b, which splits coordinator's L1 submitter roles while keeping the user-facing setup to one Sepolia-funded deployer key.

**Signer role split — wired for fresh first boots, pending live validation** (#46b)

The upstream e2e flow does not run coordinator L1 submissions through one account. It grants rollup operator roles to separate relayer/operator accounts: one for blob/data-submission and one for aggregation/finalization. Message anchoring is also distinct and uses the L2 message anchorer role, not the user's Sepolia deployer.

The quickstart now mirrors that shape while preserving the one-secret UX. During first boot, `account-setup.sh` derives two deterministic L1 submitter keys from `L1_DEPLOYER_PRIVATE_KEY`, keeps the pre-baked L2 message-anchoring key for the L2 role, writes all four web3signer keystores, and records the signer addresses/pubkeys in `/shared/addresses-precomputed.json`. `config-render` then injects the matching pubkeys into coordinator config. `deploy-contracts.sh` grants rollup operator roles to the deployer plus both derived L1 submitters, grants the L2 message setter role to the L2 anchorer, and funds the two derived L1 submitters before coordinator starts.

This is designed to fit the first-time boot sequence without asking users to manage three Sepolia keys. Existing Docker volumes from the single-key era will not contain the new signer metadata; run `docker compose --env-file versions.env --env-file .env --profile stack-partial-prover down -v` before validating this change on an old checkout.

**L2 Blockscout frontend — fixed** (#47)

Added `l2-blockscout-frontend` using the official Blockscout frontend image, pinned by digest in `versions.env`, and exposed it on `http://localhost:4001`.

The frontend uses `NEXT_PUBLIC_USE_NEXT_JS_PROXY=true` and points `NEXT_PUBLIC_API_HOST` at the Docker service name `l2-blockscout:4000`, so server-side startup work and browser API traffic both reach the existing backend. Optional marketplace/ad/gas/web3 widgets are disabled to keep the local explorer minimal.

Verification on the live stack: frontend env validation passed, sitemap generation fetched `/api/v2/{addresses,transactions,blocks,tokens,smart-contracts}` successfully from `l2-blockscout:4000`, `curl -I http://127.0.0.1:4001` returned `200 OK`, and the backend returned current indexed L2 blocks.

**Fresh signer-split validation + funding correction — fixed for quickstart defaults** (#48)

A clean `down -v` Sepolia boot on 2026-05-08 validated the first-time boot sequence with split coordinator signers. `account-setup` derived separate L1 blob and finalization submitter accounts from the user's deployer key, kept the pre-baked L2 message anchorer, and `config-render` injected three distinct coordinator pubkeys. Contract deployment completed, L2 Blockscout indexed blocks, coordinator opened `9545`/`9546`, the dev prover produced responses, and coordinator submitted L1 blob plus aggregation transactions.

The signer split itself was not the remaining blocker. The live run showed the previous funding default was too low: submitters were topped up to 0.03 ETH, but the first blob-submission path needed slightly more than 0.033 ETH of upfront max-fee headroom. Manually adding 0.2 ETH to each derived L1 submitter unblocked blob and aggregation submissions. Defaults now top up each derived L1 submitter with 0.25 ETH when its balance is below 0.10 ETH. Live logs still show transient nonce/underpriced retries during catch-up, but blob and aggregation submissions continue and finalization advances; that is a retry-noise caveat, not the earlier single-key `address already reserved` failure.

Follow-up cleanup in the same pass:

- `deploy-contracts.sh` now checks RPC reachability through Node `fetch` with the URL held in the process environment, instead of passing the provider URL as a `curl` argv value.
- `log4j2-dev.xml` drops the coordinator `App configs` startup dump so rendered provider endpoints are not printed in normal quickstart logs.
