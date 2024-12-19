package distributed

import (
	"strings"

	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/variables"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
)

// ModuleDiscoverer a set of methods responsible for the horizontal splittings (i.e., splitting to modules)
type ModuleDiscoverer interface {
	// Analyze is responsible for letting the module discoverer compute how to
	// group best the columns into modules.
	Analyze(comp *wizard.CompiledIOP)
	ModuleList(comp *wizard.CompiledIOP) []ModuleName
	FindModule(col ifaces.Column) ModuleName
	// given a query and a module name it checks if the query is inside the module
	QueryIsInModule(ifaces.Query, ModuleName) bool
	ExpressionIsInModule(*symbolic.Expression, ModuleName) bool
}

type ModuleName string

// Example struct implementing ModuleDiscoverer
type PeriodSeperatingModuleDiscoverer struct {
	modules map[ModuleName][]ifaces.Column
}

// Analyze groups columns into modules
func (p *PeriodSeperatingModuleDiscoverer) Analyze(comp *wizard.CompiledIOP) {
	p.modules = make(map[ModuleName][]ifaces.Column)
	numRounds := comp.NumRounds()
	for i := range numRounds {
		for _, col := range comp.Columns.AllHandlesAtRound(i) { // Assume comp.Columns exists
			module := periodLogicToDetermineModule(col)
			p.modules[module] = append(p.modules[module], col)
		}
	}
}

func periodLogicToDetermineModule(col ifaces.Column) ModuleName {
	colName := col.GetColID()
	return ModuleName(periodSeparator(string(colName)))
}

func periodSeparator(name string) string {
	// Find the index of the first occurrence of a period
	index := strings.Index(name, ".")
	if index == -1 {
		// If no period is found, return the original string
		return name
	}
	// Return the substring before the first period
	return name[:index]
}

// NbModules returns the number of modules
func (p *PeriodSeperatingModuleDiscoverer) NbModules() int {
	return len(p.modules)
}

// ModuleList returns the list of module names
func (p *PeriodSeperatingModuleDiscoverer) ModuleList(comp *wizard.CompiledIOP) []ModuleName {
	moduleNames := make([]ModuleName, 0, len(p.modules))
	for moduleName := range p.modules {
		moduleNames = append(moduleNames, moduleName)
	}
	return moduleNames
}

// FindModule finds the module name for a given column
func (p *PeriodSeperatingModuleDiscoverer) FindModule(col ifaces.Column) ModuleName {
	for moduleName, columns := range p.modules {
		for _, c := range columns {
			if c == col {
				return moduleName
			}
		}
	}
	return "no column found" // Return a default or error value
}

// ColumnIsInModule checks that the given column is inside the given module.
func (p *PeriodSeperatingModuleDiscoverer) ColumnIsInModule(col ifaces.Column, name ModuleName) bool {
	for _, c := range p.modules[name] {
		if c.GetColID() == col.GetColID() {
			return true
		}
	}
	return false
}

//	ExpressionIsInModule checks that all the columns in the expression are from the given module.
//
// It does not check the presence of the coins and other metadata in the module.
func (p *PeriodSeperatingModuleDiscoverer) ExpressionIsInModule(expr *symbolic.Expression, name ModuleName) bool {
	var (
		board    = expr.Board()
		metadata = board.ListVariableMetadata()
		b        = true
		nCols    = 0
	)

	// by contradiction, if there is no metadata it belongs to the module.
	if len(metadata) == 0 {
		return true
	}

	for _, m := range metadata {
		switch v := m.(type) {
		case ifaces.Column:
			if !p.ColumnIsInModule(v, name) {
				b = b && false
			}
			nCols++
			// The expression can involve random coins
		case coin.Info, variables.X, variables.PeriodicSample, ifaces.Accessor:
			// Do nothing
		default:
			utils.Panic("unknown type %T", metadata)
		}
	}

	if nCols == 0 {
		panic("could not find any column in the expression")
	} else {
		return b
	}
}

func (p *PeriodSeperatingModuleDiscoverer) QueryIsInModule(ifaces.Query, ModuleName) bool {
	panic("unimplemented")
}
