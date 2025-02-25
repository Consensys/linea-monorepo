package experiment

import (
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/horner"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/logderivativesum"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/mimc"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/permutation"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/specialqueries"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

// DistributedWizard represents a wizard protocol that has undergone a
// distributed compilation process.
type DistributedWizard struct {

	// ModuleNames is the list of the names of the modules that compose
	// the distributed protocol.
	ModuleNames []ModuleName

	// LPPs is the list of the LPP parts for every modules
	LPPs []*ModuleLPP

	// GLs is the list of the GL parts for every modules
	GLs []*ModuleGL

	// Bootstrapper is the original compiledIOP precompiled with a few
	// preparation steps.
	Bootstrapper *wizard.CompiledIOP

	// Disc is the [ModuleDiscoverer] used to delimitate the scope for
	// each module.
	Disc ModuleDiscoverer
}

// Distribute returns a [DistributedWizard] from a [wizard.CompiledIOP]. It
// takes ownership of the input [wizard.CompiledIOP]. And uses disc to design
// the scope of each module.
func Distribute(comp *wizard.CompiledIOP, disc ModuleDiscoverer) DistributedWizard {

	distributedWizard := DistributedWizard{
		Bootstrapper: precompileInitialWizard(comp),
	}

	disc.Analyze(distributedWizard.Bootstrapper)
	distributedWizard.ModuleNames = disc.ModuleList()

	for _, moduleName := range distributedWizard.ModuleNames {

		moduleFilter := moduleFilter{
			Disc:   disc,
			Module: moduleName,
		}

		filteredModuleInputs := moduleFilter.FilterCompiledIOP(
			distributedWizard.Bootstrapper,
		)

		distributedWizard.LPPs = append(
			distributedWizard.LPPs,
			BuildModuleLPP(&filteredModuleInputs),
		)

		distributedWizard.GLs = append(
			distributedWizard.GLs,
			BuildModuleGL(&filteredModuleInputs),
		)
	}

	return distributedWizard
}

// CompileModules applies the compilation steps to each modules identically.
func (dist *DistributedWizard) CompileModules(compilers ...func(*wizard.CompiledIOP)) {
	for i := range dist.ModuleNames {
		for _, compile := range compilers {
			compile(dist.LPPs[i].Wiop)
			compile(dist.GLs[i].Wiop)
		}
	}
}

// precompileInitialWizard pre-compiles the initial wizard protocol by applying all the
// compilation steps needing to be run before the splitting phase. Its role is to
// ensure that all of the queries that cannot be processed by the splitting phase
// are removed from the compiled IOP.
func precompileInitialWizard(comp *wizard.CompiledIOP) *wizard.CompiledIOP {
	mimc.CompileMiMC(comp)
	specialqueries.RangeProof(comp)
	specialqueries.CompileFixedPermutations(comp)
	logderivativesum.LookupIntoLogDerivativeSum(comp)
	permutation.CompileIntoGdProduct(comp)
	horner.ProjectionToHorner(comp)
	return comp
}
