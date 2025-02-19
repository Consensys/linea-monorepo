package global

import (
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	edc "github.com/consensys/linea-monorepo/prover/zkevm/prover/publicInput/execution_data_collector"
)

type proverActionForBoundaries struct {
	provider         ifaces.Column
	receiver         ifaces.Column
	providerOpenings []query.LocalOpening
	receiverOpenings []query.LocalOpening

	hashOpeningProvider query.LocalOpening
	hashOpeningReceiver query.LocalOpening
	mimicHasherProvider edc.MIMCHasher
	mimicHasherReceiver edc.MIMCHasher
}

// it assigns all the LocalOpening  covering the boundaries
func (pa proverActionForBoundaries) Run(run *wizard.ProverRuntime) {
	var (
		providerWit = run.GetColumn(pa.provider.GetColID()).IntoRegVecSaveAlloc()
		receiverWit = run.GetColumn(pa.receiver.GetColID()).IntoRegVecSaveAlloc()
	)

	for i := range pa.providerOpenings {

		run.AssignLocalPoint(pa.providerOpenings[i].ID, providerWit[i])
		run.AssignLocalPoint(pa.receiverOpenings[i].ID, receiverWit[i])
	}

	pa.mimicHasherProvider.AssignHasher(run)
	pa.mimicHasherReceiver.AssignHasher(run)

	var (
		hashProvider = run.GetColumnAt(pa.mimicHasherProvider.HashFinal.GetColID(), 0)
		hashReceiver = run.GetColumnAt(pa.mimicHasherReceiver.HashFinal.GetColID(), 0)
	)

	run.AssignLocalPoint(pa.hashOpeningProvider.ID, hashProvider)
	run.AssignLocalPoint(pa.hashOpeningReceiver.ID, hashReceiver)
}
