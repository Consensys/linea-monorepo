package distributed

import (
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

type ModuleName = string

type DistributedWizard struct {
	Bootstrapper       *wizard.CompiledIOP
	DistributedModules []DistributedModule
	Aggregator         *wizard.CompiledIOP
}

// DistributedModule implements the utilities relevant to a single segment.
type DistributedModule struct {
	LookupPermProj *wizard.CompiledIOP
	GlobalLocal    *wizard.CompiledIOP
	VK             [2]field.Element
}

type ModuleDiscoverer interface {
	// Analyze is responsible for letting the module discoverer compute how to
	// group best the columns into modules.
	Analyze(comp *wizard.CompiledIOP)
	NbModules() int
	ModuleList(comp *wizard.CompiledIOP) []ModuleName
	FindModule(col ifaces.Column) ModuleName
}

// This transforms the initial wizard. So it is not really the initial
// wizard anymore. That means the caller can forget about "initialWizard"
// after calling the function.
// maxNbSegment is a large max for the number of segments in a module.
func Distribute(initialWizard *wizard.CompiledIOP, disc ModuleDiscoverer, maxNbSegments int) DistributedWizard {
	// it updates the map of Modules-Columns (that is a field of initialWizard).
	disc.Analyze(initialWizard)

	moduleLs := disc.ModuleList(initialWizard)
	distModules := []DistributedModule{}

	for _, modName := range moduleLs {
		// Segment Compilation;
		// Compile every dist module with the same sequence of compilation steps for uniformity
		distMod := extractDistModule(initialWizard, disc, modName)
		distModules = append(distModules, distMod)
	}

	// for each [DistributedModule] it checks the consistency among
	// its replications where the number of replications is maxNbSegments.
	aggr := aggregator(maxNbSegments, distModules, moduleLs)

	return DistributedWizard{
		Bootstrapper:       initialWizard,
		DistributedModules: distModules,
		Aggregator:         aggr,
	}
}

func addSplittingStep(comp *wizard.CompiledIOP, disc ModuleDiscoverer) {
	panic("unimplemented")
}

func extractDistModule(comp *wizard.CompiledIOP, disc ModuleDiscoverer, moduleName ModuleName) DistributedModule {
	panic("unimplemented")
}

func aggregator(n int, idsModules []DistributedModule, moduleNames []string) *wizard.CompiledIOP {
	panic("unimplemented")
}
