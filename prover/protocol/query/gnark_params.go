package query

import (
	fiatshamir "github.com/consensys/linea-monorepo/prover/crypto/fiatshamir_koalabear"
	"github.com/consensys/linea-monorepo/prover/maths/common/vectorext"
	"github.com/consensys/linea-monorepo/prover/maths/field/gnarkfext"
	"github.com/consensys/linea-monorepo/prover/maths/zk"
)

// A gnark circuit version of the LocalOpeningResult
type GnarkLocalOpeningParams struct {
	BaseY  zk.WrappedVariable
	ExtY   gnarkfext.E4Gen
	IsBase bool
}

func (p LocalOpeningParams) GnarkAssign() GnarkLocalOpeningParams {

	exty := gnarkfext.NewE4Gen(p.ExtY)
	return GnarkLocalOpeningParams{
		BaseY:  zk.ValueOf(p.BaseY.String()),
		ExtY:   exty,
		IsBase: p.IsBase,
	}
}

// A gnark circuit version of LogDerivSumParams
type GnarkLogDerivSumParams struct {
	Sum gnarkfext.E4Gen
}

// A gnark circuit version of GrandProductParams
type GnarkGrandProductParams struct {
	Prod gnarkfext.E4Gen
}

// HornerParamsPartGnark is a [HornerParamsPart] in a gnark circuit.
type HornerParamsPartGnark struct {
	// N0 is an initial offset of the Horner query
	N0 zk.WrappedVariable
	// N1 is the second offset of the Horner query
	N1 zk.WrappedVariable
}

// GnarkHornerParams represents the parameters of the Horner evaluation query
// in a gnark circuit.
type GnarkHornerParams struct {
	// Final result is the result of summing the Horner parts for every
	// queries.
	FinalResult gnarkfext.E4Gen
	// Parts are the parameters of the Horner parts
	Parts []HornerParamsPartGnark
}

func (p LogDerivSumParams) GnarkAssign() GnarkLogDerivSumParams {
	// return GnarkLogDerivSumParams{Sum: p.Sum}
	tmp := p.Sum.GetExt()
	return GnarkLogDerivSumParams{Sum: gnarkfext.NewE4Gen(tmp)} // TODO @thomas fixme (ext vs base)
}

// A gnark circuit version of InnerProductParams
type GnarkInnerProductParams struct {
	Ys []gnarkfext.E4Gen
}

func (p InnerProduct) GnarkAllocate() GnarkInnerProductParams {
	return GnarkInnerProductParams{Ys: make([]gnarkfext.E4Gen, len(p.Bs))}
}

func (p InnerProductParams) GnarkAssign() GnarkInnerProductParams {
	return GnarkInnerProductParams{Ys: vectorext.IntoGnarkAssignment(p.Ys)}
}

// A gnark circuit version of univariate eval params
type GnarkUnivariateEvalParams struct {
	ExtX  gnarkfext.E4Gen
	ExtYs []gnarkfext.E4Gen
}

func (p UnivariateEval) GnarkAllocate() GnarkUnivariateEvalParams {
	// no need to preallocate the x because its size is already known
	return GnarkUnivariateEvalParams{
		ExtYs: make([]gnarkfext.E4Gen, len(p.Pols)),
	}
}

// Returns a gnark assignment for the present parameters
func (p UnivariateEvalParams) GnarkAssign() GnarkUnivariateEvalParams {
	return GnarkUnivariateEvalParams{
		ExtYs: vectorext.IntoGnarkAssignment(p.ExtYs),
		ExtX:  gnarkfext.NewE4Gen(p.ExtX),
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
		FinalResult: gnarkfext.NewE4Gen(p.FinalResult),
		Parts:       parts,
	}
}

// Update the fiat-shamir state with the the present parameters
func (p GnarkInnerProductParams) UpdateFS(fs *fiatshamir.GnarkFS) {
	fs.UpdateExt(p.Ys...)
}

// Update the fiat-shamir state with the the present parameters
func (p GnarkLocalOpeningParams) UpdateFS(fs *fiatshamir.GnarkFS) {
	fs.Update(p.BaseY.AsNative())
}

// Update the fiat-shamir state with the the present parameters
func (p GnarkLogDerivSumParams) UpdateFS(fs *fiatshamir.GnarkFS) {
	fs.UpdateExt(p.Sum)
}

// Update the fiat-shamir state with the the present parameters
func (p GnarkGrandProductParams) UpdateFS(fs *fiatshamir.GnarkFS) {
	fs.UpdateExt(p.Prod)
}

// Update the fiat-shamir state with the the present parameters
func (p GnarkUnivariateEvalParams) UpdateFS(fs *fiatshamir.GnarkFS) {
	fs.UpdateExt(p.ExtYs...)
}

// Update the fiat-shamir state with the the present field extension parameters
func (p GnarkUnivariateEvalParams) UpdateFSExt(fs *fiatshamir.GnarkFS) {
	fs.UpdateExt(p.ExtYs...)
}

// Update the fiat-shamir state with the the present parameters
func (p GnarkHornerParams) UpdateFS(fs *fiatshamir.GnarkFS) {
	fs.UpdateExt(p.FinalResult)

	for _, part := range p.Parts {
		fs.Update(part.N0.AsNative(), part.N1.AsNative())
	}
}
