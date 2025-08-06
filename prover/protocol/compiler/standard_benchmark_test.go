package compiler_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/consensys/linea-monorepo/prover/crypto/ringsis"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/vortex"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/internal/testtools"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	sym "github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
)

// StdBenchmarkCase represents a benchmark case for the Arcane and Vortex compilers
type StdBenchmarkCase struct {
	// Name is the name of the benchmark case
	Name string

	Permutations SubModuleParameters
	Lookup       SubModuleParameters
	Projection   SubModuleParameters
	Fibo         SubModuleParameters
}

// SubModuleParameters represents the parameters of a sub-module in [StdBenchmarkCase]
type SubModuleParameters struct {
	// Count specifies the number of instances of the sub-module
	Count int
	// NumCol specifies the number of columns to define for the module if this
	// is useful to specify.
	NumCol int
	// NumRow specifies the number of rows to define for the module
	NumRow int
	// NumRowAux is an optional parameter for module requiring a second set of
	// columns with a different number of rows such as Lookups or Projections
	NumRowAux int
}

var (
	benchCases = []StdBenchmarkCase{
		{
			Name: "minimal",
			Permutations: SubModuleParameters{
				Count:  1,
				NumCol: 1,
				NumRow: 1 << 10,
			},
			Lookup: SubModuleParameters{
				Count:     1,
				NumCol:    1,
				NumRow:    1 << 10,
				NumRowAux: 1 << 10,
			},
			Projection: SubModuleParameters{
				Count:     1,
				NumCol:    1,
				NumRow:    1 << 10,
				NumRowAux: 1 << 10,
			},
			Fibo: SubModuleParameters{
				Count:  1,
				NumRow: 1 << 10,
			},
		},
	}
)

func BenchmarkCompiler(b *testing.B) {
	for _, bc := range benchCases {
		b.Run(bc.Name, func(b *testing.B) {
			benchmarkCompiler(b, bc)
		})
	}
}

func benchmarkCompiler(b *testing.B, sbc StdBenchmarkCase) {

	comp := wizard.Compile(
		sbc.Define,
		compiler.Arcane(
			compiler.WithTargetColSize(1<<18),
			compiler.WithStitcherMinSize(1<<8),
		),
		vortex.Compile(
			2,
			vortex.WithOptionalSISHashingThreshold(64),
			vortex.ForceNumOpenedColumns(256),
			vortex.WithSISParams(&ringsis.StdParams),
		),
	)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = wizard.Prove(comp, sbc.NewAssigner(b))
	}
}

// Define defines the benchmark case
func (sbc *StdBenchmarkCase) Define(b *wizard.Builder) {

	for n := 0; n < sbc.Permutations.Count; n++ {
		definePermutationModule(b.CompiledIOP, n, sbc.Permutations)
	}

	for n := 0; n < sbc.Lookup.Count; n++ {
		defineLookupModule(b.CompiledIOP, n, sbc.Lookup)
	}

	for n := 0; n < sbc.Projection.Count; n++ {
		defineProjectionModule(b.CompiledIOP, n, sbc.Projection)
	}

	for n := 0; n < sbc.Fibo.Count; n++ {
		defineFiboModule(b.CompiledIOP, n, sbc.Fibo)
	}
}

// NewAssigner returns a function assigning the modules of the benchmark case. The
// function also stops the timer during that time.
func (sbc *StdBenchmarkCase) NewAssigner(b *testing.B) func(run *wizard.ProverRuntime) {
	return func(run *wizard.ProverRuntime) {

		b.StopTimer()

		for n := 0; n < sbc.Permutations.Count; n++ {
			assignPermutationModule(run, n, sbc.Permutations)
		}

		for n := 0; n < sbc.Lookup.Count; n++ {
			assignLookupModule(run, n, sbc.Lookup)
		}

		for n := 0; n < sbc.Projection.Count; n++ {
			assignProjectionModule(run, n, sbc.Projection)
		}

		for n := 0; n < sbc.Fibo.Count; n++ {
			assignFiboModule(run, n, sbc.Fibo)
		}

		b.StartTimer()
	}
}

