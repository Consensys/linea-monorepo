# Reference Benchmarks

Reference wall-clock benchmarks for the prover on the current CPU reference
host, using real request/response data from `/home/ubuntu/provertestdata2`.

## Environment

- Date: 2026-05-04
- Host class: AWS `r8a.24xlarge`
- CPU: 96 vCPU, AMD EPYC 9R45, 1 thread per core
- Memory: 741 GiB
- Kernel: Linux 6.17.0-1012-aws, Ubuntu 24.04
- Git commit: `331a64144807582f4010569c0c10c4b3bc3b14b0`
- Go: `go1.26.0 linux/amd64`
- Build command: `make bin/prover`
- Prover binary: `bin/prover`, built with default `GO_BUILD_TAGS=debug`
- GPU: `nvidia-smi` could not communicate with an NVIDIA driver on this host.
  These numbers are CPU/default-Plonk numbers, not GPU-Plonk numbers.

## Data Source

The active benchmark source is `/home/ubuntu/provertestdata2`.

| Path | Contents |
| --- | ---: |
| `prover-execution/responses` | 2915 execution responses |
| `prover-compression/requests` | 288 compression requests |
| `prover-compression/responses` | 288 compression responses |
| `prover-aggregation/requests` | 48 aggregation requests |

Execution proofs were not regenerated. Aggregation benchmarks consumed the
existing execution responses referenced by the aggregation requests.

## Configuration

The run used
`reference-benchmarks/config-mainnet-limitless-7.1.0-provertestdata2.toml`.
It is based on `config/config-mainnet-limitless.toml`, with:

- `version = "7.1.0"`
- `execution.requests_root_dir = "/home/ubuntu/provertestdata2/prover-execution"`
- `data_availability.requests_root_dir = "/home/ubuntu/provertestdata2/prover-compression"`
- `aggregation.requests_root_dir = "/home/ubuntu/provertestdata2/prover-aggregation"`
- `aggregation.num_proofs = [10, 20, 50, 100, 200, 400]`

Keeping the full `num_proofs` list matters: `aggregation-100` must be at
`setupPos=3` so the BN254 emulation setup matches the circuit verification key
layout.

## Circuit Metadata

| Circuit | Curve | Constraints |
| --- | --- | ---: |
| `data-availability-v2` | BLS12-377 | 126,587,080 |
| `public-input-interconnection` | BLS12-377 | 56,396,760 |
| `aggregation-100` | BW6-761 | 27,009,624 |
| `emulation` | BN254 | 14,407,969 |

The setup-load measurements come from the `loaded setup ... duration=...` log
line emitted after reading circuit, proving key, verifying key, and manifest
data from disk.

## Compression Proofs

Command shape:

```sh
/usr/bin/time -v -o <time-file> \
  bin/prover prove \
    --config reference-benchmarks/config-mainnet-limitless-7.1.0-provertestdata2.toml \
    --in <compression-request-json> \
    --out <compression-response-json>
```

Results for the first five sorted requests:

| Run | Block range | Wall time | Setup load | Solver | Prover | Max RSS | CPU |
| --- | --- | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | `30388561-30389025` | 4:40.85 | 17.286s | 31.9s | 258.7s | 203.0 GiB | 4701% |
| 2 | `30389026-30389504` | 4:38.50 | 17.335s | 31.9s | 256.2s | 207.3 GiB | 4740% |
| 3 | `30389505-30390023` | 4:40.37 | 17.610s | 31.7s | 257.7s | 214.5 GiB | 4715% |
| 4 | `30390024-30390503` | 4:41.73 | 17.301s | 31.9s | 259.5s | 204.9 GiB | 4718% |
| 5 | `30390504-30390918` | 4:40.09 | 17.500s | 30.7s | 257.8s | 204.5 GiB | 4707% |

Average wall time: 4:40.31.

## Aggregation Proofs

Command shape:

```sh
/usr/bin/time -v -o <time-file> \
  bin/prover prove \
    --config reference-benchmarks/config-mainnet-limitless-7.1.0-provertestdata2.toml \
    --in <aggregation-request-json> \
    --out <aggregation-response-json>
```

Results for the first five sorted requests:

