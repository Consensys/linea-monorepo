package polynomials

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover-ray/maths/koalabear/field"
)

// ---------------------------------------------------------------------------
// FoldInto
// ---------------------------------------------------------------------------

// TestFoldBaseKnownValue verifies the concrete fold result
// [0, 1, 2, 3] folded at r=5 → [10, 11]:
//   - res[0] = 0 + 5·(2−0) = 10
//   - res[1] = 1 + 5·(3−1) = 11
func TestFoldBaseKnownValue(t *testing.T) {
	table := field.VecFromBase(field.VecFromInts(0, 1, 2, 3))
	resSlice := make([]field.Element, 2)
	res := field.VecFromBase(resSlice)

	var r field.Element
	r.SetUint64(5)

	FoldInto(res, table, field.ElemFromBase(r))

	want := field.VecFromInts(10, 11)
	if !field.VecEqualBase(res.AsBase(), want) {
		t.Fatalf("FoldBaseKnownValue: got %v, want %v", field.VecPrettifyBase(res.AsBase()), field.VecPrettifyBase(want))
	}
}

// TestFoldTypesAgree checks that all four typed fold kernels produce the same
// result when inputs happen to be base-field elements (lifted to ext as needed).
func TestFoldTypesAgree(t *testing.T) {
	rng := newRng()
	const size = 16

	for range testN {
		tableBase := field.VecPseudoRandBase(rng, size)
		r := field.PseudoRand(rng)
		rExt := field.Lift(r)
		tableExt := make([]field.Ext, size)
		for i, e := range tableBase {
			tableExt[i] = field.Lift(e)
		}

		// Reference: base/base/base
		resRef := make([]field.Element, size/2)
		foldBaseBaseInto(resRef, tableBase, r)

		// base table, ext r → ext result
		res1 := make([]field.Ext, size/2)
		foldBaseExtInto(res1, tableBase, rExt)

		// ext table, base r → ext result
		res2 := make([]field.Ext, size/2)
		foldExtBaseInto(res2, tableExt, r)

		// ext table, ext r → ext result
		res3 := make([]field.Ext, size/2)
		foldExtExtInto(res3, tableExt, rExt)

		for i := range resRef {
			want := field.Lift(resRef[i])
			if !res1[i].Equal(&want) {
				t.Fatalf("i=%d foldBaseExtInto mismatch: got %v want %v", i, res1[i], want)
			}
			if !res2[i].Equal(&want) {
				t.Fatalf("i=%d foldExtBaseInto mismatch: got %v want %v", i, res2[i], want)
			}
			if !res3[i].Equal(&want) {
				t.Fatalf("i=%d foldExtExtInto mismatch: got %v want %v", i, res3[i], want)
			}
		}
	}
}

// TestFoldIntoDispatch verifies the public FoldInto dispatcher against the
// base-only typed kernel.
func TestFoldIntoDispatch(t *testing.T) {
	rng := newRng()
	const size = 32

	for range testN {
		tableBase := field.VecPseudoRandBase(rng, size)
		r := field.PseudoRand(rng)

		// via dispatcher (base/base/base)
		resDisp := make([]field.Element, size/2)
		FoldInto(field.VecFromBase(resDisp), field.VecFromBase(tableBase), field.ElemFromBase(r))

		// via typed kernel
		resRef := make([]field.Element, size/2)
		foldBaseBaseInto(resRef, tableBase, r)

		if !field.VecEqualBase(resDisp, resRef) {
			t.Fatal("FoldInto dispatch (base/base/base) disagrees with typed kernel")
		}
	}
}

