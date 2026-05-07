# SCAFFOLD-PLAN — `docs/getting-started/linea-stack/`

_v0 of "Streamlined Linea Stack deployment". Technical Feature Owner: Moris Iarossi._
_Last updated: 2026-05-06._

This is the planned file inventory across both stages. Review before stage 1 starts. Each entry is marked **ready-to-run** (will land working in v0) or **TODO** (will land as a stub or a documented gap that needs follow-up before public release).

## Stage boundaries

- **Stage 1** (after this plan is approved): `docker-compose.yml`, `versions.env`, `.env.example`, `README.md` skeleton.
- **Stage 2** (after stage 1 review): `scripts/`, `config/l1/`, `config/l2/`, `config/web3signer/`, `config/observability/`, `config/explorer/`.

## Status legend

- 🟢 **ready-to-run** — file lands functional in v0; lifted from internal or written against a clear pattern.
- 🟡 **ready-to-run, validation pending** — written from scratch or non-trivially adapted, needs first-boot test before we promote v0.
- 🔴 **TODO** — known gap at v0 (binary file, hardhat task wiring not finalised, etc.). Documented in the file itself.

## Decisions (resolved with TFO 2026-05-06)

1. **Postgres = isolated instances per consumer.** Four services: `coordinator-pg`, `postman-pg`, `blockscout-l1-pg`, `blockscout-l2-pg`. (Postman gets its own per "isolation > shared".) Each gets its own data volume.
2. **Blockscout configs under `config/explorer/`** (not `config/observability/`). Spec scoped observability to Prometheus/Grafana/Loki/Promtail.
3. **`l2-node-besu` follower included** alongside sequencer. Sequencer RPC stays internal; users hit `l2-node-besu:8545`. Consistent with `linea-mainnet/` and `linea-sepolia/` siblings.
4. **`transaction-exclusion-api` dropped.** Confirmed via `besu-plugins/linea-sequencer/docs/plugins.md:152` — `--plugin-linea-rejected-tx-endpoint` defaults to `null` and is optional. Flag removed from sequencer + l2-node-besu commands. No stub.
5. **L1-EL = Besu** (consensys/linea-besu-package), matching the internal validation path. (Tech-design A4 will be updated separately by TFO.) Cleaner Consensys-stack story: Besu on L1, Besu sequencer on L2, Teku consensus on L1.
6. **Image tag pins copied from `compose-spec-l2-services.yml` defaults** (the `${VAR:-tag}` `tag` portion). Source of truth.
7. **`addresses.json` shared-volume handoff.** `deploy-contracts.sh` writes it to a `linea-shared-config` volume; coordinator/sequencer/postman/prover all mount the same path read-only. No cross-service env var plumbing.
8. **Pre-baked dev keys copied verbatim** from `docker/web3signer/key-files/` and `docker/config/maru/`, banner header prepended. The TLS `.p12` keystore is binary and cannot be Write'd by my tools — explicit 🔴 TODO with documented `cp` step in README.
9. **Hardhat wrapper bind-mounts `../../../contracts`** into the `deploy-contracts` container. Means v0 lives inside the linea-monorepo and isn't a fully portable tarball — that's the v1 packaging story. README will be explicit.

## File tree (with status)

