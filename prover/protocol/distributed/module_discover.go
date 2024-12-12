package distributed

import (
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

// it implement [ModuleDiscoverer], it splits the compiler horizontally.
type HorizontalSplitting struct {
	modules []string
}

func (split HorizontalSplitting) Analyze(comp *wizard.CompiledIOP) {

}

func (split HorizontalSplitting) Split(comp *wizard.CompiledIOP) {

}

// QueryIsInModule checks if the given query is inside the given module
func (split HorizontalSplitting) QueryIsInModule(ifaces.Query, moduleName) bool {
	panic("unimplemented")

}
