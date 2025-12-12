package limbs

import (
	"math/big"
	"slices"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/distributed/pragmas"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
)

const (
	// limbBitWidth is the number of bits in a limb
	limbBitWidth = 16

	// limbByteWidth is the number of bytes in a limb
	limbByteWidth = limbBitWidth / 8
)

// Limbed is an interface for groups of columns collectively representing
// limbs.
type Limbed interface {
	symbolic.Metadata
	NumRow() int
	NumLimbs() int
	ToBigEndian() Limbed
	ToLittleEndian() Limbed
	Limbs() []ifaces.Column
}

// limbs represents a register represented by a list of columns.
type limbs[E Endianness] struct {
	c    []ifaces.Column
	name ifaces.ColID
	_    E // this field is needed to tag the struct with E
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
	return limbs[E]{c: c, name: name}
}

// Size returns the number of rows in the provided columns
func (l limbs[E]) Size() int {
	return l.c[0].Size()
}

// BitSize returns the total number of bits represented by the provided columns
func (l limbs[E]) BitSize() int {
	return len(l.c) * limbBitWidth
}

// NumLimbs returns the number of limbs in the provided columns
func (l limbs[E]) NumLimbs() int {
	return len(l.c)
}

// LimbBitWidth returns the number of bits in a limb, which is a constant?
func (l limbs[E]) LimbBitWidth() int {
	return limbBitWidth
}

// GetRow returns the typed row for the provided field element.
func (l limbs[E]) GetRow(run ifaces.Runtime, r int) row[E] {
	if r < 0 || r >= l.c[0].Size() {
		utils.Panic("row out of bound: %v, max %v", r, l.c[0].Size())
	}
	rowF := make(row[E], len(l.c))
	for i := range l.c {
		rowF[i] = l.c[i].GetColAssignmentAt(run, r)
	}
	return rowF
}

// GetRowAsBytes returns the represented bytes for the provided field element. The
// function panics if the requested row is out of bound or if one of the columns
// has been called.
func (l limbs[E]) GetRowAsBytes(run ifaces.Runtime, row int) []byte {
	rowF := l.GetRow(run, row)
	return limbsToBytes[E](rowF)
}

// GetRowAsBigInt returns the represented big.Int for the provided field element.
func (l limbs[E]) GetRowAsBigInt(run ifaces.Runtime, row int) *big.Int {
	rowF := l.GetRow(run, row)
	return limbToBigInt[E](rowF)
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

// GetAssignmentAsBigInt returns the represented big.Int for the provided field
// elements.
func (l limbs[E]) GetAssignmentAsBigInt(run ifaces.Runtime) []*big.Int {
	res := make([]*big.Int, 0, l.Size())
	for i := 0; i < l.Size(); i++ {
		res = append(res, l.GetRowAsBigInt(run, i))
	}
	return res
}

// AssignBytes assigns the provided bytes to the provided field elements.
func (l limbs[E]) AssignBytes(run *wizard.ProverRuntime, bytes [][]byte) {

	var (
		numLimbs = utils.DivExact(len(bytes[0]), limbByteWidth)
		numRow   = len(bytes)
	)

	if numLimbs != len(l.c) {
		utils.Panic("provided number of limbs must be equal to the number of bytes, got %v and %v", numLimbs, len(l.c))
	}

	if l.c[0].Size() != numRow {
		utils.Panic("number of bytes must be equal to the number of limbs, got %v and %v", len(bytes), len(l.c))
	}

	limbs := bytesToLimbsVec[E](bytes, numLimbs)

	for c := range l.c {
		run.AssignColumn(l.c[c].GetColID(), smartvectors.NewRegular(limbs[c]))
	}
}

// AssignBigInts assigns the provided big.Ints to the provided field elements.
func (l limbs[E]) AssignBigInts(run *wizard.ProverRuntime, bigints []*big.Int) {

	var (
		numRow      = len(bigints)
		numLimbs    = len(l.c)
		uintBitSize = numLimbs * limbBitWidth
	)

	if l.c[0].Size() != numRow {
		utils.Panic("number of bytes must be equal to the number of limbs, got %v and %v", numRow, len(l.c))
	}

	res := bigIntToLimbsVec[E](bigints, len(l.c), uintBitSize)

	for c := range l.c {
		run.AssignColumn(l.c[c].GetColID(), smartvectors.NewRegular(res[c]))
	}
}

// ToBigEndian returns the limbs in big endian form
func (l limbs[E]) ToBigEndian() limbs[BigEndian] {
	new := limbs[BigEndian]{name: l.name, c: make([]ifaces.Column, len(l.c))}
	copy(new.c, l.c)
	if isLittleEndian[E]() {
		slices.Reverse(new.c)
	}
	return new
}

// ToLittleEndian returns the limbs in little endian form
func (l limbs[E]) ToLittleEndian() limbs[LittleEndian] {
	new := limbs[LittleEndian]{name: l.name, c: make([]ifaces.Column, len(l.c))}
	copy(new.c, l.c)
	if isBigEndian[E]() {
		slices.Reverse(new.c)
	}
	return new
}

// String implements the [github.com/consensys/linea-monorepo/prover/symbolic.Metadata]
// interface.
func (l limbs[E]) String() string {
	return string(l.name)
}
