package statemanager

import (
	"fmt"

	"github.com/consensys/linea-monorepo/prover/zkevm/arithmetization"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/common"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/column/verifiercol"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	sym "github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/statemanager/lineacodehash"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/statemanager/statesummary"
)

const (
	ACP             = "acp"
	SCP             = "scp"
	ADDR_MULTIPLIER = "340282366920938463463374607431768211456" // 2^{16*8}
)

// romLex returns the columns of the arithmetization.RomLex module of interest
// to justify the consistency between them and the MiMCCodeHash module
func romLex(comp *wizard.CompiledIOP, arith *arithmetization.Arithmetization) *lineacodehash.RomLexInput {
	return &lineacodehash.RomLexInput{
		CFIRomLex: arith.LimbColumnsOfArr2(comp, "romlex", "CODE_FRAGMENT_INDEX"),
		CodeHash:  arith.LimbColumnsOfArr16(comp, "romlex", "CODE_HASH"),
	}
}

// rom returns the columns of the arithmetization corresponding to the Rom module
// that are of interest to justify consistency with the MiMCCodeHash module
func rom(comp *wizard.CompiledIOP, arith *arithmetization.Arithmetization) *lineacodehash.RomInput {
	res := &lineacodehash.RomInput{
		NBytes:   comp.Columns.GetHandle("rom.nBYTES"),
		Counter:  comp.Columns.GetHandle("rom.COUNTER"),
		Acc:      arith.LimbColumnsOfArr8(comp, "rom", "ACC"),
		CFI:      arith.LimbColumnsOfArr2(comp, "rom", "CODE_FRAGMENT_INDEX"),
		CodeSize: arith.LimbColumnsOfArr2(comp, "rom", "CODE_SIZE"),
	}

	return res
}

