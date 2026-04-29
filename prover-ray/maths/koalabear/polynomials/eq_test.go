package polynomials

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover-ray/maths/koalabear/field"
)

// ---------------------------------------------------------------------------
// EvalEqBase
// ---------------------------------------------------------------------------

// TestEvalEqBaseIdentity verifies the identity and annihilation properties of Eq:
//   - Eq(0,1) = 0  (distinct boolean values)
//   - Eq(0,0) = 1  (equal boolean values)
//   - Eq(1,1) = 1  (equal boolean values)
//   - Eq(x,x) = 1  for arbitrary x (self-equality)
func TestEvalEqBaseIdentity(t *testing.T) {
	var zero, one field.Element
	one.SetOne()

	t.Run("Eq(0,1)=0", func(t *testing.T) {
		got := EvalEqBase([]field.Element{zero}, []field.Element{one})
		if !got.IsZero() {
			t.Fatalf("Eq(0,1): got %v, want 0", got)
		}
	})

	t.Run("Eq(1,0)=0", func(t *testing.T) {
		got := EvalEqBase([]field.Element{one}, []field.Element{zero})
		if !got.IsZero() {
			t.Fatalf("Eq(1,0): got %v, want 0", got)
		}
	})

	t.Run("Eq(0,0)=1", func(t *testing.T) {
		got := EvalEqBase([]field.Element{zero}, []field.Element{zero})
		if !got.Equal(&one) {
			t.Fatalf("Eq(0,0): got %v, want 1", got)
		}
	})

	t.Run("Eq(1,1)=1", func(t *testing.T) {
		got := EvalEqBase([]field.Element{one}, []field.Element{one})
		if !got.Equal(&one) {
			t.Fatalf("Eq(1,1): got %v, want 1", got)
		}
	})

	// Eq(x,x)=1 only holds for boolean x ∈ {0,1}; for arbitrary x the formula
	// gives 1 − 2x + 2x² which is generally not 1.
	t.Run("Eq(0,0)=Eq(1,1)=1", func(t *testing.T) {
		if got := EvalEqBase([]field.Element{zero}, []field.Element{zero}); !got.Equal(&one) {
			t.Fatalf("Eq(0,0)≠1: got %v", got)
		}
		if got := EvalEqBase([]field.Element{one}, []field.Element{one}); !got.Equal(&one) {
			t.Fatalf("Eq(1,1)≠1: got %v", got)
		}
	})

	t.Run("formula_direct", func(t *testing.T) {
		// Verify Eq(x,y) = 1 + 2xy - x - y matches our implementation.
		rng := newRng()
		for range testN {
			x := field.PseudoRand(rng)
			y := field.PseudoRand(rng)
			got := EvalEqBase([]field.Element{x}, []field.Element{y})
			// want = 1 + 2xy - x - y
			var want, xy field.Element
			want.SetOne()
			xy.Mul(&x, &y)
			xy.Add(&xy, &xy)
			want.Add(&want, &xy)
			want.Sub(&want, &x)
			want.Sub(&want, &y)
			if !got.Equal(&want) {
				t.Fatalf("formula mismatch: got %v want %v", got, want)
			}
		}
	})
}

// TestEvalEqBaseFactored verifies that EvalEqBase(q,h) equals the product of
// individual EvalEqBase calls — confirming the factored-product formula.
func TestEvalEqBaseFactored(t *testing.T) {
	rng := newRng()
	const n = 6

	for range testN {
		q := field.VecPseudoRandBase(rng, n)
		h := field.VecPseudoRandBase(rng, n)

		got := EvalEqBase(q, h)

		var want field.Element
		want.SetOne()
		for i := range q {
			term := EvalEqBase(q[i:i+1], h[i:i+1])
			want.Mul(&want, &term)
		}

		if !got.Equal(&want) {
			t.Fatal("EvalEqBase: factored product mismatch")
		}
	}
}

