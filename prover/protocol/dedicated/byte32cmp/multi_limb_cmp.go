package byte32cmp

import (
	"fmt"
	"sync"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/dedicated"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/protocol/wizardutils"
	sym "github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/parallel"
)

// multiLimbCmp is a dedicated wizard which can compare two [LimbColumns] and
// thereby constructing indicator columns indicating whether the first operand
// is lower, greater or if the two operands are equals.
//
// This object implements the [wizard.ProverAction] interface and is meant to
// be run to compute the assignment of the returned column by [CmpMultiLimbs].
type multiLimbCmp struct {

	// isGreater and isLower are columns to be assigned by the context. The
	// reason isEqual is missing is because its handling is defered
	// to a dedicated sub-context.
	isGreater, isLower ifaces.Column

	// isEqualCtx is the dedicated [wizard.ProverAction] responsible for
	// assigning the returned isEqual column.
	isEqualCtx wizard.ProverAction

	// nonNegative syndrom is an internal column created such that it should
	// always represent a number of size 1 << numLimbs. It is constructed using
	// the result of the other limb-by-limb comparison.
	nonNegativeSyndrom ifaces.Column

	// syndromBoard is an expression board used to assign the non-negative
	// syndromColumn its is the same as (isGreater - isLower).
	syndromBoard sym.ExpressionBoard

	// The subCtxs are the [wizard.ProverAction] responsible for doing the
	// assignment part for each limb-by-limb comparison.
	subCtxs []wizard.ProverAction
}

// CmpMultiLimbs returns three columns: isGreater, isEqual and isLower which
// are mutually-exclusive pre-constrainted binary columns and a
// [wizard.ProverAction] computing their assignment. The returned columns
// indicates whether the words represented by `a` are greater, equal or
// smaller compared to `b`. It can be used to perform comparison between large
// numbers.
//
// The function assumes that `a` and `b` are already constrained to be well-formed.
//
// The function works by using the [CmpSmallCols] limb-wise to obtain a partial
// comparison limb-by-limbs (gi, ei, li). From there, isEqual is constrained
// by "all the ei == 1" <==> "isEqual = 1" using a [dedicated.IsZero]
// sub-routine. isGreater and isLower are constrained using what we call a
// syndrom = \sum_i (gi - li) 2**i. We have that isGreater = 1 <=> syndrom > 0
// and isLower <=> syndrom < 0.
//
// The function is limited to 64 limbs.
func CmpMultiLimbs(comp *wizard.CompiledIOP, a, b LimbColumns) (isGreater, isEqual, isLower ifaces.Column, pa wizard.ProverAction) {

	if len(a.Limbs) != len(b.Limbs) {
		utils.Panic("a and b don't have the same number of limbs: %v %v", len(a.Limbs), len(b.Limbs))
	}

	if len(a.Limbs) == 0 || len(b.Limbs) == 0 {
		utils.Panic("a and b cannot have zero limbs %v %v", len(a.Limbs), len(b.Limbs))
	}

	if a.LimbBitSize != b.LimbBitSize {
		utils.Panic("a and b don't use the same limb-bit-size: %v %v", a.LimbBitSize, b.LimbBitSize)
	}

	if a.IsBigEndian != b.IsBigEndian {
		utils.Panic("a and b don't have the same endianness")
	}

	var (
		nRows   = ifaces.AssertSameLength(a.Limbs...)
		nRowsB  = ifaces.AssertSameLength(b.Limbs...)
		ctxName = func(subName string) string {
			return fmt.Sprintf("CMP_MULTI_LIMB_%v_%v", len(comp.Columns.AllKeys()), subName)
		}
	)

	if nRows != nRowsB {
		utils.Panic("a and b must have the same length")
	}

	var (
		isBigEndian     = a.IsBigEndian
		roundA          = wizardutils.MaxRound(a.Limbs...)
		round           = max(roundA, wizardutils.MaxRound(b.Limbs...))
		numLimbs        = len(a.Limbs)
		numBitsPerLimbs = a.LimbBitSize
		ctx             = &multiLimbCmp{
			isGreater:          comp.InsertCommit(round, ifaces.ColIDf(ctxName("IS_GREATER")), nRows),
			isLower:            comp.InsertCommit(round, ifaces.ColIDf(ctxName("IS_LOWER")), nRows),
			nonNegativeSyndrom: comp.InsertCommit(round, ifaces.ColIDf(ctxName("MUST_BE_POSITIVE")), nRows),
		}

		syndromExpression = sym.NewConstant(0)
		allLimbsEqual     = sym.NewConstant(0)
	)

	for i := 0; i < numLimbs; i++ {
		var (
			g, e, l, lCtx = CmpSmallCols(comp, a.Limbs[i], b.Limbs[i], numBitsPerLimbs)
		)

		ctx.subCtxs = append(ctx.subCtxs, lCtx)
		allLimbsEqual = sym.Add(allLimbsEqual, sym.Sub(e, 1))

		factor := 1 << i
		if isBigEndian {
			factor = 1 << (numLimbs - i - 1)
		}

		syndromExpression = sym.Add(
			syndromExpression,
			sym.Mul(factor, sym.Sub(g, l)),
		)
	}

	comp.InsertGlobal(
		round,
		ifaces.QueryID(ctxName("IS_GREATER_IS_BINARY")),
		sym.Mul(ctx.isGreater, sym.Sub(ctx.isGreater, 1)),
	)

	comp.InsertGlobal(
		round,
		ifaces.QueryID(ctxName("IS_LOWER_IS_BINARY")),
		sym.Mul(ctx.isLower, sym.Sub(ctx.isLower, 1)),
	)

	isEqual, ctx.isEqualCtx = dedicated.IsZero(comp, allLimbsEqual)

	comp.InsertGlobal(
		round,
		ifaces.QueryIDf(ctxName("FLAGS_MUTUALLY_EXCLUSIVE")),
		sym.Sub(1, ctx.isGreater, isEqual, ctx.isLower),
	)

	ctx.syndromBoard = syndromExpression.Board()

	comp.InsertGlobal(
		round,
		ifaces.QueryID(ctxName("COMPUTE_NN_SYNDROME")),
		sym.Sub(
			ctx.nonNegativeSyndrom,
			sym.Mul(
				sym.Sub(ctx.isGreater, ctx.isLower),
				syndromExpression,
			),
		),
	)

	comp.InsertRange(
		round,
		ifaces.QueryID(ctxName("RANGE_CHECK_NN_SYNDROM")),
		ctx.nonNegativeSyndrom,
		1<<numLimbs,
	)

	return ctx.isGreater, isEqual, ctx.isLower, ctx
}

