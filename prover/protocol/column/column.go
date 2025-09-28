package column

import (
	"reflect"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/variables"
	"github.com/consensys/linea-monorepo/prover/protocol/zk"
	"github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
)

const (
	shift       string = "SHIFT"
	interleaved string = "INTERLEAVED"
	// Generalizes the concept of Natural[T] to
	// the case of verifier defined columns
	nonComposite string = "NONCOMPOSITE"
)

// Constructs a natural column. input validation. Not exported, use
// [store.AddToRound] which will output a well-formed object. This ensures the
// invariant that a `Natural[T]` always have a store.
func newNatural[T zk.Element](name ifaces.ColID, position columnPosition, store *Store[T]) Natural[T] {
	if len(name) == 0 {
		utils.Panic("empty name")
	}
	if store == nil {
		utils.Panic("null store (%v)", name)
	}
	return Natural[T]{ID: name, position: position, store: store}
}

// RootParents returns the underlying base [Natural[T]] of the current handle.
func RootParents[T zk.Element](h ifaces.Column[T]) ifaces.Column[T] {

	if !h.IsComposite() {
		return h
	}

	switch inner := h.(type) {
	case Natural[T]:
		// No changes
		return h
	case Shifted[T]:
		return RootParents(inner.Parent)
	default:
		utils.Panic("unexpected type %v", reflect.TypeOf(inner))
	}
	panic("unreachable")
}

// RootsOf returns a deduplicated list of [column.Natural[T]] (or
// [verifiercol.VerifierCol]] underlying the definition of the input columns.
// If `skipVCol` is true, the function will skip the verifiercol and not
// the [column.Natural[T]] columns. The function also skips nil columns.
func RootsOf[T zk.Element](cols []ifaces.Column[T], skipVCol bool) []ifaces.Column[T] {

	var (
		roots    = make([]ifaces.Column[T], 0, len(cols))
		rootSets = make(map[ifaces.ColID]struct{}, len(cols))
	)

	for i := range cols {

		if cols[i] == nil {
			continue
		}

		root := RootParents(cols[i])

		// This clause checks if the root is a verifiercol and if it
		// should be skipped following "skipVCol". However, testing this
		// directly would create a circular dependency. So we instead use
		// [IsComposite] == false and hasType([column.Natural[T]]) == true
		// to perform the test.
		_, isNat := root.(Natural[T])
		if !isNat && !root.IsComposite() && skipVCol {
			continue
		}

		if _, ok := rootSets[root.GetColID()]; ok {
			continue
		}

		roots = append(roots, root)
		rootSets[root.GetColID()] = struct{}{}
	}

	return roots
}

// StackOffset sums all the offsets contained in the handle and return the result
func StackOffsets[T zk.Element](h ifaces.Column[T]) int {

	if !h.IsComposite() {
		return 0
	}

	switch inner := h.(type) {
	case Natural[T]:
		// No changes
		return 0
	case Shifted[T]:
		return inner.Offset + StackOffsets(inner.Parent)
	default:
		utils.Panic("unexpected type %v", reflect.TypeOf(inner))
	}
	panic("unreachable")
}

// NbLeaves returns the number of underlying [Natural[T]] columns for `h`. If `h`
// is neither an [Interleaved] nor derived from an [Interleaved], the function
// returns 1.
func NbLeaves[T zk.Element](h ifaces.Column[T]) int {
	switch inner := h.(type) {
	case Natural[T]:
		// No changes
		return 1
	case Shifted[T]:
		return NbLeaves(inner.Parent)
	default:
		utils.Panic("unexpected type %v", reflect.TypeOf(inner))
	}
	panic("unreachable")
}

