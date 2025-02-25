package experiment

import (
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
)

// ModuleWitness is a structure collecting the witness of a module. And
// stores all the informations that are necessary to build the witness.
type ModuleWitness struct {
	// ModuleName indicates the name of the module
	ModuleName string
	// IsLPP indicates if the current instance of [ModuleWitness] is for
	// an LPP segment. In the contrary case, it is understood to be for
	// a GL segment.
	IsLPP bool
	// ModuleIndex indicates the vertical split of the current module
	ModuleIndex int
	// IsFirst, IsLast indicates if the module is the first or the last
	// segment of the module. When [ModuleIndex] == 0, [IsFirst] is true.
	IsFirst, IsLast bool
	// Columns maps the column id to their witness values
	Columns map[ifaces.ColID]smartvectors.SmartVector
	// ReceivedValuesGlobal stores the received values (for the global
	// constraints) of the current segment.
	ReceivedValuesGlobal []field.Element
	// N0 values are the parameters to the Horner queries in the same order
	// as in the [FilteredModuleInputs.HornerArgs]
	N0Values []int
}
