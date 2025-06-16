// Could be done only looking at the state (e.g. no need to run the EVM, just verifying a Merkle proof of inclusion and comparing the tx nonce and the account nonce.
// we need to proof the account is the right account (state meemgerhsip)
// load the right account from the merkleproof (state roof hash) - pi of the circuit

package nonce

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/utils/types"
)

type InvalidNonceCircuit struct {
	PublicInput frontend.Variable `gnark:",public"` //hash of the tx, stateroothash

	StateRootHash frontend.Variable `gnark:",secret"`

	RawTx []frontend.Variable `gnark:",secret"` // RLP-encoded transaction as a slice of bytes

	TxNonce frontend.Variable `gnark:",secret"` // Expected nonce for the transaction (?)

	Account []types.Account `gnark:",secret"`
}

func ExtractNonceFromRLP(api frontend.API, rawTx []frontend.Variable) frontend.Variable {
	// Extract the transaction type (first byte of rawTx)
	txType := rawTx[0]

	// Check if the transaction is typed (Access List or Dynamic Fee)
	isAccessList := api.IsZero(api.Sub(txType, frontend.Variable(0x01))) // txType == 0x01
	isDynamicFee := api.IsZero(api.Sub(txType, frontend.Variable(0x02))) // txType == 0x02
	isTyped := api.Or(isAccessList, isDynamicFee)

	// Directly access the nonce field based on the transaction type
	firstByte := api.Select(isTyped, rawTx[1], rawTx[0]) // Use hardcoded indices (1 for typed, 0 for legacy)

	// Check if the nonce is a single-byte value
	isSingleByte := api.IsZero(api.Sub(frontend.Variable(0x80), firstByte)) // 1 if firstByte < 0x80

	// Extract single-byte value
	singleByteValue := firstByte

	// Extract short length-prefixed value
	shortLength := api.Sub(firstByte, frontend.Variable(0x80)) // Length of the nonce
	shortValue := frontend.Variable(0)

	// remove isInRange,

	// Iterate over all possible indices in rawTx to simulate dynamic indexing
	for i := 0; i < 8; i++ {
		// Check if the current index falls within the range of the nonce bytes
		isInRange := api.And(
			api.IsZero(api.Sub(frontend.Variable(i), frontend.Variable(0))), // i >= 0
			api.IsZero(api.Sub(shortLength, frontend.Variable(i+1))),        // i < shortLength
		)

		// Select the value at the current index and reconstruct the nonce
		shortValue = api.Add(api.Mul(shortValue, frontend.Variable(256)), api.Select(isInRange, rawTx[i], frontend.Variable(0)))
	}

	// Select the correct nonce value based on the field type
	nonce := api.Select(isSingleByte, singleByteValue, shortValue)

	return nonce
}

func (c *InvalidNonceCircuit) Define(api frontend.API) error {

	// Extract the nonce field from the RLP-encoded transaction
	extractedTxNonce := ExtractNonceFromRLP(api, c.RawTx)

	// Validate the nonce
	validateNonce(api, extractedNonce, c.TxNonce)

	// Extract account nonce
	extractedAccountNonce := 

	return nil
}

func validateNonce(api frontend.API, extractedNonce frontend.Variable, expectedNonce frontend.Variable) {
	// Assert that the extracted nonce matches the expected nonce
	api.AssertIsEqual(extractedNonce, expectedNonce)
}
