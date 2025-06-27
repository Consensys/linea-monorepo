package permutation

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

func TestPermutationPass(t *testing.T) {

	testCases := []struct {
		Define     func(*wizard.Builder)
		Prove      func(*wizard.ProverRuntime)
		Title      string
		ShouldPass bool
	}{
		{
			Define: func(builder *wizard.Builder) {
				a := builder.RegisterCommit("A", 16)
				b := builder.RegisterCommit("B", 16)
				builder.Permutation("PERM", []ifaces.Column{a}, []ifaces.Column{b})
			},

			Prove: func(run *wizard.ProverRuntime) {
				run.AssignColumn("A", smartvectors.ForTest(0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15))
				run.AssignColumn("B", smartvectors.ForTest(15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1, 0))
			},

			Title:      "single-column",
			ShouldPass: true,
		},
		{
			Define: func(builder *wizard.Builder) {
				a := []ifaces.Column{
					builder.RegisterCommit("A1", 16),
					builder.RegisterCommit("A2", 16),
				}
				b := []ifaces.Column{
					builder.RegisterCommit("B1", 16),
					builder.RegisterCommit("B2", 16),
				}
				// the same permutation amon columns of a and b
				builder.Permutation("PERM", a, b)
			},

			Prove: func(run *wizard.ProverRuntime) {
				run.AssignColumn("A1", smartvectors.ForTest(0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15))
				run.AssignColumn("A2", smartvectors.ForTest(10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 110, 111, 112, 113, 114, 115))
				run.AssignColumn("B1", smartvectors.ForTest(15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1, 0))
				run.AssignColumn("B2", smartvectors.ForTest(115, 114, 113, 112, 111, 110, 19, 18, 17, 16, 15, 14, 13, 12, 11, 10))
			},

			Title:      "two-columnss",
			ShouldPass: true,
		},
		{
			Define: func(builder *wizard.Builder) {
				a := []ifaces.Column{
					builder.RegisterCommit("A1", 16),
					builder.RegisterCommit("A2", 16),
				}
				b := []ifaces.Column{
					builder.RegisterCommit("B1", 16),
					builder.RegisterCommit("B2", 16),
				}
				// PERM1 does not have to be the same permutations as PERM2
				builder.Permutation("PERM1", a[:1], b[:1])
				builder.Permutation("PERM2", a[1:], b[1:])
			},

			Prove: func(run *wizard.ProverRuntime) {
				run.AssignColumn("A1", smartvectors.ForTest(0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15))
				run.AssignColumn("A2", smartvectors.ForTest(11, 10, 12, 13, 14, 15, 16, 17, 18, 19, 110, 111, 112, 113, 114, 115))
				run.AssignColumn("B1", smartvectors.ForTest(15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1, 0))
				run.AssignColumn("B2", smartvectors.ForTest(115, 114, 113, 112, 111, 110, 19, 18, 17, 16, 15, 14, 13, 12, 11, 10))
			},

			Title:      "two queries for one column",
			ShouldPass: true,
		},
		{
			Define: func(builder *wizard.Builder) {
				a := []ifaces.Column{
					builder.RegisterCommit("A1", 16),
					builder.RegisterCommit("A2", 16),
					builder.RegisterCommit("A3", 16),
					builder.RegisterCommit("A4", 16),
				}
				b := []ifaces.Column{
					builder.RegisterCommit("B1", 16),
					builder.RegisterCommit("B2", 16),
					builder.RegisterCommit("B3", 16),
					builder.RegisterCommit("B4", 16),
				}
				builder.Permutation("PERM1", a[0:1], b[0:1])
				builder.Permutation("PERM2", a[1:2], b[1:2])
				builder.Permutation("PERM3", a[2:3], b[2:3])
				builder.Permutation("PERM4", a[3:4], b[3:4])
			},

			Prove: func(run *wizard.ProverRuntime) {
				run.AssignColumn("A1", smartvectors.ForTest(0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15))
				run.AssignColumn("A2", smartvectors.ForTest(11, 10, 12, 13, 14, 15, 16, 17, 18, 19, 110, 111, 112, 113, 114, 115))
				run.AssignColumn("A3", smartvectors.ForTest(22, 20, 22, 23, 24, 25, 26, 27, 28, 29, 220, 222, 222, 223, 224, 225))
				run.AssignColumn("A4", smartvectors.ForTest(122, 120, 122, 123, 124, 125, 126, 127, 128, 129, 1220, 1222, 1222, 1223, 1224, 1225))
				run.AssignColumn("B1", smartvectors.ForTest(15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1, 0))
				run.AssignColumn("B2", smartvectors.ForTest(115, 114, 113, 112, 111, 110, 19, 18, 17, 16, 15, 14, 13, 12, 11, 10))
				run.AssignColumn("B3", smartvectors.ForTest(225, 224, 223, 222, 222, 220, 29, 28, 27, 26, 25, 24, 23, 22, 22, 20))
				run.AssignColumn("B4", smartvectors.ForTest(1225, 1224, 1223, 1222, 1222, 1220, 129, 128, 127, 126, 125, 124, 123, 122, 122, 120))
			},

			Title:      "4 queries for one column, all on one column",
			ShouldPass: true,
		},
		{
			Define: func(builder *wizard.Builder) {
				a := []ifaces.Column{
					builder.RegisterCommit("A1", 8),
					builder.RegisterCommit("A2", 8),
					builder.RegisterCommit("A3", 8),
					builder.RegisterCommit("A4", 8),
					builder.RegisterCommit("A5", 8),
					builder.RegisterCommit("A6", 8),
					builder.RegisterCommit("A7", 8),
					builder.RegisterCommit("A8", 8),
				}
				b := []ifaces.Column{
					builder.RegisterCommit("B1", 8),
					builder.RegisterCommit("B2", 8),
					builder.RegisterCommit("B3", 8),
					builder.RegisterCommit("B4", 8),
					builder.RegisterCommit("B5", 8),
					builder.RegisterCommit("B6", 8),
					builder.RegisterCommit("B7", 8),
					builder.RegisterCommit("B8", 8),
				}
				// the permutation between a[0] and b [0] is the same permutation as the one between a[1] and b[1]
				builder.Permutation("PERM1", a[0:2], b[0:2])
				builder.Permutation("PERM2", a[2:4], b[2:4])
				builder.Permutation("PERM3", a[4:6], b[4:6])
				builder.Permutation("PERM4", a[6:], b[6:])
			},

			Prove: func(run *wizard.ProverRuntime) {
				run.AssignColumn("A1", smartvectors.ForTest(0, 1, 2, 3, 4, 5, 6, 7))
				run.AssignColumn("A2", smartvectors.ForTest(1, 2, 3, 4, 5, 6, 7, 8))
				run.AssignColumn("A3", smartvectors.ForTest(2, 3, 4, 5, 6, 7, 8, 9))
				run.AssignColumn("A4", smartvectors.ForTest(3, 4, 5, 6, 7, 8, 9, 10))
				run.AssignColumn("A5", smartvectors.ForTest(10, 11, 12, 13, 14, 15, 16, 17))
				run.AssignColumn("A6", smartvectors.ForTest(11, 12, 13, 14, 15, 16, 17, 18))
				run.AssignColumn("A7", smartvectors.ForTest(12, 13, 14, 15, 16, 17, 18, 19))
				run.AssignColumn("A8", smartvectors.ForTest(13, 14, 15, 16, 17, 18, 19, 110))

				run.AssignColumn("B1", smartvectors.ForTest(1, 2, 3, 4, 5, 6, 7, 0))
				run.AssignColumn("B2", smartvectors.ForTest(2, 3, 4, 5, 6, 7, 8, 1))
				run.AssignColumn("B3", smartvectors.ForTest(3, 4, 5, 6, 7, 8, 9, 2))
				run.AssignColumn("B4", smartvectors.ForTest(4, 5, 6, 7, 8, 9, 10, 3))
				run.AssignColumn("B5", smartvectors.ForTest(11, 12, 13, 14, 15, 16, 17, 10))
				run.AssignColumn("B6", smartvectors.ForTest(12, 13, 14, 15, 16, 17, 18, 11))
				run.AssignColumn("B7", smartvectors.ForTest(13, 14, 15, 16, 17, 18, 19, 12))
				run.AssignColumn("B8", smartvectors.ForTest(14, 15, 16, 17, 18, 19, 110, 13))
			},

			Title:      "4 queries for one column, all on two columns",
			ShouldPass: true,
		},
		{
			Define: func(builder *wizard.Builder) {
				a := []ifaces.Column{
					builder.RegisterCommit("A1", 4),
					builder.RegisterCommit("A2", 4),
					builder.RegisterCommit("A3", 4),
					builder.RegisterCommit("A4", 4),
					builder.RegisterCommit("A5", 8),
					builder.RegisterCommit("A6", 8),
					builder.RegisterCommit("A7", 8),
					builder.RegisterCommit("A8", 8),
				}
				b := []ifaces.Column{
					builder.RegisterCommit("B1", 4),
					builder.RegisterCommit("B2", 4),
					builder.RegisterCommit("B3", 4),
					builder.RegisterCommit("B4", 4),
					builder.RegisterCommit("B5", 8),
					builder.RegisterCommit("B6", 8),
					builder.RegisterCommit("B7", 8),
					builder.RegisterCommit("B8", 8),
				}
				builder.Permutation("PERM1", a[0:2], b[0:2])
				builder.Permutation("PERM2", a[2:4], b[2:4])
				builder.Permutation("PERM3", a[4:6], b[4:6])
				builder.Permutation("PERM4", a[6:], b[6:])
			},

			Prove: func(run *wizard.ProverRuntime) {
				run.AssignColumn("A1", smartvectors.ForTest(0, 1, 2, 3))
				run.AssignColumn("A2", smartvectors.ForTest(1, 2, 3, 4))
				run.AssignColumn("A3", smartvectors.ForTest(2, 3, 4, 5))
				run.AssignColumn("A4", smartvectors.ForTest(3, 4, 5, 6))
				run.AssignColumn("A5", smartvectors.ForTest(10, 11, 12, 13, 14, 15, 16, 17))
				run.AssignColumn("A6", smartvectors.ForTest(11, 12, 13, 14, 15, 16, 17, 18))
				run.AssignColumn("A7", smartvectors.ForTest(12, 13, 14, 15, 16, 17, 18, 19))
				run.AssignColumn("A8", smartvectors.ForTest(13, 14, 15, 16, 17, 18, 19, 110))

				run.AssignColumn("B1", smartvectors.ForTest(1, 2, 3, 0))
				run.AssignColumn("B2", smartvectors.ForTest(2, 3, 4, 1))
				run.AssignColumn("B3", smartvectors.ForTest(3, 4, 5, 2))
				run.AssignColumn("B4", smartvectors.ForTest(4, 5, 6, 3))
				run.AssignColumn("B5", smartvectors.ForTest(11, 12, 13, 14, 15, 16, 17, 10))
				run.AssignColumn("B6", smartvectors.ForTest(12, 13, 14, 15, 16, 17, 18, 11))
				run.AssignColumn("B7", smartvectors.ForTest(13, 14, 15, 16, 17, 18, 19, 12))
				run.AssignColumn("B8", smartvectors.ForTest(14, 15, 16, 17, 18, 19, 110, 13))
			},

			Title:      "4 queries for one column, all on two columns",
			ShouldPass: true,
		},
		{
			Define: func(builder *wizard.Builder) {
				a := []ifaces.Column{
					builder.RegisterCommit("A1", 4),
					builder.RegisterCommit("A2", 4),
					builder.RegisterCommit("A3", 4),
					builder.RegisterCommit("A4", 4),
					builder.RegisterCommit("A5", 8),
					builder.RegisterCommit("A6", 8),
					builder.RegisterCommit("A7", 8),
					builder.RegisterCommit("A8", 8),
				}
				b := []ifaces.Column{
					builder.RegisterCommit("B1", 4),
					builder.RegisterCommit("B2", 4),
					builder.RegisterCommit("B3", 4),
					builder.RegisterCommit("B4", 4),
					builder.RegisterCommit("B5", 8),
					builder.RegisterCommit("B6", 8),
					builder.RegisterCommit("B7", 8),
					builder.RegisterCommit("B8", 8),
				}
				// each fragment has its own permutation, but the permutation is the same for the columns from the same fragment.
				builder.CompiledIOP.InsertFragmentedPermutation(0, "PERM1", [][]ifaces.Column{a[0:2], a[2:4]}, [][]ifaces.Column{b[0:2], b[2:4]})
				builder.CompiledIOP.InsertFragmentedPermutation(0, "PERM2", [][]ifaces.Column{a[4:6], a[6:8]}, [][]ifaces.Column{b[4:6], b[6:8]})
			},

			Prove: func(run *wizard.ProverRuntime) {
				run.AssignColumn("A1", smartvectors.ForTest(0, 1, 2, 3))
				run.AssignColumn("A2", smartvectors.ForTest(1, 2, 3, 4))
				run.AssignColumn("A3", smartvectors.ForTest(2, 3, 4, 5))
				run.AssignColumn("A4", smartvectors.ForTest(3, 4, 5, 6))
				run.AssignColumn("A5", smartvectors.ForTest(10, 11, 12, 13, 14, 15, 16, 17))
				run.AssignColumn("A6", smartvectors.ForTest(11, 12, 13, 14, 15, 16, 17, 18))
				run.AssignColumn("A7", smartvectors.ForTest(12, 13, 14, 15, 16, 17, 18, 19))
				run.AssignColumn("A8", smartvectors.ForTest(13, 14, 15, 16, 17, 18, 19, 110))

				run.AssignColumn("B1", smartvectors.ForTest(1, 2, 3, 0))
				run.AssignColumn("B2", smartvectors.ForTest(2, 3, 4, 1))
				run.AssignColumn("B3", smartvectors.ForTest(2, 4, 3, 5))
				run.AssignColumn("B4", smartvectors.ForTest(3, 5, 4, 6))
				run.AssignColumn("B5", smartvectors.ForTest(11, 12, 13, 14, 15, 16, 17, 10))
				run.AssignColumn("B6", smartvectors.ForTest(12, 13, 14, 15, 16, 17, 18, 11))
				run.AssignColumn("B7", smartvectors.ForTest(13, 14, 15, 16, 17, 18, 19, 12))
				run.AssignColumn("B8", smartvectors.ForTest(14, 15, 16, 17, 18, 19, 110, 13))
			},

			Title:      "2 fragmented multi-column queries using different sizes",
			ShouldPass: true,
		},
		{
			Define: func(builder *wizard.Builder) {
				a := [][]ifaces.Column{
					{builder.RegisterCommit("A1", 4)},
					{builder.RegisterCommit("A2", 4)},
					{builder.RegisterCommit("A3", 4)},
					{builder.RegisterCommit("A4", 4)},
				}
				b := [][]ifaces.Column{
					{builder.RegisterCommit("B1", 16)},
				}
				builder.CompiledIOP.InsertFragmentedPermutation(0, "PERM1", a, b)
			},

			Prove: func(run *wizard.ProverRuntime) {
				run.AssignColumn("A1", smartvectors.ForTest(0, 1, 2, 3))
				run.AssignColumn("A2", smartvectors.ForTest(4, 5, 6, 7))
				run.AssignColumn("A3", smartvectors.ForTest(8, 9, 10, 11))
				run.AssignColumn("A4", smartvectors.ForTest(12, 13, 14, 15))

				run.AssignColumn("B1", smartvectors.ForTest(0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15))
			},

			Title:      "dissymetric fractional query",
			ShouldPass: true,
		},

		// test cases that should Not Pass
		{
			Define: func(builder *wizard.Builder) {
				a := builder.RegisterCommit("A", 16)
				b := builder.RegisterCommit("B", 16)
				builder.Permutation("PERM", []ifaces.Column{a}, []ifaces.Column{b})
			},

			Prove: func(run *wizard.ProverRuntime) {
				run.AssignColumn("A", smartvectors.ForTest(0, 2, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15))
				run.AssignColumn("B", smartvectors.ForTest(15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1, 0))
			},

			Title:      "single-column, should not pass",
			ShouldPass: false,
		},
		{
			Define: func(builder *wizard.Builder) {
				a := []ifaces.Column{
					builder.RegisterCommit("A1", 4),
					builder.RegisterCommit("A2", 4),
					builder.RegisterCommit("A3", 4),
					builder.RegisterCommit("A4", 4),
					builder.RegisterCommit("A5", 8),
					builder.RegisterCommit("A6", 8),
					builder.RegisterCommit("A7", 8),
					builder.RegisterCommit("A8", 8),
				}
				b := []ifaces.Column{
					builder.RegisterCommit("B1", 4),
					builder.RegisterCommit("B2", 4),
					builder.RegisterCommit("B3", 4),
					builder.RegisterCommit("B4", 4),
					builder.RegisterCommit("B5", 8),
					builder.RegisterCommit("B6", 8),
					builder.RegisterCommit("B7", 8),
					builder.RegisterCommit("B8", 8),
				}
				builder.CompiledIOP.InsertFragmentedPermutation(0, "PERM1", [][]ifaces.Column{a[0:2], a[2:4]}, [][]ifaces.Column{b[0:2], b[2:4]})
				builder.CompiledIOP.InsertFragmentedPermutation(0, "PERM2", [][]ifaces.Column{a[4:6], a[6:8]}, [][]ifaces.Column{b[4:6], b[6:8]})
			},

			Prove: func(run *wizard.ProverRuntime) {
				run.AssignColumn("A1", smartvectors.ForTest(0, 1, 2, 3))
				run.AssignColumn("A2", smartvectors.ForTest(2, 2, 3, 4))
				run.AssignColumn("A3", smartvectors.ForTest(2, 3, 4, 5))
				run.AssignColumn("A4", smartvectors.ForTest(3, 4, 5, 6))
				run.AssignColumn("A5", smartvectors.ForTest(10, 11, 12, 13, 14, 15, 16, 17))
				run.AssignColumn("A6", smartvectors.ForTest(11, 12, 13, 14, 15, 16, 17, 18))
				run.AssignColumn("A7", smartvectors.ForTest(12, 13, 14, 15, 16, 17, 18, 19))
				run.AssignColumn("A8", smartvectors.ForTest(13, 14, 15, 16, 17, 18, 19, 110))

				run.AssignColumn("B1", smartvectors.ForTest(1, 2, 3, 0))
				run.AssignColumn("B2", smartvectors.ForTest(2, 3, 4, 1))
				run.AssignColumn("B3", smartvectors.ForTest(3, 4, 5, 2))
				run.AssignColumn("B4", smartvectors.ForTest(4, 5, 6, 3))
				run.AssignColumn("B5", smartvectors.ForTest(11, 12, 13, 14, 15, 16, 17, 10))
				run.AssignColumn("B6", smartvectors.ForTest(12, 13, 14, 15, 16, 17, 18, 11))
				run.AssignColumn("B7", smartvectors.ForTest(13, 14, 15, 16, 17, 18, 19, 12))
				run.AssignColumn("B8", smartvectors.ForTest(14, 15, 16, 17, 18, 19, 110, 13))
			},

			Title:      "2 fragmented multi-column queries using different sizes, should not pass",
			ShouldPass: false,
		},

		{
			Define: func(builder *wizard.Builder) {
				a := [][]ifaces.Column{
					{builder.RegisterCommit("A1", 4)},
					{builder.RegisterCommit("A2", 4)},
					{builder.RegisterCommit("A3", 4)},
					{builder.RegisterCommit("A4", 4)},
				}
				b := [][]ifaces.Column{
					{builder.RegisterCommit("B1", 16)},
				}
				builder.CompiledIOP.InsertFragmentedPermutation(0, "PERM1", a, b)
			},

			Prove: func(run *wizard.ProverRuntime) {
				run.AssignColumn("A1", smartvectors.ForTest(1, 1, 2, 3))
				run.AssignColumn("A2", smartvectors.ForTest(4, 5, 6, 7))
				run.AssignColumn("A3", smartvectors.ForTest(8, 9, 10, 11))
				run.AssignColumn("A4", smartvectors.ForTest(12, 13, 14, 15))

				run.AssignColumn("B1", smartvectors.ForTest(0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15))
			},

			Title:      "dissymetric fractional query, should not pass",
			ShouldPass: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Title, func(t *testing.T) {
			comp := wizard.Compile(testCase.Define, CompileViaGrandProduct, dummy.Compile)
			proof := wizard.Prove(comp, testCase.Prove)
			if err := wizard.Verify(comp, proof); err != nil && testCase.ShouldPass {
				t.Fatalf("verifier did not pass: %v", err.Error())
			}
			if err := wizard.Verify(comp, proof); err == nil && !testCase.ShouldPass {
				t.Fatalf("verifier is passing for a false claim")
			}
		})
	}
}

