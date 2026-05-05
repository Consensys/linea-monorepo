//go:build !cuda

package vortex

import (
	"fmt"

	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/crypto/ringsis"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/crypto/state-management/smt"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/maths/field"
)

func tryCommitSISGPU(_ EncodedMatrix, _ *ringsis.Key) (*smt.Tree, []field.Element, bool) {
	return nil, nil, false
}

func tryBuildSISMiMCTreeGPU(_ []field.Element, _ int) (*smt.Tree, bool) {
	return nil, false
}

func tryCommitNoSISMiMCGPU(_ []smartvectors.SmartVector) (*smt.Tree, []field.Element, bool) {
	return nil, nil, false
}

func buildSISMiMCTreeGPU(_ []field.Element, _ int) (*smt.Tree, error) {
	return nil, fmt.Errorf("cuda build tag required")
}
