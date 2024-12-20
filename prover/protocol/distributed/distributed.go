package distributed

import (
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/innerproduct"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/mimc"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/specialqueries"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

type ModuleName = string

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
	NbModules() int
	ModuleList(comp *wizard.CompiledIOP) []ModuleName
	FindModule(col ifaces.Column) ModuleName
	// given a query and a module name it checks if the query is inside the module
	QueryIsInModule(ifaces.Query, ModuleName) bool
}

// This transforms the initial wizard. So it is not really the initial
// wizard anymore. That means the caller can forget about "initialWizard"
// after calling the function.
// maxSegmentSize is a static parameter for the max size of the columns in segments.
// maxNumSegment give the max number of segments in a module.
func Distribute(initialWizard *wizard.CompiledIOP, disc ModuleDiscoverer, maxSegmentSize, maxNumSegment int) DistributedWizard {

	prepare(initialWizard)
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
	moduleName ModuleName,
	maxSegmentSize, maxNumSegment int,
) DistributedModule {
	// initialize  two compiledIOPs, for LPP and GL.
	disModule := DistributedModule{
		LookupPermProj: &wizard.CompiledIOP{},
		GlobalLocal:    &wizard.CompiledIOP{},
	}

	for _, qName := range comp.QueriesNoParams.AllUnignoredKeys() {

		q := comp.QueriesNoParams.Data(qName)
		if disc.QueryIsInModule(q, moduleName) {
			continue
		}

		switch v := q.(type) {
		// Filter LPP queries
		case query.Inclusion:
			addToLookupPermProj(disModule.LookupPermProj, v)

		case query.Permutation:
			addToLookupPermProj(disModule.LookupPermProj, v)

		case query.Projection:
			addToLookupPermProj(disModule.LookupPermProj, v)

			// Filter LG queries
		case query.GlobalConstraint:
			addToGlobalLocal(disModule.GlobalLocal, v)

		case query.LocalConstraint:
			addToGlobalLocal(disModule.GlobalLocal, v)

		case query.LocalOpening:
			addToGlobalLocal(disModule.GlobalLocal, v)

		default:
			// Handle other types if necessary
			panic("Other type queries are not handled")
		}

	}
	return disModule

}

func addToLookupPermProj(comp *wizard.CompiledIOP, q ifaces.Query) {
	panic("unimplemented")
}

func addToGlobalLocal(comp *wizard.CompiledIOP, q ifaces.Query) {
	panic("unimplemented")
}

// It builds a CompiledIOP object that contains the consistency checks among the segments.
func aggregator(distModules []DistributedModule, maxNumSegments int) *wizard.CompiledIOP {
	panic("unimplemented")
}

// prepare reduces any query to LPP or GL.
// it prepares the columns that depends on whole the witness,e.g., M column for lookups.
func prepare(comp *wizard.CompiledIOP) {
	mimc.CompileMiMC(comp)
	specialqueries.RangeProof(comp)
	specialqueries.CompileFixedPermutations(comp)
	innerproduct.Compile(comp)

	// prepareLookup(comp)
}
