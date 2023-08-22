package json

import (
	"fmt"

	"github.com/consensys/accelerated-crypto-monorepo/backend/jsonutil"
	"github.com/consensys/accelerated-crypto-monorepo/backend/statemanager/eth"
	"github.com/sirupsen/logrus"
	"github.com/valyala/fastjson"
)

// json field names
const (
	TYPE               string = "type"
	LOCATION           string = "location"
	NEXT_FREE_NODE     string = "nextFreeNode"
	NEW_NEXT_FREE_NODE string = "newNextFreeNode"
	SUBROOT            string = "subRoot"
	LEAF               string = "leaf"
	PROOF              string = "proof"
	KEY                string = "key"
	VALUE              string = "value"
	OLD_VALUE          string = "oldValue"
	NEW_VALUE          string = "newValue"
	DELETED_VALUE      string = "deletedValue"
	OLD_SUBROOT        string = "oldSubRoot"
	NEW_SUBROOT        string = "newSubRoot"
	LEFT_PROOF         string = "leftProof"
	DELETED_PROOF      string = "deletedProof"
	NEW_PROOF          string = "newProof"
	RIGHT_PROOF        string = "rightProof"
	LEFT_LEAF          string = "leftLeaf"
	RIGHT_LEAF         string = "rightLeaf"
	PRIOR_LEFT_LEAF    string = "priorLeftLeaf"
	PRIOR_DELETED_LEAF string = "priorDeletedLeaf"
	PRIOR_RIGHT_LEAF   string = "priorRightLeaf"
	PRIOR_UPDATED_LEAF string = "priorUpdatedLeaf"
)

// Code to identify the type of an update
const (
	READ_TRACE_CODE      int = 0
	READ_ZERO_TRACE_CODE int = 1
	INSERTION_TRACE_CODE int = 2
	UPDATE_TRACE_CODE    int = 3
	DELETION_TRACE_CODE  int = 4
)

// Parse state-manager output
func ParseStateManagerTraces(v fastjson.Value) (traces [][]any, err error) {

	// Attempt to parse the JSON and panic on failure
	blocks, err := jsonutil.TryGetArray(v)
	if err != nil {
		return nil, err
	}

	if len(blocks) == 0 {
		return nil, fmt.Errorf("did not find any blocks in the traces")
	}

	logrus.Infof("state-manager inspector : got %v blocks", len(blocks))

	// And parse it block by block
	res := make([][]any, len(blocks))
	for i, block := range blocks {
		res[i], err = ParseBlockTraces(block)
		if err != nil {
			return nil, fmt.Errorf("error parsing the block %v : %v", i, err)
		}

		logrus.Infof("state-manager inspector : got %v traces for block %v", len(res[i]), i)
	}

	return res, nil
}

// Parse a JSON-text string into an array of traces
func ParseBlockTraces(v fastjson.Value) ([]any, error) {

	// Attempt to parse the JSON and panic on failure
	array, err := jsonutil.TryGetArray(v)
	if err != nil {
		return nil, err
	}

	// For each type of the array, we attempt to deserialize
	// it into a string
	parsed := make([]any, len(array))
	for i := range array {

		// To distinguish which type of update, we use the field "type"
		typeCode, err := jsonutil.TryGetInt(array[i], TYPE)
		if err != nil {
			return nil, fmt.Errorf("can't get the trace code : %v (position %v)", err, i)
		}

		// The location help us track to which sub-tree relates the current trace
		location, err := jsonutil.TryGetString(array[i], LOCATION)
		if err != nil {
			return nil, fmt.Errorf("can't get the location : %v (position %v)", err, i)
		}

		switch typeCode {
		case READ_TRACE_CODE:
			switch location {
			case eth.WS_LOCATION:
				parsed[i], err = ParseReadNonZeroTraceWS(array[i])
			default:
				parsed[i], err = ParseReadNonZeroTraceST(array[i])
			}
		case READ_ZERO_TRACE_CODE:
			switch location {
			case eth.WS_LOCATION:
				parsed[i], err = ParseReadZeroTraceWS(array[i])
			default:
				parsed[i], err = ParseReadZeroTraceST(array[i])
			}
		case INSERTION_TRACE_CODE:
			switch location {
			case eth.WS_LOCATION:
				parsed[i], err = ParseInsertionTraceWS(array[i])
			default:
				parsed[i], err = ParseInsertionTraceST(array[i])
			}
		case UPDATE_TRACE_CODE:
			switch location {
			case eth.WS_LOCATION:
				parsed[i], err = ParseUpdateTraceWS(array[i])
			default:
				parsed[i], err = ParseUpdateTraceST(array[i])
			}
		case DELETION_TRACE_CODE:
			switch location {
			case eth.WS_LOCATION:
				parsed[i], err = ParseDeletionTraceWS(array[i])
			default:
				parsed[i], err = ParseDeletionTraceST(array[i])
			}
		default:
			return nil, fmt.Errorf("unknown trace code : %v", typeCode)
		}

		// Finally check the error
		if err != nil {
			return nil, fmt.Errorf("could not parse the trace : %v (position %v = %v)", err, i, array[i].String())
		}
	}

	return parsed, nil
}
