package distributedprojection

import (
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

type distributedProjectionVerifierAction struct {
	Name     ifaces.QueryID
	Horner0s []query.LocalOpening
	onlyB []bool
	skipped  bool
}

func (va *distributedProjectionVerifierAction) Run(run *wizard.VerifierRuntime) error {
	var (
		actualHorner = field.Zero()
	)
	for index, elemHorner := range va.Horner0s {
		if !va.onlyB[index] {
			elemHornerVal := run.GetLocalPointEvalParams(va.Horner0s[index].ID).Y
			actualHorner.Add(&actualHorner, &elemHornerVal)
}

} 
	return nil
}