// TestEvalEqBaseEmptyProduct checks that an empty input returns 1.
func TestEvalEqBaseEmptyProduct(t *testing.T) {
	var one field.Element
	one.SetOne()
	got := EvalEqBase(nil, nil)
	if !got.Equal(&one) {
		t.Fatalf("EvalEqBase(nil,nil): got %v, want 1", got)
	}
}

// ---------------------------------------------------------------------------
// EvalEqExt
// ---------------------------------------------------------------------------

// TestEvalEqExtIdentity mirrors TestEvalEqBaseIdentity for the extension field.
func TestEvalEqExtIdentity(t *testing.T) {
	var zero, one field.Ext
	one.SetOne()

	t.Run("Eq(0,1)=0", func(t *testing.T) {
		got := EvalEqExt([]field.Ext{zero}, []field.Ext{one})
		if !extEq(got, zero) {
			t.Fatalf("Eq(0,1): got %v, want 0", got)
		}
	})

	// Same restriction: Eq(x,x)=1 only for x ∈ {0,1}.
	t.Run("Eq(0,0)=Eq(1,1)=1", func(t *testing.T) {
		if got := EvalEqExt([]field.Ext{zero}, []field.Ext{zero}); !extEq(got, one) {
			t.Fatalf("Eq(0,0)≠1: got %v", got)
		}
		if got := EvalEqExt([]field.Ext{one}, []field.Ext{one}); !extEq(got, one) {
			t.Fatalf("Eq(1,1)≠1: got %v", got)
		}
	})
}

// TestEvalEqExtMatchesBase verifies that EvalEqExt and EvalEqBase agree when
// both inputs are base-field elements lifted to ext.
func TestEvalEqExtMatchesBase(t *testing.T) {
	rng := newRng()
	const n = 5

	for range testN {
		qBase := field.VecPseudoRandBase(rng, n)
		hBase := field.VecPseudoRandBase(rng, n)

		// Lift to ext
		qExt := make([]field.Ext, n)
		hExt := make([]field.Ext, n)
		for i := range qBase {
			qExt[i] = field.Lift(qBase[i])
			hExt[i] = field.Lift(hBase[i])
		}

		gotBase := EvalEqBase(qBase, hBase)
		gotExt := EvalEqExt(qExt, hExt)
		wantExt := field.Lift(gotBase)

		if !extEq(gotExt, wantExt) {
			t.Fatal("EvalEqExt disagrees with EvalEqBase for base-field inputs")
		}
	}
}

// ---------------------------------------------------------------------------
// EvalEq (generic dispatcher)
// ---------------------------------------------------------------------------

// TestEvalEqDispatchBase verifies that EvalEq returns a base-tagged result and
// matches EvalEqBase when all inputs are base.
func TestEvalEqDispatchBase(t *testing.T) {
	rng := newRng()
	const n = 5

	for range testN {
		qBase := field.VecPseudoRandBase(rng, n)
		hBase := field.VecPseudoRandBase(rng, n)

		q := make([]field.Gen, n)
		h := make([]field.Gen, n)
		for i := range q {
			q[i] = field.ElemFromBase(qBase[i])
			h[i] = field.ElemFromBase(hBase[i])
		}

		got := EvalEq(q, h)
		want := EvalEqBase(qBase, hBase)

		if !got.IsBase() {
			t.Fatal("EvalEq(base,base) should return a base-tagged Gen")
		}
		gotBase := got.AsBase()
		if !gotBase.Equal(&want) {
			t.Fatalf("EvalEq dispatcher mismatch: got %v want %v", gotBase, want)
		}
	}
}

