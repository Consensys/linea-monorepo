package keccak

import (
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"

	"github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/generic"
	keccak "github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/keccak/glue"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/keccak/keccakf/iokeccakf"
)

type module struct {
	keccak keccak.KeccakOverBlocks
}

// NewCustomizedKeccak declares the columns and the constraints for proving hash over EXPECTED blocks.
// The correctness of original blocks is checked outside of the module,
// where one also should assert that the expected blocks are the same as correct original blocks.
func NewCustomizedKeccak(comp *wizard.CompiledIOP, maxNbKeccakF int) *module {
	var (
		size = utils.NextPowerOfTwo(generic.KeccakUsecase.NbOfLanesPerBlock() * maxNbKeccakF)

		inp = keccak.KeccakOverBlockInputs{
			LaneInfo: iokeccakf.LaneInfo{
				Lane:                 comp.InsertProof(0, "Lane", size, true),
				IsBeginningOfNewHash: comp.InsertProof(0, "IsFirstLaneOfNewHash", size, true),
				IsLaneActive:         comp.InsertProof(0, "IsLaneActive", size, true),
			},
			KeccakfSize: maxNbKeccakF,
		}
		m = keccak.NewKeccakOverBlocks(comp, inp)
	)

	for i := range m.Outputs.Hash {
		comp.Columns.SetStatus(m.Outputs.Hash[i].GetColID(), column.Proof)
	}

	return &module{
		keccak: *m,
	}
}

// AssignCustomizedKeccak assigns the module.
func (m *module) AssignCustomizedKeccak(run *wizard.ProverRuntime, providers [][]byte) {

	//assign Lane-Info
	keccak.AssignLaneInfo(run, &m.keccak.Inputs.LaneInfo, providers)
	// assign keccak
	m.keccak.Inputs.Provider = providers
	m.keccak.Run(run)

}
