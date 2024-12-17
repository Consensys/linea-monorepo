package distributed

import (
	"strings"

	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

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
		for _, col := range comp.Columns.AllHandlesAtRound(i) {
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
