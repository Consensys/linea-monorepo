package hash

import (
	"encoding/binary"
	"encoding/hex"

	"github.com/consensys/accelerated-crypto-monorepo/utils"
)

// padding is applied to pad the message to a multiple
// of the rate, which involves adding a "1" bit, zero or more "0" bits, and
// a final "1" bit.
func pad(a []byte) []byte {
	if len(a) == 0 {
		panic("the input is empty")
	}
	n := (Rate - (len(a) % Rate))

	// generate 10^*1
	x := []byte{0x01}
	for i := 1; i < n; i++ {
		x = append(x, 0)
	}
	x[n-1] ^= 0x80

	a = append(a, x...)

	//sanity checks
	if len(x) != n {
		panic("padding is not correct")
	}

	if len(a)%Rate != 0 {
		utils.Panic("length extension is not correctly done, len(a) = %v should be multiple of Rate = %v", len(a), Rate)
	}

	return a
}

// decodeHex converts a hex-encoded string into a raw byte string.
func DecodeHex(s string) []byte {
	b, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return b
}

// it absorb a block of [Rate]Byte to the state
func (s *State) absorbBlock(blockByte []byte) {

	aUint64 := s.aUint64
	// separate each 64 bits=8byte
	var c [Rate / 8]uint64
	for k := 0; k < len(blockByte); k += 8 {
		b := blockByte[k : k+8]
		c[k/8] = binary.LittleEndian.Uint64(b)
	}
	for k := range c {
		j := k / 5
		i := k % 5
		if 5*j+i != k {
			panic("error")
		}
		aUint64[i][j] = aUint64[i][j] ^ c[k]
	}
	s.aUint64 = aUint64

}

func messageBlocks(t []byte) [][]byte {
	//cut each 136 byte to 17 Field Element

	blockByte := make([][]byte, len(t)/Rate)
	for m := 0; m < len(t); m += Rate {
		p := t[m : m+Rate]
		blockByte[m/Rate] = p
	}
	return blockByte
}

type hashInput []byte

// it compute the number of keccak permutation needed for a batch of hashes
func NumberofKeccakf(data []hashInput) int {
	np := 0
	for u := range data {
		np += len(data[u])

	}
	return np
}
func isStateZero(a [5][5]uint64) (b bool) {
	b = true
	for i := 0; i < 5; i++ {
		for j := 0; j < 5; j++ {
			if a[i][j] != 0 {
				b = false
			}
		}
	}
	return b
}
