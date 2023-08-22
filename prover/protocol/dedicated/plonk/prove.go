package plonk

import (
	"reflect"

	"github.com/consensys/accelerated-crypto-monorepo/maths/common/smartvectors"
	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/wizard"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	cs "github.com/consensys/gnark/constraint/bn254"
	"github.com/consensys/gnark/frontend"
)

func (ctx *Ctx) RegisterNoCommitProver() {

	// Sanity-check
	if ctx.HasCommitment() {
		panic("this function should only be used when the commitment is not used")
	}

	lroProver := func(run *wizard.ProverRuntime) {

		for i := range ctx.Columns.L {
			// Let the assigner return an assignment
			assignment := ctx.Plonk.Assigner[i]()

			// Check that both the assignment and the base
			// circuit have the same type
			if reflect.TypeOf(ctx.Plonk.Circuit) != reflect.TypeOf(assignment) {
				utils.Panic("circuit and assignment do not have the same type")
			}

			// Parse it as witness
			witness, err := frontend.NewWitness(assignment, ecc.BN254.ScalarField())
			if err != nil {
				utils.Panic("Could not parse the assignment into a witness")
			}

			publicWitness, err := frontend.NewWitness(assignment, ecc.BN254.ScalarField(), frontend.PublicOnly())
			if err != nil {
				utils.Panic("Could not parse the assignment into a public witness")
			}

			// Converts it as a smart-vector
			pubWitSV := smartvectors.RightZeroPadded(
				[]field.Element(publicWitness.Vector().(fr.Vector)),
				ctx.DomainSize(),
			)

			// Assign the public witness
			run.AssignColumn(ctx.Columns.PI[i].GetColID(), pubWitSV)

			// Solve the circuit
			sol_, err := ctx.Plonk.SPR.Solve(witness)
			if err != nil {
				utils.Panic("Error in the solver")
			}

			// And parse the solution into a witness
			solution := sol_.(*cs.SparseR1CSSolution)
			run.AssignColumn(ctx.Columns.L[i].GetColID(), smartvectors.NewRegular(solution.L))
			run.AssignColumn(ctx.Columns.R[i].GetColID(), smartvectors.NewRegular(solution.R))
			run.AssignColumn(ctx.Columns.O[i].GetColID(), smartvectors.NewRegular(solution.O))
		}

	}

	ctx.comp.SubProvers.AppendToInner(ctx.round, lroProver)
}
