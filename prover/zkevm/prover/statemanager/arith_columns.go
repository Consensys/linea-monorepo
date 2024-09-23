package statemanager

import (
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/statemanager/mimccodehash"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/statemanager/statesummary"
)

// romLex returns the columns of the arithmetization.RomLex module of interest
// to justify the consistency between them and the MiMCCodeHash module
func romLex(comp *wizard.CompiledIOP) *mimccodehash.RomLexInput {
	return &mimccodehash.RomLexInput{
		CFIRomLex:  comp.Columns.GetHandle("romlex.CODE_FRAGMENT_INDEX"),
		CodeHashHi: comp.Columns.GetHandle("romlex.CODE_HASH_HI"),
		CodeHashLo: comp.Columns.GetHandle("romlex.CODE_HASH_LO"),
	}
}

// rom returns the columns of the arithmetization corresponding to the Rom module
// that are of interest to justify consistency with the MiMCCodeHash module
func rom(comp *wizard.CompiledIOP) *mimccodehash.RomInput {
	res := &mimccodehash.RomInput{
		CFI:      comp.Columns.GetHandle("rom.CODE_FRAGMENT_INDEX"),
		Acc:      comp.Columns.GetHandle("rom.ACC"),
		NBytes:   comp.Columns.GetHandle("rom.nBYTES"),
		Counter:  comp.Columns.GetHandle("rom.COUNTER"),
		CodeSize: comp.Columns.GetHandle("rom.CODE_SIZE"),
	}
	return res
}

// acp returns the columns of the arithmetization corresponding to the ACP
// perspective of the Hub that are of interest for checking consistency with
// the stateSummary
func acp(comp *wizard.CompiledIOP) statesummary.HubColumnSet {
	panic("not available yet")
}

// scp returns the columns of the arithmetization corresponding to the SCP
// perspective of the Hub that are of interest for checking consistency with
// the stateSummary
func scp(comp *wizard.CompiledIOP) statesummary.HubColumnSet {
	panic("not available yet")
}
