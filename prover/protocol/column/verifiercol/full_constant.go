package verifiercol

import (
	"fmt"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext/gnarkfext"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
)

// Represents a constant column
type ConstCol struct {
	Base   field.Element
	Ext    fext.Element
	isBase bool
	Size_  int
}

// NewConstantCol creates a new ConstCol column
func NewConstantCol(elem field.Element, size int) ifaces.Column {
	return ConstCol{
		Base:   elem,
		Ext:    fext.NewFromBase(elem),
		isBase: true,
		Size_:  size,
	}
}

func NewConstantColExt(elem fext.Element, size int) ifaces.Column {
	return ConstCol{
		Base:   field.Zero(),
		Ext:    elem,
		isBase: false,
		Size_:  size,
	}
}

// Returns the round of definition of the column (always zero)
// Even though this is more of a convention than a meaningful
// return value.
func (cc ConstCol) Round() int {
	return 0
}

// Returns a generic name from the column. Defined from the coin's.
func (cc ConstCol) GetColID() ifaces.ColID {
	if cc.isBase {
		return ifaces.ColIDf("CONSTCOL_%v_%v", cc.Base.String(), cc.Size_)
	} else {
		return ifaces.ColIDf("CONSTCOL_%v_%v", cc.Ext.String(), cc.Size_)
	}
}

// Always return true
func (cc ConstCol) MustExists() {}

// Returns the size of the column
func (cc ConstCol) Size() int {
	return cc.Size_
}

// Returns a constant smart-vector
func (cc ConstCol) GetColAssignment(_ ifaces.Runtime) ifaces.ColAssignment {
	return smartvectors.NewConstant(cc.Base, cc.Size_)
}

func (cc ConstCol) GetColAssignmentAtBase(_ ifaces.Runtime, _ int) (field.Element, error) {
	if cc.isBase {
		return cc.Base, nil
	} else {
		return field.Zero(), fmt.Errorf("requested a Base element from a verifier col over field extensions")
	}
}

func (cc ConstCol) GetColAssignmentAtExt(_ ifaces.Runtime, _ int) fext.Element {
	return cc.Ext
}

// Returns the column as a list of gnark constants
func (cc ConstCol) GetColAssignmentGnark(_ ifaces.GnarkRuntime) []frontend.Variable {
	res := make([]frontend.Variable, cc.Size_)
	for i := range res {
		res[i] = cc.Base
	}
	return res
}

func (cc ConstCol) GetColAssignmentGnarkBase(run ifaces.GnarkRuntime) ([]frontend.Variable, error) {
	if cc.isBase {
		res := make([]frontend.Variable, cc.Size_)
		for i := range res {
			res[i] = cc.Base
		}
		return res, nil
	} else {
		return nil, fmt.Errorf("requested Base elements but column defined over field extensions")
	}
}

func (cc ConstCol) GetColAssignmentGnarkExt(run ifaces.GnarkRuntime) []gnarkfext.Variable {
	res := make([]gnarkfext.Variable, cc.Size_)
	for i := range res {
		res[i] = gnarkfext.NewFromExtension(cc.Ext)
	}
	return res
}

// Returns a particular position of the coin value
func (cc ConstCol) GetColAssignmentAt(run ifaces.Runtime, pos int) field.Element {
	return cc.Base
}

// Returns a particular position of the coin value
func (cc ConstCol) GetColAssignmentGnarkAt(run ifaces.GnarkRuntime, pos int) frontend.Variable {
	if cc.isBase {
		return cc.Base
	} else {
		panic("requested a Base element from a verifier col over field extensions")
	}
}

// Returns a particular position of the coin value
func (cc ConstCol) GetColAssignmentGnarkAtBase(run ifaces.GnarkRuntime, pos int) (frontend.Variable, error) {
	if cc.isBase {
		return cc.Base, nil
	} else {
		return field.Zero(), fmt.Errorf("requested a Base element from a verifier col over field extensions")
	}
}

// Returns a particular position of the coin value
func (cc ConstCol) GetColAssignmentGnarkAtExt(run ifaces.GnarkRuntime, pos int) gnarkfext.Variable {
	return gnarkfext.ExtToVariable(cc.Ext)
}

// Since the column is directly defined from the
// values of a random coin it does not count as a
// composite column.
func (cc ConstCol) IsComposite() bool {
	return false
}

// Returns the name of the column.
func (cc ConstCol) String() string {
	return string(cc.GetColID())
}

// Splits the column and return a handle of it
func (cc ConstCol) Split(comp *wizard.CompiledIOP, from, to int) ifaces.Column {

	if to < from || to-from > cc.Size() {
		utils.Panic("Can't split %++v into [%v, %v]", cc, from, to)
	}

	// Copy the underlying cc, and assigns the new from and to
	return NewConstantCol(cc.Base, to-from)
}

func (cc ConstCol) IsBase() bool {
	if cc.isBase {
		return true
	} else {
		return false
	}
}
