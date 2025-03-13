package statesummary

import (
	"sync"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/dedicated"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	sym "github.com/consensys/linea-monorepo/prover/symbolic"
)

// arithmetizationLink collects columns from the hub that are of interest for
// checking consistency between the StateSummary and the rest of the
// arithmetization.
type arithmetizationLink struct {
	Acp, Scp    HubColumnSet
	scpSelector scpSelector
}

// ConnectToHub generates all the constraints attesting that the state-summary
// and the Hub relates to the same state operations.
func (ss *Module) ConnectToHub(comp *wizard.CompiledIOP, acp, scp HubColumnSet) {

	al := &arithmetizationLink{
		Acp:         acp,
		Scp:         scp,
		scpSelector: newScpSelector(comp, scp),
	}

	storageIntegrationDefineInitial(comp, *ss, scp, al.scpSelector)
	storageIntegrationDefineFinal(comp, *ss, scp, al.scpSelector)
	accountIntegrationDefineInitial(comp, *ss, acp)
	accountIntegrationDefineFinal(comp, *ss, acp)

	ss.arithmetizationLink = al
}

func (ss *Module) assignArithmetizationLink(run *wizard.ProverRuntime) {

	storageIntegrationAssignInitial(run, *ss, ss.arithmetizationLink.Scp)
	storageIntegrationAssignFinal(run, *ss, ss.arithmetizationLink.Scp)
	accountIntegrationAssignInitial(run, *ss, ss.arithmetizationLink.Acp)
	accountIntegrationAssignFinal(run, *ss, ss.arithmetizationLink.Acp)

	// @alex: this should be commonized utility or should be simplified to not
	// use a closure because the closure is used only once.
	runConcurrent := func(pas []wizard.ProverAction) {
		wg := &sync.WaitGroup{}
		for _, pa := range pas {
			wg.Add(1)
			go func(pa wizard.ProverAction) {
				pa.Run(run)
				wg.Done()
			}(pa)
		}

		wg.Wait()
	}

	runConcurrent([]wizard.ProverAction{
		ss.arithmetizationLink.scpSelector.ComputeSelectorMinDeplBlock,
		ss.arithmetizationLink.scpSelector.ComputeSelectorMaxDeplBlock,
	})

}

/*
HubColumnSet is a struct that corresponds to the HUB columns present in the ACP (account consistency permutation)
and the SCP (storage consistency permutation)
*/
type HubColumnSet struct {
	// helper column
	Address ifaces.Column
	// account data
	AddressHI, AddressLO                                 ifaces.Column
	Nonce, NonceNew                                      ifaces.Column
	CodeHashHI, CodeHashLO, CodeHashHINew, CodeHashLONew ifaces.Column
	CodeSizeOld, CodeSizeNew                             ifaces.Column
	BalanceOld, BalanceNew                               ifaces.Column
	// storage data
	KeyHI, KeyLO                                       ifaces.Column
	ValueHICurr, ValueLOCurr, ValueHINext, ValueLONext ifaces.Column
	// helper numbers
	DeploymentNumber, DeploymentNumberInf ifaces.Column
	BlockNumber                           ifaces.Column
	// helper columns
	Exists, ExistsNew ifaces.Column
	PeekAtAccount     ifaces.Column
	PeekAtStorage     ifaces.Column
	// first and last marker columns
	FirstAOC, LastAOC ifaces.Column
	FirstKOC, LastKOC ifaces.Column
	// first and last block marker columns
	FirstAOCBlock, LastAOCBlock ifaces.Column
	FirstKOCBlock, LastKOCBlock ifaces.Column
	// block deployment
	MinDeplBlock, MaxDeplBlock ifaces.Column
}

/*
scpSelector contains two columns SelectorMinDeplBlock and SelectorMaxDeplBlock
These columns are 1 at indices where the deployment number is equal to MinDeplBlock/MaxDeplBlock, and 0 otherwise
*/
type scpSelector struct {
	SelectorMinDeplBlock, SelectorMaxDeplBlock               ifaces.Column
	ComputeSelectorMinDeplBlock, ComputeSelectorMaxDeplBlock wizard.ProverAction
}

