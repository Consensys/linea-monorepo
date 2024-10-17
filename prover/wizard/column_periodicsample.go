package wizard

import (
	"math/big"
	"strconv"

	"github.com/consensys/gnark/frontend"
	sv "github.com/consensys/zkevm-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/zkevm-monorepo/prover/maths/common/vector"
	"github.com/consensys/zkevm-monorepo/prover/maths/fft"
	"github.com/consensys/zkevm-monorepo/prover/maths/field"
	"github.com/consensys/zkevm-monorepo/prover/utils"
)

var _ Column = ColPeriodicSampling{}

// ColPeriodicSampling refers to an abstract value that is 0 on every entries
// and one on every entries i such that i % T == 0. It is also used to instantiate
// concretely Lagrange polynomials.
type ColPeriodicSampling struct {
	T             int
	Offset        int // Should be < T
	size          int
	shiftByMulGen bool
	cosetID       int
	cosetRatio    int
}

// NewColPeriodicSample constructs a new [ColPeriodicSampling], will panic if the
// offset it larger than T
func NewColPeriodicSampling(period, offset, size int) ColPeriodicSampling {
	return newPeriodicSample(period, offset, size)
}

// NewColLagrange constructs a new column representing a Lagrange polynomial
func NewColLagrange(size, offset int) ColPeriodicSampling {
	return newPeriodicSample(size, offset, size)
}

func newPeriodicSample(period, offset, size int) ColPeriodicSampling {

	if !utils.IsPowerOfTwo(period) {
		utils.Panic("non power of two period %v", period)
	}

	if period > size {
		utils.Panic("The period %v cannot be larger than the size %v", period, size)
	}

	if offset < 0 {
		utils.Panic("negative offset %v", offset)
	}

	if offset >= period {
		utils.Panic("The offset can't be larger than the period. offset = %v, period = %v", offset, period)
	}

	return ColPeriodicSampling{
		T:          period,
		Offset:     offset,
		cosetRatio: 1,
		size:       size,
	}
}

// OnCoset returns an equivalent column defined on a coset.
func (t ColPeriodicSampling) OnCoset(cosetID, cosetRatio int, shiftByMulGen bool) ColPeriodicSampling {

	var (
		newT = t
	)

	if cosetID > 0 {
		gcd := utils.GCD(cosetID, cosetRatio)
		newT.cosetID = cosetID / gcd
		newT.cosetRatio = cosetRatio / gcd
	}

	if cosetID == 0 {
		newT.cosetID = 0
		newT.cosetRatio = 1
	}

	newT.shiftByMulGen = shiftByMulGen

	return newT
}

// to implement symbolic.Metadata
func (t ColPeriodicSampling) String() string {
	// Double append/prepend to avoid confusion
	res := "periodic-sampling/t=" + strconv.Itoa(t.T) + "/offset=" + strconv.Itoa(t.Offset)

	if t.cosetID > 0 {
		res += "/" + strconv.Itoa(t.cosetID) + "/" + strconv.Itoa(t.cosetRatio)
	}

	if t.shiftByMulGen {
		res += "/shifted-by-mul-gen"
	}

	return res
}

func (t ColPeriodicSampling) Shift(n int) Column {
	newT := t
	t.Offset -= n
	return newT
}

func (t ColPeriodicSampling) Size() int {
	return t.size
}

func (t ColPeriodicSampling) GetAssignment(_ Runtime) sv.SmartVector {

	// When the periodic sampling is not over a coset
	if !t.shiftByMulGen && t.cosetID == 0 {
		res := make([]field.Element, t.size)
		for i := range res {
			if i%t.T == t.Offset {
				res[i].SetOne()
			}
		}
		return sv.NewRegular(res)
	}

	var (
		a           = field.One()
		l           = t.size / t.T
		denominator = make([]field.Element, t.T)
		al, an      field.Element
		one         = field.One()
		omegal      = fft.GetOmega(t.T) // It's the canonical t-root of unity
	)

	// Compute the coset shifting value
	if t.shiftByMulGen {
		a = field.NewElement(field.MultiplicativeGen)
	}

	// Skip if there is no coset ratio
	if t.cosetRatio > 0 {
		omegaN := fft.GetOmega(t.size * t.cosetRatio)
		omegaN.Exp(omegaN, big.NewInt(int64(t.cosetID)))
		a.Mul(&a, &omegaN)
	}

	// If there is an offset in the sample we also adjust here
	if t.Offset > 0 {
		omegalInv := fft.GetOmega(t.size)
		omegalInv.Exp(omegalInv, big.NewInt(int64(-t.Offset)))
		a.Mul(&a, &omegalInv)
	}

	// Precomputes the values of a^n and a^l and omega^l
	al.Exp(a, big.NewInt(int64(l)))
	an.Exp(a, big.NewInt(int64(t.size)))

	// Compute the denominator
	denominator[0] = al

	for i := 1; i < t.T; i++ {
		denominator[i].Mul(&denominator[i-1], &omegal)
		denominator[i-1].Sub(&denominator[i-1], &one)
	}

	denominator[t.T-1].Sub(&denominator[t.T-1], &one)
	denominator = field.BatchInvert(denominator)

	var (
		constTerm = an
		lField    = field.NewElement(uint64(l))
		nField    = field.NewElement(uint64(t.size))
	)

	// Compute the constant term l / n (a^n - 1)
	constTerm.Sub(&constTerm, &one)
	nField.Inverse(&nField)
	constTerm.Mul(&constTerm, &lField)
	constTerm.Mul(&constTerm, &nField)

	vector.ScalarMul(denominator, denominator, constTerm)

	// Now, we just need to repeat it "l" time and we can return
	res := make([]field.Element, t.T, t.size)
	copy(res, denominator)
	for len(res) < t.size {
		res = append(res, res...)
	}

	return sv.NewRegular(res)
}

func (t ColPeriodicSampling) GetAssignmentGnark(_ frontend.API, _ RuntimeGnark) []frontend.Variable {

	var (
		res_ = t.GetAssignment(nil)
		res  = make([]frontend.Variable, res_.Len())
	)

	for i := range res {
		res[i] = res_.Get(i)
	}
	return res
}

func (t ColPeriodicSampling) Round() int {
	return 0
}
