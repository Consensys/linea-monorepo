package wizard

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/zkevm-monorepo/prover/maths/field"
	"github.com/consensys/zkevm-monorepo/prover/symbolic"
)

type Accessor interface {
	symbolic.Metadata
	Round() int
	GetVal(run Runtime) field.Element
	GetValGnark(api frontend.API, run GnarkRuntime) frontend.Variable
}
