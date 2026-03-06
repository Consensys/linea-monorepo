package vortex

import (
	"github.com/consensys/linea-monorepo/prover/utils"
)

const (
	// fieldElementSizeBytes is the number of bytes used to represent a single
	// field element (BLS12-377 scalar field element = 32 bytes).
	fieldElementSizeBytes = 32
)

// EstimateOpeningProofSizeBytes returns an upper-bound estimate of the
// serialized size (in bytes) of the Vortex opening proof that this [Ctx] will
// produce after compilation.
//
// The estimate covers:
//   - The Reed-Solomon encoded linear combination (numEncodedCols field elements)
//   - The opened column entries (numColsToOpen × totalRows field elements)
//   - The Merkle proof siblings (depth × numCommitments × numColsToOpen × 32 bytes)
//     plus an int for each path
//
// This method must be called after [generateVortexParams] has been invoked,
// i.e., after [Compile] has finished processing.
func (ctx *Ctx) EstimateOpeningProofSizeBytes() int {
	var (
		numEncodedCols = ctx.NumEncodedCols()
		numColsToOpen  = ctx.NbColsToOpen()
		totalRows      = utils.NextPowerOfTwo(ctx.CommittedRowsCount)
		depth          = utils.Log2Ceil(numEncodedCols)
		numCommitments = ctx.NumCommittedRounds()
	)

	// Account for the precomputed columns commitment
	if ctx.IsNonEmptyPrecomputed() {
		numCommitments++
		totalRows += utils.NextPowerOfTwo(len(ctx.Items.Precomputeds.PrecomputedColums))
	}

	// Linear combination codeword: one field element per encoded column
	linearCombSize := numEncodedCols * fieldElementSizeBytes

	// Opened columns: for each opened column position, for each row across all
	// commitments, one field element.
	openedColsSize := numColsToOpen * totalRows * fieldElementSizeBytes

	// Merkle proofs: for each commitment and each opened column position, one
	// Merkle proof of `depth` siblings (32 bytes each) plus a 4-byte path int.
	merkleProofSize := numCommitments * numColsToOpen * (depth*fieldElementSizeBytes + 4)

	return linearCombSize + openedColsSize + merkleProofSize
}
