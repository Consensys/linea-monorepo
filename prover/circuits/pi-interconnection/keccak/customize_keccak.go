package keccak

import (
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/largefield/crypto/ringsis"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/largefield/protocol/column"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/largefield/protocol/compiler"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/largefield/protocol/compiler/cleanup"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/largefield/protocol/compiler/logdata"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/largefield/protocol/compiler/mimc"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/largefield/protocol/compiler/selfrecursion"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/largefield/protocol/compiler/vortex"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/largefield/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"

	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/largefield/zkevm/prover/hash/generic"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/largefield/zkevm/prover/hash/keccak"
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
			LaneInfo: keccak.LaneInfo{
				Lanes:                comp.InsertProof(0, "Lane", size),
				IsFirstLaneOfNewHash: comp.InsertProof(0, "IsFirstLaneOfNewHash", size),
				IsLaneActive:         comp.InsertProof(0, "IsLaneActive", size),
			},
			MaxNumKeccakF: maxNbKeccakF,
		}
		m = keccak.NewKeccakOverBlocks(comp, inp)
	)

	comp.Columns.SetStatus(m.HashHi.GetColID(), column.Proof)
	comp.Columns.SetStatus(m.HashLo.GetColID(), column.Proof)

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
	var (
		sisInstance = ringsis.Params{LogTwoBound: 16, LogTwoDegree: 6}

		fullCompilationSuite = compilationSuite{

			compiler.Arcane(
				compiler.WithTargetColSize(1<<18),
				compiler.WithStitcherMinSize(1<<8),
			),
			logdata.Log("after vortex"),
			vortex.Compile(
				2,
				vortex.ForceNumOpenedColumns(256),
				vortex.WithSISParams(&sisInstance),
			),

			selfrecursion.SelfRecurse,
			cleanup.CleanUp,
			mimc.CompileMiMC,
			compiler.Arcane(
				compiler.WithTargetColSize(1<<16),
				compiler.WithStitcherMinSize(1<<8),
			),
			vortex.Compile(
				8,
				vortex.ForceNumOpenedColumns(64),
				vortex.WithSISParams(&sisInstance),
			),

			selfrecursion.SelfRecurse,
			cleanup.CleanUp,
			mimc.CompileMiMC,
			compiler.Arcane(
				compiler.WithTargetColSize(1<<13),
				compiler.WithStitcherMinSize(1<<8),
			),
			vortex.Compile(
				8,
				vortex.ForceNumOpenedColumns(64),
				vortex.WithOptionalSISHashingThreshold(1<<20),
			),
		}
	)

	return fullCompilationSuite

}
