package statemanager

import (
	"fmt"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/column/verifiercol"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/statemanager/mimccodehash"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/statemanager/statesummary"
)

const (
	ACP = "acp"
	SCP = "scp"
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
	size := comp.Columns.GetHandle("hub.acp_ADDRESS_HI").Size()

	// the prover-side state manager uses a single field element for 20-bytes addresses
	// and we need to create this column ourselves
	if !comp.Columns.Exists("HUB_acp_PROVER_SIDE_ADDRESS_IDENTIFIER") {
		comp.InsertCommit(0,
			"HUB_acp_PROVER_SIDE_ADDRESS_IDENTIFIER",
			size,
		)
	}

	res := statesummary.HubColumnSet{
		Address:             comp.Columns.GetHandle("HUB_acp_PROVER_SIDE_ADDRESS_IDENTIFIER"),
		AddressHI:           comp.Columns.GetHandle("hub.acp_ADDRESS_HI"),
		AddressLO:           comp.Columns.GetHandle("hub.acp_ADDRESS_LO"),
		Nonce:               comp.Columns.GetHandle("hub.acp_NONCE"),
		NonceNew:            comp.Columns.GetHandle("hub.acp_NONCE_NEW"),
		CodeHashHI:          comp.Columns.GetHandle("hub.acp_CODE_HASH_HI"),
		CodeHashLO:          comp.Columns.GetHandle("hub.acp_CODE_HASH_LO"),
		CodeHashHINew:       comp.Columns.GetHandle("hub.acp_CODE_HASH_HI_NEW"),
		CodeHashLONew:       comp.Columns.GetHandle("hub.acp_CODE_HASH_LO_NEW"),
		CodeSizeOld:         comp.Columns.GetHandle("hub.acp_CODE_SIZE"),
		CodeSizeNew:         comp.Columns.GetHandle("hub.acp_CODE_SIZE_NEW"),
		BalanceOld:          comp.Columns.GetHandle("hub.acp_BALANCE"),
		BalanceNew:          comp.Columns.GetHandle("hub.acp_BALANCE_NEW"),
		KeyHI:               comp.Columns.GetHandle("HUB_acp_PROVER_SIDE_ADDRESS_IDENTIFIER"),
		KeyLO:               comp.Columns.GetHandle("HUB_acp_PROVER_SIDE_ADDRESS_IDENTIFIER"),
		ValueHICurr:         comp.Columns.GetHandle("HUB_acp_PROVER_SIDE_ADDRESS_IDENTIFIER"),
		ValueLOCurr:         comp.Columns.GetHandle("HUB_acp_PROVER_SIDE_ADDRESS_IDENTIFIER"),
		ValueHINext:         comp.Columns.GetHandle("HUB_acp_PROVER_SIDE_ADDRESS_IDENTIFIER"),
		ValueLONext:         comp.Columns.GetHandle("HUB_acp_PROVER_SIDE_ADDRESS_IDENTIFIER"),
		DeploymentNumber:    comp.Columns.GetHandle("hub.acp_DEPLOYMENT_NUMBER"),
		DeploymentNumberInf: comp.Columns.GetHandle("hub.acp_DEPLOYMENT_NUMBER"),
		BlockNumber:         comp.Columns.GetHandle("hub.acp_REL_BLK_NUM"),
		Exists:              comp.Columns.GetHandle("hub.acp_EXISTS"),
		ExistsNew:           comp.Columns.GetHandle("hub.acp_EXISTS_NEW"),
		PeekAtAccount:       comp.Columns.GetHandle("hub.acp_PEEK_AT_ACCOUNT"),
		PeekAtStorage:       comp.Columns.GetHandle("HUB_acp_PROVER_SIDE_ADDRESS_IDENTIFIER"),
		FirstAOC:            comp.Columns.GetHandle("hub.acp_FIRST_IN_CNF"),
		LastAOC:             comp.Columns.GetHandle("hub.acp_FINAL_IN_CNF"),
		FirstKOC:            comp.Columns.GetHandle("HUB_acp_PROVER_SIDE_ADDRESS_IDENTIFIER"),
		LastKOC:             comp.Columns.GetHandle("HUB_acp_PROVER_SIDE_ADDRESS_IDENTIFIER"),
		FirstAOCBlock:       comp.Columns.GetHandle("hub.acp_FIRST_IN_BLK"),
		LastAOCBlock:        comp.Columns.GetHandle("hub.acp_FINAL_IN_BLK"),
		FirstKOCBlock:       comp.Columns.GetHandle("HUB_acp_PROVER_SIDE_ADDRESS_IDENTIFIER"),
		LastKOCBlock:        comp.Columns.GetHandle("HUB_acp_PROVER_SIDE_ADDRESS_IDENTIFIER"),
		MinDeplBlock:        comp.Columns.GetHandle("hub.acp_DEPLOYMENT_NUMBER_FIRST_IN_BLOCK"),
		MaxDeplBlock:        comp.Columns.GetHandle("hub.acp_DEPLOYMENT_NUMBER_FINAL_IN_BLOCK"),
	}
	return res
}

