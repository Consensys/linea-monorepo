package ethereum

import (
	"math/big"

	"github.com/ethereum/go-ethereum/core/types"
)

// Get the signer of a transaction
func GetSigner(tx *types.Transaction) types.Signer {
	if tx.Protected() {
		return getSigner(tx.ChainId())
	}
	return getUnprotectedSigner()
}

// Get the signer
func getSigner(chainID *big.Int) types.Signer {
	return types.NewLondonSigner(chainID)
}

// Get the unprotected signer
func getUnprotectedSigner() types.Signer {
	return types.HomesteadSigner{}
}
