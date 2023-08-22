package hashtypes

import (
	"fmt"
	"io"
	"math/big"

	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
)

// Abstract representation of a MiMC hash
type Digest [32]byte

// size of the field, used for the input size of Mimc
const (
	NumBytes = field.Bytes
)

// Writes the Digest
func (d Digest) WriteTo(b io.Writer) (int64, error) {
	_, err := b.Write(d[:])
	if err != nil {
		panic(err) // hard forbid any error
	}
	return 32, nil
}

func BytesToDigest(b []byte) (d Digest) {
	// Sanity-check the length of the digest
	if len(b) != len(Digest{}) {
		utils.Panic("Passed a string of %v bytes but expected %v", len(b), Digest{})
	}
	copy(d[:], b)
	return d
}

// Big int are assumed to fit on 32 bytes and are written as a single
// block of 32 bytes in bigendian form (i.e, zero-padded on the left)
func WriteBigIntTo(w io.Writer, b *big.Int) (int64, error) {
	balanceBig := b.FillBytes(make([]byte, 32))
	n, err := w.Write(balanceBig)
	return int64(n), err
}

// Writes an integer as a 32bytes word like big ints
func WriteInt64To(w io.Writer, i int64) (int64, error) {
	b := big.NewInt(int64(i))
	return WriteBigIntTo(w, b)
}

/*
Cmp two digests. The digest are interpreted as big-endians big integers and
then are compared. Returns:
  - a < b : -1
  - a == b : 0
  - a > b : 1
*/
func Cmp(a, b Digest) int {
	var bigA, bigB big.Int
	bigA.SetBytes(a[:])
	bigB.SetBytes(b[:])
	return bigA.Cmp(&bigB)
}

/*
Returns the max of two digest. The digest are interpreted as big-endian big
integers.
*/
func Max(a, b Digest) Digest {
	if Cmp(a, b) == 1 {
		return a
	}
	return b
}

/*
Returns the min of two digest. The digest are interpreted as big-endian big
integers.
*/
func Min(a, b Digest) Digest {
	if Cmp(a, b) == -1 {
		return a
	}
	return b
}

// Returns an hexstring representation of the digest
func (d Digest) Hex() string {
	return fmt.Sprintf("0x%x", [32]byte(d))
}

// Constructs a dummy digest from an integer
func DummyDigest(i int) (d Digest) {
	d[31] = byte(i)
	return d
}
