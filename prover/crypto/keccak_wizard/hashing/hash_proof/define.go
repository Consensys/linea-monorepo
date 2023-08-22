package hash_proof

import (
	keccakf "github.com/consensys/accelerated-crypto-monorepo/crypto/keccak_wizard/keccakf"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/wizard"
)

// define the keccakFModule via the parameter numPerm
func DefineMultiHash(comp *wizard.CompiledIOP, round, numPerm int) (keccakFModule keccakf.KeccakFModule) {
	// set the number of permutations for keccakFModule to numPerm
	keccakFModule.NP = numPerm
	// define the keccakFModule
	keccakFModule.DefineKeccakF(comp, round)
	return keccakFModule
}
