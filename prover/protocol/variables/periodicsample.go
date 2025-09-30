package variables

import (
	"fmt"
	"math/big"

	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/gnark-crypto/field/koalabear/fft"
	"github.com/consensys/gnark/frontend"
	sv "github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/common/vector"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/protocol/zk"
	"github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/sirupsen/logrus"
)

// Refers to an abstract value that is 0 on every entries and one
// on every entries i such that i % T == 0
type PeriodicSample[T zk.Element] struct {
	W      int
	Offset int // Should be < T
}

// Constructs a new PeriodicSample, will panic if the offset it larger
// than T
func NewPeriodicSample[T zk.Element](period, offset int) *symbolic.Expression[T] {

	offset = utils.PositiveMod(offset, period)

	if !utils.IsPowerOfTwo(period) {
		utils.Panic("non power of two period %v", period)
	}

	if offset < 0 {
		utils.Panic("negative offset %v", offset)
	}

	if offset >= period {
		utils.Panic("The offset can't be larger than the period. offset = %v, period = %v", offset, period)
	}

	return symbolic.NewVariable[T](PeriodicSample[T]{
		W:      period,
		Offset: offset,
	})
}

// Lagrange returns a PeriodicSampling representing a Lagrange polynomial
func Lagrange[T zk.Element](n, pos int) *symbolic.Expression[T] {
	return NewPeriodicSample[T](n, pos)
}

// to implement symbolic.Metadata
func (t PeriodicSample[T]) String() string {
	// Double append/prepend to avoid confusion
	return fmt.Sprintf("__PERIODIC_SAMPLE_%v_OFFSET_%v__", t.W, t.Offset)
}

func (t PeriodicSample[T]) EvalAtOnDomain(pos int) field.Element {
	if pos%t.W == t.Offset {
		return field.One()
	}
	return field.Zero()
}

// Evaluates the expression outside of the domain
func (t PeriodicSample[T]) EvalAtOutOfDomain(size int, x field.Element) field.Element {
	n := size
	l := n / t.W
	one := field.One()
	lField := field.NewElement(uint64(l))
	nField := field.NewElement(uint64(n))
	evalPoint := x

	// If there is an offset in the sample we also adjust here
	if t.Offset > 0 {
		var shift field.Element
		omegaN, _ := fft.Generator(uint64(size))
		evalPoint.Mul(&evalPoint, shift.Exp(omegaN, big.NewInt(int64(-t.Offset))))
	}

	var denominator, numerator field.Element
	denominator.Exp(evalPoint, big.NewInt(int64(l)))
	denominator.Sub(&denominator, &one)
	denominator.Mul(&denominator, &nField)
	numerator.Exp(evalPoint, big.NewInt(int64(n)))
	numerator.Sub(&numerator, &one)
	numerator.Mul(&numerator, &lField)

	if denominator.IsZero() {
		panic("denominator was zero")
	}

	var res field.Element
	res.Div(&numerator, &denominator)
	return res
}

// Evaluates the expression outside of the domain
func (t PeriodicSample[T]) EvalAtOutOfDomainExt(size int, x fext.Element) fext.Element {
	l := size / t.W
	one := fext.One()
	var lField, nField fext.Element
	nField.B0.A0.SetUint64(uint64(size))
	lField.B0.A0.SetUint64(uint64(l))

	// If there is an offset in the sample we also adjust here
	if t.Offset > 0 {
		var shift field.Element
		omegaN, _ := fft.Generator(uint64(size))
		x.MulByElement(&x, shift.Exp(omegaN, big.NewInt(int64(-t.Offset))))
	}

	var denominator, numerator fext.Element
	denominator.Exp(x, big.NewInt(int64(l)))
	denominator.Sub(&denominator, &one)
	denominator.Mul(&denominator, &nField)
	numerator.Exp(x, big.NewInt(int64(size)))
	numerator.Sub(&numerator, &one)
	numerator.Mul(&numerator, &lField)

	if denominator.IsZero() {
		panic("denominator was zero")
	}

	var res fext.Element
	res.Div(&numerator, &denominator)
	return res
}

// Evaluate a particular position on the domain
func (t PeriodicSample[T]) GnarkEvalAtOnDomain(api frontend.API, pos int) T {
	return t.GnarkEvalNoCoset(t.W)[pos%t.W]
}

