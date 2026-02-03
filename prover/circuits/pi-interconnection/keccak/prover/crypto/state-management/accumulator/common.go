package accumulator

import (
	"io"

	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/crypto/state-management/smt"

	//lint:ignore ST1001 -- the package contains a list of standard types for this repo
	. "github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/utils/types"
)

// Generic hashing for object satisfying the io.WriterTo interface
func hash[T io.WriterTo](conf *smt.Config, m T) Bytes32 {
	hasher := conf.HashFunc()
	m.WriteTo(hasher)
	Bytes32 := hasher.Sum(nil)
	return AsBytes32(Bytes32)
}

// Trace is an interface shared by all the "traces" types. Used to
// collect MerkleProof verifications claims.
type Trace interface {
	// DeferMerkleChecks appends all the merkle-proofs checks happening in a trace
	// verification into a slice of smt.ProvedClaim
	DeferMerkleChecks(config *smt.Config, appendTo []smt.ProvedClaim) []smt.ProvedClaim
	// HKey returns the HKey of the trace
	HKey(cfg *smt.Config) Bytes32
	// RWInt returns 0 is the trace is a read-only operation and 1 if it is a
	// read-write operation.
	RWInt() int
}
