package limbs

import (
	"math/big"
	"slices"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/distributed/pragmas"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
)

const (
	// limbBitWidth is the number of bits in a limb
	limbBitWidth = 16

	// limbByteWidth is the number of bytes in a limb
	limbByteWidth = limbBitWidth / 8
)

// limbs represents a register represented by a list of columns.
type limbs[E Endianness] struct {
	C []ifaces.Column
	_ E // this field is needed to tag the struct with E
}

// createLimb creates a set of columns in the wizard. The columns are always of
// type [column.Commitment] and are for the round zero.
func createLimb[E Endianness](
	comp *wizard.CompiledIOP,
	name ifaces.ColID,
	numLimbs, size int,
	prags ...pragmas.PragmaPair,
) limbs[E] {
	c := make([]ifaces.Column, numLimbs)
	for i := range c {
		cname := ifaces.ColIDf("%v_L%02v", name, i)
		c[i] = comp.InsertCommit(0, cname, size, true)
		for _, pragma := range prags {
			c := c[i].(column.Natural)
			c.SetPragma(pragma.Pragma, pragma.Value)
		}
	}
	return limbs[E]{C: c}
}

// Size returns the number of rows in the provided columns
func (l limbs[E]) Size() int {
	return l.C[0].Size()
}

// BitSize returns the total number of bits represented by the provided columns
func (l limbs[E]) BitSize() int {
	return len(l.C) * limbBitWidth
}

// NumLimbs returns the number of limbs in the provided columns
func (l limbs[E]) NumLimbs() int {
	return len(l.C)
}

// LimbBitWidth returns the number of bits in a limb, which is a constant?
func (l limbs[E]) LimbBitWidth() int {
	return limbBitWidth
}

// GetRowAsBytes returns the represented bytes for the provided field element. The
// function panics if the requested row is out of bound or if one of the columns
// has been called.
func (l limbs[E]) GetRowAsBytes(run ifaces.Runtime, row int) []byte {

	if row < 0 || row >= l.C[0].Size() {
		utils.Panic("row out of bound: %v, max %v", row, l.C[0].Size())
	}

	rowF := make([]field.Element, len(l.C))
	for i := range l.C {
		rowF[i] = l.C[i].GetColAssignmentAt(run, row)
	}

	return limbsToBytes[E](rowF)
}

// GetAssignmentAsBytes returns the represented bytes for the provided field
// elements.
func (l limbs[E]) GetAssignmentAsBytes(run ifaces.Runtime) [][]byte {
	res := make([][]byte, 0, l.Size())
	for i := 0; i < l.Size(); i++ {
		res = append(res, l.GetRowAsBytes(run, i))
	}
	return res
}

// AssignBytes assigns the provided bytes to the provided field elements.
func (l limbs[E]) AssignBytes(run *wizard.ProverRuntime, bytes [][]byte) {

	var (
		numLimbs = utils.DivExact(len(bytes[0]), limbByteWidth)
		numRow   = len(bytes)
	)

	if numLimbs != len(l.C) {
		utils.Panic("provided number of limbs must be equal to the number of bytes, got %v and %v", numLimbs, len(l.C))
	}

	if l.C[0].Size() != numRow {
		utils.Panic("number of bytes must be equal to the number of limbs, got %v and %v", len(bytes), len(l.C))
	}

	limbs := bytesToLimbsVec[E](bytes, numLimbs)

	for c := range l.C {
		run.AssignColumn(l.C[c].GetColID(), smartvectors.NewRegular(limbs[c]))
	}
}

// GetRowAsBigInt returns the represented big.Int for the provided field element.
func (l limbs[E]) GetRowAsBigInt(run ifaces.Runtime, row int) *big.Int {

	if row < 0 || row >= l.C[0].Size() {
		utils.Panic("row out of bound: %v, max %v", row, l.C[0].Size())
	}

	rowF := make([]field.Element, len(l.C))
	for i := range l.C {
		rowF[i] = l.C[i].GetColAssignmentAt(run, row)
	}

	return limbToBigInt[E](rowF)
}

