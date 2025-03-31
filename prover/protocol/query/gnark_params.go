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
	baseYs []frontend.Variable
	extYs  []gnarkfext.Variable
	isBase bool
}

func (p InnerProduct) GnarkAllocate() GnarkInnerProductParams {
	if p.isBase {
		return GnarkInnerProductParams{
			baseYs: make([]frontend.Variable, len(p.Bs)),
			extYs:  nil,
			isBase: true,
		}
	} else {
		return GnarkInnerProductParams{
			baseYs: nil,
			extYs:  make([]gnarkfext.Variable, len(p.Bs)),
			isBase: false,
		}
	}

}

func (p InnerProductParams) GnarkAssign() GnarkInnerProductParams {
	return GnarkInnerProductParams{
		baseYs: vector.IntoGnarkAssignment(p.baseYs),
		extYs:  vectorext.IntoGnarkAssignment(p.extYs),
		isBase: p.isBase,
	}
}

// A gnark circuit version of univariate eval params
type GnarkUnivariateEvalParams struct {
	baseX  frontend.Variable
	baseYs []frontend.Variable
	extX   gnarkfext.Variable
	extYs  []gnarkfext.Variable
	isBase bool
}

func (p UnivariateEval) GnarkAllocate() GnarkUnivariateEvalParams {
	// no need to preallocate the x because its size is already known
	if p.Pols[0].IsBase() {
		return GnarkUnivariateEvalParams{
			baseYs: make([]frontend.Variable, len(p.Pols)),
			extYs:  nil,
			isBase: true,
		}
	} else {
		return GnarkUnivariateEvalParams{
			baseYs: nil,
			extYs:  make([]gnarkfext.Variable, len(p.Pols)),
			isBase: false,
		}
	}
}

// Returns a gnark assignment for the present parameters
func (p UnivariateEvalParams) GnarkAssign() GnarkUnivariateEvalParams {
	if p.isBase {
		return GnarkUnivariateEvalParams{
			baseYs: vector.IntoGnarkAssignment(p.baseYs),
			baseX:  p.baseX,
			extYs:  nil,
			extX:   gnarkfext.NewZero(),
			isBase: true,
		}
	} else {
		return GnarkUnivariateEvalParams{
			baseYs: nil,
			baseX:  nil,
			extYs:  vectorext.IntoGnarkAssignment(p.extYs),
			extX:   gnarkfext.ExtToVariable(p.extX),
			isBase: false,
		}
	}
}

// Update the fiat-shamir state with the the present parameters
func (p GnarkInnerProductParams) UpdateFS(fs *fiatshamir.GnarkFiatShamir) {
	if p.isBase {
		fs.Update(p.baseYs...)
	} else {
		fs.UpdateExt(p.extYs...)
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
	if p.isBase {
		fs.Update(p.baseYs...)
	} else {
		fs.UpdateExt(p.extYs...)
	}
}
