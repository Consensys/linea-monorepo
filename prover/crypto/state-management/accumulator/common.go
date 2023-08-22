package accumulator

import (
	"io"

	"github.com/consensys/accelerated-crypto-monorepo/crypto/state-management/hashtypes"
	"github.com/consensys/accelerated-crypto-monorepo/crypto/state-management/smt"
)

// convenience alias
type Digest = hashtypes.Digest

// Generic hashing for object satisfying the io.WriterTo interface
func hash[T io.WriterTo](conf *smt.Config, m T) Digest {
	hasher := conf.HashFunc()
	m.WriteTo(hasher)
	digest := hasher.Sum(nil)
	return hashtypes.BytesToDigest(digest)
}

// Interaface that is shared by all the "traces" types. Used to collect traces
type DeferableCheck interface {
	DeferMerkleChecks(config *smt.Config, appendTo []smt.ProvedClaim) []smt.ProvedClaim
}
