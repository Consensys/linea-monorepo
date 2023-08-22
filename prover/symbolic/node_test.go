package symbolic_test

import (
	"testing"

	"github.com/consensys/accelerated-crypto-monorepo/symbolic"
	"github.com/stretchr/testify/require"
)

func TestNodeID(t *testing.T) {

	levels := []int{0, 0, 1, 2, 15, 64, 1 << 10}
	pos := []int{0, 1, 1, 0, 4, 12, 1 << 10, 42}

	for i := range levels {
		nodeID := symbolic.NewNodeID(levels[i], pos[i])

		require.Equal(t, levels[i], nodeID.Level(), "nodeid %v", nodeID)
		require.Equal(t, pos[i], nodeID.PosInLevel(), "nodeid %v", nodeID)
	}

}