| Run | Block range | Inputs | Wall time | Setup load | Prover phases | Max RSS | CPU |
| --- | --- | --- | ---: | --- | --- | ---: | ---: |
| 1 | `30388561-30391349` | 58 execution, 6 compression | 6:46.32 | PI 16.783s, BW6 14.027s, BN254 2.517s | PI 128.6s, BW6 136.3s, BN254 26.7s | 199.2 GiB | 6060% |
| 2 | `30391350-30394103` | 51 execution, 6 compression | 6:47.74 | PI 16.871s, BW6 15.488s, BN254 2.391s | PI 127.7s, BW6 137.5s, BN254 26.5s | 200.4 GiB | 6037% |
| 3 | `30394104-30396213` | 52 execution, 6 compression | 6:47.80 | PI 17.720s, BW6 15.018s, BN254 2.458s | PI 129.8s, BW6 136.3s, BN254 26.8s | 197.4 GiB | 6065% |
| 4 | `30396214-30398752` | 68 execution, 6 compression | 6:43.39 | PI 18.121s, BW6 15.767s, BN254 2.495s | PI 126.3s, BW6 135.2s, BN254 26.6s | 202.8 GiB | 6100% |
| 5 | `30398753-30401287` | 75 execution, 6 compression | 6:48.52 | PI 16.376s, BW6 12.917s, BN254 2.410s | PI 131.6s, BW6 136.8s, BN254 26.5s | 180.5 GiB | 6054% |

Average wall time: 6:46.75.

Solver phases were also stable:

| Stage | Constraints | Solver time range |
| --- | ---: | ---: |
| PI | 56,396,760 | 25.0s-26.1s |
| BW6 | 27,009,624 | 4.7s-4.8s |
| BN254 | 14,407,969 | 3.3s-3.6s |

## GPU Plonk Benchmark Addendum

The same first five sorted compression and aggregation requests were rerun on
an AWS `g7e.8xlarge` host using the GPU Plonk prover.

Environment:

- Date: 2026-05-04
- Host class: AWS `g7e.8xlarge`
- CPU: 32 vCPU, Intel Xeon Platinum 8559C
- Memory: 249 GiB
- GPU: NVIDIA RTX PRO 6000 Blackwell Server Edition, 97,887 MiB VRAM
- NVIDIA driver: 590.48.01
- Git commit before local GPU changes: `091e487f5e4a27149c81f59aa0e90373b613402d`
- Go: `go1.26.0 linux/amd64`
- Build command: `make bin/prover`
- GPU mode:
  `LINEA_PROVER_GPU_PLONK2=1 GNARK_GPU_PLONK2_LOG_MSM_PHASES=1`

All GPU runs used `gpu/plonk2` with strict fallback disabled. The logs show
`cpuFallback=false` for every Plonk proof and no CPU MSM fallback was used.

### GPU Compression Proofs

Baseline GPU run, before local GPU prover optimizations:

| Run | Block range | Wall time | Setup load | Solver | GPU prover | Peak VRAM | Max RSS | CPU |
| --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | `30388561-30389025` | 5:31.18 | 2:29.38 | 33.565s | 2:47.25 | 79.6 GiB | 220.6 GiB | 150% |
| 2 | `30389026-30389504` | 4:15.15 | 1:17.01 | 30.139s | 2:44.00 | 79.6 GiB | 211.1 GiB | 191% |
| 3 | `30389505-30390023` | 3:34.82 | 30.284s | 33.023s | 2:49.58 | 79.6 GiB | 232.4 GiB | 229% |
| 4 | `30390024-30390503` | 5:52.78 | 2:46.39 | 33.714s | 2:51.65 | 79.6 GiB | 215.0 GiB | 160% |
| 5 | `30390504-30390918` | 3:46.26 | 35.203s | 35.952s | 2:55.78 | 79.6 GiB | 232.6 GiB | 251% |

Average wall time: 4:36.04.
Average GPU prover phase: 2:49.65.
Average setup load: 1:31.65.

Optimized single compression reference, after local GPU prover changes:

| Block range | Wall time | Setup load | Solver | GPU prover | MSM policy | Max RSS | CPU |
| --- | ---: | ---: | ---: | ---: | --- | ---: | ---: |
| `30389505-30390023` | 2:35.68 | 23.315s | 54.341s | 2:02.51 | strict GPU, window 20, split/no CPU fallback | 204.5 GiB | 321% |

The optimized run used `bin/prover-cuda` with
`LINEA_PROVER_GPU_PLONK2=1` and a BLS12-377 MSM window of 20. Its log reports
`cpuFallback=false`; no CPU MSM fallback was used.

Final optimized 3-proof compression batch, rerun on 2026-05-05:

| Run | Block range | Wall time | Setup load | Solver | GPU prover | Max RSS | CPU |
| --- | --- | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | `30388561-30389025` | 2:59.45 | 51.409s | 48.52s | 1:57.50 | 208.5 GiB | 269% |
| 2 | `30389026-30389504` | 3:02.01 | 51.947s | 49.99s | 1:58.65 | 208.6 GiB | 293% |
| 3 | `30389505-30390023` | 2:38.80 | 31.978s | 47.89s | 1:56.61 | 208.4 GiB | 337% |

