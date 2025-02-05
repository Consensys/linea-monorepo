package global

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
)

type verifierChecks struct {
	providerOpenings   []query.LocalOpening
	receiverOpenings   []query.LocalOpening
	colAgainstProvider []query.LocalOpening
	colAgainstReceiver []query.LocalOpening
	skipped            bool
}

func (v *verifierChecks) Run(run *wizard.VerifierRuntime) error {

	for i, pOpening := range v.providerOpenings {

		var (
			p    = run.GetLocalPointEvalParams(pOpening.ID).Y
			colP = run.GetLocalPointEvalParams(v.colAgainstProvider[i].ID).Y
			//			r    = run.GetLocalPointEvalParams(v.receiverOpenings[i].ID).Y
			//			colR = run.GetLocalPointEvalParams(v.colAgainstReceiver[i].ID).Y
		)

		if p != colP {
			utils.Panic("LocalOpenings from columns and the provider are not consistence,"+
				" from column %v, from provider %v", colP.String(), p.String())
		}

		/*		if r != colR {
					panic("LocalOpenings from columns and the receiver are not consistence")
				}
		*/

	}
	return nil

}

func (v *verifierChecks) Skip() {
	v.skipped = true
}

func (v *verifierChecks) IsSkipped() bool {
	return v.skipped

}

func (v *verifierChecks) RunGnark(api frontend.API, run *wizard.WizardVerifierCircuit) {

	for i, pOpening := range v.providerOpenings {

		var (
			p    = run.GetLocalPointEvalParams(pOpening.ID).Y
			colP = run.GetLocalPointEvalParams(v.colAgainstProvider[i].ID).Y
			r    = run.GetLocalPointEvalParams(v.receiverOpenings[i].ID).Y
			colR = run.GetLocalPointEvalParams(v.colAgainstReceiver[i].ID).Y
		)

		api.AssertIsEqual(p, colP)
		api.AssertIsEqual(r, colR)

	}

}
