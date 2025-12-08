package common

import (
	"encoding/binary"

	"github.com/consensys/linea-monorepo/prover/utils/types"
)

// LimbBytes is the size of one limb in bytes
const LimbBytes = 2

// SplitBytes splits the input slice into subarrays of the provided size.
func SplitBytes(input []byte, limbBytes ...int) [][]byte {
	limbSize := LimbBytes
	if len(limbBytes) > 0 {
		limbSize = limbBytes[0]
	}
	if len(input) == 0 {
		return [][]byte{}
	}

	var result [][]byte
	for i := 0; i < len(input); i += limbSize {
		end := i + limbSize
		if end > len(input) {
			end = len(input)
		}
		result = append(result, input[i:end])
	}
	return result
}

// SplitBigEndianUint64 splits the uint64 input into big endian subarrays of the provided size.
func SplitBigEndianUint64(input uint64) [][]byte {
	inputBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(inputBytes, input)
	// we left padd with zero bytes to
	// make it a 64 byte array.
	// Also accumulate 4 bytes per limb
	res := types.LeftPadded48Zeros(types.LeftPadded(inputBytes[:]))
	return SplitBytes(res, 4)
}

// SplitLittleEndianUint64 splits the uint64 input into little endian subarrays of the provided size.
func SplitLittleEndianUint64(input uint64) [][]byte {
	inputBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(inputBytes, input)
	return SplitBytes(inputBytes)
}
