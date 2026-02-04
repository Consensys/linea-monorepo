package pi_interconnection

import (
	"hash"

	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/backend/blobsubmission"
	public_input "github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/public-input"
)

type Request struct {
	Decompressions []blobsubmission.Response
	Executions     []public_input.Execution
	Aggregation    public_input.Aggregation
}

// MerkleRoot computes the merkle root of data using the given hasher.
// TODO modify aggregation.PackInMiniTrees to optionally take a hasher instead of reimplementing
func MerkleRoot(hsh hash.Hash, treeNbLeaves int, data [][32]byte) [32]byte {
	if len(data) == 0 || len(data) > treeNbLeaves {
		panic("unacceptable tree size")
	}

	// duplicate; pad if necessary
	b := make([][32]byte, treeNbLeaves)
	copy(b, data)

	for len(b) != 1 {
		n := len(b) / 2
		for i := 0; i < n; i++ {
			hsh.Reset()
			hsh.Write(b[2*i][:])
			hsh.Write(b[2*i+1][:])
			copy(b[i][:], hsh.Sum(nil))
		}
		b = b[:n]
	}

	return b[0]
}
