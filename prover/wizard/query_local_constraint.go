package wizard

import (
	"fmt"
	"reflect"

	"github.com/consensys/gnark/frontend"
	sv "github.com/consensys/zkevm-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/zkevm-monorepo/prover/symbolic"
	"github.com/consensys/zkevm-monorepo/prover/utils"
)

// A local constraint is an arithmetic relation between prespecified
// polynomial commitment openings and random coin.
// the local constraint is evaluated at 0
// in order to obtain evaluations at different points, the vector should be shifted first
// and the constraint applied after
type QueryLocalConstraint struct {
	*symbolic.Expression
	metadata *metadata
	*subQuery
}

// Construct a new local constraint
func (api *API) NewQueryLocalConstraint(expr *symbolic.Expression) QueryLocalConstraint {

	expr.AssertValid()

	var (
		e     = NewColExpression(expr)
		round = e.Round()
		_     = e.Size()
		res   = QueryLocalConstraint{
			Expression: expr,
			subQuery: &subQuery{
				round: round,
			},
			metadata: api.newMetadata(),
		}
	)

	api.queries.addToRound(round, &res)
	return res
}

// Test the polynomial identity
func (cs QueryLocalConstraint) Check(run Runtime) error {

	var (
		board     = cs.Board()
		metadatas = board.ListVariableMetadata()
		inputs    = make([]sv.SmartVector, len(metadatas))
	)

	for i, metadataInterface := range metadatas {
		switch metadata := metadataInterface.(type) {
		case Column:
			// The offsets are already tested to be in range
			// 	- should be between 0 and N-1 (included) where N is the size of
			// 		the polynomials
			val := metadata.GetAssignment(run).Get(0)
			inputs[i] = sv.NewConstant(val, 1)
		case Accessor:
			inputs[i] = sv.NewConstant(metadata.GetVal(run), 1)
		default:
			utils.Panic("Unknown variable type %v in local constraint %v", reflect.TypeOf(metadataInterface), cs.String())
		}
	}

	res := board.Evaluate(inputs).(*sv.Constant).Val()

	if !res.IsZero() {
		debugMap := make(map[string]string)

		for k, metadataInterface := range metadatas {
			inpx := inputs[k].Get(0)
			debugMap[string(metadataInterface.String())] = fmt.Sprintf("%v", inpx.String())
		}

		return fmt.Errorf("the local constraint %v check failed \n\tinput details : %v \n\tres: %v\n\t", cs.String(), debugMap, res.String())
	}

	return nil
}

// Test the polynomial identity in a circuit setting
func (cs QueryLocalConstraint) CheckGnark(api frontend.API, run RuntimeGnark) {

	var (
		board     = cs.Board()
		metadatas = board.ListVariableMetadata()
		inputs    = make([]frontend.Variable, len(metadatas))
	)

	for i, metadataInterface := range metadatas {
		switch metadata := metadataInterface.(type) {
		case Column:
			/*
				The offsets are already tested to be in range
					- should be between 0 and N-1 (included)
						where N is the size of the polynomials
			*/
			val := metadata.GetAssignmentGnark(api, run)[0]
			inputs[i] = val
		case Accessor:
			inputs[i] = metadata.GetValGnark(api, run)
		default:
			utils.Panic("Unknown variable type %v in local constraint %v", reflect.TypeOf(metadataInterface), cs.String())
		}
	}

	// Sanity-check : n (the number of element used for the evaluation)
	// should be equal to the length of metadata
	res := board.GnarkEval(api, inputs)
	api.AssertIsEqual(res, 0)
}
