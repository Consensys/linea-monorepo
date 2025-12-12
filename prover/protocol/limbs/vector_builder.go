package limbs

import (
	"math/big"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

// VectorBuilder is a helper to iteratively construct the assignment of a column
// by iteratively pushing values and then pad.
type VectorBuilder[E Endianness] struct {
	limbs  limbs[E]
	slices [][]field.Element
}

// Height returns the total number of rows that have been pushed
func (b *VectorBuilder[E]) Height() int {
	return len(b.slices)
}

// Push adds a new row to the builder. This does not check for the endianness.
func (b *VectorBuilder[E]) Push(row row[E]) {
	if len(row) != b.limbs.NumLimbs() {
		panic("wrong number of columns")
	}
	for i := range row {
		b.slices[i] = append(b.slices[i], row[i])
	}
}

// PushBigInt pushes a new big.Int to the builder
func (b *VectorBuilder[E]) PushBigInt(x *big.Int) {
	row := bigIntToLimbs[E](x, b.limbs.BitSize())
	b.Push(row)
}

// PushBytes pushes a new bytes to the builder
func (b *VectorBuilder[E]) PushBytes(x []byte) {
	row := bytesToLimbs[E](x)
	b.Push(row)
}

// PushZero pushes a new zero to the builder
func (b *VectorBuilder[E]) PushZero() {
	row := make([]field.Element, b.limbs.NumLimbs())
	b.Push(row)
}

// PushOne pushes a row storing the big integer one, respecting the endianness
// of the limbs.
func (b *VectorBuilder[E]) PushOne() {
	row := bigIntToLimbs[E](big.NewInt(1), b.limbs.BitSize())
	b.Push(row)
}

// PushInt pushes a small integer to the builder. It will spread it into limbs
// if needed.
func (b *VectorBuilder[E]) PushInt(x int) {
	row := bigIntToLimbs[E](big.NewInt(int64(x)), b.limbs.BitSize())
	b.Push(row)
}

// PeekAt returns the last pushed row in native form.
func (b *VectorBuilder[E]) PeekAt(r int) row[E] {
	row := make([]field.Element, r)
	for i := range row {
		row[i] = b.slices[i][r]
	}
	return row
}

// PeekBigIntAt returns the last pushed row in big.Int form.
func (b *VectorBuilder[E]) PeekBigIntAt(r int) *big.Int {
	return limbToBigInt[E](b.PeekAt(r))
}

// PeekBytesAt returns the last pushed row in bytes form.
func (b *VectorBuilder[E]) PeekBytesAt(r int) []byte {
	return limbsToBytes[E](b.PeekAt(r))
}

// PeekLast returns the last pushed row in native form.
func (b *VectorBuilder[E]) PeekLast() row[E] {
	return b.PeekAt(len(b.slices[0]) - 1)
}

// RepushLast pushes a value equal to the last pushed value of `vb`
func (b *VectorBuilder[E]) RepushLast() {
	if len(b.slices[0]) == 0 {
		panic("attempted to repush the last item of an empty builder")
	}
	last := b.PeekAt(len(b.slices[0]) - 1)
	b.Push(last)
}

// PushInc repushes that last element incremented by 1
func (b *VectorBuilder[E]) PushInc() {
	last := b.PeekBigIntAt(len(b.slices[0]) - 1)
	last.Add(last, big.NewInt(1))
	b.PushBigInt(last)
}

// PadAndAssignZero assigns the content of the builder to the column.
func (b *VectorBuilder[E]) PadAndAssignZero(run *wizard.ProverRuntime) {
	for i := range b.slices {
		run.AssignColumn(
			b.limbs.C[i].GetColID(),
			smartvectors.RightZeroPadded(b.slices[i], b.limbs.C[i].Size()),
		)
	}
}