// defineLookupModule adds a lookup module to the benchmark case
func defineLookupModule(comp *wizard.CompiledIOP, index int, params SubModuleParameters) {

	// This is needed because both tables are assigned using [testtools.RandomFromSeed]
	// and if NumRowAux was smaller, then S (parametrized with NumRow) would be
	// would contain more rows than T and would therefore contain entries not
	// contained in S.
	if params.NumRow < params.NumRowAux {
		utils.Panic("Lookup module requires NumRow < NumRowAux, got %++v for index %v", params, index)
	}

	var (
		s = []ifaces.Column{}
		t = []ifaces.Column{}
	)

	for i := 0; i < params.NumCol; i++ {
		si := comp.InsertCommit(0, formatName[ifaces.ColID]("Lookup", index, "S", i), params.NumRow)
		ti := comp.InsertCommit(0, formatName[ifaces.ColID]("Lookup", index, "T", i), params.NumRowAux)

		s = append(s, si)
		t = append(t, ti)
	}

	comp.InsertInclusion(0, formatName[ifaces.QueryID]("Lookup", index, "Query"), t, s)
}

// assignLookupModule assigns random values to the columns of a lookup module using
// [testtools.RandomFromSeed].
func assignLookupModule(run *wizard.ProverRuntime, index int, params SubModuleParameters) {

	ti := testtools.RandomFromSeed(params.NumRowAux, int64(index))
	si := testtools.RandomFromSeed(params.NumRow, int64(index))

	for i := 0; i < params.NumCol; i++ {
		run.AssignColumn(formatName[ifaces.ColID]("Lookup", index, "S", i), si)
		run.AssignColumn(formatName[ifaces.ColID]("Lookup", index, "T", i), ti)
	}
}

// definePermutationModule creates a permutation module with the requested number
// rows and columns on each side of the permutation.
func definePermutationModule(comp *wizard.CompiledIOP, index int, params SubModuleParameters) {

	var (
		a = []ifaces.Column{}
		b = []ifaces.Column{}
	)

	for i := 0; i < params.NumCol; i++ {
		ai := comp.InsertCommit(0, formatName[ifaces.ColID]("Permutation", index, "A", i), params.NumRow)
		bi := comp.InsertCommit(0, formatName[ifaces.ColID]("Permutation", index, "B", i), params.NumRow)

		a = append(a, ai)
		b = append(b, bi)
	}

	comp.InsertPermutation(0, formatName[ifaces.QueryID]("Permutation", index, "Query"), a, b)
}

// assignPermutationModule assigns random values to the columns of a permutation module using
// [testtools.RandomFromSeed] for a and the same reverted vector for b.
func assignPermutationModule(run *wizard.ProverRuntime, index int, params SubModuleParameters) {

	ai := testtools.RandomFromSeed(params.NumRow, int64(index))
	bi := testtools.Reverse(ai)

	for i := 0; i < params.NumCol; i++ {
		run.AssignColumn(formatName[ifaces.ColID]("Permutation", index, "A", i), ai)
		run.AssignColumn(formatName[ifaces.ColID]("Permutation", index, "B", i), bi)
	}
}

// defineProjectionModule adds a projection module to the benchmark case.
func defineProjectionModule(comp *wizard.CompiledIOP, index int, params SubModuleParameters) {

	var (
		a       = []ifaces.Column{}
		b       = []ifaces.Column{}
		aFilter = comp.InsertCommit(0, formatName[ifaces.ColID]("Projection", index, "AFilter"), params.NumRow)
		bFilter = comp.InsertCommit(0, formatName[ifaces.ColID]("Projection", index, "BFilter"), params.NumRowAux)
	)

	for i := 0; i < params.NumCol; i++ {
		ai := comp.InsertCommit(0, formatName[ifaces.ColID]("Projection", index, "A", i), params.NumRow)
		bi := comp.InsertCommit(0, formatName[ifaces.ColID]("Projection", index, "B", i), params.NumRowAux)

		a = append(a, ai)
		b = append(b, bi)
	}

	comp.InsertProjection(
		formatName[ifaces.QueryID]("Projection", index, "Query"),
		query.ProjectionInput{
			ColumnA: a,
			ColumnB: b,
			FilterA: aFilter,
			FilterB: bFilter,
		})
}

