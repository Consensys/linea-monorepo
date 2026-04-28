// Package plonk2 is the validation-first GPU foundation for curve-generic
// PlonK proving code.
//
// The existing gpu/plonk package is a production-oriented BLS12-377 prover
// path. It uses fixed-width CUDA field arithmetic and a BLS12-377 G1 MSM in
// twisted Edwards coordinates. This package intentionally starts lower in the
// stack: it exposes curve-indexed scalar-field vectors, NTT domains, PlonK
// quotient kernels, and a short-Weierstrass affine MSM/KZG commitment backend
// for the curves needed by recursive proof composition. Tests compare every
// operation against gnark-crypto.
//
// Design constraints:
//   - keep the C ABI flat and curve-indexed;
//   - store GPU field vectors in SoA layout for coalesced limb access;
//   - keep host buffers in gnark-crypto AoS Montgomery layout;
//   - share CUDA context, streams, and staging memory with the top-level gpu
//     package;
//   - avoid curve-specific Go wrappers where a small curve descriptor suffices.
//
// The generic MSM is correctness-first: it implements the production data
// model and all-curve KZG semantics, but its reduction kernels are not yet
// competitive with gpu/plonk's BLS12-377 twisted-Edwards path. Full GPU PlonK
// proof generation still needs to wire these primitives under the existing
// prover orchestration.
package plonk2
