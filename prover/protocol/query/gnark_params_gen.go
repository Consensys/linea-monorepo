package query

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/crypto/fiatshamir"
	"github.com/consensys/linea-monorepo/prover/maths/common/vector"
	"github.com/consensys/linea-monorepo/prover/maths/common/vectorext"
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

// Allocate allocates a [GnarkHornerParams] with the right dimensions
func (g *GnarkHornerParamsGen[T]) Allocate(p HornerParams) {
	g.Parts = make([]HornerParamsPartGnarkGen[T], len(p.Parts))
}

func (g *GnarkHornerParamsGen[T]) Assign(p HornerParams) {
	g.Parts = make([]HornerParamsPartGnarkGen[T], len(p.Parts))
	for i := 0; i < len(p.Parts); i++ {
		g.Parts[i] = HornerParamsPartGnarkGen[T]{
			N0: zk.ValueOf[T](p.Parts[i].N0),
			N1: zk.ValueOf[T](p.Parts[i].N1),
		}
	}
}

// func (p LogDerivSumParams) GnarkAssign() GnarkLogDerivSumParams {
// 	return GnarkLogDerivSumParams{Sum: p.Sum}
// }

// A gnark circuit version of InnerProductParams
type GnarkInnerProductParamsGen[T zk.Element] struct {
	Ys []gnarkfext.E4Gen[T]
}

// Allocate allocates a [GnarkHornerParams] with the right dimensions
func (g *GnarkInnerProductParamsGen[T]) Allocate(p InnerProduct) {
	g.Ys = make([]gnarkfext.E4Gen[T], len(p.Bs))
}

func (g *GnarkInnerProductParamsGen[T]) Assign(p InnerProductParams) {
	g.Ys = vectorext.IntoGnarkAssignmentGen[T](p.Ys)
}

// A gnark circuit version of univariate eval params
type GnarkUnivariateEvalParamsGen[T zk.Element] struct {
	X      T
	Ys     []T
	ExtX   gnarkfext.E4Gen[T]
	ExtYs  []gnarkfext.E4Gen[T]
	IsBase bool
}

func (g *GnarkUnivariateEvalParamsGen[T]) Allocate(p UnivariateEval) {
	g.Ys = make([]T, len(p.Pols))
	g.ExtYs = make([]gnarkfext.E4Gen[T], len(p.Pols))

}

// Returns a gnark assignment for the present parameters
func (g *GnarkUnivariateEvalParamsGen[T]) GnarkAssign(p UnivariateEvalParams) {
	if p.IsBase {
		g.Ys = vector.IntoGnarkAssignmentGen[T](p.Ys)
		g.X = zk.ValueOf[T](p.X)
	} else {
		// extension query
		g.Ys = nil
		g.X = zk.ValueOf[T](0)
	}
	g.ExtYs = vectorext.IntoGnarkAssignmentGen[T](p.ExtYs)
	g.ExtX = gnarkfext.NewE4Gen[T](p.ExtX)
}

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
