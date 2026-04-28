package sumcheck

import (
	"math/bits"

	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
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

