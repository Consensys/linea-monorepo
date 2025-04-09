package query

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/crypto/fiatshamir"
	"github.com/consensys/linea-monorepo/prover/maths/common/vector"
	"github.com/consensys/linea-monorepo/prover/maths/common/vectorext"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext/gnarkfext"
)

// A gnark circuit version of the LocalOpeningResult
type GnarkLocalOpeningParams struct {
	BaseY  frontend.Variable
	ExtY   gnarkfext.Variable
	isBase bool
}

func (p LocalOpeningParams) GnarkAssign() GnarkLocalOpeningParams {
	return GnarkLocalOpeningParams{
		BaseY: p.BaseY,
		ExtY: gnarkfext.Variable{
			A0: p.ExtY.A0,
			A1: p.ExtY.A1,
		},
		isBase: p.IsBase,
	}
}

// A gnark circuit version of InnerProductParams
type GnarkInnerProductParams struct {
	BaseYs []frontend.Variable
	ExtYs  []gnarkfext.Variable
	IsBase bool
}

func (p InnerProduct) GnarkAllocate() GnarkInnerProductParams {
	if p.isBase {
		return GnarkInnerProductParams{
			BaseYs: make([]frontend.Variable, len(p.Bs)),
			ExtYs:  nil,
			IsBase: true,
		}
	} else {
		return GnarkInnerProductParams{
			BaseYs: nil,
			ExtYs:  make([]gnarkfext.Variable, len(p.Bs)),
			IsBase: false,
		}
	}

}

func (p InnerProductParams) GnarkAssign() GnarkInnerProductParams {
	return GnarkInnerProductParams{
		BaseYs: vector.IntoGnarkAssignment(p.baseYs),
		ExtYs:  vectorext.IntoGnarkAssignment(p.extYs),
		IsBase: p.isBase,
	}
}

// A gnark circuit version of univariate eval params
type GnarkUnivariateEvalParams struct {
	BaseX frontend.Variable
	ExtX  gnarkfext.Variable
	// isBaseX is true if the x value is in the base field, false otherwise
	IsBaseX bool
	BaseYs  []frontend.Variable
	ExtYs   []gnarkfext.Variable
}

func (p UnivariateEval) GnarkAllocate() GnarkUnivariateEvalParams {
	// no need to preallocate the x because its size is already known
	if p.Pols[0].IsBase() {
		return GnarkUnivariateEvalParams{
			BaseYs: make([]frontend.Variable, len(p.Pols)),
			ExtYs:  nil,
			IsBase: true,
		}
	} else {
		return GnarkUnivariateEvalParams{
			BaseYs: nil,
			ExtYs:  make([]gnarkfext.Variable, len(p.Pols)),
			IsBase: false,
		}
	}
}

// Returns a gnark assignment for the present parameters
func (p UnivariateEvalParams) GnarkAssign() GnarkUnivariateEvalParams {
	if p.IsBase {
		return GnarkUnivariateEvalParams{
			BaseYs: vector.IntoGnarkAssignment(p.BaseYs),
			BaseX:  p.BaseX,
			ExtYs:  nil,
			ExtX:   gnarkfext.NewZero(),
			IsBase: true,
		}
	} else {
		return GnarkUnivariateEvalParams{
			BaseYs: nil,
			BaseX:  nil,
			ExtYs:  vectorext.IntoGnarkAssignment(p.ExtYs),
			ExtX:   gnarkfext.ExtToVariable(p.ExtX),
			IsBase: false,
		}
	}
}

// Update the fiat-shamir state with the the present parameters
func (p GnarkInnerProductParams) UpdateFS(fs *fiatshamir.GnarkFiatShamir) {
	if p.IsBase {
		fs.Update(p.BaseYs...)
	} else {
		fs.UpdateExt(p.ExtYs...)
	}
}

// Update the fiat-shamir state with the the present parameters
func (p GnarkLocalOpeningParams) UpdateFS(fs *fiatshamir.GnarkFiatShamir) {
	if p.isBase {
		fs.Update(p.BaseY)
	} else {
		fs.UpdateExt(p.ExtY)
	}
}

// Update the fiat-shamir state with the the present parameters
func (p GnarkUnivariateEvalParams) UpdateFS(fs *fiatshamir.GnarkFiatShamir) {
	if p.IsBase {
		fs.Update(p.BaseYs...)
	} else {
		fs.UpdateExt(p.ExtYs...)
	}
}
