package limbs

import (
	"math/big"

	"github.com/consensys/linea-monorepo/prover/maths/field"
)

// row represents a single row in the assignment of a limb register. The type
// may only be constructed by converting explicit types.
type row[E Endianness] []field.Element

// NumLimbs returns the number of limbs in the row
func (r row[E]) NumLimbs() int {
	return len(r)
}

func RowFromBigInt[E Endianness](bi *big.Int, bitSize int) row[E] {
	return bigIntToLimbs[E](bi, bitSize)
}

func RowFromInt[E Endianness](x int, bitSize int) row[E] {
	return bigIntToLimbs[E](big.NewInt(int64(x)), bitSize)
}

func RowFromBytes[E Endianness](x []byte) row[E] {
	return bytesToLimbs[E](x)
}
