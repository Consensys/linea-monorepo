package globalcs

import (
	"math/big"
	"reflect"

	"github.com/consensys/gnark-crypto/field/koalabear/fft"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/variables"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/protocol/zk"
	"github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
)

// mergingCtx collects all the compilation input, output and artefacts pertaining
// the merging of the global constraints
type mergingCtx[T zk.Element] struct {

	// DomainSize corresponds to the shared domain size of all the uncompiled global
	// constraints. The fact that all the global constraints share the same domain
	// size is a pre-condition for the compiler and is enforced by Split/Stick compilers.
	DomainSize int

	// RatioBuckets stores arranged by "ratio".
	//
	// The "ratio" denotes the ratio
	// between the quotient size for that particular constraint and the domain
	// size of the constraint.
	//
	//
	// All the constraints happening here have the same domain size even when they
	// are in different bucket. The only difference is then the degree and the
	// and the offset of the constraint.
	//
	// The constraints that are stored here are modified w.r.t. the one initially
	// present in the compiledIOP when the compiler starts because they are multiplied
	// by X-1, X-\omega, X-\omega^2 to account for the bound cancellation.
	RatioBuckets map[int][]*symbolic.Expression[T]

	// Ratios stores the list of the "ratios" in the order in which they are
	// encountered in the compiled IOP.
	//
	// The "ratio" denotes the ratio
	// between the quotient size for that particular constraint and the domain
	// size of the constraint.
	Ratios []int
}

// accumulateConstraints scans comp to collect uncompiled global constraints and
// aggregate them into unified global constraints per "ratio".
//
// See [mergingCtx.Ratios] for an explanation for "ratio".
func accumulateConstraints[T zk.Element](comp *wizard.CompiledIOP[T]) (mergingCtx[T], bool) {

	ctx := mergingCtx[T]{
		RatioBuckets: make(map[int][]*symbolic.Expression[T]),
	}

	for _, qName := range comp.QueriesNoParams.AllUnignoredKeys() {
		// Filter only the global constraints
		cs, ok := comp.QueriesNoParams.Data(qName).(query.GlobalConstraint)
		if !ok {
			// Not a global constraint
			continue
		}

		// For the first iteration, the domain size is unset so we need to initialize
		// it. This works because the domain size of a constraint cannot legally
		// be 0.
		if ctx.DomainSize == 0 {
			ctx.DomainSize = cs.DomainSize
		}

		// This enforces the precondition that all the global constraint must
		// share the same domain.
		if cs.DomainSize != ctx.DomainSize {
			utils.Panic("At this point in the compilation process, we expect all constraints to have the same domain")
		}

		// Mark the constraint as ignored, so that it does not get compiled a
		// second time by a sub-sequent round of compilation.
		comp.QueriesNoParams.MarkAsIgnored(qName)
		ctx.registerCs(cs)
	}

	if ctx.DomainSize == 0 {
		// There is no global constraint to compile
		return mergingCtx[T]{}, false
	}

	return ctx, true
}

// aggregateConstraints returns the list of the aggregated constraints
func (ctx *mergingCtx[T]) aggregateConstraints(comp *wizard.CompiledIOP[T]) []*symbolic.Expression[T] {

	var (
		aggregateExpressions = make([]*symbolic.Expression[T], len(ctx.Ratios))
		initialRound         = comp.NumRounds()
		mergingCoin          = comp.InsertCoin(initialRound, coin.Name(deriveName(comp, DEGREE_RANDOMNESS)), coin.FieldExt)
	)

	for i, ratio := range ctx.Ratios {
		aggregateExpressions[i] = symbolic.NewPolyEval(mergingCoin.AsVariable(), ctx.RatioBuckets[ratio])
	}

	return aggregateExpressions
}

// registerCs determines the ratio of a constraint and appends it to the corresponding
// bucket.
func (ctx *mergingCtx[T]) registerCs(cs query.GlobalConstraint[T]) {

	var (
		bndCancelledExpr = getBoundCancelledExpression(cs)
		ratio            = getExprRatio(bndCancelledExpr)
	)

	// Initialize the outer-maps / slices if the entries are not already allocated
	if _, ok := ctx.RatioBuckets[ratio]; !ok {
		ctx.RatioBuckets[ratio] = []*symbolic.Expression[T]{}
		ctx.Ratios = append(ctx.Ratios, ratio)
	}

	ctx.RatioBuckets[ratio] = append(ctx.RatioBuckets[ratio], bndCancelledExpr)
}

