package vortex

import (
	"github.com/consensys/linea-monorepo/prover/crypto/ringsis"
)

// Option to be passed to vortex
type VortexOp func(ctx *Ctx)

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
func WithOptionalSISHashingThreshold(sisHashingTHreshold int) VortexOp {
	return func(ctx *Ctx) {
		ctx.ApplySISHashingThreshold = sisHashingTHreshold
	}
}

// Replace SIS with a custom hasher
func ReplaceSisByMimc() VortexOp {
	return func(ctx *Ctx) {
		ctx.ReplaceSisByMimc = true
		ctx.SisParams = nil
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
func AddMerkleRootToPublicInputs(name string, round int) VortexOp {
	return func(ctx *Ctx) {
		ctx.AddMerkleRootToPublicInputsOpt = struct {
			Enabled bool
			Name    string
			Round   int
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
			Enabled bool
			Name    string
		}{Enabled: true, Name: name}
	}
}
