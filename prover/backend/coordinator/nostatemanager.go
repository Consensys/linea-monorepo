package coordinator

import "github.com/consensys/accelerated-crypto-monorepo/utils"

// Adds the state-manager data from the keccak blocks
func PopulateWithKeccakRootHashes(pi *ProverInput, po *ProverOutput) {

	// Check that the keccak state root hash is set properly and sanity-check over it

	// Pass the hash as the ParentStateRootHash.
	parent, err := pi.KeccakParentStateRootHash()
	if err != nil {
		// If we can get the parent root hash we use the default value
		// 0x00000... . But we log a warning about it because we cannot
		// be sure that the error does not come from
		utils.Panic("the keccak root hash was not set properly: " + err.Error())
	}

	po.ParentStateRootHash = parent.Hex()

	// And pass the Keccak root hash
	blocks := pi.Blocks()
	for i := range blocks {
		po.BlocksData[i].RootHash = blocks[i].Root().Hex()
	}
}