// Run implements the [wizard.ProverAction] interface.
func (mCmp *multiLimbCmp) Run(run *wizard.ProverRuntime) {

	// This will assign the per-limbs comparision contexts
	parallel.Execute(len(mCmp.subCtxs), func(start, stop int) {
		for i := start; i < stop; i++ {
			mCmp.subCtxs[i].Run(run)
		}
	})

	// This will assign the IsEqual column. It can be done in parallel of the
	// the rest. But it requires the per-limb context to be run prior to this.
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		mCmp.isEqualCtx.Run(run)
		wg.Done()
	}()

	var (
		syndrom   = wizardutils.EvalExprColumn(run, mCmp.syndromBoard)
		isGreater = make([]field.Element, mCmp.isGreater.Size())
		isLower   = make([]field.Element, mCmp.isLower.Size())
		nnSyndrom = make([]field.Element, mCmp.isLower.Size())
	)

	for i := range isGreater {
		var (
			sF = syndrom.Get(i)
		)

		if sF.IsUint64() && !sF.IsZero() {
			isGreater[i] = field.One()
			nnSyndrom[i] = sF
		}

		if !sF.IsUint64() {
			isLower[i] = field.One()
			nnSyndrom[i].Neg(&sF)
		}
	}

	run.AssignColumn(mCmp.isGreater.GetColID(), smartvectors.NewRegular(isGreater))
	run.AssignColumn(mCmp.isLower.GetColID(), smartvectors.NewRegular(isLower))
	run.AssignColumn(mCmp.nonNegativeSyndrom.GetColID(), smartvectors.NewRegular(nnSyndrom))

	wg.Wait()
}
