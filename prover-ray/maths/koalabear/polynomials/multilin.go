package polynomials

import (
	"fmt"

	"github.com/consensys/linea-monorepo/prover-ray/maths/koalabear/field"
)

// FoldInto folds a multilinear polynomial table on its first variable.
//
// A multilinear polynomial in n variables is stored as a length-2ⁿ slice of
// evaluations on the boolean hypercube {0,1}ⁿ. Folding at r produces the
// unique multilinear polynomial in n−1 variables obtained by substituting r
// for the first variable:
//
//	res[i] = table[i] + r·(table[i+mid] − table[i]),   mid = len(table)/2.
//
// Supported (res, table, r) type combinations:
//
//	(base, base, base)  all-base arithmetic
//	(ext,  base, ext)   table values promoted to ext on the fly
//	(ext,  ext,  base)  cheaper MulByElement path for the scalar
//	(ext,  ext,  ext)   full extension arithmetic
//
// Panics if res.Len()·2 ≠ table.Len() or the type combination is unsupported.
func FoldInto(res, table field.Vec, r field.Gen) {
	if res.Len()*2 != table.Len() {
		panic(fmt.Sprintf(
			"polynomial: FoldInto: res.Len()=%d must be half of table.Len()=%d",
			res.Len(), table.Len(),
		))
	}
	switch {
	case res.IsBase():
		foldBaseBaseInto(res.AsBase(), table.AsBase(), r.AsBase())
	case table.IsBase():
		foldBaseExtInto(res.AsExt(), table.AsBase(), r.AsExt())
	case r.IsBase():
		foldExtBaseInto(res.AsExt(), table.AsExt(), r.AsBase())
	default:
		foldExtExtInto(res.AsExt(), table.AsExt(), r.AsExt())
	}
}

// EvalMultilin evaluates a multilinear polynomial at the given coordinates.
//
// The polynomial P is the unique multilinear polynomial satisfying
// P(x) = table[x] for every x ∈ {0,1}ⁿ (x interpreted as an n-bit integer
// with coords[0] selecting the most-significant bit). Evaluation proceeds by
// n successive folds — one per coordinate — without modifying the input table.
//
// len(table) must equal 2^len(coords).
//
// The result is tagged base iff table and every coordinate are base-field.
func EvalMultilin(table field.Vec, coords []field.Gen) field.Gen {
	n := len(coords)
	if table.Len() != 1<<n {
		panic(fmt.Sprintf(
			"polynomial: EvalMultilin: table.Len()=%d must equal 2^%d=%d",
			table.Len(), n, 1<<n,
		))
	}
	if n == 0 {
		if table.IsBase() {
			return field.ElemFromBase(table.AsBase()[0])
		}
		return field.ElemFromExt(table.AsExt()[0])
	}

	// Stay in base only when every input is base-field.
	outIsBase := table.IsBase()
	if outIsBase {
		for _, c := range coords {
			if !c.IsBase() {
				outIsBase = false
				break
			}
		}
	}

	if outIsBase {
		work := make([]field.Element, table.Len())
		copy(work, table.AsBase())
		for _, r := range coords {
			mid := len(work) / 2
			foldBaseBaseInto(work[:mid], work, r.AsBase())
			work = work[:mid]
		}
		return field.ElemFromBase(work[0])
	}

	// Extension path: lift table to ext if needed, then fold.
	work := make([]field.Ext, table.Len())
	if table.IsBase() {
		for i, e := range table.AsBase() {
			work[i] = field.Lift(e)
		}
	} else {
		copy(work, table.AsExt())
	}
	for _, r := range coords {
		mid := len(work) / 2
		if r.IsBase() {
			foldExtBaseInto(work[:mid], work, r.AsBase())
		} else {
			foldExtExtInto(work[:mid], work, r.AsExt())
		}
		work = work[:mid]
	}
	return field.ElemFromExt(work[0])
}

// ---------------------------------------------------------------------------
// Typed fold kernels (unexported)
// ---------------------------------------------------------------------------

// foldBaseBaseInto sets res[i] = bottom[i] + r·(top[i]−bottom[i]) for all i,
// where bottom = table[:len(res)], top = table[len(res):].
// Safe when res aliases table[:len(res)].
func foldBaseBaseInto(res, table []field.Element, r field.Element) {
	mid := len(res)
	bottom, top := table[:mid], table[mid:]
	for i := range res {
		var diff field.Element
		diff.Sub(&top[i], &bottom[i])
		diff.Mul(&diff, &r)
		res[i].Add(&bottom[i], &diff)
	}
}

// foldBaseExtInto folds a base-field table with an extension-field scalar,
// promoting base values to ext on the fly.
// res[i] = Lift(bottom[i]) + r·Lift(top[i]−bottom[i])
func foldBaseExtInto(res []field.Ext, table []field.Element, r field.Ext) {
	mid := len(res)
	bottom, top := table[:mid], table[mid:]
	for i := range res {
		var diff field.Element
		diff.Sub(&top[i], &bottom[i])
		res[i] = field.Lift(bottom[i])
		var rDiff field.Ext
		rDiff.MulByElement(&r, &diff)
		res[i].Add(&res[i], &rDiff)
	}
}

// foldExtBaseInto folds an extension-field table with a base-field scalar.
// Uses the cheaper MulByElement path. Safe when res aliases table[:len(res)].
func foldExtBaseInto(res []field.Ext, table []field.Ext, r field.Element) {
	mid := len(res)
	bottom, top := table[:mid], table[mid:]
	for i := range res {
		var diff, rDiff field.Ext
		diff.Sub(&top[i], &bottom[i])
		rDiff.MulByElement(&diff, &r)
		res[i].Add(&bottom[i], &rDiff)
	}
}

// foldExtExtInto folds an extension-field table with an extension-field scalar.
// Safe when res aliases table[:len(res)].
func foldExtExtInto(res []field.Ext, table []field.Ext, r field.Ext) {
	mid := len(res)
	bottom, top := table[:mid], table[mid:]
	for i := range res {
		var diff, rDiff field.Ext
		diff.Sub(&top[i], &bottom[i])
		rDiff.Mul(&diff, &r)
		res[i].Add(&bottom[i], &rDiff)
	}
}