// EvalExprColumn resolves an expression to a column assignment. The expression
// must be converted to a board prior to evaluating the expression.
//
//   - If the expression does not uses ifaces.Column[T] as metadata, the function
//     will panic.
//
//   - If the expression contains several columns and they don't contain all
//     have the same size.
func EvalExprColumn[T zk.Element](run ifaces.Runtime, board symbolic.ExpressionBoard[T]) smartvectors.SmartVector {

	var (
		metadata = board.ListVariableMetadata()
		inputs   = make([]smartvectors.SmartVector, len(metadata))
		length   = ExprIsOnSameLengthHandles(&board)
		v        field.Element
	)

	// Attempt to recover the size of the
	for i := range inputs {
		switch m := metadata[i].(type) {
		case ifaces.Column[T]:
			inputs[i] = m.GetColAssignment(run)
		case coin.Info:
			if m.IsBase() {
				v = run.GetRandomCoinField(m.Name)
				inputs[i] = smartvectors.NewConstant(v, length)
			} else {
				vExt := run.GetRandomCoinFieldExt(m.Name)
				inputs[i] = smartvectors.NewConstantExt(vExt, length)
			}

		case ifaces.Accessor[T]:
			if m.IsBase() {
				v, _ := m.GetValBase(run)
				inputs[i] = smartvectors.NewConstant(v, length)
			} else {
				v := m.GetValExt(run)
				inputs[i] = smartvectors.NewConstantExt(v, length)
			}

		case variables.PeriodicSample[T]:
			v := m.EvalCoset(length, 0, 1, false)
			inputs[i] = v
		case variables.X[T]:
			v := m.EvalCoset(length, 0, 1, false)
			inputs[i] = v
		}
	}

	return board.Evaluate(inputs)
}

// GnarkEvalExprColumn evaluates an expression in a gnark circuit setting
func GnarkEvalExprColumn[T zk.Element](api frontend.API, run ifaces.GnarkRuntime[T], board symbolic.ExpressionBoard[T]) []T {

	var (
		metadata = board.ListVariableMetadata()
		length   = ExprIsOnSameLengthHandles(&board)
		res      = make([]T, length)
	)

	apiGen, err := zk.NewApi[T](api)
	if err != nil {
		panic(err)
	}

	for k := 0; k < length; k++ {

		inputs := make([]T, len(metadata))

		for i := range inputs {
			switch m := metadata[i].(type) {
			case ifaces.Column[T]:
				inputs[i] = m.GetColAssignmentGnarkAt(run, k)
			case coin.Info:
				if m.IsBase() {
					inputs[i] = run.GetRandomCoinField(m.Name)
				} else {
					// inputs[i] = run.GetRandomCoinFieldExt(m.Name)
				}
			case ifaces.Accessor[T]:
				inputs[i] = m.GetFrontendVariable(apiGen, run)
			case variables.PeriodicSample[T]:
				// inputs[i] = m.EvalAtOnDomain(k)
			case variables.X[T]:
				// there is no theoritical problem with this but there are
				// no cases known where this happens so we just don't
				// support it.
				panic("unexpected")
			}
		}

		res[k] = board.GnarkEval(api, inputs)
	}

	return res
}

// ExprIsOnSameLengthHandles checks that all the variables of the expression
// that are [ifaces.Column[T]] have the same size (and panics if it does not), then
// returns the match.
func ExprIsOnSameLengthHandles[T zk.Element](board *symbolic.ExpressionBoard[T]) int {

	var (
		metadatas = board.ListVariableMetadata()
		length    = 0
	)

	for _, m := range metadatas {
		switch metadata := m.(type) {
		case ifaces.Column[T]:
			// Initialize the length with the first commitment
			if length == 0 {
				length = metadata.Size()
			}

			// Sanity-check the vector should all have the same length
			if length != metadata.Size() {
				utils.Panic("Inconsistent length for %v (has size %v, but expected %v)", metadata.GetColID(), metadata.Size(), length)
			}
		// The expression can involve random coins
		case coin.Info, variables.X[T], variables.PeriodicSample[T], ifaces.Accessor[T]:
			// Do nothing
		default:
			utils.Panic("unknown type %T", metadata)
		}
	}

	// No commitment were found in the metadata, thus this call is broken
	if length == 0 {
		utils.Panic("declared a handle from an expression which does not contains any handle")
	}

	return length
}

