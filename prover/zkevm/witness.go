package zkevm

import (
	"math/big"

	"github.com/consensys/zkevm-monorepo/prover/backend/ethereum"
	"github.com/consensys/zkevm-monorepo/prover/backend/execution/statemanager"
	"github.com/ethereum/go-ethereum/common"
)

// Witness is a collection of prover inputs used to derive an assignment to the
// full proving scheme.
type Witness struct {
	// ExecTracesFPath is the filepath toward the execution traces to use for
	// proof trace generation.
	ExecTracesFPath string
	// StateManager traces
	SMTraces        [][]statemanager.DecodedTrace
	TxSignatures    map[[32]byte]ethereum.Signature
	L2BridgeAddress common.Address
	ChainID         uint
}

func (w Witness) TxSignatureGetter(txHash []byte) (r, s, v *big.Int, err error) {
	panic("unimplemented")
}
