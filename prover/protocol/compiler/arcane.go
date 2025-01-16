package compiler

import (
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/cleanup"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/globalcs"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/innerproduct"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/localcs"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/logdata"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/lookup"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/permutation"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/specialqueries"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/stitch_split/splitter"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/stitch_split/stitcher"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/univariates"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
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
		permutation.CompileGrandProduct(comp)
		lookup.CompileLogDerivative(comp)
		innerproduct.Compile(comp)
		if withLog_ {
			logdata.Log("after-expansion")(comp)
		}
		stitcher.Stitcher(minStickSize, targetColSize)(comp)
		splitter.Splitter(targetColSize)(comp)
		if withLog_ {
			logdata.Log("post-rectangularization")(comp)
		}
		cleanup.CleanUp(comp)
		localcs.Compile(comp)
		globalcs.Compile(comp)
		univariates.CompileLocalOpening(comp)
		univariates.Naturalize(comp)
		univariates.MultiPointToSinglePoint(targetColSize)(comp)
		if withLog_ {
			logdata.Log("end-of-arcane")(comp)
		}
	}
}
