package sumcheck

import (
	"math/bits"

	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/utils/parallel"
)

// MultiLin is a dense multilinear polynomial in n = log2(len) variables,
// stored as the vector of evaluations on the boolean hypercube.
//
// Indexing convention (matches gnark-crypto's MultiLin): the entry
//
//	m[ b_1 << (n-1) | b_2 << (n-2) | ... | b_n ]
//
// equals P(b_1, ..., b_n). Variable X_1 is the most-significant bit of the
// index, X_n is the least-significant bit. Fold binds X_1 first.
type MultiLin []fext.Element

// NumVars returns log2(len(m)). Panics if len(m) is not a power of two.
func (m MultiLin) NumVars() int {
	if len(m) == 0 || len(m)&(len(m)-1) != 0 {
		panic("MultiLin length must be a power of two")
	}
	return bits.TrailingZeros(uint(len(m)))
}

// Clone returns a deep copy of m.
func (m MultiLin) Clone() MultiLin {
	out := make(MultiLin, len(m))
	copy(out, m)
	return out
}

// Fold partially evaluates m at X_1 = r, mutating it in place to a polynomial
// of one fewer variable (length halves). After Fold,
//
//	m'[i] = (1-r)*m[i] + r*m[i+mid]   for i ∈ [0, mid)
func (m *MultiLin) Fold(r fext.Element) {
	mid := len(*m) / 2
	if mid == 0 {
		panic("Fold called on a constant multilinear")
	}
	bottom := (*m)[:mid]
	top := (*m)[mid:]
	var t fext.Element
	for i := 0; i < mid; i++ {
		t.Sub(&top[i], &bottom[i])
		t.Mul(&t, &r)
		bottom[i].Add(&bottom[i], &t)
	}
	*m = (*m)[:mid]
}

// ParallelFold is like Fold but parallelises the inner loop using
// parallel.Execute when mid >= 1<<13 (8192 elements). For smaller sizes it
// falls through to the sequential Fold. Call it only from a context that owns
// the full CPU budget (e.g. K == 1 in the outer parallel.Execute).
func (m *MultiLin) ParallelFold(r fext.Element) {
	mid := len(*m) / 2
	if mid == 0 {
		panic("Fold called on a constant multilinear")
	}
	const threshold = 1 << 13
	if mid < threshold {
		m.Fold(r)
		return
	}
	bottom := (*m)[:mid]
	top := (*m)[mid:]
	parallel.Execute(mid, func(start, stop int) {
		var t fext.Element
		for i := start; i < stop; i++ {
			t.Sub(&top[i], &bottom[i])
			t.Mul(&t, &r)
			bottom[i].Add(&bottom[i], &t)
		}
	})
	*m = (*m)[:mid]
}

// FoldFromBase fuses the base-field→fext conversion with the first fold step.
// It takes a base-field slice src (length 2^n) and an extension-field challenge
// r, and returns a MultiLin of length 2^(n-1) whose i-th entry is
//
//	src[i] + r*(src[i+mid] - src[i])
//
// This avoids the full-size fext allocation that IntoRegVecSaveAllocExt would
// produce, cutting peak memory in half for large polynomials.
func FoldFromBase(src []field.Element, r fext.Element) MultiLin {
	n := len(src)
	if n < 2 || n&(n-1) != 0 {
		panic("FoldFromBase: src length must be a power of two >= 2")
	}
	mid := n / 2
	out := make(MultiLin, mid)
	for i := 0; i < mid; i++ {
		var diff field.Element
		diff.Sub(&src[i+mid], &src[i])
		var t fext.Element
		t.MulByElement(&r, &diff)
		fext.SetFromBase(&out[i], &src[i])
		out[i].Add(&out[i], &t)
	}
	return out
}

// ParallelFoldFromBase is like [FoldFromBase] but parallelises the inner loop
// when mid >= 1<<13. Call it from a context that owns the full CPU budget.
func ParallelFoldFromBase(src []field.Element, r fext.Element) MultiLin {
	n := len(src)
	if n < 2 || n&(n-1) != 0 {
		panic("ParallelFoldFromBase: src length must be a power of two >= 2")
	}
	mid := n / 2
	out := make(MultiLin, mid)
	const threshold = 1 << 13
	if mid < threshold {
		return FoldFromBase(src, r)
	}
	parallel.Execute(mid, func(start, stop int) {
		for i := start; i < stop; i++ {
			var diff field.Element
			diff.Sub(&src[i+mid], &src[i])
			var t fext.Element
			t.MulByElement(&r, &diff)
			fext.SetFromBase(&out[i], &src[i])
			out[i].Add(&out[i], &t)
		}
	})
	return out
}

// Evaluate computes m(point) where len(point) == m.NumVars(). It does not
// mutate m. Folds variables in order: point[0] binds X_1, point[1] binds X_2,
// and so on.
func (m MultiLin) Evaluate(point []fext.Element) fext.Element {
	if len(point) != m.NumVars() {
		panic("Evaluate: point length must equal NumVars")
	}
	work := m.Clone()
	for _, r := range point {
		work.Fold(r)
	}
	return work[0]
}

