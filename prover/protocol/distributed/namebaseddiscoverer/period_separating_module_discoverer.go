package discoverer

import (
	"strings"

	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/distributed"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
)

// Type alias for distributed.ModuleName
type ModuleName = distributed.ModuleName

// Example struct implementing ModuleDiscoverer
type PeriodSeperatingModuleDiscoverer struct {
	modules map[ModuleName][]ifaces.Column
}

// Analyze groups columns into modules
func (p *PeriodSeperatingModuleDiscoverer) Analyze(comp *wizard.CompiledIOP) {
	p.modules = make(map[ModuleName][]ifaces.Column)
	numRounds := comp.NumRounds()
	for i := range numRounds {
		for _, col := range comp.Columns.AllHandlesAtRound(i) {
			module := periodLogicToDetermineModule(col)
			p.modules[module] = append(p.modules[module], col)
		}
	}
}

func periodLogicToDetermineModule(col ifaces.Column) ModuleName {
	colName := col.GetColID()
	// for multiplicity Column it is "TABLE_moduleName." So we should separate the ModuleName from this.
	name := ModuleName(periodSeparator(string(colName)))
	index := strings.LastIndex(name, "_")
	if index != -1 {
		name = name[index+1:]
	}
	return name
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

// ModuleList returns the list of module names
func (p *PeriodSeperatingModuleDiscoverer) ModuleList() []ModuleName {
	moduleNames := make([]ModuleName, 0, len(p.modules))
	for moduleName := range p.modules {
		moduleNames = append(moduleNames, moduleName)
	}
	return moduleNames
}

func (p *PeriodSeperatingModuleDiscoverer) ListColumns(modulaName ModuleName) []ifaces.Column {
	return p.modules[modulaName]
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

// QueryIsInModule checks if the given query is inside the given module
func (p *PeriodSeperatingModuleDiscoverer) QueryIsInModule(ifaces.Query, ModuleName) bool {
	panic("unimplemented")

}

// ColumnIsInModule checks that the given column is inside the given module.
func (p *PeriodSeperatingModuleDiscoverer) ColumnIsInModule(col ifaces.Column, name ModuleName) bool {
	colID := col.GetColID()
	if shifted, ok := col.(column.Shifted); ok {
		colID = shifted.Parent.GetColID()
	}
	for _, c := range p.modules[name] {
		if c.GetColID() == colID {
			return true
		}
	}
	return false
}

func (p *PeriodSeperatingModuleDiscoverer) HasModule(col ifaces.Column) (ModuleName, bool) {
	for moduleName, columns := range p.modules {
		for _, c := range columns {
			if c == col {
				return moduleName, true
			}
		}
	}
	return "", false
}

//	ExpressionIsInModule checks if any column, from the expression,
//
// is assigned to the given module. If so, it return true.
func (p *PeriodSeperatingModuleDiscoverer) ExpressionIsInModule(expr *symbolic.Expression, name ModuleName) bool {
	var (
		cols = distributed.ListColumnsFromExpr(expr, true)
	)

	// by contradiction, if there is no column it belongs to the module.
	if len(cols) == 0 {
		return true
	}

	for _, col := range cols {

		// verifer column can be common among modules (since they have the same ID),
		//so they are not good for decision making.
		if !distributed.IsVerifierColumn(col) {

			if p.ColumnIsInModule(col, name) {
				return true
			}

		}
	}

	return false

}

// ExpressionIsInModule checks if any column, from the given slice, is assigned to the given module. If so, it return true.
func (p *PeriodSeperatingModuleDiscoverer) SliceIsInModule(cols []ifaces.Column, name ModuleName) bool {

	// by contradiction, if there is no column it belongs to the module.
	if len(cols) == 0 {
		return true
	}

	for _, col := range cols {

		if !distributed.IsVerifierColumn(col) {
			if p.ColumnIsInModule(col, name) {
				return true
			}
		}

	}

	return false

}

// UpdateDiscoverer assign all the unassigned columns, from the given expression, to the given module.
// if a column is already assigned to a different module, it panics.
func (p *PeriodSeperatingModuleDiscoverer) UpdateDiscoverer(cols []ifaces.Column, name ModuleName) {

	for _, col := range cols {

		// sanity check; panic if the column is already in a different module
		if actualModule, ok := p.HasModule(col); ok {
			if actualModule != name {
				if !distributed.IsVerifierColumn(col) {
					utils.Panic("column %v is from module %v and not from the given module %v",
						col.GetColID(), actualModule, name)
				}
			}
		}

		// assign the column to the module, if the column is not assigned to any module.
		if !p.ColumnIsInModule(col, name) {
			p.modules[name] = append(p.modules[name], col)

		}

	}

}

func (p *PeriodSeperatingModuleDiscoverer) NewSizeOf(ifaces.Column) int {
	panic("unimplemented")
}
