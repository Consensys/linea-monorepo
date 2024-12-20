package query

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/crypto/fiatshamir"
	"github.com/consensys/linea-monorepo/prover/maths/common/vector"
)

// A gnark circuit version of the LocalOpeningResult
type GnarkLocalOpeningParams struct {
	Y frontend.Variable
}

func (p LocalOpeningParams) GnarkAssign() GnarkLocalOpeningParams {
	return GnarkLocalOpeningParams{Y: p.Y}
}

type GnarkLogDerivSumParams struct {
	Y frontend.Variable
}

func (p LogDerivSumParams) GnarkAssign() GnarkLogDerivSumParams {
	return GnarkLogDerivSumParams{Y: p.Sum}
}

// A gnark circuit version of InnerProductParams
type GnarkInnerProductParams struct {
	Ys []frontend.Variable
}

func (p InnerProduct) GnarkAllocate() GnarkInnerProductParams {
	return GnarkInnerProductParams{Ys: make([]frontend.Variable, len(p.Bs))}
}

func (p InnerProductParams) GnarkAssign() GnarkInnerProductParams {
	return GnarkInnerProductParams{Ys: vector.IntoGnarkAssignment(p.Ys)}
}

// A gnark circuit version of univariate eval params
type GnarkUnivariateEvalParams struct {
	X  frontend.Variable
	Ys []frontend.Variable
}

func (p UnivariateEval) GnarkAllocate() GnarkUnivariateEvalParams {
	// no need to preallocate the x because its size is already known
	return GnarkUnivariateEvalParams{Ys: make([]frontend.Variable, len(p.Pols))}
}

// Returns a gnark assignment for the present parameters
func (p UnivariateEvalParams) GnarkAssign() GnarkUnivariateEvalParams {
	return GnarkUnivariateEvalParams{Ys: vector.IntoGnarkAssignment(p.Ys), X: p.X}
}

// Update the fiat-shamir state with the the present parameters
func (p GnarkInnerProductParams) UpdateFS(fs *fiatshamir.GnarkFiatShamir) {
	fs.Update(p.Ys...)
}

// Update the fiat-shamir state with the the present parameters
func (p GnarkLocalOpeningParams) UpdateFS(fs *fiatshamir.GnarkFiatShamir) {
	fs.Update(p.Y)
}

// Update the fiat-shamir state with the the present parameters
func (p GnarkLogDerivSumParams) UpdateFS(fs *fiatshamir.GnarkFiatShamir) {
	fs.Update(p.Y)
}

// Update the fiat-shamir state with the the present parameters
func (p GnarkUnivariateEvalParams) UpdateFS(fs *fiatshamir.GnarkFiatShamir) {
	fs.Update(p.Ys...)
}
