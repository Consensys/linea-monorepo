package limbs

import (
	"math/big"
	"slices"

	"github.com/consensys/linea-monorepo/prover/maths/field"
)

// Since row is not constructible. We still expose an alias for vectors of
// rows.
type VecRow[E Endianness] = []row[E]

// row represents a single row in the assignment of a limb register. The type
// may only be constructed by converting explicit types.
type row[E Endianness] struct {
	T []field.Element
	_ E
}

// NumLimbs returns the number of limbs in the row
func (r row[E]) NumLimbs() int {
	return len(r.T)
}

// RowFromBigInt converts a big.Int into a row. The mapping between int and
// bytes is always big endian. But the mapping between bytes and limbs is E.
func RowFromBigInt[E Endianness](bi *big.Int, bitSize int) row[E] {
	raw := bigIntToLimbs[E](bi, bitSize)
	return row[E]{T: raw}
}

// RowFromInt converts an int into a row. The mapping between int and bytes is
// always big endian. But the mapping between bytes and limbs is E.
func RowFromInt[E Endianness](x int, bitSize int) row[E] {
	raw := bigIntToLimbs[E](big.NewInt(int64(x)), bitSize)
	return row[E]{T: raw}
}

// RowFromBytes converts bytes into a row
func RowFromBytes[E Endianness](x []byte) row[E] {
	raw := bytesToLimbs[E](x)
	return row[E]{T: raw}
}

// RowFromKoala converts a koalabear element into a row. The mapping between
// koalabear and bytes is always big endian. But the mapping between bytes and
// limbs is E.
func RowFromKoala[E Endianness](x field.Element, bitSize int) row[E] {
	xInt := int(x.Uint64())
	return RowFromInt[E](xInt, bitSize)
}

// ToBigEndianLimbs returns the row in big endian form
func (r row[E]) ToBigEndianLimbs() row[BigEndian] {
	t := slices.Clone(r.T)
	if isLittleEndian[E]() {
		slices.Reverse(t)
	}
	return row[BigEndian]{T: t}
}

// ToLittleEndianLimbs returns the row in little endian form
func (r row[E]) ToLittleEndianLimbs() row[LittleEndian] {
	t := slices.Clone(r.T)
	if isBigEndian[E]() {
		slices.Reverse(t)
	}
	return row[LittleEndian]{T: t}
}

// ToBigInt converts the row into a big.Integer assuming big-endian order for
// the mapping (underlying bytes) -> (big.Int).
func (r row[E]) ToBigInt() *big.Int {
	return limbToBigInt[E](r.T)
}

// ToBytes converts the row into bytes assuming big-endian order for the
// mapping (underlying bytes) -> (big.Int).
func (r row[E]) ToBytes() []byte {
	return limbsToBytes[E](r.T)
}

// ToBytes16 converts the row into bytes assuming big-endian order for the
// mapping (underlying bytes) -> (big.Int).
func (r row[E]) ToBytes16() [16]byte {
	s := limbsToBytes[E](r.T)
	return [16]byte(s)
}

// ToBytes32 converts the row into bytes assuming big-endian order for the
// mapping (underlying bytes) -> (big.Int).
func (r row[E]) ToBytes32() [32]byte {
	s := limbsToBytes[E](r.T)
	return [32]byte(s)
}

// SplitOnBit splits the row into two rows. The first row contains the first
// half of the limbs and the second row contains the second half of the limbs.
func (r row[E]) SplitOnBit(at int) (row[E], row[E]) {
	buf := r.ToBytes()
	hi, lo := buf[:at/8], buf[at/8:]
	return RowFromBytes[E](hi), RowFromBytes[E](lo)
}