func TestPermutationPassMixed(t *testing.T) {

	testCases := []struct {
		Define     func(*wizard.Builder)
		Prove      func(*wizard.ProverRuntime)
		Title      string
		ShouldPass bool
	}{
		{
			Define: func(builder *wizard.Builder) {
				a := builder.RegisterCommit("A", 16)
				b := builder.RegisterCommit("B", 16)
				builder.Permutation("PERM", []ifaces.Column{a}, []ifaces.Column{b})
			},

			Prove: func(run *wizard.ProverRuntime) {
				run.AssignColumn("A", smartvectors.ForTest(0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15))
				run.AssignColumn("B", smartvectors.ForTestExt(15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1, 0))
			},

			Title:      "single-column",
			ShouldPass: true,
		},
		{
			Define: func(builder *wizard.Builder) {
				a := []ifaces.Column{
					builder.RegisterCommit("A1", 16),
					builder.RegisterCommit("A2", 16),
				}
				b := []ifaces.Column{
					builder.RegisterCommit("B1", 16),
					builder.RegisterCommit("B2", 16),
				}
				// the same permutation amon columns of a and b
				builder.Permutation("PERM", a, b)
			},

			Prove: func(run *wizard.ProverRuntime) {
				run.AssignColumn("A1", smartvectors.ForTest(0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15))
				run.AssignColumn("A2", smartvectors.ForTestFromPairs(
					10, 30010, // second component is 30000 + first component
					11, 30011,
					12, 30012,
					13, 30013,
					14, 30014,
					15, 30015,
					16, 30016,
					17, 30017,
					18, 30018,
					19, 30019,
					110, 30110,
					111, 30111,
					112, 30112,
					113, 30113,
					114, 30114,
					115, 30115,
				))
				run.AssignColumn("B1", smartvectors.ForTest(15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1, 0))
				run.AssignColumn("B2", smartvectors.ForTestFromPairs(
					115, 30115, // second component is 30000 + first components
					114, 30114,
					113, 30113,
					112, 30112,
					111, 30111,
					110, 30110,
					19, 30019,
					18, 30018,
					17, 30017,
					16, 30016,
					15, 30015,
					14, 30014,
					13, 30013,
					12, 30012,
					11, 30011,
					10, 30010,
				))
			},

			Title:      "two-columnss",
			ShouldPass: true,
		},
		{
			Define: func(builder *wizard.Builder) {
				a := []ifaces.Column{
					builder.RegisterCommit("A1", 16),
					builder.RegisterCommit("A2", 16),
				}
				b := []ifaces.Column{
					builder.RegisterCommit("B1", 16),
					builder.RegisterCommit("B2", 16),
				}
				// PERM1 does not have to be the same permutations as PERM2
				builder.Permutation("PERM1", a[:1], b[:1])
				builder.Permutation("PERM2", a[1:], b[1:])
			},

			Prove: func(run *wizard.ProverRuntime) {
				run.AssignColumn("A1", smartvectors.ForTest(0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15))
				run.AssignColumn("A2", smartvectors.ForTestFromPairs(
					11, 100011, // second component is 100000 + first component
					10, 100010,
					12, 100012,
					13, 100013,
					14, 100014,
					15, 100015,
					16, 100016,
					17, 100017,
					18, 100018,
					19, 100019,
					110, 100110,
					111, 100111,
					112, 100112,
					113, 100113,
					114, 100114,
					115, 100115,
				))
				run.AssignColumn("B1", smartvectors.ForTest(15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1, 0))
				run.AssignColumn("B2", smartvectors.ForTestFromPairs(
					115, 100115, // second component is 100000 + first component
					114, 100114,
					113, 100113,
					112, 100112,
					111, 100111,
					110, 100110,
					19, 100019,
					18, 100018,
					17, 100017,
					16, 100016,
					15, 100015,
					14, 100014,
					13, 100013,
					12, 100012,
					11, 100011,
					10, 100010,
				))
			},

			Title:      "two queries for one column",
			ShouldPass: true,
		},
		{
			Define: func(builder *wizard.Builder) {
				a := []ifaces.Column{
					builder.RegisterCommit("A1", 16),
					builder.RegisterCommit("A2", 16),
					builder.RegisterCommit("A3", 16),
					builder.RegisterCommit("A4", 16),
				}
				b := []ifaces.Column{
					builder.RegisterCommit("B1", 16),
					builder.RegisterCommit("B2", 16),
					builder.RegisterCommit("B3", 16),
					builder.RegisterCommit("B4", 16),
				}
				builder.Permutation("PERM1", a[0:1], b[0:1])
				builder.Permutation("PERM2", a[1:2], b[1:2])
				builder.Permutation("PERM3", a[2:3], b[2:3])
				builder.Permutation("PERM4", a[3:4], b[3:4])
			},

			Prove: func(run *wizard.ProverRuntime) {
				run.AssignColumn("A1", smartvectors.ForTest(0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15))
				run.AssignColumn("A2", smartvectors.ForTest(11, 10, 12, 13, 14, 15, 16, 17, 18, 19, 110, 111, 112, 113, 114, 115))
				run.AssignColumn("A3", smartvectors.ForTestFromPairs(
					22, 9900022, // second component is 9900000 + first component
					20, 9900020,
					22, 9900022,
					23, 9900023,
					24, 9900024,
					25, 9900025,
					26, 9900026,
					27, 9900027,
					28, 9900028,
					29, 9900029,
					220, 9900220,
					222, 9900222,
					222, 9900222,
					223, 9900223,
					224, 9900224,
					225, 9900225,
				))
				run.AssignColumn("A4", smartvectors.ForTest(122, 120, 122, 123, 124, 125, 126, 127, 128, 129, 1220, 1222, 1222, 1223, 1224, 1225))
				run.AssignColumn("B1", smartvectors.ForTest(15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1, 0))
				run.AssignColumn("B2", smartvectors.ForTest(115, 114, 113, 112, 111, 110, 19, 18, 17, 16, 15, 14, 13, 12, 11, 10))
				run.AssignColumn("B3", smartvectors.ForTestFromPairs(
					225, 9900225, // second component is 9900000 + first component
					224, 9900224,
					223, 9900223,
					222, 9900222,
					222, 9900222,
					220, 9900220,
					29, 9900029,
					28, 9900028,
					27, 9900027,
					26, 9900026,
					25, 9900025,
					24, 9900024,
					23, 9900023,
					22, 9900022,
					22, 9900022,
					20, 9900020,
				))
				run.AssignColumn("B4", smartvectors.ForTest(1225, 1224, 1223, 1222, 1222, 1220, 129, 128, 127, 126, 125, 124, 123, 122, 122, 120))
			},

			Title:      "4 queries for one column, all on one column",
			ShouldPass: true,
		},
		{
			Define: func(builder *wizard.Builder) {
				a := []ifaces.Column{
					builder.RegisterCommit("A1", 8),
					builder.RegisterCommit("A2", 8),
					builder.RegisterCommit("A3", 8),
					builder.RegisterCommit("A4", 8),
					builder.RegisterCommit("A5", 8),
					builder.RegisterCommit("A6", 8),
					builder.RegisterCommit("A7", 8),
					builder.RegisterCommit("A8", 8),
				}
				b := []ifaces.Column{
					builder.RegisterCommit("B1", 8),
					builder.RegisterCommit("B2", 8),
					builder.RegisterCommit("B3", 8),
					builder.RegisterCommit("B4", 8),
					builder.RegisterCommit("B5", 8),
					builder.RegisterCommit("B6", 8),
					builder.RegisterCommit("B7", 8),
					builder.RegisterCommit("B8", 8),
				}
				// the permutation between a[0] and b [0] is the same permutation as the one between a[1] and b[1]
				builder.Permutation("PERM1", a[0:2], b[0:2])
				builder.Permutation("PERM2", a[2:4], b[2:4])
				builder.Permutation("PERM3", a[4:6], b[4:6])
				builder.Permutation("PERM4", a[6:], b[6:])
			},

			Prove: func(run *wizard.ProverRuntime) {
				run.AssignColumn("A1", smartvectors.ForTest(0, 1, 2, 3, 4, 5, 6, 7))
				run.AssignColumn("A2", smartvectors.ForTestFromPairs(
					1, 7770001, // second component is 7770000 + first component
					2, 7770002,
					3, 7770003,
					4, 7770004,
					5, 7770005,
					6, 7770006,
					7, 7770007,
					8, 7770008,
				))
				run.AssignColumn("A3", smartvectors.ForTest(2, 3, 4, 5, 6, 7, 8, 9))
				run.AssignColumn("A4", smartvectors.ForTest(3, 4, 5, 6, 7, 8, 9, 10))
				run.AssignColumn("A5", smartvectors.ForTest(10, 11, 12, 13, 14, 15, 16, 17))
				run.AssignColumn("A6", smartvectors.ForTest(11, 12, 13, 14, 15, 16, 17, 18))
				run.AssignColumn("A7", smartvectors.ForTest(12, 13, 14, 15, 16, 17, 18, 19))
				run.AssignColumn("A8", smartvectors.ForTest(13, 14, 15, 16, 17, 18, 19, 110))

				run.AssignColumn("B1", smartvectors.ForTest(1, 2, 3, 4, 5, 6, 7, 0))
				run.AssignColumn("B2", smartvectors.ForTestFromPairs(
					2, 7770002, // second component is 7770000 + first component
					3, 7770003,
					4, 7770004,
					5, 7770005,
					6, 7770006,
					7, 7770007,
					8, 7770008,
					1, 7770001,
				))
				run.AssignColumn("B3", smartvectors.ForTest(3, 4, 5, 6, 7, 8, 9, 2))
				run.AssignColumn("B4", smartvectors.ForTest(4, 5, 6, 7, 8, 9, 10, 3))
				run.AssignColumn("B5", smartvectors.ForTest(11, 12, 13, 14, 15, 16, 17, 10))
				run.AssignColumn("B6", smartvectors.ForTest(12, 13, 14, 15, 16, 17, 18, 11))
				run.AssignColumn("B7", smartvectors.ForTest(13, 14, 15, 16, 17, 18, 19, 12))
				run.AssignColumn("B8", smartvectors.ForTest(14, 15, 16, 17, 18, 19, 110, 13))
			},

			Title:      "4 queries for one column, all on two columns",
			ShouldPass: true,
		},
		{
			Define: func(builder *wizard.Builder) {
				a := []ifaces.Column{
					builder.RegisterCommit("A1", 4),
					builder.RegisterCommit("A2", 4),
					builder.RegisterCommit("A3", 4),
					builder.RegisterCommit("A4", 4),
					builder.RegisterCommit("A5", 8),
					builder.RegisterCommit("A6", 8),
					builder.RegisterCommit("A7", 8),
					builder.RegisterCommit("A8", 8),
				}
				b := []ifaces.Column{
					builder.RegisterCommit("B1", 4),
					builder.RegisterCommit("B2", 4),
					builder.RegisterCommit("B3", 4),
					builder.RegisterCommit("B4", 4),
					builder.RegisterCommit("B5", 8),
					builder.RegisterCommit("B6", 8),
					builder.RegisterCommit("B7", 8),
					builder.RegisterCommit("B8", 8),
				}
				builder.Permutation("PERM1", a[0:2], b[0:2])
				builder.Permutation("PERM2", a[2:4], b[2:4])
				builder.Permutation("PERM3", a[4:6], b[4:6])
				builder.Permutation("PERM4", a[6:], b[6:])
			},

			Prove: func(run *wizard.ProverRuntime) {
				run.AssignColumn("A1", smartvectors.ForTest(0, 1, 2, 3))
				run.AssignColumn("A2", smartvectors.ForTest(1, 2, 3, 4))
				run.AssignColumn("A3", smartvectors.ForTest(2, 3, 4, 5))
				run.AssignColumn("A4", smartvectors.ForTest(3, 4, 5, 6))
				run.AssignColumn("A5", smartvectors.ForTest(10, 11, 12, 13, 14, 15, 16, 17))
				run.AssignColumn("A6", smartvectors.ForTest(11, 12, 13, 14, 15, 16, 17, 18))
				run.AssignColumn("A7", smartvectors.ForTest(12, 13, 14, 15, 16, 17, 18, 19))
				run.AssignColumn("A8", smartvectors.ForTestFromPairs(
					13, 8880013, // second component is 8880000 + first component
					14, 8880014,
					15, 8880015,
					16, 8880016,
					17, 8880017,
					18, 8880018,
					19, 8880019,
					110, 8880110,
				))

				run.AssignColumn("B1", smartvectors.ForTest(1, 2, 3, 0))
				run.AssignColumn("B2", smartvectors.ForTest(2, 3, 4, 1))
				run.AssignColumn("B3", smartvectors.ForTest(3, 4, 5, 2))
				run.AssignColumn("B4", smartvectors.ForTest(4, 5, 6, 3))
				run.AssignColumn("B5", smartvectors.ForTest(11, 12, 13, 14, 15, 16, 17, 10))
				run.AssignColumn("B6", smartvectors.ForTest(12, 13, 14, 15, 16, 17, 18, 11))
				run.AssignColumn("B7", smartvectors.ForTest(13, 14, 15, 16, 17, 18, 19, 12))
				run.AssignColumn("B8", smartvectors.ForTestFromPairs(
					14, 8880014, // second component is 8880000 + first component
					15, 8880015,
					16, 8880016,
					17, 8880017,
					18, 8880018,
					19, 8880019,
					110, 8880110,
					13, 8880013,
				))
			},

			Title:      "4 queries for one column, all on two columns",
			ShouldPass: true,
		},
		{
			Define: func(builder *wizard.Builder) {
				a := []ifaces.Column{
					builder.RegisterCommit("A1", 4),
					builder.RegisterCommit("A2", 4),
					builder.RegisterCommit("A3", 4),
					builder.RegisterCommit("A4", 4),
					builder.RegisterCommit("A5", 8),
					builder.RegisterCommit("A6", 8),
					builder.RegisterCommit("A7", 8),
					builder.RegisterCommit("A8", 8),
				}
				b := []ifaces.Column{
					builder.RegisterCommit("B1", 4),
					builder.RegisterCommit("B2", 4),
					builder.RegisterCommit("B3", 4),
					builder.RegisterCommit("B4", 4),
					builder.RegisterCommit("B5", 8),
					builder.RegisterCommit("B6", 8),
					builder.RegisterCommit("B7", 8),
					builder.RegisterCommit("B8", 8),
				}
				// each fragment has its own permutation, but the permutation is the same for the columns from the same fragment.
				builder.CompiledIOP.InsertFragmentedPermutation(0, "PERM1", [][]ifaces.Column{a[0:2], a[2:4]}, [][]ifaces.Column{b[0:2], b[2:4]})
				builder.CompiledIOP.InsertFragmentedPermutation(0, "PERM2", [][]ifaces.Column{a[4:6], a[6:8]}, [][]ifaces.Column{b[4:6], b[6:8]})
			},

			Prove: func(run *wizard.ProverRuntime) {
				run.AssignColumn("A1", smartvectors.ForTestFromPairs(
					0, 7000000, // second component is 7000000 + first component
					1, 7000001,
					2, 7000002,
					3, 7000003,
				))
				run.AssignColumn("A2", smartvectors.ForTest(1, 2, 3, 4))
				run.AssignColumn("A3", smartvectors.ForTest(2, 3, 4, 5))
				run.AssignColumn("A4", smartvectors.ForTest(3, 4, 5, 6))
				run.AssignColumn("A5", smartvectors.ForTestFromPairs(
					10, 65555510, // second component is 65555500 + first component
					11, 65555511,
					12, 65555512,
					13, 65555513,
					14, 65555514,
					15, 65555515,
					16, 65555516,
					17, 65555517,
				))
				run.AssignColumn("A6", smartvectors.ForTest(11, 12, 13, 14, 15, 16, 17, 18))
				run.AssignColumn("A7", smartvectors.ForTest(12, 13, 14, 15, 16, 17, 18, 19))
				run.AssignColumn("A8", smartvectors.ForTest(13, 14, 15, 16, 17, 18, 19, 110))

				run.AssignColumn("B1", smartvectors.ForTestFromPairs(
					1, 7000001, // second component is 7000000 + first component
					2, 7000002,
					3, 7000003,
					0, 7000000,
				))
				run.AssignColumn("B2", smartvectors.ForTest(2, 3, 4, 1))
				run.AssignColumn("B3", smartvectors.ForTest(2, 4, 3, 5))
				run.AssignColumn("B4", smartvectors.ForTest(3, 5, 4, 6))
				run.AssignColumn("B5", smartvectors.ForTestFromPairs(
					11, 65555511, // second component is 65555500 + first component
					12, 65555512,
					13, 65555513,
					14, 65555514,
					15, 65555515,
					16, 65555516,
					17, 65555517,
					10, 65555510,
				))
				run.AssignColumn("B6", smartvectors.ForTest(12, 13, 14, 15, 16, 17, 18, 11))
				run.AssignColumn("B7", smartvectors.ForTest(13, 14, 15, 16, 17, 18, 19, 12))
				run.AssignColumn("B8", smartvectors.ForTest(14, 15, 16, 17, 18, 19, 110, 13))
			},

			Title:      "2 fragmented multi-column queries using different sizes",
			ShouldPass: true,
		},
		{
			Define: func(builder *wizard.Builder) {
				a := [][]ifaces.Column{
					{builder.RegisterCommit("A1", 4)},
					{builder.RegisterCommit("A2", 4)},
					{builder.RegisterCommit("A3", 4)},
					{builder.RegisterCommit("A4", 4)},
				}
				b := [][]ifaces.Column{
					{builder.RegisterCommit("B1", 16)},
				}
				builder.CompiledIOP.InsertFragmentedPermutation(0, "PERM1", a, b)
			},

			Prove: func(run *wizard.ProverRuntime) {
				run.AssignColumn("A1", smartvectors.ForTest(0, 1, 2, 3))
				run.AssignColumn("A2", smartvectors.ForTest(4, 5, 6, 7))
				run.AssignColumn("A3", smartvectors.ForTestFromPairs(
					8, 333000008, // second component is 333000000 + first component
					9, 333000009,
					10, 333000010,
					11, 333000011,
				))
				run.AssignColumn("A4", smartvectors.ForTest(12, 13, 14, 15))

				run.AssignColumn("B1", smartvectors.ForTestFromPairs(
					0, 0,
					1, 0,
					2, 0,
					3, 0,
					4, 0,
					5, 0,
					6, 0,
					7, 0,
					8, 333000008,
					9, 333000009,
					10, 333000010,
					11, 333000011,
					12, 0,
					13, 0,
					14, 0,
					15, 0,
				))
			},

			Title:      "dissymetric fractional query",
			ShouldPass: true,
		},

		// test cases that should Not Pass
		{
			Define: func(builder *wizard.Builder) {
				a := builder.RegisterCommit("A", 16)
				b := builder.RegisterCommit("B", 16)
				builder.Permutation("PERM", []ifaces.Column{a}, []ifaces.Column{b})
			},

			Prove: func(run *wizard.ProverRuntime) {
				run.AssignColumn("A", smartvectors.ForTestFromPairs(
					0, 78700000, // second component is 78700000 + first component
					2, 78700002,
					2, 78700002,
					3, 78700003,
					4, 78700004,
					5, 78700005,
					6, 78700006,
					7, 78700007,
					8, 78700008,
					9, 78700009,
					10, 78700010,
					11, 78700011,
					12, 78700012,
					13, 78700013,
					14, 78700014,
					15, 78700015,
				))
				run.AssignColumn("B", smartvectors.ForTestFromPairs(
					15, 78700015, // second component is 78700000 + first component
					14, 78700014,
					13, 78700013,
					12, 78700012,
					11, 78700011,
					10, 78700010,
					9, 78700009,
					8, 78700008,
					7, 78700007,
					6, 78700006,
					5, 78700005,
					4, 78700004,
					3, 78700003,
					2, 78700002,
					1, 78700001,
					0, 78700000,
				))
			},

			Title:      "single-column, should not pass",
			ShouldPass: false,
		},
		{
			Define: func(builder *wizard.Builder) {
				a := []ifaces.Column{
					builder.RegisterCommit("A1", 4),
					builder.RegisterCommit("A2", 4),
					builder.RegisterCommit("A3", 4),
					builder.RegisterCommit("A4", 4),
					builder.RegisterCommit("A5", 8),
					builder.RegisterCommit("A6", 8),
					builder.RegisterCommit("A7", 8),
					builder.RegisterCommit("A8", 8),
				}
				b := []ifaces.Column{
					builder.RegisterCommit("B1", 4),
					builder.RegisterCommit("B2", 4),
					builder.RegisterCommit("B3", 4),
					builder.RegisterCommit("B4", 4),
					builder.RegisterCommit("B5", 8),
					builder.RegisterCommit("B6", 8),
					builder.RegisterCommit("B7", 8),
					builder.RegisterCommit("B8", 8),
				}
				builder.CompiledIOP.InsertFragmentedPermutation(0, "PERM1", [][]ifaces.Column{a[0:2], a[2:4]}, [][]ifaces.Column{b[0:2], b[2:4]})
				builder.CompiledIOP.InsertFragmentedPermutation(0, "PERM2", [][]ifaces.Column{a[4:6], a[6:8]}, [][]ifaces.Column{b[4:6], b[6:8]})
			},

			Prove: func(run *wizard.ProverRuntime) {
				run.AssignColumn("A1", smartvectors.ForTestFromPairs(
					0, 55555000, // second component is 55555000 + first component
					1, 55555001,
					2, 55555002,
					3, 55555003,
				))
				run.AssignColumn("A2", smartvectors.ForTest(2, 2, 3, 4))
				run.AssignColumn("A3", smartvectors.ForTest(2, 3, 4, 5))
				run.AssignColumn("A4", smartvectors.ForTest(3, 4, 5, 6))
				run.AssignColumn("A5", smartvectors.ForTest(10, 11, 12, 13, 14, 15, 16, 17))
				run.AssignColumn("A6", smartvectors.ForTest(11, 12, 13, 14, 15, 16, 17, 18))
				run.AssignColumn("A7", smartvectors.ForTest(12, 13, 14, 15, 16, 17, 18, 19))
				run.AssignColumn("A8", smartvectors.ForTest(13, 14, 15, 16, 17, 18, 19, 110))

				run.AssignColumn("B1", smartvectors.ForTestFromPairs(
					1, 55555001, // second component is 55555000 + first component
					2, 55555002,
					3, 55555003,
					0, 55555000,
				))
				run.AssignColumn("B2", smartvectors.ForTest(2, 3, 4, 1))
				run.AssignColumn("B3", smartvectors.ForTest(3, 4, 5, 2))
				run.AssignColumn("B4", smartvectors.ForTest(4, 5, 6, 3))
				run.AssignColumn("B5", smartvectors.ForTest(11, 12, 13, 14, 15, 16, 17, 10))
				run.AssignColumn("B6", smartvectors.ForTest(12, 13, 14, 15, 16, 17, 18, 11))
				run.AssignColumn("B7", smartvectors.ForTest(13, 14, 15, 16, 17, 18, 19, 12))
				run.AssignColumn("B8", smartvectors.ForTest(14, 15, 16, 17, 18, 19, 110, 13))
			},

			Title:      "2 fragmented multi-column queries using different sizes, should not pass",
			ShouldPass: false,
		},

		{
			Define: func(builder *wizard.Builder) {
				a := [][]ifaces.Column{
					{builder.RegisterCommit("A1", 4)},
					{builder.RegisterCommit("A2", 4)},
					{builder.RegisterCommit("A3", 4)},
					{builder.RegisterCommit("A4", 4)},
				}
				b := [][]ifaces.Column{
					{builder.RegisterCommit("B1", 16)},
				}
				builder.CompiledIOP.InsertFragmentedPermutation(0, "PERM1", a, b)
			},

			Prove: func(run *wizard.ProverRuntime) {
				run.AssignColumn("A1", smartvectors.ForTest(1, 1, 2, 3))
				run.AssignColumn("A2", smartvectors.ForTest(4, 5, 6, 7))
				run.AssignColumn("A3", smartvectors.ForTestFromPairs(
					8, 76600008, // second component is 76600000 + first component
					9, 76600009,
					10, 76600010,
					11, 76600011,
				))
				run.AssignColumn("A4", smartvectors.ForTest(12, 13, 14, 15))

				run.AssignColumn("B1", smartvectors.ForTestFromPairs(
					0, 0,
					1, 1,
					2, 2,
					3, 3,
					4, 4,
					5, 5,
					6, 6,
					7, 7,
					8, 76600008,
					9, 76600009,
					10, 76600010,
					11, 76600011,
					12, 0,
					13, 0,
					14, 0,
					15, 0,
				))
			},

			Title:      "dissymetric fractional query, should not pass",
			ShouldPass: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Title, func(t *testing.T) {
			comp := wizard.Compile(testCase.Define, CompileViaGrandProduct, dummy.Compile)
			proof := wizard.Prove(comp, testCase.Prove)
			if err := wizard.Verify(comp, proof); err != nil && testCase.ShouldPass {
				t.Fatalf("verifier did not pass: %v", err.Error())
			}
			if err := wizard.Verify(comp, proof); err == nil && !testCase.ShouldPass {
				t.Fatalf("verifier is passing for a false claim")
			}
		})
	}
}
