# Get Started

### Requirements:

- Node.js >= 24.14.1 (see `.nvmrc`)
- Docker v24 or higher
  - Docker should ideally have ~16 GB of Memory and 4+ CPUs to run the entire stack.
- Docker Compose version v2.19+
- Make v3.81+
- Pnpm >= 10.32.1 (https://pnpm.io/installation)

### Run stack locally

#### Install Node dependencies

```
make pnpm-install
```

#### Start the stack and run end-to-end tests

```
make start-env-with-tracing-v2
pnpm -F e2e run test:local
```

To stop the stack:

```
make clean-environment
```

While running the end-to-end tests, you should observe files being generated in `tmp/local/` directory.

```
├── local
│  ├── prover
│  │  ├── request
│  │  │  └── 4-4-etv0.2.0-stv1.2.0-getZkProof.json
│  │  ├── requests-done
│  │  │  ├── 1-1-etv0.2.0-stv1.2.0-getZkProof.json.success
│  │  │  ├── 2-2-etv0.2.0-stv1.2.0-getZkProof.json.success
│  │  │  └── 3-3-etv0.2.0-stv1.2.0-getZkProof.json.success
│  │  └── response
│  │      ├── 1-1-etv0.2.0-stv1.2.0-getZkProof.json.0.2.0.json
│  │      ├── 2-2-etv0.2.0-stv1.2.0-getZkProof.json.0.2.0.json
│  │      └── 3-3-etv0.2.0-stv1.2.0-getZkProof.json.0.2.0.json
│  └── traces
│      ├── conflated
│      │  ├── 1-1.conflated.v0.2.0.json.gz
│      │  ├── 2-2.conflated.v0.2.0.json.gz
│      │  ├── 3-3.conflated.v0.2.0.json.gz
│      │  └── 4-4.conflated.v0.2.0.json.gz
│      ├── raw
│      │  ├── 1-0x2e1a3f506c0d5f11310301a86f608d840d3db0e28c545eaf9e9c9812e2b795e0.v0.2.0.json.gz
│      │  ├── 2-0x3e5b3bd8e21a94488bf93776480271d3fef8033152effd4e19fe6519dea53379.v0.2.0.json.gz
│      │  ├── 3-0xa5046c13502a619a7a3f091b397234dc020f6cbda1942d247d1003d4c73899b6.v0.2.0.json.gz
│      │  ├── 4-0xe9203ede2114bf9c291692c4bd2dcc7207973c267ed411d65568d1138b3ecfcc.v0.2.0.json.gz
│      │  ├── 5-0x2c8ec07d4222bed8285be3de83f0fccc989134c49826baed5340bf7aa8e3ce8f.v0.2.0.json.gz
│      │  └── 6-0x3c7b7ee369d5fe02a6865415a2d0ef4ec385812351723e35a3b54d972f9f4ceb.v0.2.0.json.gz
│      └── raw-non-canonical
```

#### Troubleshooting

Docker: Sometimes restarting the stack several times may lead to network/state issues.
The following commands may help.
**Note:** Please be aware this will permanently remove all docker images, containers and **docker volumes**
and any data saved it them.

```
make clean-environment
docker system prune --volumes
```

## Tuning in conflation

For local testing and development conflation deadline is set to 6s `conflation-deadline=PT6S` in
`config/coordinator/coordinator-config-v2.toml` (Docker-based `make start-env*` stacks often mount a different file from `docker/config/`).
Hence, only a two-block conflation.
If you want bigger conflations, increase the deadline accordingly.

### Target checkpoints, L1 finalization, and API resume

Under `[conflation.proof-aggregation]` in the coordinator TOML configuration,
you can define **target checkpoints** and optional **gates** that pause L2 block import / conflation until release
conditions are met.

**`target-end-blocks`** — Block numbers that mark checkpoints. The coordinator uses the same target set across the
conflation pipeline so that, for each configured **`N`**, work is **cut off with end block number `N`**: an execution
**batch** ends at **`N`**, the enclosing **blob** is closed so its range ends at **`N`**, and **proof aggregation** is
triggered so that aggregation’s block range also ends at **`N`**. When the wait flags below are enabled, **import /
conflation then pauses** at that boundary: the pause is recognized when block **`N + 1`** is handed to conflation—that
is, after the batch/blob/aggregation slice through **`N`** has been completed and the coordinator is about to progress
past **`N`**.

**`timestamp-based-hard-forks`** — Ordered list of L2 block timestamp thresholds (ISO-8601 strings or Unix epoch
seconds, as in other coordinator TOML `Instant` fields). The list must be **strictly ascending** with **no
duplicates**. Conflation uses the same crossing rule as `TimestampHardForkConflationCalculator`: the **first** L2 block
whose timestamp is **at or above** a threshold **T**, while the previous block’s timestamp was **strictly below** **T**,
is the block **`N + 1`** for that fork. The coordinator closes conflation so that the execution **batch**,
enclosing **blob**, and **aggregation** all have **end block number `N`** (the last block still strictly before the
fork in time). The crossing block **`N + 1`** starts the next batch after the boundary. For the optional checkpoint
**pause** (when the wait flags below are enabled), the same **`N` / `N + 1`** split applies as for `target-end-blocks`:
the pause is tied to handing **`N + 1`** into conflation once the work through **`N`** has been completed.

**`wait-target-block-l1-finalization`** — When `true`, after a checkpoint the coordinator waits until the
**L1-finalized** L2 block number reported by the rollup is at least the checkpoint height before it may resume
importing.

**`wait-api-resume-after-target-block`** — When `true`, after a checkpoint the coordinator also waits for an operator
**JSON-RPC** call before resuming. If both wait flags are `true`, **both** the L1-finalization condition and the API
signal must be satisfied before the pause clears. If only one flag is `true`, only that gate applies.

The pause machinery is active when **either** wait flag is `true`. If both are `false`, `target-end-blocks` and
`timestamp-based-hard-forks` do not trigger this import pause (they can still affect proof aggregation layout
elsewhere).

**Resume API** — JSON-RPC method name: `conflation_signalTargetCheckpointResume`. The body is standard JSON-RPC 2.0;
`params` may be an empty array. The `result` is a boolean: `true` if the resume signal was accepted (API gate enabled
and a checkpoint pause was actually waiting on the API), `false` otherwise.

Coordinator API settings live under `[api]` (`json-rpc-port`, optional `json-rpc-path`; default path is `/`). In the
default Docker tracing stack, `json-rpc-port` is **9546** and is **not** published to the host; call it from inside the
`coordinator` container or add a port mapping if you need host access.

Example:

```bash
curl -sS -X POST "http://127.0.0.1:9546/" \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":1,"method":"conflation_signalTargetCheckpointResume","params":[]}'
```

If you set a non-default `json-rpc-path` (for example `/jsonrpc`), append that path to the URL instead of `/`.

## Next steps

Consider reviewing the [Linea architecture](architecture-description.md) description.

For detailed instructions on local development and building services locally, see the [Local Development Guide](local-development-guide.md).