Average wall time: 2:53.42.

All three final compression proofs used strict GPU Plonk
(`cpuFallback=false`) and passed the Plonk sanity check.

### GPU Aggregation Proofs

| Run | Block range | Wall time | Setup load | GPU prover phases | Peak VRAM | Max RSS | CPU |
| --- | --- | ---: | --- | --- | ---: | ---: | ---: |
| 1 | `30388561-30391349` | 13:24.75 | PI 2:37.35, BW6 4:26.08, BN254 21.844s | PI 1:27.48, BW6 1:08.28, BN254 14.783s | 72.9 GiB | 189.8 GiB | 1291% |
| 2 | `30391350-30394103` | 9:02.14 | PI 27.192s, BW6 25.899s, BN254 3.770s | PI 1:30.35, BW6 1:09.62, BN254 14.020s | 72.9 GiB | 172.2 GiB | 1886% |
| 3 | `30394104-30396213` | 9:46.22 | PI 26.540s, BW6 1:17.56, BN254 3.748s | PI 1:29.78, BW6 1:03.40, BN254 13.918s | 72.9 GiB | 189.5 GiB | 1742% |
| 4 | `30396214-30398752` | 9:31.63 | PI 28.312s, BW6 1:00.17, BN254 3.743s | PI 1:29.00, BW6 1:02.65, BN254 13.684s | 72.9 GiB | 190.9 GiB | 1803% |
| 5 | `30398753-30401287` | 9:51.55 | PI 28.395s, BW6 1:17.65, BN254 3.858s | PI 1:29.07, BW6 1:03.65, BN254 13.668s | 72.9 GiB | 189.0 GiB | 1760% |

Average wall time: 10:19.26.
Average GPU prover phase sum: 2:48.67.
Average setup load sum: 2:42.42.

Optimized single aggregation reference, after local aggregation prover changes:

| Block range | Wall time | Setup load | GPU prover phases | Setup policy | Max RSS | CPU |
| --- | ---: | --- | --- | --- | ---: | ---: |
| `30391350-30394103` | 7:46.41 | PI 23.121s, BW6 32.729s, BN254 7.440s | PI 59.79s, BW6 55.65s, BN254 10.37s | strict GPU, canonical SRS only, BW6/BN254 setup prefetched under PI | 195.4 GiB | 2156% |
| `30391350-30394103` | 5:36.22 | PI 20.017s, BW6 30.541s, BN254 5.121s | PI 1:07.59, BW6 56.59s, BN254 10.49s | above, plus PI Vortex GPU MiMC for SIS-tree and no-SIS commitments, `GOMEMLIMIT=180GiB GOGC=75` | 199.2 GiB | 1669% |
| `30391350-30394103` | 4:56.04 | PI 20.511s, BW6 31.017s, BN254 5.148s | PI 1:07.09, BW6 56.15s, BN254 11.55s | above, plus PI Vortex GPU ring-SIS for commitments with at least 512 rows | 199.9 GiB | 1226% |
| `30391350-30394103` | 4:42.94 | PI 20.120s, BW6 30.737s, BN254 5.129s | PI 1:03.43, BW6 57.64s, BN254 10.32s | above, plus PI quotient coefficient cache across cosets | 203.5 GiB | 1155% |
| `30391350-30394103` | 4:39.77 | PI 20.297s, BW6 31.029s, BN254 5.265s | PI 1:03.00, BW6 55.38s, BN254 10.47s | above, plus quotient metadata/domain/map cleanup and timing breakdown | 201.6 GiB | 1168% |
| `30391350-30394103` | 4:41.91 | PI 20.217s, BW6 30.991s, BN254 5.321s | PI 1:03.60, BW6 55.49s, BN254 10.58s | above, plus cached BLS SIS FFT/MiMC static data and flattened SIS keys; single spot check, not included in the 3-proof average | 202.6 GiB | 1160% |
| `30391350-30394103` | 4:34.66 | PI 20.656s, BW6 31.815s, BN254 5.372s | PI 1:07.76, BW6 56.91s, BN254 10.71s | above, plus a 2-column CUDA tile for the large BLS SIS leaf kernel; single spot check, not included in the 3-proof average | 201.9 GiB | 1167% |

Final optimized 3-proof aggregation batch, rerun on 2026-05-05:

