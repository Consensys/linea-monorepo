//go:build !race

package hash_proof

import (
	"crypto/rand"
	"testing"

	"golang.org/x/crypto/sha3"

	keccakHash "github.com/consensys/accelerated-crypto-monorepo/crypto/keccak_wizard/hashing/hash"
	keccakf "github.com/consensys/accelerated-crypto-monorepo/crypto/keccak_wizard/keccakf"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/compiler/dummy"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/wizard"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

// it proves and verifies the hash-computation for multihash
func TestHashProof(t *testing.T) {
	// announcing the number of hashes
	numHash := 8
	multiHashModule := ParamForMultiHashModule(numHash)

	// numPerm for the keccakf module
	numPerm := 32

	var keccakFModule keccakf.KeccakFModule
	define := func(build *wizard.Builder) {
		// Note : for the moment there is no constraint over the hashes (but just over permutations),
		// thus the Define is not involved with multiHashModule
		keccakFModule = DefineMultiHash(build.CompiledIOP, 0, numPerm)
	}
	//compilation step, declaring the constraints
	compiled := wizard.Compile(
		define,
		// dummy.LazyCommit,
		dummy.Compile,
	)

	// proof and verification of the permutations KeccakF provoked by multiHashModule
	proof := wizard.Prove(compiled, func(run *wizard.ProverRuntime) { ProverAssign(run, multiHashModule, keccakFModule, numHash) })
	err := wizard.Verify(compiled, proof)
	require.NoError(t, err)
}
func ParamForMultiHashModule(k int) (l MultiHash) {
	l.InputHash = make([]Bytes, k)
	l.OutputHash = make([]Bytes, k)

	for i := range l.InputHash {
		//choose the number of bytes for the input to i-t hash
		l.InputHash[i] = make(Bytes, 3*i+4)
	}
	return l
}

func ProverAssign(run *wizard.ProverRuntime, hashModule MultiHash, keccakModule keccakf.KeccakFModule, k int) {
	for i := 0; i < k; i++ {
		_, err := rand.Read(hashModule.InputHash[i])
		if err != nil {
			logrus.Fatalf("error while generating random bytes: %s", err)
		}
		hashModule.OutputHash[i] = make([]byte, keccakHash.OutputLen)
		// obtain the output via the original keccakhash
		h := sha3.NewLegacyKeccak256()
		h.Write(hashModule.InputHash[i])
		hashModule.OutputHash[i] = h.Sum(nil)

	}
	// assign the keccakf module
	hashModule.Prover(run, keccakModule)
}
