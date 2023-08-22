package eth

import (
	"fmt"
	"reflect"

	"github.com/consensys/accelerated-crypto-monorepo/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"
)

// Inspect the traces and check if they are consistent
// with what the spec allows
func CheckTraces(traces []any) (oldStateRootHash Digest, newStateRootHash Digest, err error) {

	var prevAddress Address

	if len(traces) == 0 {
		utils.Panic("no state-manager traces, that's impossible.")
	}

	// Dispatch the traces to separate traces relating to
	// different accounts
	traceByAccount := [][]any{}
	traceWs := []any{}
	digestErr := Digest{}

	// Traces for the same account should be continuous
	alreadyFoundAcc := map[Address]struct{}{}

	// Collect all the traces in their respective slices. We also
	// check that all checks done relative to an account have been
	// done contiguously.
	for i, trace := range traces {
		// Push if a WS
		if isWorldStateTrace(trace) {
			traceWs = append(traceWs, trace)
		}

		address, err := getTraceAccount(trace)
		if err != nil {
			return digestErr, digestErr, err
		}

		// Ensures we have at most one segment for each address
		if _, ok := alreadyFoundAcc[address]; ok && address != prevAddress && i > 0 {
			return digestErr, digestErr, fmt.Errorf("two segments for address %v", address.Hex())
		}

		// If the account changed, push into a new slice
		if i == 0 || address != prevAddress {
			traceByAccount = append(traceByAccount, []any{})
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

// true iff the trace is on the world-state
func isWorldStateTrace(trace any) bool {
	// We use reflection to recover the location. The policy is to
	// panic on failure in this function. The JSON parsing covers
	// the fact that the `Location` must exist and be a string.
	location := getLocation(trace)
	return location == WS_LOCATION
}

// return the account of a WS trace. panic if not WS
func accountForWSTrace(trace any) Address {
	// case of an account trie access. the locator is the `Key` field
	// of the traces. We crucially rely on the fact that the name of
	// the field is "Key" and it's a hard error if not.
	address := reflect.ValueOf(trace).FieldByName("Key").Interface().(Address)

	// sanity-check : log. It's not illegal to have the "0x0" address
	// but it's strange so we log it.
	if (address == Address{}) {
		logrus.Warnf("found locator for address 0x0. Might be a parsing error. The underlying trace is %++v. Please check if consistent", trace)
	}
	return address
}

// return the location of a trace. panic if not found
func getLocation(trace any) string {
	return reflect.ValueOf(trace).FieldByName("Location").Interface().(string)
}

// return the account of a storage trie trace. Panic if no
// locator is found. soft-error if the locator is not a valid
// address.
func accountForSTTrace(trace any) (Address, error) {
	location := getLocation(trace)
	if !IsHexAddress(location) {
		return Address{}, fmt.Errorf("location is not a valid address : %v", location)
	}
	return Address(common.HexToAddress(location)), nil
}

// return the account of a trace. location for storage trie
// updates) and key for world-state update
func getTraceAccount(trace any) (address Address, err error) {
	// The location should cast to either an address or to the
	// WORLD_STATE locator. This has not been checked yet. So
	// it's a soft error.
	if isWorldStateTrace(trace) {
		return accountForWSTrace(trace), nil
	}

	address, err = accountForSTTrace(trace)
	if err != nil {
		return Address{}, fmt.Errorf("the locator %v is invalid : %v", getLocation(trace), err)
	}

	return address, nil
}
