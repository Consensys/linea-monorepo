package plonk

import (
	"reflect"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	cs "github.com/consensys/gnark/constraint/bls12-377"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/zkevm-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/zkevm-monorepo/prover/maths/field"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
	"github.com/consensys/zkevm-monorepo/prover/utils"
)

// This function is responsible for scheduling the assignment of the Wizard
// columns related to the currently compiled Plonk circuit. It is used
// specifically for when we do not wish to use BBS commitment as part of the
// circuit.
//
// In essence, the function works by computing the Plonk witness by calling the
// gnark solver over the circuit and assign the LRO columns from the resulting
// solution.
func (ctx *compilationCtx) registerNoCommitProver() {

	// Sanity-check
	if ctx.HasCommitment() {
		panic("this function should only be used when the commitment is not used")
	}

	lroProver := func(run *wizard.ProverRuntime) {

		for i := range ctx.Columns.L {
			// Let the assigner return an assignment
			assignment := ctx.Plonk.WitnessAssigner[i]()

			// Check that both the assignment and the base
			// circuit have the same type
			if reflect.TypeOf(ctx.Plonk.Circuit) != reflect.TypeOf(assignment) {
				utils.Panic("circuit and assignment do not have the same type (%T != %T)", ctx.Plonk.Circuit, assignment)
			}

			// Parse it as witness
			witness, err := frontend.NewWitness(assignment, ecc.BLS12_377.ScalarField())
			if err != nil {
				utils.Panic("Could not cast the assignment into a witness: %v", err)
			}

			publicWitness, err := frontend.NewWitness(assignment, ecc.BLS12_377.ScalarField(), frontend.PublicOnly())
			if err != nil {
				utils.Panic("Could not cast the assignment into a public witness: %v", err)
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