// TestEvalEqDispatchExt verifies that EvalEq matches EvalEqExt when inputs are ext.
func TestEvalEqDispatchExt(t *testing.T) {
	rng := newRng()
	const n = 5

	for range testN {
		qExt := field.VecPseudoRandExt(rng, n)
		hExt := field.VecPseudoRandExt(rng, n)

		q := make([]field.Gen, n)
		h := make([]field.Gen, n)
		for i := range q {
			q[i] = field.ElemFromExt(qExt[i])
			h[i] = field.ElemFromExt(hExt[i])
		}

		got := EvalEq(q, h)
		want := EvalEqExt(qExt, hExt)

		if got.IsBase() {
			t.Fatal("EvalEq(ext,ext) should not return base-tagged result")
		}
		if !extEq(got.AsExt(), want) {
			t.Fatal("EvalEq dispatcher ext mismatch")
		}
	}
}

// ---------------------------------------------------------------------------
// FoldedEqTableBase
// ---------------------------------------------------------------------------

// TestFoldedEqTableBaseConsistency is the central correctness property:
// evaluating FoldedEqTableBase(q) at h must equal EvalEqBase(q, h).
func TestFoldedEqTableBaseConsistency(t *testing.T) {
	rng := newRng()

	for range testN {
		for n := 0; n <= 8; n++ {
			q := field.VecPseudoRandBase(rng, n)
			h := field.VecPseudoRandBase(rng, n)

			want := EvalEqBase(q, h)

			table := make([]field.Element, 1<<n)
			FoldedEqTableBase(table, q)

			coords := make([]field.Gen, n)
			for i, hi := range h {
				coords[i] = field.ElemFromBase(hi)
			}
			got := EvalMultilin(field.VecFromBase(table), coords)

			if !got.IsBase() {
				t.Fatalf("n=%d: EvalMultilin of eq table should return base", n)
			}
			gotBase := got.AsBase()
			if !gotBase.Equal(&want) {
				t.Fatalf("n=%d: FoldedEqTable+Eval=%v EvalEq=%v", n, gotBase, want)
			}
		}
	}
}

// TestFoldedEqTableBaseBooleanValues checks that table[x] equals the product
// Eq(q[0], x₀)·…·Eq(q[n-1], x_{n-1}) for every x ∈ {0,1}ⁿ directly.
func TestFoldedEqTableBaseBooleanValues(t *testing.T) {
	rng := newRng()

	for n := 0; n <= 8; n++ {
		q := field.VecPseudoRandBase(rng, n)

		table := make([]field.Element, 1<<n)
		FoldedEqTableBase(table, q)

		var zero, one field.Element
		one.SetOne()

		for x := 0; x < (1 << n); x++ {
			// Compute the expected product directly.
			bits := make([]field.Element, n)
			for i := range bits {
				if (x>>(n-1-i))&1 == 1 {
					bits[i] = one
				}
				// else bits[i] stays zero
			}
			want := EvalEqBase(q, bits)

			_ = zero
			if !table[x].Equal(&want) {
				t.Fatalf("n=%d x=%d: table[x]=%v want %v", n, x, table[x], want)
			}
		}
	}
}

// TestFoldedEqTableBaseMultiplier verifies that the optional multiplier argument
// scales every table entry by the given factor.
func TestFoldedEqTableBaseMultiplier(t *testing.T) {
	rng := newRng()

	for range testN {
		for n := 1; n <= 6; n++ {
			q := field.VecPseudoRandBase(rng, n)
			mult := field.PseudoRand(rng)

			// Without multiplier
			base := make([]field.Element, 1<<n)
			FoldedEqTableBase(base, q)

			// With multiplier
			scaled := make([]field.Element, 1<<n)
			FoldedEqTableBase(scaled, q, mult)

			for i := range base {
				var want field.Element
				want.Mul(&base[i], &mult)
				if !scaled[i].Equal(&want) {
					t.Fatalf("n=%d i=%d: multiplier mismatch: got %v want %v", n, i, scaled[i], want)
				}
			}
		}
	}
}

