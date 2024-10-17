package wizard

import (
	"math/big"
	"strconv"

	"github.com/consensys/gnark/frontend"
	sv "github.com/consensys/zkevm-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/zkevm-monorepo/prover/maths/fft"
	"github.com/consensys/zkevm-monorepo/prover/maths/field"
	"github.com/consensys/zkevm-monorepo/prover/utils"
)

var _ Column = &ColumnX{}

// ColumnX refers to an abstract variable X over which all polynomials are defined. It
// implements the [Column] interface and represents a column storing the
// successive powers of the generator root of unity.
//
// This column is used to cancel expressions at specific points.
type ColumnX struct {
	size          int
	cosetID       int
	cosetRatio    int
	shiftByMulGen bool
}

// Construct a new variable for coin
func NewXVar(size int) *ColumnX {
	return &ColumnX{
		size:       size,
		cosetRatio: 1,
	}
}

// OnCoset shifts the X variable over a different coset.
func (x *ColumnX) OnCoset(cosetID int, cosetRatio int, shiftByMulGen bool) *ColumnX {
	var (
		newX = *x
	)

	if cosetID > 0 {
		gcd := utils.GCD(cosetID, cosetRatio)
		newX.cosetID = cosetID / gcd
		newX.cosetRatio = cosetRatio / gcd
	}

	if cosetID == 0 {
		newX.cosetID = 0
		newX.cosetRatio = 1
	}

	newX.shiftByMulGen = shiftByMulGen

	return &newX
}

// to implement symbolic.Metadata
func (x ColumnX) String() string {
	// Double append/prepend to avoid confusion
	res := "x"

	if x.cosetID > 0 {
		res += "/" + strconv.Itoa(x.cosetID) + "/" + strconv.Itoa(x.cosetRatio)
	}

	if x.shiftByMulGen {
		res += "/shifted-by-mul-gen"
	}

	return res
}

// Returns an evaluation of the X, possibly over a coset. Pass
// `GetAssignment(size, 0, 0, false)` to directly evaluate over a coset
func (x *ColumnX) GetAssignment(run Runtime) sv.SmartVector {
	var (
		size         = x.size
		cosetRatio   = x.cosetRatio
		cosetId      = x.cosetID
		omega        = fft.GetOmega(size)
		omegaQNumCos = fft.GetOmega(size * cosetRatio)
	)
	omegaQNumCos.Exp(omegaQNumCos, big.NewInt(int64(cosetId)))

	omegaI := field.One()
	if x.shiftByMulGen {
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
func (x *ColumnX) GetAssignmentGnark(_ frontend.API, _ RuntimeGnark) []frontend.Variable {
	var (
		res_ = x.GetAssignment(nil)
		res  = make([]frontend.Variable, res_.Len())
	)
	for i := range res {
		res[i] = res_.Get(i)
	}
	return res
}

func (x *ColumnX) Round() int {
	return 0
}

func (x *ColumnX) Size() int {
	return x.size
}

func (x *ColumnX) Shift(n int) Column {
	panic("unimplemented")
}
