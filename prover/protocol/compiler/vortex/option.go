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

// WithStreamingCommitment enables the streaming SIS commitment path (Level 1)
// which processes rows in batches for better cache efficiency during Ring-SIS
// hashing. The full encoded matrix W' is still materialized and stored for the
// linear combination and opening phases.
// batchSize controls how many rows are processed at a time. If batchSize <= 0,
// a default of max(1, NbRows/8) is used.
func WithStreamingCommitment(batchSize int) VortexOp {
	return func(ctx *Ctx) {
		ctx.StreamingCommitment = true
		ctx.StreamingBatchSize = batchSize
	}
}

// WithStreamingCommitmentL2 enables Level 2 streaming: the full encoded matrix
// W' is never materialized. During commitment, rows are RS-encoded in batches
// and discarded after SIS hashing. During linear combination and column opening,
// rows are re-encoded on demand from the original matrix W. This eliminates
// ~50% of peak memory (at rate=2) at the cost of one additional RS encoding pass.
func WithStreamingCommitmentL2(batchSize int) VortexOp {
	return func(ctx *Ctx) {
		ctx.StreamingCommitment = true
		ctx.StreamingBatchSize = batchSize
		ctx.StreamingNoMaterialize = true
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
