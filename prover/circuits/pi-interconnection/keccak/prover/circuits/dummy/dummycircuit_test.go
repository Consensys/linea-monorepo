package dummy_test

import (
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/circuits"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/circuits/dummy"
	"github.com/stretchr/testify/require"
)

// Test that the generated proofs works as expected
func TestDummyCircuitProofGen(t *testing.T) {
	assert := require.New(t)

	x := fr.NewElement(uint64(0xabcdef0123456789))

	srsProvider := circuits.NewUnsafeSRSProvider()

	for _, id := range []circuits.MockCircuitID{0, 1, 0xfff, 78970} {
		pp, err := dummy.MakeUnsafeSetup(srsProvider, id, ecc.BN254.ScalarField())
		assert.NoError(err)

		_ = dummy.MakeProof(&pp, x, id)
	}
}
