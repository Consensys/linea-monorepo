package permutation

import (
	"errors"
	"fmt"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

// The verifier gets all the query openings and multiply them together and
// expect them to be one. It is represented by an array of ZCtx holding for
// the same round. (we have the guarantee that they come from the same query).
type VerifierCtx struct {
	Ctxs    []*ZCtx
	skipped bool
}

// Run implements the [wizard.VerifierAction] interface and checks that the
// product of the products given by the ZCtx is equal to one.
func (v *VerifierCtx) Run(run wizard.Runtime) error {

	mustBeOne := field.One()

	for _, zCtx := range v.Ctxs {
		for _, opening := range zCtx.ZOpenings {
			y := run.GetLocalPointEvalParams(opening.ID).Y
			mustBeOne.Mul(&mustBeOne, &y)
		}
	}

	if mustBeOne != field.One() {
		return errors.New("the permutation check compiler did not pass")
	}

	return nil
}

// Run implements the [wizard.VerifierAction] interface and is as
// [VerifierCtx.Run] but in the context of a gnark circuit.
func (v *VerifierCtx) RunGnark(api frontend.API, run wizard.GnarkRuntime) {

	mustBeOne := frontend.Variable(1)

	for _, zCtx := range v.Ctxs {
		for _, opening := range zCtx.ZOpenings {
			y := run.GetLocalPointEvalParams(opening.ID).Y
			mustBeOne = api.Mul(mustBeOne, y)
		}
	}

	api.AssertIsEqual(mustBeOne, frontend.Variable(1))
}

func (v *VerifierCtx) Skip() {
	v.skipped = true
}

func (v *VerifierCtx) IsSkipped() bool {
	return v.skipped
}

// CheckGrandProductIsOne is a verifier action checking that the grand product
// is one.
type CheckGrandProductIsOne struct {
	Query   *query.GrandProduct
	Skipped bool
}

func (c *CheckGrandProductIsOne) Run(run wizard.Runtime) error {
	y := run.GetGrandProductParams(c.Query.ID).Y
	if !y.IsOne() {
		return fmt.Errorf("[Permutation -> GrandProduct] the outcome of the grand-product query should be one")
	}
	return nil
}

func (c *CheckGrandProductIsOne) RunGnark(api frontend.API, run wizard.GnarkRuntime) {
	y := run.GetGrandProductParams(c.Query.ID).Prod
	api.AssertIsEqual(y, frontend.Variable(1))
}

func (c *CheckGrandProductIsOne) Skip() {
	c.Skipped = true
}

func (c *CheckGrandProductIsOne) IsSkipped() bool {
	return c.Skipped
}
