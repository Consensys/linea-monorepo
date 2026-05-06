# Reference Execution Benchmark

Date: 2026-05-06 UTC
Branch: `perf/limitless-onthefly`
Merge update: `origin/main` at `4ba5708501d3d712bc65d056635371cf6491f426` is merged into `HEAD` at `ebe784467abb159e9dfaad87dd4e78b57fb2e089`.

## Configuration

The benchmark targeted `config/config-mainnet-limitless.toml` in limitless execution mode:

- `version = "7.1.0"`
- `prover_mode = "limitless"`
- `conflated_traces_dir = "/home/ubuntu/testdata"`
- `requests_root_dir = "/home/ubuntu/testdata/execution"`
- `serialization = false`, so the on-the-fly path is used.
- `ignore_compatibility_check = false`

The request trace paths are absolute-style `/shared/...` paths. To make those resolve under `/home/ubuntu/testdata`, this symlink is present:

```text
/home/ubuntu/testdata/shared/v3/conflation-backtesting -> /home/ubuntu/testdata/conflated
```

Note: literal `GOMEMLIMIT=700GB` is rejected by the Go runtime as malformed, so the benchmark commands used `GOMEMLIMIT=700000000000` with `GOGC=500`.

## Setup

The selected requests use trace engine `linea-9cb6f11`, so `zkevm/arithmetization/zkevm.bin` was rebuilt for trace commit `9cb6f117bb1a693d42e5ca9bcb9443873aef54db`.

The pre-existing `7.1.0` `execution-limitless` setup was stale for the rebuilt circuit. It had circuit checksum `0x095c49cb3968381115ee1d6414f5b50f5db1a6055d3b5ec83df1ed5926b5d6d2`; the current circuit checksum is `0xc9a191e1ef23ea672f025f96582a4dbcd30446ab78b5533be52236aa8670de6a`.

Regenerated `execution-limitless` setup:

- Command: `./bin/prover setup --config config/config-mainnet-limitless.toml --circuits execution-limitless`
- Wall time: `7:26.05`
- Max RSS: `640,137,972 KB`
- Exit status: `0`
- Verifying key: `0x6284bdc7f7b67e88eaf349ef90d0cd6afe77896823826062d5f006cf4ed5552e`
- Circuit checksum: `0xc9a191e1ef23ea672f025f96582a4dbcd30446ab78b5533be52236aa8670de6a`
- Config checksum: `0xcf09b310e6c0147cd58d0956c2e2d9eafd35e793fa07a5bd899a5f4c36ecd6ac`
- Constraints: `86,241,368`

Retained setup logs:

```text
benchmark_results/reference_execution_setup_skip/
```

## Build And Validation

- Built the prover binary with `make bin/prover` before running the benchmark.
- Ran `gofmt` over touched Go files; `gofmt -l` returned no output.
- Passed: `go test ./zkevm/prover/ecdsa -tags nocorset,fuzzlight -run 'TestEcDataAssignData'`
- Passed: `go test ./backend/execution/limitless ./crypto/vortex ./zkevm/prover/ecdsa -tags nocorset,fuzzlight -timeout 30m`
- `golangci-lint run` could not be executed because `golangci-lint` is not installed in this environment.
- Passed: `git diff --check`

## Selected Requests

| Request | Blocks | Request bytes | Trace bytes | Trace engine |
|---|---:|---:|---:|---|
| `30006397-30006404-getZkProof` | 8 | 805,556 | 2,269,716 | `linea-9cb6f11` |
| `30015036-30015073-getZkProof` | 38 | 11,457,001 | 46,232,161 | `linea-9cb6f11` |
| `30016678-30016741-getZkProof` | 64 | 13,300,148 | 39,202,267 | `linea-9cb6f11` |

## Results

All three selected requests completed successfully and wrote execution proof responses under `benchmark_results/reference_execution_regenerated/`.

| Request | Blocks | Exit | Wall time | Max RSS | Proof file | Public input |
|---|---:|---:|---:|---:|---:|---|
| `30006397-30006404-getZkProof` | 8 | 0 | `12:08.07` | `671,388,696 KB` | 14K | `0x070ef12a12d95c6871d3a978d69f6470e3ddff01fb114c22da42da24bb092f51` |
| `30015036-30015073-getZkProof` | 38 | 0 | `19:39.04` | `721,440,408 KB` | 114K | `0x02a61edc178fb29bf2f07312ef973073272f2c9c92d77534a25ef4531ffbcb0c` |
| `30016678-30016741-getZkProof` | 64 | 0 | `17:02.97` | `682,338,888 KB` | 154K | `0x1188122ff70dc2d797fe64d2a85622c1c2281d9045c0ba07ea019a38efdbd7b9` |

All three proof responses report verifying key `0x6284bdc7f7b67e88eaf349ef90d0cd6afe77896823826062d5f006cf4ed5552e`, and each serialized proof string is 2,450 characters.

Outer proof milestones:

| Request | Conglomeration done | Witness | Outer proof | Sanity check |
|---|---|---|---|---|
| `30006397-30006404-getZkProof` | 19:31:48 | 19:31:49 | 19:31:53 | 19:35:53 |
| `30015036-30015073-getZkProof` | 20:07:17 | 20:07:18 | 20:07:22 | 20:11:24 |
| `30016678-30016741-getZkProof` | 20:24:59 | 20:25:00 | 20:25:04 | 20:29:05 |

## Gnark PR 1761 Attempt

