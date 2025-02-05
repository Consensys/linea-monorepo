package global

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
)

type paramAssignments struct {
	provider           ifaces.Column
	receiver           ifaces.Column
	providerOpenings   []query.LocalOpening
	receiverOpenings   []query.LocalOpening
	colAgainstProvider []query.LocalOpening
	colAgainstReceiver []query.LocalOpening
}

type verifierChecks struct {
	providerOpenings   []query.LocalOpening
	receiverOpenings   []query.LocalOpening
	colAgainstProvider []query.LocalOpening
	colAgainstReceiver []query.LocalOpening
	skipped            bool
}

func (pa paramAssignments) Run(run *wizard.ProverRuntime) {
	var (
		providerWit = run.GetColumn(pa.provider.GetColID()).IntoRegVecSaveAlloc()
		receiverWit = run.GetColumn(pa.receiver.GetColID()).IntoRegVecSaveAlloc()
	)

	for i := range pa.providerOpenings {
		run.AssignLocalPoint(pa.providerOpenings[i].ID, providerWit[i])
		run.AssignLocalPoint(pa.colAgainstProvider[i].ID, providerWit[i])

		run.AssignLocalPoint(pa.receiverOpenings[i].ID, receiverWit[i])
		run.AssignLocalPoint(pa.colAgainstReceiver[i].ID, receiverWit[i])
	}
}

func (v *verifierChecks) Run(run *wizard.VerifierRuntime) error {

	for i, pOpening := range v.providerOpenings {

		var (
			p    = run.GetLocalPointEvalParams(pOpening.ID).Y
			colP = run.GetLocalPointEvalParams(v.colAgainstProvider[i].ID).Y
			r    = run.GetLocalPointEvalParams(v.receiverOpenings[i].ID).Y
			colR = run.GetLocalPointEvalParams(v.colAgainstReceiver[i].ID).Y
		)

		if p != colP {
			utils.Panic("LocalOpenings from columns and the provider are not consistence,"+
				" from column %v, from provider %v", colP.String(), p.String())
		}

		if r != colR {
			panic("LocalOpenings from columns and the receiver are not consistence")
		}

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