/*
newScpSelector creates the selector columns needed for the connector between the state summary and the HUB arithmetization
these two selectors are only defined for the arithmetization columns
*/
func newScpSelector(comp *wizard.CompiledIOP, smc HubColumnSet) scpSelector {

	SelectorMinDeplNoBlock, ComputeSelectorMinDeplNoBlock := dedicated.IsZero(
		comp,
		sym.Sub(smc.DeploymentNumber, smc.MinDeplBlock),
	)

	SelectorMaxDeplNoBlock, ComputeSelectorMaxDeplNoBlock := dedicated.IsZero(
		comp,
		sym.Sub(smc.DeploymentNumber, smc.MaxDeplBlock),
	)

	res := scpSelector{
		SelectorMinDeplBlock:        SelectorMinDeplNoBlock,
		SelectorMaxDeplBlock:        SelectorMaxDeplNoBlock,
		ComputeSelectorMinDeplBlock: ComputeSelectorMinDeplNoBlock,
		ComputeSelectorMaxDeplBlock: ComputeSelectorMaxDeplNoBlock,
	}

	return res
}

/*
accountIntegrationDefineInitial defines the bidirectional lookups used to check initial account data consistency between
a StateSummary struct corresponding to Shomei traces and a StateManagerColumns struct
(which corresponds to a permutation of the arithmetization's HUB columns, in this case an ACP—account consistency permutation)
For each block, these lookups will check the consistency of the initial account data from the Shomei traces with
the corresponding columns in the arithmetization.
*/
func accountIntegrationDefineInitial(comp *wizard.CompiledIOP, ss Module, smc HubColumnSet) {

	var (
		filterArith = comp.InsertCommit(0,
			"FILTER_CONNECTOR_SUMMARY_ARITHMETIZATION_ACCOUNT_INITIAL_ARITHMETIZATION",
			smc.AddressHI.Size(),
		)

		filterSummary = comp.InsertCommit(0,
			"FILTER_CONNECTOR_SUMMARY_ARITHMETIZATION_ACCOUNT_INITIAL_SUMMARY",
			ss.IsStorage.Size(),
		)

		stateSummaryTable = []ifaces.Column{ss.Account.Address,
			ss.Account.Initial.Balance,
			ss.Account.Initial.Nonce,
			ss.Account.Initial.CodeSize,
			ss.Account.Initial.ExpectedHubCodeHash.Hi,
			ss.Account.Initial.ExpectedHubCodeHash.Lo,
			ss.BatchNumber,
			ss.Account.Initial.Exists,
		}

		arithTable = []ifaces.Column{smc.Address,
			smc.BalanceOld,
			smc.Nonce,
			smc.CodeSizeOld,
			smc.CodeHashHI,
			smc.CodeHashLO,
			smc.BlockNumber,
			smc.Exists,
		}
	)

	comp.InsertInclusionDoubleConditional(0,
		"LOOKUP_STATE_MGR_ARITH_TO_STATE_SUMMARY_INIT_ACCOUNT",
		stateSummaryTable,
		arithTable,
		filterSummary,
		filterArith,
	)

	comp.InsertInclusionDoubleConditional(0,
		"LOOKUP_STATE_MGR_ARITH_TO_STATE_SUMMARY_INIT_ACCOUNT_REVERSED",
		arithTable,
		stateSummaryTable,
		filterArith,
		filterSummary,
	)

	// Now we define the constraints for our filters
	comp.InsertGlobal(
		0,
		ifaces.QueryIDf("CONSTRAINT_FILTER_CONNECTOR_SUMMARY_ARITHMETIZATION_ACCOUNT_INITIAL_ARITHMETIZATION"),
		sym.Sub(
			filterArith,
			sym.Mul(
				smc.PeekAtAccount,
				smc.FirstAOCBlock,
			),
		),
	)

	comp.InsertGlobal(
		0,
		ifaces.QueryIDf("CONSTRAINT_FILTER_CONNECTOR_SUMMARY_ARITHMETIZATION_ACCOUNT_INITIAL_SUMMARY"),
		sym.Sub(
			filterSummary,
			sym.Mul(
				ss.IsInitialDeployment,
				sym.Sub(
					1,
					ss.IsStorage,
				),
			),
		),
	)
}

