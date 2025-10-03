package verifiercol

import (
	"fmt"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/maths/field/gnarkfext"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/protocol/zk"
	"github.com/consensys/linea-monorepo/prover/utils"
)

// Represents a constant column
type ConstCol[T zk.Element] struct {
	Base       field.Element
	Ext        fext.Element
	IsBaseFlag bool
	Size_      int
	Name       string
}

// NewConstantCol creates a new ConstCol column
func NewConstantCol[T zk.Element](elem field.Element, size int, name string) ifaces.Column[T] {
	return ConstCol[T]{
		Base:       elem,
		Ext:        fext.Lift(elem),
		IsBaseFlag: true,
		Size_:      size,
		Name:       name,
	}
}

func NewConstantColExt[T zk.Element](elem fext.Element, size int, name string) ifaces.Column[T] {
	return ConstCol[T]{
		Base:       field.Zero(),
		Ext:        elem,
		IsBaseFlag: false,
		Size_:      size,
		Name:       name,
	}
}

// Returns the round of definition of the column (always zero)
// Even though this is more of a convention than a meaningful
// return value.
func (cc ConstCol[T]) Round() int {
	return 0
}

// Returns a generic name from the column. Defined from the coin's.
func (cc ConstCol[T]) GetColID() ifaces.ColID {

	val := cc.Base.String()
	if !cc.IsBaseFlag {
		val = cc.Ext.String()
	}

	if len(cc.Name) > 0 {
		return ifaces.ColIDf("CONSTCOL_%v_%v", val, cc.Name)
	}

	return ifaces.ColIDf("CONSTCOL_%v_%v", val, cc.Size_)
}

// Always return true
func (cc ConstCol[T]) MustExists() {}

// Returns the size of the column
func (cc ConstCol[T]) Size() int {
	return cc.Size_
}

// Returns a constant smart-vector
func (cc ConstCol[T]) GetColAssignment(_ ifaces.Runtime) ifaces.ColAssignment {
	return smartvectors.NewConstant(cc.Base, cc.Size_)
}

func (cc ConstCol[T]) GetColAssignmentAtBase(_ ifaces.Runtime, _ int) (field.Element, error) {
	if cc.IsBaseFlag {
		return cc.Base, nil
	} else {
		return field.Zero(), fmt.Errorf("requested a base element from a verifier col over field extensions")
	}
}

func (cc ConstCol[T]) GetColAssignmentAtExt(_ ifaces.Runtime, _ int) fext.Element {
	return cc.Ext
}

// Returns the column as a list of gnark constants
func (cc ConstCol[T]) GetColAssignmentGnark(_ ifaces.GnarkRuntime[T]) []T {

	res := make([]T, cc.Size_)
	for i := range res {
		res[i] = *zk.ValueOf[T](cc.Base)
	}
	return res
}

func (cc ConstCol[T]) GetColAssignmentGnarkBase(run ifaces.GnarkRuntime[T]) ([]T, error) {
	if cc.IsBaseFlag {
		res := make([]T, cc.Size_)
		for i := range res {
			res[i] = *zk.ValueOf[T](cc.Base)
		}
		return res, nil
	} else {
		return nil, fmt.Errorf("requested base elements but column defined over field extensions")
	}
}

func (cc ConstCol[T]) GetColAssignmentGnarkExt(run ifaces.GnarkRuntime[T]) []gnarkfext.E4Gen[T] {
	res := make([]gnarkfext.E4Gen[T], cc.Size_)
	for i := range res {
		var temp gnarkfext.E4Gen[T]
		temp.FromExt(cc.Ext)
		res[i] = temp
	}
	return res
}

// Returns a particular position of the coin value
func (cc ConstCol[T]) GetColAssignmentAt(run ifaces.Runtime, pos int) field.Element {
	return cc.Base
}

// Returns a particular position of the coin value
func (cc ConstCol[T]) GetColAssignmentGnarkAt(run ifaces.GnarkRuntime[T], pos int) T {
	var res T
	if cc.IsBaseFlag {
		res = *zk.ValueOf[T](cc.Base)
		return res
	} else {
		panic("requested a base element from a verifier col over field extensions")
	}
}

// Returns a particular position of the coin value
func (cc ConstCol[T]) GetColAssignmentGnarkAtBase(run ifaces.GnarkRuntime[T], pos int) (T, error) {
	var res T
	if cc.IsBaseFlag {
		res = *zk.ValueOf[T](cc.Base)
		return res, nil
	} else {
		return res, fmt.Errorf("requested a base element from a verifier col over field extensions")
	}
}

// Returns a particular position of the coin value
func (cc ConstCol[T]) GetColAssignmentGnarkAtExt(run ifaces.GnarkRuntime[T], pos int) gnarkfext.E4Gen[T] {
	var temp gnarkfext.E4Gen[T]
	temp.FromExt(cc.Ext)
	return temp
}

// Since the column is directly defined from the
// values of a random coin it does not count as a
// composite column.
func (cc ConstCol[T]) IsComposite() bool {
	return false
}

// Returns the name of the column.
func (cc ConstCol[T]) String() string {
	return string(cc.GetColID())
}

// Splits the column and return a handle of it
func (cc ConstCol[T]) Split(comp *wizard.CompiledIOP[T], from, to int) ifaces.Column[T] {

	if to < from || to-from > cc.Size() {
		utils.Panic("Can't split %++v into [%v, %v]", cc, from, to)
	}

	// Copy the underlying cc, and assigns the new from and to
	return NewConstantCol[T](cc.Base, to-from, cc.Name)
}

func (cc ConstCol[T]) IsBase() bool {
	return cc.IsBaseFlag
}

func (cc ConstCol[T]) IsZero() bool {
	if cc.IsBaseFlag {
		return cc.Base.IsZero()
	} else {
		return cc.Ext.IsZero()
	}
}

func (cc ConstCol[T]) IsOne() bool {
	if cc.IsBaseFlag {
		return cc.Base.IsOne()
	} else {
		return cc.Ext.IsOne()
	}
}

// Returns the string representation of the underlying field element
func (cc ConstCol[T]) StringField() string {
	if cc.IsBaseFlag {
		return cc.Base.String()
	} else {
		return cc.Ext.String()
	}
}
