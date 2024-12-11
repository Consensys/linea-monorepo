package inclusion

import (
	"slices"

	"github.com/consensys/linea-monorepo/prover/protocol/column"
	lookUp "github.com/consensys/linea-monorepo/prover/protocol/compiler/lookup"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	sym "github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
)

const (
	packingArity = 3
)

// zCtx is an internal compilation structure responsible for grouping the
// "Sigma" columns together so that we can trade-off commitment complexity with
// global constraint degree.
//
// All the grouped "Sigma" columns are of the same size and same round for the
// same context.
// M is computing the appearance of the rows of T in the rows of S.

// M:       tableCtx.M,
// S:       checkTable,
// T:       lookupTable,
// SFilter: includedFilters,

type logDerivativeSumCtx struct {
	zCtx *lookUp.ZCtx
	// the global sum in LogDerivative
	PI ifaces.Column
}

// check permutation and see how/where compile is called (see how to constracut z there)
// when constructing z, check if z is T or S
// and change T -> -M, S -> +Filter
// S or T -> ({S,T} + X)
// compile should be called inside CompileGrandSum
func (lgSum *logDerivativeSumCtx) compileLogDerivativeSum(comp *wizard.CompiledIOP) {

	var (
		z     = lgSum.zCtx
		numZs = utils.DivCeil(
			len(z.SigmaDenominator),
			packingArity,
		)
	)
	z.Zs = make([]ifaces.Column, numZs)
	z.ZOpenings = make([]query.LocalOpening, numZs)
	z.ZNumeratorBoarded = make([]sym.ExpressionBoard, numZs)
	z.ZDenominatorBoarded = make([]sym.ExpressionBoard, numZs)

	lgSum.PI = comp.InsertColumn(
		z.Round,
		lookUp.DeriveName[ifaces.ColID]("PI", comp.SelfRecursionCount, z.Round, z.Size),
		1,
		column.PublicInput,
	)

	for i := range z.Zs {

		var (
			packedNum = safeAnySubSlice(z.SigmaNumerator, i*packingArity, (i+1)*packingArity)
			packedDen = safeAnySubSlice(z.SigmaDenominator, i*packingArity, (i+1)*packingArity)

			zNumerator   = sym.NewConstant(0)
			zDenominator = sym.Mul(packedDen...)
		)

		for j := range packedNum {
			term := packedNum[j]
			for k := range packedDen {
				if k != j {
					term = sym.Mul(term, packedDen[k])
				}
			}
			zNumerator = sym.Add(zNumerator, term)
		}

		z.ZNumeratorBoarded[i] = zNumerator.Board()
		z.ZDenominatorBoarded[i] = zDenominator.Board()

		z.Zs[i] = comp.InsertCommit(
			z.Round,
			lookUp.DeriveName[ifaces.ColID]("Z", comp.SelfRecursionCount, z.Round, z.Size, i),
			z.Size,
		)

		// initial condition
		comp.InsertLocal(
			z.Round,
			lookUp.DeriveName[ifaces.QueryID]("Z_CONSISTENCY_START", comp.SelfRecursionCount, z.Round, z.Size, i),
			sym.Sub(
				zNumerator,
				sym.Mul(
					z.Zs[i],
					zDenominator,
				),
			),
		)

		// consistency check
		comp.InsertGlobal(
			z.Round,
			lookUp.DeriveName[ifaces.QueryID]("Z_CONSISTENCY", comp.SelfRecursionCount, z.Round, z.Size, i),
			sym.Sub(
				zNumerator,
				sym.Mul(
					sym.Sub(z.Zs[i], column.Shift(z.Zs[i], -1)),
					zDenominator,
				),
			),
		)

		// local opening of the final value of the Z polynomial
		z.ZOpenings[i] = comp.InsertLocalOpening(
			z.Round,
			lookUp.DeriveName[ifaces.QueryID]("Z_FINAL", comp.SelfRecursionCount, z.Round, z.Size, i),
			column.Shift(z.Zs[i], -1),
		)
	}
}

// attempt to take the subslice of a slice, and truncates or returns an empty
// slice if the parameters are out of bounds.
func safeAnySubSlice[T any](t []T, start, stop int) []any {

	if stop < start {
		panic("invalid argument")
	}

	var tmp []T

	switch {
	case start >= len(t):
		return []any{}
	case stop >= len(t):
		tmp = t[start:]
	default:
		tmp = t[start:stop]
	}

	res := make([]any, len(tmp))
	for i := range res {
		res[i] = tmp[i]
	}

	return res
}

// IntoLogDerivativeSum handles the remaining process after PushToZCatalog.
func InsertLogDerivativeSum(comp *wizard.CompiledIOP, lastRound int, zCatalog map[[2]int]*lookUp.ZCtx) {
	var (
		zEntries = [][2]int{}
		va       = finalEvaluationCheck{}
	)

	// This loop is necessary to build a sorted list of the entries of zCatalog.
	// Without it, if we tried to loop over zCatalog directly, the entries would
	// be processed in a non-deterministic order. The sorting order itself is
	// without importance, what matters is that zEntries is in deterministic
	// order.
	for entry := range zCatalog {
		zEntries = append(zEntries, entry)
	}

	slices.SortFunc(zEntries, func(a, b [2]int) int {
		switch {
		case a[0] < b[0]:
			return -1
		case a[0] > b[0]:
			return 1
		case a[1] < b[1]:
			return -1
		case a[1] > b[1]:
			return 1
		default:
			return 0
		}
	})

	// compile zCatalog
	for _, entry := range zEntries {
		zC := zCatalog[entry]
		logDerivSumCtx :=
			logDerivativeSumCtx{
				zCtx: zC,
			}
		// z-packing compile
		logDerivSumCtx.compileLogDerivativeSum(comp)
		// entry[0]:round, entry[1]: size
		// the round that Gamma was registered.

		// pushZAssignment(zAssignmentTask(*zC))
		va.ZOpenings = append(va.ZOpenings, zC.ZOpenings...)
		va.Name = zC.Name
	}

	comp.RegisterVerifierAction(lastRound, &va)
}
