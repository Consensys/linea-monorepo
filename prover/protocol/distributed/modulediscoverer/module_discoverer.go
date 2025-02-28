package modulediscoverer

import (
	"fmt"
	"sync"

	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/symbolic"
)

type ModuleName string

// ModuleDiscoverer defines methods responsible for horizontal splitting (i.e., splitting into modules).
type ModuleDiscoverer interface {
	// Analyze computes how to group columns into modules.
	Analyze(comp *wizard.CompiledIOP)
	// ModuleList returns the list of module names.
	ModuleList() []ModuleName
	// FindModule returns the module corresponding to a column.
	FindModule(col ifaces.Column) ModuleName
	// NewSizeOf returns the split-size of a column in the module.
	NewSizeOf(col ifaces.Column) int

	ExpressionIsInModule(*symbolic.Expression, ModuleName) bool
	QueryIsInModule(ifaces.Query, ModuleName) bool
	ColumnIsInModule(col ifaces.Column, name ModuleName) bool
}

// DisjointSet represents a union-find data structure, which efficiently groups elements (columns)
// into disjoint sets (modules). It supports fast union and find operations with path compression.
type DisjointSet struct {
	parent map[ifaces.Column]ifaces.Column // Maps a column to its representative parent.
	rank   map[ifaces.Column]int           // Stores the rank (tree depth) for optimization.
}

// NewDisjointSet initializes a new DisjointSet with empty mappings.
func NewDisjointSet() *DisjointSet {
	return &DisjointSet{
		parent: make(map[ifaces.Column]ifaces.Column),
		rank:   make(map[ifaces.Column]int),
	}
}

// Find returns the representative (root) of a column using path compression for optimization.
// Path compression ensures that the structure remains nearly flat, reducing the time complexity to O(α(n)),
// where α(n) is the inverse Ackermann function, which is nearly constant in practice.
//
// Example:
// Suppose we have the following sets:
//
//	A -> B -> C (C is the root)
//	D -> E -> F (F is the root)
//
// Calling Find(A) will compress the path so that:
//
//	A -> C
//	B -> C
//	C remains the root
//
// Similarly, calling Find(D) will compress the path so that:
//
//	D -> F
//	E -> F
//	F remains the root
func (ds *DisjointSet) Find(col ifaces.Column) ifaces.Column {
	if _, exists := ds.parent[col]; !exists {
		ds.parent[col] = col
		ds.rank[col] = 0
	}
	if ds.parent[col] != col {
		ds.parent[col] = ds.Find(ds.parent[col])
	}
	return ds.parent[col]
}

// Union merges two sets by linking the root of one to the root of another, optimizing with rank.
// The smaller tree is always attached to the larger tree to keep the depth minimal.
//
// Time Complexity: O(α(n)) (nearly constant due to path compression and union by rank).
//
// Example:
// Suppose we have:
//
//	Set 1: A -> B (B is the root)
//	Set 2: C -> D (D is the root)
//
// Calling Union(A, C) will merge the sets:
//
//	If B has a higher rank than D:
//	    D -> B
//	    C -> D -> B
//	If D has a higher rank than B:
//	    B -> D
//	    A -> B -> D
//	If B and D have equal rank:
//	    D -> B (or B -> D)
//	    Rank of the new root increases by 1
func (ds *DisjointSet) Union(col1, col2 ifaces.Column) {
	root1 := ds.Find(col1)
	root2 := ds.Find(col2)

	if root1 != root2 {
		if ds.rank[root1] > ds.rank[root2] {
			ds.parent[root2] = root1
		} else if ds.rank[root1] < ds.rank[root2] {
			ds.parent[root1] = root2
		} else {
			ds.parent[root2] = root1
			ds.rank[root1]++
		}
	}
}

// Module represents a set of columns grouped by constraints.
type Module struct {
	moduleName ModuleName
	ds         *DisjointSet // Uses a disjoint set to track relationships among columns.
	size       int
	numColumns int
}

// Discoverer tracks modules using DisjointSet.
type Discoverer struct {
	mutex           sync.Mutex
	modules         []*Module
	moduleNames     []ModuleName
	columnsToModule map[ifaces.Column]ModuleName
}

// NewDiscoverer initializes a new Discoverer.
func NewDiscoverer() *Discoverer {
	return &Discoverer{
		modules:         []*Module{},
		moduleNames:     []ModuleName{},
		columnsToModule: make(map[ifaces.Column]ModuleName),
	}
}

// CreateModule initializes a new module with a disjoint set and populates it with columns.
func (disc *Discoverer) CreateModule(columns []ifaces.Column) *Module {
	module := &Module{
		moduleName: ModuleName(fmt.Sprintf("Module_%d", len(disc.modules))),
		ds:         NewDisjointSet(),
	}
	for _, col := range columns {
		module.ds.parent[col] = col
		module.ds.rank[col] = 0
		fmt.Println("Assigned parent for column:", col)
	}
	for i := 0; i < len(columns); i++ {
		for j := i + 1; j < len(columns); j++ {
			module.ds.Union(columns[i], columns[j])
		}
	}
	fmt.Println("Final parent map for module:", module.moduleName, module.ds.parent)
	disc.moduleNames = append(disc.moduleNames, module.moduleName)
	disc.modules = append(disc.modules, module)
	return module
}

// MergeModules merges a list of overlapping modules into a single module.
func (disc *Discoverer) MergeModules(modules []*Module, moduleCandidates *[]*Module) *Module {
	if len(modules) == 0 {
		return nil
	}

	// Select the first module as the base
	mergedModule := modules[0]

	// Merge all remaining modules into the base
	for _, module := range modules[1:] {
		for col := range module.ds.parent {
			mergedModule.ds.Union(mergedModule.ds.Find(col), col)
		}

		// Remove merged module from moduleCandidates
		*moduleCandidates = removeModule(*moduleCandidates, module)
	}

	return mergedModule
}

