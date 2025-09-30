package variables

import (
	"math/big"

	"github.com/consensys/gnark-crypto/field/koalabear/fft"
	sv "github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/zk"
	"github.com/consensys/linea-monorepo/prover/symbolic"
)

/*
Refers to an abstract variable X over which all polynomials
are defined.
*/
type X[T zk.Element] struct{}

// Construct a new variable for coin
func NewXVar[T zk.Element]() *symbolic.Expression {
	return symbolic.NewVariable(X[T]{})
}

// to implement symbolic.Metadata
func (x X[T]) String() string {
	// Double append/prepend to avoid confusion
	return "__X__"
}

// Returns an evaluation of the X, possibly over a coset. Pass
// `EvalCoset(size, 0, 0, false)` to directly evaluate over a coset
func (x X[T]) EvalCoset(size, cosetId, cosetRatio int, shiftGen bool) sv.SmartVector {
	omega, err := fft.Generator(uint64(size))
	if err != nil {
		panic(err)
	}
	omegaQNumCos, err1 := fft.Generator(uint64(size * cosetRatio))
	if err1 != nil {
		panic(err1)
	}
	omegaQNumCos.Exp(omegaQNumCos, big.NewInt(int64(cosetId)))

	omegaI := field.One()
	if shiftGen {
		omegaI = field.NewElement(field.MultiplicativeGen)
	}
	omegaI.Mul(&omegaI, &omegaQNumCos)

	// precomputations of the powers of omega, can be optimized if useful
	omegas := make([]field.Element, size)
	for i := range omegas {
		omegas[i] = omegaI
		omegaI.Mul(&omegaI, &omega)
	}

	return sv.NewRegular(omegas)
}

// Evaluate the variable, but not over a coset
func (x X[T]) GnarkEvalNoCoset(size int) []T {
	res_ := x.EvalCoset(size, 0, 1, false)
	res := make([]T, res_.Len())
	for i := range res {
		res[i] = *zk.ValueOf[T](res_.Get(i))
	}
	return res
}

func (x X[T]) IsBase() bool {
	return true
}
