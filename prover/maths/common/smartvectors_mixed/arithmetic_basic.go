package smartvectors_mixed

import (
	sv "github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectorsext"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
)

func IsBase(vec sv.SmartVector) bool {
	_, isBaseError := vec.GetBase(0)
	if isBaseError == nil {
		return true
	} else {
		return false
	}
}

func LiftToExt(vec sv.SmartVector) sv.SmartVector {
	if !IsBase(vec) {
		panic("cannot lift to extension, vector is not base")
	}
	switch v := vec.(type) {
	case *sv.Constant:
		res := smartvectorsext.NewConstantExt(
			fext.NewFromBase(v.Val()),
			v.Len(),
		)
		return res
	case *sv.PaddedCircularWindow:
		windowExt := make([]fext.Element, len(v.Window()))
		for i := range v.Window() {
			windowExt[i].SetFromBase(&v.Window()[i])
		}
		res := smartvectorsext.NewPaddedCircularWindowExt(
			windowExt,
			fext.NewFromBase(v.PaddingVal()),
			v.Offset(),
			v.Len(),
		)
		return res
	case *sv.Rotated:
		/*		for i := range v.Window() {
					windowExt[i].SetFromBase(&v.Window()[i])
				}
				return smartvectorsext.NewRotatedExt(
					,
					v.offset,
			)*/
		vecExt := make([]fext.Element, v.Len())
		v.WriteInSliceExt(vecExt)
		return smartvectorsext.NewRegularExt(vecExt)
	case *sv.Pooled:
		vecExt := make([]fext.Element, v.Len())
		v.WriteInSliceExt(vecExt)
		return smartvectorsext.NewRegularExt(vecExt)
	case *sv.Regular:
		vecExt := make([]fext.Element, v.Len())
		v.WriteInSliceExt(vecExt)
		return smartvectorsext.NewRegularExt(vecExt)
	}
	panic("unsupported type")
}

// Add returns a smart-vector obtained by position-wise adding [SmartVector].
//   - all inputs `vecs` must have the same size, or the function panics
//   - the output smart-vector has the same size as the input vectors
//   - if no input vectors are provided, the function panics
func Add(vecs ...sv.SmartVector) sv.SmartVector {
	vecsBase := make([]sv.SmartVector, 0, len(vecs))
	vecsExt := make([]sv.SmartVector, 0, len(vecs))
	for _, vec := range vecs {
		if IsBase(vec) {
			vecsBase = append(vecsBase, vec)
		} else {
			vecsExt = append(vecsExt, vec)
		}
	}

	var res sv.SmartVector = sv.NewConstant(field.Zero(), vecs[0].Len())
	if len(vecsBase) > 0 {
		res = sv.Add(vecsBase...)
	}

	if len(vecsExt) == 0 {
		return res
	} else {
		addExt := smartvectorsext.Add(vecsExt...)
		liftedBase := LiftToExt(res)
		return smartvectorsext.Add(liftedBase, addExt)
	}

}

// Mul returns a smart-vector obtained by position-wise multiplying [SmartVector].
//   - all inputs `vecs` must have the same size, or the function panics
//   - the output smart-vector has the same size as the input vectors
//   - if no input vectors are provided, the function panics
func Mul(vecs ...SmartVector) SmartVector {
	coeffs := make([]int, len(vecs))
	for i := range coeffs {
		coeffs[i] = 1
	}

	return Product(coeffs, vecs)
}
