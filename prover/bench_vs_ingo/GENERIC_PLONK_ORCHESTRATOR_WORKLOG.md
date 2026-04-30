# Generic PlonK GPU Orchestrator Worklog

Date: 2026-04-30

Scope: `gpu/plonk2` generic prepared-state orchestration and solved-wire
Lagrange commitment waves for BN254, BLS12-377, and BW6-761. Full generic
proof generation is still not wired for every curve; this work targets the
generic GPU state and the L/R/O commitment wave that every generic prover port
will depend on.

## Flow

The prepared prover now treats memory in explicit phases:

1. Prepare persistent state once per proving key:
   - canonical SRS point table
   - Lagrange SRS point table
   - FFT twiddles
   - permutation table
   - fixed selector and permutation polynomials
2. Pin only the MSM scratch needed for a commitment wave.
3. Commit L/R/O solved-wire vectors against the resident Lagrange SRS.
4. Release MSM scratch before later quotient/permutation phases need VRAM.

`NewProver` and `Prove` bind the calling OS thread to the selected CUDA device
before allocating or launching kernels. This matters on multi-GPU hosts because
CUDA's current device is thread-local.

## Fixes

- `gnark_gpu_plonk2_msm_create` no longer allocates MSM scratch eagerly. It
  uploads resident points only; sort/key/bucket buffers are allocated lazily or
  explicitly by `PinWorkBuffers`.
- The memory plan now counts both canonical and Lagrange SRS residency, plus
  the large-bucket overflow metadata used by the MSM accumulator.
- The solved-wire wave benchmark now reports per-wire MSM phase timings, so a
  slow L/R/O wire cannot be hidden by the last commitment in a batch.
- The generic short-Weierstrass MSM now has a bounded two-phase bucket
  accumulator:
  - serial phase processes at most a dynamic cap per bucket
  - overflow buckets are compacted into a device list
  - a block-parallel tree reduction handles each large-bucket tail

The large-bucket fix is required for real PlonK wire distributions. Random
MSM inputs were already fast, but solved wires can concentrate millions of
entries into a small number of buckets.

## Benchmarks

Command shape:

```bash
PLONK2_PLONK_BENCH_CONSTRAINTS=<steps> \
go test ./gpu/plonk2 -tags cuda,nocorset -run '^$' \
  -bench '^BenchmarkGenericSolvedWireCommitmentWave_CUDA/<curve>' \
  -benchmem -benchtime=3x -count=1 -timeout=<timeout>
```

The chain benchmark has two constraints per step. Inputs used:

| Requested domain | Benchmark steps | Reported constraints |
|---:|---:|---:|
| `1<<23` | `4194302` | `8388604` |
| `1<<24` | `8388606` | `16777212` |

Results, averaged over three timed L/R/O waves:

| Curve | Domain | Wave Time | Planned Peak | Per-Wave Scratch | Persistent |
|---|---:|---:|---:|---:|---:|
| BN254 | `1<<23` | 447.968 ms | 10.740 GiB | 4.303 GiB | 3.438 GiB |
| BN254 | `1<<24` | 871.679 ms | 21.428 GiB | 8.553 GiB | 6.875 GiB |
| BLS12-377 | `1<<23` | 1140.263 ms | 11.264 GiB | 4.326 GiB | 3.938 GiB |
| BLS12-377 | `1<<24` | 2144.233 ms | 22.451 GiB | 8.576 GiB | 7.875 GiB |
| BW6-761 | `1<<23` | 4872.834 ms | 17.457 GiB | 6.394 GiB | 6.563 GiB |
| BW6-761 | `1<<24` | 9131.384 ms | 34.144 GiB | 12.019 GiB | 13.125 GiB |

Raw output and per-phase timings are in:

- `raw/plonk2_generic_orchestrator_bn254_domain_2p23_20260430.txt`
- `raw/plonk2_generic_orchestrator_bn254_domain_2p24_20260430.txt`
- `raw/plonk2_generic_orchestrator_bls12377_domain_2p23_20260430.txt`
- `raw/plonk2_generic_orchestrator_bls12377_domain_2p24_20260430.txt`
- `raw/plonk2_generic_orchestrator_bw6761_domain_2p23_20260430.txt`
- `raw/plonk2_generic_orchestrator_bw6761_domain_2p24_20260430.txt`

## Interpretation

- BN254 scales almost linearly from `1<<23` to `1<<24` and is now stable.
- BLS12-377 is slower mainly because G1 arithmetic uses a six-limb base field.
- BW6-761 remains the largest cost center. The catastrophic serial large-bucket
  behavior is removed, but 12-limb Jacobian additions dominate the wave.
- Planned peak memory is under 35 GiB even for BW6-761 at `1<<24` on this
  benchmark, leaving headroom on the 97 GiB RTX PRO 6000 used here.

## Multi-GPU Scaling Notes

The current safe scaling model is one prepared prover per GPU. State is not
shared between devices: each GPU owns its CUDA context, resident SRS tables,
FFT domain, fixed polynomials, and pinned MSM scratch. This is the right
isolation boundary for several concurrent segments or proofs.

Memory scales per GPU, not globally. For `k` GPUs running independent proofs,
host memory and SRS loading scale roughly by `k`, while wall-clock throughput
can scale if the scheduler assigns independent proof segments to devices and
keeps each goroutine pinned to its OS thread while using that device.

The host used for this work exposes one GPU, so multi-GPU throughput was not
measured here. The device binding changes are in place for correctness on
multi-GPU hosts.

## Remaining Work

1. Port the full PlonK proof orchestration to the generic backend for all
   curves. Today only BLS12-377 has the legacy full-GPU bridge.
2. Reduce BW6-761 bucket-add cost with generated/specialized 12-limb G1
   addition and squaring code. The large-bucket scheduler is no longer the
   fundamental bottleneck.
3. Add an optional benchmark/test that runs two independent generic prepared
   provers on two GPUs when `LIMITLESS_GPU_COUNT>=2`.
4. For full-proof orchestration, enforce a strict phase memory contract:
   release MSM scratch before quotient vectors, repin before commitment waves,
   and reject plans whose peak exceeds a configured per-device limit.