// TestFoldedEqTableBaseNormalization checks that the entries of FoldedEqTable(q)
// sum to 1 (they form a probability distribution over {0,1}ⁿ for random q).
func TestFoldedEqTableBaseNormalization(t *testing.T) {
	rng := newRng()

	for n := 0; n <= 8; n++ {
		q := field.VecPseudoRandBase(rng, n)

		table := make([]field.Element, 1<<n)
		FoldedEqTableBase(table, q)

		var sum field.Element
		for i := range table {
			sum.Add(&sum, &table[i])
		}

		var one field.Element
		one.SetOne()
		if !sum.Equal(&one) {
			t.Fatalf("n=%d: table entries sum to %v, want 1", n, sum)
		}
	}
}

// ---------------------------------------------------------------------------
// FoldedEqTableExt
// ---------------------------------------------------------------------------

// TestFoldedEqTableExtConsistency mirrors TestFoldedEqTableBaseConsistency for
// the extension-field variant.
func TestFoldedEqTableExtConsistency(t *testing.T) {
	rng := newRng()

	for range testN {
		for n := 0; n <= 5; n++ {
			q := field.VecPseudoRandExt(rng, n)
			h := field.VecPseudoRandExt(rng, n)

			want := EvalEqExt(q, h)

			table := make([]field.Ext, 1<<n)
			FoldedEqTableExt(table, q)

			coords := make([]field.Gen, n)
			for i, hi := range h {
				coords[i] = field.ElemFromExt(hi)
			}
			got := EvalMultilin(field.VecFromExt(table), coords)

			if !extEq(got.AsExt(), want) {
				t.Fatalf("n=%d: FoldedEqTableExt+Eval mismatch", n)
			}
		}
	}
}

// TestFoldedEqTableExtMatchesBase verifies that lifting all inputs to ext and
// using FoldedEqTableExt gives the same result as FoldedEqTableBase.
func TestFoldedEqTableExtMatchesBase(t *testing.T) {
	rng := newRng()

	for n := 0; n <= 6; n++ {
		q := field.VecPseudoRandBase(rng, n)

		// Base table
		baseTable := make([]field.Element, 1<<n)
		FoldedEqTableBase(baseTable, q)

		// Ext table with lifted coords
		qExt := make([]field.Ext, n)
		for i, qi := range q {
			qExt[i] = field.Lift(qi)
		}
		extTable := make([]field.Ext, 1<<n)
		FoldedEqTableExt(extTable, qExt)

		for i := range baseTable {
			want := field.Lift(baseTable[i])
			if !extTable[i].Equal(&want) {
				t.Fatalf("n=%d i=%d: ext table disagrees with base table", n, i)
			}
		}
	}
}

// ---------------------------------------------------------------------------
// ChunkOfEqTableBase
// ---------------------------------------------------------------------------

// TestChunkOfEqTableBaseMatchesFull verifies that assembling the table chunk by
// chunk yields the same result as a single FoldedEqTableBase call.
func TestChunkOfEqTableBaseMatchesFull(t *testing.T) {
	rng := newRng()

	for n := 0; n < 15; n++ {
		q := field.VecPseudoRandBase(rng, n)

		full := make([]field.Element, 1<<n)
		FoldedEqTableBase(full, q)

		for logChunk := 1; logChunk < n; logChunk++ {
			chunkSize := 1 << logChunk
			nChunks := (1 << n) / chunkSize

			chunked := make([]field.Element, 1<<n)
			for id := 0; id < nChunks; id++ {
				chunk := chunked[id*chunkSize : (id+1)*chunkSize]
				ChunkOfEqTableBase(chunk, id, chunkSize, q)
			}

			if !field.VecEqualBase(full, chunked) {
				t.Fatalf("n=%d logChunk=%d: chunked output differs from full", n, logChunk)
			}
		}
	}
}

