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

	h1 := hashtypes.NewPoseidon2FieldHasher()
	h2 := hashtypes.NewPoseidon2FieldHasher()
	h3 := hashtypes.NewPoseidon2FieldHasher()
	randInputs := make(field.Vector, 10)
	randInputs.MustSetRandom()

	// test SumElements
	dgst1 := h1.SumElements(randInputs)

	// test Write
	for i := range randInputs {
		h2.Write(randInputs[i].Marshal())
	}
	dgst2, err := h2.Sum(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assert.Equal(dgst1, types.Bytes32ToHash(dgst2), "hashes do not match")

	// test SumElement
	for i := range randInputs {
		h3.WriteElement(randInputs[i])
	}
	dgst3 := h3.SumElement()
	assert.Equal(dgst1, dgst3, "hashes do not match")

}
