package sumcheck

import "github.com/consensys/linea-monorepo/prover/maths/field/fext"

// EvalEq evaluates the multilinear extension of the equality predicate at
// (q, h):
//
//	eq(q, h) = ∏_i ( q_i*h_i + (1-q_i)*(1-h_i) )
//	        = ∏_i ( 1 - q_i - h_i + 2*q_i*h_i )
//
// On boolean inputs eq(q,h) is 1 iff q == h, else 0; this multilinearly
// extends to all of fext^n.
func EvalEq(q, h []fext.Element) fext.Element {
	if len(q) != len(h) {
		panic("EvalEq: q and h must have equal length")
	}
	var res fext.Element
	res.SetOne()
	if len(q) == 0 {
		return res
	}

	var one, prod, sum, term fext.Element
	one.SetOne()
	for i := range q {
		// term = 1 - q_i - h_i + 2*q_i*h_i
		prod.Mul(&q[i], &h[i])
		term.Double(&prod)
		term.Add(&term, &one)
		sum.Add(&q[i], &h[i])
		term.Sub(&term, &sum)
		res.Mul(&res, &term)
	}
	return res
}

// BuildEqTable returns the multilinear polynomial eq(r, *) materialized over
// the boolean hypercube, in MultiLin layout (X_1 = MSB of index). For every
// h ∈ {0,1}^n, the returned table satisfies T[idx(h)] = eq(r, h), where
// idx(h) = (h_1 << (n-1)) | ... | h_n.
//
// Cost: O(2^n) field operations and 2^n storage. For batched sumcheck we keep
// one table per claim and fold it alongside the polynomial each round.
func BuildEqTable(r []fext.Element) MultiLin {
	n := len(r)
	size := 1 << n
	table := make(MultiLin, size)
	if n == 0 {
		table[0].SetOne()
		return table
	}
	table[0].SetOne()

	var one, oneMinus, hi fext.Element
	one.SetOne()
	// At iteration i we bind X_{i+1} := r[i]. The "active" positions before
	// this iteration are j << (n-i) for j ∈ [0, 1<<i), each holding
	// eq(r[:i], (b_1, ..., b_i)) for the corresponding bits. We expand each
	// active position by writing the X_{i+1}=1 child at j1 and rescaling the
	// X_{i+1}=0 child in place.
	for i := 0; i < n; i++ {
		oneMinus.Sub(&one, &r[i])
		for j := 0; j < (1 << i); j++ {
			j0 := j << (n - i)
			j1 := j0 + (1 << (n - 1 - i))
			hi.Mul(&r[i], &table[j0])
			table[j1] = hi
			table[j0].Mul(&oneMinus, &table[j0])
		}
	}
	return table
}
