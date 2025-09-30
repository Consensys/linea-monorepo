package query

import (
	"github.com/consensys/linea-monorepo/prover/crypto/fiatshamir"
	"github.com/consensys/linea-monorepo/prover/maths/common/vector"
	"github.com/consensys/linea-monorepo/prover/maths/common/vectorext"
	"github.com/consensys/linea-monorepo/prover/maths/field/gnarkfext"
	"github.com/consensys/linea-monorepo/prover/protocol/zk"
)

// A gnark circuit version of the LocalOpeningResult
type GnarkLocalOpeningParams[T zk.Element] struct {
	BaseY  T
	ExtY   gnarkfext.E4Gen[T]
	IsBase bool
}

func (g *GnarkLocalOpeningParams[T]) GnarkAssign(p LocalOpeningParams) {
	g.BaseY = *zk.ValueOf[T](p.BaseY)
	g.ExtY = gnarkfext.NewE4Gen[T](p.ExtY)
	g.IsBase = p.IsBase
}

// A gnark circuit version of LogDerivSumParams
type GnarkLogDerivSumParams[T zk.Element] struct {
	Sum T
}

func (g *GnarkLogDerivSumParams[T]) GnarkAssign(p LogDerivSumParams) {
	g.Sum = *zk.ValueOf[T](p.Sum)
}

// A gnark circuit version of GrandProductParams
type GnarkGrandProductParams[T zk.Element] struct {
	Prod T
}

// HornerParamsPartGnark is a [HornerParamsPart] in a gnark circuit.
type HornerParamsPartGnark[T zk.Element] struct {
	// N0 is an initial offset of the Horner query
	N0 T
	// N1 is the second offset of the Horner query
	N1 T
}

// GnarkHornerParams represents the parameters of the Horner evaluation query
// in a gnark circuit.
type GnarkHornerParams[T zk.Element] struct {
	// Final result is the result of summing the Horner parts for every
	// queries.
	FinalResult T
	// Parts are the parameters of the Horner parts
	Parts []HornerParamsPartGnark[T]
}

// Allocate allocates a [GnarkHornerParams] with the right dimensions
func (g *GnarkHornerParams[T]) Allocate(p HornerParams[T]) {
	g.Parts = make([]HornerParamsPartGnark[T], len(p.Parts))
}

func (g *GnarkHornerParams[T]) Assign(p HornerParams[T]) {
	g.Parts = make([]HornerParamsPartGnark[T], len(p.Parts))
	for i := 0; i < len(p.Parts); i++ {
		g.Parts[i] = HornerParamsPartGnark[T]{
			N0: *zk.ValueOf[T](p.Parts[i].N0),
			N1: *zk.ValueOf[T](p.Parts[i].N1),
		}
	}
}

// A gnark circuit version of InnerProductParams
type GnarkInnerProductParams[T zk.Element] struct {
	Ys []gnarkfext.E4Gen[T]
}

// Allocate allocates a [GnarkHornerParams] with the right dimensions
func (g *GnarkInnerProductParams[T]) Allocate(p InnerProduct[T]) {
	g.Ys = make([]gnarkfext.E4Gen[T], len(p.Bs))
}

func (g *GnarkInnerProductParams[T]) Assign(p InnerProductParams) {
	g.Ys = vectorext.IntoGnarkAssignment[T](p.Ys)
}

// A gnark circuit version of univariate eval params
type GnarkUnivariateEvalParams[T zk.Element] struct {
	X      T
	Ys     []T
	ExtX   gnarkfext.E4Gen[T]
	ExtYs  []gnarkfext.E4Gen[T]
	IsBase bool
}

func (p UnivariateEval[T]) GnarkAllocate() GnarkUnivariateEvalParams[T] {
	// no need to preallocate the x because its size is already known
	return GnarkUnivariateEvalParams[T]{
		Ys:    make([]T, len(p.Pols)),
		ExtYs: make([]gnarkfext.E4Gen[T], len(p.Pols)),
	}
}

// Returns a gnark assignment for the present parameters
func (g *GnarkUnivariateEvalParams[T]) GnarkAssign(p UnivariateEvalParams) {
	if p.IsBase {
		g.Ys = vector.IntoGnarkAssignment[T](p.Ys)
		g.X = *zk.ValueOf[T](p.X)
	} else {
		// extension query
		g.Ys = nil
		g.X = *zk.ValueOf[T](0)
	}
	g.ExtYs = vectorext.IntoGnarkAssignment[T](p.ExtYs)
	g.ExtX = gnarkfext.NewE4Gen[T](p.ExtX)
}

// Update the fiat-shamir state with the the present parameters
func (p GnarkInnerProductParams[T]) UpdateFS(fs *fiatshamir.GnarkFiatShamir) {
	// TODO @thomas update FS gen
	//fs.UpdateExt(p.Ys...)
}

// Update the fiat-shamir state with the the present parameters
func (p GnarkLocalOpeningParams[T]) UpdateFS(fs *fiatshamir.GnarkFiatShamir) {
	// TODO @thomas update FS gen
	// fs.Update(p.BaseY)
}

// Update the fiat-shamir state with the the present parameters
func (p GnarkGrandProductParams[T]) UpdateFS(fs *fiatshamir.GnarkFiatShamir) {
	// TODO @thomas update FS gen
	// fs.Update(p.Prod)
}

// Update the fiat-shamir state with the the present parameters
func (p GnarkUnivariateEvalParams[T]) UpdateFS(fs *fiatshamir.GnarkFiatShamir) {
	// TODO @thomas update FS gen
	// fs.Update(p.Ys...)
}

// Update the fiat-shamir state with the the present field extension parameters
func (p GnarkUnivariateEvalParams[T]) UpdateFSExt(fs *fiatshamir.GnarkFiatShamir) {
	// TODO @thomas update FS gen
	// fs.UpdateExt(p.ExtYs...)
}

// Update the fiat-shamir state with the the present parameters
func (p GnarkHornerParams[T]) UpdateFS(fs *fiatshamir.GnarkFiatShamir) {
	// TODO @thomas update FS gen
	// fs.Update(p.FinalResult)

	// for _, part := range p.Parts {
	// 	fs.Update(part.N0, part.N1)
	// }
}
