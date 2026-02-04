package statemanager

import (
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/utils"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/utils/types"
)

// return the account of a trace. location for storage trie updates) and key for
// world-state update
func (trace DecodedTrace) GetRelatedAccount() (address Address, err error) {
	// The location should cast to either an address or to the WORLD_STATE
	// locator. This has not been checked yet. So it's a soft error.
	switch u := trace.Underlying.(type) {
	case ReadNonZeroTraceST, ReadZeroTraceST, InsertionTraceST, UpdateTraceST, DeletionTraceST:
		return types.AddressFromHex(trace.Location)
	case ReadNonZeroTraceWS:
		return u.Key, nil
	case ReadZeroTraceWS:
		return u.Key, nil
	case InsertionTraceWS:
		return u.Key, nil
	case UpdateTraceWS:
		return u.Key, nil
	case DeletionTraceWS:
		return u.Key, nil
	}
	utils.Panic("unknown underlying type: %T", trace.Underlying)
	return Address{}, nil
}
