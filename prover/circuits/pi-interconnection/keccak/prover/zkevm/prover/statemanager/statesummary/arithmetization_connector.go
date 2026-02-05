package statesummary

import (
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/protocol/wizard"
)

// arithmetizationLink collects columns from the hub that are of interest for
// checking consistency between the StateSummary and the rest of the
// arithmetization.
type arithmetizationLink struct {
	Acp, Scp    HubColumnSet
	ScpSelector ScpSelector
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
	// account existence information, to detect account pattern if needed
	ExistsFirstInBlock ifaces.Column
	ExistsFinalInBlock ifaces.Column
}

/*
ScpSelector contains two columns SelectorMinDeplBlock and SelectorMaxDeplBlock
These columns are 1 at indices where the deployment number is equal to MinDeplBlock/MaxDeplBlock, and 0 otherwise
*/
type ScpSelector struct {
	SelectorMinDeplBlock, SelectorMaxDeplBlock               ifaces.Column
	ComputeSelectorMinDeplBlock, ComputeSelectorMaxDeplBlock wizard.ProverAction
	// selectors for empty keys, current values
	SelectorEmptySTValueHi, SelectorEmptySTValueLo               ifaces.Column
	ComputeSelectorEmptySTValueHi, ComputeSelectorEmptySTValueLo wizard.ProverAction
	// selectors for empty keys, next values
	SelectorEmptySTValueNextHi, SelectorEmptySTValueNextLo               ifaces.Column
	ComputeSelectorEmptySTValueNextHi, ComputeSelectorEmptySTValueNextLo wizard.ProverAction
	// storage key difference selectors
	SelectorSTKeyDiffHi, SelectorSTKeyDiffLo               ifaces.Column
	ComputeSelectorSTKeyDiffHi, ComputeSelectorSTKeyDiffLo wizard.ProverAction
	// Account Address Diff
	SelectorAccountAddressDiff        ifaces.Column
	ComputeSelectorAccountAddressDiff wizard.ProverAction
	// block number key difference selectors
	SelectorBlockNoDiff        ifaces.Column
	ComputeSelectorBlockNoDiff wizard.ProverAction
}
