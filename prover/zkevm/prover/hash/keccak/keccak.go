// The keccak package specifies all the mechanism through which the zkevm
// keccaks are proven and extracted from the arithmetization of the zk-EVM.
package keccak

import (
	"runtime"

	"github.com/consensys/zkevm-monorepo/prover/crypto/keccak"
	"github.com/consensys/zkevm-monorepo/prover/protocol/ifaces"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
	"github.com/consensys/zkevm-monorepo/prover/utils"
	"github.com/consensys/zkevm-monorepo/prover/zkevm/prover/hash/datatransfer/acc_module"
	"github.com/consensys/zkevm-monorepo/prover/zkevm/prover/hash/datatransfer/datatransfer"
	"github.com/consensys/zkevm-monorepo/prover/zkevm/prover/hash/datatransfer/dedicated"
	g "github.com/consensys/zkevm-monorepo/prover/zkevm/prover/hash/generic"
	"github.com/consensys/zkevm-monorepo/prover/zkevm/prover/hash/keccak/keccakf"
)

// A trace provider represents a module capable sending data to the keccak
// module. It does so by appending keccak block traces to a given list. The
// trace provider may access the current Prover runtime to extract information.
type TraceProvider interface {
	AppendTraces(run *wizard.ProverRuntime, traces *keccak.PermTraces, genTrace *g.GenTrace)
}

// Module provides the Keccak component of the zk-EVM
type Module struct {
	Settings     *Settings
	Keccakf      keccakf.Module
	DataTransfer datatransfer.Module
	dataTrace    acc_module.DataModule
	infoTrace    acc_module.InfoModule
	providers    []g.GenericByteModule
	// for inputs that are slices of bytes
	SliceProviders [][]byte
	MaxNumKeccakF  int
}

// Registers the keccak module module within the zkevm arithmetization
func (m *Module) Define(comp *wizard.CompiledIOP, providers []g.GenericByteModule, nbKeccakF int) {
	m.providers = providers
	m.MaxNumKeccakF = nbKeccakF

	// Pass 0 as the definition round as the present context implies that the
	// keccakf module is here for the zkevm.

	// unify the data from different providers in a single provider, in a provable way
	m.dataTrace.NewDataModule(comp, 0, m.MaxNumKeccakF, m.providers)

	// assign the provider to DataTransfer module
	m.DataTransfer.Provider = m.dataTrace.Provider

	// check the correct dataSerialization of limbs to blocks via dataTransfer module
	m.DataTransfer.NewDataTransfer(comp, 0, m.MaxNumKeccakF, 0)

	// run keccakF over the blocks, in a provable way
	m.Keccakf = keccakf.NewModule(comp, 0, m.MaxNumKeccakF)

	// assign the blocks from DataTransfer to keccakF,
	// also take the output from keccakF and give it back to DataTransfer
	m.CsConnectDataTransferToKeccakF(comp, 0)

	// project back the hash results to the providers
	m.infoTrace.NewInfoModule(comp, 0, m.MaxNumKeccakF, m.providers, m.DataTransfer.HashOutput, m.dataTrace)

}

// Assigns the  keccak module. This module does not require external
// output as everything is readily available from the arithmetization traces.
func (m *Module) AssignKeccak(run *wizard.ProverRuntime) {

	// assign the aggregated Provider
	m.dataTrace.AssignDataModule(run, m.providers)

	// Construct the traces for the aggregated Provider
	permTrace := keccak.PermTraces{}
	genTrace := g.GenTrace{}
	m.DataTransfer.Provider.AppendTraces(run, &genTrace, &permTrace)

	// If we have too many permutations, truncate them
	limit := m.Keccakf.MaxNumKeccakf
	if len(permTrace.Blocks) > limit {
		utils.Panic("got too many keccakf. Limit is  %v, but received %v", limit, len(permTrace.Blocks))
	}

	// And manually assign the module from the content of genTrace and permTrace.
	m.DataTransfer.AssignModule(run, permTrace, genTrace)
	m.Keccakf.Assign(run, permTrace)
	m.infoTrace.AssignInfoModule(run, m.providers)

	// We empirically found that a forced GC here was improving the runtime.
	runtime.GC()
}

// It connect the data-transfer module to the keccakf module via a projection query over the blocks.
func (mod *Module) CsConnectDataTransferToKeccakF(comp *wizard.CompiledIOP, round int) {
	dt := mod.DataTransfer
	keccakf := mod.Keccakf

	// constraints over Data-module (inputs)
	var filterIsBlock []ifaces.Column
	for j := 0; j < 17; j++ {
		filterIsBlock = append(filterIsBlock, keccakf.IO.IsBlcok)
	}

	spaghettiSize := dt.BaseConversion.LaneX.Size()
	dedicated.InsertIsSpaghetti(comp, round, ifaces.QueryIDf("ExportBlocks"),
		[][]ifaces.Column{keccakf.Blocks[:]}, filterIsBlock, []ifaces.Column{dt.BaseConversion.LaneX}, spaghettiSize)

	// constraints over Info-module (outputs)
	//	(i.e., from hashOutputSlicesBaseB to hashHiSlices/hashLowSlices)
	offSet := 2
	for j := range dt.HashOutput.HashLoSlices {
		for k := range dt.HashOutput.HashLoSlices[0] {
			comp.InsertInclusion(round, ifaces.QueryIDf("BaseConversion_HashOutput_%v_%v", j, k),
				[]ifaces.Column{dt.LookUps.ColBaseBDirty, dt.LookUps.ColUint4},
				[]ifaces.Column{keccakf.IO.HashOutputSlicesBaseB[j][k], dt.HashOutput.HashHiSlices[j][k]})

			comp.InsertInclusion(round, ifaces.QueryIDf("BaseConversion_HashOutput_%v_%v", j+offSet, k),
				[]ifaces.Column{dt.LookUps.ColBaseBDirty, dt.LookUps.ColUint4},
				[]ifaces.Column{keccakf.IO.HashOutputSlicesBaseB[j+offSet][k], dt.HashOutput.HashLoSlices[j][k]})
		}
	}

}
