package nonce

import "github.com/consensys/gnark/frontend"

// To do: We will only use txType == 0x02 (Dynamic Fee Transactions)
// Legacy and AccessList can be removed
func ExtractTxNonceFromRLP(api frontend.API, rawTx []frontend.Variable) frontend.Variable {
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