// acp returns the columns of the arithmetization corresponding to the ACP
// perspective of the Hub that are of interest for checking consistency with
// the stateSummary
func acp(comp *wizard.CompiledIOP, arith *arithmetization.Arithmetization) statesummary.HubColumnSet {
	size := comp.Columns.GetHandle("hub.acp_ADDRESS_HI").Size()

	// the prover-side state manager uses a single field element for 20-bytes addresses
	// and we need to create this column ourselves
	if !comp.Columns.Exists("HUB_acp_PROVER_SIDE_ADDRESS_IDENTIFIER") {
		combinedAddr := comp.InsertCommit(0,
			"HUB_acp_PROVER_SIDE_ADDRESS_IDENTIFIER",
			size,
		)

		// constrain the processed HUB addresses
		addrHI := arith.LimbColumnsOfArr2(comp, "hub", "acp_ADDRESS_HI")
		addrLO := arith.LimbColumnsOfArr8(comp, "hub", "acp_ADDRESS_LO")
		comp.InsertGlobal(
			0,
			ifaces.QueryIDf("STATE_MANAGER_ACP_HUB_PROCESSED_ADDRESSES_GLOBAL_CONSTRAINT"),
			sym.Sub(
				combinedAddr,
				sym.Mul(
					addrHI,
					field.NewFromString(ADDR_MULTIPLIER),
				),
				addrLO,
			),
		)
	}

	constantZero := verifiercol.NewConstantCol(field.Zero(), size, "hub.acp_connection")
	constantZero8 := [8]ifaces.Column{constantZero, constantZero, constantZero, constantZero, constantZero, constantZero, constantZero, constantZero}

	res := statesummary.HubColumnSet{
		Address:             common.GetMultiHandleEthAddress(comp, "HUB_acp_PROVER_SIDE_ADDRESS_IDENTIFIER"),
		AddressHI:           arith.LimbColumnsOfArr2(comp, "hub", "acp_ADDRESS_HI"),
		AddressLO:           arith.LimbColumnsOfArr8(comp, "hub", "acp_ADDRESS_LO"),
		Nonce:               arith.LimbColumnsOfArr4(comp, "hub", "acp_NONCE"),
		NonceNew:            arith.LimbColumnsOfArr4(comp, "hub", "acp_NONCE_NEW"),
		CodeHashHI:          arith.LimbColumnsOfArr8(comp, "hub", "acp_CODE_HASH_HI"),
		CodeHashLO:          arith.LimbColumnsOfArr8(comp, "hub", "acp_CODE_HASH_LO"),
		CodeHashHINew:       arith.LimbColumnsOfArr8(comp, "hub", "acp_CODE_HASH_HI_NEW"),
		CodeHashLONew:       arith.LimbColumnsOfArr8(comp, "hub", "acp_CODE_HASH_LO_NEW"),
		CodeSizeOld:         arith.LimbColumnsOfArr4(comp, "hub", "acp_CODE_SIZE"),
		CodeSizeNew:         arith.LimbColumnsOfArr4(comp, "hub", "acp_CODE_SIZE_NEW"),
		BalanceOld:          arith.LimbColumnsOfArr8(comp, "hub", "acp_BALANCE"),
		BalanceNew:          arith.LimbColumnsOfArr8(comp, "hub", "acp_BALANCE_NEW"),
		KeyHI:               constantZero8,
		KeyLO:               constantZero8,
		ValueHICurr:         constantZero8,
		ValueLOCurr:         constantZero8,
		ValueHINext:         constantZero8,
		ValueLONext:         constantZero8,
		DeploymentNumber:    arith.LimbColumnsOfArr2(comp, "hub", "acp_DEPLOYMENT_NUMBER"),
		DeploymentNumberInf: arith.LimbColumnsOfArr2(comp, "hub", "acp_DEPLOYMENT_NUMBER"),
		BlockNumber:         arith.LimbColumnsOfArr4(comp, "hub", "acp_BLK_NUMBER"),
		Exists:              arith.ColumnOf(comp, "hub", "acp_EXISTS"),
		ExistsNew:           arith.ColumnOf(comp, "hub", "acp_EXISTS_NEW"),
		PeekAtAccount:       arith.ColumnOf(comp, "hub", "acp_PEEK_AT_ACCOUNT"),
		PeekAtStorage:       constantZero,
		FirstAOC:            arith.ColumnOf(comp, "hub", "acp_FIRST_IN_CNF"),
		LastAOC:             arith.ColumnOf(comp, "hub", "acp_FINAL_IN_CNF"),
		FirstKOC:            constantZero,
		LastKOC:             constantZero,
		FirstAOCBlock:       arith.ColumnOf(comp, "hub", "acp_FIRST_IN_BLK"),
		LastAOCBlock:        arith.ColumnOf(comp, "hub", "acp_FINAL_IN_BLK"),
		FirstKOCBlock:       constantZero,
		LastKOCBlock:        constantZero,
		MinDeplBlock:        arith.LimbColumnsOfArr2(comp, "hub", "acp_DEPLOYMENT_NUMBER_FIRST_IN_BLOCK"),
		MaxDeplBlock:        arith.LimbColumnsOfArr2(comp, "hub", "acp_DEPLOYMENT_NUMBER_FINAL_IN_BLOCK"),
		ExistsFirstInBlock:  constantZero,
		ExistsFinalInBlock:  constantZero,
	}

	return res
}

