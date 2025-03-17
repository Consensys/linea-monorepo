package experiment

import (
	"fmt"
	"math"
	"sort"
	"sync"
	"unsafe"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/internal/plonkinternal"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/protocol/wizardutils"
	"github.com/consensys/linea-monorepo/prover/utils"
)

// StandardModuleDiscoverer groups modules using two layers. In the first layer,
// the [QueryBasedModuleDiscoverer] groups columns that are part of the same
// queries (for global, local, plonk, log-derivative, grand-product and horner
// queries).
type StandardModuleDiscoverer struct {
	// TargetWeight is the target weight for each module.
	TargetWeight    int
	modules         []*StandardModule
	columnsToModule map[ifaces.Column]ModuleName
	columnsToSize   map[ifaces.Column]int
}

// QueryBasedModuleDiscoverer tracks modules using DisjointSet.
type QueryBasedModuleDiscoverer struct {
	mutex           sync.Mutex
	modules         []*QueryBasedModule
	moduleNames     []ModuleName
	columnsToModule map[ifaces.Column]ModuleName
}

// StandardModule is a structure coalescing a set of [QueryBasedModule]s.
type StandardModule struct {
	moduleName ModuleName
	subModules []*QueryBasedModule
	newSizes   []int
}

// QueryBasedModule represents a set of columns grouped by constraints.
type QueryBasedModule struct {
	moduleName ModuleName
	ds         *DisjointSet[column.Natural] // Uses a disjoint set to track relationships among columns.
	size       int
	// nbConstraintsOfPlonkCirc counts the number of constraints in a Plonk
	// in wizard module if one is found. If several circuits are stores
	// the number is the sum for all the circuits where the nb constraint
	// for circuit is padded to the next power of two. For instance, if
	// we have a circuit with 5 constraints, and one with 33 constraints
	// the value of nbConstraintsPlonk will be 64 + 8 = 72.
	nbConstraintsOfPlonkCirc int
	// nbInstancesOfPlonkCirc indicates the max number of instances in a
	// Plonk in wizard module if one is found.
	nbInstancesOfPlonkCirc int
	// nbInstancesOfPlonkQuery indicates the number of Plonk in wizard query in the
	// present module.
	nbInstancesOfPlonkQuery int
	// nbSegmentCache caches the results of SegmentBoundaries
	nbSegmentCache map[unsafe.Pointer][2]int
}

// DisjointSet represents a union-find data structure, which efficiently groups elements (columns)
// into disjoint sets (modules). It supports fast union and find operations with path compression.
type DisjointSet[T comparable] struct {
	parent map[T]T   // Maps a column to its representative parent.
	rank   map[T]int // Stores the rank (tree depth) for optimization.
}

// NewQueryBasedDiscoverer initializes a new Discoverer.
func NewQueryBasedDiscoverer() *QueryBasedModuleDiscoverer {
	return &QueryBasedModuleDiscoverer{
		modules:         []*QueryBasedModule{},
		moduleNames:     []ModuleName{},
		columnsToModule: make(map[ifaces.Column]ModuleName),
	}
}

