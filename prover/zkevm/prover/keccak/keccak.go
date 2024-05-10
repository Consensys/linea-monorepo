// The keccak package specifies all the mechanism through which the zkevm
// keccaks are proven and extracted from the arithmetization of the zk-EVM.
package keccak

import (
	"runtime"

	"github.com/consensys/zkevm-monorepo/prover/crypto/keccak"
	"github.com/consensys/zkevm-monorepo/prover/protocol/ifaces"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
	"github.com/consensys/zkevm-monorepo/prover/zkevm/prover/keccak/datatransfer/datatransfer"
	"github.com/consensys/zkevm-monorepo/prover/zkevm/prover/keccak/datatransfer/dedicated"
	g "github.com/consensys/zkevm-monorepo/prover/zkevm/prover/keccak/generic"
	"github.com/consensys/zkevm-monorepo/prover/zkevm/prover/keccak/keccakf"
	"github.com/sirupsen/logrus"
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
	dataTransfer datatransfer.DataTransferModule
	// indicates if two modules keccakf and datatransfer are connected
	connected bool
}

// Registers the keccak module module within the zkevm arithmetization
func (m *Module) Define(comp *wizard.CompiledIOP) {
	m.connected = false
	// Pass 0 as the definition round as the present context implies that the
	// keccakf module is here for the zkevm.
	m.Keccakf = keccakf.NewModule(comp, 0, m.Settings.MaxNumKeccakf)
	if m.connected {
		m.dataTransfer.NewDataTransfer(comp, 0, m.Settings.MaxNumKeccakf)
		m.csConnectDataTransferToKeccakF(comp, 0)
	}
}

// Assigns the  keccak module. This module does not require external
// output as everything is readily available from the arithmetization traces.
func (m *Module) AssignKeccak(run *wizard.ProverRuntime) {

	// The list of the providers in use by the zk-EVM module. Today, we do not
	// link it to the actual zk-EVM column (e.g., we don't prove that we use the
	// same values as what is actually in the arithmetization modules). When we
	// turn into that, we will want to have it defined in the module.
	providers := []TraceProvider{
		// g.NewGenericByteModule(run.Spec, g.PHONEY_RLP),
		// NewGenericByteModule(run.Spec, TX_RLP), (disabled)
	}

	// Construct the traces
	permTrace := keccak.PermTraces{}
	genTrace := g.GenTrace{}
	for _, provider := range providers {
		provider.AppendTraces(run, &permTrace, &genTrace)
	}

	// If we have too many permutations, truncate them
	limit := m.Keccakf.MaxNumKeccakf
	if len(permTrace.Blocks) > limit {
		logrus.Warnf(
			"Got too many keccakf's (%d) : truncating to %d",
			len(permTrace.Blocks), limit,
		)
		permTrace.Blocks = permTrace.Blocks[:limit]
		permTrace.KeccakFInps = permTrace.KeccakFInps[:limit]
		permTrace.KeccakFOuts = permTrace.KeccakFOuts[:limit]

		// TBD; truncating genTrace as well or os.Exit(77)
	}

	// And manually assign the module from the content of gt and permTrace.
	if m.connected {
		m.dataTransfer.AssignModule(run, permTrace, genTrace)
	}
	m.Keccakf.Assign(run, permTrace)

	// We empirically found that a forced GC here was improving the runtime.
	runtime.GC()
}

// It connect the data-transfer module to the keccakf module via a projection query over the blocks.
func (mod *Module) csConnectDataTransferToKeccakF(comp *wizard.CompiledIOP, round int) {
	dt := mod.dataTransfer
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
	for j := range dt.HashOutput.HashHiSlices {
		for k := range dt.HashOutput.HashHiSlices[0] {
			comp.InsertInclusion(round, ifaces.QueryIDf("BaseConversion_HashOutput_%v_%v", j, k),
				[]ifaces.Column{dt.LookUps.ColBaseBDirty, dt.LookUps.ColUint4},
				[]ifaces.Column{keccakf.IO.HashOutputSlicesBaseB[j][k], dt.HashOutput.HashLowSlices[j][k]})

			comp.InsertInclusion(round, ifaces.QueryIDf("BaseConversion_HashOutput_%v_%v", j+offSet, k),
				[]ifaces.Column{dt.LookUps.ColBaseBDirty, dt.LookUps.ColUint4},
				[]ifaces.Column{keccakf.IO.HashOutputSlicesBaseB[j+offSet][k], dt.HashOutput.HashHiSlices[j][k]})
		}
	}

}
