// The keccak package specifies all the mechanism through which the zkevm
// keccaks are proven and extracted from the arithmetization of the zk-EVM.
package keccak

import (
	"github.com/consensys/zkevm-monorepo/prover/crypto/keccak"
	"github.com/consensys/zkevm-monorepo/prover/protocol/ifaces"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
	"github.com/consensys/zkevm-monorepo/prover/symbolic"
	"github.com/consensys/zkevm-monorepo/prover/utils"
	"github.com/consensys/zkevm-monorepo/prover/zkevm/prover/hash/generic"
	"github.com/consensys/zkevm-monorepo/prover/zkevm/prover/hash/importpad"
	gen_acc "github.com/consensys/zkevm-monorepo/prover/zkevm/prover/hash/keccak/acc_module"
	"github.com/consensys/zkevm-monorepo/prover/zkevm/prover/hash/keccak/base_conversion.go"
	"github.com/consensys/zkevm-monorepo/prover/zkevm/prover/hash/keccak/keccakf"
	"github.com/consensys/zkevm-monorepo/prover/zkevm/prover/packing"
	"github.com/consensys/zkevm-monorepo/prover/zkevm/prover/packing/dedicated/spaghettifier"
)

const (
	numLanesPerBlock = 17
)

type KeccakInput struct {
	Settings  *Settings
	Providers []generic.GenDataModule
}

// Module provides the keccakHash component of the zk-EVM
type keccakHash struct {
	Inputs         *KeccakInput
	HashHi, HashLo ifaces.Column
	MaxNumKeccakF  int

	// prover actions for  internal modules
	pa_acc                   wizard.ProverAction
	pa_importPad, pa_packing wizard.ProverAction
	pa_blockBaseConversion   wizard.ProverAction
	pa_hashBaseConversion    wizard.ProverAction
	pa_spaghetti             wizard.ProverAction
	keccakF                  keccakf.Module

	// the result of genericAccumulator
	Provider generic.GenDataModule
}

// Registers the keccak module module within the zkevm arithmetization
func NewKeccak(comp *wizard.CompiledIOP, inp KeccakInput) *keccakHash {
	var (
		maxNumKeccakF        = inp.Settings.MaxNumKeccakf
		size                 = utils.NextPowerOfTwo(maxNumKeccakF * generic.KeccakUsecase.BlockSizeBytes())
		lookupBaseConversion = base_conversion.NewLookupTables(comp)

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

		// apply base conversion over the blocks
		inpBcBlock = base_conversion.BlockBaseConversionInputs{
			Lane:                 packing.Repacked.Lanes,
			IsFirstLaneOfNewHash: packing.Repacked.IsFirstLaneOfNewHash,
			IsLaneActive:         packing.Repacked.IsLaneActive,
			Lookup:               lookupBaseConversion,
		}

		bcForBlock = base_conversion.NewBlockBaseConversion(comp, inpBcBlock)

		// run keccakF
		keccakf = keccakf.NewModule(comp, 0, maxNumKeccakF)

		// bring the hash result to the natural base (uint)
		inpBcHash = base_conversion.HashBaseConversionInput{
			LimbsHiB: append(
				keccakf.IO.HashOutputSlicesBaseB[0][:],
				keccakf.IO.HashOutputSlicesBaseB[1][:]...,
			),

			LimbsLoB: append(
				keccakf.IO.HashOutputSlicesBaseB[2][:],
				keccakf.IO.HashOutputSlicesBaseB[3][:]...,
			),
			MaxNumKeccakF: maxNumKeccakF,
			Lookup:        lookupBaseConversion,
		}

		bcForHash = base_conversion.NewHashBaseConversion(comp, inpBcHash)
	)

	// keccakF does not directly take the blocks, but rather build them via a trace
	// thus, we need to check that the blocks in keccakf matches the one from base conversion.
	// blocks in keccakf are the spaghetti form of LaneX.
	inpSpaghetti := spaghettifier.SpaghettificationInput{
		Name:          "KECCAK",
		ContentMatrix: [][]ifaces.Column{keccakf.Blocks[:]},
		Filter:        isBlock(keccakf.IO.IsBlcok),
		SpaghettiSize: bcForBlock.LaneX.Size(),
	}

	blockSpaghetti := spaghettifier.Spaghettify(comp, inpSpaghetti)
	comp.InsertGlobal(0, "BLOCK_Is_LANEX",
		symbolic.Sub(blockSpaghetti.ContentSpaghetti[0], bcForBlock.LaneX),
	)

	// set the module
	m := &keccakHash{
		Inputs:                 &inp,
		MaxNumKeccakF:          maxNumKeccakF,
		HashHi:                 bcForHash.HashHi,
		HashLo:                 bcForHash.HashLo,
		pa_acc:                 acc,
		pa_importPad:           imported,
		pa_packing:             packing,
		pa_blockBaseConversion: bcForBlock,
		keccakF:                keccakf,
		pa_hashBaseConversion:  bcForHash,
		pa_spaghetti:           blockSpaghetti,
		Provider:               acc.Provider,
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
	// assign blockBaseConversion
	m.pa_blockBaseConversion.Run(run)
	// assign keccakF
	// first, construct the traces for the accumulated Provider
	permTrace := generateTrace(m.Provider.ScanStreams(run))
	m.keccakF.Assign(run, permTrace)
	// assign HashBaseConversion
	m.pa_hashBaseConversion.Run(run)
	//assign blockSpaghetti
	m.pa_spaghetti.Run(run)
}

func isBlock(col ifaces.Column) []ifaces.Column {
	var isBlock []ifaces.Column
	for j := 0; j < numLanesPerBlock; j++ {
		isBlock = append(isBlock, col)
	}
	return isBlock
}

func generateTrace(streams [][]byte) (t keccak.PermTraces) {
	for _, stream := range streams {
		keccak.Hash(stream, &t)
	}
	return t
}
