# Reference Benchmarks

This directory holds the canonical reference benchmark for the linea-monorepo
prover. It pins the host class, request set, and runtime configuration so that
the GPU compression path can be reproduced and audited.

## What this branch ships

- **Compression (data-availability-v2)**: GPU-accelerated through `gpu/plonk2`.
  The prover automatically uses the GPU whenever a CUDA device is reachable
  (`gpu.HasDevice()`); no environment variable is needed. On the reference host
  this brings the per-proof wall-clock time below 2 min 30 s.
- **Aggregation (PI / BW6 / BN254)**: the GPU pipes are wired in (gpu/plonk2
  for the three Plonk phases, gpu/vortex for the public-input wizard
  commitments, gpu/quotient for the wizard quotient evaluation), but they are
  **disabled by default** behind the master flag
  `LINEA_PROVER_GPU_AGGREGATION=1`. Performance of the aggregation GPU path is
  not yet at production target — leave the flag off in production today.
- **Controller**: when launched on a GPU host, `cmd/controller` only accepts
  compression jobs. Execution / aggregation / invalidity files are ignored
  even if the corresponding `Enable*` toggles are on, since the GPU host
  would otherwise fall back to a slow CPU prover for those job types.

## Reference host

| Property | Value |
| --- | --- |
| Host class | AWS `g7e.8xlarge` |
| GPU | NVIDIA RTX PRO 6000 Blackwell Server Edition, 97887 MiB VRAM |
| NVIDIA driver | 590.48.01 |
| CPU | 32 vCPU, Intel Xeon Platinum 8559C |
| Memory | 249 GiB |
| Kernel | Linux 6.17.0-1013-aws, Ubuntu 24.04 |
| Go | `go1.26.0 linux/amd64` |
| Build command | `make GO_BUILD_TAGS=debug,cuda bin/prover` |
| Config | `reference-benchmarks/config-mainnet-limitless-7.1.0-provertestdata2.toml` |
| Data | `/home/ubuntu/provertestdata2/prover-compression` |

## Compression — GPU reference (3-proof batch, 2026-05-08)

The first three sorted compression requests under
`/home/ubuntu/provertestdata2/prover-compression/requests/` were each proved
in a fresh process (cold-cache for assets is paid once before the batch — the
first proof of a fresh boot is slower until the OS page-cache holds the
~46 GiB of canonical SRS).

Command shape:

```sh
GOMEMLIMIT=180GiB GOGC=75 \
  /usr/bin/time -v -o <time-file> \
  bin/prover prove \
    --config reference-benchmarks/config-mainnet-limitless-7.1.0-provertestdata2.toml \
    --in <compression-request-json> \
    --out <compression-response-json>
```

| Run | Block range | Wall time | Setup load | Solver | GPU prover | Max RSS | CPU |
| --- | --- | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | `30388561-30389025` | 2:10.41 | 16.81s | 33.12s | 1:43.61 | 200.7 GiB | 285% |
| 2 | `30389026-30389504` | 2:10.21 | 16.86s | 33.11s | 1:43.31 | 200.7 GiB | 285% |
| 3 | `30389505-30390023` | 2:09.96 | 16.90s | 33.08s | 1:43.12 | 200.6 GiB | 286% |

**Average wall time: 2:10.19** (`GPU prover total` average 1:43.35).

Each run is a single `bin/prover prove` process. The GPU prover sub-phases
(from `gpu/plonk2/bls12377/prove.go` instrumentation) decompose roughly as:
solve 33 s ▸ trace-ready / init GPU instance 19 s ▸ MSM commit L,R,O 4 s ▸
build Z 5 s ▸ iFFT+commit Z 3 s ▸ quotient GPU 25 s ▸ MSM commit h1,h2,h3 4 s
▸ eval+linearize+open Z 7 s ▸ MSM commit linPol 1.5 s ▸ batch opening 4 s.

Raw artifacts:

- `results/2026-05-08-g7e-8xlarge-gpu-compression-final/compression/` —
  proof responses
- `results/2026-05-08-g7e-8xlarge-gpu-compression-final/logs/` — raw prover
  logs and per-run `/usr/bin/time -v` output
- `results/2026-05-08-g7e-8xlarge-gpu-compression-final/env.txt` — host /
  build environment captured at run time

## Reproducing the compression run

1. Build the cuda binary:
   `make GO_BUILD_TAGS=debug,cuda bin/prover`
2. Make sure the canonical SRS and the 7.1.0 setup directory are populated
   under `prover-assets/7.1.0/data-availability-v2/`. The setup load is the
   first ~17 s of each proof.
3. Run any compression request from `provertestdata2`:
   ```sh
   GOMEMLIMIT=180GiB GOGC=75 bin/prover prove \
     --config reference-benchmarks/config-mainnet-limitless-7.1.0-provertestdata2.toml \
     --in <request.json> \
     --out <response.json>
   ```
4. The GPU is detected automatically. Expect the first compression after a
   fresh boot to pay the SRS page-cache fault (~2 min extra); subsequent
   runs should track the table above.

## Aggregation — gated GPU path

Set `LINEA_PROVER_GPU_AGGREGATION=1` to opt the aggregation pipeline into
GPU dispatch. With the flag set:

- the PI / BW6 / BN254 PlonK phases use `gpu/plonk2` (per-curve packages);
- the public-input wizard's Vortex MiMC and ring-SIS commitments use
  `gpu/vortex` (the keccak-vendored `gpu_mimc_cuda.go` path);
- the wizard quotient evaluation in `protocol/compiler/globalcs/quotient.go`
  uses `gpu/quotient`.

The flag is off by default. We are not benchmarking the aggregation GPU path
in this branch — those numbers belong to a follow-up PR once the path
matches CPU on production hosts.

## Diagnostic env vars (advanced)

These are tuning / debugging knobs and should not be set in normal operation:

| Variable | Purpose |
| --- | --- |
| `LINEA_PROVER_GPU_DEVICE_ID` | Pin the prover process to a specific GPU index (default 0). |
| `GNARK_GPU_PLONK2_MSM_WINDOW_BITS` | Override the auto-selected Pippenger window size for the BLS12-377 MSMs (default 20 for n > 2^26). |
| `GNARK_GPU_PLONK2_LOG_MSM_PHASES` | Log per-phase MSM timings from each `MultiExp` call. |
| `LINEA_PROVER_GPU_VORTEX_RECOMMIT` | Force the wizard Vortex commit prover to recompute the host-side encoded matrix at open time instead of keeping a device snapshot. |
| `LINEA_PROVER_GPU_VORTEX_SNAPSHOT_BUDGET_GIB` | VRAM budget for the cross-round Vortex snapshot (default 48 GiB). |

The keccak-vendored PI Vortex tuning knobs
(`LINEA_PROVER_GPU_PI_SIS_MIN_ROWS`, `LINEA_PROVER_GPU_PI_SIS_SPLIT_MIN_ROWS`,
etc.) are documented in
`circuits/pi-interconnection/keccak/prover/crypto/vortex/gpu_mimc_cuda.go`.
The master flag `LINEA_PROVER_GPU_AGGREGATION=1` automatically opts in to
PI MiMC, ring-SIS, and quotient-reevaluation when the operator does not
explicitly override the per-knob env vars.
