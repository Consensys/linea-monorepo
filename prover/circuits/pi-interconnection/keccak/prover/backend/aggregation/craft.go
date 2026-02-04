package aggregation

import (
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/crypto/state-management/hashtypes"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/crypto/state-management/smt"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/utils"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/utils/types"
)

const (
	// Indicates the depth of a Merkle-tree for L2 messages, and implicitly how
	// many messages can be stored in a single root hash.
	l2MsgMerkleTreeDepth     = 5
	l2MsgMerkleTreeMaxLeaves = 1 << l2MsgMerkleTreeDepth
)

// Pack an array of boolean into an offset list. The offset list encodes the
// position of all the boolean whose value is true. Each position is encoded
// as a big-endian uint16.

// Hash the L2 messages into Merkle trees or arity 2 and depth
// `l2MsgMerkleTreeDepth`. The leaves are zero-padded on the right.
func PackInMiniTrees(l2MsgHashes []string) []string {

	paddedLen := utils.NextMultipleOf(len(l2MsgHashes), l2MsgMerkleTreeMaxLeaves)
	paddedL2MsgHashes := make([]string, paddedLen)
	copy(paddedL2MsgHashes, l2MsgHashes)

	res := []string{}

	for i := 0; i < paddedLen; i += l2MsgMerkleTreeMaxLeaves {

		digests := make([]types.Bytes32, l2MsgMerkleTreeMaxLeaves)

		// Convert the leaves into digests that can be processed by the smt
		// package.
		for j := range digests {
			leaf := paddedL2MsgHashes[i+j]
			decoded, err := utils.HexDecodeString(leaf)
			copy(digests[j][:], decoded)

			if err != nil {
				panic(err)
			}
		}

		tree := smt.BuildComplete(digests, hashtypes.Keccak)
		res = append(res, tree.Root.Hex())
	}

	return res
}
