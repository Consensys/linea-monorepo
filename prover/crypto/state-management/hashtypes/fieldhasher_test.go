package hashtypes_test

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/crypto/state-management/hashtypes"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/utils/types"
	"github.com/stretchr/testify/require"
)

func TestFieldHasher(t *testing.T) {
	assert := require.New(t)

	h1 := hashtypes.Poseidon2()
	h2 := hashtypes.Poseidon2()
	h3 := hashtypes.Poseidon2()
	randInputs := make(field.Vector, 10)
	randInputs.MustSetRandom()

	// test SumElements
	dgst1 := h1.SumElements(randInputs)

	// test Write + Sum
	for i := range randInputs {
		h2.Write(randInputs[i].Marshal())
	}
	dgst2 := h2.Sum(nil)

	var dgst2Byte32 types.Bytes32
	copy(dgst2Byte32[:], dgst2[:])
	assert.Equal(dgst1, types.Bytes32ToOctuplet(dgst2Byte32), "hashes do not match")

	// test WriteElement + SumElement
	for i := range randInputs {
		h3.WriteElement(randInputs[i])
	}
	dgst3 := h3.SumElements(nil)
	assert.Equal(dgst1, dgst3, "hashes do not match")

}
