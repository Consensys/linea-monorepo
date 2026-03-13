package vortex

import (
	"github.com/consensys/linea-monorepo/prover/crypto/encoding"

	"github.com/consensys/linea-monorepo/prover/crypto/ringsis"
	"github.com/consensys/linea-monorepo/prover/maths/field"
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
func WithOptionalSISHashingThreshold(sisHashingThreshold int) VortexOp {
	return func(ctx *Ctx) {
		ctx.ApplySISHashThreshold = sisHashingThreshold
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
			Enabled             bool
			Name                string
			PrecomputedValue    [blockSize]field.Element
			PrecomputedBLSValue [encoding.KoalabearChunks]field.Element
		}{Enabled: true, Name: name}
	}
}

// WithUAlphaCoefficients instructs the Vortex compiler to send the polynomial
// coefficients of U_alpha (T E4 elements) instead of the full evaluation form
// (T×RS E4 elements). This reduces proof cells by ~(RS-1)/RS × T × 4 cells
// while maintaining the same security level via the column consistency checks.
// For T=4096 and RS=16 this saves 245,760 proof cells (~40% reduction overall).
func WithUAlphaCoefficients() VortexOp {
	return func(ctx *Ctx) {
		ctx.UseUAlphaCoefficients = true
	}
}

// SkipSelfRecursionProofColumns suppresses the registration of the
// OpenedSISColumns and OpenedNonSISColumns proof columns. These are needed
// only when a SelfRecurse step will follow this Vortex compilation. Passing
// this option to the outermost / final Vortex (where no further self-recursion
// occurs) eliminates dead-weight proof cells: the verifier never reads these
// columns, yet the prover would otherwise pay K × NextPowerOfTwo(rows) cells.
func SkipSelfRecursionProofColumns() VortexOp {
	return func(ctx *Ctx) {
		ctx.SkipSelfRecursionProofCols = true
	}
}

// SkipPrecomputedMerkleProof omits the precomputed columns from the Merkle
// column-inclusion check. The precomputed polynomial evaluations are already
// authenticated by the Schwartz-Zippel linear-combination check (the Ys used
// in the verifier come directly from the wizard's own evaluation of the
// precomputed polynomials, not from the proof). The extra Merkle proof is
// therefore redundant and can be safely dropped. Skipping it reduces
// MerkleProofSize when the removal of one commitment slot causes the product
// depth × numComs × K to cross a power-of-two boundary downward.
func SkipPrecomputedMerkleProof() VortexOp {
	return func(ctx *Ctx) {
		ctx.SkipPrecomputedMerkleProof = true
	}
}
