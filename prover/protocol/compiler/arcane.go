package compiler

import (
	"github.com/consensys/accelerated-crypto-monorepo/protocol/compiler/arithmetics"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/compiler/cleanup"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/compiler/logdata"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/compiler/specialqueries"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/compiler/splitter"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/compiler/splitter/sticker"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/compiler/univariates"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/wizard"
)

// Arcane is a grouping of all compilers. It compiles
// any wizard into a single-point polynomial-IOP
func Arcane(minStickSize, targetColSize int, noLog ...bool) func(comp *wizard.CompiledIOP) {
	withLog_ := false
	if len(noLog) > 0 {
		withLog_ = !noLog[0]
	}

	return func(comp *wizard.CompiledIOP) {
		specialqueries.RangeProof(comp)
		specialqueries.CompileFixedPermutations(comp)
		specialqueries.CompilePermutations(comp)
		specialqueries.LogDerivativeLookupCompiler(comp)
		specialqueries.CompileInnerProduct(comp)
		if withLog_ {
			logdata.Log("after-expansion")(comp)
		}
		sticker.Sticker(minStickSize, targetColSize)(comp)
		splitter.SplitColumns(targetColSize)(comp)
		if withLog_ {
			logdata.Log("post-rectangularization")(comp)
		}
		cleanup.CleanUp(comp)
		arithmetics.CompileLocal(comp)
		arithmetics.CompileGlobal(comp)
		univariates.CompileLocalOpening(comp)
		univariates.Naturalize(comp)
		univariates.MultiPointToSinglePoint(targetColSize)(comp)
		if withLog_ {
			logdata.Log("end-of-arcane")(comp)
		}
	}
}
