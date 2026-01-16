package keccak

import (
	"github.com/consensys/linea-monorepo/prover/crypto/ringsis"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/cleanup"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/logdata"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/selfrecursion"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/vortex"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"

	mimcComp "github.com/consensys/linea-monorepo/prover/protocol/compiler/poseidon2"

	"github.com/consensys/linea-monorepo/prover/utils"

	"github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/generic"
	keccak "github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/keccak/glue"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/keccak/keccakf/iokeccakf"
)

type module struct {
	keccak keccak.KeccakOverBlocks
}

// NewCustomizedKeccak declares the columns and the constraints for proving hash over EXPECTED blocks.
// The correctness of original blocks is checked outside the module,
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

// type alias to denote a wizard-compilation suite. This is used when calling
// compile and provides internal parameters for the wizard package.
type compilationSuite = []func(*wizard.CompiledIOP)

func WizardCompilationParameters() []func(iop *wizard.CompiledIOP) {
	utils.Panic("use the right self-recursion parameters")

	var (
		sisInstance = ringsis.Params{LogTwoBound: 16, LogTwoDegree: 6}

		fullCompilationSuite = compilationSuite{

			compiler.Arcane(
				compiler.WithTargetColSize(1<<18),
				compiler.WithStitcherMinSize(1<<8),
			),
			logdata.Log("after vortex"),
			vortex.Compile(
				2, true,
				vortex.ForceNumOpenedColumns(256),
				vortex.WithSISParams(&sisInstance),
			),

			selfrecursion.SelfRecurse,
			cleanup.CleanUp,
			mimcComp.CompilePoseidon2,
			compiler.Arcane(
				compiler.WithTargetColSize(1<<16),
				compiler.WithStitcherMinSize(1<<8),
			),
			vortex.Compile(
				8,
				true,
				vortex.ForceNumOpenedColumns(64),
				vortex.WithSISParams(&sisInstance),
			),

			selfrecursion.SelfRecurse,
			cleanup.CleanUp,
			mimcComp.CompilePoseidon2,
			compiler.Arcane(
				compiler.WithTargetColSize(1<<13),
				compiler.WithStitcherMinSize(1<<8),
			),
			vortex.Compile(
				8, true,
				vortex.ForceNumOpenedColumns(64),
				vortex.WithOptionalSISHashingThreshold(1<<20),
			),
		}
	)

	return fullCompilationSuite

}
