package codegen

import (
	"bytes"
	"strings"
	"testing"

	"github.com/consensys/linea-monorepo/prover-ray/maths/koalabear/field"
	"github.com/consensys/linea-monorepo/prover-ray/wiop"
	"github.com/consensys/linea-monorepo/prover-ray/wiop/compilers/logderivativesum"
	"github.com/consensys/linea-monorepo/prover-ray/wiop/compilers/lookuptologderivsum"
)

// newSingleFractionLDS builds a size-4 module with a single LogDerivativeSum
// query Σ col[i]/1, compiled to one Z column. It returns the system after the
// logderivativesum compiler pass has registered the verifier action.
func newSingleFractionLDS(t *testing.T) *wiop.System {
	t.Helper()
	sys := wiop.NewSystemf("ld-codegen")
	r0 := sys.NewRound()
	sys.NewRound() // result round, following the column round
	mod := sys.NewSizedModule(sys.Context.Childf("mod"), 4, wiop.PaddingDirectionNone)
	col := mod.NewColumn(sys.Context.Childf("col"), wiop.VisibilityOracle, r0)
	one := wiop.NewConstantVector(mod, field.NewFromString("1"))
	sys.NewLogDerivativeSum(sys.Context.Childf("ld"), []wiop.Fraction{
		{Numerator: col.View(), Denominator: one},
	})
	logderivativesum.Compile(sys)
	return sys
}

func TestBuildLogDerivSystemExtractsQuery(t *testing.T) {
	sys := newSingleFractionLDS(t)

	ld, err := BuildLogDerivSystem(sys)
	if err != nil {
		t.Fatalf("BuildLogDerivSystem() error = %v", err)
	}
	if len(ld.Queries) != 1 {
		t.Fatalf("expected exactly one query, got %d", len(ld.Queries))
	}
	q := ld.Queries[0]
	if len(q.ZFinals) != 1 {
		t.Fatalf("a single fraction packs into one Z column, got %d z-finals", len(q.ZFinals))
	}
	if q.ResultIsZero {
		t.Fatalf("a plain LogDerivativeSum query must not be marked result-is-zero")
	}
	// The endpoint opening and the Result cell both live in the result round.
	if q.ZFinals[0].Round != q.Result.Round {
		t.Fatalf("z-final round %d and result round %d should match", q.ZFinals[0].Round, q.Result.Round)
	}
}

func TestWriteLogDerivSystemZigRendersQuery(t *testing.T) {
	sys := newSingleFractionLDS(t)
	ld, err := BuildLogDerivSystem(sys)
	if err != nil {
		t.Fatalf("BuildLogDerivSystem() error = %v", err)
	}

	var out bytes.Buffer
	if err := WriteLogDerivSystemZig(&out, 0, ld); err != nil {
		t.Fatalf("WriteLogDerivSystemZig() error = %v", err)
	}
	got := out.String()
	for _, want := range []string{
		"const logderivativesum = @import",
		"system_0_logderiv_query_0_zfinals = [_]logderivativesum.ScalarRef{",
		"system_0_logderiv_queries = [_]logderivativesum.Query{",
		".result_is_zero = false",
		"const system_0_logderiv = logderivativesum.System{ .queries = &system_0_logderiv_queries };",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Zig missing %q:\n%s", want, got)
		}
	}
}

func TestBuildLogDerivSystemMarksLookupResultZero(t *testing.T) {
	// A single-column inclusion lookup reduces to a LogDerivativeSum whose
	// Result must be zero; the extracted query must carry ResultIsZero.
	sys := wiop.NewSystemf("lk-codegen")
	r0 := sys.NewRound()
	modT := sys.NewSizedModule(sys.Context.Childf("modT"), 4, wiop.PaddingDirectionNone)
	modS := sys.NewSizedModule(sys.Context.Childf("modS"), 4, wiop.PaddingDirectionNone)
	colT := modT.NewColumn(sys.Context.Childf("T"), wiop.VisibilityOracle, r0)
	colS := modS.NewColumn(sys.Context.Childf("S"), wiop.VisibilityOracle, r0)
	sys.NewInclusion(
		sys.Context.Childf("inc"),
		[]wiop.Table{wiop.NewTable(colS.View())},
		[]wiop.Table{wiop.NewTable(colT.View())},
	)

	lookuptologderivsum.Compile(sys)
	logderivativesum.Compile(sys)

	ld, err := BuildLogDerivSystem(sys)
	if err != nil {
		t.Fatalf("BuildLogDerivSystem() error = %v", err)
	}
	if len(ld.Queries) != 1 {
		t.Fatalf("the lookup reduces to exactly one LogDerivativeSum query, got %d", len(ld.Queries))
	}
	if !ld.Queries[0].ResultIsZero {
		t.Fatalf("a lookup-reduced query must be marked result-is-zero")
	}
}

func TestBuildLogDerivSystemNoQueries(t *testing.T) {
	sys := wiop.NewSystemf("ld-none")
	sys.NewRound()
	sys.NewSizedModule(sys.Context.Childf("mod"), 4, wiop.PaddingDirectionNone)

	ld, err := BuildLogDerivSystem(sys)
	if err != nil {
		t.Fatalf("BuildLogDerivSystem() error = %v", err)
	}
	if len(ld.Queries) != 0 {
		t.Fatalf("a system without LogDerivativeSum queries must yield no queries, got %d", len(ld.Queries))
	}
}