// assignProjectionModule assigns random values to the columns of a projection
// using [testtools.RandomFromSeed]. The large side of the projection is zero-padded
// to match the length of the small side and make the projection work.
func assignProjectionModule(run *wizard.ProverRuntime, index int, params SubModuleParameters) {

	var (
		sizeSmall        = min(params.NumRow, params.NumRowAux)
		sizeLarge        = max(params.NumRow, params.NumRowAux)
		valueSmall       = testtools.RandomFromSeed(sizeSmall, int64(index))
		valueLarge       = smartvectors.RightZeroPadded(valueSmall.IntoRegVecSaveAlloc(), sizeLarge)
		filterSmall      = smartvectors.NewConstant(field.One(), sizeSmall)
		filterLarge      = smartvectors.RightZeroPadded(filterSmall.IntoRegVecSaveAlloc(), sizeLarge)
		ai, bi           smartvectors.SmartVector
		aFilter, bFilter smartvectors.SmartVector
	)

	if params.NumRow < params.NumRowAux {
		ai = valueSmall
		bi = valueLarge
		aFilter = filterSmall
		bFilter = filterLarge
	} else {
		ai = valueLarge
		bi = valueSmall
		aFilter = filterLarge
		bFilter = filterSmall
	}

	run.AssignColumn(formatName[ifaces.ColID]("Projection", index, "AFilter"), aFilter)
	run.AssignColumn(formatName[ifaces.ColID]("Projection", index, "BFilter"), bFilter)

	for i := 0; i < params.NumCol; i++ {
		run.AssignColumn(formatName[ifaces.ColID]("Projection", index, "A", i), ai)
		run.AssignColumn(formatName[ifaces.ColID]("Projection", index, "B", i), bi)
	}
}

// defineFibo creates a single column constrained to be the fibonacci sequence
// starting at 1, 1
func defineFiboModule(comp *wizard.CompiledIOP, index int, params SubModuleParameters) {

	a := comp.InsertCommit(0, formatName[ifaces.ColID]("Fibo", index), params.NumRow)

	comp.InsertGlobal(0, formatName[ifaces.QueryID]("Fibo", index, "Global"), sym.Sub(
		a,
		column.Shift(a, -1),
		column.Shift(a, -2),
	))

	comp.InsertLocal(
		0,
		formatName[ifaces.QueryID]("Fibo", index, "Local_0"),
		sym.Sub(a, 1),
	)

	comp.InsertLocal(
		0,
		formatName[ifaces.QueryID]("Fibo", index, "Local_1"),
		sym.Sub(column.Shift(a, 1), 1),
	)
}

func assignFiboModule(run *wizard.ProverRuntime, index int, params SubModuleParameters) {

	fibo := make([]field.Element, 0, params.NumRow)
	fibo = append(fibo, field.One(), field.One())

	for i := 2; i < params.NumRow; i++ {
		var next field.Element
		next.Add(&fibo[i-1], &fibo[i-2])
		fibo = append(fibo, next)
	}

	run.AssignColumn(
		formatName[ifaces.ColID]("Fibo", index),
		smartvectors.NewRegular(fibo),
	)
}

func formatName[T ~string](args ...any) T {
	argsStr := []string{"BENCHMARK"}
	for _, arg := range args {
		argStr := fmt.Sprintf("%v", arg)
		argStr = strings.ToUpper(argStr)
		argsStr = append(argsStr, argStr)
	}
	return T(strings.Join(argsStr, "_"))
}
