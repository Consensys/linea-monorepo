package statesummary

import (
	"github.com/consensys/zkevm-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/zkevm-monorepo/prover/maths/field"
	"github.com/consensys/zkevm-monorepo/prover/protocol/ifaces"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
	"github.com/consensys/zkevm-monorepo/prover/utils"
	"github.com/consensys/zkevm-monorepo/prover/utils/types"
)

// vectorBuilder is a convenience structure to assign columns by appending
// field elements on top of field elements.
type vectorBuilder struct {

	// column is the column being assigned by the current builder.
	column ifaces.Column

	// slice is the current stage of the field element
	slice []field.Element
}

// newVectorBuilder initializes a new vectorBuilder with a new column.
func newVectorBuilder(col ifaces.Column) *vectorBuilder {
	return &vectorBuilder{
		column: col,
		slice:  make([]field.Element, 0, col.Size()),
	}
}

// Resize resizes the slice up to `newLen`. Panics if the newLen is smaller
// than the current length.
func (vb *vectorBuilder) Resize(newLen int) {
	if len(vb.slice) > newLen {
		utils.Panic("the old length %v is larger than the newLen %v", len(vb.slice), newLen)
	}

	vb.slice = append(vb.slice, make([]field.Element, newLen-len(vb.slice))...)
}

func (vb *vectorBuilder) PushBoolean(bo bool) {
	if bo {
		vb.PushOne()
	} else {
		vb.PushZero()
	}
}

func (vb *vectorBuilder) PushZero() {
	vb.slice = append(vb.slice, field.Zero())
}

func (vb *vectorBuilder) PushOne() {
	vb.slice = append(vb.slice, field.One())
}

func (vb *vectorBuilder) PushField(f field.Element) {
	vb.slice = append(vb.slice, f)
}

func (vb *vectorBuilder) PushInt(x int) {
	f := field.NewElement(uint64(x))
	vb.PushField(f)
}

func (vb *vectorBuilder) PushHi(fb types.FullBytes32) {
	var f field.Element
	f.SetBytes(fb[:16])
	vb.PushField(f)
}

func (vb *vectorBuilder) PushLo(fb types.FullBytes32) {
	var f field.Element
	f.SetBytes(fb[16:])
	vb.PushField(f)
}

func (vb *vectorBuilder) PushBytes32(b32 types.Bytes32) {
	var f field.Element
	if err := f.SetBytesCanonical(b32[:]); err != nil {
		panic(err)
	}
	vb.PushField(f)
}

func (vb *vectorBuilder) Pop() {
	vb.slice = vb.slice[:len(vb.slice)-1]
}

func (vb *vectorBuilder) RepushLast() {
	last := vb.slice[len(vb.slice)-1]
	vb.PushField(last)
}

func (vb *vectorBuilder) PushAddr(addr types.EthAddress) {
	var f field.Element
	f.SetBytes(addr[:])
	vb.PushField(f)
}

func (vb *vectorBuilder) PadAndAssign(run *wizard.ProverRuntime, v ...field.Element) {
	paddingValue := field.Zero()
	if len(v) > 0 {
		paddingValue = v[0]
	}

	run.AssignColumn(
		vb.column.GetColID(),
		smartvectors.RightPadded(vb.slice, paddingValue, vb.column.Size()),
	)
}
