package verifiercol

import (
	"fmt"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectorsext"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext/gnarkfext"
	"strings"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
)

// compile check to enforce the struct to belong to the corresponding interface
var _ VerifierCol = FromAccessors{}

// FromAccessors is a [VerifierCol] implementation that is built by stacking the
// values of [ifaces.Accessor] on top of each others. It is the most type of
// column and subsumes almost all the other types.
type FromAccessors struct {
	// Accessors stores the list of accessors building the column.
	Accessors  []ifaces.Accessor
	Padding    field.Element
	PaddingExt fext.Element
	Size_      int
	// Round_ caches the round value of the column.
	Round_ int
}

func (f FromAccessors) IsBase() bool {
	return f.Accessors[0].IsBase()
}

// NewFromAccessors instantiates a [FromAccessors] column from a list of
// [ifaces.Accessor]. The column's size must be a power of 2.
//
// You should not pass accessors of type [expressionAsAccessor] as their
// evaluation within a gnark circuit requires using the frontend.API which we
// can't access in the context currently.
func NewFromAccessors(accessors []ifaces.Accessor, padding field.Element, size int) ifaces.Column {
	if !utils.IsPowerOfTwo(size) {
		utils.Panic("the column must be a power of two (size=%v)", size)
	}
	round := 0
	for i := range accessors {
		round = max(round, accessors[i].Round())
	}
	return FromAccessors{
		Accessors: accessors,
		Round_:    round,
		Padding:   padding,
		Size_:     size,
	}
}

func NewFromAccessorsExt(accessors []ifaces.Accessor, padding fext.Element, size int) ifaces.Column {
	if !utils.IsPowerOfTwo(size) {
		utils.Panic("the column must be a power of two (size=%v)", size)
	}
	round := 0
	for i := range accessors {
		round = max(round, accessors[i].Round())
	}
	return FromAccessors{
		Accessors:  accessors,
		Round_:     round,
		PaddingExt: padding,
		Size_:      size,
	}
}

// Round returns the round ID of the column and implements the [ifaces.Column]
// interface.
func (f FromAccessors) Round() int {
	return f.Round_
}

// GetColID returns the column ID
func (f FromAccessors) GetColID() ifaces.ColID {
	accessorNames := make([]string, len(f.Accessors))
	for i := range f.Accessors {
		accessorNames[i] = f.Accessors[i].Name()
	}
	return ifaces.ColIDf("FROM_ACCESSORS_%v_PADDING=%v_SIZE=%v", strings.Join(accessorNames, "_"), f.Padding.String(), f.Size_)
}

// MustExists implements the [ifaces.Column] interface and always returns true.
func (f FromAccessors) MustExists() {}

// Size returns the size of the colum and implements the [ifaces.Column]
// interface.
func (f FromAccessors) Size() int {
	return f.Size_
}

// GetColAssignment returns the assignment of the current column
func (f FromAccessors) GetColAssignment(run ifaces.Runtime) ifaces.ColAssignment {
	if f.IsBase() {
		res := make([]field.Element, len(f.Accessors))
		for i := range res {
			res[i] = f.Accessors[i].GetVal(run)
		}
		return smartvectors.RightPadded(res, f.Padding, f.Size_)
	} else {
		res := make([]fext.Element, len(f.Accessors))
		for i := range res {
			res[i] = f.Accessors[i].GetValExt(run)
		}
		return smartvectorsext.RightPadded(res, f.PaddingExt, f.Size_)
	}

}

// GetColAssignment returns a gnark assignment of the current column
func (f FromAccessors) GetColAssignmentGnark(run ifaces.GnarkRuntime) []frontend.Variable {

	res := make([]frontend.Variable, f.Size_)
	for i := range f.Accessors {
		res[i] = f.Accessors[i].GetFrontendVariable(nil, run)
	}

	for i := len(f.Accessors); i < f.Size_; i++ {
		res[i] = f.Padding
	}

	return res
}

func (f FromAccessors) GetColAssignmentGnarkBase(run ifaces.GnarkRuntime) ([]frontend.Variable, error) {
	if f.IsBase() {
		res := make([]frontend.Variable, f.Size_)
		for i := range f.Accessors {
			res[i] = f.Accessors[i].GetFrontendVariable(nil, run)
		}

		for i := len(f.Accessors); i < f.Size_; i++ {
			res[i] = f.Padding
		}

		return res, nil
	} else {
		return nil, fmt.Errorf("requested a base element from a field extension")
	}
}

