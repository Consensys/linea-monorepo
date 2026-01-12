package query

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/crypto/fiatshamir"
	"github.com/consensys/linea-monorepo/prover/maths/common/vectorext"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/gnarkfext"
	"github.com/consensys/linea-monorepo/prover/maths/zk"
)

// A gnark circuit version of the LocalOpeningResult
type GnarkLocalOpeningParams struct {
	BaseY  frontend.Variable
	ExtY   gnarkfext.Element
	IsBase bool
}

func (p LocalOpeningParams) GnarkAssign() GnarkLocalOpeningParams {

	ExtY := gnarkfext.AssignFromExt(p.ExtY)
	return GnarkLocalOpeningParams{
		BaseY:  field.NewFromKoala(p.BaseY),
		ExtY:   ExtY,
		IsBase: p.IsBase,
	}
}

// A gnark circuit version of LogDerivSumParams
type GnarkLogDerivSumParams struct {
	Sum gnarkfext.Element
}

// A gnark circuit version of GrandProductParams
type GnarkGrandProductParams struct {
	Prod gnarkfext.Element
}

// HornerParamsPartGnark is a [HornerParamsPart] in a gnark circuit.
type HornerParamsPartGnark struct {
	// N0 is an initial offset of the Horner query
	N0 frontend.Variable
	// N1 is the second offset of the Horner query
	N1 frontend.Variable
}

// GnarkHornerParams represents the parameters of the Horner evaluation query
// in a gnark circuit.
type GnarkHornerParams struct {
	// Final result is the result of summing the Horner parts for every
	// queries.
	FinalResult gnarkfext.Element
	// Parts are the parameters of the Horner parts
	Parts []HornerParamsPartGnark
}

func (p LogDerivSumParams) GnarkAssign() GnarkLogDerivSumParams {
	// return GnarkLogDerivSumParams{Sum: p.Sum}
	tmp := p.Sum.GetExt()
	return GnarkLogDerivSumParams{Sum: gnarkfext.AssignFromExt(tmp)} // TODO @thomas fixme (ext vs base)
}

// A gnark circuit version of InnerProductParams
type GnarkInnerProductParams struct {
	Ys []gnarkfext.Element
}

func (p InnerProduct) GnarkAllocate() GnarkInnerProductParams {
	return GnarkInnerProductParams{Ys: make([]gnarkfext.Element, len(p.Bs))}
}

func (p InnerProductParams) GnarkAssign() GnarkInnerProductParams {
	return GnarkInnerProductParams{Ys: vectorext.IntoGnarkAssignment(p.Ys)}
}

// A gnark circuit version of univariate eval params
type GnarkUnivariateEvalParams struct {
	ExtX  gnarkfext.Element
	ExtYs []gnarkfext.Element
}

func (p UnivariateEval) GnarkAllocate() GnarkUnivariateEvalParams {
	// no need to preallocate the x because its size is already known
	return GnarkUnivariateEvalParams{
		ExtYs: make([]gnarkfext.Element, len(p.Pols)),
	}
}

// Returns a gnark assignment for the present parameters
func (p UnivariateEvalParams) GnarkAssign() GnarkUnivariateEvalParams {
	return GnarkUnivariateEvalParams{
		ExtYs: vectorext.IntoGnarkAssignment(p.ExtYs),
		ExtX:  gnarkfext.AssignFromExt(p.ExtX),
	}

}

// GnarkAllocate allocates a [GnarkHornerParams] with the right dimensions
func (p HornerParams) GnarkAllocate() GnarkHornerParams {
	return GnarkHornerParams{
		Parts: make([]HornerParamsPartGnark, len(p.Parts)),
	}
}

// GnarkAssign returns a gnark assignment for the present parameters.
func (p HornerParams) GnarkAssign() GnarkHornerParams {

	parts := make([]HornerParamsPartGnark, len(p.Parts))
	for i, part := range p.Parts {
		parts[i] = HornerParamsPartGnark{
			N0: zk.ValueOf(part.N0),
			N1: zk.ValueOf(part.N1),
		}
	}

	return GnarkHornerParams{
		FinalResult: gnarkfext.AssignFromExt(p.FinalResult),
		Parts:       parts,
	}
}

// Update the fiat-shamir state with the the present parameters
func (p GnarkInnerProductParams) UpdateFS(fs *fiatshamir.GnarkFS) {
	(*fs).UpdateExt(p.Ys...)
}

// Update the fiat-shamir state with the the present parameters
func (p GnarkLocalOpeningParams) UpdateFS(fs *fiatshamir.GnarkFS) {
	if p.IsBase {
		(*fs).Update(zk.WrapFrontendVariable(p.BaseY))
	} else {
		(*fs).UpdateExt(p.ExtY)
	}
}

// Update the fiat-shamir state with the the present parameters
func (p GnarkLogDerivSumParams) UpdateFS(fs *fiatshamir.GnarkFS) {

	(*fs).UpdateExt(p.Sum)
}

// Update the fiat-shamir state with the the present parameters
func (p GnarkGrandProductParams) UpdateFS(fs *fiatshamir.GnarkFS) {
	(*fs).UpdateExt(p.Prod)
}

// Update the fiat-shamir state with the the present parameters
func (p GnarkUnivariateEvalParams) UpdateFS(fs *fiatshamir.GnarkFS) {

	(*fs).UpdateExt(p.ExtYs...)
}

// Update the fiat-shamir state with the present parameters
func (p GnarkHornerParams) UpdateFS(fs *fiatshamir.GnarkFS) {
	(*fs).UpdateExt(p.FinalResult)

	for _, part := range p.Parts {
		(*fs).Update(zk.WrapFrontendVariable(part.N0), zk.WrapFrontendVariable(part.N1))
	}
}
