package polynomials

import (
	"fmt"
	"math/bits"

	"github.com/consensys/linea-monorepo/prover-ray/maths/koalabear/field"
)

// EvalEq computes the product Π_{i} Eq(q[i], h[i]) using the equality polynomial
//
//	Eq(x, y) = 1 − x − y + 2xy = (1−x)(1−y) + xy,
//
// which interpolates the equality function on {0,1}: Eq(a,b) = 1 iff a=b ∈ {0,1}.
//
// q and h must have the same length. The result is tagged base iff every element
// of q and h is base-field. An empty product returns the multiplicative identity.
func EvalEq(q, h []field.Gen) field.Gen {
	if len(q) != len(h) {
		panic(fmt.Sprintf("polynomial: EvalEq: len(q)=%d != len(h)=%d", len(q), len(h)))
	}
	res := field.ElemOne()
	one := field.ElemOne()
	for i := range q {
		// nxt = 1 + 2·q[i]·h[i] − q[i] − h[i]
		qihi := q[i].Mul(h[i])
		nxt := one.Add(qihi).Add(qihi).Sub(q[i]).Sub(h[i])
		res = res.Mul(nxt)
	}
	return res
}

// EvalEqBase computes Π_{i} Eq(q[i], h[i]) over the base field.
// q and h must have the same length. An empty product returns 1.
func EvalEqBase(q, h []field.Element) field.Element {
	if len(q) != len(h) {
		panic(fmt.Sprintf("polynomial: EvalEqBase: len(q)=%d != len(h)=%d", len(q), len(h)))
	}
	var res, one field.Element
	one.SetOne()
	res.SetOne()
	for i := range q {
		var nxt, xy, sum field.Element
		xy.Mul(&q[i], &h[i])
		xy.Add(&xy, &xy)   // 2·q[i]·h[i]
		nxt.Add(&one, &xy) // 1 + 2·q[i]·h[i]
		sum.Add(&q[i], &h[i])
		nxt.Sub(&nxt, &sum) // 1 + 2·q[i]·h[i] − q[i] − h[i]
		res.Mul(&res, &nxt)
	}
	return res
}

// EvalEqExt computes Π_{i} Eq(q[i], h[i]) over the extension field.
// q and h must have the same length. An empty product returns 1.
func EvalEqExt(q, h []field.Ext) field.Ext {
	if len(q) != len(h) {
		panic(fmt.Sprintf("polynomial: EvalEqExt: len(q)=%d != len(h)=%d", len(q), len(h)))
	}
	var res, one field.Ext
	one.SetOne()
	res.SetOne()
	for i := range q {
		var nxt, xy, sum field.Ext
		xy.Mul(&q[i], &h[i])
		xy.Add(&xy, &xy)
		nxt.Add(&one, &xy)
		sum.Add(&q[i], &h[i])
		nxt.Sub(&nxt, &sum)
		res.Mul(&res, &nxt)
	}
	return res
}

// FoldedEqTableBase fills table with the evaluations of Eq(coords, ·) over the
// boolean hypercube {0,1}ⁿ, where n = len(coords).
//
// On return, table[x] = Π_{i} Eq(coords[i], xᵢ) where x is decoded as
//
//	x₀ = x >> (n−1),   x₁ = (x >> (n−2)) & 1,   …,   x_{n-1} = x & 1,
//
// so coords[0] selects the most-significant bit of the index. This matches the
// coordinate convention used by [EvalMultilin] and [FoldInto].
//
// table must be pre-allocated with length 2^len(coords).
//
// An optional multiplier scales every entry: each table[x] is multiplied by
// multiplier[0] before the function returns. Only the first element is used.
func FoldedEqTableBase(table []field.Element, coords []field.Element, multiplier ...field.Element) {
	n := len(coords)
	if len(table) != 1<<n {
		panic(fmt.Sprintf(
			"polynomial: FoldedEqTableBase: table length %d != 2^%d",
			len(table), n,
		))
	}
	table[0].SetOne()
	if len(multiplier) > 0 {
		table[0] = multiplier[0]
	}
	for i, r := range coords {
		for j := 0; j < (1 << i); j++ {
			J := j << (n - i)
			JNext := J + (1 << (n - 1 - i))
			table[JNext].Mul(&r, &table[J])
			table[J].Sub(&table[J], &table[JNext])
		}
	}
}

