package vortex

import (
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/utils"
)

// EstimateOpeningProofSizeBytes returns an upper-bound estimate of the
// serialized size (in bytes) of the Vortex opening proof that this [Ctx] will
// produce after compilation.
//
// The estimate covers:
//   - The Reed-Solomon encoded linear combination (numEncodedCols extension-field elements,
//     each [fext.ExtensionDegree] × [field.Bytes] bytes)
//   - The opened column entries (numColsToOpen × totalRows base-field elements)
//   - The Merkle proof siblings (depth × numCommitments × numColsToOpen base-field elements)
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

	// Linear combination codeword: one extension-field element per encoded column.
	// Each extension-field element spans ExtensionDegree base-field elements.
	linearCombSize := numEncodedCols * fext.ExtensionDegree * field.Bytes

	// Opened columns: for each opened column position, for each row across all
	// commitments, one base-field element.
	openedColsSize := numColsToOpen * totalRows * field.Bytes

	// Merkle proofs: for each commitment and each opened column position, one
	// Merkle proof of `depth` siblings (each a base-field element) plus a 4-byte
	// path integer.
	merkleProofSize := numCommitments * numColsToOpen * (depth*field.Bytes + 4)

	return linearCombSize + openedColsSize + merkleProofSize
}
