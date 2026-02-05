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
