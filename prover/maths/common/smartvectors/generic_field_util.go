package smartvectors

import (
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
)

func ToConstantSmartvector(e *fext.GenericFieldElem, length int) SmartVector {
	if e.IsBase() {
		baseElem, _ := e.GetBase()
		return NewConstant(baseElem, length)
	} else {
		return NewConstantExt(e.GetExt(), length)
	}
}

func GetFirstElemOfSmartvector(vector SmartVector) *fext.GenericFieldElem {
	if IsBase(vector) {
		elem, _ := vector.GetBase(0)
		return fext.NewESHashFromBase(&elem)
	}
	// If the vector is not over base elements, we assume it is over extensions
	elem := vector.GetExt(0)
	return fext.NewESHashFromExt(&elem)
}
