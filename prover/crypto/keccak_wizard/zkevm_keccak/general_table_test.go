//go:build !race

package zkevm_keccak_test

import (
	"testing"

	hp "github.com/consensys/accelerated-crypto-monorepo/crypto/keccak_wizard/hashing/hash_proof"
	keccakf "github.com/consensys/accelerated-crypto-monorepo/crypto/keccak_wizard/keccakf"
	"github.com/consensys/accelerated-crypto-monorepo/crypto/keccak_wizard/zkevm_keccak"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/compiler/dummy"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/wizard"
	"github.com/stretchr/testify/require"
)

// it testes the Multihash-proof for the random zkevm tables
func TestZkEVMKeccak(t *testing.T) {
	size := 256
	//generate zkEVM tables
	table := TableForTest(size)
	numPerm := 32

	var KeccakFmodule keccakf.KeccakFModule
	define := func(build *wizard.Builder) {
		KeccakFmodule = hp.DefineMultiHash(build.CompiledIOP, 0, numPerm)
	}
	//compile, declare the constraints for multi keccak hash
	compiled := wizard.Compile(
		define,
		dummy.Compile,
	)

	// proof and verification of the multi hashes
	proof := wizard.Prove(compiled, func(run *wizard.ProverRuntime) {
		multiHash := table.MultiHashFromTable(numPerm, zkevm_keccak.RAND)
		multiHash.Prover(run, KeccakFmodule)
	})
	err := wizard.Verify(compiled, proof)
	require.NoError(t, err)

}
