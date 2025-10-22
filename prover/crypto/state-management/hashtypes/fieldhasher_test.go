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
	randInputs := make(field.Vector, 10)
	randInputs.MustSetRandom()

	// test Write + Sum
	var bytes []byte
	for _, elem := range randInputs {
		bytes = append(bytes, elem.Marshal()...)
	}
	h1.Write(bytes) // Write the whole byte slice
	dgst1 := h1.Sum(nil)
	var dgst1Byte32 types.Bytes32
	copy(dgst1Byte32[:], dgst1[:])

	// test WriteElement + SumElement
	h2.WriteElements(randInputs)
	dgst2 := h2.SumElement()
	assert.Equal(types.Bytes32ToHash(dgst1Byte32), dgst2, "hashes do not match")

}
