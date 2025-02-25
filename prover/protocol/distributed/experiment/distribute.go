package experiment

import (
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/horner"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/logderivativesum"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/mimc"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/permutation"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/specialqueries"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

// precompileInitialWizard pre-compiles the initial wizard protocol by applying all the
// compilation steps needing to be run before the splitting phase. Its role is to
// ensure that all of the queries that cannot be processed by the splitting phase
// are removed from the compiled IOP.
func precompileInitialWizard(comp *wizard.CompiledIOP) {
	mimc.CompileMiMC(comp)
	specialqueries.RangeProof(comp)
	specialqueries.CompileFixedPermutations(comp)
	logderivativesum.LookupIntoLogDerivativeSum(comp)
	permutation.CompileIntoGdProduct(comp)
	horner.ProjectionToHorner(comp)
}