/*
accountIntegrationAssignInitial assigns the columns used to check initial account
data consistency using the lookups from AccountIntegrationDefineInitial
*/
func accountIntegrationAssignInitial(run *wizard.ProverRuntime, ss Module, smc HubColumnSet) {

	svfilterArith := smartvectors.Mul(smc.PeekAtAccount.GetColAssignment(run), smc.FirstAOCBlock.GetColAssignment(run))

	run.AssignColumn("FILTER_CONNECTOR_SUMMARY_ARITHMETIZATION_ACCOUNT_INITIAL_ARITHMETIZATION", svfilterArith)

	selectorNotStorage := make([]field.Element, ss.IsStorage.Size())

	for index := range selectorNotStorage {
		isStorage := ss.IsStorage.GetColAssignmentAt(run, index)
		if isStorage.IsZero() {
			selectorNotStorage[index].SetOne()
		}
	}

	svSelectorNotStorage := smartvectors.NewRegular(selectorNotStorage)
	svfilterSummary := smartvectors.Mul(svSelectorNotStorage, ss.IsInitialDeployment.GetColAssignment(run))

	run.AssignColumn("FILTER_CONNECTOR_SUMMARY_ARITHMETIZATION_ACCOUNT_INITIAL_SUMMARY", svfilterSummary)
}

/*
accountIntegrationDefineFinal defines the bidirectional lookups used to check final account data consistency between
a StateSummary struct corresponding to Shomei traces and a StateManagerColumns struct
(which corresponds to a permutation of the arithmetization's HUB columns, in this case an ACP—account consistency permutation)
For each block, these lookups will check the consistency of the final account data from the Shomei traces with
the corresponding columns in the arithmetization.
*/
func accountIntegrationDefineFinal(comp *wizard.CompiledIOP, ss Module, smc HubColumnSet) {
	filterArith := comp.InsertCommit(0, "FILTER_CONNECTOR_SUMMARY_ARITHMETIZATION_ACCOUNT_FINAL_ARITHMETIZATION", smc.AddressHI.Size())
	filterSummary := comp.InsertCommit(0, "FILTER_CONNECTOR_SUMMARY_ARITHMETIZATION_ACCOUNT_FINAL_SUMMARY", ss.IsStorage.Size())
	stateSummaryTable := []ifaces.Column{
		ss.Account.Address,
		ss.Account.Final.Balance,
		ss.Account.Final.Nonce,
		ss.Account.Final.CodeSize,
		ss.Account.Final.ExpectedHubCodeHash.Hi,
		ss.Account.Final.ExpectedHubCodeHash.Lo,
		ss.BatchNumber,
		ss.Account.Final.Exists,
	}
	arithTable := []ifaces.Column{
		smc.Address,
		smc.BalanceNew,
		smc.NonceNew,
		smc.CodeSizeNew,
		smc.CodeHashHINew,
		smc.CodeHashLONew,
		smc.BlockNumber,
		smc.ExistsNew,
	}

	comp.InsertInclusionDoubleConditional(0, "LOOKUP_STATE_MGR_ARITH_TO_STATE_SUMMARY_FINAL_ACCOUNT", stateSummaryTable, arithTable, filterSummary, filterArith)
	comp.InsertInclusionDoubleConditional(0, "LOOKUP_STATE_MGR_ARITH_TO_STATE_SUMMARY_FINAL_ACCOUNT_REVERSED", arithTable, stateSummaryTable, filterArith, filterSummary)

	// Now we define the constraints for our filters
	comp.InsertGlobal(
		0,
		ifaces.QueryIDf("CONSTRAINT_FILTER_CONNECTOR_SUMMARY_ARITHMETIZATION_ACCOUNT_FINAL_ARITHMETIZATION"),
		sym.Sub(
			filterArith,
			sym.Mul(
				smc.PeekAtAccount,
				smc.LastAOCBlock,
			),
		),
	)

	comp.InsertGlobal(
		0,
		ifaces.QueryIDf("CONSTRAINT_FILTER_CONNECTOR_SUMMARY_ARITHMETIZATION_ACCOUNT_FINAL_SUMMARY"),
		sym.Sub(
			filterSummary,
			sym.Mul(
				ss.IsFinalDeployment,
				sym.Sub(
					1,
					ss.IsStorage,
				),
			),
		),
	)
}