| Run | Block range | Wall time | Setup load | GPU prover phases | Max RSS | CPU |
| --- | --- | ---: | --- | --- | ---: | ---: |
| 1 | `30388561-30391349` | 4:44.51 | PI 51.196s, BW6 1:22.82, BN254 37.402s | PI 1:00.98, BW6 57.13s, BN254 11.36s | 201.8 GiB | 1156% |
| 2 | `30391350-30394103` | 4:44.21 | PI 20.733s, BW6 31.169s, BN254 5.336s | PI 1:03.53, BW6 58.14s, BN254 10.42s | 200.9 GiB | 1157% |
| 3 | `30394104-30396213` | 4:42.49 | PI 20.445s, BW6 30.919s, BN254 5.500s | PI 1:03.71, BW6 56.52s, BN254 10.67s | 199.8 GiB | 1163% |

Average wall time: 4:43.74.

All nine final aggregation sub-proofs used strict GPU Plonk
(`cpuFallback=false`) and passed their sanity checks.

The optimized run used `bin/prover` built with `GO_BUILD_TAGS=debug,cuda` and
`LINEA_PROVER_GPU_PLONK2=1 GNARK_GPU_PLONK2_LOG_MSM_PHASES=1`. The log reports
`loading canonical SRS only for strict gpu/plonk2 proving` for PI, BW6, and
BN254. No Lagrange SRS was loaded in GPU mode.

For the same request, the baseline GPU run was 9:02.14. Setup prefetch reduced
that request by 1:15.73 by hiding BW6 and BN254 setup loading under the PI
wizard/prover phase. The remaining critical path is the public-input wizard
proof before PI Plonk: it is dominated by repeated Vortex commitments,
especially SIS hashing and MiMC Merkleization. A specialized MiMC Merkle-tree
builder reduced Merkle allocations and shaved a few percent off individual
Merkleization calls, but it is not enough by itself to reach a 4 minute
wall-clock target. The next high-leverage optimization is GPU or algorithmic
acceleration of the PI Vortex/SIS commitment path.

The PI Vortex GPU MiMC run used the same request with
`LINEA_PROVER_GPU_PI_MIMC=1`. It adds CUDA kernels for BLS12-377 MiMC
Merkleization in the PI Vortex path and reuses them for no-SIS column
commitments. Against the optimized single reference, total wall time improved
from 7:46.41 to 5:36.22. The strict wizard window, measured from
`generating wizard proof for 6602 hashes from 7052 permutations` to the first
PI `Creating the witness`, improved from 5:05 to 3:06.

The PI Vortex GPU ring-SIS run additionally used `LINEA_PROVER_GPU_PI_SIS=1`.
The SIS offload is gated to commitments with at least 512 rows by default
(`LINEA_PROVER_GPU_PI_SIS_MIN_ROWS` overrides it). An unthresholded run was
slower because small SIS commitments paid about 12s of fixed GPU setup/copy
overhead. With the threshold, only the 1880-row and 812-row commitments used
the fused GPU SIS+MiMC tree path, reducing total wall time to 4:56.04 and the
strict wizard window to 2:26.

The 512-row threshold remains the default. Before the tiled SIS kernel, the
288-row SIS commitment was slower on GPU than CPU; after the tiled kernel it is
only break-even once Go-side materialization is included.

The quotient-cache run keeps the same BLS12-377 PI Vortex GPU SIS/MiMC path and
caches each quotient root column in coefficient form before evaluating cosets.
This avoids repeated inverse FFTs for the same roots across quotient actions.
It reduced total wall time to 4:42.94 and the strict wizard window to 2:16. The
first `quotient compute` timer dropped from 40.32s to 30.65s; the later quotient
computes were 2.85s and 0.55s.

The quotient metadata/domain/map cleanup removes hot `sync.Map` use during root
reevaluation, reuses one coset domain per global coset, and logs quotient
sub-timings. It reduced the best observed end-to-end wall time to 4:39.77 and
kept the strict wizard window at 2:15. The first quotient breakdown was:
coefficient cache 4.08s, coset reevaluation 17.71s, input prep 0.16s, and
expression evaluation / scaling / assignment 7.09s.

The cached-static-data spot check keeps the same retained algorithm and only
avoids rebuilding BLS SIS FFT tables, MiMC constants, and flattened SIS keys
inside repeated GPU commitment calls. It measured 4:41.91 on the same request,
which is within normal run-to-run noise of the quotient-map cleanup run; it is
kept as a small code cleanup rather than as a new benchmark headline.

A timing-instrumented diagnostic run on the same request measured 4:44.64 with
`LINEA_PROVER_GPU_PI_VORTEX_TIMINGS=1`. The extra instrumentation adds CUDA
synchronizations around internal phases, so this is diagnostic rather than a
headline benchmark. It showed the large fused BLS SIS commitments are now
compute-bound in the leaf kernel:

