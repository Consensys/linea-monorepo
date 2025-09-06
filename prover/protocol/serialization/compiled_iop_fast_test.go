package serialization_test

import (
	"fmt"
	"testing"

	"github.com/consensys/linea-monorepo/prover/crypto/ringsis"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/cleanup"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/globalcs"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/localcs"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/logderivativesum"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/mimc"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/mpts"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/plonkinwizard"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/recursion"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/selfrecursion"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/univariates"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/vortex"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/serialization"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/sirupsen/logrus"
)

var ser = serialization.NewSerializer()

func TestSerdeIOP1_Fast(t *testing.T) {

	for _, tc := range testcases {
		t.Run(fmt.Sprintf("testcase-%++v", tc), func(subT *testing.T) {
			define := generateProtocol(tc)
			sisInstances := tc.SisInstance

			comp := wizard.Compile(
				define,
				vortex.Compile(
					2,
					vortex.ForceNumOpenedColumns(16),
					vortex.WithSISParams(&sisInstances),
				),
				selfrecursion.SelfRecurse,
				dummy.Compile,
			)

			_, err := ser.PackCompiledIOPFast(comp)
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestSerdeIOP2_Fast(t *testing.T) {

	tc := TestCase{Numpoly: 32, NumRound: 3, PolSize: 32, NumOpenCol: 16, SisInstance: sisInstances[0]}
	t.Run(fmt.Sprintf("testcase-%++v", tc), func(subT *testing.T) {
		define := generateProtocol(tc)

		comp := wizard.Compile(
			define,
			vortex.Compile(
				2,
				vortex.ForceNumOpenedColumns(tc.NumOpenCol),
				vortex.WithSISParams(&tc.SisInstance),
			),
			selfrecursion.SelfRecurse,
			mimc.CompileMiMC,
			compiler.Arcane(
				compiler.WithTargetColSize(1<<10)),
			vortex.Compile(
				2,
				vortex.ForceNumOpenedColumns(tc.NumOpenCol),
				vortex.WithSISParams(&tc.SisInstance),
			),
			selfrecursion.SelfRecurse,
			mimc.CompileMiMC,
			compiler.Arcane(
				compiler.WithTargetColSize(1<<13)),
			vortex.Compile(
				2,
				vortex.ForceNumOpenedColumns(tc.NumOpenCol),
				vortex.WithSISParams(&tc.SisInstance),
			),
			dummy.Compile,
		)

		_, err := ser.PackCompiledIOPFast(comp)
		if err != nil {
			t.Fatal(err)
		}
	})
}

// Test for committing to the precomputed polynomials
func TestSerdeIOP3_Fast(t *testing.T) {

	for _, tc := range testcases_precomp {
		t.Run(fmt.Sprintf("testcase-%++v", tc), func(subT *testing.T) {
			define := generateProtocol(tc)
			sisInstances := tc.SisInstance
			comp := wizard.Compile(
				define,
				vortex.Compile(
					2,
					vortex.ForceNumOpenedColumns(16),
					vortex.WithSISParams(&sisInstances),
				),
				selfrecursion.SelfRecurse,
				dummy.Compile,
			)

			_, err := ser.PackCompiledIOPFast(comp)
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}

// Test for precomputed polys with multilayered self recursion
func TestSerdeIOP4_Fast(t *testing.T) {

	logrus.SetLevel(logrus.FatalLevel)

	tc := TestCase{Numpoly: 32, NumRound: 3, PolSize: 32, NumOpenCol: 16, SisInstance: sisInstances[0],
		NumPrecomp: 4, IsCommitPrecomp: true}
	t.Run(fmt.Sprintf("testcase-%++v", tc), func(subT *testing.T) {
		define := generateProtocol(tc)

		comp := wizard.Compile(
			define,
			vortex.Compile(
				2,
				vortex.ForceNumOpenedColumns(16),
				vortex.WithSISParams(&tc.SisInstance),
			),
			selfrecursion.SelfRecurse,
			mimc.CompileMiMC,
			compiler.Arcane(
				compiler.WithTargetColSize(1<<10)),
			vortex.Compile(
				2,
				vortex.ForceNumOpenedColumns(16),
				vortex.WithSISParams(&tc.SisInstance),
			),
			selfrecursion.SelfRecurse,
			mimc.CompileMiMC,
			compiler.Arcane(
				compiler.WithTargetColSize(1<<13)),
			vortex.Compile(
				2,
				vortex.ForceNumOpenedColumns(16),
				vortex.WithSISParams(&tc.SisInstance),
			),
			dummy.Compile,
		)

		_, err := ser.PackCompiledIOPFast(comp)
		if err != nil {
			t.Fatal(err)
		}
	})
}

func TestSerdeIOP5_Fast(t *testing.T) {

	numRow := 1 << 10
	tc := distributeTestCase{numRow: numRow}
	sisInstance := ringsis.Params{LogTwoBound: 16, LogTwoDegree: 6}

	t.Run(fmt.Sprintf("testcase-%++v", tc), func(subT *testing.T) {

		comp := wizard.Compile(
			func(build *wizard.Builder) {
				tc.define(build.CompiledIOP)
			},
			mimc.CompileMiMC,
			plonkinwizard.Compile,
			compiler.Arcane(
				compiler.WithTargetColSize(1<<17),
				compiler.WithDebugMode("conglomeration"),
			),
			vortex.Compile(
				2,
				vortex.ForceNumOpenedColumns(256),
				vortex.WithSISParams(&sisInstance),
				vortex.AddMerkleRootToPublicInputs(lppMerkleRootPublicInput, []int{0}),
			),
			selfrecursion.SelfRecurse,
			cleanup.CleanUp,
			mimc.CompileMiMC,
			compiler.Arcane(
				compiler.WithTargetColSize(1<<15),
			),
			vortex.Compile(
				8,
				vortex.ForceNumOpenedColumns(64),
				vortex.WithSISParams(&sisInstance),
			),
			selfrecursion.SelfRecurse,
			cleanup.CleanUp,
			mimc.CompileMiMC,
			compiler.Arcane(
				compiler.WithTargetColSize(1<<13),
			),
			vortex.Compile(
				8,
				vortex.ForceNumOpenedColumns(64),
				vortex.WithOptionalSISHashingThreshold(1<<20),
			),
		)

		_, err := ser.PackCompiledIOPFast(comp)
		if err != nil {
			t.Fatal(err)
		}
	})
}

func TestSerdeIOP6_Fast(t *testing.T) {

	define1 := func(bui *wizard.Builder) {

		var (
			a = bui.RegisterCommit("A", 8)
			b = bui.RegisterCommit("B", 8)
		)

		bui.Inclusion("Q", []ifaces.Column{a}, []ifaces.Column{b})
	}

	suites := [][]func(*wizard.CompiledIOP){
		{
			logderivativesum.CompileLookups,
			localcs.Compile,
			globalcs.Compile,
			univariates.Naturalize,
			mpts.Compile(),
			vortex.Compile(
				2,
				vortex.ForceNumOpenedColumns(4),
				vortex.WithSISParams(&ringsis.StdParams),
				vortex.PremarkAsSelfRecursed(),
				vortex.WithOptionalSISHashingThreshold(0),
			),
		},
	}

	for i, s := range suites {

		t.Run(fmt.Sprintf("case-%v", i), func(t *testing.T) {
			comp1 := wizard.Compile(define1, s...)
			// runSerdeTest(t, comp1, fmt.Sprintf("iop6-recursion-comp1-%v", i), true, false)
			_, err := ser.PackCompiledIOPFast(comp1)
			if err != nil {
				t.Fatal(err)
			}
			define2 := func(build2 *wizard.Builder) {
				recursion.DefineRecursionOf(build2.CompiledIOP, comp1, recursion.Parameters{
					Name:        "test",
					WithoutGkr:  true,
					MaxNumProof: 1,
				})
			}
			comp2 := wizard.Compile(define2, dummy.CompileAtProverLvl())
			// runSerdeTest(t, comp2, fmt.Sprintf("iop6-recursion-comp2-%v", i), true, false)
			_, err = ser.PackCompiledIOPFast(comp2)
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}
