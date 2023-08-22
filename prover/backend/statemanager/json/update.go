package json

import (
	"github.com/consensys/accelerated-crypto-monorepo/backend/jsonutil"
	"github.com/consensys/accelerated-crypto-monorepo/backend/statemanager/eth"
	"github.com/valyala/fastjson"
)

/*
	Schema:

	location
	newNextFreeNode
	oldSubRoot
	newSubRoot
	proof
	key
	oldValue
	newValue
	priorUpdatedLeaf
*/

// Parse update trace for a storage trie
func ParseUpdateTraceST(v fastjson.Value) (eth.UpdateTraceST, error) {

	trace, traceErr := eth.UpdateTraceST{}, eth.UpdateTraceST{}
	var err error

	trace.Location, err = jsonutil.TryGetString(v, LOCATION)
	if err != nil {
		return traceErr, err
	}

	trace.NewNextFreeNode, err = jsonutil.TryGetInt(v, NEW_NEXT_FREE_NODE)
	if err != nil {
		return traceErr, err
	}

	trace.OldSubRoot, err = jsonutil.TryGetDigest(v, OLD_SUBROOT)
	if err != nil {
		return traceErr, err
	}

	trace.NewSubRoot, err = jsonutil.TryGetDigest(v, NEW_SUBROOT)
	if err != nil {
		return traceErr, err
	}

	trace.Proof, err = TryParseMerkleProof(v, PROOF)
	if err != nil {
		return traceErr, err
	}

	trace.Key, err = jsonutil.TryGetBytes32(v, KEY)
	if err != nil {
		return traceErr, err
	}

	trace.OldValue, err = jsonutil.TryGetBytes32(v, OLD_VALUE)
	if err != nil {
		return traceErr, err
	}

	trace.NewValue, err = jsonutil.TryGetBytes32(v, NEW_VALUE)
	if err != nil {
		return traceErr, err
	}

	trace.OldOpening, err = TryParseLeafOpening(v, PRIOR_UPDATED_LEAF)
	if err != nil {
		return traceErr, err
	}

	return trace, nil
}

// Parse update trace for the world state
func ParseUpdateTraceWS(v fastjson.Value) (eth.UpdateTraceWS, error) {

	trace, traceErr := eth.UpdateTraceWS{}, eth.UpdateTraceWS{}
	var err error

	trace.Location, err = jsonutil.TryGetString(v, LOCATION)
	if err != nil {
		return traceErr, err
	}

	trace.NewNextFreeNode, err = jsonutil.TryGetInt(v, NEW_NEXT_FREE_NODE)
	if err != nil {
		return traceErr, err
	}

	trace.OldSubRoot, err = jsonutil.TryGetDigest(v, OLD_SUBROOT)
	if err != nil {
		return traceErr, err
	}

	trace.NewSubRoot, err = jsonutil.TryGetDigest(v, NEW_SUBROOT)
	if err != nil {
		return traceErr, err
	}

	trace.Proof, err = TryParseMerkleProof(v, PROOF)
	if err != nil {
		return traceErr, err
	}

	trace.Key, err = jsonutil.TryGetAddress(v, KEY)
	if err != nil {
		return traceErr, err
	}

	trace.OldValue, err = tryParseAccount(v, OLD_VALUE)
	if err != nil {
		return traceErr, err
	}

	trace.NewValue, err = tryParseAccount(v, NEW_VALUE)
	if err != nil {
		return traceErr, err
	}

	trace.OldOpening, err = TryParseLeafOpening(v, PRIOR_UPDATED_LEAF)
	if err != nil {
		return traceErr, err
	}

	return trace, nil
}
