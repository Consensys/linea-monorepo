package mimc

import (
	"fmt"
	"strings"

	"github.com/consensys/accelerated-crypto-monorepo/crypto/mimc"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/dedicated/expr_handle"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/ifaces"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/wizard"
	"github.com/consensys/accelerated-crypto-monorepo/symbolic"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
)

// Internally checks the correctness of hashing a MiMC blocks in parallel.
func manualCheckMiMCBlock(comp *wizard.CompiledIOP, blocks, oldStates, newStates ifaces.Column) {

	ctx := mimcCtx{
		comp:      comp,
		oldStates: oldStates,
		blocks:    blocks,
		newStates: newStates,
	}

	// Creates an intermediate column for each round
	s := blocks
	for i := 0; i < len(mimc.Constants)-1; i++ {
		s = ctx.manualPermRound(s, i)
	}

	// And checks consistency of the last one with the alleged resulting states
	ctx.manualCheckFinalRoundPerm(s)
}

// Utility struct wrapping all the intermediate values of the MiMC wizard
type mimcCtx struct {
	comp      *wizard.CompiledIOP
	oldStates ifaces.Column
	blocks    ifaces.Column
	newStates ifaces.Column
}

// Applies the round permutation #i, works for all except the last round
func (ctx *mimcCtx) manualPermRound(s ifaces.Column, i int) ifaces.Column {

	// Sanity-check that the function was not called over the last round
	if i == len(mimc.Constants)-1 {
		utils.Panic("For the last round, use the FinalRoundPerm function instead")
	}

	sVar := ifaces.ColumnAsVariable(s)
	kVar := ifaces.ColumnAsVariable(ctx.oldStates)
	ark := symbolic.NewConstant(mimc.Constants[i])

	// Computes the MiMC round expression s' = (s + k + ark)^5
	expr := sVar.Add(kVar).Add(ark).Pow(5)

	return expr_handle.ExprHandle(ctx.comp, expr, mimcName(ctx.comp, ctx.newStates.GetColID(), i))
}

// Applies the final round of permutation
func (ctx *mimcCtx) manualCheckFinalRoundPerm(s ifaces.Column) {

	// Sanity-check that the function was not called over the last round
	if len(mimc.Constants) == 1 {
		utils.Panic("For the last round, use the PermRound function instead")
	}

	sVar := ifaces.ColumnAsVariable(s)
	kVar := ifaces.ColumnAsVariable(ctx.oldStates)
	ark := symbolic.NewConstant(mimc.Constants[len(mimc.Constants)-1])
	oldState := ifaces.ColumnAsVariable(ctx.oldStates)
	newState := ifaces.ColumnAsVariable(ctx.newStates)
	block := ifaces.ColumnAsVariable(ctx.blocks)

	// Computes the MiMC round expression s' = (s + k + ark)^5
	// We add twice the oldState, because the MiMC specification requires
	// adding it in order to turn the MiMC keyed-permutation into a cipher
	// and a second time because the Miyaguchi-Preneel construction requires
	// adding it to the cipertext to turn the cipher into a hash function.
	// The final sub is because we want to assess the correctness of the new
	// state
	expr := sVar.
		Add(kVar).Add(ark).Pow(5).
		Add(oldState).Add(oldState).
		Add(block).Sub(newState)

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
