package distributed

import (
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

type moduleName = string

type DistributedWizard struct {
	// initializedWizard
	Bootstrapper *wizard.CompiledIOP
	// compiledIOPs for a segment of each module.
	// Since splits are fair/uniform over a module. One segment can represent all the segments.
	DistributedModules []DistributedModule
	// it is a compiledIOP object representing the consistency checks among the segments.
	Aggregator *wizard.CompiledIOP
}

// DistributedModule implements the utilities relevant to a single segment from a module.
type DistributedModule struct {
	LookupPermProj *wizard.CompiledIOP
	GlobalLocal    *wizard.CompiledIOP
}

// ModuleDiscoverer a set of methods responsible for the horizontal splittings (i.e., splitting to modules)
type ModuleDiscoverer interface {
	// Analyze is responsible for letting the module discoverer compute how to
	// group best the columns into modules.
	Analyze(comp *wizard.CompiledIOP)
	ModuleList(comp *wizard.CompiledIOP) []string
	FindModule(col ifaces.Column) moduleName
	FindNbSegments(moduleName string, SegParam map[string]int) int
}

// This transforms the initial wizard. So it is not really the initial
// wizard anymore. That means the caller can forget about "initialWizard"
// after calling the function.
// maxSegmentSize is a static parameter for the max size of the columns in segments.
// maxNumSegment give the max number of segments in a module.
func Distribute(initialWizard *wizard.CompiledIOP, disc ModuleDiscoverer, maxSegmentSize, maxNumSegment int) DistributedWizard {

	// analyze the initialWizard to split it to modules.
	disc.Analyze(initialWizard)

	moduleLs := disc.ModuleList(initialWizard)
	distModules := []DistributedModule{}

	for _, modName := range moduleLs {
		// CompiledIOP of Segment; it mainly represent the queries over the segment
		// Due to the fair distribution over a module. the CompiledIOP for the segments from the same module are the same.
		// Compile every dist module with the same sequence of compilation steps for uniformity.
		distMod := extractDistModule(initialWizard, disc, modName, maxSegmentSize, maxNumSegment)
		distModules = append(distModules, distMod)
	}

	// it output a compiledIOP object declaring the queries for the consistency among segments/modules
	aggr := aggregator(distModules, maxNumSegment)

	return DistributedWizard{
		Bootstrapper:       initialWizard,
		DistributedModules: distModules,
		Aggregator:         aggr,
	}
}

// it should scan comp and based on module name build compiledIOP for LPP and for GL.
func extractDistModule(
	comp *wizard.CompiledIOP, disc ModuleDiscoverer,
	moduleName moduleName,
	maxSegmentSize, maxNumSegment int,
) DistributedModule {
	panic("unimplemented")
}

// It builds a CompiledIOP object that contains the consistency checks among the segments.
func aggregator(distModules []DistributedModule, maxNumSegments int) *wizard.CompiledIOP {
	panic("unimplemented")
}