// Analyze processes scans the comp and generate the modules. It works by
// grouping columns that are part of the same query using the [QueryBasedModuleDiscoverer].
// After that it sorts the generated modules by weight and
func (disc *StandardModuleDiscoverer) Analyze(comp *wizard.CompiledIOP) {

	subDiscover := NewQueryBasedDiscoverer()
	subDiscover.Analyze(comp)

	modulesQBased := subDiscover.modules

	sort.Slice(modulesQBased, func(i, j int) bool {
		return modulesQBased[i].Weight(0) < modulesQBased[j].Weight(0)
	})

	var (
		groups        = [][]*QueryBasedModule{{}}
		currWeightSum = 0
	)

	for i := range modulesQBased {

		var (
			moduleNext = modulesQBased[i]
			weightNext = moduleNext.Weight(0)
			distPrev   = utils.Abs(disc.TargetWeight - currWeightSum)
			distNext   = utils.Abs(disc.TargetWeight - currWeightSum - weightNext)
		)

		if weightNext == 0 {
			continue
		}

		if distNext > distPrev {
			groups = append(groups, []*QueryBasedModule{})
			currWeightSum = 0
		}

		currWeightSum += weightNext
		groups[len(groups)-1] = append(groups[len(groups)-1], moduleNext)
	}

	disc.modules = make([]*StandardModule, len(groups))

	for i := range disc.modules {

		disc.modules[i] = &StandardModule{
			moduleName: ModuleName(fmt.Sprintf("Module_%d", i)),
			subModules: groups[i],
			newSizes:   make([]int, len(groups[i])),
		}

		// weightTotalInitial is the weight of the module using the initial
		// number of rows.
		weightTotalInitial := 0
		// maxNumRows is the number of rows of the "shallowest" submodule of in
		// group[i]
		minNumRows := math.MaxInt

		// initializes the newSizes using the number of rows from the initial
		// comp.
		for j := range groups[i] {
			numRows := groups[i][j].NumRow()
			disc.modules[i].newSizes[j] = numRows
			minNumRows = min(minNumRows, numRows)
			weightTotalInitial += groups[i][j].Weight(0)
		}

		// Computes optimal newSizes by successively dividing by two the numbers
		// of rows and checks if this brings the total weight of the module to
		// the target weight.
		var (
			bestReduction = 1
			bestWeight    = weightTotalInitial
		)

		for reduction := 2; reduction < minNumRows; reduction *= 2 {

			currWeight := 0
			for j := range groups[i] {
				numRow := groups[i][j].NumRow() / reduction
				if numRow < 1 {
					panic("the 'reduction' is bounded by the min number of rows so it should not be smaller than 1")
				}
				currWeight += groups[i][j].Weight(numRow)
			}

			currDist := utils.Abs(currWeight - disc.TargetWeight)
			bestDist := utils.Abs(bestWeight - disc.TargetWeight)

			if currDist < bestDist {
				bestReduction = reduction
				bestWeight = currWeight
			}
		}

		for j := range groups[i] {
			disc.modules[i].newSizes[j] /= bestReduction
		}
	}

	disc.columnsToModule = make(map[ifaces.Column]ModuleName)
	disc.columnsToSize = make(map[ifaces.Column]int)

	for i := range disc.modules {

		moduleName := disc.modules[i].moduleName

		for j := range disc.modules[i].subModules {

			subModule := disc.modules[i].subModules[j]
			newSize := disc.modules[i].newSizes[j]

			for col := range subModule.ds.parent {
				disc.columnsToModule[col] = moduleName
				disc.columnsToSize[col] = newSize
			}
		}
	}
}

// ModuleList returns a list of all module names.
func (disc *StandardModuleDiscoverer) ModuleList() []ModuleName {
	modulesNames := make([]ModuleName, len(disc.modules))
	for i := range disc.modules {
		modulesNames[i] = disc.modules[i].moduleName
	}
	return modulesNames
}

// NumColumnOf counts the number of columns found for the current module
func (disc *StandardModuleDiscoverer) NumColumnOf(moduleName ModuleName) int {

	for i := range disc.modules {

		if disc.modules[i].moduleName != moduleName {
			continue
		}

		res := 0
		for j := range disc.modules[i].subModules {
			res += disc.modules[i].subModules[j].NumColumn()
		}
		return res
	}

	utils.Panic("module not found")
	return 0
}

// ModuleOf returns the module name for a given column“
func (disc *StandardModuleDiscoverer) ModuleOf(col column.Natural) ModuleName {
	return disc.columnsToModule[col]
}

// NewSizeOf returns the size (length) of a column.
func (disc *StandardModuleDiscoverer) NewSizeOf(col column.Natural) int {
	return disc.columnsToSize[col]
}

// SegmentBoundaryOfColumn returns the starting point and the ending point of the
// segmentation of column. The implementation works by identifying the corresponding
// [StandardModule] and then the corresponding inner [QueryBasedModule]. The function
// will then scans the assignment of the columns in the query based module and examine
// their padding to return the boundaries of the segmented area.
//
// To clarify, the segmentation of the column corresponds to the part that will be
// kept by the segmentation process of the column (e.g. the area covered by the
// reunion of all the segments). The rest of the assignment corresponds to padding.
func (disc *StandardModuleDiscoverer) SegmentBoundaryOf(run *wizard.ProverRuntime, col column.Natural) (int, int) {

	var (
		rootCol          = column.RootParents(col).(column.Natural)
		stdModuleName    = disc.columnsToModule[col]
		stdModule        *StandardModule
		queryBasedModule *QueryBasedModule
		segmentSize      int
	)

	for i := range disc.modules {
		if disc.modules[i].moduleName == stdModuleName {
			stdModule = disc.modules[i]
			break
		}
	}

	for i := range stdModule.subModules {
		if _, ok := stdModule.subModules[i].ds.parent[rootCol]; ok {
			segmentSize = stdModule.newSizes[i]
			queryBasedModule = stdModule.subModules[i]
			break
		}
	}

	return queryBasedModule.SegmentBoundaries(run, segmentSize)
}

