package mpts

import (
	"math/rand/v2"
	"testing"

	"github.com/consensys/gnark-crypto/field/koalabear/fft"
	"github.com/consensys/linea-monorepo/prover/maths/common/mempool"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/common/vector"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/protocol/column/verifiercol"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/internal/testtools"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
)

func TestWithProfile(t *testing.T) {
	var rng = rand.New(utils.NewRandSource(0))

	testcases := []struct {
		Profile []int
		UTC     *testtools.UnivariateTestcase
	}{
		{
			Profile: []int{3},
			UTC: &testtools.UnivariateTestcase{
				NameStr: "profile-3",
				Polys: []smartvectors.SmartVector{
					smartvectors.ForTest(1, 2, 3, 4),
				},
				QueryXs: []fext.Element{
					fext.PseudoRand(rng),
				},
				QueryPols: [][]int{
					{0},
				},
			},
		},
	}

	for _, tc := range testcases {

		suite := []func(*wizard.CompiledIOP){
			Compile(WithNumColumnProfileOpt(tc.Profile, 0)),
			dummy.Compile,
		}

		t.Run(tc.UTC.NameStr, func(t *testing.T) {
			testtools.RunTestcase(t, tc.UTC, suite)
		})
	}

}

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

/*
TODO@yao

func TestCompilerWithGnark(t *testing.T) {

		suite := []func(*wizard.CompiledIOP){
			Compile(),
			dummy.Compile,
		}

		for _, tc := range testtools.ListOfUnivariateTestcasesPositive {
			t.Run(tc.Name(), func(t *testing.T) {
				testtools.RunTestShouldPassWithGnark(t, tc, suite)
			})
		}
	}
*/
func TestWithVerifierCol(t *testing.T) {

	suite := []func(*wizard.CompiledIOP){
		Compile(),
		dummy.Compile,
	}

	testcases := []*testtools.AnonymousTestcase{
		{
			NameStr: "with-constant-col",
			DefineFunc: func(comp *wizard.CompiledIOP) {
				u := verifiercol.NewConstantCol(field.Zero(), 8)
				comp.InsertUnivariate(0, "U", []ifaces.Column{u})
			},
			AssignFunc: func(run *wizard.ProverRuntime) {
				run.AssignUnivariate("U", fext.Zero(), fext.Zero())
			},
		},
		{
			NameStr: "with-constant-col-2",
			DefineFunc: func(comp *wizard.CompiledIOP) {
				u := verifiercol.NewConstantCol(field.NewElement(42), 4)
				comp.InsertUnivariate(0, "U", []ifaces.Column{u})
			},
			AssignFunc: func(run *wizard.ProverRuntime) {
				run.AssignUnivariate("U", fext.Zero(), fext.NewElement(42, 0, 0, 0))
			},
		},
	}

	for i, tc := range testcases {
		t.Run(tc.NameStr, func(t *testing.T) {
			testtools.RunTestcase(t, testcases[i], suite)
		})
	}
}

func TestLdeOf(t *testing.T) {

	gen_8, err := fft.Generator(8)
	if err != nil {
		panic(err)
	}

	gen_4, err1 := fft.Generator(4)
	if err1 != nil {
		panic(err1)
	}
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
			LDE:  vector.PowerVec(gen_8, 8),
		},
		{
			Name: "x-poly-2",
			Poly: vector.PowerVec(gen_4, 4),
			LDE:  vector.PowerVec(gen_8, 8),
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
