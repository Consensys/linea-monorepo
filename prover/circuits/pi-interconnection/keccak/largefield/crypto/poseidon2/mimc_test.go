package poseidon2_test

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/largefield/crypto/poseidon2"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/largefield/maths/field"
	"github.com/stretchr/testify/require"
)

func TestPoseidon2Block(t *testing.T) {

	for i := 0; i < 100; i++ {

		hasher := poseidon2.NewPoseidon2()

		// old is set to zero
		var x, old field.Element

		// s is set to a random value. Each run of the test will
		// generate a different value.
		x.SetRandom()
		xBytes := x.Bytes()

		newState := poseidon2.BlockCompression(old, x)

		hasher.Write(xBytes[:])
		newBytes := hasher.Sum(nil)
		var newFromHasher field.Element
		newFromHasher.SetBytes(newBytes)

		require.Equal(t, newFromHasher.String(), newState.String())
	}
}
