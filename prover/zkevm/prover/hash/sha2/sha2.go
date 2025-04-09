// The sha2 package provides all the necessary tools to verify the calls to the
// sha2 precompiles in the Linea's zkevm.
package sha2

import (
	"github.com/consensys/linea-monorepo/prover/protocol/dedicated"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/generic"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/importpad"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/packing"
)

type Settings struct {
	MaxNumSha2F                    int
	NbInstancesPerCircuitSha2Block int
}

// Sha2SingleProviderInput stores the inputs for [newSha2SingleProvider]
type Sha2SingleProviderInput struct {
	Settings
	Provider generic.GenericByteModule
}

// Sha2SingleProvider stores the hash result and [wizard.ProverAction] of the submodules.
type Sha2SingleProvider struct {
	Inputs         *Sha2SingleProviderInput
	HashHi, HashLo ifaces.Column
	// indicates the active part of HashHi/HashLo
	IsActive    ifaces.Column
	MaxNumSha2F int

	// prover actions for  internal modules
	pa_importPad, pa_packing wizard.ProverAction
	pa_cSha2                 *sha2BlockModule
}

// NewSha2ZkEvm constructs the Sha2 module as used in Linea's zkEVM.
func NewSha2ZkEvm(comp *wizard.CompiledIOP, s Settings) *Sha2SingleProvider {

	sha2ProviderInput := Sha2SingleProviderInput{
		Settings: s,
		Provider: generic.GenericByteModule{
			Data: generic.GenDataModule{
				HashNum: comp.Columns.GetHandle("shakiradata.ID"),
				Index:   comp.Columns.GetHandle("shakiradata.INDEX"),
				Limb:    comp.Columns.GetHandle("shakiradata.LIMB"),
				NBytes:  comp.Columns.GetHandle("shakiradata.nBYTES"),
				ToHash:  comp.Columns.GetHandle("shakiradata.IS_SHA2_DATA"),
			},
			Info: generic.GenInfoModule{
				HashNum: comp.Columns.GetHandle("shakiradata.ID"),
				HashLo:  comp.Columns.GetHandle("shakiradata.LIMB"),
				HashHi:  comp.Columns.GetHandle("shakiradata.LIMB"),
				// Before, we usse to pass column.Shift(IsHashHi, -1) but this does
				// not work with the prover distribution as the column is used as
				// a filter for a projection query.
				IsHashHi: comp.Columns.GetHandle("shakiradata.SELECTOR_SHA2_RES_HI"),
			},
		},
	}

	sha2ProviderInput.Provider.Info.IsHashLo = dedicated.ManuallyShift(
		comp,
		sha2ProviderInput.Provider.Info.IsHashHi,
		-1,
	)

	return newSha2SingleProvider(comp, sha2ProviderInput)
}

// newSha2SingleProvider implements the utilities for proving sha2 hash
// over the streams which are encoded inside a set of structs [generic.GenDataModule].
// It calls;
// -  Padding module to insure the correct padding of the streams.
// -  packing module to insure the correct packing of padded-stream into blocks.
// -  sha2Blocks to insures the correct hash computation over the given blocks.
func newSha2SingleProvider(comp *wizard.CompiledIOP, inp Sha2SingleProviderInput) *Sha2SingleProvider {
	var (
		maxNumSha2F = inp.MaxNumSha2F
		size        = utils.NextPowerOfTwo(maxNumSha2F * generic.Sha2Usecase.BlockSizeBytes())

		// apply import and pad
		inpImportPadd = importpad.ImportAndPadInputs{
			Name: "SHA2",
			Src: generic.GenericByteModule{
				Data: inp.Provider.Data,
			},
			PaddingStrategy: generic.Sha2Usecase,
		}

		imported = importpad.ImportAndPad(comp, inpImportPadd, size)

		// apply packing
		inpPck = packing.PackingInput{
			MaxNumBlocks: maxNumSha2F,
			PackingParam: generic.Sha2Usecase,
			Imported: packing.Importation{
				Limb:      imported.Limbs,
				NByte:     imported.NBytes,
				IsNewHash: imported.IsNewHash,
				IsActive:  imported.IsActive,
			},
			Name: "SHA2",
		}

		packing = packing.NewPack(comp, inpPck)

		// this ensures the correctness of the block hashing
		cSha2Inp = &sha2BlocksInputs{
			Name:                 "SHA2_OVER_BLOCK",
			MaxNbBlockPerCirc:    inp.NbInstancesPerCircuitSha2Block,
			MaxNbCircuit:         utils.DivCeil(maxNumSha2F, inp.NbInstancesPerCircuitSha2Block),
			PackedUint32:         packing.Repacked.Lanes,
			Selector:             packing.Repacked.IsLaneActive,
			IsFirstLaneOfNewHash: packing.Repacked.IsFirstLaneOfNewHash,
		}
		cSha2 = newSha2BlockModule(comp, cSha2Inp).WithCircuit(comp)
	)

	comp.InsertProjection("SHA2_RES_HI",
		query.ProjectionInput{ColumnA: []ifaces.Column{cSha2.HashHi},
			ColumnB: []ifaces.Column{inp.Provider.Info.HashHi},
			FilterA: cSha2.IsEffFirstLaneOfNewHash,
			FilterB: inp.Provider.Info.IsHashHi})

	comp.InsertProjection("SHA2_RES_LO",
		query.ProjectionInput{ColumnA: []ifaces.Column{cSha2.HashLo},
			ColumnB: []ifaces.Column{inp.Provider.Info.HashLo},
			FilterA: cSha2.IsEffFirstLaneOfNewHash,
			FilterB: inp.Provider.Info.IsHashLo.(*dedicated.ManuallyShifted).Natural})

	// set the module
	m := &Sha2SingleProvider{
		Inputs:       &inp,
		MaxNumSha2F:  maxNumSha2F,
		HashHi:       cSha2.HashHi,
		HashLo:       cSha2.HashLo,
		IsActive:     cSha2.IsActive,
		pa_importPad: imported,
		pa_packing:   packing,
		pa_cSha2:     cSha2,
	}

	return m
}

// It implements [wizard.ProverAction] for sha2.
func (m *Sha2SingleProvider) Run(run *wizard.ProverRuntime) {

	m.Inputs.Provider.Info.IsHashLo.(*dedicated.ManuallyShifted).Assign(run)

	// assign ImportAndPad module
	m.pa_importPad.Run(run)
	// assign packing module
	m.pa_packing.Run(run)
	m.pa_cSha2.Run(run)
}