// return the runtime assignments of a linear combination column
// that is computed on the fly from the columns stored in hs
func RandLinCombColAssignment[T zk.Element](run ifaces.Runtime, coinVal field.Element, hs []ifaces.Column[T]) smartvectors.SmartVector {
	var colTableWit smartvectors.SmartVector
	var witnessCollapsed smartvectors.SmartVector
	x := field.One()
	witnessCollapsed = smartvectors.NewConstant(field.Zero(), hs[0].Size())
	for tableCol := range hs {
		colTableWit = hs[tableCol].GetColAssignment(run)
		witnessCollapsed = smartvectors.Add(witnessCollapsed, smartvectors.Mul(colTableWit, smartvectors.NewConstant(x, hs[0].Size())))
		x.Mul(&x, &coinVal)
	}
	return witnessCollapsed
}

// maximal round of declaration for a list of commitment
func MaxRound[T zk.Element](handles ...ifaces.Column[T]) int {
	res := 0
	for _, handle := range handles {
		res = utils.Max(res, handle.Round())
	}
	return res
}

// ShiftExpr returns a shifted version of the expression. The function will
// panic if called with an expression that uses [variables.PeriodicSampling]
func ShiftExpr[T zk.Element](expr *symbolic.Expression[T], offset int) *symbolic.Expression[T] {

	return expr.ReconstructBottomUp(func(e *symbolic.Expression[T], children []*symbolic.Expression[T]) (new *symbolic.Expression[T]) {

		vari, isVar := e.Operator.(symbolic.Variable[T])
		if !isVar {
			return e.SameWithNewChildren(children)
		}

		if per, isPeriodic := vari.Metadata.(variables.PeriodicSample[T]); isPeriodic {
			return symbolic.NewVariable[T](variables.PeriodicSample[T]{W: per.W, Offset: utils.PositiveMod(per.Offset-offset, per.W)})
		}

		col, isCol := vari.Metadata.(ifaces.Column[T])
		if !isCol {
			return e.SameWithNewChildren(children)
		}

		return symbolic.NewVariable[T](Shift(col, offset))
	})
}

// StatusOf returns the status of the column
func StatusOf[T zk.Element](c ifaces.Column[T]) Status {

	switch col := c.(type) {
	case Natural[T]:
		return col.Status()
	case Shifted[T]:
		return StatusOf(col.Parent)
	default:
		// Expectedly, a verifier col
		return VerifierDefined
	}
}

// IsPublicExpression returns true if the expression is fully public
func IsPublicExpression[T zk.Element](expr *symbolic.Expression[T]) bool {

	var (
		board        = expr.Board()
		meta         = board.ListVariableMetadata()
		numPublic    = 0
		numNonPublic = 0
		statusMap    = map[ifaces.ColID]string{}
	)

	for _, m := range meta {
		if c, ok := m.(ifaces.Column[T]); ok {
			status := StatusOf(c)
			statusMap[c.GetColID()] = status.String()
			if status.IsPublic() {
				numPublic++
			} else {
				numNonPublic++
			}
		}
	}

	if numNonPublic > 0 && numPublic > 0 {
		utils.Panic("the expression is mixed: %v", statusMap)
	}

	return numPublic > 0
}

// ColumnsOfExpression returns the list of all the columns occuring as variables of
// the provided expression.
func ColumnsOfExpression[T zk.Element](expr *symbolic.Expression[T]) []ifaces.Column[T] {

	var (
		board    = expr.Board()
		metadata = board.ListVariableMetadata()
		res      []ifaces.Column[T]
	)

	for i := range metadata {
		if _, ok := metadata[i].(ifaces.Column[T]); ok {
			res = append(res, metadata[i].(ifaces.Column[T]))
		}
	}

	return res
}