func (t PeriodicSample[T]) GnarkEvalAtOutOfDomain(api frontend.API, size int, x T) T {
	n := size
	l := n / t.W
	lField := field.NewElement(uint64(l))
	nField := field.NewElement(uint64(n))

	apiGen, err := zk.NewApi[T](api)
	if err != nil {
		panic(err)
	}

	// If there is an offset in the sample we also adjust here
	if t.Offset > 0 {
		omegaN, err := fft.Generator(uint64(n))
		if err != nil {
			panic(err)
		}
		//x = api.Mul(x, gnarkutil.Exp(api, omegaN, -t.Offset))
		var negOmegaN koalabear.Element
		negOmegaN.Neg(&omegaN)
		tmp := field.Exp(apiGen, *zk.ValueOf[T](negOmegaN), big.NewInt(int64(t.Offset)))
		x = *apiGen.Mul(&x, tmp)
	}

	denominator := field.Exp(apiGen, x, big.NewInt(int64(l)))
	denominator = apiGen.Sub(denominator, zk.ValueOf[T](1))
	denominator = apiGen.Mul(denominator, zk.ValueOf[T](nField))
	numerator := field.Exp(apiGen, x, big.NewInt(int64(n)))
	numerator = apiGen.Sub(numerator, zk.ValueOf[T](1))
	numerator = apiGen.Mul(numerator, zk.ValueOf[T](lField))

	return *apiGen.Div(numerator, denominator)
}

// Returns the result in gnark form. This returns a vector of constant
// on the form of Ts.
func (t PeriodicSample[T]) GnarkEvalNoCoset(size int) []T {
	res_ := t.EvalCoset(size, 0, 1, false)
	res := make([]T, res_.Len())
	for i := range res {
		res[i] = *zk.ValueOf[T](res_.Get(i))
	}
	return res
}

// Returns an evaluation of the periodic sample (possibly) over a coset.
// To not push it over a coset : pass EvalCoset(size, 0, {0,1}, false)
func (t PeriodicSample[T]) EvalCoset(size, cosetId, cosetRatio int, shiftGen bool) sv.SmartVector {
	// The return value is T periodic so we only pay for `X^n - 1` on the coset
	// https://hackmd.io/S78bJUa0Tk-T256iduE22g#Computing-the-evaluations-for-Z

	// Computes the integers constant
	n := size
	l := n / t.W
	one := field.One()

	// sanity-check the evaluation domain is too small for this to make sense
	if n < t.W {
		utils.Panic("tried evaluation on a domain of size %v but the period is %v", n, t.W)
	}

	// sanity-check : normally this code is never called for a coset if shiftGen is false
	if cosetRatio >= 2 != shiftGen {
		logrus.Infof("passed coset ratio %v but the shiftGen is %v", cosetRatio, shiftGen)
	}

	// Edge case, the evaluation is not done on a coset. Directly return the ideal value
	// Without this special handling, we would divide by zero.
	if !shiftGen && cosetId == 0 {
		res := make([]field.Element, size)
		for i := range res {
			if i%t.W == t.Offset {
				res[i].SetOne()
			}
		}
		return sv.NewRegular(res)
	}

	// Compute the coset shifting value
	a := field.One()
	if shiftGen {
		a = field.NewElement(field.MultiplicativeGen)
	}

	// Skip if there is no coset ratio
	if cosetRatio > 0 {
		omegaN, err := fft.Generator(uint64(n * cosetRatio))
		if err != nil {
			panic(err)
		}
		omegaN.Exp(omegaN, big.NewInt(int64(cosetId)))
		a.Mul(&a, &omegaN)
	}

	// If there is an offset in the sample we also adjust here
	if t.Offset > 0 {
		omegalInv, err := fft.Generator(uint64(n))
		if err != nil {
			panic(err)
		}
		omegalInv.Exp(omegalInv, big.NewInt(int64(-t.Offset)))
		a.Mul(&a, &omegalInv)
	}

	// Precomputes the values of a^n and a^l and omega^l
	var al, an field.Element
	al.Exp(a, big.NewInt(int64(l)))
	an.Exp(a, big.NewInt(int64(n)))
	omegal, err := fft.Generator(uint64(t.W)) // It's the canonical t-root of unity
	if err != nil {
		panic(err)
	}

	// Denominator
	denominator := make([]field.Element, t.W)
	denominator[0] = al

	for i := 1; i < t.W; i++ {
		denominator[i].Mul(&denominator[i-1], &omegal)
		denominator[i-1].Sub(&denominator[i-1], &one)
	}

	denominator[t.W-1].Sub(&denominator[t.W-1], &one)
	denominator = field.BatchInvert(denominator)

	/*
		Compute the constant term l / n (a^n - 1)
	*/
	constTerm := an
	constTerm.Sub(&constTerm, &one)
	lField := field.NewElement(uint64(l))
	nField := field.NewElement(uint64(n))
	nField.Inverse(&nField)
	constTerm.Mul(&constTerm, &lField)
	constTerm.Mul(&constTerm, &nField)

	vector.ScalarMul(denominator, denominator, constTerm)

	// Now, we just need to repeat it "l" time and we can return
	res := make([]field.Element, t.W, n)
	copy(res, denominator)
	for len(res) < n {
		res = append(res, res...)
	}

	return sv.NewRegular(res)
}

func (t PeriodicSample[T]) IsBase() bool {

	return true
}
