package publicInput

import (
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/common"
)

// FunctionalInputExtractor is a collection over LocalOpeningQueries that can be
// used to check the values contained in the Wizard witness are consistent with
// the statement of the outer-proof.
type FunctionalInputExtractor struct {

	// DataNbBytes fetches the byte size of the execution data. It is important
	// to include it as the execution data hashing would be vulnerable to padding
	// attacks.
	DataNbBytes query.LocalOpening

	// DataChecksum returns the hash of the execution data
	DataChecksum [common.NbLimbU128]query.LocalOpening
	DataSZX      ifaces.Column
	DataSZY      ifaces.Accessor
	L2Messages   [16][common.NbLimbU256]query.LocalOpening

	// InitialStateRootHash and FinalStateRootHash are resp the initial and
	// root hash of the state for the
	InitialStateRootHash, FinalStateRootHash                  [common.NbElemPerHash]query.LocalOpening
	InitialBlockNumber, FinalBlockNumber                      [common.NbLimbU48]query.LocalOpening
	InitialBlockTimestamp, FinalBlockTimestamp                [common.NbLimbU128]query.LocalOpening
	FirstRollingHashUpdate, LastRollingHashUpdate             [common.NbLimbU256]query.LocalOpening
	FirstRollingHashUpdateNumber, LastRollingHashUpdateNumber [common.NbLimbU128]query.LocalOpening

	ChainID              [common.NbLimbU256]query.LocalOpening
	NBytesChainID        query.LocalOpening
	L2MessageServiceAddr [common.NbLimbEthAddress]query.LocalOpening
	CoinBase             [common.NbLimbEthAddress]query.LocalOpening
	BaseFee              [common.NbLimbU128]query.LocalOpening
}

// Run assigns all the local opening queries
func (fie *FunctionalInputExtractor) Run(run *wizard.ProverRuntime) {

	assignLO := func(q query.LocalOpening) {
		run.AssignLocalPoint(q.ID, q.Pol.GetColAssignmentAt(run, 0))
	}

	assignLOs := func(qs []query.LocalOpening) {
		for _, q := range qs {
			assignLO(q)
		}
	}

	assignLO(fie.DataNbBytes)
	assignLO(fie.NBytesChainID)
	assignLOs(fie.DataChecksum[:])
	assignLOs(fie.InitialStateRootHash[:])
	assignLOs(fie.FinalStateRootHash[:])
	assignLOs(fie.InitialBlockNumber[:])
	assignLOs(fie.FinalBlockNumber[:])
	assignLOs(fie.InitialBlockTimestamp[:])
	assignLOs(fie.FinalBlockTimestamp[:])
	assignLOs(fie.FirstRollingHashUpdate[:])
	assignLOs(fie.LastRollingHashUpdate[:])
	assignLOs(fie.FirstRollingHashUpdateNumber[:])
	assignLOs(fie.LastRollingHashUpdateNumber[:])
	assignLOs(fie.ChainID[:])
	assignLOs(fie.L2MessageServiceAddr[:])
	assignLOs(fie.CoinBase[:])
	assignLOs(fie.BaseFee[:])

	for i := range fie.L2Messages {
		assignLOs(fie.L2Messages[i][:])
	}
}