func (f FromAccessors) GetColAssignmentGnarkExt(run ifaces.GnarkRuntime) []gnarkfext.Variable {
	res := make([]gnarkfext.Variable, f.Size_)
	for i := range f.Accessors {
		res[i] = f.Accessors[i].GetFrontendVariableExt(nil, run)
	}

	for i := len(f.Accessors); i < f.Size_; i++ {
		res[i] = gnarkfext.ExtToVariable(f.PaddingExt)
	}

	return res
}

// GetColAssignmentAt returns a particular position of the column
func (f FromAccessors) GetColAssignmentAt(run ifaces.Runtime, pos int) field.Element {

	if pos >= f.Size_ {
		utils.Panic("out of bound: size=%v pos=%v", f.Size_, pos)
	}

	if pos >= len(f.Accessors) {
		return f.Padding
	}

	return f.Accessors[pos].GetVal(run)
}

func (f FromAccessors) GetColAssignmentAtBase(run ifaces.Runtime, pos int) (field.Element, error) {
	if f.IsBase() {
		if pos >= f.Size_ {
			utils.Panic("out of bound: size=%v pos=%v", f.Size_, pos)
		}

		if pos >= len(f.Accessors) {
			return f.Padding, nil
		}

		return f.Accessors[pos].GetVal(run), nil
	} else {
		return field.Zero(), fmt.Errorf("requested a base element from an underlying field extension")
	}
}

func (f FromAccessors) GetColAssignmentAtExt(run ifaces.Runtime, pos int) fext.Element {
	if f.IsBase() {
		baseElem, _ := f.GetColAssignmentAtBase(run, pos)
		return fext.NewFromBase(baseElem)
	} else {
		if pos >= f.Size_ {
			utils.Panic("out of bound: size=%v pos=%v", f.Size_, pos)
		}

		if pos >= len(f.Accessors) {
			return f.PaddingExt
		}

		return f.Accessors[pos].GetValExt(run)

	}
}

// GetColAssignmentGnarkAt returns a particular position of the column in a gnark circuit
func (f FromAccessors) GetColAssignmentGnarkAt(run ifaces.GnarkRuntime, pos int) frontend.Variable {
	if pos >= f.Size_ {
		utils.Panic("out of bound: size=%v pos=%v", f.Size_, pos)
	}

	if pos >= len(f.Accessors) {
		return f.Padding
	}

	return f.Accessors[pos].GetFrontendVariable(nil, run)
}

func (f FromAccessors) GetColAssignmentGnarkAtBase(run ifaces.GnarkRuntime, pos int) (frontend.Variable, error) {
	if f.IsBase() {
		if pos >= f.Size_ {
			utils.Panic("out of bound: size=%v pos=%v", f.Size_, pos)
		}

		if pos >= len(f.Accessors) {
			return f.Padding, nil
		}

		return f.Accessors[pos].GetFrontendVariable(nil, run), nil
	} else {
		return nil, fmt.Errorf("requested a base element from an underlying field extension")
	}
}

func (f FromAccessors) GetColAssignmentGnarkAtExt(run ifaces.GnarkRuntime, pos int) gnarkfext.Variable {
	if f.IsBase() {
		if pos >= f.Size_ {
			utils.Panic("out of bound: size=%v pos=%v", f.Size_, pos)
		}

		if pos >= len(f.Accessors) {
			return gnarkfext.NewFromBase(f.Padding)
		}

		elem, _ := f.Accessors[pos].GetFrontendVariableBase(nil, run)
		return gnarkfext.NewFromBase(elem)
	} else {
		if pos >= f.Size_ {
			utils.Panic("out of bound: size=%v pos=%v", f.Size_, pos)
		}

		if pos >= len(f.Accessors) {
			return gnarkfext.ExtToVariable(f.PaddingExt)
		}

		return f.Accessors[pos].GetFrontendVariableExt(nil, run)
	}
}

// IsComposite implements the [ifaces.Column] interface
func (f FromAccessors) IsComposite() bool {
	return false
}

// String implements the [symbolic.Metadata] interface
func (f FromAccessors) String() string {
	return string(f.GetColID())
}

// Split implements the [VerifierCol] interface
func (f FromAccessors) Split(_ *wizard.CompiledIOP, from, to int) ifaces.Column {

	if from >= len(f.Accessors) {
		return NewConstantCol(f.Padding, to-from)
	}

	var subAccessors = f.Accessors[from:]

	if to < len(f.Accessors) {
		subAccessors = f.Accessors[from:to]
	}

	// We don't call the accessor to ensure that the segment has the same round
	// definition as the original column.
	return FromAccessors{
		Accessors: subAccessors,
		Round_:    f.Round_,
		Padding:   f.Padding,
		Size_:     to - from,
	}
}
