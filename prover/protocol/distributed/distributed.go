package distributed

import (
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/mimc"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/specialqueries"
	"github.com/consensys/linea-monorepo/prover/protocol/distributed/compiler/global"
	"github.com/consensys/linea-monorepo/prover/protocol/distributed/compiler/inclusion"
	"github.com/consensys/linea-monorepo/prover/protocol/distributed/compiler/local"
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
	NbModules() moduleName
	ModuleList(comp *wizard.CompiledIOP) []string
	FindModule(col ifaces.Column) moduleName
}

// This transforms the initial wizard. So it is not really the initial
// wizard anymore. That means the caller can forget about "initialWizard"
// after calling the function.
// maxNbSegment is a large max for the number of segments in a module.
func Distribute(initialWizard *wizard.CompiledIOP, disc ModuleDiscoverer, maxNbSegments int) DistributedWizard {

	// prepare the  initialWizard for the distribution. e.g.,
	// adding auxiliary columns or dividing a lookup query to two queries one over T and the other over S.
	prepare(initialWizard)
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

// It adds the compilation steps
func prepare(comp *wizard.CompiledIOP) {

	mimc.CompileMiMC(comp)
	specialqueries.RangeProof(comp)
	specialqueries.CompileFixedPermutations(comp)

	inclusion.IntoLogDerivativeSum(comp)
	permutation.IntoGrandProduct(comp)
	projection.IntoGrandSum(comp)
	local.IntoDistributedLocal(comp)
	global.IntoDistributedGlobal(comp)
}

func addSplittingStep(comp *wizard.CompiledIOP, disc ModuleDiscoverer) {
	panic("unimplemented")
}

func extractDistModule(comp *wizard.CompiledIOP, disc ModuleDiscoverer, moduleName moduleName) DistributedModule {
	panic("unimplemented")
}

func aggregator(n int, idsModules []DistributedModule, moduleNames []string) *wizard.CompiledIOP {
	panic("unimplemented")
}
