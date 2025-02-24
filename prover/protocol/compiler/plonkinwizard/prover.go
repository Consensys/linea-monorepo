package plonkinwizard

import (
	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend/witness"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
)

// circAssignment is the prover step responsible for assigning the plonk circuit
type circAssignment struct{ *context }

// assignSelOpening is the prover step responsible for assigning the selector opening
type assignSelOpening struct{ *context }

func (a *circAssignment) Run(run *wizard.ProverRuntime) {
	a.PlonkCtx.GetPlonkProverAction().Run(run, a.getWitnesses(run))
}

func (a circAssignment) getWitnesses(run *wizard.ProverRuntime) []witness.Witness {

	var (
		data           = a.Q.Data.GetColAssignment(run).IntoRegVecSaveAlloc()
		sel            = a.Q.Selector.GetColAssignment(run).IntoRegVecSaveAlloc()
		nbPublic       = a.PlonkCtx.Plonk.SPR.GetNbPublicVariables()
		nbPublicPadded = utils.NextPowerOfTwo(nbPublic)
		witnesses      = make([]witness.Witness, 0, a.Q.Data.Size()/nbPublicPadded)
	)

	for i := 0; i < len(sel) && !sel[i].IsZero(); i += nbPublicPadded {

		var (
			locPubInputs  = data[i : i+nbPublic]
			locSelector   = sel[i : i+nbPublic]
			witness, _    = witness.New(ecc.BLS12_377.ScalarField())
			witnessFiller = make(chan any, nbPublic)
		)

		for currPos := 0; currPos < nbPublic; currPos++ {

			// NB: this will make the dummy verifier fail but not the
			// actual one as this is not checked by the query. Still,
			// if it happens it legitimately means there is a bug.
			if locSelector[currPos].IsZero() {
				panic("[plonkInWizard] incomplete assignment")
			}

			witnessFiller <- locPubInputs[currPos]
		}

		// closing the channel is necessary to prevent leaking and
		// also to let the witness "know" it is complete.
		close(witnessFiller)
		if err := witness.Fill(nbPublic, 0, witnessFiller); err != nil {
			utils.Panic("[witness.Fill] failed %v", err.Error())
		}

		witnesses = append(witnesses, witness)
	}

	return witnesses
}

func (a assignSelOpening) Run(run *wizard.ProverRuntime) {

	var (
		maxNbInstances    = len(a.SelOpenings)
		nbPubInputsPadded = utils.NextPowerOfTwo(a.Q.GetNbPublicInputs())
	)

	for i := 0; i < maxNbInstances; i++ {

		var (
			openedPos = i * nbPubInputsPadded
			openedVal = a.Q.Selector.GetColAssignment(run).Get(openedPos)
		)

		run.AssignLocalPoint(a.SelOpenings[i].ID, openedVal)
	}
}
