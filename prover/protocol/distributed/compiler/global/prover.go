package global

import (
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

type paramAssignments struct {
	provider           ifaces.Column
	receiver           ifaces.Column
	providerOpenings   []query.LocalOpening
	receiverOpenings   []query.LocalOpening
	colAgainstProvider []query.LocalOpening
	colAgainstReceiver []query.LocalOpening
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