After the baseline run, the benchmark was repeated with
[`Consensys/gnark#1761`](https://github.com/Consensys/gnark/pull/1761),
`perf: generalize sparse solver optimizations`.

- PR head: `3460cedcac438acd8bb16b255e0e335f9133895f`
- Go module version: `github.com/consensys/gnark v0.14.1-0.20260505192735-3460cedcac43`
- `github.com/consensys/gnark-crypto` remained `v0.20.2-0.20260402204920-39238e584b99`
- GitHub reported merge commit `9ffdc029ee4143141c3f8b9f5ec6f9969602527a`, but that revision was not fetchable as a Go module revision, so the benchmark used the PR head.

Built the prover with:

```bash
go get github.com/consensys/gnark@3460cedcac438acd8bb16b255e0e335f9133895f
make bin/prover
```

### Setup compatibility

A first attempt reused the baseline `execution-limitless` setup and failed at the outer Plonk proof stage:

- Request: `30006397-30006404-getZkProof`
- Wall time: `9:16.81`
- Max RSS: `678,951,164 KB`
- Exit status: `2`
- Failure: `constraint #29427927 is not satisfied` in `execution.(*CircuitExecution).CheckLimitlessConglomerationCompletion`

The diagnostic stack referenced the previous gnark setup code path:

```text
github.com/consensys/gnark@v0.14.1-0.20260219004710-bbfb2f70a565
```

That made the old setup incompatible with the PR binary, so the `execution-limitless` setup was regenerated.
The regeneration used a temporary local setup-only bypass for the final limitless asset-store copy; that patch
was reverted before building the prover used for proving.

Regenerated PR-1761 setup:

- Command:
  `LINEA_SKIP_LIMITLESS_ASSET_STORE=1 GOGC=500 GOMEMLIMIT=700000000000 ./bin/prover setup --config config/config-mainnet-limitless.toml --circuits execution-limitless --force`
- Wall time: `6:22.18`
- Max RSS: `637,486,632 KB`
- Exit status: `0`
- Verifying key: `0x32679a695a51f57c1ec61e2a220ae55902f801f7ae4e47490e59c9946768a4d0`
- Circuit checksum: `0x4e7a1450e1c327fa79b9eb91b9dc5bbdbccfbe26dd79e9b152c564db74d25c57`
- Config checksum: `0xcf09b310e6c0147cd58d0956c2e2d9eafd35e793fa07a5bd899a5f4c36ecd6ac`
- Constraints: `84,819,973`

### Prove result

After regenerating setup with PR 1761, the smallest reference request reached the outer proof stage but panicked
inside gnark before writing a proof response:

| Request | Blocks | Exit | Wall time | Max RSS | Proof file | Failure |
|---|---:|---:|---:|---:|---:|---|
| `30006397-30006404-getZkProof` | 8 | 2 | `8:30.83` | `674,750,608 KB` | none | `panic: level 142: unknown proving level type *constraint.GkrSkipLevel` |

Outer proof milestones before the panic:

| Request | Conglomeration done | Witness | Outer proof | Sanity check |
|---|---|---|---|---|
| `30006397-30006404-getZkProof` | 21:05:04 | 21:05:05 | 21:05:10 | not reached |

Failure signature:

```text
panic: level 142: unknown proving level type *constraint.GkrSkipLevel
github.com/consensys/gnark/internal/gkr/bls12-377.Prove
    .../github.com/consensys/gnark@v0.14.1-0.20260505192735-3460cedcac43/internal/gkr/bls12-377/gkr.go:154
```

At that location, the PR code handles `constraint.GkrSkipLevel` values, while the loaded proving schedule
contained `*constraint.GkrSkipLevel`. This is consistent with the PR's own note that full generated
constraint-package tests still fail around a GKR blueprint serialization round-trip mismatch. The Linea
execution benchmark depends on serialized setup assets and hits that failure in the end-to-end proving path.

No comparable wall-time improvement table is available for PR 1761 because the smallest selected request does
not complete. The 38-block and 64-block requests were not rerun after this failure because they share the same
outer BLS12-377 Plonk/GKR proving path and would not produce comparable proof timings until the gnark failure is
fixed.

## Issues Found And Fixed

### Stale outer setup

Before regenerating setup, the smallest request reached the outer proof stage and failed in `execution.checkL2MSgHashes`:

```text
constraint #29375959 is not satisfied: [assertIsEqual] 0 == 804848009
execution.checkL2MSgHashes
execution.checkPublicInputs
execution.(*CircuitExecution).Define
```

The same failure reproduced with downloaded setup versions `7.1.0`, `7.0.7`, and `7.0.1`. `804848009` is `0x2ff90189`, matching the tail of the request's `zkParentStateRootHash`. A direct consistency check between the current inner proof and the current execution circuit passed, which isolated the failure to stale outer setup assets rather than trace parsing or public input propagation.

### ECDSA antichamber capacity

After setup regeneration, the 38-block request initially failed during bootstrap:

```text
[LIMIT OVERFLOW] limit=8192 requested=132402 err=column ECDSA_ANTICHAMBER_ECRECOVER_ECRECOVER_ID slice size 132402 is larger than column size 8192
```

The retry at scaling 32 failed with the same antichamber column at `requested=4195634`. The root cause was that `assignFromEcDataSource` skipped non-ecrecover source rows while also pushing zero rows into the packed ECDSA antichamber vectors. Removing those zero pushes keeps the antichamber packed by actual ecrecover rows, matching the existing comment and capacity model.

Added regression test: `TestEcDataAssignData_SkipsNonEcrecoverRows`.

## Retained Artifacts

```text
benchmark_results/reference_execution_regenerated/
benchmark_results/reference_execution_setup_skip/
benchmark_results/reference_execution_gnark1761_setup/
benchmark_results/reference_execution_gnark1761_old_setup_failed/
benchmark_results/reference_execution_gnark1761/
```

Earlier failed-attempt logs are also retained for comparison:

```text
benchmark_results/reference_execution_attempts/
benchmark_results/reference_execution/
```
