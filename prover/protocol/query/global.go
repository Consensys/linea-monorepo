package query

import (
	"fmt"
	"math"
	"reflect"

	sv "github.com/consensys/accelerated-crypto-monorepo/maths/common/smartvectors"
	"github.com/consensys/accelerated-crypto-monorepo/maths/fft"
	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/coin"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/column"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/ifaces"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/variables"
	"github.com/consensys/accelerated-crypto-monorepo/symbolic"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
	"github.com/consensys/accelerated-crypto-monorepo/utils/gnarkutil"
	"github.com/consensys/gnark/frontend"
	"github.com/sirupsen/logrus"
)

/*
A global constraint is an arithmetic constraint that is applied on several vectors
For instance A[i - 1] * B[i] = A[i] for all i \in 0..1000. The expression can also
use random coins as variables.
*/
type GlobalConstraint struct {
	/*
		Symbolic expression representing the global constraint
	*/
	*symbolic.Expression
	/*
		ifaces.QueryID of the coinstraint
	*/
	ID ifaces.QueryID
	/*
		Domain over which the constraints applies
	*/
	DomainSize int
	/*
		Optional field: prevents the constraint to be cancelled on the
		bound. This is used in the "permutation" global constraint for
		instance. False by default
	*/
	NoBoundCancel bool
}

/*
Constructor for global constraints

  - getSize is a function that gives the size of a commitment from its ifaces.QueryID.
    It is aimed at abstracting the "CompiledIOP" from this package to avoid
    circular dependencies. It used mainly for validating the constraint and
    computing the size of the domain on which the constraint applies.
*/
func NewGlobalConstraint(id ifaces.QueryID, expr *symbolic.Expression, noBoundCancel ...bool) GlobalConstraint {

	/*
		Sanity-check : the querie's ifaces.QueryID cannot be empty or nil
	*/
	if len(id) <= 0 {
		utils.Panic("Given an empty ifaces.QueryID for global constraint query")
	}

	expr.AssertValid()

	res := GlobalConstraint{
		Expression: expr,
		ID:         id,
	}

	if len(noBoundCancel) > 0 {
		utils.Require(len(noBoundCancel) == 1, "there should be only 2 bound cancel got %v", len(noBoundCancel))
		res.NoBoundCancel = noBoundCancel[0]
	}

	// performs validation for the
	res.DomainSize = res.validatedDomainSize()
	return res
}

/*
Test a polynomial identity relation
*/
func (cs GlobalConstraint) Check(run ifaces.Runtime) error {

	logrus.Debugf("checking global : %v\n", cs.ID)

	boarded := cs.Board()
	metadatas := boarded.ListVariableMetadata()

	/*
		Sanity-check : All witnesses should have a size at least
		larger than end.
	*/
	for _, metadataInterface := range metadatas {
		if handle, ok := metadataInterface.(ifaces.Column); ok {
			witness := handle.GetColAssignment(run)
			if witness.Len() != cs.DomainSize {
				utils.Panic(
					"Query %v - Witness of %v has size %v  which is below %v",
					cs.ID, handle.GetColID(), witness.Len(), cs.DomainSize,
				)
			}
		}
	}

	/*
		Collects the relevant datas into a slice for the evaluation
	*/
	evalInputs := make([]sv.SmartVector, len(metadatas))

	/*
		Omega is a root of unity which generates the domain of evaluation
		of the constraint. Its size coincide with the size of the domain
		of evaluation. For each value of `i`, X will evaluate to omega^i.
	*/
	omega := fft.GetOmega(cs.DomainSize)
	omegaI := field.One()

	// precomputations of the powers of omega, can be optimized if useful
	omegas := make([]field.Element, cs.DomainSize)
	for i := 0; i < cs.DomainSize; i++ {
		omegas[i] = omegaI
		omegaI.Mul(&omegaI, &omega)
	}

	/*
		Collect the relevants inputs for evaluating the constraint
	*/
	for k, metadataInterface := range metadatas {
		switch meta := metadataInterface.(type) {
		case ifaces.Column:
			w := meta.GetColAssignment(run)
			evalInputs[k] = w
		case coin.Info:
			evalInputs[k] = sv.NewConstant(run.GetRandomCoinField(meta.Name), cs.DomainSize)
		case variables.X:
			evalInputs[k] = meta.EvalCoset(cs.DomainSize, 0, 1, false)
		case variables.PeriodicSample:
			evalInputs[k] = meta.EvalCoset(cs.DomainSize, 0, 1, false)
		case *ifaces.Accessor:
			evalInputs[k] = sv.NewConstant(meta.GetVal(run), cs.DomainSize)
		default:
			utils.Panic("Not a variable type %v in query %v", reflect.TypeOf(metadataInterface), cs.ID)
		}
	}

	// This panics if the global constraints doesn't use any commitment
	res := boarded.Evaluate(evalInputs)

	offsetRange := cs.MinMaxOffset()

	start, stop := 0, res.Len()
	if !cs.NoBoundCancel {
		start -= offsetRange.Min
		stop -= offsetRange.Max
	}

	for i := start; i < stop; i++ {

		// if i < 7 {
		// 	input := make([]field.Element, len(evalInputs))
		// 	for k := range input {
		// 		input[k] = evalInputs[k].Get(i)
		// 	}
		// 	fmt.Printf("(%v) inputs for position %v, %v\n", cs.ID, i, vector.Prettify(input))
		// }

		resx := res.Get(i)
		// The proper test
		if !resx.IsZero() {
			debugMap := make(map[string]string)

			for k, metadataInterface := range metadatas {
				inpx := evalInputs[k].Get(i)
				debugMap[string(metadataInterface.String())] = fmt.Sprintf("%v", inpx.String())
			}

			return fmt.Errorf("the global constraint check failed at row %v \n\tinput details : %v \n\tres: %v\n\t", i, debugMap, resx.String())
		}
	}

	// Update the value of omega^i
	omegaI.Mul(&omegaI, &omega)

	// Nil indicate the test passes
	return nil
}

