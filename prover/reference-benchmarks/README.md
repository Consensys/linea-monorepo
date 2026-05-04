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

## Raw Artifacts

Generated responses:

- `reference-benchmarks/results/2026-05-04-r8a-24xlarge-provertestdata2-7.1.0/compression/`
- `reference-benchmarks/results/2026-05-04-r8a-24xlarge-provertestdata2-7.1.0/aggregation/`

Raw prover logs and GNU `time -v` outputs:

- `reference-benchmarks/results/2026-05-04-r8a-24xlarge-provertestdata2-7.1.0/logs/`

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