/*
accountIntegrationAssignFinal assigns the columns used to check initial account data consistency using the lookups from accountIntegrationAssignFinal
*/
func accountIntegrationAssignFinal(run *wizard.ProverRuntime, ss Module, smc HubColumnSet) {
	filterArith := smartvectors.Mul(smc.PeekAtAccount.GetColAssignment(run), smc.LastAOCBlock.GetColAssignment(run))

	run.AssignColumn("FILTER_CONNECTOR_SUMMARY_ARITHMETIZATION_ACCOUNT_FINAL_ARITHMETIZATION", filterArith)

	selectorNotStorage := make([]field.Element, ss.IsStorage.Size())

	for index := range selectorNotStorage {
		isStorage := ss.IsStorage.GetColAssignmentAt(run, index)
		if isStorage.IsZero() {
			selectorNotStorage[index].SetOne()
		}
	}

	svSelectorNotStorage := smartvectors.NewRegular(selectorNotStorage)
	svfilterSummary := smartvectors.Mul(svSelectorNotStorage, ss.IsFinalDeployment.GetColAssignment(run))

	run.AssignColumn("FILTER_CONNECTOR_SUMMARY_ARITHMETIZATION_ACCOUNT_FINAL_SUMMARY", svfilterSummary)
}

/*
storageIntegrationDefineInitial defines the bidirectional lookups used to check initial storage data consistency between
a StateSummary struct corresponding to Shomei traces and a StateManagerColumns struct
(which corresponds to a permutation of the arithmetization's HUB columns, in this case an SCP—storage consistency permutation)
For each block, these lookups will check the consistency of the initial storage data from the Shomei traces with
the corresponding columns in the arithmetization.
*/
func storageIntegrationDefineInitial(comp *wizard.CompiledIOP, ss Module, smc HubColumnSet, sc scpSelector) {
	filterArith := comp.InsertCommit(0, "FILTER_CONNECTOR_SUMMARY_ARITHMETIZATION_STORAGE_INITIAL_ARITHMETIZATION", smc.AddressHI.Size())
	filterSummary := comp.InsertCommit(0, "FILTER_CONNECTOR_SUMMARY_ARITHMETIZATION_STORAGE_INITIAL_SUMMARY", ss.Account.Address.Size())

	filterArithReversed := comp.InsertCommit(0, "FILTER_CONNECTOR_SUMMARY_ARITHMETIZATION_STORAGE_INITIAL_ARITHMETIZATION_REVERSED", smc.AddressHI.Size())

	summaryTable := []ifaces.Column{
		ss.Account.Address,
		ss.Storage.Key.Hi,
		ss.Storage.Key.Lo,
		ss.Storage.OldValue.Hi,
		ss.Storage.OldValue.Lo,
		ss.BatchNumber,
	}
	arithTable := []ifaces.Column{
		smc.Address,
		smc.KeyHI,
		smc.KeyLO,
		smc.ValueHICurr,
		smc.ValueLOCurr,
		smc.BlockNumber,
	}
	comp.InsertInclusionDoubleConditional(
		0,
		"LOOKUP_STATE_MGR_ARITH_TO_STATE_SUMMARY_INIT_STORAGE",
		summaryTable,
		arithTable,
		filterSummary,
		filterArith,
	)
	comp.InsertInclusionDoubleConditional(
		0,
		"LOOKUP_STATE_MGR_ARITH_TO_STATE_SUMMARY_INIT_STORAGE_REVERSE",
		arithTable,
		summaryTable,
		filterArithReversed,
		filterSummary,
	)

	// Now we define the constraints for our filters
	comp.InsertGlobal(
		0,
		ifaces.QueryIDf("CONSTRAINT_FILTER_CONNECTOR_SUMMARY_ARITHMETIZATION_STORAGE_INITIAL_ARITHMETIZATION"),
		sym.Sub(
			filterArith,
			sym.Mul(
				sc.SelectorMinDeplBlock,
				smc.PeekAtStorage,
				smc.FirstKOCBlock,
			),
		),
	)

	comp.InsertGlobal(
		0,
		ifaces.QueryIDf("CONSTRAINT_FILTER_CONNECTOR_SUMMARY_ARITHMETIZATION_STORAGE_INITIAL_ARITHMETIZATION_REVERSED"),
		sym.Sub(
			filterArithReversed,
			sym.Mul(
				smc.PeekAtStorage,
				smc.FirstKOCBlock,
			),
		),
	)

	comp.InsertGlobal(
		0,
		ifaces.QueryIDf("CONSTRAINT_FILTER_CONNECTOR_SUMMARY_ARITHMETIZATION_STORAGE_INITIAL_SUMMARY"),
		sym.Sub(
			filterSummary,
			sym.Mul(
				ss.IsStorage,
				ss.IsInitialDeployment,
			),
		),
	)
}