| Commitment | Total C time | H2D rows | SIS leaf kernel | MiMC tree | D2H |
| --- | ---: | ---: | ---: | ---: | ---: |
| 1880 rows x 524288 encoded cols | 17.33s | 3.12s | 14.14s | 12.5ms | 56.7ms |
| 812 rows x 524288 encoded cols | 14.16s | 0.91s | 12.71s | 12.5ms | 519.2ms |
| Warm SIS-digest MiMC tree only, 524288 leaves | 0.30s-0.32s | 0.10s-0.12s | 0.18s-0.19s | 12ms-13ms | 1.8ms-16ms |

The production-size GPU MiMC tree benchmark after Go-side tree materialization
cleanup measured about 328ms/op for 524288 leaves, close to the warmed CUDA
timing above. That confirms the next substantial win is not another static
table cache; it is a faster or device-resident BLS SIS/Vortex commitment design.

A follow-up 2-column CUDA tile for the fused BLS SIS leaf kernel keeps two
encoded columns in one 128-thread block. It preserves the same host-visible SIS
hashes and Merkle tree layout, but improves memory coalescing and reduces block
scheduling overhead. Production-shaped microbenchmarks with regular rows showed:

| Commitment | Before tile | 2-column tile | Leaf-kernel delta |
| --- | ---: | ---: | ---: |
| 1880 rows x 524288 encoded cols | 17.33s C total, 14.14s leaf | 12.09s C total, 8.44s leaf | 40.3% faster leaf |
| 812 rows x 524288 encoded cols | 14.16s C total, 12.71s leaf | 9.08s C total, 7.21s leaf | 43.3% faster leaf |

The same aggregation request then measured 4:34.66 wall clock without timing
instrumentation, and the strict wizard window from `generating wizard proof...`
to the PI `Creating the witness` log was about 2:03. A 4-column tile was also
tested after fixing lane indexing, but it regressed the same microbenchmarks to
14.40s and 17.62s C total for the 812-row and 1880-row cases respectively, so
the retained tile width is 2.

The faster kernel does not justify lowering the SIS GPU threshold globally. A
288-row regular-row microbenchmark measured 7.60s C total but 9.16s including
Go-side materialization, which is only break-even with the CPU SIS plus GPU
MiMC-tree path observed in the full proof. A 108-row regular-row case measured
6.17s C total and is slower than the CPU path. The default threshold therefore
stays at 512 rows.

### GPU Path Review

The retained GPU work is witness-side only. It does not modify circuit structs,
`Define()` methods, public-input layouts, verification keys, wizard constraint
registration, compiled-IOP checks, Fiat-Shamir ordering, or verifier logic. The
GPU kernels replace deterministic BLS12-377 field computations that already
existed on the prover side:

- BLS12-377 MiMC leaf and parent hashing for Vortex Merkle trees.
- BLS12-377 ring-SIS column hashing for large PI Vortex commitments.
- BLS12-377 quotient root coefficient caching and host-side coset
  reevaluation cleanup.

The cryptography-specific checks from the review were:

- MiMC uses the same 62 BLS12-377 constants from gnark-crypto and the same
  Miyaguchi-Preneel-style absorb relation as the CPU implementation.
- SIS is gated to the reviewed `degree=64` and `logTwoBound=16` parameters.
  Any different SIS shape falls back to CPU.
- The CUDA SIS path converts Montgomery field elements to canonical raw limbs
  only for the 16-bit limb decomposition, then stores SIS outputs back as
  Montgomery field elements before MiMC hashing. CPU/GPU tests compare the full
  SIS digest vector, Merkle leaves, internal nodes, and root.
- The 2-column tile uses a separate lane index for each tile; the earlier
  indexing bug found while testing 4-column tiling was fixed before retaining
  the 2-column kernel.
- Unsupported smartvector row shapes, non-power-of-two encoded widths, too many
  rows for the SIS key, CUDA errors, and allocation failures all fall back to
  the CPU path rather than producing a partial GPU commitment.

No soundness issue was found in the retained path. The main remaining risk is
engineering complexity rather than protocol semantics: the BLS PI Vortex path
still materializes encoded matrices on the host and only offloads selected
commitment work. The existing `gpu/vortex` package already demonstrates the
right architecture for KoalaBear, with a device-resident commit state, GPU
UAlpha computation, and selected-column extraction. That implementation cannot
be reused directly here because this PI path is BLS12-377 with MiMC/ring-SIS,
not KoalaBear with Poseidon2. The analogous next step is therefore a
BLS12-377 device-resident Vortex state:

- Commit stores encoded BLS rows or a column-major snapshot on the GPU.
- After the `Alpha` coin, UAlpha is computed on device without another full
  host/device transfer.
- After the `Q` coin, only selected encoded columns and selected SIS digests
  are downloaded for the opening and self-recursion assignments.
