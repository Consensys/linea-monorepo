package fiatshamir

import "github.com/consensys/accelerated-crypto-monorepo/utils"

// BitReader allows to read slices of bit of designed size
type BitReader struct {
	numBits    int
	curBits    int
	underlying []byte
}

/*
Underlying is assumed to be a big endian field.
*/
func NewBitReader(underlying []byte, numBits int) BitReader {
	// Check numBits
	if 8*len(underlying) < numBits {
		utils.Panic("numBits %v can't be larger than the buffer size %v", numBits, 8*len(underlying))
	}

	return BitReader{
		numBits:    numBits,
		curBits:    0,
		underlying: bigToLittle(underlying),
	}
}

// `30` is chosen as a limit to prevent overflow
// issues due to reading an int
const READ_LIMIT int = 30

func (r *BitReader) ReadInt(n int) int {

	if n > READ_LIMIT {
		utils.Panic("Can't read more than %v, got %v", READ_LIMIT, n)
	}

	if n <= 0 {
		panic("Asked to read 0 or negative bytes, probably a mistake")
	}

	if n+r.curBits > r.numBits {
		panic("Overflowed the reader")
	}

	res := 0

	// Add the bytes one by one
	// Pretty inefficient but simple
	for i := 0; i < n; i++ {
		curr := r.curBits + i
		selected := r.underlying[curr/8]
		selected >>= curr % 8
		selected &= 1
		res |= int(selected) << i
	}

	r.curBits += n
	return res
}

/*
Converts out of place a sequence of bytes from bigToLittle

(i.e) just reverse the order
*/
func bigToLittle(b []byte) []byte {
	res := make([]byte, len(b))
	for i := range res {
		res[i] = b[len(b)-1-i]
	}
	return res
}
