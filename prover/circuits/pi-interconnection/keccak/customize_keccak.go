package keccak

import (
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
	"github.com/consensys/zkevm-monorepo/prover/utils"

	"github.com/consensys/zkevm-monorepo/prover/zkevm/prover/hash/generic"
	wKeccak "github.com/consensys/zkevm-monorepo/prover/zkevm/prover/hash/keccak"
)

type module struct {
	keccak wKeccak.CustomizedkeccakHash
}

// NewCustomizedKeccak declares the columns and the constraints for proving hash over EXPECTED blocks.
// The correctness of original blocks is checked outside of the module,
// where one also should assert that the expected blocks are the same as correct original blocks.
func NewCustomizedKeccak(comp *wizard.CompiledIOP, maxNbKeccakF int) *module {
	var (
		size = utils.NextPowerOfTwo(generic.KeccakUsecase.NbOfLanesPerBlock() * maxNbKeccakF)

		inp = wKeccak.CustomizedKeccakInputs{
			LaneInfo: wKeccak.LaneInfo{
				Lanes:                comp.InsertCommit(0, "Lane", size),
				IsFirstLaneOfNewHash: comp.InsertCommit(0, "IsFirstLaneOfNewHash", size),
				IsLaneActive:         comp.InsertCommit(0, "IsLaneActive", size),
			},
			MaxNumKeccakF: maxNbKeccakF,
		}
		m = wKeccak.NewCustomizedKeccak(comp, inp)
	)
	return &module{
		keccak: *m,
	}
}

// AssignCustomizedKeccak assigns the module.
func (m *module) AssignCustomizedKeccak(run *wizard.ProverRuntime, providers [][]byte) {

	//assign Lane-Info
	wKeccak.AssignLaneInfo(run, &m.keccak.Inputs.LaneInfo, providers)
	// assign keccak
	m.keccak.Inputs.Provider = providers
	m.keccak.Run(run)

}
