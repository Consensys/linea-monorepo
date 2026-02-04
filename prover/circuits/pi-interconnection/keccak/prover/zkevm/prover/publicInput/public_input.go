package publicInput

import (
	arith "github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/zkevm/prover/publicInput/arith_struct"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/zkevm/prover/publicInput/logs"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/zkevm/prover/statemanager/statesummary"
)

var (
	DataNbBytes                  = "DataNbBytes"
	DataChecksum                 = "DataChecksum"
	L2MessageHash                = "L2MessageHash"
	InitialStateRootHash         = "InitialStateRootHash"
	FinalStateRootHash           = "FinalStateRootHash"
	InitialBlockNumber           = "InitialBlockNumber"
	FinalBlockNumber             = "FinalBlockNumber"
	InitialBlockTimestamp        = "InitialBlockTimestamp"
	FinalBlockTimestamp          = "FinalBlockTimestamp"
	FirstRollingHashUpdate_0     = "FirstRollingHashUpdate_0"
	FirstRollingHashUpdate_1     = "FirstRollingHashUpdate_1"
	LastRollingHashUpdate_0      = "LastRollingHashUpdate_0"
	LastRollingHashUpdate_1      = "LastRollingHashUpdate_1"
	FirstRollingHashUpdateNumber = "FirstRollingHashUpdateNumber"
	LastRollingHashNumberUpdate  = "LastRollingHashNumberUpdate"
	ChainID                      = "ChainID"
	NBytesChainID                = "NBytesChainID"
	L2MessageServiceAddrHi       = "L2MessageServiceAddrHi"
	L2MessageServiceAddrLo       = "L2MessageServiceAddrLo"
)

// Settings contains options for proving and verifying that the public inputs are computed properly.
type Settings struct {
	Name string
}

// InputModules groups several arithmetization modules needed to compute the public input.
type InputModules struct {
	BlockData    *arith.BlockDataCols
	TxnData      *arith.TxnData
	RlpTxn       *arith.RlpTxn
	LogCols      logs.LogColumns
	StateSummary *statesummary.Module
}
