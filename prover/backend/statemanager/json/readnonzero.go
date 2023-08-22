package json

import (
	"github.com/consensys/accelerated-crypto-monorepo/backend/jsonutil"
	"github.com/consensys/accelerated-crypto-monorepo/backend/statemanager/eth"
	"github.com/valyala/fastjson"
)

// Parse a read non zero trace for the storage trie
func ParseReadNonZeroTraceST(v fastjson.Value) (eth.ReadNonZeroTraceST, error) {

	trace, traceErr := eth.ReadNonZeroTraceST{}, eth.ReadNonZeroTraceST{}

	var err error
	trace.NextFreeNode, err = jsonutil.TryGetInt(v, NEXT_FREE_NODE)
	if err != nil {
		return traceErr, err
	}

	trace.LeafOpening, err = TryParseLeafOpening(v, LEAF)
	if err != nil {
		return traceErr, err
	}

	trace.SubRoot, err = jsonutil.TryGetDigest(v, SUBROOT)
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

	trace.Value, err = jsonutil.TryGetBytes32(v, VALUE)
	if err != nil {
		return traceErr, err
	}

	trace.Location, err = jsonutil.TryGetString(v, LOCATION)
	if err != nil {
		return traceErr, err
	}

	return trace, nil
}

// Parse a read non zero trace for the world state
func ParseReadNonZeroTraceWS(v fastjson.Value) (eth.ReadNonZeroTraceWS, error) {

	trace, traceErr := eth.ReadNonZeroTraceWS{}, eth.ReadNonZeroTraceWS{}

	var err error
	trace.NextFreeNode, err = jsonutil.TryGetInt(v, NEXT_FREE_NODE)
	if err != nil {
		return traceErr, err
	}

	trace.LeafOpening, err = TryParseLeafOpening(v, LEAF)
	if err != nil {
		return traceErr, err
	}

	trace.SubRoot, err = jsonutil.TryGetDigest(v, SUBROOT)
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

	trace.Value, err = tryParseAccount(v, VALUE)
	if err != nil {
		return traceErr, err
	}

	trace.Location, err = jsonutil.TryGetString(v, LOCATION)
	if err != nil {
		return traceErr, err
	}

	return trace, nil
}
