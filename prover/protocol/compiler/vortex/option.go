package vortex

import (
	"github.com/consensys/linea-monorepo/prover/crypto/ringsis"
	"github.com/consensys/linea-monorepo/prover/maths/field"
)

// Option to be passed to vortex
type VortexOp func(ctx *Ctx)

// ForceNumTotalColumns pads the Vortex committed-row count to at least n by
// inserting zero-valued shadow polynomials into the last committed round.
// This ensures that circuit loops depending on CommittedRowsCount have a fixed
// upper bound, making the gnark verifier circuit size uniform across tree depths.
// If the actual count already meets or exceeds n, no padding is added.
func ForceNumTotalColumns(n int) VortexOp {
	return func(ctx *Ctx) {
		ctx.MinTotalCommittedCols = n
	}
}

// ForceNumTotalRounds pads the number of committed IOP rounds to at least n by
// converting empty rounds into dummy committed rounds (each with one zero-valued
// shadow polynomial). This fixes MerkleProofSize and Merkle-root proof column
// counts, making the gnark verifier circuit size uniform across tree depths.
// Empty rounds within [0, MaxCommittedRound] are filled first.
// If the actual count already meets or exceeds n, no padding is added.
func ForceNumTotalRounds(n int) VortexOp {
	return func(ctx *Ctx) {
		ctx.MinTotalCommittedRounds = n
	}
}

// Overrides the number of opened columns (should
// not be used in production)
func ForceNumOpenedColumns(nbCol int) VortexOp {
	return func(ctx *Ctx) {
		ctx.NumOpenedCol = nbCol
	}
}

// Allows passing a SIS instance
func WithSISParams(params *ringsis.Params) VortexOp {
	return func(ctx *Ctx) {
		ctx.SisParams = params
	}
}

// Allows skipping the SIS hashing of columns of the round matrices
// if the number of polynomials to commit to for the particular round
// is less than the threshold
func WithOptionalSISHashingThreshold(sisHashingThreshold int) VortexOp {
	return func(ctx *Ctx) {
		ctx.ApplySISHashThreshold = sisHashingThreshold
	}
}

// ForceNumPrecomputed pads the precomputed column count to at least n by
// inserting zero-valued shadow polynomials. This normalizes the precomputed
// column count across tree depths, ensuring that totalCommitted (dynamic +
// precomputed) in the first Vortex round is depth-independent, which keeps
// the BN254 wrap circuit constraint count constant.
// If the actual count already meets or exceeds n, no padding is added.
func ForceNumPrecomputed(n int) VortexOp {
	return func(ctx *Ctx) {
		ctx.MinTotalPrecomputedCols = n
	}
}

// PremarkAsSelfRecursed marks the ctx as selfrecursed. This is useful
// toward conglomerating the receiver comp but is not needed for
// self-recursion or full-recursion.
func PremarkAsSelfRecursed() VortexOp {
	return func(ctx *Ctx) {
		ctx.IsSelfrecursed = true
	}
}

// AddMerkleRootToPublicInputs tells the compiler to additionally adds
// a merkle root to the public inputs of the comp. This is useful for
// the distributed prover. The name argument is used to set the Name
// field of the public-input.
func AddMerkleRootToPublicInputs(name string, round []int) VortexOp {
	return func(ctx *Ctx) {
		ctx.AddMerkleRootToPublicInputsOpt = struct {
			Enabled bool
			Name    string
			Round   []int
		}{Enabled: true, Name: name, Round: round}
	}
}

// AddPrecomputedMerkleRootToPublicInputs tells the compiler to adds
// a precomputed merkle root to the public inputs of the comp. This is
// useful for the distributed prover. The name argument is used to set
// the Name field of the public-input.
func AddPrecomputedMerkleRootToPublicInputs(name string) VortexOp {
	return func(ctx *Ctx) {
		ctx.AddPrecomputedMerkleRootToPublicInputsOpt = struct {
			Enabled          bool
			Name             string
			PrecomputedValue [blockSize]field.Element
		}{Enabled: true, Name: name}
	}
}
