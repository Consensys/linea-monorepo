package horner

import (
	"math/rand"
	"strconv"
	"testing"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
)

// projectionTestcase represents a test-case for a projection query
type projectionTestcase struct {
	Name             string
	FilterA, FilterB smartvectors.SmartVector
	As, Bs           []smartvectors.SmartVector
	ShouldFail       bool
}

var rng = rand.New(rand.NewSource(0))

var projectionTCs = []projectionTestcase{

	{
		Name:    "positive/selector-full-zeroes",
		FilterA: smartvectors.NewConstant(field.Zero(), 16),
		FilterB: smartvectors.NewConstant(field.Zero(), 8),
		As: []smartvectors.SmartVector{
			smartvectors.PseudoRand(rng, 16),
		},
		Bs: []smartvectors.SmartVector{
			smartvectors.PseudoRand(rng, 8),
		},
	},

	{
		Name:    "positive/counting-values",
		FilterA: onesAt(16, []int{2, 4, 6, 8, 10}),
		FilterB: onesAt(8, []int{1, 2, 3, 4, 5}),
		As: []smartvectors.SmartVector{
			countingAt(16, 0, []int{2, 4, 6, 8, 10}),
		},
		Bs: []smartvectors.SmartVector{
			countingAt(8, 0, []int{1, 2, 3, 4, 5}),
		},
	},
	{
		Name:    "positive/selector-full-zeroes-multicolumn",
		FilterA: smartvectors.NewConstant(field.Zero(), 16),
		FilterB: smartvectors.NewConstant(field.Zero(), 8),
		As: []smartvectors.SmartVector{
			smartvectors.PseudoRand(rng, 16),
			smartvectors.PseudoRand(rng, 16),
		},
		Bs: []smartvectors.SmartVector{
			smartvectors.PseudoRand(rng, 8),
			smartvectors.PseudoRand(rng, 8),
		},
	},

	{
		Name:    "positive/counting-values-multicolumn",
		FilterA: onesAt(16, []int{2, 4, 6, 8, 10}),
		FilterB: onesAt(8, []int{1, 2, 3, 4, 5}),
		As: []smartvectors.SmartVector{
			countingAt(16, 0, []int{2, 4, 6, 8, 10}),
			countingAt(16, 5, []int{2, 4, 6, 8, 10}),
		},
		Bs: []smartvectors.SmartVector{
			countingAt(8, 0, []int{1, 2, 3, 4, 5}),
			countingAt(8, 5, []int{1, 2, 3, 4, 5}),
		},
	},

	{
		Name:    "negative/full-random-with-full-ones-selectors",
		FilterA: smartvectors.NewConstant(field.One(), 16),
		FilterB: smartvectors.NewConstant(field.One(), 8),
		As: []smartvectors.SmartVector{
			smartvectors.PseudoRand(rng, 16),
		},
		Bs: []smartvectors.SmartVector{
			smartvectors.PseudoRand(rng, 8),
		},
		ShouldFail: true,
	},

	{
		Name:    "negative/counting-too-many",
		FilterA: onesAt(16, []int{2, 4, 6, 8, 10}),
		FilterB: onesAt(8, []int{1, 2, 3, 4, 5, 6}),
		As: []smartvectors.SmartVector{
			countingAt(16, 0, []int{2, 4, 6, 8, 10}),
		},
		Bs: []smartvectors.SmartVector{
			countingAt(8, 0, []int{1, 2, 3, 4, 5, 6}),
		},
		ShouldFail: true,
	},

	{
		Name:    "negative/counting-misaligned",
		FilterA: onesAt(16, []int{2, 4, 6, 8, 10}),
		FilterB: onesAt(8, []int{1, 2, 3, 4, 5}),
		As: []smartvectors.SmartVector{
			countingAt(16, 0, []int{2, 4, 6, 8, 10}),
		},
		Bs: []smartvectors.SmartVector{
			countingAt(8, 0, []int{1, 2, 3, 4, 6}),
		},
		ShouldFail: true,
	},
}

// onesAt returns a smartvector for size n, whose values are 1 at the given indices
// and 0 at all other positions.
func onesAt(size int, at []int) smartvectors.SmartVector {
	res := make([]field.Element, size)
	for _, n := range at {
		res[n] = field.One()
	}
	return smartvectors.NewRegular(res)
}