// Analyze processes columns and assigns them to modules.
//
// {1,2,3,4,5}
// {100}
// {6,7,8}
// {9,10}
// {3,6,20}
// {2,99}
//
// Processing:
// First Iteration - {1,2,3,4,5}
// No existing module.
// Create Module_0 → {1,2,3,4,5}
// Assign columns {1,2,3,4,5} to Module_0.
//
// Second Iteration - {100}
// No overlap with existing modules.
// Create Module_1 → {100}
// Assign {100} to Module_1.
//
// Third Iteration - {6,7,8}
// No overlap.
// Create Module_2 → {6,7,8}
// Assign {6,7,8} to Module_2.
//
// Fourth Iteration - {9,10}
// No overlap.
// Create Module_3 → {9,10}
// Assign {9,10} to Module_3.
//
// Fifth Iteration - {3,6,20}
// {3} is in Module_0, {6} is in Module_2 → Overlap detected.
// Merge Module_0 and Module_2 into Module_0.
// Module_0 now contains {1,2,3,4,5,6,7,8,20}.
// Remove Module_2 from moduleCandidates.
// Assign {3,6,20} to Module_0.
//
// Sixth Iteration - {2,99}
// {2} is in Module_0 → Overlap detected.
// Add {99} to Module_0.
// Module_0 now contains {1,2,3,4,5,6,7,8,20,99}.
// Assign {2,99} to Module_0.
//
// Final Modules:
// Module_0 → {1,2,3,4,5,6,7,8,20,99}
// Module_1 → {100}
// Module_3 → {9,10}
func (disc *QueryBasedModuleDiscoverer) Analyze(comp *wizard.CompiledIOP) {
	disc.mutex.Lock()
	defer disc.mutex.Unlock()

	disc.columnsToModule = make(map[ifaces.Column]ModuleName)

	moduleCandidates := []*QueryBasedModule{}

	for _, qName := range comp.QueriesNoParams.AllKeys() {

		queryData := comp.QueriesNoParams.Data(qName)

		// Permutation queries are expectedly already compiled into a GdProduct
		// queries. Still we still need to group the columns into the same modules
		// because otherwise we can have a situation where both side of the permutation
		// have  different number of segments and different sizes for the segment, causing
		// the gd-product query to fail.
		if perm, ok := queryData.(query.Permutation); ok {

			group := []column.Natural{}

			// If it is a permutation, we need to group the columns that are used
			// in the permutation. Both sides must be in the same module.
			for i := range perm.A {
				group = append(group, rootsOfColumns(perm.A[i])...)
			}

			for i := range perm.B {
				group = append(group, rootsOfColumns(perm.B[i])...)
			}

			moduleCandidates = disc.GroupColumns(group, moduleCandidates, 0, 0, 0)
			continue
		}

		if comp.QueriesNoParams.IsIgnored(qName) {
			continue
		}

		var (
			// toGroup lists sets of columns who need to be grouped.
			toGroup = [][]column.Natural{}

			// nbConstraintsOfPlonkCirc tells if there are Plonk-in-Wizard overheads
			// associated with the modules columns. And how many constraints there
			// are.
			nbConstraintsOfPlonkCirc = 0
			nbInstancesOfPlonkCirc   = 0
			nbInstancesOfPlonkQuery  = 0
		)

		switch q := queryData.(type) {

		default:
			utils.Panic("unexpected query: type=%T name=%v", q, qName)

		case query.Range:
			// The query is expected but no grouping required.

		case query.GlobalConstraint:
			cols := wizardutils.ColumnsOfExpression(q.Expression)
			toGroup = append(toGroup, rootsOfColumns(cols))

		case query.LocalConstraint:
			cols := wizardutils.ColumnsOfExpression(q.Expression)
			toGroup = append(toGroup, rootsOfColumns(cols))

		case *query.PlonkInWizard:
			// Note: [q.CircuitMask] might be "nil" and it is ok. This is the reason
			// why we need that [column.RootsOf] filters out 'nil' inputs.
			cols := []ifaces.Column{q.Selector, q.Data}
			toGroup = append(toGroup, rootsOfColumns(cols))
			// Since there is only one possible option, we know that it can
			// only be a range-check.
			nbConstraintsOfPlonkCirc = utils.NextPowerOfTwo(countConstraintsOfPlonkCirc(q))
			nbInstancesOfPlonkCirc = q.GetMaxNbCircuitInstances()
			nbInstancesOfPlonkQuery = 1
		}

		for _, columns := range toGroup {
			moduleCandidates = disc.GroupColumns(
				columns,
				moduleCandidates,
				nbConstraintsOfPlonkCirc,
				nbInstancesOfPlonkCirc,
				nbInstancesOfPlonkQuery,
			)
		}
	}

	for _, qName := range comp.QueriesParams.AllUnignoredKeys() {

		var (
			// toGroup lists sets of columns who need to be grouped.
			toGroup = [][]column.Natural{}

			// nbConstraintsOfPlonkCirc tells if there are Plonk-in-Wizard overheads
			// associated with the modules columns. And how many constraints there
			// are.
			nbConstraintsOfPlonkCirc = 0
			nbInstancesOfPlonkCirc   = 0
			nbInstancesOfPlonkQuery  = 0
		)

		switch q := comp.QueriesParams.Data(qName).(type) {

		default:
			utils.Panic("unexpected query: type=%T name=%v", q, qName)

		case query.LocalOpening:
			// Nothing to do

		case query.LogDerivativeSum:

			sizes := utils.SortedKeysOf(q.Inputs, func(a, b int) bool { return a < b })
			for _, size := range sizes {
				inpForSize := q.Inputs[size]
				for i := range inpForSize.Numerator {
					colNums := wizardutils.ColumnsOfExpression(inpForSize.Numerator[i])
					colDens := wizardutils.ColumnsOfExpression(inpForSize.Denominator[i])
					rootNums, rootDens := rootsOfColumns(colNums), rootsOfColumns(colDens)
					toGroup = append(toGroup, append(rootNums, rootDens...))
				}
			}

		// Note: this is case is already super-seded by our handling of the permutation
		// queries. So this should not bring any additional grouping.
		case query.GrandProduct:
			sizes := utils.SortedKeysOf(q.Inputs, func(a, b int) bool { return a < b })
			for _, size := range sizes {
				inpForSize := q.Inputs[size]
				for i := range inpForSize.Numerators {
					colNums := wizardutils.ColumnsOfExpression(inpForSize.Numerators[i])
					colDens := wizardutils.ColumnsOfExpression(inpForSize.Denominators[i])
					rootNums, rootDens := rootsOfColumns(colNums), rootsOfColumns(colDens)
					toGroup = append(toGroup, append(rootNums, rootDens...))
				}
			}

		case *query.Horner:
			for _, part := range q.Parts {
				group := append(wizardutils.ColumnsOfExpression(part.Coefficient), part.Selector)
				toGroup = append(toGroup, rootsOfColumns(group))
			}
		}

		for _, columns := range toGroup {
			moduleCandidates = disc.GroupColumns(
				columns,
				moduleCandidates,
				nbConstraintsOfPlonkCirc,
				nbInstancesOfPlonkCirc,
				nbInstancesOfPlonkQuery,
			)
		}
	}

	// Remove empty modules
	disc.RemoveNils()

	// Assign final module names after all processing
	for _, module := range disc.modules {
		for col := range module.ds.parent {
			disc.columnsToModule[col] = module.moduleName
		}
	}
}