// getBoundCancelledExpression computes the "bound cancelled expression" for the
// constraint cs. Namely, the constraints expression is multiplied by terms of the
// form X-\omega^k to cancel the expression at position "k" if required. If the
// constraint uses the "noBoundCancel" feature, then the constraint expression is
// directly returned.
func getBoundCancelledExpression[T zk.Element](cs query.GlobalConstraint[T]) *symbolic.Expression[T] {

	if cs.NoBoundCancel {
		return cs.Expression
	}

	var (
		cancelRange = query.MinMaxOffset(cs.Expression)
		res         = cs.Expression
		domainSize  = cs.DomainSize
		x           = variables.NewXVar[T]()
		omega, _    = fft.Generator(uint64(domainSize))
		// factors is a list of expression to multiply to obtain the return expression. It
		// is initialized with "only" the initial expression and we iteratively add the
		// terms (X-i) to it. At the end, we call [sym.Mul[T]] a single time. This structure
		// is important because it [sym.Mul[T]] operates a sequence of optimization routines
		// that are everytime we call it. In an earlier version, we were calling [sym.Mul[T]]
		// for every factor and this were making the function have a quadratic/cubic runtime.
		factors = make([]any, 0, utils.Abs(cancelRange.Max)+utils.Abs(cancelRange.Min)+1)
	)

	factors = append(factors, res)

	// appendFactor appends an expressions representing $X-\rho^i$ to [factors]
	appendFactor := func(i int) {
		var root field.Element
		root.Exp(omega, big.NewInt(int64(i)))
		factors = append(factors, symbolic.Sub[T](x, root))
	}

	if cancelRange.Min < 0 {
		// Cancels the expression on the range [0, -cancelRange.Min)
		for i := 0; i < -cancelRange.Min; i++ {
			appendFactor(i)
		}
	}

	if cancelRange.Max > 0 {
		// Cancels the expression on the range (N-cancelRange.Max-1, N-1]
		for i := 0; i < cancelRange.Max; i++ {
			point := domainSize - i - 1 // point at which we want to cancel the constraint
			appendFactor(point)
		}
	}

	// When factors is of length 1, it means the expression does not need to be
	// bound-cancelled and we can directly return the original expression
	// without calling [sym.Mul[T]].
	if len(factors) == 1 {
		return res
	}

	return symbolic.Mul[T](factors...)
}

// getExprRatio computes the ratio of the expression and ceil to the next power
// of two. The input expression should be pre-bound-cancelled. The domainSize
func getExprRatio[T zk.Element](expr *symbolic.Expression[T]) int {
	var (
		board        = expr.Board()
		domainSize   = column.ExprIsOnSameLengthHandles(&board)
		exprDegree   = board.Degree(GetDegree[T](domainSize))
		quotientSize = exprDegree - domainSize + 1
		ratio        = utils.DivCeil(quotientSize, domainSize)
	)
	return utils.NextPowerOfTwo(max(1, ratio))
}

// GetDegree is a generator returning a DegreeGetter that can be passed to
// [symbolic.Expression[T]Board.Degree]. The generator takes the domain size as
// input.
func GetDegree[T zk.Element](size int) func(iface interface{}) int {
	return func(iface interface{}) int {
		switch v := iface.(type) {
		case ifaces.Column[T]:
			// Univariate polynomials is X. We pad them with zeroes so it is safe
			// to return the domainSize directly.
			if size != v.Size() {
				panic("unconsistent sizes for the commitments")
			}
			// The size gives the number of coefficients , but we return the degree
			// hence the - 1
			return v.Size() - 1
		case coin.Info, ifaces.Accessor[T]:
			// Coins are treated
			return 0
		case variables.X[T]:
			return 1
		case variables.PeriodicSample[T]:
			return size - size/v.W
		default:
			utils.Panic("Unknown type %v\n", reflect.TypeOf(v))
		}
		panic("unreachable")
	}
}
