package compiler_test

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/protocol/compiler"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/plonkinwizard"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/poseidon2"
	"github.com/consensys/linea-monorepo/prover/protocol/internal/testtools"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/sirupsen/logrus"
)

var totalSuite = []func(comp *wizard.CompiledIOP){
	poseidon2.CompilePoseidon2,

	plonkinwizard.Compile,
	compiler.Arcane(
		compiler.WithDebugMode("debug"),
		compiler.WithStitcherMinSize(1),
		compiler.WithTargetColSize(8),
	),
	dummy.Compile,

	// vortex.Compile(2, vortex.ReplaceSisByMimc(), vortex.ForceNumOpenedColumns(2)),
}

func TestCompilers(t *testing.T) {

	logrus.SetLevel(logrus.FatalLevel)

	// runTestList(t, "fixed-permutation", testtools.ListOfFixedPermutationTestcasePositive)
	// runTestList(t, "global", testtools.ListOfGlobalTestcasePositive)
	// runTestList(t, "global", testtools.ListOfGlobalTestcaseNegative)
	// runTestList(t, "grand-product", testtools.ListOfGrandProductTestcasePositive)
	// runTestList(t, "grand-product", testtools.ListOfGrandProductTestcaseNegative)
	// runTestList(t, "horner", testtools.ListOfHornerTestcaseNegative)
	// runTestList(t, "horner", testtools.ListOfHornerTestcasePositive)
	// runTestList(t, "innerproduct", testtools.ListOfInnerProductTestcasePositive)
	// runTestList(t, "logderivativesum", testtools.ListOfLogDerivativeSumTestcasePositive)
	// runTestList(t, "logderivativesum", testtools.ListOfLogDerivativeSumTestcaseNegative)
	// runTestList(t, "permutation", testtools.ListOfPermutationTestcasePositive)
	// runTestList(t, "permutation", testtools.ListOfPermutationTestcaseNegative)
	// runTestList(t, "projection", testtools.ListOfProjectionTestcasePositive)
	// runTestList(t, "projection", testtools.ListOfProjectionTestcaseNegative)
	// runTestList(t, "mimc", testtools.ListOfMiMCTestcase)
	runTestList(t, "poseidon2", testtools.ListOfPoseidon2Testcase)

}

func TestCompilersWithGnarkVerifier(t *testing.T) {

	logrus.SetLevel(logrus.FatalLevel)

	runTestListGnark(t, "global", testtools.ListOfGlobalTestcasePositive)
	runTestListGnark(t, "horner", testtools.ListOfHornerTestcasePositive)
	runTestListGnark(t, "grand-product", testtools.ListOfGrandProductTestcasePositive)
	runTestListGnark(t, "projection", testtools.ListOfProjectionTestcasePositive)
	runTestListGnark(t, "permutation", testtools.ListOfPermutationTestcasePositive)
	runTestListGnark(t, "logderivativesum", testtools.ListOfLogDerivativeSumTestcasePositive)
	runTestListGnark(t, "mimc", testtools.ListOfMiMCTestcase)
	runTestListGnark(t, "fixed-permutation", testtools.ListOfFixedPermutationTestcasePositive)
}

func runTestList[T testtools.Testcase](t *testing.T, prefix string, list []T) {

	t.Run(prefix, func(t *testing.T) {
		for _, tc := range list {
			t.Run(tc.Name(), func(t *testing.T) {

				defer func() {
					if r := recover(); r != nil {
						t.Errorf("Test got a panic: %v", r)
					}
				}()

				testtools.RunTestcase(t, tc, totalSuite)
			})
		}
	})
}

func runTestListGnark[T testtools.Testcase](t *testing.T, prefix string, list []T) {

	t.Run(prefix, func(t *testing.T) {
		for _, tc := range list {
			t.Run(tc.Name(), func(t *testing.T) {
				testtools.RunTestShouldPassWithGnark(t, tc, totalSuite)
			})
		}
	})
}