// RemoveNils remove the empty modules from the list of modules.
func (disc *QueryBasedModuleDiscoverer) RemoveNils() {
	for i := len(disc.modules) - 1; i >= 0; i-- {
		if len(disc.modules[i].ds.parent) == 0 {
			disc.modules = append(disc.modules[:i], disc.modules[i+1:]...)
		}
	}
}

// CreateModule initializes a new module with a disjoint set and populates it with columns.
func (disc *QueryBasedModuleDiscoverer) CreateModule(columns []column.Natural) *QueryBasedModule {

	var colID ifaces.ColID
	if len(columns) > 0 {
		colID = columns[0].GetColID()
	} else {
		colID = "default" // columns is empty
	}

	module := &QueryBasedModule{
		moduleName: ModuleName(fmt.Sprintf("Module_%d_%s", len(disc.modules), colID)),
		ds:         NewDisjointSet[column.Natural](),
	}
	for _, col := range columns {
		module.ds.parent[col] = col
		module.ds.rank[col] = 0
	}
	for i := 0; i < len(columns); i++ {
		for j := i + 1; j < len(columns); j++ {
			module.ds.Union(columns[i], columns[j])
		}
	}

	disc.moduleNames = append(disc.moduleNames, module.moduleName)
	disc.modules = append(disc.modules, module)
	return module
}

