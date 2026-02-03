package statemanager

import (
	"fmt"

	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/utils"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/utils/types"
)

// Inspect the traces and check if they are consistent with what the spec allows
func CheckTraces(traces []DecodedTrace) (oldStateRootHash Digest, newStateRootHash Digest, err error) {

	var prevAddress Address

	if len(traces) == 0 {
		utils.Panic("no state-manager traces, that's impossible.")
	}

	// Dispatch the traces to separate traces relating to different accounts
	traceByAccount := [][]DecodedTrace{}
	traceWs := []DecodedTrace{}
	digestErr := Digest{}

	// Traces for the same account should be continuous
	alreadyFoundAcc := map[Address]struct{}{}

	// Collect all the traces in their respective slices. We also check that all
	// checks done relative to an account have been done contiguously.
	for i, trace := range traces {
		// Push if a WS
		if trace.isWorldState() {
			traceWs = append(traceWs, trace)
		}

		address, err := trace.GetRelatedAccount()
		if err != nil {
			return digestErr, digestErr, err
		}

		// Ensures we have at most one segment for each address
		if _, ok := alreadyFoundAcc[address]; ok && address != prevAddress && i > 0 {
			return digestErr, digestErr, fmt.Errorf("two segments for address %v", address.Hex())
		}

		// If the account changed, push into a new slice
		if i == 0 || address != prevAddress {
			traceByAccount = append(traceByAccount, []DecodedTrace{})
		}

		last := len(traceByAccount) - 1
		traceByAccount[last] = append(traceByAccount[last], trace)
		alreadyFoundAcc[address] = struct{}{}
		prevAddress = address
	}

	// Then audit the traces by account
	for _, traces := range traceByAccount {
		// run the pattern inspection before the proof verification
		if err := inspectPattern(traces); err != nil {
			return digestErr, digestErr, err
		}
		// run the proof verification on the account
		if err := checkProofsForAccount(traces); err != nil {
			return digestErr, digestErr, err
		}
	}

	// Finally check the proof for the world state
	return checkProofsWorldState(traceWs)
}

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
