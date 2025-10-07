package query

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/crypto/fiatshamir"
	"github.com/consensys/linea-monorepo/prover/maths/common/vector"
	"github.com/consensys/linea-monorepo/prover/maths/common/vectorext"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/gnarkfext"
)

// A gnark circuit version of the LocalOpeningResult
type GnarkLocalOpeningParams struct {
	BaseY  zk.WrappedVariable
	ExtY   gnarkfext.E4Gen
	IsBase bool
}

func (p LocalOpeningParams) GnarkAssign() GnarkLocalOpeningParams {
	var exty gnarkfext.E4Gen
	exty.Assign(p.ExtY)
	return GnarkLocalOpeningParams{
		BaseY:  p.BaseY,
		ExtY:   exty,
		IsBase: p.IsBase,
	}
}

// A gnark circuit version of LogDerivSumParams
type GnarkLogDerivSumParams struct {
	Sum zk.WrappedVariable
}

// A gnark circuit version of GrandProductParams
type GnarkGrandProductParams struct {
	Prod zk.WrappedVariable
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
	FinalResult zk.WrappedVariable
	// Parts are the parameters of the Horner parts
	Parts []HornerParamsPartGnark
}

func (p LogDerivSumParams) GnarkAssign() GnarkLogDerivSumParams {
	return GnarkLogDerivSumParams{Sum: p.Sum}
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
	X      zk.WrappedVariable
	Ys     []zk.WrappedVariable
	ExtX   gnarkfext.E4Gen
	ExtYs  []gnarkfext.E4Gen
	IsBase bool
}

func (p UnivariateEval) GnarkAllocate() GnarkUnivariateEvalParams {
	// no need to preallocate the x because its size is already known
	return GnarkUnivariateEvalParams{
		Ys:    make([]zk.WrappedVariable, len(p.Pols)),
		ExtYs: make([]gnarkfext.E4Gen, len(p.Pols)),
	}
}

// Returns a gnark assignment for the present parameters
func (p UnivariateEvalParams) GnarkAssign() GnarkUnivariateEvalParams {
	if p.IsBase {
		return GnarkUnivariateEvalParams{
			Ys:    vector.IntoGnarkAssignment(p.Ys),
			X:     p.X,
			ExtYs: vectorext.IntoGnarkAssignment(p.ExtYs),
			ExtX:  gnarkfext.SetFromExt(p.ExtX),
		}
	} else {
		// extension query
		return GnarkUnivariateEvalParams{
			Ys:    nil,
			X:     field.Zero(),
			ExtYs: vectorext.IntoGnarkAssignment(p.ExtYs),
			ExtX:  gnarkfext.SetFromExt(p.ExtX),
		}
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
			N0: part.N0,
			N1: part.N1,
		}
	}

	return GnarkHornerParams{
		FinalResult: p.FinalResult,
		Parts:       parts,
	}
}

// Update the fiat-shamir state with the the present parameters
func (p GnarkInnerProductParams) UpdateFS(fs *fiatshamir.GnarkFiatShamir) {
	fs.UpdateExt(p.Ys...)
}

// Update the fiat-shamir state with the the present parameters
func (p GnarkLocalOpeningParams) UpdateFS(fs *fiatshamir.GnarkFiatShamir) {
	fs.Update(p.BaseY)
}

// Update the fiat-shamir state with the the present parameters
func (p GnarkLogDerivSumParams) UpdateFS(fs *fiatshamir.GnarkFiatShamir) {
	fs.Update(p.Sum)
}

// Update the fiat-shamir state with the the present parameters
func (p GnarkGrandProductParams) UpdateFS(fs *fiatshamir.GnarkFiatShamir) {
	fs.Update(p.Prod)
}

// Update the fiat-shamir state with the the present parameters
func (p GnarkUnivariateEvalParams) UpdateFS(fs *fiatshamir.GnarkFiatShamir) {
	fs.Update(p.Ys...)
}

// Update the fiat-shamir state with the the present field extension parameters
func (p GnarkUnivariateEvalParams) UpdateFSExt(fs *fiatshamir.GnarkFiatShamir) {
	fs.UpdateExt(p.ExtYs...)
}

// Update the fiat-shamir state with the the present parameters
func (p GnarkHornerParams) UpdateFS(fs *fiatshamir.GnarkFiatShamir) {
	fs.Update(p.FinalResult)

	for _, part := range p.Parts {
		fs.Update(part.N0, part.N1)
	}
}
