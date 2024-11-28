package distributed

import (
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/mimc"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/specialqueries"
	"github.com/consensys/linea-monorepo/prover/protocol/distributed/compiler/inclusion"
	"github.com/consensys/linea-monorepo/prover/protocol/distributed/compiler/permutation"
	"github.com/consensys/linea-monorepo/prover/protocol/distributed/compiler/projection"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

type moduleName = string

type DistributedWizard struct {
	Bootstrapper       *wizard.CompiledIOP
	DistributedModules []DistributedModule
	Aggregator         *wizard.CompiledIOP
}

type DistributedModule struct {
	LookupPermProj *wizard.CompiledIOP
	GlobalLocal    *wizard.CompiledIOP
	VK             [2]field.Element
}

type ModuleDiscoverer interface {
	// Analyze is responsible for letting the module discoverer compute how to
	// group best the columns into modules.
	Analyze(comp *wizard.CompiledIOP)
	NbModules() moduleName
	ModuleList() []string
	FindModule(col ifaces.Column) moduleName
}

// This transforms the initial wizard. So it is not really the initial
// wizard anymore. That means the caller can forget about "initialWizard"
// after calling the function.
func Distribute(initialWizard *wizard.CompiledIOP, disc ModuleDiscoverer) DistributedWizard {

	prepare(initialWizard)
	disc.Analyze(initialWizard)

	addSplittingStep(disc, initialWizard) // adds a prover step to "push" all the sub-witness assignment

	moduleLs := disc.ModuleList()
	distModules := []DistributedModule{}

	for _, modName := range moduleLs {
		distMod := extractDistModule(initialWizard, disc, modName)
		distModules = append(distModules, distMod)
	}

	// Compile every dist module with the same sequence of compilation steps for uniformity

	return DistributedWizard{
		Bootstrapper: initialWizard,
	}
}

// It adds the compilation steps
func prepare(comp *wizard.CompiledIOP) {

	mimc.CompileMiMC(comp)
	specialqueries.RangeProof(comp)
	specialqueries.CompileFixedPermutations(comp)

	inclusion.IntoLogDerivativeSum(comp)
	permutation.IntoGrandProduct(comp)
	projection.IntoGrandSum(comp)
}

func addSplittingStep(comp *wizard.CompiledIOP) {
	panic("unimplemented")
}

func extractDistModule(comp *wizard.CompiledIOP, disc ModuleDiscoverer, moduleName moduleName) DistributedModule {
	panic("unimplemented")
}