// scp returns the columns of the arithmetization correspoanding to the SCP
// perspective of the Hub that are of interest for checking consistency with
// the stateSummary
func scp(comp *wizard.CompiledIOP, arith *arithmetization.Arithmetization) statesummary.HubColumnSet {
	size := comp.Columns.GetHandle("hub.scp_ADDRESS_HI").Size()

	// the prover-side state manager uses a single field element for 20-bytes addresses
	// and we need to create this column ourselves
	if !comp.Columns.Exists("HUB_scp_PROVER_SIDE_ADDRESS_IDENTIFIER") {
		combinedAddr := comp.InsertCommit(0,
			"HUB_scp_PROVER_SIDE_ADDRESS_IDENTIFIER",
			size,
		)

		// constrain the processed HUB addresses
		addrHI := arith.LimbColumnsOfArr2(comp, "hub", "scp_ADDRESS_HI")
		addrLO := arith.LimbColumnsOfArr8(comp, "hub", "scp_ADDRESS_LO")
		comp.InsertGlobal(
			0,
			ifaces.QueryIDf("STATE_MANAGER_SCP_HUB_PROCESSED_ADDRESSES_GLOBAL_CONSTRAINT"),
			sym.Sub(
				combinedAddr,
				sym.Mul(
					addrHI,
					field.NewFromString(ADDR_MULTIPLIER),
				),
				addrLO,
			),
		)
	}

	constantZero := verifiercol.NewConstantCol(field.Zero(), size, "hub.scp_connection")

	constantZero4 := [4]ifaces.Column{
		constantZero, constantZero,
		constantZero, constantZero,
	}

	constantZero8 := [8]ifaces.Column{
		constantZero, constantZero,
		constantZero, constantZero,
		constantZero, constantZero,
		constantZero, constantZero,
	}

	res := statesummary.HubColumnSet{
		Address:             common.GetMultiHandleEthAddress(comp, "HUB_scp_PROVER_SIDE_ADDRESS_IDENTIFIER"),
		AddressHI:           arith.LimbColumnsOfArr2(comp, "hub", "scp_ADDRESS_HI"),
		AddressLO:           arith.LimbColumnsOfArr8(comp, "hub", "scp_ADDRESS_LO"),
		Nonce:               constantZero4,
		NonceNew:            constantZero4,
		CodeHashHI:          constantZero8,
		CodeHashLO:          constantZero8,
		CodeHashHINew:       constantZero8,
		CodeHashLONew:       constantZero8,
		CodeSizeOld:         constantZero4,
		CodeSizeNew:         constantZero4,
		BalanceOld:          constantZero8,
		BalanceNew:          constantZero8,
		KeyHI:               arith.LimbColumnsOfArr8(comp, "hub", "scp_STORAGE_KEY_HI"),
		KeyLO:               arith.LimbColumnsOfArr8(comp, "hub", "scp_STORAGE_KEY_LO"),
		ValueHICurr:         arith.LimbColumnsOfArr8(comp, "hub", "scp_VALUE_CURR_HI"),
		ValueLOCurr:         arith.LimbColumnsOfArr8(comp, "hub", "scp_VALUE_CURR_LO"),
		ValueHINext:         arith.LimbColumnsOfArr8(comp, "hub", "scp_VALUE_NEXT_HI"),
		ValueLONext:         arith.LimbColumnsOfArr8(comp, "hub", "scp_VALUE_NEXT_LO"),
		DeploymentNumber:    arith.LimbColumnsOfArr2(comp, "hub", "scp_DEPLOYMENT_NUMBER"),
		DeploymentNumberInf: arith.LimbColumnsOfArr2(comp, "hub", "scp_DEPLOYMENT_NUMBER"),
		BlockNumber:         arith.LimbColumnsOfArr4(comp, "hub", "scp_BLK_NUMBER"),
		Exists:              constantZero,
		ExistsNew:           constantZero,
		PeekAtAccount:       constantZero,
		PeekAtStorage:       arith.ColumnOf(comp, "hub", "scp_PEEK_AT_STORAGE"),
		FirstAOC:            constantZero,
		LastAOC:             constantZero,
		FirstKOC:            arith.ColumnOf(comp, "hub", "scp_FIRST_IN_CNF"),
		LastKOC:             arith.ColumnOf(comp, "hub", "scp_FINAL_IN_CNF"),
		FirstAOCBlock:       constantZero,
		LastAOCBlock:        constantZero,
		FirstKOCBlock:       arith.ColumnOf(comp, "hub", "scp_FIRST_IN_BLK"),
		LastKOCBlock:        arith.ColumnOf(comp, "hub", "scp_FINAL_IN_BLK"),
		MinDeplBlock:        arith.LimbColumnsOfArr2(comp, "hub", "scp_DEPLOYMENT_NUMBER_FIRST_IN_BLOCK"),
		MaxDeplBlock:        arith.LimbColumnsOfArr2(comp, "hub", "scp_DEPLOYMENT_NUMBER_FINAL_IN_BLOCK"),
		ExistsFirstInBlock:  arith.ColumnOf(comp, "hub", "scp_EXISTS_FIRST_IN_BLOCK"),
		ExistsFinalInBlock:  arith.ColumnOf(comp, "hub", "scp_EXISTS_FINAL_IN_BLOCK"),
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
			wizard.DisableAssignmentSizeReduction,
		)
	}
	// assign the addresses column in each of the submodules
	assignHubAddressesSubdomain(ACP)
	assignHubAddressesSubdomain(SCP)
}