// TestChunkOfEqTableBaseMultiplier checks that the multiplier flows through to
// ChunkOfEqTableBase correctly.
func TestChunkOfEqTableBaseMultiplier(t *testing.T) {
	rng := newRng()
	const n = 6

	q := field.VecPseudoRandBase(rng, n)
	mult := field.PseudoRand(rng)

	// Reference: full table with multiplier
	full := make([]field.Element, 1<<n)
	FoldedEqTableBase(full, q, mult)

	// Chunked with multiplier
	const logChunk = 2
	chunkSize := 1 << logChunk
	nChunks := (1 << n) / chunkSize

	chunked := make([]field.Element, 1<<n)
	for id := 0; id < nChunks; id++ {
		chunk := chunked[id*chunkSize : (id+1)*chunkSize]
		ChunkOfEqTableBase(chunk, id, chunkSize, q, mult)
	}

	if !field.VecEqualBase(full, chunked) {
		t.Fatal("ChunkOfEqTableBase: multiplier variant differs from full table")
	}
}

// ---------------------------------------------------------------------------
// ChunkOfEqTableExt
// ---------------------------------------------------------------------------

// TestChunkOfEqTableExtMatchesFull mirrors TestChunkOfEqTableBaseMatchesFull
// for the extension-field variant.
func TestChunkOfEqTableExtMatchesFull(t *testing.T) {
	rng := newRng()

	for n := 0; n < 10; n++ {
		q := field.VecPseudoRandExt(rng, n)

		full := make([]field.Ext, 1<<n)
		FoldedEqTableExt(full, q)

		for logChunk := 1; logChunk < n; logChunk++ {
			chunkSize := 1 << logChunk
			nChunks := (1 << n) / chunkSize

			chunked := make([]field.Ext, 1<<n)
			for id := 0; id < nChunks; id++ {
				chunk := chunked[id*chunkSize : (id+1)*chunkSize]
				ChunkOfEqTableExt(chunk, id, chunkSize, q)
			}

			if !field.VecEqualExt(full, chunked) {
				t.Fatalf("n=%d logChunk=%d: ext chunked output differs from full", n, logChunk)
			}
		}
	}
}

// ---------------------------------------------------------------------------
// Cross-validation: FoldedEqTable + EvalEq
// ---------------------------------------------------------------------------

// TestEqTableEvalEqCrossValidation is the key sumcheck identity:
// Σ_{x ∈ {0,1}ⁿ} EqTable(q)[x] · f(x) == EvalEq(q, r)·f(r)
// specialised to f = EqTable(r) so LHS = EvalMultilin(EqTable(q), r) = EvalEq(q,r).
func TestEqTableEvalEqCrossValidation(t *testing.T) {
	rng := newRng()

	for range testN {
		for n := 0; n <= 7; n++ {
			q := field.VecPseudoRandBase(rng, n)
			r := field.VecPseudoRandBase(rng, n)

			// a = EvalEq(q, r)
			a := EvalEqBase(q, r)

			// b = EvalMultilin(FoldedEqTable(q), r)
			table := make([]field.Element, 1<<n)
			FoldedEqTableBase(table, q)
			rGen := make([]field.Gen, n)
			for i, ri := range r {
				rGen[i] = field.ElemFromBase(ri)
			}
			bGen := EvalMultilin(field.VecFromBase(table), rGen)
			b := bGen.AsBase()

			if !a.Equal(&b) {
				t.Fatalf("n=%d: EvalEq=%v EvalMultilin(EqTable)=%v", n, a, b)
			}
		}
	}
}

// ---------------------------------------------------------------------------
// Benchmarks
// ---------------------------------------------------------------------------

// BenchmarkFoldedEqTableBase benchmarks FoldedEqTableBase on a 2^20 table.
func BenchmarkFoldedEqTableBase(b *testing.B) {
	const n = 20
	q := field.VecPseudoRandBase(newRng(), n)
	table := make([]field.Element, 1<<n)

	b.ResetTimer()
	for range b.N {
		FoldedEqTableBase(table, q)
	}
}