```
docs/getting-started/linea-stack/
├── SCAFFOLD-PLAN.md                              [you are here]
├── README.md                                     stage 1 — 🟢 ready-to-run (skeleton; full prose lands now)
├── docker-compose.yml                            stage 1 — 🟡 validation pending (15 services, 2 profiles)
├── versions.env                                  stage 1 — 🟢 ready-to-run (pins lifted from internal)
├── .env.example                                  stage 1 — 🟢 ready-to-run
│
├── scripts/
│   ├── deploy-contracts.sh                       stage 2 — 🟡 hardhat wrapper; calls existing tasks in ../../../contracts via pnpm exec hardhat
│   └── seed-funds.sh                             stage 2 — 🟢 ready-to-run (cast send loop; idempotent via balance check)
│
├── config/
│   ├── l1/
│   │   ├── genesis-generator/
│   │   │   ├── network-config.yml                stage 2 — 🟢 lifted from docker/config/l1-node/cl/network-config.yml
│   │   │   ├── mnemonics.yaml                    stage 2 — 🟢 lifted from docker/config/l1-node/cl/mnemonics.yaml (DEV-ONLY banner)
│   │   │   └── generate-genesis.sh               stage 2 — 🟢 lifted from docker/config/l1-node/generate-genesis.sh
│   │   ├── el/  (Besu)
│   │   │   ├── config.toml                       stage 2 — 🟢 lifted from docker/config/l1-node/el/config.toml
│   │   │   ├── besu.key                          stage 2 — 🟢 lifted (DEV-ONLY banner)
│   │   │   ├── log4j.xml                         stage 2 — 🟢 lifted
│   │   │   └── jwtsecret.txt                     stage 2 — 🟢 lifted (DEV-ONLY banner)
│   │   └── cl/  (Teku)
│   │       ├── teku-config.yaml                  stage 2 — 🟢 lifted from docker/config/l1-node/cl/config.yaml
│   │       ├── teku.key                          stage 2 — 🟢 lifted (DEV-ONLY banner)
│   │       ├── teku-keys/                        stage 2 — 🟢 directory lifted from docker/config/l1-node/cl/teku-keys
│   │       ├── teku-secrets/                     stage 2 — 🟢 directory lifted from docker/config/l1-node/cl/teku-secrets
│   │       └── log4j.xml                         stage 2 — 🟢 lifted
│   │
│   ├── l2/
│   │   ├── genesis-init/
│   │   │   ├── init.sh                           stage 2 — 🟢 lifted from docker/config/l2-genesis-initialization/init.sh
│   │   │   ├── genesis-besu.json.template        stage 2 — 🟢 lifted (template that init.sh fills in)
│   │   │   └── genesis-maru.json.template        stage 2 — 🟢 lifted
│   │   ├── sequencer/
│   │   │   ├── sequencer.config.toml             stage 2 — 🟡 lifted from docker/config/linea-besu-sequencer/sequencer.config.toml; tx-exclusion-api plugin flag dropped
│   │   │   ├── key                               stage 2 — 🟢 lifted (DEV-ONLY banner)
│   │   │   ├── deny-list.txt                     stage 2 — 🟢 empty placeholder file (sequencer bind-mounts it)
│   │   │   └── log4j.xml                         stage 2 — 🟢 lifted
│   │   ├── l2-node-besu/
│   │   │   ├── l2-node-besu.config.toml          stage 2 — 🟢 lifted from docker/config/l2-node-besu/l2-node-besu-config.toml
│   │   │   └── log4j.xml                         stage 2 — 🟢 lifted
│   │   ├── maru/
│   │   │   ├── config.toml                       stage 2 — 🟢 lifted from docker/config/maru/sequencer.config.toml
│   │   │   ├── private-key                       stage 2 — 🟢 lifted from docker/config/maru/0x083e...key (DEV-ONLY banner)
│   │   │   └── log4j.xml                         stage 2 — 🟢 lifted
│   │   ├── coordinator/
│   │   │   ├── coordinator-config.toml           stage 2 — 🟡 distilled from docker/config/l2-genesis-initialization/coordinator-config-v2-hardforks.toml + config/coordinator/coordinator-config-v2.toml; reads addresses.json from shared volume
│   │   │   ├── vertx-options.json                stage 2 — 🟢 lifted from config/coordinator/vertx-options.json
│   │   │   ├── log4j2-dev.xml                    stage 2 — 🟢 lifted
│   │   │   └── tls-files/                        stage 2 — 🔴 TODO (TLS keystore is binary; document copy step from config/coordinator/tls-files)
│   │   ├── prover/
│   │   │   └── prover-config-partial.toml        stage 2 — 🟢 lifted from docker/config/prover/v3/prover-config-partial.toml
│   │   ├── postman/
│   │   │   └── env                               stage 2 — 🟡 lifted from docker/config/postman/env, addresses sourced from /shared/addresses.json at runtime
│   │   └── shomei/
│   │       └── log4j.xml                         stage 2 — 🟢 lifted
│   │
│   ├── web3signer/
│   │   ├── key-files/
│   │   │   ├── anchoring-signer.yaml             stage 2 — 🟢 lifted from docker/web3signer/key-files (DEV-ONLY banner enlarged)
│   │   │   ├── data-submission-signer.yaml       stage 2 — 🟢 lifted (DEV-ONLY)
│   │   │   ├── finalization-signer.yaml          stage 2 — 🟢 lifted (DEV-ONLY)
│   │   │   └── liveness-signer.yaml              stage 2 — 🟢 lifted (DEV-ONLY)
│   │   ├── tls-files/
│   │   │   ├── known-clients.txt                 stage 2 — 🟢 lifted
│   │   │   ├── web3signer-keystore-password.txt  stage 2 — 🟢 lifted (DEV-ONLY)
│   │   │   └── web3signer-keystore.p12           stage 2 — 🔴 TODO (binary file; cannot write via tools, README documents `cp` from internal)
│   │   └── conf/
│   │       └── config.yaml                       stage 2 — 🟢 lifted from docker/web3signer/conf/config.yaml
│   │
│   ├── observability/
│   │   ├── prometheus.yml                        stage 2 — 🟡 newly written; scrapes sequencer, l2-node-besu, coordinator, postman, prover (when up), web3signer, both blockscouts
│   │   ├── loki-config.yaml                      stage 2 — 🟢 lifted from docker/config/observability/loki-config.yaml
│   │   ├── promtail-config.yaml                  stage 2 — 🟢 lifted from docker/config/observability/promtail-config.yaml
│   │   ├── grafana.ini                           stage 2 — 🟢 lifted from docker/config/observability/grafana.ini
│   │   └── grafana-provisioning/
│   │       ├── datasources/datasources.yaml      stage 2 — 🟡 newly written; Prometheus + Loki datasources
│   │       ├── dashboards/dashboards.yaml        stage 2 — 🟢 standard provisioning manifest
│   │       └── dashboards/linea-l2-health.json   stage 2 — 🟡 hand-rolled, 5 panels (block height, tx count, L1 submission lag, coordinator queue depth, prover throughput)
│   │
│   └── explorer/
│       ├── l1-blockscout.env                     stage 2 — 🟢 lifted from config/blockscout/l1-blockscout.env
│       └── l2-blockscout.env                     stage 2 — 🟢 lifted from config/blockscout/l2-blockscout.env
│
└── config/postgres/
    ├── coordinator-init.sql                      stage 2 — 🟢 schema bootstrap for coordinator-pg (lifted patterns from docker/postgres/init)
    ├── postman-init.sql                          stage 2 — 🟢 schema bootstrap for postman-pg
    ├── blockscout-l1-init.sql                    stage 2 — 🟢 createdb stub
    └── blockscout-l2-init.sql                    stage 2 — 🟢 createdb stub
```

