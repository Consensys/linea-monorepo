package statemanager

import (
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/crypto/state-management/hashtypes"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/crypto/state-management/smt"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/utils/types"
)

var MIMC_CONFIG = &smt.Config{
	HashFunc: hashtypes.MiMC,
	Depth:    40,
}

// Returns an empty code hash : the hash of an empty bytes32
func EmptyCodeHash(config *smt.Config) Digest {
	hasher := config.HashFunc()
	hasher.Write(make([]byte, 32))
	return types.AsBytes32(hasher.Sum(nil))
}