/*
Scan the metadatas, performs validation and return the domainSize of the constraint.
Panic if there is a problem
*/
func (cs *GlobalConstraint) validatedDomainSize() int {

	boarded := cs.Board()
	metadatas := boarded.ListVariableMetadata()

	/*
		Flag assessing is any `Commitment` metadata was found. If None was found,
		we emit a panic as it is likely a non-usecase.
	*/
	foundAny := false
	domainSize := 0

	// From the min/max offset
	for _, metadataInterface := range metadatas {
		if handle, ok := metadataInterface.(ifaces.Column); ok {
			// All domains should be the same length
			if !foundAny {
				// Mark that we found at least one commitment
				foundAny = true
				/*
					We take the first commitment as a reference size, that
					all the other must have
				*/
				domainSize = handle.Size()
			}

			// validation of the metadata
			if handle.Size() != domainSize {
				utils.Panic("found a commitment %v with domain size %v, but also found domainSize %v", handle.GetColID(), handle.Size(), domainSize)
			}
			if handle.Size() == 0 {
				utils.Panic("size 0 is forbidden (%v)", handle.GetColID())
			}
		}
	}

	// that would be a global constraint on nothing
	if !foundAny {
		utils.Panic("query %v - Could not find any commitment in the metadatas: %v", cs.ID, metadatas)
	}

	return domainSize
}

// Returns the min and max offset happening in the expression
func (cs *GlobalConstraint) MinMaxOffset() utils.Range {

	minOffset := math.MaxInt
	maxOffset := math.MinInt

	/*
		Flag detecting if we indeed found at least one correct metadata
		There for sanity-checks
	*/
	foundAny := false

	exprBoard := cs.Expression.Board()

	for _, metadataUncasted := range exprBoard.ListVariableMetadata() {
		if handle, ok := metadataUncasted.(ifaces.Column); ok {
			foundAny = true

			offset := column.StackOffsets(handle)
			minOffset = utils.Min(minOffset, offset)
			maxOffset = utils.Max(maxOffset, offset)
		}
	}

	if !foundAny {
		panic("did not find any")
	}

	res := utils.Range{Min: minOffset, Max: maxOffset}

	return res
}

/*
Test a polynomial identity relation
*/
func (cs GlobalConstraint) CheckGnark(api frontend.API, run ifaces.GnarkRuntime) {
	boarded := cs.Board()
	metadatas := boarded.ListVariableMetadata()

	/*
		Sanity-check : All witnesses should have a size at least
		larger than end.
	*/
	for _, metadataInterface := range metadatas {
		if handle, ok := metadataInterface.(ifaces.Column); ok {
			witness := handle.GetColAssignmentGnark(run)
			if len(witness) != cs.DomainSize {
				utils.Panic(
					"Query %v - Witness of %v has size %v which is below %v",
					cs.ID, handle.GetColID(), len(witness), cs.DomainSize,
				)
			}
		}
	}

	/*
		Collects the relevant datas into a slice for the evaluation
	*/
	evalInputs := make([][]frontend.Variable, len(metadatas))

	/*
		Omega is a root of unity which generates the domain of evaluation
		of the constraint. Its size coincide with the size of the domain
		of evaluation. For each value of `i`, X will evaluate to omega^i.
	*/
	omega := fft.GetOmega(cs.DomainSize)
	omegaI := field.One()

	// precomputations of the powers of omega, can be optimized if useful
	omegas := make([]frontend.Variable, cs.DomainSize)
	for i := 0; i < cs.DomainSize; i++ {
		omegas[i] = omegaI
		omegaI.Mul(&omegaI, &omega)
	}

	/*
		Collect the relevants inputs for evaluating the constraint
	*/
	for k, metadataInterface := range metadatas {
		switch meta := metadataInterface.(type) {
		case ifaces.Column:
			w := meta.GetColAssignmentGnark(run)
			evalInputs[k] = w
		case coin.Info:
			evalInputs[k] = gnarkutil.RepeatedVariable(run.GetRandomCoinField(meta.Name), cs.DomainSize)
		case variables.X:
			evalInputs[k] = meta.GnarkEvalNoCoset(cs.DomainSize)
		case variables.PeriodicSample:
			evalInputs[k] = meta.GnarkEvalNoCoset(cs.DomainSize)
		case *ifaces.Accessor:
			evalInputs[k] = gnarkutil.RepeatedVariable(meta.GetFrontendVariable(api, run), cs.DomainSize)
		default:
			utils.Panic("Not a variable type %v in query %v", reflect.TypeOf(metadataInterface), cs.ID)
		}
	}

	offsetRange := cs.MinMaxOffset()

	start, stop := 0, cs.DomainSize
	if !cs.NoBoundCancel {
		start -= offsetRange.Min
		stop -= offsetRange.Max
	}

	for i := start; i < stop; i++ {
		// This panics if the global constraints doesn't use any commitment
		inputs := make([]frontend.Variable, len(evalInputs))
		for j := range inputs {
			inputs[j] = evalInputs[j][i]
		}
		res := boarded.GnarkEval(api, inputs)

		// if i < 7 {
		// 	toPrint := []frontend.Variable{cs.ID, "pos=", i, "values="}
		// 	toPrint = append(toPrint, inputs...)
		// 	toPrint = append(toPrint, "res", res)
		// 	api.Println(toPrint...)
		// }
		api.AssertIsEqual(res, 0)
	}

	// Update the value of omega^i
	omegaI.Mul(&omegaI, &omega)
}