// scp returns the columns of the arithmetization correspoanding to the SCP
// perspective of the Hub that are of interest for checking consistency with
// the stateSummary
func scp(comp *wizard.CompiledIOP) statesummary.HubColumnSet {
	size := comp.Columns.GetHandle("hub.scp_ADDRESS_HI").Size()

	// the prover-side state manager uses a single field element for 20-bytes addresses
	// and we need to create this column ourselves
	if !comp.Columns.Exists("HUB_scp_PROVER_SIDE_ADDRESS_IDENTIFIER") {
		comp.InsertCommit(0,
			"HUB_scp_PROVER_SIDE_ADDRESS_IDENTIFIER",
			size,
		)
	}

	res := statesummary.HubColumnSet{
		Address:             comp.Columns.GetHandle("HUB_scp_PROVER_SIDE_ADDRESS_IDENTIFIER"),
		AddressHI:           comp.Columns.GetHandle("hub.scp_ADDRESS_HI"),
		AddressLO:           comp.Columns.GetHandle("hub.scp_ADDRESS_LO"),
		Nonce:               verifiercol.NewConstantCol(field.Zero(), size),
		NonceNew:            verifiercol.NewConstantCol(field.Zero(), size),
		CodeHashHI:          verifiercol.NewConstantCol(field.Zero(), size),
		CodeHashLO:          verifiercol.NewConstantCol(field.Zero(), size),
		CodeHashHINew:       verifiercol.NewConstantCol(field.Zero(), size),
		CodeHashLONew:       verifiercol.NewConstantCol(field.Zero(), size),
		CodeSizeOld:         verifiercol.NewConstantCol(field.Zero(), size),
		CodeSizeNew:         verifiercol.NewConstantCol(field.Zero(), size),
		BalanceOld:          verifiercol.NewConstantCol(field.Zero(), size),
		BalanceNew:          verifiercol.NewConstantCol(field.Zero(), size),
		KeyHI:               comp.Columns.GetHandle("hub.scp_STORAGE_KEY_HI"),
		KeyLO:               comp.Columns.GetHandle("hub.scp_STORAGE_KEY_LO"),
		ValueHICurr:         comp.Columns.GetHandle("hub.scp_VALUE_CURR_HI"),
		ValueLOCurr:         comp.Columns.GetHandle("hub.scp_VALUE_CURR_LO"),
		ValueHINext:         comp.Columns.GetHandle("hub.scp_VALUE_NEXT_HI"),
		ValueLONext:         comp.Columns.GetHandle("hub.scp_VALUE_NEXT_LO"),
		DeploymentNumber:    comp.Columns.GetHandle("hub.scp_DEPLOYMENT_NUMBER"),
		DeploymentNumberInf: comp.Columns.GetHandle("hub.scp_DEPLOYMENT_NUMBER"),
		BlockNumber:         comp.Columns.GetHandle("hub.scp_REL_BLK_NUM"),
		Exists:              verifiercol.NewConstantCol(field.Zero(), size),
		ExistsNew:           verifiercol.NewConstantCol(field.Zero(), size),
		PeekAtAccount:       verifiercol.NewConstantCol(field.Zero(), size),
		PeekAtStorage:       comp.Columns.GetHandle("hub.scp_PEEK_AT_STORAGE"),
		FirstAOC:            verifiercol.NewConstantCol(field.Zero(), size),
		LastAOC:             verifiercol.NewConstantCol(field.Zero(), size),
		FirstKOC:            comp.Columns.GetHandle("hub.scp_FIRST_IN_CNF"),
		LastKOC:             comp.Columns.GetHandle("hub.scp_FINAL_IN_CNF"),
		FirstAOCBlock:       verifiercol.NewConstantCol(field.Zero(), size),
		LastAOCBlock:        verifiercol.NewConstantCol(field.Zero(), size),
		FirstKOCBlock:       comp.Columns.GetHandle("hub.scp_FIRST_IN_BLK"),
		LastKOCBlock:        comp.Columns.GetHandle("hub.scp_FINAL_IN_BLK"),
		MinDeplBlock:        comp.Columns.GetHandle("hub.scp_DEPLOYMENT_NUMBER_FIRST_IN_BLOCK"),
		MaxDeplBlock:        comp.Columns.GetHandle("hub.scp_DEPLOYMENT_NUMBER_FINAL_IN_BLOCK"),
	}
	return res
}

/*
assignHubAddresses is a function that combines addressHI and addressLO from
the arithmetization columns into a single column.
*/
func assignHubAddresses(run *wizard.ProverRuntime) {
	assignHubAddressesSubdomain := func(domainName string) {
		addressHI := run.GetColumn(ifaces.ColID(fmt.Sprintf("hub.%s_ADDRESS_HI", domainName)))
		addressLO := run.GetColumn(ifaces.ColID(fmt.Sprintf("hub.%s_ADDRESS_LO", domainName)))

		size := addressHI.Len()
		newVect := make([]field.Element, size)
		for i := range newVect {
			elemHi := addressHI.Get(i)
			bytesHi := elemHi.Bytes()

			elemLo := addressLO.Get(i)
			bytesLo := elemLo.Bytes()
			newBytes := make([]byte, field.Bytes)
			// set the high part
			for j := 0; j < 4; j++ {
				newBytes[12+j] = bytesHi[32-(4-j)]
			}
			// set the low part
			for j := 4; j < 20; j++ {
				newBytes[12+j] = bytesLo[16+(j-4)]
			}
			newVect[i].SetBytes(newBytes)
		}
		run.AssignColumn(
			ifaces.ColID(fmt.Sprintf("HUB_%s_PROVER_SIDE_ADDRESS_IDENTIFIER", domainName)),
			smartvectors.NewRegular(newVect),
		)

		run.AssignColumn(
			ifaces.ColID(fmt.Sprintf("HUB_%s_PROVER_SIDE_ZERO_COLUMN", domainName)),
			smartvectors.NewConstant(field.Zero(), size),
		)
	}
	// assign the addresses column in each of the submodules
	assignHubAddressesSubdomain(ACP)
	assignHubAddressesSubdomain(SCP)
}
