package hashtypes_test

// import (
// 	"testing"

// 	"github.com/consensys/linea-monorepo/prover/crypto/state-management/hashtypes"
// 	"github.com/consensys/linea-monorepo/prover/maths/field"
// 	"github.com/stretchr/testify/require"
// )

// func TestFieldHasher(t *testing.T) {
// 	assert := require.New(t)

// 	h1 := hashtypes.Poseidon2()
// 	h3 := hashtypes.Poseidon2()
// 	randInputs := make(field.Vector, 8)
// 	randInputs.MustSetRandom()

// 	// test SumElements
// 	dgst1 := h1.SumElements(randInputs)

// 	// test Write + Sum

// 	// test WriteElement + SumElement

// 	h3.WriteElements(randInputs)

// 	dgst3 := h3.SumElement()
// 	assert.Equal(dgst1, dgst3, "hashes do not match")

// }
