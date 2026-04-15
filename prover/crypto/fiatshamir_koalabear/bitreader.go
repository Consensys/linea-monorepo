package fiatshamir_koalabear

// BitReader is a utility to read a slice of bytes bit by bit. It should be
// initialized via the NewBitReader method. Its main use-case is to help
// converting generate random field elements into a list of bounded integers
// as the list of random column to open in Vortex.
type BitReader struct {
	numBits    int
	curBits    int
	underlying []byte
}

// readLimit characterizes the number of bits that can be read in a single go
// by the [BitReader.ReadInt].
const readLimit int = 30
