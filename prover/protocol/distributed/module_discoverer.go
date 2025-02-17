package distributed

import (
	"sync"

	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/column/verifiercol"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/variables"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
)

// ModuleDiscoverer implements the ModuleDiscovererInterface
type ModuleDiscoverer struct {
	moduleMapping map[ifaces.Column]ModuleName
	moduleNames   []ModuleName
	mutex         sync.Mutex
}

// ModuleDiscovererInterface defines methods for horizontal splitting (i.e., splitting into modules).
type ModuleDiscovererInterface interface {
	// Analyze computes how to group columns into modules.
	Analyze(comp *wizard.CompiledIOP)
	NbModules() int
	ModuleList() []ModuleName
	FindModule(col ifaces.Column) ModuleName
	ExpressionIsInModule(*symbolic.Expression, ModuleName) bool
	QueryIsInModule(ifaces.Query, ModuleName) bool
	ColumnIsInModule(col ifaces.Column, name ModuleName) bool
}

// NewModuleDiscoverer initializes and returns a new ModuleDiscoverer instance.
func NewModuleDiscoverer() *ModuleDiscoverer {
	return &ModuleDiscoverer{
		moduleMapping: make(map[ifaces.Column]ModuleName),
		moduleNames:   []ModuleName{},
	}
}

// Analyze clusters columns into modules by iterating through global constraints (QueriesNoParams).
// Columns sharing the same global constraints are grouped into the same module.
func (md *ModuleDiscoverer) Analyze(comp *wizard.CompiledIOP) {
	md.mutex.Lock()
	defer md.mutex.Unlock()

	moduleIndex := 0
	columnClusters := make(map[ModuleName]map[ifaces.Column]bool)

	for _, qName := range comp.QueriesNoParams.AllUnignoredKeys() {
		query := comp.QueriesNoParams.Data(qName)

		// Determine the columns connected by this global constraint
		connectedColumns := getColumnsFromQuery(query)

		// Check if these columns belong to an existing module
		moduleFound := false
		for moduleName, cluster := range columnClusters {
			if sharesColumns(cluster, connectedColumns) {
				mergeClusters(cluster, connectedColumns)
				moduleFound = true
				md.assignModule(moduleName, connectedColumns)
				break
			}
		}

		// If no module matches, create a new one
		if !moduleFound {
			moduleName := ModuleName("Module_" + string(moduleIndex))
			moduleIndex++
			columnClusters[moduleName] = connectedColumns
			md.moduleNames = append(md.moduleNames, moduleName)
			md.assignModule(moduleName, connectedColumns)
		}
	}
}

// NbModules returns the total number of discovered modules.
func (md *ModuleDiscoverer) NbModules() int {
	md.mutex.Lock()
	defer md.mutex.Unlock()
	return len(md.moduleNames)
}

// ModuleList returns the list of all module names.
func (md *ModuleDiscoverer) ModuleList() []ModuleName {
	md.mutex.Lock()
	defer md.mutex.Unlock()
	return append([]ModuleName(nil), md.moduleNames...)
}

// FindModule returns the module name for the given column.
func (md *ModuleDiscoverer) FindModule(col ifaces.Column) ModuleName {
	md.mutex.Lock()
	defer md.mutex.Unlock()
	return md.moduleMapping[col]
}

// ExpressionIsInModule checks that all the columns  (except verifiercol) in the expression are from the given module.
func (md *ModuleDiscoverer) ExpressionIsInModule(expr *symbolic.Expression, name ModuleName) bool {
	board := expr.Board()
	metadata := board.ListVariableMetadata()

	// by contradiction, if there is no metadata it belongs to the module.
	if len(metadata) == 0 {
		return true
	}

	md.mutex.Lock()
	defer md.mutex.Unlock()

	b := true
	nCols := 0

	for _, m := range metadata {
		switch v := m.(type) {
		case ifaces.Column:
			if _, ok := v.(verifiercol.VerifierCol); !ok {
				if !md.ColumnIsInModule(v, name) {
					b = false
				}
				nCols++
			}
			// The expression can involve random coins
		case coin.Info, variables.X, variables.PeriodicSample, ifaces.Accessor:
			// Do nothing
		default:
			utils.Panic("unknown type %T", metadata)
		}
	}

	if nCols == 0 {
		panic("could not find any column in the expression")
	}
	return b
}

// QueryIsInModule checks if the given query is inside the given module
func (md *ModuleDiscoverer) QueryIsInModule(query ifaces.Query, name ModuleName) bool {
	md.mutex.Lock()
	defer md.mutex.Unlock()
	for _, col := range query.Columns() {
		if md.FindModule(col) != name {
			return false
		}
	}
	return true
}

// ColumnIsInModule checks that the given column is inside the given module.
func (md *ModuleDiscoverer) ColumnIsInModule(col ifaces.Column, name ModuleName) bool {
	md.mutex.Lock()
	defer md.mutex.Unlock()
	return md.moduleMapping[col] == name
}

// CoinIsInModule (placeholder): Extend logic to handle coins if needed.
func (md *ModuleDiscoverer) CoinIsInModule(coin ifaces.Coin, name ModuleName) bool {
	// Logic for associating coins with modules goes here
	return false
}

// Utility: Extracts columns involved in a query.
func getColumnsFromQuery(query ifaces.Query) map[ifaces.Column]bool {
	columns := make(map[ifaces.Column]bool)
	for _, col := range query.Columns() {
		columns[col] = true
	}
	return columns
}

// Utility: Checks if two column clusters share any columns.
func sharesColumns(cluster map[ifaces.Column]bool, columns map[ifaces.Column]bool) bool {
	for col := range columns {
		if cluster[col] {
			return true
		}
	}
	return false
}

// Utility: Merges columns into an existing cluster.
func mergeClusters(cluster map[ifaces.Column]bool, columns map[ifaces.Column]bool) {
	for col := range columns {
		cluster[col] = true
	}
}

// Utility: Assigns module information to columns in the mapping.
func (md *ModuleDiscoverer) assignModule(moduleName ModuleName, columns map[ifaces.Column]bool) {
	for col := range columns {
		md.moduleMapping[col] = moduleName
	}
}