- Device buffers are freed before PI/BW6/BN254 Plonk proving starts, keeping
  the current strict GPU Plonk memory envelope intact.

This is the likely path to a larger wizard speedup. More local threshold tuning
or another static-data cache will not produce an order-of-magnitude gain.

In the final 3-proof aggregation batch, the first large quotient computation
was stable at 29.1s-29.6s. The dominant sub-step remained BLS12-377 coset
reevaluation, at 17.66s-18.00s for 2173 non-constant roots over a 262144
domain.

Strict wizard Vortex timing on the same request:

| Phase | Optimized reference | PI Vortex GPU MiMC | PI Vortex GPU SIS>=512 + MiMC | SIS>=512 + MiMC + quotient cache | Quotient map cleanup | Delta vs optimized reference |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| SIS Merkleization, 11 calls | 96.27s total, 8.75s avg | 3.64s total, 0.33s avg | 2.98s total, 0.27s avg | 2.96s total, 0.27s avg | 2.96s total, 0.27s avg | 32.5x faster |
| No-SIS MiMC commitments, 7 calls | 27.85s total, 3.98s avg | 1.63s total, 0.23s avg | 1.61s total, 0.23s avg | 1.95s total, 0.28s avg | 1.42s total, 0.20s avg | 19.6x faster |
| SIS hashing, 11 calls | 89.62s total | 89.99s total | 48.21s total | 48.12s total | 48.09s total | 1.9x faster |

The remaining blocker for an order-of-magnitude whole-wizard speedup is the
non-large SIS commitment work and residual quotient/coset evaluation. The fused
GPU SIS path is worthwhile on large matrices, but needs resident buffers or a
lower-overhead upload path before it should be applied to every SIS commitment.
For quotient, the largest measured sub-step is BLS12-377 coset reevaluation
FFT: 2173 non-constant roots over a 262144 domain took 17.71s. A useful GPU
quotient path needs a batched BLS12-377 coset FFT API that avoids one cgo call
and one H2D/D2H round trip per root. A one-root-at-a-time GPU quotient
prototype was tested and discarded because it serialized thousands of
host/device transfers and was slower than the parallel CPU path. A BLS12-377
variant using the existing `gpu/plonk2/bls12377` FFTs had to bit-reverse the
cached coefficients before `CosetFFT` to match the CPU `FFT(..., DIT,
OnCoset())` ordering, but still worsened the first large quotient
reevaluation from ~17.7s to 20.6s. A follow-up batched C API removed the
per-root cgo overhead and validated against the CPU transform, but it still
worsened the same quotient reevaluation to 27.5s because it loops over all
roots on-device and pays a large packed transfer. That prototype was removed
with the other unsuccessful experiments.

### Cost Per Proof

Pricing was refreshed for `us-east-2` Linux instances on 2026-05-05.
On-Demand prices came from the AWS Price List Bulk API current Amazon EC2 CSV
for `us-east-2`.

The previous GPU spot cost used `$0.7697/h`, which is a lowest-availability-zone
current value, not a robust regional average. AWS documents the public Spot
pricing table as the lowest price per instance type in a region, updated every
5 minutes:
<https://aws.amazon.com/ec2/spot/pricing/>. Cross-checks:

