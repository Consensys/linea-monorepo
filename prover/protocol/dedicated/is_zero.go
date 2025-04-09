package dedicated

import (
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/protocol/wizardutils"
	sym "github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
)

type isZeroCtx struct {
	ctxID     int
	size      int
	round     int
	c         *sym.Expression
	mask      *sym.Expression
	invOrZero ifaces.Column
	isZero    ifaces.Column
}

// IsZero returns a fully constrained binary column z such that
// c[i] = 0 <=> z[i] = 1 and c[i] != 0  <=> z[i] = 0. `c` can be either a
// symbolic expression or a column. If the input expression contains a shift,
// then the resulting column `z` is computed and constrained assuming the shift
// wraps around.
//
// The function also returns a context object that can be invoked to perform the
// assignment of `z` and the intermediate internal columns. It has to be called
// explictly by the the caller during the prover runtime.
func IsZero(comp *wizard.CompiledIOP, c any) (ifaces.Column, wizard.ProverAction) {

	var (
		ctx = &isZeroCtx{
			ctxID: len(comp.QueriesNoParams.AllKeys()),
		}
	)

	switch c1 := c.(type) {
	case ifaces.Column:
		ctx.round = c1.Round()
		ctx.size = c1.Size()
		ctx.c = sym.NewVariable(c1)
	case *sym.Expression:
		board := c1.Board()
		ctx.c = c1
		ctx.size = column.ExprIsOnSameLengthHandles(&board)
		ctx.round = wizardutils.LastRoundToEval(c1)
	}

	compileIsZeroWithSize(comp, ctx)

	return ctx.isZero, ctx
}

// IsZeroMasked is an [IsZero] but allows passing an additional `mask`,
// expectedly a binary column. The caller is responsible for
// ensuring/constraining that the column is binary. All the generated columns
// will be zeroied when the mask is up.
func IsZeroMask(comp *wizard.CompiledIOP, c, mask any) (ifaces.Column, wizard.ProverAction) {

	var (
		ctx = &isZeroCtx{
			ctxID: len(comp.QueriesNoParams.AllKeys()),
		}
	)

	ctx.c, ctx.round, ctx.size = wizardutils.AsExpr(c)
	m, roundMask, sizeMask := wizardutils.AsExpr(mask)
	ctx.round = max(roundMask, ctx.round)

	if sizeMask != ctx.size {
		utils.Panic("the size of the mask if %v but the column's size is %v", sizeMask, ctx.size)
	}

	ctx.mask = m

	compileIsZeroWithSize(comp, ctx)

	return ctx.isZero, ctx
}

func compileIsZeroWithSize(comp *wizard.CompiledIOP, ctx *isZeroCtx) {

	ctx.isZero = comp.InsertCommit(
		ctx.round,
		ifaces.ColIDf("IS_ZERO_%v_RES", ctx.ctxID),
		ctx.size,
	)

	ctx.invOrZero = comp.InsertCommit(
		ctx.round,
		ifaces.ColIDf("IS_ZERO_%v_INVERSE_OR_ZERO", ctx.ctxID),
		ctx.size,
	)

	var mask = any(1)
	if ctx.mask != nil {
		mask = ctx.mask
	}

	comp.InsertGlobal(
		0,
		ifaces.QueryIDf("IS_ZERO_%v_RES_IS_ONE_IF_C_ISZERO", ctx.ctxID),
		sym.Add(ctx.isZero, sym.Mul(ctx.invOrZero, ctx.c), sym.Neg(mask)),
	)

	comp.InsertGlobal(
		0,
		ifaces.QueryIDf("IS_ZERO_%v_RES_IS_ZERO_IF_C_ISNONZERO", ctx.ctxID),
		sym.Mul(ctx.isZero, ctx.c),
	)

	if ctx.mask != nil {
		comp.InsertGlobal(
			0,
			ifaces.QueryIDf("IS_ZERO_%v_RES_IS_MASKED", ctx.ctxID),
			sym.Sub(ctx.isZero, sym.Mul(mask, ctx.isZero)),
		)
	}
}

// Run implements the [wizard.ProverAction] interface
func (ctx *isZeroCtx) Run(run *wizard.ProverRuntime) {
	var (
		c         = column.EvalExprColumn(run, ctx.c.Board())
		invOrZero = smartvectors.BatchInvert(c)
		isZero    = smartvectors.IsZero(c)
	)

	if ctx.mask != nil {
		mask := column.EvalExprColumn(run, ctx.mask.Board())
		invOrZero = smartvectors.Mul(invOrZero, mask)
		isZero = smartvectors.Mul(isZero, mask)
	}

	run.AssignColumn(ctx.invOrZero.GetColID(), invOrZero, wizard.DisableAssignmentSizeReduction)
	run.AssignColumn(ctx.isZero.GetColID(), isZero, wizard.DisableAssignmentSizeReduction)
}