## Open questions

All resolved. Stage 1 is unblocked.

## Stage 2 — required-before-merge TODOs

These are the items that must be resolved during stage 2 (not deferrable to a v0.1 follow-up):

1. **L1 Blockscout first-boot validation.** The internal stack uses `consensys/linea-besu-package` for L1-EL with the `--profile` mechanism, but Blockscout was historically validated against the Geth-based L1 in the orphaned `compose-spec-extra-observability.yml`. Need to verify:
   - The `genesis.json` mounted at `/app/genesis.json` is the EL-genesis form Blockscout expects (not the Beacon SSZ).
   - `ETHEREUM_JSONRPC_TRACE_URL` works against Linea Besu's `debug_traceTransaction` (Geth-style trace API). If Linea Besu only exposes Besu-style traces, Blockscout L1 needs a different `JSONRPC_VARIANT` or won't index internal txs.
   - WebSocket subscription works on Linea Besu's `:8546`.
   Action: bring up `stack-no-prover`, watch `l1-blockscout` logs for the first 5 min, confirm blocks index. If it fails, switch L1 EL Blockscout config to use a different Geth client OR keep Linea Besu but disable the trace-based features in `l1-blockscout.env`.

2. **IP → hostname replacement when lifting configs.** The internal stack hardcodes IPs in two places that I'll be lifting; both must be rewritten to use Docker hostnames (single `linea` network in v0):
   - `docker/config/l2-genesis-initialization/coordinator-config-v2-hardforks.toml` — the coordinator config has multiple `http://11.11.11.x` and `http://10.10.10.x` references for sequencer, l2-node-besu, shomei, l1-el-node, web3signer endpoints. Each must become its container hostname (`http://sequencer:8545`, `http://l2-node-besu:8545`, `http://shomei:8888`, `http://l1-el-node:8545`, `https://web3signer:9000`).
   - `docker/compose-spec-l2-services.yml:455-465` — shomei command line uses `--besu-rpc-http-host=11.11.11.119` and `--rpc-http-host=11.11.11.114`. Already swapped in our compose-file shomei command (`--besu-rpc-http-host=l2-node-besu`, `--rpc-http-host=0.0.0.0`); just calling out so the lifted shomei log4j and any sibling configs follow suit.
   - L2 sequencer's bootnode enode is referenced by `@11.11.11.101:30303` in internal l2-node-besu commands. Already swapped to `@sequencer:30303` in our compose; verify the enode public key still matches the lifted sequencer key file (it does — same key file lifted).

3. **transaction-exclusion-api plugin flag — confirmed OPTIONAL.** Per `besu-plugins/linea-sequencer/docs/plugins.md:152`:
   > `--plugin-linea-rejected-tx-endpoint` | default: `null` | "A valid URL e.g. `http://localhost:9363` to enable reporting"
   The flag is omitted from sequencer + l2-node-besu commands in our compose. No stub service required. Re-verified at stage 1 sign-off.

## Estimated tool-call budget

- Stage 1: ~6 calls (4 file writes + this plan edit + final summary). _Done._
- Stage 2: ~50 calls (many small lifts + config writes + 2 scripts + dashboard JSON).