/*
storageIntegrationAssignInitial assigns the columns used to check initial storage data consistency using the lookups from StorageIntegrationDefineInitial
*/
func storageIntegrationAssignInitial(run *wizard.ProverRuntime, ss Module, smc HubColumnSet) {
	filterSummary := smartvectors.Mul(ss.IsStorage.GetColAssignment(run), ss.IsInitialDeployment.GetColAssignment(run))
	run.AssignColumn("FILTER_CONNECTOR_SUMMARY_ARITHMETIZATION_STORAGE_INITIAL_SUMMARY", filterSummary)

	selectorMinDeplBlock := make([]field.Element, smc.AddressHI.Size())

	for index := range selectorMinDeplBlock {
		minDeplBlock := smc.MinDeplBlock.GetColAssignmentAt(run, index)
		deplNumber := smc.DeploymentNumber.GetColAssignmentAt(run, index)
		if minDeplBlock.Equal(&deplNumber) {
			selectorMinDeplBlock[index].SetOne()
		}
	}
	svSelectorMinDeplBlock := smartvectors.NewRegular(selectorMinDeplBlock)

	filterArith := smartvectors.Mul(svSelectorMinDeplBlock, smc.PeekAtStorage.GetColAssignment(run), smc.FirstKOCBlock.GetColAssignment(run))
	run.AssignColumn("FILTER_CONNECTOR_SUMMARY_ARITHMETIZATION_STORAGE_INITIAL_ARITHMETIZATION", filterArith)

	/*
		When looking up with including = {arithmetization} and included = {State summary}, we remove the MinDeplBlock filter selector
		(arithmetization keys might be read after the first deployment in the block)
	*/
	filterArithReversed := smartvectors.Mul(smc.PeekAtStorage.GetColAssignment(run), smc.FirstKOCBlock.GetColAssignment(run))
	run.AssignColumn("FILTER_CONNECTOR_SUMMARY_ARITHMETIZATION_STORAGE_INITIAL_ARITHMETIZATION_REVERSED", filterArithReversed)
}