// AddColumnsToModule adds columns to an existing module.
func (disc *Discoverer) AddColumnsToModule(module *Module, columns []ifaces.Column) {
	for _, col := range columns {
		module.ds.parent[col] = col
		module.ds.rank[col] = 0
		module.ds.Union(module.ds.Find(columns[0]), col) // Union with the first column
	}
}

// Helper function to remove a module from the slice
func removeModule(modules []*Module, target *Module) []*Module {
	var updatedModules []*Module
	for _, mod := range modules {
		if mod != target {
			updatedModules = append(updatedModules, mod)
		}
	}
	return updatedModules
}

// Analyze processes columns and assigns them to modules.

// {1,2,3,4,5}
// {100}
// {6,7,8}
// {9,10}
// {3,6,20}
// {2,99}

// Processing:
// First Iteration - {1,2,3,4,5}
// No existing module.
// Create Module_0 → {1,2,3,4,5}
// Assign columns {1,2,3,4,5} to Module_0.

// Second Iteration - {100}
// No overlap with existing modules.
// Create Module_1 → {100}
// Assign {100} to Module_1.

// Third Iteration - {6,7,8}
// No overlap.
// Create Module_2 → {6,7,8}
// Assign {6,7,8} to Module_2.

// Fourth Iteration - {9,10}
// No overlap.
// Create Module_3 → {9,10}
// Assign {9,10} to Module_3.

// Fifth Iteration - {3,6,20}
// {3} is in Module_0, {6} is in Module_2 → Overlap detected.
// Merge Module_0 and Module_2 into Module_0.
// Module_0 now contains {1,2,3,4,5,6,7,8,20}.
// Remove Module_2 from moduleCandidates.
// Assign {3,6,20} to Module_0.

// Sixth Iteration - {2,99}
// {2} is in Module_0 → Overlap detected.
// Add {99} to Module_0.
// Module_0 now contains {1,2,3,4,5,6,7,8,20,99}.
// Assign {2,99} to Module_0.

// Final Modules:
// Module_0 → {1,2,3,4,5,6,7,8,20,99}
// Module_1 → {100}
// Module_3 → {9,10}

func (disc *Discoverer) Analyze(comp *wizard.CompiledIOP) {
	disc.mutex.Lock()
	defer disc.mutex.Unlock()

	moduleCandidates := []*Module{}

	for _, qName := range comp.QueriesNoParams.AllUnignoredKeys() {
		cs, ok := comp.QueriesNoParams.Data(qName).(query.GlobalConstraint)
		if !ok {
			continue // Skip non-global constraints
		}

		columns := getColumnsFromQuery(cs)
		overlappingModules := []*Module{}

		// Find overlapping modules
		for _, module := range moduleCandidates {
			if HasOverlap(module, columns) {
				overlappingModules = append(overlappingModules, module)
			}
		}

		var assignedModule *Module

		// Merge if necessary
		if len(overlappingModules) > 0 {
			assignedModule = disc.MergeModules(overlappingModules, &moduleCandidates)
			disc.AddColumnsToModule(assignedModule, columns)
		} else {
			// Create a new module
			assignedModule = disc.CreateModule(columns)
			moduleCandidates = append(moduleCandidates, assignedModule)
		}
	}

	// Assign final module names after all processing
	for _, module := range moduleCandidates {
		for col := range module.ds.parent {
			disc.columnsToModule[col] = module.moduleName
		}
	}
}

// getColumnsFromQuery extracts columns from a global constraint query.
func getColumnsFromQuery(q ifaces.Query) []ifaces.Column {
	gc, ok := q.(query.GlobalConstraint)
	if !ok {
		return nil // Not a global constraint, return nil
	}

	// Extract columns from the constraint expression
	var columns []ifaces.Column
	board := gc.Expression.Board()
	for _, metadata := range board.ListVariableMetadata() {
		if col, ok := metadata.(ifaces.Column); ok {
			columns = append(columns, col)
		}
	}

	return columns
}

// assignModule assigns a module name to a set of columns.
func (disc *Discoverer) assignModule(moduleName ModuleName, columns []ifaces.Column) {
	for _, col := range columns {
		disc.columnsToModule[col] = moduleName
	}
}

// NewSizeOf returns the size (length) of a column.
func (disc *Discoverer) NewSizeOf(col ifaces.Column) int {
	return col.Size()
}

// ModuleList returns a list of all module names.
func (disc *Discoverer) ModuleList() []ModuleName {
	disc.mutex.Lock()
	defer disc.mutex.Unlock()
	return disc.moduleNames
}

// ModuleOf returns the module name for a given column.
func (disc *Discoverer) ModuleOf(col ifaces.Column) ModuleName {
	disc.mutex.Lock()
	defer disc.mutex.Unlock()

	if moduleName, exists := disc.columnsToModule[col]; exists {
		return moduleName
	}
	return ""
}

// HasOverlap checks if a module shares at least one column with a set of columns.
func HasOverlap(module *Module, columns []ifaces.Column) bool {
	for _, col := range columns {
		fmt.Println("Checking column:", col, "against module:", module.moduleName)
		if _, exists := module.ds.parent[col]; exists {
			fmt.Println("Overlap found between:", col, "and module:", module.moduleName)
			return true
		}
	}
	return false
}

// NbModules returns the total number of discovered modules.
func (disc *Discoverer) NbModules() int {
	disc.mutex.Lock()
	defer disc.mutex.Unlock()
	return len(disc.moduleNames)
}
