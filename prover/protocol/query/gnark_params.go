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

func (p LocalOpeningParams[T]) GnarkAssign() GnarkLocalOpeningParams[T] {
	return GnarkLocalOpeningParams[T]{
		BaseY:  *zk.ValueOf[T](p.BaseY),
		ExtY:   gnarkfext.NewE4Gen[T](p.ExtY),
		IsBase: p.IsBase,
	}
}

// A gnark circuit version of LogDerivSumParams
type GnarkLogDerivSumParams[T zk.Element] struct {
	Sum T
}

func (g *GnarkLogDerivSumParams[T]) GnarkAssign(p LogDerivSumParams[T]) {
	g.Sum = *zk.ValueOf[T](p.Sum)
}

// A gnark circuit version of GrandProductParams
type GnarkGrandProductParams[T zk.Element] struct {
	Prod gnarkfext.E4Gen[T]
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
	FinalResult gnarkfext.E4Gen[T]
	// Parts are the parameters of the Horner parts
	Parts []HornerParamsPartGnark[T]
}

// GnarkAllocate allocates a [GnarkHornerParams] with the right dimensions
func (p HornerParams[T]) GnarkAllocate() GnarkHornerParams[T] {
	return GnarkHornerParams[T]{
		Parts: make([]HornerParamsPartGnark[T], len(p.Parts)),
	}
}

// GnarkAssign returns a gnark assignment for the present parameters.
func (p HornerParams[T]) GnarkAssign() GnarkHornerParams[T] {

	parts := make([]HornerParamsPartGnark[T], len(p.Parts))
	for i, part := range p.Parts {
		parts[i] = HornerParamsPartGnark[T]{
			N0: *zk.ValueOf[T](part.N0),
			N1: *zk.ValueOf[T](part.N1),
		}
	}

	return GnarkHornerParams[T]{
		FinalResult: gnarkfext.NewE4Gen[T](p.FinalResult),
		Parts:       parts,
	}
}

func (p LogDerivSumParams[T]) GnarkAssign() GnarkLogDerivSumParams[T] {
	var tmp T
	// return GnarkLogDerivSumParams[T]{Sum: p.Sum}
	return GnarkLogDerivSumParams[T]{Sum: tmp} // TODO @thomas fixme
}

// A gnark circuit version of InnerProductParams
type GnarkInnerProductParams[T zk.Element] struct {
	Ys []gnarkfext.E4Gen[T]
}

func (p InnerProduct[T]) GnarkAllocate() GnarkInnerProductParams[T] {
	return GnarkInnerProductParams[T]{Ys: make([]gnarkfext.E4Gen[T], len(p.Bs))}
}

func (p InnerProductParams[T]) GnarkAssign() GnarkInnerProductParams[T] {
	return GnarkInnerProductParams[T]{Ys: vectorext.IntoGnarkAssignment[T](p.Ys)}
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
func (p UnivariateEvalParams[T]) GnarkAssign() GnarkUnivariateEvalParams[T] {

	var g GnarkUnivariateEvalParams[T]

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

	return g
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
func (p GnarkLogDerivSumParams[T]) UpdateFS(fs *fiatshamir.GnarkFiatShamir) {
	fs.Update(p.Sum)
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
