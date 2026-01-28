package limbs

import (
	"math/big"
	"slices"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/column/verifiercol"
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

var (
	_ Limbed = (*Limbs[LittleEndian])(nil)
)

// Limbed is an interface for groups of columns collectively representing
// limbs.
type Limbed interface {
	symbolic.Metadata
	NumRow() int
	NumLimbs() int
	ToBigEndianLimbs() Limbs[BigEndian]
	ToLittleEndianLimbs() Limbs[LittleEndian]
	Limbs() []ifaces.Column
	ColumnNames() []string
}

type limbs[E Endianness] = Limbs[E]

// Limbs represents a register represented by a list of columns.
type Limbs[E Endianness] struct {
	c    []ifaces.Column
	name ifaces.ColID
	_    E // this field is needed to tag the struct with E
}

const (
	NbLimbU32        = 2
	NbLimbU64        = 4
	NbLimbU128       = 8
	NbLimbEthAddress = 10
	NbLimbU256       = 16
)

// NumLimbsOf returns the number of limbs for the provided size
func NumLimbsOf[S BitSize]() int {
	return utils.DivExact(uintSize[S](), limbBitWidth)
}

// NewLimbs creates a set of columns in the wizard. The columns are always of
// type [column.Commitment] and are for the round zero.
func NewLimbs[E Endianness](comp *wizard.CompiledIOP, name ifaces.ColID,
	numLimbs, size int, prags ...pragmas.PragmaPair) Limbs[E] {
	c := make([]ifaces.Column, numLimbs)
	for i := range c {
		cname := ifaces.ColIDf("%v_%v", name, i)
		c[i] = comp.InsertCommit(0, cname, size, true)
		for _, pragma := range prags {
			c := c[i].(column.Natural)
			c.SetPragma(pragma.Pragma, pragma.Value)
		}
	}
	return Limbs[E]{c: c, name: name}
}

// KoalaAsLimb creates a limb object representing a large bigint equal to the
// provided small column. Namely, it zero-extends the input following the
// provided endianness.
func KoalaAsLimb[E Endianness](col ifaces.Column, bitSize int) Limbs[E] {

	numLimbs := utils.DivExact(bitSize, limbBitWidth)
	res := make([]ifaces.Column, numLimbs)
	for i := range numLimbs {
		res[i] = verifiercol.NewConstantCol(field.Zero(), col.Size(), "0")
	}

	switch any(E{}).(type) {
	case LittleEndian:
		res[0] = col
	case BigEndian:
		res[numLimbs-1] = col
	}

	return Limbs[E]{c: res, name: col.GetColID()}
}

// Size returns the number of rows in the provided columns
func (l Limbs[E]) Size() int {
	return l.c[0].Size()
}

// BitSize returns the total number of bits represented by the provided columns
func (l Limbs[E]) BitSize() int {
	return len(l.c) * limbBitWidth
}

// NumLimbs returns the number of limbs in the provided columns
func (l Limbs[E]) NumLimbs() int {
	return len(l.c)
}

// LimbBitWidth returns the number of bits in a limb, which is a constant.
func (l Limbs[E]) LimbBitWidth() int {
	return limbBitWidth
}

// Limbs returns the raw limbs of the [limbs] object.
func (l Limbs[E]) Limbs() []ifaces.Column {
	return l.c
}

// LimbsArr2 returns a fixed sized array of limbs.
func (l Limbs[E]) LimbsArr2() [2]ifaces.Column {
	return [2]ifaces.Column(l.c)
}

// LimbsArr3 returns a fixed sized array of limbs.
func (l Limbs[E]) LimbsArr3() [3]ifaces.Column {
	return [3]ifaces.Column(l.c)
}

// LimbsArr4 returns a fixed sized array of limbs.
func (l Limbs[E]) LimbsArr4() [4]ifaces.Column {
	return [4]ifaces.Column(l.c)
}

// LimbsArr8 returns a fixed sized array of limbs.
func (l Limbs[E]) LimbsArr8() [8]ifaces.Column {
	return [8]ifaces.Column(l.c)
}

// LimbsArr10 returns the fixed sized array of limbs.
func (l Limbs[E]) LimbsArr10() [10]ifaces.Column {
	return [10]ifaces.Column(l.c)
}

// LimbsArr16 returns a fixed sized array of limbs.
func (l Limbs[E]) LimbsArr16() [16]ifaces.Column {
	return [16]ifaces.Column(l.c)
}

// NumRow returns the total number of rows of the [limbs] object.
func (l Limbs[E]) NumRow() int {
	return l.c[0].Size()
}

// GetRow returns the typed row for the provided field element.
func (l Limbs[E]) GetRow(run ifaces.Runtime, r int) row[E] {
	if r < 0 || r >= l.c[0].Size() {
		utils.Panic("row out of bound: %v, max %v", r, l.c[0].Size())
	}
	rowF := make([]field.Element, len(l.c))
	for i := range l.c {
		rowF[i] = l.c[i].GetColAssignmentAt(run, r)
	}
	return row[E]{T: rowF}
}

// GetRowAsBytes returns the represented bytes for the provided field element. The
// function panics if the requested row is out of bound or if one of the columns
// has been called.
func (l Limbs[E]) GetRowAsBytes(run ifaces.Runtime, row int) []byte {
	rowF := l.GetRow(run, row)
	return limbsToBytes[E](rowF.T)
}

// GetRowAsBigInt returns the represented big.Int for the provided field element.
func (l Limbs[E]) GetRowAsBigInt(run ifaces.Runtime, row int) *big.Int {
	rowF := l.GetRow(run, row)
	return limbToBigInt[E](rowF.T)
}

// GetAssignment returns the assignment in the form of endianness-tagged rows
func (l Limbs[E]) GetAssignment(run ifaces.Runtime) []row[E] {
	res := make([]row[E], 0, l.Size())
	for i := 0; i < l.Size(); i++ {
		res = append(res, l.GetRow(run, i))
	}
	return res
}

// GetAssignmentAsBytes returns the represented bytes for the provided field
// elements.
func (l Limbs[E]) GetAssignmentAsBytes(run ifaces.Runtime) [][]byte {
	res := make([][]byte, 0, l.Size())
	for i := 0; i < l.Size(); i++ {
		res = append(res, l.GetRowAsBytes(run, i))
	}
	return res
}

// GetAssignmentAsByte16Exact returns the represented bytes for the provided field
// elements. It will panic if the number of bytes stored in each row is not
// exactly 16.
func (l Limbs[E]) GetAssignmentAsByte16Exact(run ifaces.Runtime) [][16]byte {
	res := make([][16]byte, 0, l.Size())
	if l.NumLimbs() != NbLimbU128 {
		utils.Panic("only 128-bit are supported")
	}
	for i := 0; i < l.Size(); i++ {
		r := l.GetRowAsBytes(run, i)
		res = append(res, [16]byte(r))
	}
	return res
}

// GetAssignmentAsBigInt returns the represented big.Int for the provided field
// elements.
func (l Limbs[E]) GetAssignmentAsBigInt(run ifaces.Runtime) []*big.Int {
	res := make([]*big.Int, 0, l.Size())
	for i := 0; i < l.Size(); i++ {
		res = append(res, l.GetRowAsBigInt(run, i))
	}
	return res
}

// AssignAndPadRows assign the provided vector of rows to the limbs
func (l Limbs[E]) AssignAndPadRows(run *wizard.ProverRuntime, rows []row[E]) {
	res := make([][]field.Element, l.NumLimbs())
	for i := range res {
		res[i] = make([]field.Element, len(rows))
		for k := range rows {
			res[i][k] = rows[k].T[i]
		}
		run.AssignColumn(l.c[i].GetColID(), smartvectors.RightZeroPadded(res[i], l.NumRow()))
	}
}

// AssignBytes assigns the provided bytes to the provided field elements.
func (l Limbs[E]) AssignBytes(run *wizard.ProverRuntime, bytes [][]byte) {

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
func (l Limbs[E]) AssignBigInts(run *wizard.ProverRuntime, bigints []*big.Int) {

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

// AssignAndZeroPadsBigInts assigns the provided big.Ints to the provided field
// elements and zero pads the rest of the limbs.
func (l Limbs[E]) AssignAndZeroPadsBigInts(run *wizard.ProverRuntime, bigints []*big.Int) {
	var (
		numRow      = len(bigints)
		numLimbs    = len(l.c)
		uintBitSize = numLimbs * limbBitWidth
	)

	if l.c[0].Size() < numRow {
		utils.Panic("number of bytes must be equal to the number of limbs, got %v and %v", numRow, len(l.c))
	}

	res := bigIntToLimbsVec[E](bigints, len(l.c), uintBitSize)

	for c := range l.c {
		run.AssignColumn(l.c[c].GetColID(), smartvectors.RightZeroPadded(res[c], l.NumRow()))
	}
}

// ToBigEndianLimbs returns the limbs in big endian form
func (l Limbs[E]) ToBigEndianLimbs() Limbs[BigEndian] {
	new := Limbs[BigEndian]{name: l.name, c: make([]ifaces.Column, len(l.c))}
	copy(new.c, l.c)
	if isLittleEndian[E]() {
		slices.Reverse(new.c)
	}
	return new
}

// ToLittleEndianLimbs returns the limbs in little endian form
func (l Limbs[E]) ToLittleEndianLimbs() Limbs[LittleEndian] {
	new := Limbs[LittleEndian]{name: l.name, c: make([]ifaces.Column, len(l.c))}
	copy(new.c, l.c)
	if isBigEndian[E]() {
		slices.Reverse(new.c)
	}
	return new
}

// FuseLimbs fuses two limbs into a single limbs. The returned limbs are in the
// named with the name of hi.
func FuseLimbs[E Endianness](hi, lo Limbs[E]) Limbs[E] {
	res := make([]ifaces.Column, len(hi.c)+len(lo.c))

	if isLittleEndian[E]() {
		copy(res[:len(lo.c)], lo.c)
		copy(res[len(lo.c):], hi.c)
		return Limbs[E]{name: hi.name, c: res}
	}

	copy(res[:len(hi.c)], hi.c)
	copy(res[len(hi.c):], lo.c)
	return Limbs[E]{name: hi.name, c: res}
}

// SplitOnByte splits the limbs into two limbs. The returned limbs are in the named
// with the name of hi. The returned limbs represent the bytes b[:at] and b[at:]
// where b are the bytes represented by l. The name of the returned limbs are
// the one of the original one appended with hi and lo
func (l Limbs[E]) SplitOnBit(at int) (hi, lo Limbs[E]) {

	splitOnLimbs := utils.DivExact(at, limbBitWidth)
	if isLittleEndian[E]() {
		n := l.NumLimbs() - splitOnLimbs
		hi = Limbs[E]{name: l.name, c: l.c[n:]}
		lo = Limbs[E]{name: l.name, c: l.c[:n]}
		return hi, lo
	}

	hi = Limbs[E]{name: l.name, c: l.c[:splitOnLimbs]}
	lo = Limbs[E]{name: l.name, c: l.c[splitOnLimbs:]}
	return hi, lo
}

// SliceOnBit returns the slice representing lbits[s0:s1]
func (l Limbs[E]) SliceOnBit(s0, s1 int) Limbs[E] {
	la, _ := l.SplitOnBit(s1)  // l[:s1]
	_, lb := la.SplitOnBit(s0) // l[s0:s1]
	return lb
}

// Rename renames the limbs. But not the underlying columns.
func (l Limbs[E]) Rename(name ifaces.ColID) Limbs[E] {
	new := l
	new.name = name
	return new
}

// String implements the [github.com/consensys/linea-monorepo/prover/symbolic.Metadata]
// interface.
func (l Limbs[E]) String() string {
	return string(l.name)
}

// ColumnNames returns the names of the columns in the limbs as a list of strings
func (l Limbs[E]) ColumnNames() []string {
	res := make([]string, len(l.c))
	for i := range l.c {
		res[i] = string(l.c[i].GetColID())
	}
	return res
}

// IsBase always return true. It is implemented so that we can use it in global
// constraints
func (l Limbs[E]) IsBase() bool {
	return true
}
