package plonk

import (
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	cs "github.com/consensys/gnark/constraint/bls12-377"
	"github.com/consensys/zkevm-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/zkevm-monorepo/prover/maths/field"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
	"github.com/consensys/zkevm-monorepo/prover/utils"
	"github.com/consensys/zkevm-monorepo/prover/utils/parallel"
)

// PlonkInWizardProverAction is an interface representing prover runtime action
// to assign the Plonk circuit and run the gnark solver to generate the witness.
type PlonkInWizardProverAction interface {
	Run(run *wizard.ProverRuntime, wa WitnessAssigner)
}

// noCommitProverAction is a wrapper-type for [compilationCtx] implementing the
// [PlonkInWizardProverAction].
type noCommitProverAction compilationCtx

var (
	_ PlonkInWizardProverAction = noCommitProverAction{}
)

// Run is responsible for scheduling the assignment of the Wizard
// columns related to the currently compiled Plonk circuit. It is used
// specifically for when we do not wish to use BBS commitment as part of the
// circuit.
//
// In essence, the function works by computing the Plonk witness by calling the
// gnark solver over the circuit and assign the LRO columns from the resulting
// solution.
//
// It implements the [PlonkInWizardProverAction] interface.
func (pa noCommitProverAction) Run(run *wizard.ProverRuntime, wa WitnessAssigner) {

	var (
		ctx             = compilationCtx(pa)
		maxNbInstance   = pa.maxNbInstances
		numEffInstances = wa.NumEffWitnesses(run)
	)

	parallel.Execute(maxNbInstance, func(start, stop int) {
		for i := start; i < stop; i++ {

			if i >= numEffInstances {
				run.AssignColumn(ctx.Columns.TinyPI[i].GetColID(), smartvectors.NewConstant(field.Zero(), ctx.Columns.TinyPI[i].Size()))
				run.AssignColumn(ctx.Columns.L[i].GetColID(), smartvectors.NewConstant(field.Zero(), ctx.Columns.L[0].Size()))
				run.AssignColumn(ctx.Columns.R[i].GetColID(), smartvectors.NewConstant(field.Zero(), ctx.Columns.R[0].Size()))
				run.AssignColumn(ctx.Columns.O[i].GetColID(), smartvectors.NewConstant(field.Zero(), ctx.Columns.O[0].Size()))
				run.AssignColumn(ctx.Columns.Activators[i].GetColID(), smartvectors.NewConstant(field.Zero(), 1))
				continue
			}

			// create the witness assignment
			witness, pubWitness, err := wa.Assign(run, i)
			if err != nil {
				utils.Panic("Could not create the witness: %v", err)
			}
			if ctx.TinyPISize() > 0 {

				// Converts it as a smart-vector
				pubWitSV := smartvectors.RightZeroPadded(
					[]field.Element(pubWitness.Vector().(fr.Vector)),
					ctx.TinyPISize(),
				)

				// Assign the public witness
				run.AssignColumn(ctx.Columns.TinyPI[i].GetColID(), pubWitSV)
			}

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
			run.AssignColumn(ctx.Columns.Activators[i].GetColID(), smartvectors.NewConstant(field.One(), 1))
		}
	})

	if ctx.RangeCheck.Enabled && !ctx.RangeCheck.wasCancelled {
		ctx.assignRangeChecked(run)
	}
}