| Source | Window / basis | `g7e.8xlarge` Linux spot price |
| --- | --- | ---: |
| [CloudPrice](https://cloudprice.net/aws/spot-history?instanceType=g7e.8xlarge&region=us-east-2&period=30days) 30-day table, `us-east-2`, time-weighted equal-AZ average derived from 231 table rows | 2026-04-04 to 2026-05-03 | $1.2464/h |
| [CloudPrice](https://cloudprice.net/aws/spot-history?instanceType=g7e.8xlarge&region=us-east-2&period=30days) 7-day slice of the same table, time-weighted equal-AZ average | 2026-04-26 to 2026-05-03 | $1.1695/h |
| [Holori](https://calculator.holori.com/aws/ec2/g7e.8xlarge?region=us-east-2) `us-east-2` page | 24h average | $1.1002/h |
| [Holori](https://calculator.holori.com/aws/ec2/g7e.8xlarge?region=us-east-2) `us-east-2` page | current AZs | $0.8047/h in `us-east-2b`, $1.4271/h in `us-east-2a` |
| [DoiT](https://compute.doit.com/spot/us-east-2/g7e.8xlarge) `us-east-2` page | current AZs | $0.7697/h in `us-east-2b`, $1.4094/h in `us-east-2a` |
| [aws-pricing.com](https://aws-pricing.com/g7e.8xlarge.html) | current region listing, last updated 2026-05-02 | $0.8642/h |
| [aws-pricing.com](https://aws-pricing.com/g7e.8xlarge.html) | all listed regions average | $1.1440/h |

The CPU host spot rate was cross-checked the same way:

| Source | Window / basis | `r8a.24xlarge` Linux spot price |
| --- | --- | ---: |
| [CloudPrice](https://cloudprice.net/aws/spot-history?instanceType=r8a.24xlarge&region=us-east-2&period=30days) 30-day table, `us-east-2`, time-weighted equal-AZ average derived from 335 table rows | 2026-04-04 to 2026-05-03 | $3.3873/h |
| [Holori](https://calculator.holori.com/aws/ec2/r8a.24xlarge?region=us-east-2) `us-east-2` page | 24h average | $3.1700/h |
| [Holori](https://calculator.holori.com/aws/ec2/r8a.24xlarge?region=us-east-2) `us-east-2` page | current AZs | $2.7274/h in `us-east-2b`, $2.9796/h in `us-east-2a`, $3.8883/h in `us-east-2c` |
| [DoiT](https://compute.doit.com/spot/us-east-2/r8a.24xlarge) `us-east-2` page | current AZs | $2.7384/h in `us-east-2b`, $2.9822/h in `us-east-2a`, $3.8439/h in `us-east-2c` |

Cost rows below use CloudPrice 30-day equal-AZ averages for spot rates on both
hosts. Lowest-AZ current rates are still useful for short opportunistic runs,
but they are too optimistic as standing benchmark costs.

| Proof | Host | Avg wall | On-Demand $/h | On-Demand $/proof | Spot $/h | Spot $/proof |
| --- | --- | ---: | ---: | ---: | ---: | ---: |
| Compression | `CPU r8a.24xlarge` | 4:40.31 | $7.66848 | $0.5971 | $3.3873 | $0.2637 |
| Compression, baseline | `GPU g7e.8xlarge` | 4:36.04 | $5.26824 | $0.4040 | $1.2464 | $0.0956 |
| Compression, optimized single reference | `GPU g7e.8xlarge` | 2:35.68 | $5.26824 | $0.2278 | $1.2464 | $0.0539 |
| Compression, final optimized 3-proof avg | `GPU g7e.8xlarge` | 2:53.42 | $5.26824 | $0.2538 | $1.2464 | $0.0600 |
| Aggregation | `CPU r8a.24xlarge` | 6:46.75 | $7.66848 | $0.8664 | $3.3873 | $0.3827 |
| Aggregation | `GPU g7e.8xlarge` | 10:19.26 | $5.26824 | $0.9062 | $1.2464 | $0.2144 |
| Aggregation, optimized single reference | `GPU g7e.8xlarge` | 7:46.41 | $5.26824 | $0.6825 | $1.2464 | $0.1615 |
| Aggregation, PI Vortex GPU MiMC | `GPU g7e.8xlarge` | 5:36.22 | $5.26824 | $0.4920 | $1.2464 | $0.1164 |
| Aggregation, PI Vortex GPU SIS>=512 + MiMC | `GPU g7e.8xlarge` | 4:56.04 | $5.26824 | $0.4332 | $1.2464 | $0.1025 |
| Aggregation, PI Vortex GPU SIS>=512 + MiMC + quotient cache | `GPU g7e.8xlarge` | 4:42.94 | $5.26824 | $0.4141 | $1.2464 | $0.0980 |
| Aggregation, PI Vortex GPU SIS>=512 + MiMC + quotient map cleanup | `GPU g7e.8xlarge` | 4:39.77 | $5.26824 | $0.4094 | $1.2464 | $0.0969 |
| Aggregation, PI Vortex 2-column SIS tile spot check | `GPU g7e.8xlarge` | 4:34.66 | $5.26824 | $0.4019 | $1.2464 | $0.0951 |
| Aggregation, final optimized 3-proof avg | `GPU g7e.8xlarge` | 4:43.74 | $5.26824 | $0.4152 | $1.2464 | $0.0982 |

Savings versus the CPU reference, using final optimized 3-proof GPU averages:

| Proof | Time saved | Time saved % | On-Demand $ saved / proof | On-Demand saved % | Spot $ saved / proof | Spot saved % |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Compression | 1:46.89 | 38.1% | $0.3433 | 57.5% | $0.2037 | 77.2% |
| Aggregation | 2:03.01 | 30.2% | $0.4512 | 52.1% | $0.2845 | 74.3% |

At 10,000 compression proofs plus 10,000 aggregation proofs per month, those
per-proof deltas imply about $7,945/month on On-Demand or $4,882/month on Spot
in compute savings, before considering scheduling and operational effects.

The baseline GPU aggregation Plonk phases are faster than the CPU Plonk phases,
but the initial end-to-end aggregation benchmark was slower on `g7e.8xlarge`
because the host has fewer CPU cores and setup/public-input work dominates the
wall clock. With setup prefetching and PI Vortex GPU MiMC enabled, the single
reference aggregation request is faster than the CPU reference row, though
small ring-SIS commitments and residual quotient work remain CPU-bound.

## Raw Artifacts

Generated responses:

- `reference-benchmarks/results/2026-05-04-r8a-24xlarge-provertestdata2-7.1.0/compression/`
- `reference-benchmarks/results/2026-05-04-r8a-24xlarge-provertestdata2-7.1.0/aggregation/`
- `reference-benchmarks/results/2026-05-04-g7e-8xlarge-gpu-plonk2-provertestdata2-7.1.0/compression/`
- `reference-benchmarks/results/2026-05-04-g7e-8xlarge-gpu-plonk2-provertestdata2-7.1.0/aggregation/`
- `reference-benchmarks/results/2026-05-04-g7e-8xlarge-gpu-plonk2-optimize-bls12377-pass8-msm-window20-cuda/`
- `reference-benchmarks/results/2026-05-04-g7e-8xlarge-gpu-aggregation-setup-prefetch/`
- `reference-benchmarks/results/2026-05-04-g7e-8xlarge-gpu-pi-mimc-all-gomem180/`
- `reference-benchmarks/results/2026-05-04-g7e-8xlarge-gpu-pi-sis-threshold512-mimc-gomem180/`
- `reference-benchmarks/results/2026-05-05-g7e-8xlarge-gpu-pi-sis-threshold512-mimc-quotient-cache-gomem180/`
- `reference-benchmarks/results/2026-05-05-g7e-8xlarge-gpu-pi-sis-threshold512-mimc-quotient-map-gomem180/`
- `reference-benchmarks/results/2026-05-05-g7e-8xlarge-gpu-pi-sis-threshold512-static-cache-gomem180/`
- `reference-benchmarks/results/2026-05-05-g7e-8xlarge-gpu-pi-vortex-timing-gomem180/`
- `reference-benchmarks/results/2026-05-05-g7e-8xlarge-gpu-pi-sis-tiled2-gomem180/`
- `reference-benchmarks/results/2026-05-05-g7e-8xlarge-gpu-final-3x/compression/`
- `reference-benchmarks/results/2026-05-05-g7e-8xlarge-gpu-final-3x/aggregation/`

Raw prover logs and GNU `time -v` outputs:

- `reference-benchmarks/results/2026-05-04-r8a-24xlarge-provertestdata2-7.1.0/logs/`
- `reference-benchmarks/results/2026-05-04-g7e-8xlarge-gpu-plonk2-provertestdata2-7.1.0/logs/`
- `reference-benchmarks/results/2026-05-04-g7e-8xlarge-gpu-aggregation-setup-prefetch/logs/`
- `reference-benchmarks/results/2026-05-04-g7e-8xlarge-gpu-pi-mimc-all-gomem180/logs/`
- `reference-benchmarks/results/2026-05-04-g7e-8xlarge-gpu-pi-sis-threshold512-mimc-gomem180/logs/`
- `reference-benchmarks/results/2026-05-05-g7e-8xlarge-gpu-pi-sis-threshold512-mimc-quotient-cache-gomem180/logs/`
- `reference-benchmarks/results/2026-05-05-g7e-8xlarge-gpu-pi-sis-threshold512-mimc-quotient-map-gomem180/logs/`
- `reference-benchmarks/results/2026-05-05-g7e-8xlarge-gpu-pi-sis-threshold512-static-cache-gomem180/logs/`
- `reference-benchmarks/results/2026-05-05-g7e-8xlarge-gpu-pi-vortex-timing-gomem180/logs/`
- `reference-benchmarks/results/2026-05-05-g7e-8xlarge-gpu-pi-sis-tiled2-gomem180/logs/`
- `reference-benchmarks/results/2026-05-05-g7e-8xlarge-gpu-final-3x/logs/`

## Diagnostic Notes

- `/home/ubuntu/provertestdata2` responses are 7.1.0-compatible. A 7.0.7
  aggregation attempt rejected the execution response verifying key
  `0xaf523823aefa029f83c1f5e740311a96fa02083deb204cc35ac25e63b4f69e19`.
- A temporary config with `aggregation.num_proofs = [100]` loaded
  `aggregation-100` at `setupPos=0` and failed BN254 emulation verification.
  The benchmark config keeps `[10, 20, 50, 100, 200, 400]`, making
  `aggregation-100` use `setupPos=3`.
- The older `/home/ubuntu/provertestdata` source did not include the execution
  response files needed by its aggregation requests, so it was not used for the
  current aggregation baseline.
