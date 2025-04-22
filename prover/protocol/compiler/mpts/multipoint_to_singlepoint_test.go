package mpts

import (
	"fmt"
	"testing"

	"github.com/consensys/linea-monorepo/prover/maths/common/mempool"
	"github.com/consensys/linea-monorepo/prover/maths/common/vector"
	"github.com/consensys/linea-monorepo/prover/maths/fft"
	"github.com/consensys/linea-monorepo/prover/maths/fft/fastpoly"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/protocol/internal/testtools"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

func TestCompiler(t *testing.T) {

	suite := []func(*wizard.CompiledIOP){
		Compile(),
		dummy.Compile,
	}

	for _, tc := range testtools.ListOfUnivariateTestcasesPositive {
		t.Run(tc.Name(), func(t *testing.T) {
			testtools.RunTestcase(t, tc, suite)
		})
	}
}

func TestLdeOf(t *testing.T) {

	testcases := []struct {
		Name string
		Poly []field.Element
		LDE  []field.Element
	}{
		{
			Name: "constant-poly",
			Poly: vector.Repeat(field.NewElement(23), 8),
			LDE:  vector.Repeat(field.NewElement(23), 32),
		},
		{
			Name: "x-poly",
			Poly: vector.ForTest(1, -1),
			LDE:  vector.PowerVec(fft.GetOmega(8), 8),
		},
		{
			Name: "x-poly-2",
			Poly: vector.PowerVec(fft.GetOmega(4), 4),
			LDE:  vector.PowerVec(fft.GetOmega(8), 8),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.Name, func(t *testing.T) {

			var (
				sizeBig = len(tc.LDE)
				memPool = mempool.CreateFromSyncPool(sizeBig)
				resPtr  = ldeOf(tc.Poly, memPool)
				res     = *resPtr
			)

			for i := range tc.LDE {
				if !tc.LDE[i].Equal(&res[i]) {
					t.Errorf("mismatch res=%v, expected=%v", vector.Prettify(res), vector.Prettify(tc.LDE))
					return
				}
			}
		})
	}

}

func TestCompilerDebug(t *testing.T) {

	var (
		omega8           = fft.GetOmega(8)
		powersOfOmega    = vector.PowerVec(omega8, 8)
		poly             = vector.ForTest(0, 0, 0, 0, 0, 1, 0, 0)
		x                = field.NewElement(0)
		y                = fastpoly.Interpolate(poly, x)
		powersOfOmegaInv = field.BatchInvert(powersOfOmega)
		quotient         = make([]field.Element, 8)
		r                = field.NewElement(2)
		rho              = field.NewFromString("987628969677496025087526830832245684720032240592929493744400676641038345321")
	)

	fmt.Printf("y = %v\n", y.String())

	for i := range quotient {
		quotient[i].Sub(&poly[i], &y)
		quotient[i].Mul(&quotient[i], &powersOfOmegaInv[i])
	}

	var (
		q0 = append([]field.Element{}, quotient...)
		q1 = append([]field.Element{}, quotient...)
	)

	vector.ScalarMul(q1, q1, rho)

	fmt.Printf("q0: %v\n", vector.Prettify(q0))
	fmt.Printf("q1: %v\n", vector.Prettify(q1))

	var rhoPlusOne = field.One()
	rhoPlusOne.Add(&rho, &rhoPlusOne)
	vector.ScalarMul(quotient, quotient, rhoPlusOne)
	fmt.Printf("quotient: %v\n", vector.Prettify(quotient))

	var (
		qr   = fastpoly.Interpolate(quotient, r)
		pr   = fastpoly.Interpolate(poly, r)
		rInv field.Element
	)

	rInv.Inverse(&r)

	fmt.Printf("qr: %v\n", qr.String())
	fmt.Printf("pr: %v\n", pr.String())
	fmt.Printf("rInv: %v\n", rInv.String())
	fmt.Printf("y: %v\n", y.String())

	tmp := pr
	tmp.Sub(&pr, &y)
	fmt.Printf("tmp: %v\n", tmp.String())
	tmp.Mul(&tmp, &rInv)
	tmp.Mul(&tmp, &rhoPlusOne)

	fmt.Printf("res: %v\n", tmp.String())
}