// MergeModules merges a list of overlapping modules into a single module.
func (disc *QueryBasedModuleDiscoverer) MergeModules(modules []*QueryBasedModule, moduleCandidates []*QueryBasedModule) *QueryBasedModule {
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

		mergedModule.nbConstraintsOfPlonkCirc += module.nbConstraintsOfPlonkCirc
		mergedModule.nbInstancesOfPlonkCirc += module.nbInstancesOfPlonkCirc
		mergedModule.nbInstancesOfPlonkQuery += module.nbInstancesOfPlonkQuery

		// nilifying the module ensures it can no longer be matched for anything
		module.Nilify()
	}

	return mergedModule
}

// AddColumnsToModule adds columns to an existing module.
func (disc *QueryBasedModuleDiscoverer) AddColumnsToModule(module *QueryBasedModule, columns []column.Natural) {
	for _, col := range columns {
		module.ds.parent[col] = col
		module.ds.rank[col] = 0
		module.ds.Union(module.ds.Find(columns[0]), col) // Union with the first column
	}
}

// GroupColumns ensures that columns are grouped together, merging
// overlapping modules if necessary.
func (disc *QueryBasedModuleDiscoverer) GroupColumns(
	columns []column.Natural,
	moduleCandidates []*QueryBasedModule,
	nbConstraintsOfPlonkCirc int,
	nbInstancesOfPlonkCirc int,
	nbInstancesOfPlonkQuery int,
) []*QueryBasedModule {

	overlappingModules := []*QueryBasedModule{}

	// Find overlapping modules
	for _, module := range moduleCandidates {
		if module.HasOverlap(columns) {
			overlappingModules = append(overlappingModules, module)
		}
	}

	var assignedModule *QueryBasedModule

	// Merge if necessary
	if len(overlappingModules) > 0 {

		assignedModule = disc.MergeModules(overlappingModules, moduleCandidates)
		assignedModule.nbConstraintsOfPlonkCirc += nbConstraintsOfPlonkCirc
		assignedModule.nbInstancesOfPlonkCirc += nbInstancesOfPlonkCirc
		assignedModule.nbInstancesOfPlonkQuery += nbInstancesOfPlonkQuery
		disc.AddColumnsToModule(assignedModule, columns)
	} else {

		// Create a new module
		assignedModule = disc.CreateModule(columns)
		assignedModule.nbConstraintsOfPlonkCirc += nbConstraintsOfPlonkCirc
		assignedModule.nbInstancesOfPlonkCirc += nbInstancesOfPlonkCirc
		assignedModule.nbInstancesOfPlonkQuery += nbInstancesOfPlonkQuery
		moduleCandidates = append(moduleCandidates, assignedModule)
	}

	return moduleCandidates
}

// SegmentBoundaries computes the density of a module given an assignment.
// This can be used to determine the number of segment of the module.
func (mod *QueryBasedModule) SegmentBoundaries(run *wizard.ProverRuntime, segmentSize int) (int, int) {

	if res, ok := mod.nbSegmentCache[unsafe.Pointer(run)]; ok {
		return res[0], res[1]
	}

	var (
		resOrientation = 0
		resMaxDensity  = 0
		size           int
	)

	for col := range mod.ds.parent {

		var (
			val               = col.GetColAssignment(run)
			density           = smartvectors.Density(val)
			orientation, oErr = smartvectors.PaddingOrientationOf(val)
		)

		size = val.Len()

		if oErr != nil {
			continue
		}

		if density == size {
			continue
		}

		if resOrientation != 0 {
			resOrientation = orientation
		}

		if orientation != resOrientation {
			panic("conflicting orientation")
		}

		resMaxDensity = max(resMaxDensity, density)
	}

	var (
		totalSegmentedArea = segmentSize * utils.DivCeil(resMaxDensity, segmentSize)
		start, stop        = 0, totalSegmentedArea
	)

	if resOrientation == -1 {
		start, stop = size-stop, size-start
	}

	mod.nbSegmentCache[unsafe.Pointer(run)] = [2]int{start, stop}
	return start, stop
}

