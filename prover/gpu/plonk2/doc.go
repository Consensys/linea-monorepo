// Package plonk2 is the curve-generic GPU PlonK prover used by the
// linea-monorepo prover binary.
//
// Layout:
//
//	plonk2/                   — multi-curve dispatcher (this package)
//	plonk2/bls12377/          — BLS12-377 prover, used for compression
//	plonk2/bn254/             — BN254 prover, used for aggregation BN254 emulation
//	plonk2/bw6761/            — BW6-761 prover, used for aggregation
//
// The three per-curve packages are produced by gpu/internal/generator/plonk
// from a shared template. Re-emit them with `go run ./gpu/internal/generator`
// after editing the templates; the curve files are otherwise identical.
//
// Build tags:
//
//	cuda     — links against gpu/cuda/build/libgnark_gpu.a; full GPU acceleration
//	!cuda    — stub types; the dispatcher falls back to gnark's CPU prover
//
// Design constraints:
//
//   - SoA layout for GPU field vectors (coalesced limb access in CUDA).
//   - AoS Montgomery layout for host buffers (matches gnark-crypto).
//   - One CUDA context per Device; the top-level gpu package owns lifecycle.
//   - Pinned host staging buffers reused across rounds (see pinned_fr.go and
//     prove.go's persistent work-buffer scope).
//   - All multi-stream work (FFT/MSM/permutation) drains via the device's
//     compute stream before any cross-stream sync.
//
// Activation in the linea-monorepo prover:
//
//   - Compression auto-enables this prover whenever a GPU is reachable, via
//     circuits.WithGPU(true) plumbed from backend/dataavailability/prove.go.
//   - Aggregation only uses it when the operator opts in via the master flag
//     LINEA_PROVER_GPU_AGGREGATION=1 (see backend/aggregation/prove.go).
//
// See gpu/plonk2/bls12377/prove.go for the per-curve top-level prover entry
// point — that is the right starting place when reviewing the GPU PlonK
// pipeline.
package plonk2
