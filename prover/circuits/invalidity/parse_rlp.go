package invalidity

import (
	"errors"

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

// RLPExtractNonce extracts the nonce from an RLP-encoded transaction.
// The function checks the transaction type (from the first byte) and selects the correct index for the nonce accordingly.
// It selects the nonce from index 0 for Legacy transactions and index 2 for Access List and Dynamic Fee transactions,
// ensuring the correct nonce is extracted for any transaction type.

// RLPExtractNonce handles all Ethereum transaction types:
// - Legacy Tx: nonce at index 0
// - Access List Tx (EIP-2930) and Dynamic Fee Tx (EIP-1559): nonce at index 2
//
// RLPExtractNonce covers all four Ethereum transaction types in
// backend/ethereum/tx_encoding.go:
// 1. Legacy Transactions** (without EIP-155 protection or with):
//   - Nonce is located at index 0 in the RLP encoding.
//
// 2. Access List Transactions** (EIP-2930):
//   - Nonce is located at index 2 in the RLP encoding (after the transaction type byte).
//
// 3. Dynamic Fee Transactions** (EIP-1559):
//   - Nonce is also located at index 2 in the RLP encoding (after the transaction type byte).
//
// 4. Legacy Transactions with EIP-155 Protection** (EIP-155):
//   - Nonce is located at index 0, similar to standard Legacy transactions, but with additional chain ID information.
func RLPExtractNonce(api frontend.API, rawTx []frontend.Variable) frontend.Variable {
	// Extract Transaction Type
	// txType = rawTx[0]
	txType := rawTx[0]

	// Check if the Transaction is Typed (AccessList or DynamicFee)
	// isTyped = (txType == 0x01) || (txType == 0x02)
	isAccessList := api.IsZero(api.Sub(txType, 0x01)) // txType == 0x01
	isDynamicFee := api.IsZero(api.Sub(txType, 0x02)) // txType == 0x02
	isTyped := api.Or(isAccessList, isDynamicFee)

	// For typed txs, nonce is at index 2 (tx[1] = list prefix, tx[2] = nonce)
	// For legacy txs, nonce is at index 0
	nonceTyped := rawTx[2]  // works if tx is typed
	nonceLegacy := rawTx[0] // works if tx is legacy

	// Select the correct one
	nonce := api.Select(isTyped, nonceTyped, nonceLegacy)

	// return tx nonce
	// Q. is it okay to return nonce as frontend.Variable?
	return nonce
}

// ExtractNonceFromRLP extracts the nonce from RLP-encoded Ethereum transaction bytes.
//
// Ethereum transactions are serialized using Recursive Length Prefix (RLP) encoding,
// where the nonce is the first field in the transaction structure. This function
// parses the RLP bytes manually to extract the nonce without fully decoding the transaction.
//
// Parameters:
// - txBytes: A byte slice containing the RLP-encoded Ethereum transaction.
//
// Returns:
// - uint64: The nonce value extracted from the transaction.
// - error: An error if the input bytes are invalid or the RLP encoding is malformed.
//
// RLP Encoding Rules for the nonce:
//   - If the first byte is less than 0x80, the nonce is encoded as a single byte.
//   - If the first byte is between 0x80 and 0xb7, the nonce is length-prefixed (short encoding).
//     The first byte indicates the length of the nonce (1–55 bytes), followed by the nonce bytes.
//   - If the first byte is between 0xb8 and 0xbf, the nonce is length-prefixed (long encoding).
//     The first byte indicates the length of the length field (1–8 bytes), followed by the length
//     of the nonce, and then the nonce bytes.
//
// Example:
// - Input: RLP-encoded transaction bytes (e.g., []byte{0x85, 0x01, 0x02, 0x03})
// - Output: Nonce = 1
//
//
// ExtractNonceFromRLP extracts the nonce from RLP-encoded Ethereum transaction bytes.
// - Legacy Transactions: The nonce is the first field, so the field index is set to 0.
// - EIP-2930/EIP-1559 Transactions: The nonce is the second field, so the field index is set to 1.
// - The function dynamically adjusts the field index based on the transaction
// type.
//
// This function works for all Ethereum transaction types:
// - Legacy Transactions: The nonce is the first field in the RLP list.
// - EIP-2930 (Access List) Transactions: The nonce is the second field in the RLP list.
// - EIP-1559 (Dynamic Fee) Transactions: The nonce is the second field in the RLP list.
//
// Detailed description for each transaction type:
// Dynamic Fee Transactions (EIP-1559):
// - The first byte is the transaction type (0x02).
// - The nonce is the second field in the RLP list (after the chain ID).
// - ExtractNonceFromRLP skips the transaction type byte and extracts the second
// field.
// Access List Transactions (EIP-2930):
// - The first byte is the transaction type (0x01).
// - The nonce is the second field in the RLP list (after the chain ID).
// - ExtractNonceFromRLP skips the transaction type byte and extracts the second field.
// Legacy Transactions with Replay Protection (EIP-155):
// - There is no transaction type byte.
// - The nonce is the first field in the RLP list.
// - ExtractNonceFromRLP extracts the first field directly.
// Legacy Transactions without Replay Protection (Homestead):
// - There is no transaction type byte.
// - The nonce is the first field in the RLP list.
// - ExtractNonceFromRLP extracts the first field directly.
//
// Steps:
// 1. Check the transaction type (first byte):
//   - If the first byte is 0x01 (EIP-2930) or 0x02 (EIP-1559), skip the transaction type byte.
//   - Set the field index for the nonce accordingly (0 for Legacy, 1 for EIP-2930/EIP-1559).
// 2. Parse the RLP list to calculate the offset of the first field.
// 3. Extract the nonce field using the calculated offset and field index.
// 4. Convert the extracted nonce bytes into a uint64.

func ExtractNonceFromRLP(txBytes []byte) (uint64, error) {
	if len(txBytes) == 0 {
		return 0, errors.New("empty transaction bytes")
	}

	// Check for transaction type (EIP-2930 or EIP-1559)
	txType := txBytes[0]
	var fieldIndex int
	if txType == 0x01 || txType == 0x02 {
		// Skip the transaction type byte
		txBytes = txBytes[1:]
		fieldIndex = 1 // For EIP-2930 and EIP-1559, nonce is the second field
	} else {
		fieldIndex = 0 // For legacy transactions, nonce is the first field
	}

	// Parse the RLP list
	offset, err := parseRLPList(txBytes)
	if err != nil {
		return 0, err
	}

	// Extract the nonce
	nonceBytes, err := extractRLPField(txBytes, offset, fieldIndex)
	if err != nil {
		return 0, err
	}

	// Convert the byte array to uint64
	return bytesToUint64(nonceBytes), nil
}

func ExtractNonceFromRLPZk(api frontend.API, rawTx []frontend.Variable) frontend.Variable {
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

// parseRLPList parses the RLP list and calculates the offset of the first field.
// - The function correctly identifies whether the list is short or long based on the first byte.
// - It calculates the offset of the first field based on the length of the
// list.
//
// RLP lists are encoded as follows:
// - Short lists (length <= 55 bytes): The first byte is between 0xc0 and 0xf7.
//   - The length of the list is encoded in the first byte (0xc0 + length).
//   - The offset of the first field is 1 (after the length byte).
//
// - Long lists (length > 55 bytes): The first byte is between 0xf8 and 0xff.
//   - The length of the length field is encoded in the first byte (0xf7 + length of length).
//   - The actual length of the list is encoded in the next bytes.
//   - The offset of the first field is after the length bytes.

func parseRLPList(data []byte) (int, error) {
	if len(data) == 0 {
		return 0, errors.New("empty RLP data")
	}

	firstByte := data[0]
	if firstByte < 0xc0 {
		return 0, errors.New("not an RLP list")
	}

	if firstByte <= 0xf7 {
		// Short list
		return 1, nil // Offset starts after the length byte
	} else {
		// Long list
		lengthOfLength := int(firstByte - 0xf7)
		if len(data) < 1+lengthOfLength {
			return 0, errors.New("invalid RLP encoding: insufficient bytes for list length")
		}
		return 1 + lengthOfLength, nil // Offset starts after the length bytes
	}
}

func parseRLPListZk(api frontend.API, rawTx []frontend.Variable) frontend.Variable {
	txType := rawTx[0]

	// Check if the transaction is typed (Access List or Dynamic Fee)
	isAccessList := api.IsZero(api.Sub(txType, 0x01)) // txType == 0x01
	isDynamicFee := api.IsZero(api.Sub(txType, 0x02)) // txType == 0x02
	isTyped := api.Or(isAccessList, isDynamicFee)

	// Hardcoded offsets for nonce
	offsetTyped := frontend.Variable(2)  // Offset for typed transactions
	offsetLegacy := frontend.Variable(0) // Offset for legacy transactions

	// Select the correct offset based on the transaction type
	offset := api.Select(isTyped, offsetTyped, offsetLegacy)

	return offset
}

func parseRLPListZkGeneral(api frontend.API, rawTx []frontend.Variable) frontend.Variable {
	firstByte := rawTx[0]

	// Simulate comparisons for short list
	isGreaterOrEqualC0 := api.Sub(api.Cmp(firstByte, 0xc0), -1) // 1 if firstByte >= 0xc0
	isLessOrEqualF7 := api.Sub(api.Cmp(0xf7, firstByte), -1)    // 1 if firstByte <= 0xf7
	isShortList := api.And(isGreaterOrEqualC0, isLessOrEqualF7)

	// Simulate comparisons for long list
	isGreaterOrEqualF8 := api.Sub(api.Cmp(firstByte, 0xf8), -1) // 1 if firstByte >= 0xf8
	isLessOrEqualFF := api.Sub(api.Cmp(0xff, firstByte), -1)    // 1 if firstByte <= 0xff
	isLongList := api.And(isGreaterOrEqualF8, isLessOrEqualFF)

	// Offset calculation
	shortListOffset := frontend.Variable(1)                                   // Offset for short lists
	longListOffset := api.Add(frontend.Variable(1), api.Sub(firstByte, 0xf7)) // Offset for long lists

	// Select the correct offset based on the list type
	offset := api.Select(isShortList, shortListOffset, api.Select(isLongList, longListOffset, frontend.Variable(0)))

	// Assert that the first byte is valid (either short list or long list)
	api.AssertIsBoolean(api.Or(isShortList, isLongList))

	return offset
}

// extractRLPField extracts the nth field from an RLP list.
// - This function handles all possible RLP field types (single-byte, short, and long).
// - It iterates through the fields in the list and stops when the specified
// field index is reached.
//
// RLP fields are encoded as follows:
// - Single-byte values (0x00 to 0x7f): The value is encoded directly in the first byte.
// - Short length-prefixed values (0x80 to 0xb7): The first byte indicates the length of the value (0x80 + length).
//   - The actual value follows the length byte.
//
// - Long length-prefixed values (0xb8 to 0xbf): The first byte indicates the length of the length field (0xb7 + length of length).
//   - The actual length of the value is encoded in the next bytes.
//   - The value follows the length bytes.
//
// Steps:
// 1. Iterate through the fields in the RLP list until the specified field index is reached.
// 2. For each field, determine its type (single-byte, short length-prefixed, or long length-prefixed).
// 3. Extract the field and return it as a byte slice.

func extractRLPField(data []byte, offset int, fieldIndex int) ([]byte, error) {
	for i := 0; i <= fieldIndex; i++ {
		if offset >= len(data) {
			return nil, errors.New("invalid RLP encoding: insufficient bytes for field")
		}

		firstByte := data[offset]
		if firstByte < 0x80 {
			// Single-byte value
			if i == fieldIndex {
				return data[offset : offset+1], nil
			}
			offset++
		} else if firstByte <= 0xb7 {
			// Short length-prefixed value
			length := int(firstByte - 0x80)
			if offset+1+length > len(data) {
				return nil, errors.New("invalid RLP encoding: insufficient bytes for field")
			}
			if i == fieldIndex {
				return data[offset+1 : offset+1+length], nil
			}
			offset += 1 + length
		} else if firstByte <= 0xbf {
			// Long length-prefixed value
			lengthOfLength := int(firstByte - 0xb7)
			if offset+1+lengthOfLength > len(data) {
				return nil, errors.New("invalid RLP encoding: insufficient bytes for field length")
			}
			length := int(bytesToUint64(data[offset+1 : offset+1+lengthOfLength]))
			if offset+1+lengthOfLength+length > len(data) {
				return nil, errors.New("invalid RLP encoding: insufficient bytes for field")
			}
			if i == fieldIndex {
				return data[offset+1+lengthOfLength : offset+1+lengthOfLength+length], nil
			}
			offset += 1 + lengthOfLength + length
		} else {
			return nil, errors.New("invalid RLP encoding: unexpected byte")
		}
	}

	return nil, errors.New("field index out of range")
}

func extractRLPFieldZkGeneral(api frontend.API, rawTx []frontend.Variable, offset frontend.Variable) frontend.Variable {
	// Get the first byte dynamically
	firstByte := getValueAtOffset(api, rawTx, offset)

	// Simulate comparisons for field type
	isSingleByte := api.IsZero(api.Sub(frontend.Variable(0x80), firstByte)) // 1 if firstByte < 0x80
	isShortLength := api.And(
		api.IsZero(api.Sub(firstByte, frontend.Variable(0x80))),
		api.IsZero(api.Sub(frontend.Variable(0xb7), firstByte)),
	) // 1 if 0x80 <= firstByte <= 0xb7
	isLongLength := api.And(
		api.IsZero(api.Sub(firstByte, frontend.Variable(0xb8))),
		api.IsZero(api.Sub(frontend.Variable(0xbf), firstByte)),
	) // 1 if 0xb8 <= firstByte <= 0xbf

	// Extract single-byte value
	singleByteValue := firstByte

	// Extract short length-prefixed value
	// shortLength := api.Sub(firstByte, frontend.Variable(0x80))
	shortValue := getValueAtOffset(api, rawTx, api.Add(offset, frontend.Variable(1)))

	// Extract long length-prefixed value
	longLengthOfLength := api.Sub(firstByte, frontend.Variable(0xb7))
	// longLength := getValueAtOffset(api, rawTx, api.Add(offset, frontend.Variable(1)))
	longValue := getValueAtOffset(api, rawTx, api.Add(offset, api.Add(frontend.Variable(1), longLengthOfLength)))

	// Select the correct value based on the field type
	nonce := api.Select(isSingleByte, singleByteValue, api.Select(isShortLength, shortValue, api.Select(isLongLength, longValue, frontend.Variable(0))))

	return nonce
}

func getValueAtOffset(api frontend.API, array []frontend.Variable, offset frontend.Variable) frontend.Variable {
	// Initialize the selected value to zero
	selectedValue := frontend.Variable(0)

	// Iterate over all possible indices in the array
	for i := 0; i < len(array); i++ {
		// Check if the current index matches the offset
		isMatch := api.IsZero(api.Sub(offset, frontend.Variable(i)))

		// Select the value at the current index if it matches the offset
		selectedValue = api.Select(isMatch, array[i], selectedValue)
	}

	return selectedValue
}

// bytesToUint64 converts a byte slice to a uint64.
// - RLP encodes integers as big-endian byte arrays.
// - This function reconstructs the integer by processing the bytes in order.
// Steps:
// 1. Iterate through the byte slice.
// 2. Shift the result left by 8 bits and OR it with the current byte.

func bytesToUint64(b []byte) uint64 {
	var result uint64
	for _, byte := range b {
		result = (result << 8) | uint64(byte)
	}
	return result
}