// TestFoldIntoAliasingSafe checks that FoldInto is correct when res aliases
// the first half of the same backing array as table.
func TestFoldIntoAliasingSafe(t *testing.T) {
	rng := newRng()
	const size = 16

	for range testN {
		backing := field.VecPseudoRandBase(rng, size)

		// Expected result using a separate copy
		tableCopy := make([]field.Element, size)
		copy(tableCopy, backing)
		resExpected := make([]field.Element, size/2)
		r := field.PseudoRand(rng)
		foldBaseBaseInto(resExpected, tableCopy, r)

		// Alias: res = backing[:size/2], table = backing
		FoldInto(
			field.VecFromBase(backing[:size/2]),
			field.VecFromBase(backing),
			field.ElemFromBase(r),
		)

		if !field.VecEqualBase(backing[:size/2], resExpected) {
			t.Fatal("FoldInto aliasing: result differs from non-aliased version")
		}
	}
}

// ---------------------------------------------------------------------------
// EvalMultilin
// ---------------------------------------------------------------------------

// TestEvalMultilinBooleanHypercube evaluates at every boolean point and
// checks that the result equals the corresponding table entry.
func TestEvalMultilinBooleanHypercube(t *testing.T) {
	rng := newRng()

	for n := 0; n <= 6; n++ {
		size := 1 << n
		tableBase := field.VecPseudoRandBase(rng, size)

		for x := 0; x < size; x++ {
			// coords[i] selects bit (n−1−i) of x (MSB first, matching FoldInto convention).
			coords := make([]field.Gen, n)
			for i := range coords {
				if (x>>(n-1-i))&1 == 1 {
					coords[i] = field.ElemOne()
				} else {
					coords[i] = field.ElemZero()
				}
			}

			got := EvalMultilin(field.VecFromBase(tableBase), coords)
			want := tableBase[x]

			if !got.IsBase() {
				t.Fatalf("n=%d x=%d: result should be tagged base", n, x)
			}
			gotBase := got.AsBase()
			if !gotBase.Equal(&want) {
				t.Fatalf("n=%d x=%d: got %v, want %v", n, x, gotBase, want)
			}
		}
	}
}

// TestEvalMultilinMatchesManualFolds checks that EvalMultilin agrees with
// manually applying foldBaseBaseInto step by step.
func TestEvalMultilinMatchesManualFolds(t *testing.T) {
	rng := newRng()

	for range testN {
		for n := 1; n <= 8; n++ {
			size := 1 << n
			tableBase := field.VecPseudoRandBase(rng, size)
			coords := make([]field.Gen, n)
			for i := range coords {
				coords[i] = field.ElemFromBase(field.PseudoRand(rng))
			}

			// Manual fold in a working copy
			work := make([]field.Element, size)
			copy(work, tableBase)
			for _, r := range coords {
				mid := len(work) / 2
				foldBaseBaseInto(work[:mid], work, r.AsBase())
				work = work[:mid]
			}
			want := work[0]

			got := EvalMultilin(field.VecFromBase(tableBase), coords)

			if !got.IsBase() {
				t.Fatalf("n=%d: result should be tagged base", n)
			}
			gotBase := got.AsBase()
			if !gotBase.Equal(&want) {
				t.Fatalf("n=%d: EvalMultilin=%v manual=%v", n, gotBase, want)
			}
		}
	}
}

// TestEvalMultilinExtCoords verifies EvalMultilin with extension-field coordinates.
func TestEvalMultilinExtCoords(t *testing.T) {
	rng := newRng()

	for range testN {
		for n := 1; n <= 5; n++ {
			size := 1 << n
			tableBase := field.VecPseudoRandBase(rng, size)
			coords := make([]field.Gen, n)
			for i := range coords {
				coords[i] = field.ElemFromExt(randExt(rng))
			}

			// Reference: manual folds in ext
			work := make([]field.Ext, size)
			for i, e := range tableBase {
				work[i] = field.Lift(e)
			}
			for _, r := range coords {
				mid := len(work) / 2
				foldExtExtInto(work[:mid], work, r.AsExt())
				work = work[:mid]
			}
			want := work[0]

			got := EvalMultilin(field.VecFromBase(tableBase), coords)

			if got.IsBase() {
				t.Fatalf("n=%d: result should not be tagged base with ext coords", n)
			}
			if !extEq(got.AsExt(), want) {
				t.Fatalf("n=%d: EvalMultilin ext coords mismatch", n)
			}
		}
	}
}