// GetAssignmentAsBigInt returns the represented big.Int for the provided field
// elements.
func (l limbs[E]) GetAssignmentAsBigInt(run ifaces.Runtime) []*big.Int {
	res := make([]*big.Int, 0, l.Size())
	for i := 0; i < l.Size(); i++ {
		res = append(res, l.GetRowAsBigInt(run, i))
	}

	return res
}

// AssignBigInts assigns the provided big.Ints to the provided field elements.
func (l limbs[E]) AssignBigInts(run *wizard.ProverRuntime, bigints []*big.Int) {

	var (
		numRow      = len(bigints)
		numLimbs    = len(l.C)
		uintBitSize = numLimbs * limbBitWidth
	)

	if l.C[0].Size() != numRow {
		utils.Panic("number of bytes must be equal to the number of limbs, got %v and %v", numRow, len(l.C))
	}

	res := bigIntToLimbsVec[E](bigints, len(l.C), uintBitSize)

	for c := range l.C {
		run.AssignColumn(l.C[c].GetColID(), smartvectors.NewRegular(res[c]))
	}
}

// bytesToLimbsVec convert a vector of byteslices into a vector of limbs of form
// [E].
func bytesToLimbsVec[E Endianness](bytes [][]byte, numLimbs int) [][]field.Element {

	var (
		numRow = len(bytes)
		res    = make([][]field.Element, numLimbs)
	)

	for c := range res {
		res[c] = make([]field.Element, 0, numRow)
	}

	for r := range bytes {
		lbs := bytesToLimbs[E](bytes[r])
		for c := range res {
			res[c] = append(res[c], lbs[c])
		}
	}

	return res

}

// bigIntToLimbsVec converts a vector of big.Int into limbs of form [E]
func bigIntToLimbsVec[E Endianness](bigints []*big.Int, numLimbs, bitSize int) [][]field.Element {

	var (
		numRow = len(bigints)
		res    = make([][]field.Element, numLimbs)
	)

	for c := range res {
		res[c] = make([]field.Element, 0, numRow)
	}

	for r := range bigints {
		lbs := bigIntToLimbs[E](bigints[r], bitSize)
		for c := range res {
			res[c] = append(res[c], lbs[c])
		}
	}

	return res
}

// limbsBeToBytes converts a vector of limbs in form [E] into bytes
func limbsToBytes[E Endianness](limbs []field.Element) []byte {

	if isLittleEndian[E]() {
		limbs = slices.Clone(limbs)
		slices.Reverse(limbs)
	}

	res := make([]byte, 0, len(limbs)*limbByteWidth)
	for _, c := range limbs {
		cbytes := c.Bytes()
		res = append(res, cbytes[field.Bytes-limbByteWidth:]...)
	}
	return res
}

// bytesToLimbs converts a vector of bytes into limbs of form [E]
func bytesToLimbs[E Endianness](bytes []byte) []field.Element {

	var (
		numLimbs = len(bytes) / limbByteWidth
		limbs    = make([]field.Element, numLimbs)
		buf      = [field.Bytes]byte{}
	)

	for i := 0; i < numLimbs; i++ {
		copy(buf[field.Bytes-limbByteWidth:], bytes[i*limbByteWidth:(i+1)*limbByteWidth])
		if err := limbs[i].SetBytesCanonical(buf[:]); err != nil {
			utils.Panic("bytesToLimbs failed: %v", err)
		}
	}

	if isLittleEndian[E]() {
		slices.Reverse(limbs)
	}

	return limbs
}

// bigIntToLimbs converts a big.Int into limbs of form [E] using bitSize to
// determine the number of required limbs and corresponds to the maximal number
// or bits that can be used by bi.
func bigIntToLimbs[E Endianness](bi *big.Int, bitSize int) []field.Element {
	numLimbs := utils.DivCeil(bitSize, limbBitWidth)
	buf := make([]byte, limbByteWidth*numLimbs)
	bi.FillBytes(buf)
	return bytesToLimbs[E](buf)
}

// limbToBigInt converts a limb of form [E] into a big.Int
func limbToBigInt[E Endianness](limb []field.Element) *big.Int {
	x := limbsToBytes[E](limb)
	return new(big.Int).SetBytes(x)
}
