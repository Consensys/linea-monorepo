// CPU fallback stubs when CUDA is not available.

//go:build !cuda

package vortex

import (
	"github.com/consensys/linea-monorepo/prover/crypto/state-management/smt_koalabear"
	"github.com/consensys/linea-monorepo/prover/crypto/vortex/vortex_koalabear"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
)

// PreWarmGPU is a no-op without CUDA.
func PreWarmGPU(_ *vortex_koalabear.Params) {}

// EvictPipelineCache is a no-op without CUDA.
func EvictPipelineCache() {}

// CommitMerkleWithSIS delegates to the CPU implementation when CUDA is not available.
func CommitMerkleWithSIS(
	params *vortex_koalabear.Params,
	polysMatrix []smartvectors.SmartVector,
) (vortex_koalabear.EncodedMatrix, vortex_koalabear.Commitment, *smt_koalabear.Tree, []field.Element) {
	return params.CommitMerkleWithSIS(polysMatrix)
}

// CommitSIS delegates to the CPU implementation when CUDA is not available.
// The needSISHashes parameter is ignored (CPU always computes hashes).
func CommitSIS(
	params *vortex_koalabear.Params,
	polysMatrix []smartvectors.SmartVector,
	needSISHashes bool,
) (*CommitState, *smt_koalabear.Tree, []field.Element) {
	encoded, _, tree, colHashes := params.CommitMerkleWithSIS(polysMatrix)
	cs := &CommitState{encodedMatrix: encoded, nRows: len(polysMatrix)}
	return cs, tree, colHashes
}