// Nilify a module. It empties its maps and sets its size to 0.
func (module *QueryBasedModule) Nilify() {
	module.ds.parent = map[column.Natural]column.Natural{}
	module.ds.rank = map[column.Natural]int{}
	module.size = 0
	module.nbConstraintsOfPlonkCirc = 0
	module.nbInstancesOfPlonkCirc = 0
	module.nbInstancesOfPlonkQuery = 0
}

// HasOverlap checks if a module shares at least one column with a set of columns.
func (module *QueryBasedModule) HasOverlap(columns []column.Natural) bool {
	for _, col := range columns {
		if _, exists := module.ds.parent[col]; exists {
			return true
		}
	}
	return false
}

// Weight returns the weight metric of a module. The returned metric corresponds to
// a number of witness elements that the module represents. The function takes an
// optional "withNumRow" arguments to let the caller use a custom hypothetical value for
// [nbRows] in the calculation. If withNumRow=0, then the function uses the actual
// number of rows.
func (module *QueryBasedModule) Weight(withNumRow int) int {

	var (
		numRow             = module.NumRow()
		numCol             = module.NumColumn()
		numOfPlonkInstance = module.nbInstancesOfPlonkCirc
	)

	if withNumRow > 0 {
		numOfPlonkInstance = numOfPlonkInstance * withNumRow / numRow
		numRow = withNumRow
	}

	// The 4 and 11 are heuristic parameters to estimate the actual witness complexity
	// of having Plonk in wizards in the module.
	plonkCost := (4*numOfPlonkInstance + 11*module.nbInstancesOfPlonkQuery) * module.nbConstraintsOfPlonkCirc

	return numCol*numRow + plonkCost
}

// Weight returns the total weight of the module
func (module *StandardModule) Weight() int {
	weight := 0
	for i := range module.subModules {
		numRow := module.newSizes[i]
		weight += module.subModules[i].Weight(numRow)
	}
	return weight
}

// NumRow returns the number of rows for the module
func (module *QueryBasedModule) NumRow() int {
	if module.size == 0 {
		for col := range module.ds.parent {
			module.size = col.Size()
			break
		}
	}
	return module.size
}

// NumColumn returns the number of columns for the module
func (module *QueryBasedModule) NumColumn() int {
	return len(module.ds.parent)
}

// NewSizeOf returns the size (length) of a column.
func (disc *QueryBasedModuleDiscoverer) NewSizeOf(col column.Natural) int {
	return col.Size()
}

// ModuleList returns a list of all module names.
func (disc *QueryBasedModuleDiscoverer) ModuleList() []ModuleName {
	disc.mutex.Lock()
	defer disc.mutex.Unlock()
	return disc.moduleNames
}

// ModuleOf returns the module name for a given column.
func (disc *QueryBasedModuleDiscoverer) ModuleOf(col column.Natural) ModuleName {
	disc.mutex.Lock()
	defer disc.mutex.Unlock()

	if moduleName, exists := disc.columnsToModule[col]; exists {
		return moduleName
	}
	return ""
}

// NewDisjointSet initializes a new DisjointSet with empty mappings.
func NewDisjointSet[T comparable]() *DisjointSet[T] {
	return &DisjointSet[T]{
		parent: make(map[T]T),
		rank:   make(map[T]int),
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
func (ds *DisjointSet[T]) Find(col T) T {
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
func (ds *DisjointSet[T]) Union(col1, col2 T) {
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

// countConstraintsOfPlonkCirc returns the number of constraint of the circuit. The function is
// quite expensive as this requires to compile the circuit. So it's better to cache
// the result when possible.
func countConstraintsOfPlonkCirc(piw *query.PlonkInWizard) int {

	hasAddGates := false
	if len(piw.PlonkOptions) > 0 {
		hasAddGates = piw.PlonkOptions[0].RangeCheckAddGateForRangeCheck
	}

	ccs, _, _ := plonkinternal.CompileCircuit(piw.Circuit, hasAddGates)
	nbConstraints := ccs.GetNbConstraints()
	return nbConstraints
}

// rootsOfColumns returns a clean and deduplicated list of the roots of the
// columns as [column.Natural] and not [ifaces.Column]
func rootsOfColumns(cols []ifaces.Column) []column.Natural {
	roots := column.RootsOf(cols, true)
	nats := make([]column.Natural, len(roots))
	for i := range roots {
		nats[i] = roots[i].(column.Natural)
	}
	return nats
}
