package wizard

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/zkevm-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/zkevm-monorepo/prover/symbolic"
)

type Column interface {
	symbolic.Metadata
	GetAssignment(run Runtime) smartvectors.SmartVector
	GetAssignmentGnark(api frontend.API, run RuntimeGnark) []frontend.Variable
	Size() int
	Round() int
	Shift(n int) Column
}
