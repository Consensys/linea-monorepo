package json

import (
	"github.com/consensys/accelerated-crypto-monorepo/backend/jsonutil"
	"github.com/consensys/accelerated-crypto-monorepo/backend/statemanager/eth"
	"github.com/sirupsen/logrus"
	"github.com/valyala/fastjson"
)

/*
	Schema

	location
    newNextFreeNode
    oldSubRoot
    newSubRoot
    leftProof
    deletedProof
    rightProof
    key
    priorLeftLeaf
    priorDeletedLeaf
    priorRightLeaf
*/

// Parse a deletion trace for a storage trie
func ParseDeletionTraceST(v fastjson.Value) (eth.DeletionTraceST, error) {

	trace, traceErr := eth.DeletionTraceST{}, eth.DeletionTraceST{}
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

	trace.ProofDeleted, err = TryParseMerkleProof(v, DELETED_PROOF)
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

	trace.OldOpenMinus, err = TryParseLeafOpening(v, PRIOR_LEFT_LEAF)
	if err != nil {
		return traceErr, err
	}

	trace.DeletedOpen, err = TryParseLeafOpening(v, PRIOR_DELETED_LEAF)
	if err != nil {
		return traceErr, err
	}

	trace.OldOpenPlus, err = TryParseLeafOpening(v, PRIOR_RIGHT_LEAF)
	if err != nil {
		return traceErr, err
	}

	trace.DeletedValue, err = jsonutil.TryGetBytes32(v, DELETED_VALUE)
	if err != nil {
		logrus.Errorf("could not parse `%v`, tolerated but not ok", DELETED_VALUE)
	}

	return trace, nil
}

// Parse a deletion trace for a storage trie
func ParseDeletionTraceWS(v fastjson.Value) (eth.DeletionTraceWS, error) {

	trace, traceErr := eth.DeletionTraceWS{}, eth.DeletionTraceWS{}
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

	trace.ProofDeleted, err = TryParseMerkleProof(v, DELETED_PROOF)
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

	trace.OldOpenMinus, err = TryParseLeafOpening(v, PRIOR_LEFT_LEAF)
	if err != nil {
		return traceErr, err
	}

	trace.DeletedOpen, err = TryParseLeafOpening(v, PRIOR_DELETED_LEAF)
	if err != nil {
		return traceErr, err
	}

	trace.OldOpenPlus, err = TryParseLeafOpening(v, PRIOR_RIGHT_LEAF)
	if err != nil {
		return traceErr, err
	}

	trace.DeletedValue, err = tryParseAccount(v, DELETED_VALUE)
	if err != nil {
		logrus.Errorf("could not parse `%v`, tolerated but not ok", DELETED_VALUE)
	}

	return trace, nil
}
