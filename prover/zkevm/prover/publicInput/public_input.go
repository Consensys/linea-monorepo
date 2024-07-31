package publicInput

import fetch "github.com/consensys/zkevm-monorepo/prover/zkevm/prover/publicInput/fetchers_arithmetization"

// PublicInput collects a number of submodules responsible for collecting the
// wizard witness data holding the public inputs of the execution circuit.
type PublicInput struct {
	TimestampFetcher fetch.TimestampFetcher
	RootHashFetcher  fetch.RootHashFetcher
}

// GetExtractor returns [FunctionalInputExtractor] giving access to the totality
// of the public inputs recovered by the public input module.
func (pi *PublicInput) GetExtractor() *FunctionalInputExtractor {
	panic("unimplemented")
}
