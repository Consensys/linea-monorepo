package collection_test

import (
	"testing"

	"github.com/consensys/accelerated-crypto-monorepo/utils/collection"
	"github.com/stretchr/testify/require"
)

func TestVecVecLenOf(t *testing.T) {
	vecvec := collection.NewVecVec[int]()
	res := vecvec.LenOf(1)
	require.Equal(t, 0, res)
}