/*
storageIntegrationDefineFinal defines the bidirectional lookups used to check final storage data consistency between
a StateSummary struct corresponding to Shomei traces and a StateManagerColumns struct
(which corresponds to a permutation of the arithmetization's HUB columns, in this case an SCP—storage consistency permutation)
For each block, these lookups will check the consistency of the final storage data from the Shomei traces with
the corresponding columns in the arithmetization.
*/
func storageIntegrationDefineFinal(comp *wizard.CompiledIOP, ss Module, smc HubColumnSet, sc scpSelector) {

	var (
		summaryTable = []ifaces.Column{
			ss.Account.Address,
			ss.Storage.Key.Hi,
			ss.Storage.Key.Lo,
			ss.Storage.NewValue.Hi,
			ss.Storage.NewValue.Lo,
			ss.BatchNumber,
		}

		arithTable = []ifaces.Column{smc.Address,
			smc.KeyHI,
			smc.KeyLO,
			smc.ValueHINext,
			smc.ValueLONext,
			smc.BlockNumber,
		}

		filterArith = comp.InsertCommit(0,
			"FILTER_CONNECTOR_SUMMARY_ARITHMETIZATION_STORAGE_FINAL_ARITHMETIZATION",
			smc.AddressHI.Size(),
		)

		filterArithReversed = comp.InsertCommit(0,
			"FILTER_CONNECTOR_SUMMARY_ARITHMETIZATION_STORAGE_FINAL_ARITHMETIZATION_REVERSED",
			smc.AddressHI.Size(),
		)

		filterSummary = comp.InsertCommit(0,
			"FILTER_CONNECTOR_SUMMARY_ARITHMETIZATION_STORAGE_FINAL_SUMMARY",
			ss.Account.Address.Size(),
		)

		filterAccountInsert = comp.InsertCommit(0,
			"FILTER_CONNECTOR_HUB_STATE_SUMMARY_ACCOUNT_INSERT_FILTER",
			smc.AddressHI.Size(),
		)
	)

	comp.InsertInclusionDoubleConditional(0,
		"LOOKUP_STATE_MGR_ARITH_TO_STATE_SUMMARY_FINAL_STORAGE",
		summaryTable,
		arithTable,
		filterSummary,
		filterArith,
	)

	comp.InsertInclusionDoubleConditional(0,
		"LOOKUP_STATE_MGR_ARITH_TO_STATE_SUMMARY_FINAL_STORAGE_REVERSED",
		arithTable,
		summaryTable,
		filterArithReversed,
		filterSummary,
	)

	comp.InsertGlobal(
		0,
		ifaces.QueryIDf("CONSTRAINT_FILTER_CONNECTOR_SUMMARY_ARITHMETIZATION_STORAGE_FINAL_ARITHMETIZATION"),
		sym.Sub(
			filterArith,
			sym.Mul(
				sc.SelectorMaxDeplBlock,
				smc.PeekAtStorage,
				smc.LastKOCBlock,
				filterAccountInsert,
			),
		),
	)

	comp.InsertGlobal(
		0,
		ifaces.QueryIDf("CONSTRAINT_FILTER_REVERSED_CONNECTOR_SUMMARY_ARITHMETIZATION_STORAGE_FINAL_ARITHMETIZATION"),
		sym.Sub(
			filterArithReversed,
			sym.Mul(
				sc.SelectorMaxDeplBlock,
				smc.PeekAtStorage,
				smc.LastKOCBlock,
			),
		),
	)

	comp.InsertGlobal(
		0,
		ifaces.QueryIDf("CONSTRAINT_FILTER_CONNECTOR_SUMMARY_ARITHMETIZATION_STORAGE_FINAL_SUMMARY"),
		sym.Sub(
			filterSummary,
			sym.Mul(
				ss.IsStorage,
				ss.IsFinalDeployment,
			),
		),
	)

	// constraint the insertion selector filter
	existsFirstInBlock := comp.Columns.GetHandle("hub.scp_EXISTS_FIRST_IN_BLOCK")
	existsFinalInBlock := comp.Columns.GetHandle("hub.scp_EXISTS_FINAL_IN_BLOCK")
	// on storage rows, we enforce that filterAccountInsert is 0 then (existsFirstInBlock = 0 and existsFinalInBlock = 1)
	// security of the following constraint relies on the fact that the underlying marker columns are binary
	comp.InsertGlobal(
		0,
		ifaces.QueryIDf("GLOBAL_CONSTRAINT_FILTER_CONNECTOR_HUB_STATE_SUMMARY_ACCOUNT_INSERT_FILTER"),
		sym.Mul(
			smc.PeekAtStorage, // when we are dealing with storage segments
			sym.Mul(
				sym.Sub(
					1,
					filterAccountInsert,
				), // if  filterAccountInsert = 0 it must be that the conditions of the filter are both satisfied
				sym.Add(
					existsFirstInBlock,
					sym.Sub(
						1,
						existsFinalInBlock,
					),
				),
			),
		),
	)
	comp.InsertGlobal(
		0,
		ifaces.QueryIDf("GLOBAL_CONSTRAINT_FILTER_CONNECTOR_HUB_STATE_SUMMARY_ACCOUNT_INSERT_FILTER_ZEROIZATION"),
		sym.Mul(
			sym.Sub(
				1,
				smc.PeekAtStorage,
			), // when we are not dealing with storage segments
			sym.Sub(
				1,
				filterAccountInsert,
			), // filterAccountInsert must be 1
		),
	)
	mustBeBinary(comp, filterAccountInsert)
}

