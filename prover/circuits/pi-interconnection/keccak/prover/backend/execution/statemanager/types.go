package statemanager

import (
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/crypto/state-management/accumulator"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/utils/types"
)

// Handy type aliases
type (
	Digest      = types.Bytes32
	Address     = types.EthAddress
	Account     = types.Account
	FullBytes32 = types.FullBytes32
)

type (
	StorageTrie = accumulator.ProverState[FullBytes32, FullBytes32]

	// Account VS
	AccountVerifier = accumulator.VerifierState[Address, Account]
	StorageVerifier = accumulator.VerifierState[FullBytes32, FullBytes32]

	// ReadNonZeroTrace
	ReadNonZeroTraceWS = accumulator.ReadNonZeroTrace[Address, Account]
	ReadNonZeroTraceST = accumulator.ReadNonZeroTrace[FullBytes32, FullBytes32]

	// ReadZeroTrace
	ReadZeroTraceWS = accumulator.ReadZeroTrace[Address, Account]
	ReadZeroTraceST = accumulator.ReadZeroTrace[FullBytes32, FullBytes32]

	// InsertionTrace
	InsertionTraceWS = accumulator.InsertionTrace[Address, Account]
	InsertionTraceST = accumulator.InsertionTrace[FullBytes32, FullBytes32]

	// UpdateTrace
	UpdateTraceWS = accumulator.UpdateTrace[Address, Account]
	UpdateTraceST = accumulator.UpdateTrace[FullBytes32, FullBytes32]

	// DeletionTrace
	DeletionTraceWS = accumulator.DeletionTrace[Address, Account]
	DeletionTraceST = accumulator.DeletionTrace[FullBytes32, FullBytes32]
)
