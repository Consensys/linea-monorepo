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

### Cost Per Proof

Pricing was refreshed for `us-east-2` Linux instances on 2026-05-04.
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
| Compression, optimized | `GPU g7e.8xlarge` | 2:35.68 | $5.26824 | $0.2278 | $1.2464 | $0.0539 |
| Aggregation | `CPU r8a.24xlarge` | 6:46.75 | $7.66848 | $0.8664 | $3.3873 | $0.3827 |
| Aggregation | `GPU g7e.8xlarge` | 10:19.26 | $5.26824 | $0.9062 | $1.2464 | $0.2144 |
| Aggregation, optimized single reference | `GPU g7e.8xlarge` | 7:46.41 | $5.26824 | $0.6825 | $1.2464 | $0.1615 |

The GPU aggregation Plonk phases are faster than the CPU Plonk phases, but the
end-to-end aggregation benchmark is slower on `g7e.8xlarge` because the host
has fewer CPU cores and setup/public-input work dominates the wall clock.

## Raw Artifacts

Generated responses:

- `reference-benchmarks/results/2026-05-04-r8a-24xlarge-provertestdata2-7.1.0/compression/`
- `reference-benchmarks/results/2026-05-04-r8a-24xlarge-provertestdata2-7.1.0/aggregation/`
- `reference-benchmarks/results/2026-05-04-g7e-8xlarge-gpu-plonk2-provertestdata2-7.1.0/compression/`
- `reference-benchmarks/results/2026-05-04-g7e-8xlarge-gpu-plonk2-provertestdata2-7.1.0/aggregation/`
- `reference-benchmarks/results/2026-05-04-g7e-8xlarge-gpu-plonk2-optimize-bls12377-pass8-msm-window20-cuda/`
- `reference-benchmarks/results/2026-05-04-g7e-8xlarge-gpu-aggregation-setup-prefetch/`

Raw prover logs and GNU `time -v` outputs:

- `reference-benchmarks/results/2026-05-04-r8a-24xlarge-provertestdata2-7.1.0/logs/`
- `reference-benchmarks/results/2026-05-04-g7e-8xlarge-gpu-plonk2-provertestdata2-7.1.0/logs/`
- `reference-benchmarks/results/2026-05-04-g7e-8xlarge-gpu-aggregation-setup-prefetch/logs/`

The logs include one failed diagnostic aggregation attempt for run 4:
`*.failed-runtime-fault-attempt1.*`. It hit a Go runtime SIGSEGV in
`gnark-crypto` BLS12-377 vector assembly during PI proof interpolation. The
same request succeeded on retry with the same benchmark settings, so the
successful retry is the row reported above.

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
