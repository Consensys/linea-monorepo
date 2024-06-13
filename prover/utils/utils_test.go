package utils_test

import (
	"testing"

	"github.com/consensys/zkevm-monorepo/prover/utils"
	"github.com/stretchr/testify/require"
)

func TestDivCeil(t *testing.T) {

	require.Equal(t, 1, utils.DivCeil(1, 10))
	require.Equal(t, 1, utils.DivCeil(2, 10))
	require.Equal(t, 1, utils.DivCeil(3, 10))
	require.Equal(t, 1, utils.DivCeil(4, 10))
	require.Equal(t, 1, utils.DivCeil(5, 10))
	require.Equal(t, 1, utils.DivCeil(6, 10))
	require.Equal(t, 1, utils.DivCeil(7, 10))
	require.Equal(t, 1, utils.DivCeil(8, 10))
	require.Equal(t, 1, utils.DivCeil(9, 10))
	require.Equal(t, 1, utils.DivCeil(10, 10))
	require.Equal(t, 2, utils.DivCeil(11, 10))
	require.Equal(t, 2, utils.DivCeil(19, 10))
	require.Equal(t, 2, utils.DivCeil(20, 10))
	require.Equal(t, 3, utils.DivCeil(21, 10))

}
