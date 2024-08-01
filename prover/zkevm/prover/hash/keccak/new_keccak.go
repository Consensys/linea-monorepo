// The keccak package specifies all the mechanism through which the zkevm
// keccaks are proven and extracted from the arithmetization of the zkEVM.
package keccak

import (
	"github.com/consensys/zkevm-monorepo/prover/crypto/keccak"
	"github.com/consensys/zkevm-monorepo/prover/protocol/ifaces"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
	"github.com/consensys/zkevm-monorepo/prover/utils"
	"github.com/consensys/zkevm-monorepo/prover/zkevm/prover/hash/generic"
	"github.com/consensys/zkevm-monorepo/prover/zkevm/prover/hash/importpad"
	gen_acc "github.com/consensys/zkevm-monorepo/prover/zkevm/prover/hash/keccak/acc_module"
	"github.com/consensys/zkevm-monorepo/prover/zkevm/prover/packing"
)

const (
	numLanesPerBlock = 17
)

// KeccakInput stores the inputs for [NewKeccak]
type KeccakInput struct {
	Settings  *Settings
	Providers []generic.GenDataModule
}

// keccakHash stores the hash result and [wizard.ProverAction] of the submodules.
type keccakHash struct {
	Inputs         *KeccakInput
	HashHi, HashLo ifaces.Column
	MaxNumKeccakF  int

	// prover actions for  internal modules
	pa_acc                   wizard.ProverAction
	pa_importPad, pa_packing wizard.ProverAction
	pa_cKeccak               *CustomizedkeccakHash

	// the result of genericAccumulator
	Provider generic.GenDataModule
}

// NewCustomizedKeccak implements the utilities for proving keccak hash
// over the streams which are encoded inside a set of structs [generic.GenDataModule].
// It calls;
// -  accumulate a set of structs [generic.GenDataModule] inside a single struct.
// -  Padding module to insure the correct padding of the streams.
// -  packing module to insure the correct packing of padded-stream into blocks.
// - customizedKeccak to insures the correct hash computation over the given blocks.
func NewKeccak(comp *wizard.CompiledIOP, inp KeccakInput) *keccakHash {
	var (
		maxNumKeccakF = inp.Settings.MaxNumKeccakf
		size          = utils.NextPowerOfTwo(maxNumKeccakF * generic.KeccakUsecase.BlockSizeBytes())

		inpAcc = gen_acc.GenericAccumulatorInputs{
			MaxNumKeccakF: maxNumKeccakF,
			Providers:     inp.Providers,
		}

		// unify the data from different providers in a single provider
		acc = gen_acc.NewGenericAccumulator(comp, inpAcc)

		// apply import and pad
		inpImportPadd = importpad.ImportAndPadInputs{
			Name: "KECCAK",
			Src: generic.GenericByteModule{
				Data: acc.Provider,
			},
			PaddingStrategy: generic.KeccakUsecase,
		}

		imported = importpad.ImportAndPad(comp, inpImportPadd, size)

		// apply packing
		inpPck = packing.PackingInput{
			MaxNumBlocks: maxNumKeccakF,
			PackingParam: generic.KeccakUsecase,
			Imported: packing.Importation{
				Limb:      imported.Limbs,
				NByte:     imported.NBytes,
				IsNewHash: imported.IsNewHash,
				IsActive:  imported.IsActive,
			},
		}

		packing = packing.NewPack(comp, inpPck)

		// apply customized keccak over the blocks
		cKeccakInp = CustomizedKeccakInputs{
			LaneInfo: LaneInfo{
				Lanes:                packing.Repacked.Lanes,
				IsFirstLaneOfNewHash: packing.Repacked.IsFirstLaneOfNewHash,
				IsLaneActive:         packing.Repacked.IsLaneActive,
			},

			MaxNumKeccakF: maxNumKeccakF,
		}
		cKeccak = NewCustomizedKeccak(comp, cKeccakInp)
	)

	// set the module
	m := &keccakHash{
		Inputs:        &inp,
		MaxNumKeccakF: maxNumKeccakF,
		HashHi:        cKeccak.HashHi,
		HashLo:        cKeccak.HashLo,
		pa_acc:        acc,
		pa_importPad:  imported,
		pa_packing:    packing,
		pa_cKeccak:    cKeccak,

		Provider: acc.Provider,
	}

	return m
}

// It implements [wizard.ProverAction] for keccak.
func (m *keccakHash) Run(run *wizard.ProverRuntime) {

	// assign the genericAccumulator module
	m.pa_acc.Run(run)
	// assign ImportAndPad module
	m.pa_importPad.Run(run)
	// assign packing module
	m.pa_packing.Run(run)
	providerBytes := m.Provider.ScanStreams(run)
	m.pa_cKeccak.Inputs.Provider = providerBytes
	m.pa_cKeccak.Run(run)
}

func isBlock(col ifaces.Column) []ifaces.Column {
	var isBlock []ifaces.Column
	for j := 0; j < numLanesPerBlock; j++ {
		isBlock = append(isBlock, col)
	}
	return isBlock
}

// it generates [keccak.PermTraces] from the given stream.
func GenerateTrace(streams [][]byte) (t keccak.PermTraces) {
	for _, stream := range streams {
		keccak.Hash(stream, &t)
	}
	return t
}