/*
storageIntegrationAssignFinal assigns the columns used to check initial storage data consistency using the lookups from StorageIntegrationDefineFinal
*/
func storageIntegrationAssignFinal(run *wizard.ProverRuntime, ss Module, smc HubColumnSet) {
	selectorMaxDeplBlock := make([]field.Element, smc.AddressHI.Size())
	for index := range selectorMaxDeplBlock {
		maxDeplBlock := smc.MaxDeplBlock.GetColAssignmentAt(run, index)
		deplNumber := smc.DeploymentNumber.GetColAssignmentAt(run, index)
		if maxDeplBlock.Equal(&deplNumber) {
			selectorMaxDeplBlock[index].SetOne()
		}
	}
	svSelectorMaxDeplBlock := smartvectors.NewRegular(selectorMaxDeplBlock)

	filterSummary := smartvectors.Mul(
		ss.IsStorage.GetColAssignment(run),
		ss.IsFinalDeployment.GetColAssignment(run),
	)
	run.AssignColumn("FILTER_CONNECTOR_SUMMARY_ARITHMETIZATION_STORAGE_FINAL_SUMMARY", filterSummary)
	filterArithReversed := smartvectors.Mul(
		svSelectorMaxDeplBlock,
		smc.PeekAtStorage.GetColAssignment(run),
		smc.LastKOCBlock.GetColAssignment(run),
	)
	run.AssignColumn("FILTER_CONNECTOR_SUMMARY_ARITHMETIZATION_STORAGE_FINAL_ARITHMETIZATION_REVERSED", filterArithReversed)

	// compute the filter that detects account inserts in order to exclude those key reads from the
	// arithmetization to state summary lookups.
	existsFirstInBlock := run.Spec.Columns.GetHandle("hub.scp_EXISTS_FIRST_IN_BLOCK")
	existsFinalInBlock := run.Spec.Columns.GetHandle("hub.scp_EXISTS_FINAL_IN_BLOCK")
	filterAccountInsert := make([]field.Element, smc.AddressHI.Size())
	lastSegmentStart := 0
	for index := range filterAccountInsert {
		filterAccountInsert[index].SetOne() // always set the filter as one, unless we detect an insertion segment
		isStorage := smc.PeekAtStorage.GetColAssignmentAt(run, index)
		if isStorage.IsOne() {
			firstKOCBlock := smc.FirstKOCBlock.GetColAssignmentAt(run, index)
			lastKOCBlock := smc.LastKOCBlock.GetColAssignmentAt(run, index)
			existsAtBlockEnd := existsFinalInBlock.GetColAssignmentAt(run, index)

			if firstKOCBlock.IsOne() {
				// remember when the segment starts
				lastSegmentStart = index
			}
			if lastKOCBlock.IsOne() && existsAtBlockEnd.IsOne() {
				existsAtBlockStart := existsFirstInBlock.GetColAssignmentAt(run, lastSegmentStart)
				if existsAtBlockStart.IsZero() {
					// we are indeed dealing with an insertion segment
					for j := lastSegmentStart; j <= index; j++ {
						// set the filter to zeros on the insertion segment
						filterAccountInsert[j].SetZero()
					}
				}
			}
		}

	}
	svfilterAccountInsert := smartvectors.NewRegular(filterAccountInsert)
	run.AssignColumn("FILTER_CONNECTOR_HUB_STATE_SUMMARY_ACCOUNT_INSERT_FILTER", svfilterAccountInsert)

	//filterTxExec := run.Spec.Columns.GetHandle("hub.scp_TX_EXEC")
	filterArith := smartvectors.Mul(
		svSelectorMaxDeplBlock,
		smc.PeekAtStorage.GetColAssignment(run),
		smc.LastKOCBlock.GetColAssignment(run),
		svfilterAccountInsert,
	)
	run.AssignColumn("FILTER_CONNECTOR_SUMMARY_ARITHMETIZATION_STORAGE_FINAL_ARITHMETIZATION", filterArith)
}
