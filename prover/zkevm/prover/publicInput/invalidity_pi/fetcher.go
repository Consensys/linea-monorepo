package invalidity

import (
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/zkevm/arithmetization"
	fetch "github.com/consensys/linea-monorepo/prover/zkevm/prover/publicInput/fetchers_arithmetization"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/publicInput/logs"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/statemanager/statesummary"
	"github.com/ethereum/go-ethereum/common"
)

// PublicInputFetcher fetches logs and root hash
type PublicInputFetcher struct {
	LogCols         logs.LogColumns
	FetchedL2L1     logs.ExtractedData
	LogSelectors    logs.Selectors
	RootHashFetcher *fetch.RootHashFetcher
	StateSummary    *statesummary.Module
}

func GetLogColumns(comp *wizard.CompiledIOP, arith *arithmetization.Arithmetization) logs.LogColumns {

	return logs.LogColumns{
		IsLog0:       arith.ColumnOf(comp, "loginfo", "IS_LOG_X_0"),
		IsLog1:       arith.ColumnOf(comp, "loginfo", "IS_LOG_X_1"),
		IsLog2:       arith.ColumnOf(comp, "loginfo", "IS_LOG_X_2"),
		IsLog3:       arith.ColumnOf(comp, "loginfo", "IS_LOG_X_3"),
		IsLog4:       arith.ColumnOf(comp, "loginfo", "IS_LOG_X_4"),
		AbsLogNum:    arith.MashedColumnOf(comp, "loginfo", "ABS_LOG_NUM"),
		AbsLogNumMax: arith.MashedColumnOf(comp, "loginfo", "ABS_LOG_NUM_MAX"),
		Ct:           arith.ColumnOf(comp, "loginfo", "CT"),
		Data:         arith.GetLimbsOfU128Be(comp, "loginfo", "DATA_LO").ZeroExtendToSize(16).LimbsArr16(),
		TxEmitsLogs:  arith.ColumnOf(comp, "loginfo", "TXN_EMITS_LOGS"),
	}
}

// NewPublicInputFetcher return properly constrained logs and root hash fetchers
func NewPublicInputFetcher(comp *wizard.CompiledIOP, ss *statesummary.Module, logCols logs.LogColumns) PublicInputFetcher {
	name := "INVALIDITY_PI"
	// Logs: Create FetchedL2L1 for extracting L2L1 logs
	fetchedL2L1 := logs.NewExtractedData(comp, logCols.Ct.Size(), name+"_L2L1LOGS")

	// Log selectors
	logSelectors := logs.NewSelectorColumns(comp, logCols)
	// Define the extracted data constraints
	logs.DefineExtractedData(comp, logCols, logSelectors, fetchedL2L1, logs.L2L1)

	// RootHash fetcher from the StateSummary
	rootHashFetcher := fetch.NewRootHashFetcher(comp, name+"_ROOT_HASH_FETCHER", ss.IsActive.Size())
	fetch.DefineRootHashFetcher(comp, rootHashFetcher, name+"_ROOT_HASH_FETCHER", *ss)

	mpi := PublicInputFetcher{
		LogCols:         logCols,
		FetchedL2L1:     fetchedL2L1,
		LogSelectors:    logSelectors,
		RootHashFetcher: rootHashFetcher,
		StateSummary:    ss,
	}

	return mpi
}

// Assign assigns values to the  columns of the fetcher.
func (pi *PublicInputFetcher) Assign(run *wizard.ProverRuntime, l2BridgeAddress common.Address) {
	// Assign the root hash fetcher (reads from StateSummary
	fetch.AssignRootHashFetcher(run, pi.RootHashFetcher, *pi.StateSummary)

	// Assign log selectors and extracted data
	pi.LogSelectors.Assign(run, l2BridgeAddress)
	logs.AssignExtractedData(run, pi.LogCols, pi.LogSelectors, pi.FetchedL2L1, logs.L2L1)
}
