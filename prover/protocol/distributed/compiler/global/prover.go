package global

import (
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	edc "github.com/consensys/linea-monorepo/prover/zkevm/prover/publicInput/execution_data_collector"
)

// it stores the data structs the prover should assigns
type boundaryAssignments struct {
	boundaries  boundaries
	hashOpening query.LocalOpening
	mimcHash    edc.MIMCHasher
}

// it stores the information that the prover should assign regarding provider/receiver
type proverActionForBoundaries struct {
	provider boundaryAssignments
	receiver boundaryAssignments
}

// it assigns all the LocalOpening covering the boundaries, and hash values
func (pa proverActionForBoundaries) Run(run *wizard.ProverRuntime) {
	var (
		provider         = pa.provider.boundaries.boundaryCol
		receiver         = pa.receiver.boundaries.boundaryCol
		providerOpenings = pa.provider.boundaries.boundaryOpenings
		receiverOpenings = pa.receiver.boundaries.boundaryOpenings

		providerWit = run.GetColumn(provider.GetColID()).IntoRegVecSaveAlloc()
		receiverWit = run.GetColumn(receiver.GetColID()).IntoRegVecSaveAlloc()
	)

	for _, loProvider := range providerOpenings.ListAllKeys() {
		index := providerOpenings.MustGet(loProvider)
		run.AssignLocalPoint(loProvider.ID, providerWit[index])
	}

	for _, loReceiver := range receiverOpenings.ListAllKeys() {
		index := receiverOpenings.MustGet(loReceiver)
		run.AssignLocalPoint(loReceiver.ID, receiverWit[index])
	}

	// compute and  assign the hash of Provider/Receiver
	pa.provider.mimcHash.AssignHasher(run)
	pa.receiver.mimcHash.AssignHasher(run)

	var (
		hashProvider = run.GetColumnAt(pa.provider.mimcHash.HashFinal.GetColID(), 0)
		hashReceiver = run.GetColumnAt(pa.receiver.mimcHash.HashFinal.GetColID(), 0)
	)

	run.AssignLocalPoint(pa.provider.hashOpening.ID, hashProvider)
	run.AssignLocalPoint(pa.receiver.hashOpening.ID, hashReceiver)
}
