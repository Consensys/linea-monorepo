package query

import (
	"fmt"
	"reflect"

	sv "github.com/consensys/accelerated-crypto-monorepo/maths/common/smartvectors"
	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/coin"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/ifaces"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/variables"
	"github.com/consensys/accelerated-crypto-monorepo/symbolic"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
	"github.com/consensys/gnark/frontend"
)

// A local constraint is an arithmetic relation between prespecified
// polynomial commitment openings and random coin.
type LocalConstraint struct {
	*symbolic.Expression
	ID ifaces.QueryID
}

// Construct a new local constraint
// getSize is a handle (intended to be an accessor CompiledIOP) that allows
// getting the size of a commitment without creating import cycles.
func NewLocalConstraint(id ifaces.QueryID, expr *symbolic.Expression) LocalConstraint {

	if len(id) == 0 {
		panic("Local constraint with an empty ID")
	}

	expr.AssertValid()

	domainSize := 0
	foundAny := false

	/*
		Upon construct, checks that all the offsets are compatible with
		the domain size and sizes
	*/
	board := expr.Board()
	metadatas := board.ListVariableMetadata()
	for _, metadataInterface := range metadatas {
		if metadata, ok := metadataInterface.(ifaces.Column); ok {
			if !foundAny {
				foundAny = true
				domainSize = metadata.Size()
			}
			if metadata.Size() != domainSize {
				utils.Panic("Unsupported : Local constraints with heterogeneous domain size")
			}
		}
	}

	if !foundAny {
		utils.Panic("No commitment found in the local constraint")
	}

	if domainSize == 0 {
		utils.Panic("All commitment given had a length of zero")
	}

	res := LocalConstraint{Expression: expr, ID: id}
	return res
}

// Test the polynomial identity
func (cs LocalConstraint) Check(run ifaces.Runtime) error {
	board := cs.Board()
	metadatas := board.ListVariableMetadata()
	/*
		Collects the relevant datas into a slice for the evaluation
	*/
	inputs := make([]sv.SmartVector, len(metadatas))
	for i, metadataInterface := range metadatas {
		switch metadata := metadataInterface.(type) {
		case ifaces.Column:
			/*
				The offsets are already tested to be in range
					- should be between 0 and N-1 (included)
						where N is the size of the polynomials
			*/
			val := metadata.GetColAssignmentAt(run, 0)
			inputs[i] = sv.NewConstant(val, 1)
		case coin.Info:
			inputs[i] = sv.NewConstant(run.GetRandomCoinField(metadata.Name), 1)
		case variables.PeriodicSample:
			v := field.One()
			if metadata.Offset != 0 {
				v.SetZero()
			}
			inputs[i] = sv.NewConstant(metadata.EvalAtOnDomain(0), 1)
		case variables.X:
			utils.Panic("In local constraint %v, Local constraints using X are not handled so far", cs.ID)
		case *ifaces.Accessor:
			inputs[i] = sv.NewConstant(metadata.GetVal(run), 1)
		default:
			utils.Panic("Unknown variable type %v in local constraint %v", reflect.TypeOf(metadataInterface), cs.ID)
		}
	}
	/*
		Sanity-check : n (the number of element used for the evaluation)
		should be equal to the length of metadata
	*/
	res := board.Evaluate(inputs).(*sv.Constant).Val()
	/*
		If the query is satisfied, the result should be zero. In
		this case, we pretty print an error.
	*/
	if !res.IsZero() {
		debugMap := make(map[string]string)

		for k, metadataInterface := range metadatas {
			inpx := inputs[k].Get(0)
			debugMap[string(metadataInterface.String())] = fmt.Sprintf("%v", inpx.String())
		}

		return fmt.Errorf("the local constraint %v check failed \n\tinput details : %v \n\tres: %v\n\t", cs.ID, debugMap, res.String())
	}
	/*
		Nil to indicate the test passed
	*/
	return nil
}

// Test the polynomial identity in a circuit setting
func (cs LocalConstraint) CheckGnark(api frontend.API, run ifaces.GnarkRuntime) {
	board := cs.Board()
	metadatas := board.ListVariableMetadata()
	/*
		Collects the relevant datas into a slice for the evaluation
	*/
	inputs := make([]frontend.Variable, len(metadatas))
	for i, metadataInterface := range metadatas {
		switch metadata := metadataInterface.(type) {
		case ifaces.Column:
			/*
				The offsets are already tested to be in range
					- should be between 0 and N-1 (included)
						where N is the size of the polynomials
			*/
			val := metadata.GetColAssignmentGnarkAt(run, 0)
			inputs[i] = val
		case coin.Info:
			inputs[i] = run.GetRandomCoinField(metadata.Name)
		case variables.X, variables.PeriodicSample:
			utils.Panic("In local constraint %v, Local constraints using X are not handled so far", cs.ID)
		case *ifaces.Accessor:
			inputs[i] = metadata.GetFrontendVariable(api, run)
		default:
			utils.Panic("Unknown variable type %v in local constraint %v", reflect.TypeOf(metadataInterface), cs.ID)
		}
	}
	/*
		Sanity-check : n (the number of element used for the evaluation)
		should be equal to the length of metadata
	*/
	res := board.GnarkEval(api, inputs)
	api.AssertIsEqual(res, 0)
}
