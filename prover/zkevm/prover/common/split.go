package common

import "encoding/binary"

// LimbBytes is the size of one limb in bytes
const LimbBytes = 2

// SplitBytes splits the input slice into subarrays of the provided size.
func SplitBytes(input []byte) [][]byte {
	if len(input) == 0 {
		return [][]byte{}
	}

	var result [][]byte
	for i := 0; i < len(input); i += LimbBytes {
		end := i + LimbBytes
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
	return SplitBytes(inputBytes)
}

// SplitLittleEndianUint64 splits the uint64 input into little endian subarrays of the provided size.
func SplitLittleEndianUint64(input uint64) [][]byte {
	inputBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(inputBytes, input)
	return SplitBytes(inputBytes)
}