// FoldedEqTableExt is the extension-field analogue of [FoldedEqTableBase].
// table must be pre-allocated with length 2^len(coords).
func FoldedEqTableExt(table []field.Ext, coords []field.Ext, multiplier ...field.Ext) {
	n := len(coords)
	if len(table) != 1<<n {
		panic(fmt.Sprintf(
			"polynomial: FoldedEqTableExt: table length %d != 2^%d",
			len(table), n,
		))
	}
	table[0].SetOne()
	if len(multiplier) > 0 {
		table[0] = multiplier[0]
	}
	for i, r := range coords {
		for j := 0; j < (1 << i); j++ {
			J := j << (n - i)
			JNext := J + (1 << (n - 1 - i))
			table[JNext].Mul(&r, &table[J])
			table[J].Sub(&table[J], &table[JNext])
		}
	}
}

// ChunkOfEqTableBase computes a contiguous chunk of the [FoldedEqTableBase]
// output table without materialising the full 2ⁿ-element result. This enables
// parallel construction: spawn one goroutine per (chunkID, chunkSize) pair.
//
//   - chunkSize must be a power of two and at most 2^len(coords).
//   - table must be pre-allocated with length chunkSize.
//   - chunkID must lie in [0, 2^len(coords)/chunkSize).
//
// The semantics are identical to:
//
//	full := make([]field.Element, 1<<len(coords))
//	FoldedEqTableBase(full, coords)
//	copy(table, full[chunkID*chunkSize : (chunkID+1)*chunkSize])
//
// An optional multiplier is forwarded to the inner [FoldedEqTableBase] call.
func ChunkOfEqTableBase(table []field.Element, chunkID, chunkSize int,
	coords []field.Element, multiplier ...field.Element) {

	nChunks := (1 << len(coords)) / chunkSize
	logNChunks := bits.TrailingZeros(uint(nChunks)) // log₂(nChunks); valid since nChunks is a power of 2

	var one field.Element
	one.SetOne()
	r := one
	if len(multiplier) > 0 {
		r = multiplier[0]
	}

	// Accumulate the prefix factor from the logNChunks outer coordinates.
	// The k-th outer coordinate (in bit-index order) is coords[logNChunks−k−1].
	for k := 0; k < logNChunks; k++ {
		rho := &coords[logNChunks-k-1]
		if (chunkID>>k)&1 == 1 {
			r.Mul(&r, rho)
		} else {
			var tmp field.Element
			tmp.Sub(&one, rho)
			r.Mul(&r, &tmp)
		}
	}

	FoldedEqTableBase(table, coords[logNChunks:], r)
}

// ChunkOfEqTableExt is the extension-field analogue of [ChunkOfEqTableBase].
func ChunkOfEqTableExt(table []field.Ext, chunkID, chunkSize int, coords []field.Ext, multiplier ...field.Ext) {
	nChunks := (1 << len(coords)) / chunkSize
	logNChunks := bits.TrailingZeros(uint(nChunks))

	var one, r field.Ext
	one.SetOne()
	r.SetOne()
	if len(multiplier) > 0 {
		r = multiplier[0]
	}

	for k := 0; k < logNChunks; k++ {
		rho := &coords[logNChunks-k-1]
		if (chunkID>>k)&1 == 1 {
			r.Mul(&r, rho)
		} else {
			var tmp field.Ext
			tmp.Sub(&one, rho)
			r.Mul(&r, &tmp)
		}
	}

	FoldedEqTableExt(table, coords[logNChunks:], r)
}
