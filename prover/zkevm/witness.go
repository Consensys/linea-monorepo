package zkevm

import (
	"fmt"
	"math/big"

	"github.com/consensys/linea-monorepo/prover/backend/ethereum"
	"github.com/consensys/linea-monorepo/prover/backend/execution/statemanager"
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
	var (
		sig, found = w.TxSignatures[[32]byte(txHash)]
	)

	if !found {
		return nil, nil, nil, fmt.Errorf("could not find signature for tx hash = 0x%x", txHash)
	}

	r, _ = new(big.Int).SetString(sig.R, 0)
	s, _ = new(big.Int).SetString(sig.S, 0)
	v, _ = new(big.Int).SetString(sig.V, 0)

	return r, s, v, nil
}
