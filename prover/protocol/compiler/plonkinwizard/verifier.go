package plonkinwizard

import (
	"fmt"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

// checkActivatorAndMask is an implementation of [wizard.VerifierAction] and is
// used to embody the verifier checks added by [checkActivators].
type checkActivatorAndMask struct {
	*context
	skipped bool
}

func (c *checkActivatorAndMask) Run(run wizard.Runtime) error {
	for i := range c.SelOpenings {
		var (
			localOpening = run.GetLocalPointEvalParams(c.SelOpenings[i].ID)
			valOpened    = localOpening.Y
			valActiv     = c.PlonkCtx.Columns.Activators[i].GetColAssignment(run).Get(0)
		)

		if valOpened != valActiv {
			return fmt.Errorf(
				"%v: activator does not match the circMask %v (activator=%v, mask=%v)",
				c.Q.ID, i, valActiv.String(), valOpened.String(),
			)
		}
	}

	return nil
}

func (c *checkActivatorAndMask) RunGnark(api frontend.API, run wizard.GnarkRuntime) {
	for i := range c.SelOpenings {
		var (
			valOpened = run.GetLocalPointEvalParams(c.SelOpenings[i].ID).Y
			valActiv  = c.PlonkCtx.Columns.Activators[i].GetColAssignmentGnarkAt(run, 0)
		)

		api.AssertIsEqual(valOpened, valActiv)
	}
}

func (c *checkActivatorAndMask) Skip() {
	c.skipped = true
}

func (c *checkActivatorAndMask) IsSkipped() bool {
	return c.skipped
}
