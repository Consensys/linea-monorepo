package json

import (
	"github.com/consensys/accelerated-crypto-monorepo/backend/jsonutil"
	"github.com/consensys/accelerated-crypto-monorepo/backend/statemanager/eth"
	"github.com/valyala/fastjson"
)

/*
	Schema

	location
	newNextFreeNode
	oldSubRoot
	newSubRoot
	leftProof
	newProof
	rightProof
	key
	value
	priorLeftLeaf
    priorRightLeaf
  }
*/

// Parse an insertion trace in a storage trie
func ParseInsertionTraceST(v fastjson.Value) (eth.InsertionTraceST, error) {

	trace, traceErr := eth.InsertionTraceST{}, eth.InsertionTraceST{}
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

	trace.ProofMinus, err = TryParseMerkleProof(v, LEFT_PROOF)
	if err != nil {
		return traceErr, err
	}

	trace.ProofNew, err = TryParseMerkleProof(v, NEW_PROOF)
	if err != nil {
		return traceErr, err
	}

	trace.ProofPlus, err = TryParseMerkleProof(v, RIGHT_PROOF)
	if err != nil {
		return traceErr, err
	}

	trace.Key, err = jsonutil.TryGetBytes32(v, KEY)
	if err != nil {
		return traceErr, err
	}

	trace.Val, err = jsonutil.TryGetBytes32(v, VALUE)
	if err != nil {
		return traceErr, err
	}

	trace.OldOpenMinus, err = TryParseLeafOpening(v, PRIOR_LEFT_LEAF)
	if err != nil {
		return traceErr, err
	}

	trace.OldOpenPlus, err = TryParseLeafOpening(v, PRIOR_RIGHT_LEAF)
	if err != nil {
		return traceErr, err
	}

	return trace, nil
}

// Parse an insertion trace in a storage trie
func ParseInsertionTraceWS(v fastjson.Value) (eth.InsertionTraceWS, error) {

	trace, traceErr := eth.InsertionTraceWS{}, eth.InsertionTraceWS{}
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

	trace.ProofMinus, err = TryParseMerkleProof(v, LEFT_PROOF)
	if err != nil {
		return traceErr, err
	}

	trace.ProofNew, err = TryParseMerkleProof(v, NEW_PROOF)
	if err != nil {
		return traceErr, err
	}

	trace.ProofPlus, err = TryParseMerkleProof(v, RIGHT_PROOF)
	if err != nil {
		return traceErr, err
	}

	trace.Key, err = jsonutil.TryGetAddress(v, KEY)
	if err != nil {
		return traceErr, err
	}

	trace.Val, err = tryParseAccount(v, VALUE)
	if err != nil {
		return traceErr, err
	}

	trace.OldOpenMinus, err = TryParseLeafOpening(v, PRIOR_LEFT_LEAF)
	if err != nil {
		return traceErr, err
	}

	trace.OldOpenPlus, err = TryParseLeafOpening(v, PRIOR_RIGHT_LEAF)
	if err != nil {
		return traceErr, err
	}

	return trace, nil
}
