package json

import (
	"github.com/consensys/accelerated-crypto-monorepo/backend/jsonutil"
	"github.com/consensys/accelerated-crypto-monorepo/backend/statemanager/eth"
	"github.com/valyala/fastjson"
)

/*
	Schema :

	location
	nextFreeNode
	subRoot
	leftLeaf
	rightLeaf
	leftProof
	rightProof
	key
	type
*/

// Parse a read zero trace for storage trie
func ParseReadZeroTraceST(v fastjson.Value) (eth.ReadZeroTraceST, error) {

	trace := eth.ReadZeroTraceST{}
	traceErr := eth.ReadZeroTraceST{}

	var err error

	trace.Key, err = jsonutil.TryGetBytes32(v, KEY)
	if err != nil {
		return traceErr, err
	}

	trace.Location, err = jsonutil.TryGetString(v, LOCATION)
	if err != nil {
		return traceErr, err
	}

	trace.NextFreeNode, err = jsonutil.TryGetInt(v, NEXT_FREE_NODE)
	if err != nil {
		return traceErr, err
	}

	trace.OpeningMinus, err = TryParseLeafOpening(v, LEFT_LEAF)
	if err != nil {
		return traceErr, err
	}

	trace.OpeningPlus, err = TryParseLeafOpening(v, RIGHT_LEAF)
	if err != nil {
		return traceErr, err
	}

	trace.ProofMinus, err = TryParseMerkleProof(v, LEFT_PROOF)
	if err != nil {
		return traceErr, err
	}

	trace.ProofPlus, err = TryParseMerkleProof(v, RIGHT_PROOF)
	if err != nil {
		return traceErr, err
	}

	trace.SubRoot, err = jsonutil.TryGetDigest(v, SUBROOT)
	if err != nil {
		return traceErr, err
	}

	return trace, nil
}

// Parse a read zero trace for the world state
func ParseReadZeroTraceWS(v fastjson.Value) (eth.ReadZeroTraceWS, error) {

	trace := eth.ReadZeroTraceWS{}
	traceErr := eth.ReadZeroTraceWS{}

	var err error

	trace.Key, err = jsonutil.TryGetAddress(v, KEY)
	if err != nil {
		return traceErr, err
	}

	trace.Location, err = jsonutil.TryGetString(v, LOCATION)
	if err != nil {
		return traceErr, err
	}

	trace.NextFreeNode, err = jsonutil.TryGetInt(v, NEXT_FREE_NODE)
	if err != nil {
		return traceErr, err
	}

	trace.OpeningMinus, err = TryParseLeafOpening(v, LEFT_LEAF)
	if err != nil {
		return traceErr, err
	}

	trace.OpeningPlus, err = TryParseLeafOpening(v, RIGHT_LEAF)
	if err != nil {
		return traceErr, err
	}

	trace.ProofMinus, err = TryParseMerkleProof(v, LEFT_PROOF)
	if err != nil {
		return traceErr, err
	}

	trace.ProofPlus, err = TryParseMerkleProof(v, RIGHT_PROOF)
	if err != nil {
		return traceErr, err
	}

	trace.SubRoot, err = jsonutil.TryGetDigest(v, SUBROOT)
	if err != nil {
		return traceErr, err
	}

	return trace, nil
}
