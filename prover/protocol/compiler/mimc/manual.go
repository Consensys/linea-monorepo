package mimc

import (
	"fmt"
	"strings"

	"github.com/consensys/linea-monorepo/prover/crypto/mimc"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/protocol/wizardutils"
	"github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
)

// Internally checks the correctness of hashing a MiMC blocks in parallel.
// Namely, on every row i of the columns (blocks, oldStates, newStates), we
// have that mimcF(oldState, blocks) == newState.
func manualCheckMiMCBlock(comp *wizard.CompiledIOP, blocks, oldStates, newStates ifaces.Column) {

	ctx := mimcCtx{
		comp:      comp,
		oldStates: oldStates,
		blocks:    blocks,
		newStates: newStates,
	}

	round := column.MaxRound(blocks, oldStates, newStates)

	// Creates an intermediate column for each round
	s := blocks
	ctx.intermediateResult = append(ctx.intermediateResult, s)
	for i := 0; i < len(mimc.Constants)-1; i++ {
		s = ctx.manualPermRound(s, i)
		ctx.intermediateResult = append(ctx.intermediateResult, s)
	}

	// And checks consistency of the last one with the alleged resulting states
	ctx.manualCheckFinalRoundPerm(s)

	comp.SubProvers.AppendToInner(round, ctx.assign)
}

// Utility struct wrapping all the intermediate values of the MiMC wizard
type mimcCtx struct {
	comp               *wizard.CompiledIOP
	oldStates          ifaces.Column
	blocks             ifaces.Column
	newStates          ifaces.Column
	intermediateResult []ifaces.Column
	intermediatePow4   []ifaces.Column
}

// Applies the round permutation #i, works for all except the last round
func (ctx *mimcCtx) manualPermRound(s ifaces.Column, i int) ifaces.Column {

	// Sanity-check that the function was not called over the last round
	if i == len(mimc.Constants)-1 {
		utils.Panic("For the last round, use the FinalRoundPerm function instead")
	}

	var (
		sVar = ifaces.ColumnAsVariable(s)
		kVar = ifaces.ColumnAsVariable(ctx.oldStates)
		ark  = symbolic.NewConstant(mimc.Constants[i])
	)

	// Computes the MiMC round expression s' = (s + k + ark)^17
	sumPow4 := sumPow4(ctx, i, sVar, kVar, ark)
	sumPow16 := ifaces.ColumnAsVariable(sumPow4).Pow(4)
	expr := sVar.Add(kVar).Add(ark).Mul(sumPow16)

	return mimcExprHandle(ctx.comp, expr, mimcName(ctx.comp, ctx.newStates.GetColID(), i))
}

// Applies the final round of permutation
func (ctx *mimcCtx) manualCheckFinalRoundPerm(s ifaces.Column) {

	// Sanity-check that the function was not called over the last round
	if len(mimc.Constants) == 1 {
		utils.Panic("For the last round, use the PermRound function instead")
	}

	var (
		sVar     = ifaces.ColumnAsVariable(s)
		kVar     = ifaces.ColumnAsVariable(ctx.oldStates)
		ark      = symbolic.NewConstant(mimc.Constants[len(mimc.Constants)-1])
		oldState = ifaces.ColumnAsVariable(ctx.oldStates)
		newState = ifaces.ColumnAsVariable(ctx.newStates)
		block    = ifaces.ColumnAsVariable(ctx.blocks)
	)

	// Computes the MiMC round expression s' = (s + k + ark)^17
	// We add twice the oldState, because the MiMC specification requires
	// adding it in order to turn the MiMC keyed-permutation into a cipher
	// and a second time because the Miyaguchi-Preneel construction requires
	// adding it to the cipertext to turn the cipher into a hash function.
	// The final sub is because we want to assess the correctness of the new
	// state

	// This creates an intermediate columns
	sumPow4 := sumPow4(ctx, len(mimc.Constants)-1, sVar, kVar, ark)

	// And we get the final expression by completing it
	sumPow16 := ifaces.ColumnAsVariable(sumPow4)
	sumPow16 = sumPow16.Pow(4)
	expr := sVar.Add(kVar).Add(ark)
	expr = expr.Mul(sumPow16)
	expr = expr.Add(oldState).Add(oldState).Add(block)
	expr = expr.Sub(newState)

	// We use the intermediate result "s" to deduce the interaction round of
	// the construction. This is not to be confused for the "permutation" round.
	round := s.Round()

	// And add this as a global constraint
	ctx.comp.InsertGlobal(round, ifaces.QueryID(mimcName(ctx.comp, ctx.newStates.GetColID(), "FINAL")), expr)

}

func mimcName(comp *wizard.CompiledIOP, args ...interface{}) string {
	// Format all the arguments independantyl
	fmttedArgs := make([]string, len(args))
	for i := range args {
		fmttedArgs[i] = fmt.Sprintf("%v", args[i])
	}

	// Join them with "_" and prefix them with an indicator for MIMC
	return fmt.Sprintf("MIMC_%v_%s", comp.SelfRecursionCount, strings.Join(fmttedArgs, "_"))
}

// returns an expr handle computing the sum of the input columns to the power4
func sumPow4(ctx *mimcCtx, mimcRound int, cols ...*symbolic.Expression) ifaces.Column {
	res := symbolic.NewConstant(0)

	for _, arg := range cols {
		res = res.Add(arg)
	}
	res = res.Pow(4)

	// expreHandle
	h := mimcExprHandle(
		ctx.comp,
		res,
		mimcName(ctx.comp, "SUMPOW4", ctx.newStates.GetColID(), mimcRound),
	)

	ctx.intermediatePow4 = append(ctx.intermediatePow4, h)
	return h
}

// helper function for MiMC
// Create a handle from an expression.
// For general-purpose ExprHandle, use package expr_handle
func mimcExprHandle(comp *wizard.CompiledIOP, expr *symbolic.Expression, name ...string) ifaces.Column {

	maxRound := wizardutils.LastRoundToEval(expr)
	board := expr.Board()
	length := column.ExprIsOnSameLengthHandles(&board)

	handleName := fmt.Sprintf("SYMBOLIC_%v", expr.ESHash.String())
	if len(name) > 0 {
		handleName = name[0]
	}

	// res
	res := comp.InsertCommit(maxRound, ifaces.ColID(handleName), length) //create column

	// cs
	comp.InsertGlobal(maxRound, ifaces.QueryID(handleName), expr.Sub(ifaces.ColumnAsVariable(res))) //create constraints

	return res
}