// TestEvalMultilinExtTable verifies EvalMultilin with an extension-field table.
func TestEvalMultilinExtTable(t *testing.T) {
	rng := newRng()

	for range testN {
		for n := 1; n <= 5; n++ {
			size := 1 << n
			tableExt := field.VecPseudoRandExt(rng, size)
			coords := make([]field.Gen, n)
			for i := range coords {
				coords[i] = field.ElemFromBase(field.PseudoRand(rng))
			}

			// Reference: manual folds
			work := make([]field.Ext, size)
			copy(work, tableExt)
			for _, r := range coords {
				mid := len(work) / 2
				foldExtBaseInto(work[:mid], work, r.AsBase())
				work = work[:mid]
			}
			want := work[0]

			got := EvalMultilin(field.VecFromExt(tableExt), coords)

			if got.IsBase() {
				t.Fatalf("n=%d: result should not be tagged base with ext table", n)
			}
			if !extEq(got.AsExt(), want) {
				t.Fatalf("n=%d: EvalMultilin ext table mismatch", n)
			}
		}
	}
}

// TestEvalMultilinPreservesTable checks that EvalMultilin never modifies the
// input table.
func TestEvalMultilinPreservesTable(t *testing.T) {
	rng := newRng()
	const n = 5
	const size = 1 << n

	tableBase := field.VecPseudoRandBase(rng, size)
	snapshot := make([]field.Element, size)
	copy(snapshot, tableBase)

	coords := make([]field.Gen, n)
	for i := range coords {
		coords[i] = field.ElemFromBase(field.PseudoRand(rng))
	}

	EvalMultilin(field.VecFromBase(tableBase), coords)

	if !field.VecEqualBase(tableBase, snapshot) {
		t.Fatal("EvalMultilin modified the input table")
	}
}

// TestEvalMultilinDegenerate covers edge cases: n=0 and single-element tables.
func TestEvalMultilinDegenerate(t *testing.T) {
	var val field.Element
	val.SetUint64(42)

	t.Run("n=0_base", func(t *testing.T) {
		table := field.VecFromBase([]field.Element{val})
		got := EvalMultilin(table, nil)
		if !got.IsBase() {
			t.Fatal("n=0 base should return base")
		}
		gotBase := got.AsBase()
		if !gotBase.Equal(&val) {
			t.Fatalf("n=0: got %v want %v", gotBase, val)
		}
	})

	t.Run("n=0_ext", func(t *testing.T) {
		valExt := field.Lift(val)
		table := field.VecFromExt([]field.Ext{valExt})
		got := EvalMultilin(table, nil)
		if got.IsBase() {
			t.Fatal("n=0 ext should return ext")
		}
		if !extEq(got.AsExt(), valExt) {
			t.Fatalf("n=0 ext: got %v want %v", got.AsExt(), valExt)
		}
	})
}

// BenchmarkFoldBaseBase benchmarks the all-base fold kernel on a 2^25 table.
func BenchmarkFoldBaseBase(b *testing.B) {
	const logN = 25
	table := field.VecPseudoRandBase(newRng(), 1<<logN)
	var r field.Element
	r.SetUint64(5)

	b.ResetTimer()
	for range b.N {
		work := make([]field.Element, len(table))
		copy(work, table)
		mid := len(work) / 2
		foldBaseBaseInto(work[:mid], work, r)
	}
}

// BenchmarkEvalMultilinBase benchmarks EvalMultilin on a 2^20 base-field table.
func BenchmarkEvalMultilinBase(b *testing.B) {
	const n = 20
	rng := newRng()
	table := field.VecFromBase(field.VecPseudoRandBase(rng, 1<<n))
	coords := make([]field.Gen, n)
	for i := range coords {
		coords[i] = field.ElemFromBase(field.PseudoRand(rng))
	}

	b.ResetTimer()
	for range b.N {
		EvalMultilin(table, coords)
	}
}
