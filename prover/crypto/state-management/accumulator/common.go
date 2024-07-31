package accumulator

import (
	"io"

	"github.com/consensys/zkevm-monorepo/prover/crypto/state-management/smt"
	//lint:ignore ST1001 -- the package contains a list of standard types
	//lint:ignore ST1001 -- the package contains a list of standard types for this repo
	. "github.com/consensys/zkevm-monorepo/prover/utils/types"
)

// Generic hashing for object satisfying the io.WriterTo interface
func hash[T io.WriterTo](conf *smt.Config, m T) Bytes32 {
	hasher := conf.HashFunc()
	m.WriteTo(hasher)
	Bytes32 := hasher.Sum(nil)
	return AsBytes32(Bytes32)
}

// DeferableCheck is an interface shared by all the "traces" types. Used to
// collect MerkleProof verifications claims.
type DeferableCheck interface {
	// DeferMerkleChecks appends all the merkle-proofs checks happening in a trace
	// verification into a slice of smt.ProvedClaim
	DeferMerkleChecks(config *smt.Config, appendTo []smt.ProvedClaim) []smt.ProvedClaim
}