// countingAt returns a smartvector for size n, are starting from "init" and
// incrementing by 1 at the given indices. The indices must be sorted in
// ascending order.
func countingAt(size int, init int, at []int) smartvectors.SmartVector {

	var (
		res      = make([]field.Element, size)
		cursorAt = 0
		one      = field.One()
	)

	for i := range res {

		if i == 0 {
			res[i] = field.NewElement(uint64(init))
		} else {
			res[i] = res[i-1]
		}

		if cursorAt < len(at) && i == at[cursorAt] {
			res[i].Add(&res[i], &one)
			cursorAt++
		}
	}

	return smartvectors.NewRegular(res)
}

// defineTestcaseWithProjection returns a [wizard.DefineFunc] constructing
// columns and a [query.Projection] as specified by the testcase.
func defineTestcaseWithProjection(tc projectionTestcase) wizard.DefineFunc {
	return func(build *wizard.Builder) {

		inp := query.ProjectionInput{
			FilterA: build.RegisterCommit(ifaces.ColID("filterA"), tc.FilterA.Len()),
			FilterB: build.RegisterCommit(ifaces.ColID("filterB"), tc.FilterB.Len()),
			ColumnA: make([]ifaces.Column, len(tc.As)),
			ColumnB: make([]ifaces.Column, len(tc.Bs)),
		}

		for i := range inp.ColumnA {
			inp.ColumnA[i] = build.RegisterCommit(ifaces.ColID("A_"+strconv.Itoa(i)), tc.As[i].Len())
		}

		for i := range inp.ColumnB {
			inp.ColumnB[i] = build.RegisterCommit(ifaces.ColID("B_"+strconv.Itoa(i)), tc.Bs[i].Len())
		}

		build.InsertProjection(
			ifaces.QueryID("PROJECTION"),
			inp,
		)
	}
}

// proverTestcaseWithProjection returns a prover function assigning the
// columns taking place in the [query.Projection] query.
func proverTestcaseWithProjection(tc projectionTestcase) func(run *wizard.ProverRuntime) {

	return func(run *wizard.ProverRuntime) {

		run.AssignColumn(ifaces.ColID("filterA"), tc.FilterA)
		run.AssignColumn(ifaces.ColID("filterB"), tc.FilterB)

		for i := range tc.As {
			run.AssignColumn(ifaces.ColID("A_"+strconv.Itoa(i)), tc.As[i])
		}

		for i := range tc.Bs {
			run.AssignColumn(ifaces.ColID("B_"+strconv.Itoa(i)), tc.Bs[i])
		}
	}
}

func TestProjectionToHorner(t *testing.T) {

	for _, tc := range projectionTCs {

		define, prover := defineTestcaseWithProjection(tc), proverTestcaseWithProjection(tc)

		t.Run(tc.Name+"/with-dummy-compiler", func(t *testing.T) {

			comp := wizard.Compile(define, dummy.Compile)

			if tc.ShouldFail {
				runTestShouldFail(t, comp, prover)
			}

			if !tc.ShouldFail {
				runTestShouldPass(t, comp, prover)
			}
		})

		t.Run(tc.Name+"/with-into-horner-compiler", func(t *testing.T) {

			comp := wizard.Compile(define, ProjectionToHorner, dummy.Compile)

			if tc.ShouldFail {
				runTestShouldFail(t, comp, prover)
			}

			if !tc.ShouldFail {
				runTestShouldPass(t, comp, prover)
			}
		})

		t.Run(tc.Name+"/with-full-projection-compiler", func(t *testing.T) {

			comp := wizard.Compile(define, CompileProjection, dummy.Compile)

			if tc.ShouldFail {
				runTestShouldFail(t, comp, prover)
			}

			if !tc.ShouldFail {
				runTestShouldPass(t, comp, prover)
			}
		})

	}

}

func runTestShouldPass(t *testing.T, comp *wizard.CompiledIOP, prover wizard.ProverStep) {
	proof := wizard.Prove(comp, prover)
	err := wizard.Verify(comp, proof)
	if err != nil {
		t.Errorf("verifier failed: %v", err)
	}
}

func runTestShouldFail(t *testing.T, comp *wizard.CompiledIOP, prover wizard.ProverStep) {

	var (
		verErr, panicErr error
		proof            wizard.Proof
	)

	panicErr = utils.RecoverPanic(func() {
		proof = wizard.Prove(comp, prover)
	})

	if panicErr != nil {
		return
	}

	panicErr = utils.RecoverPanic(func() {
		verErr = wizard.Verify(comp, proof)
	})

	if panicErr == nil && verErr == nil {
		t.Error("test was expected to fail but did not")
	}
}
