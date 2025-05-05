package distributed

import (
	"fmt"
	"math"
	"sort"
	"sync"
	"unsafe"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/distributed/pragmas"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/internal/plonkinternal"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/collection"
)

// StandardModuleDiscoverer groups modules using two layers. In the first layer,
// the [QueryBasedModuleDiscoverer] groups columns that are part of the same
// queries (for global, local, plonk, log-derivative, grand-product and horner
// queries).
type StandardModuleDiscoverer struct {
	// TargetWeight is the target weight for each module.
	TargetWeight int
	// Affinities indicates groups of columns (potentially spanning over
	// multiple query-based modules) that are "alike" in the sense that
	// they would be opportunistic to group in the same StandardModule.
	Affinities [][]column.Natural
	// Predivision indicates that all the inputs column size should be
	// divided by some values before being added in a [QueryBasedModule].
	Predivision     int
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
	predivision     int
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
	ds         *utils.DisjointSet[column.Natural] // Uses a disjoint set to track relationships among columns.
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
	nbSegmentCache      map[unsafe.Pointer][2]int
	nbSegmentCacheMutex *sync.Mutex
	predivision         int
	hasPrecomputed      bool
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
	subDiscover.predivision = disc.Predivision
	subDiscover.Analyze(comp)

	groupedByAffinity := groupQBModulesByAffinity(subDiscover.modules, disc.Affinities)

	sort.Slice(groupedByAffinity, func(i, j int) bool {
		return weightOfGroupOfQBModules(groupedByAffinity[i]) < weightOfGroupOfQBModules(groupedByAffinity[j])
	})

	var (
		groups        = [][]*QueryBasedModule{{}}
		currWeightSum = 0
	)

	for i := range groupedByAffinity {

		var (
			groupNext  = groupedByAffinity[i]
			weightNext = weightOfGroupOfQBModules(groupNext)
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
		groups[len(groups)-1] = append(groups[len(groups)-1], groupNext...)
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
				numRow := groups[i][j].NumRow()
				if !groups[i][j].hasPrecomputed {
					numRow /= reduction
				}

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

			for col := range subModule.ds.Iter() {
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
		if stdModule.subModules[i].ds.Has(rootCol) {
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

		if comp.QueriesNoParams.IsIgnored(qName) {
			continue
		}

		queryData := comp.QueriesNoParams.Data(qName)

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
			cols := column.ColumnsOfExpression(q.Expression)
			toGroup = append(toGroup, rootsOfColumns(cols))

		case query.LocalConstraint:
			cols := column.ColumnsOfExpression(q.Expression)
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

		if disc.predivision > 0 {
			nbInstancesOfPlonkCirc /= disc.predivision
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
					colNums := column.ColumnsOfExpression(inpForSize.Numerator[i])
					colDens := column.ColumnsOfExpression(inpForSize.Denominator[i])
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
					colNums := column.ColumnsOfExpression(inpForSize.Numerators[i])
					colDens := column.ColumnsOfExpression(inpForSize.Denominators[i])
					rootNums, rootDens := rootsOfColumns(colNums), rootsOfColumns(colDens)
					toGroup = append(toGroup, append(rootNums, rootDens...))
				}
			}

		case *query.Horner:
			for _, part := range q.Parts {
				group := append(column.ColumnsOfExpression(part.Coefficient), part.Selector)
				toGroup = append(toGroup, rootsOfColumns(group))
			}
		}

		for _, columns := range toGroup {
			moduleCandidates = disc.GroupColumns(
				columns,
				moduleCandidates,
				0, 0, 0,
			)
		}
	}

	// Remove empty modules
	disc.RemoveNils()

	// Assign final module names after all processing
	for _, module := range disc.modules {

		hasPrecomputed := false

		for col := range module.ds.Iter() {
			disc.columnsToModule[col] = module.moduleName
			hp := col.Status() == column.Precomputed || col.Status() == column.VerifyingKey
			hasPrecomputed = hp || hasPrecomputed
		}

		module.hasPrecomputed = hasPrecomputed
		if !hasPrecomputed {
			module.predivision = disc.predivision
		}
		module.mustHaveConsistentLength()
	}
}

// RemoveNils remove the empty modules from the list of modules.
func (disc *QueryBasedModuleDiscoverer) RemoveNils() {
	for i := len(disc.modules) - 1; i >= 0; i-- {
		if disc.modules[i].ds.Size() == 0 {
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
		moduleName:          ModuleName(fmt.Sprintf("Module_%d_%s", len(disc.modules), colID)),
		ds:                  utils.NewDisjointSetFromList(columns),
		nbSegmentCache:      make(map[unsafe.Pointer][2]int),
		nbSegmentCacheMutex: &sync.Mutex{},
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

		for col := range module.ds.Iter() {
			mergedModule.ds.Union(mergedModule.ds.Find(col), col)
		}

		mergedModule.nbConstraintsOfPlonkCirc += module.nbConstraintsOfPlonkCirc
		mergedModule.nbInstancesOfPlonkCirc += module.nbInstancesOfPlonkCirc
		mergedModule.nbInstancesOfPlonkQuery += module.nbInstancesOfPlonkQuery
		mergedModule.hasPrecomputed = mergedModule.hasPrecomputed || module.hasPrecomputed

		// nilifying the module ensures it can no longer be matched for anything
		module.Nilify()
	}

	return mergedModule
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
		assignedModule.ds.AddList(columns)

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
//
// The function works by iterating over the columns of the module and
// computing the density of each column. The density is defined as the
// number of non-padding elements. The function does a few quirks around
// the regular columns: they are not taken into account when evaluating
// the overall density of the module. That is because, it often happens
// that we suboptimally assign a whole regular column while it could in fact
// have a sparser representation. However, if all columns are regular,
// then we consider that the density is equal to the size of the module.
//
// Beside we consider that constants columns have a density of 0, and the
// function will return 0, 0 if all columns in the module are constants.
func (mod *QueryBasedModule) SegmentBoundaries(run *wizard.ProverRuntime, segmentSize int) (int, int) {

	mod.nbSegmentCacheMutex.Lock()
	if res, ok := mod.nbSegmentCache[unsafe.Pointer(run)]; ok {
		mod.nbSegmentCacheMutex.Unlock()
		return res[0], res[1]
	}
	mod.nbSegmentCacheMutex.Unlock()

	var (
		resMaxDensity     = 0
		size              int
		areAnyNonRegular  = false
		areAnyLeftPadded  = false
		areAnyRightPadded = false
		areAnyFull        = false
		firstLeftPadded   column.Natural
		firstRightPadded  column.Natural
		firstColumn       column.Natural
	)

	for col := range mod.ds.Iter() {

		if size == 0 {
			size = col.Size()
			firstColumn = col
		}

		// This should not happen for a well-formed module, query-based modules
		// are expected to contain the same size for all columns.
		if size != col.Size() {
			utils.Panic("all columns must have the same size, first=%v col=%v", firstColumn.ID, col.ID)
		}

		// As the function is meant to be called **during** the bootstrapping
		// compilation, we cannot expect all columns to be available at this
		// point. This is for instance the case with the lookup "M" columns.
		if !run.HasColumn(col.ID) {
			continue
		}

		var (
			val           = col.GetColAssignment(run)
			start, stop   = smartvectors.CoWindowRange(val)
			isLeftPadded  = start == 0
			isRightPadded = stop == size
			density       = stop - start
			isFullColumn  = pragmas.IsFullColumn(col)
		)

		if isFullColumn {
			areAnyFull = true
			resMaxDensity = density
			break
		}

		if isLeftPadded && isRightPadded {
			continue
		} else {
			areAnyNonRegular = true
		}

		// It is important to exclude constant column from toggling up the
		// areAnyLeftPadded flag. Otherwise, the panic condition stating that
		// there should not be both left and right padded columns will be
		// activated when we mix right-padded with a constant column.
		if isLeftPadded && density > 0 {
			areAnyLeftPadded = true
			firstLeftPadded = col
		}

		if isRightPadded {
			areAnyRightPadded = true
			firstRightPadded = col
		}

		if !isLeftPadded && !isRightPadded {
			utils.Panic("column is neither left nor right padded, col=%v", col.ID)
		}

		if density > resMaxDensity {
			resMaxDensity = density
		}
	}

	if !areAnyNonRegular || areAnyFull {
		start, stop := 0, size
		mod.nbSegmentCacheMutex.Lock()
		defer mod.nbSegmentCacheMutex.Unlock()
		mod.nbSegmentCache[unsafe.Pointer(run)] = [2]int{start, stop}
		return start, stop
	}

	if resMaxDensity == 0 {
		start, stop := 0, 0
		mod.nbSegmentCacheMutex.Lock()
		defer mod.nbSegmentCacheMutex.Unlock()
		mod.nbSegmentCache[unsafe.Pointer(run)] = [2]int{start, stop}
		return start, stop
	}

	if areAnyLeftPadded && areAnyRightPadded {
		utils.Panic("the module cannot contain at the same time left and right padded columns, oneLeftPadded=%v, oneRightPadded=%v", firstLeftPadded.ID, firstRightPadded.ID)
	}

	var (
		localNbSegment     = utils.DivCeil(resMaxDensity, segmentSize)
		totalSegmentedArea = segmentSize * localNbSegment
	)

	if areAnyLeftPadded {
		start, stop := 0, totalSegmentedArea
		mod.nbSegmentCacheMutex.Lock()
		defer mod.nbSegmentCacheMutex.Unlock()
		mod.nbSegmentCache[unsafe.Pointer(run)] = [2]int{start, stop}
		return start, stop
	}

	if areAnyRightPadded {
		start, stop := size-totalSegmentedArea, size
		mod.nbSegmentCacheMutex.Lock()
		defer mod.nbSegmentCacheMutex.Unlock()
		mod.nbSegmentCache[unsafe.Pointer(run)] = [2]int{start, stop}
		return start, stop
	}

	panic("unreachable")
}

// Nilify a module. It empties its maps and sets its size to 0.
func (module *QueryBasedModule) Nilify() {
	module.ds.Reset()
	module.size = 0
	module.nbConstraintsOfPlonkCirc = 0
	module.nbInstancesOfPlonkCirc = 0
	module.nbInstancesOfPlonkQuery = 0
}

// HasOverlap checks if a module shares at least one column with a set of columns.
func (module *QueryBasedModule) HasOverlap(columns []column.Natural) bool {
	for _, col := range columns {
		if module.ds.Has(col) {
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
		for col := range module.ds.Iter() {
			module.size = col.Size()

			if module.hasPrecomputed {
				break
			}

			if module.predivision > col.Size() || module.predivision == 0 {
				break
			}

			module.size /= module.predivision
			break
		}
	}

	return module.size
}

// NumColumn returns the number of columns for the module
func (module *QueryBasedModule) NumColumn() int {
	return module.ds.Size()
}

// NewSizeOf returns the size (length) of a column.
func (disc *QueryBasedModuleDiscoverer) NewSizeOf(col column.Natural) int {
	size := col.Size()

	mod := disc.ModuleOf(col)
	for i := range disc.modules {
		if disc.modules[i].moduleName == mod {
			qbm := disc.modules[i]
			if qbm.hasPrecomputed {
				return size
			}
		}
	}

	if disc.predivision > 0 && disc.predivision < col.Size() {
		return size / disc.predivision
	}

	return size
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

// countConstraintsOfPlonkCirc returns the number of constraint of the circuit. The function is
// quite expensive as this requires to compile the circuit. So it's better to cache
// the result when possible.
func countConstraintsOfPlonkCirc(piw *query.PlonkInWizard) int {

	hasAddGates := false
	if len(piw.PlonkOptions) > 0 {
		hasAddGates = piw.PlonkOptions[0].RangeCheckAddGateForRangeCheck
	}

	ccs, _, _ := plonkinternal.CompileCircuitWithRangeCheck(piw.Circuit, hasAddGates)
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

// audit checks that all the columns taking part in the [QueryBasedModule] have
// the same length. It panics
func (m *QueryBasedModule) mustHaveConsistentLength() {

	size := -1

	for col := range m.ds.Iter() {
		if size == -1 {
			size = col.Size()
		}

		if size != col.Size() {
			utils.Panic("col=%v does not have a consistent size %v != %v", col.ID, size, col.Size())
		}
	}
}

// groupQBModulesByAffinity groups the [QueryBasedModule] by affinity. Affinity
// consists in a group of columns that are to be grouped in the same standard
// module.
func groupQBModulesByAffinity(qbModules []*QueryBasedModule, affinities [][]column.Natural) (groups [][]*QueryBasedModule) {

	sets := make([]*collection.Set[*QueryBasedModule], len(qbModules))

	for i := range qbModules {
		s := collection.NewSet[*QueryBasedModule]()
		sets[i] = &s
		sets[i].Insert(qbModules[i])
	}

	for _, aff := range affinities {

		matched := make([]*collection.Set[*QueryBasedModule], 0)
		for i := range sets {

			isSetMatched := false

			for k := range aff {
				for qbm := range sets[i].Iter() {
					if qbm.ds.Has(aff[k]) {
						isSetMatched = true
						continue
					}
				}
			}

			if isSetMatched {
				matched = append(matched, sets[i])
			}
		}

		if len(matched) <= 1 {
			continue
		}

		mergedModule := matched[0]

		for i := 1; i < len(matched); i++ {
			mergedModule.Merge(matched[i])
			matched[i].Clear()
		}
	}

	groups = make([][]*QueryBasedModule, 0, len(sets))

	for i := range sets {
		if sets[i].Size() == 0 {
			continue
		}

		groups = append(
			groups,
			sets[i].SortKeysBy(
				func(qbm1, qbm2 *QueryBasedModule) bool {
					return string(qbm1.moduleName) < string(qbm2.moduleName)
				},
			),
		)
	}

	return groups
}

func weightOfGroupOfQBModules(group []*QueryBasedModule) int {
	var weight int
	for i := range group {
		weight += group[i].Weight(0)
	}
	return weight
}
