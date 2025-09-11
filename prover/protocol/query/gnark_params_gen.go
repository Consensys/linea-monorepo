package query

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/crypto/fiatshamir"
	"github.com/consensys/linea-monorepo/prover/maths/field/gnarkfext"
	"github.com/consensys/linea-monorepo/prover/protocol/zk"
)

// A gnark circuit version of the LocalOpeningResult
type GnarkLocalOpeningParamsGen[T zk.Element] struct {
	BaseY  frontend.Variable
	ExtY   gnarkfext.Element
	IsBase bool
}

// func (p LocalOpeningParams) GnarkAssign() GnarkLocalOpeningParamsGen {
// 	var exty gnarkfext.Element
// 	exty.Assign(p.ExtY)
// 	return GnarkLocalOpeningParams{
// 		BaseY:  p.BaseY,
// 		ExtY:   exty,
// 		IsBase: p.IsBase,
// 	}
// }

// A gnark circuit version of LogDerivSumParams
type GnarkLogDerivSumParamsGen[T zk.Element] struct {
	Sum T
}

// A gnark circuit version of GrandProductParams
type GnarkGrandProductParamsGen[T zk.Element] struct {
	Prod T
}

// HornerParamsPartGnark is a [HornerParamsPart] in a gnark circuit.
type HornerParamsPartGnarkGen[T zk.Element] struct {
	// N0 is an initial offset of the Horner query
	N0 T
	// N1 is the second offset of the Horner query
	N1 T
}

// GnarkHornerParams represents the parameters of the Horner evaluation query
// in a gnark circuit.
type GnarkHornerParamsGen[T zk.Element] struct {
	// Final result is the result of summing the Horner parts for every
	// queries.
	FinalResult T
	// Parts are the parameters of the Horner parts
	Parts []HornerParamsPartGnarkGen[T]
}

// func (p LogDerivSumParams) GnarkAssign() GnarkLogDerivSumParams {
// 	return GnarkLogDerivSumParams{Sum: p.Sum}
// }

// A gnark circuit version of InnerProductParams
type GnarkInnerProductParamsGen[T zk.Element] struct {
	Ys []gnarkfext.E4Gen[T]
}

// func (p InnerProduct) GnarkAllocate() GnarkInnerProductParams {
// 	return GnarkInnerProductParams{Ys: make([]gnarkfext.Element, len(p.Bs))}
// }

// func (p InnerProductParams) GnarkAssign() GnarkInnerProductParams {
// 	return GnarkInnerProductParams{Ys: vectorext.IntoGnarkAssignment(p.Ys)}
// }

// A gnark circuit version of univariate eval params
type GnarkUnivariateEvalParamsGen[T zk.Element] struct {
	X      T
	Ys     []T
	ExtX   gnarkfext.E4Gen[T]
	ExtYs  []gnarkfext.Element
	IsBase bool
}

// func (p UnivariateEval) GnarkAllocate() GnarkUnivariateEvalParams {
// 	// no need to preallocate the x because its size is already known
// 	return GnarkUnivariateEvalParams{
// 		Ys:    make([]frontend.Variable, len(p.Pols)),
// 		ExtYs: make([]gnarkfext.Element, len(p.Pols)),
// 	}
// }

// Returns a gnark assignment for the present parameters
// func (p UnivariateEvalParams) GnarkAssign() GnarkUnivariateEvalParams {
// 	if p.IsBase {
// 		return GnarkUnivariateEvalParams{
// 			Ys:    vector.IntoGnarkAssignment(p.Ys),
// 			X:     p.X,
// 			ExtYs: vectorext.IntoGnarkAssignment(p.ExtYs),
// 			ExtX:  gnarkfext.SetFromExt(p.ExtX),
// 		}
// 	} else {
// 		// extension query
// 		return GnarkUnivariateEvalParams{
// 			Ys:    nil,
// 			X:     field.Zero(),
// 			ExtYs: vectorext.IntoGnarkAssignment(p.ExtYs),
// 			ExtX:  gnarkfext.SetFromExt(p.ExtX),
// 		}
// 	}
// }

// GnarkAllocate allocates a [GnarkHornerParams] with the right dimensions
// func (p HornerParams) GnarkAllocate() GnarkHornerParams {
// 	return GnarkHornerParams{
// 		Parts: make([]HornerParamsPartGnark, len(p.Parts)),
// 	}
// }

// GnarkAssign returns a gnark assignment for the present parameters.
// func (p HornerParams) GnarkAssign() GnarkHornerParams {

// 	parts := make([]HornerParamsPartGnark, len(p.Parts))
// 	for i, part := range p.Parts {
// 		parts[i] = HornerParamsPartGnark{
// 			N0: part.N0,
// 			N1: part.N1,
// 		}
// 	}

// 	return GnarkHornerParams{
// 		FinalResult: p.FinalResult,
// 		Parts:       parts,
// 	}
// }

// Update the fiat-shamir state with the the present parameters
func (p GnarkInnerProductParamsGen[T]) UpdateFS(fs *fiatshamir.GnarkFiatShamir) {
	// TODO @thomas update FS gen
	//fs.UpdateExt(p.Ys...)
}

// Update the fiat-shamir state with the the present parameters
func (p GnarkLocalOpeningParamsGen[T]) UpdateFS(fs *fiatshamir.GnarkFiatShamir) {
	// TODO @thomas update FS gen
	// fs.Update(p.BaseY)
}

// Update the fiat-shamir state with the the present parameters
func (p GnarkLogDerivSumParamsGen[T]) UpdateFS(fs *fiatshamir.GnarkFiatShamir) {
	// TODO @thomas update FS gen
	// fs.Update(p.Sum)
}

// Update the fiat-shamir state with the the present parameters
func (p GnarkGrandProductParamsGen[T]) UpdateFS(fs *fiatshamir.GnarkFiatShamir) {
	// TODO @thomas update FS gen
	// fs.Update(p.Prod)
}

// Update the fiat-shamir state with the the present parameters
func (p GnarkUnivariateEvalParamsGen[T]) UpdateFS(fs *fiatshamir.GnarkFiatShamir) {
	// TODO @thomas update FS gen
	// fs.Update(p.Ys...)
}

// Update the fiat-shamir state with the the present field extension parameters
func (p GnarkUnivariateEvalParamsGen[T]) UpdateFSExt(fs *fiatshamir.GnarkFiatShamir) {
	// TODO @thomas update FS gen
	// fs.UpdateExt(p.ExtYs...)
}

// Update the fiat-shamir state with the the present parameters
func (p GnarkHornerParamsGen[T]) UpdateFS(fs *fiatshamir.GnarkFiatShamir) {
	// TODO @thomas update FS gen
	// fs.Update(p.FinalResult)

	// for _, part := range p.Parts {
	// 	fs.Update(part.N0, part.N1)
	// }
}
