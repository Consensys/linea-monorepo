package projection

import (
	"fmt"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

// projectionVerifierAction is a compilation artifact generated during the
// execution of the [InsertProjection] and which implements the [wizard.VerifierAction]
// interface. It is meant to perform the verifier checks that the first values
// of the two Horner are equals.
type projectionVerifierAction struct {
	Name               ifaces.QueryID
	HornerA0, HornerB0 query.LocalOpening
	skipped            bool
}

// Run implements the [wizard.VerifierAction] interface.
func (va *projectionVerifierAction) Run(run wizard.Runtime) error {

	var (
		a = run.GetLocalPointEvalParams(va.HornerA0.ID).Y
		b = run.GetLocalPointEvalParams(va.HornerB0.ID).Y
	)

	if a != b {
		return fmt.Errorf("the horner check of the projection query `%v` did not pass", va.Name)
	}

	return nil
}

// RunGnark implements the [wizard.VerifierAction] interface.
func (va *projectionVerifierAction) RunGnark(api frontend.API, run wizard.GnarkRuntime) {

	var (
		a = run.GetLocalPointEvalParams(va.HornerA0.ID).Y
		b = run.GetLocalPointEvalParams(va.HornerB0.ID).Y
	)

	api.AssertIsEqual(a, b)
}

func (va *projectionVerifierAction) Skip() {
	va.skipped = true
}

func (va *projectionVerifierAction) IsSkipped() bool {
	return va.skipped
}
