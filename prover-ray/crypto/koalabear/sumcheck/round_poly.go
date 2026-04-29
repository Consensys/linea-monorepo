package sumcheck

import (
	"github.com/consensys/linea-monorepo/prover-ray/maths/koalabear/field"
)

// RoundPoly is a sumcheck round polynomial stored in Gruen compressed format:
// evaluations at {0, 2, 3, …, d} where d = gate.Degree().
//
// P(1) is intentionally omitted; it can be recovered as P(1) = claim − P(0),
// where claim is the expected sum for that round. This saves one gate evaluation
// pass in the inner loop and one extension-field element per round in the proof
// transcript (Gruen 2024 optimisation).
//
// Concretely:
//
//	rp[0]   = P(0)
//	rp[1]   = P(2)
//	rp[2]   = P(3)
//	…
//	rp[d-1] = P(d)
type RoundPoly []field.Ext

// EvalAt evaluates the round polynomial at r using Lagrange interpolation over
// the nodes {0, 1, 2, …, d}. claim is the current round claim; it is used to
// reconstruct the missing P(1) = claim − P(0).
func (rp RoundPoly) EvalAt(r, claim field.Ext) field.Ext {
	d := len(rp) // = gate.Degree()

	// Reconstruct P(1) = claim − P(0).
	var p1 field.Ext
	p1.Sub(&claim, &rp[0])

	// Build the full evaluation vector: vals[i] = P(i), i = 0 … d.
	vals := make([]field.Ext, d+1)
	vals[0] = rp[0]
	vals[1] = p1
	for i := 2; i <= d; i++ {
		vals[i] = rp[i-1]
	}

	// Lagrange interpolation at r over nodes {0, 1, …, d}.
	//
	//   P(r) = Σ_{i=0}^{d} P(i) · L_i(r)
	//   L_i(r) = Π_{j≠i} (r − j) / (i − j)
	//
	// The denominators (i − j) are integer constants; we compute them
	// in the field. For small d (≤ 5) this is a handful of operations.
	var result field.Ext

	for i := 0; i <= d; i++ {
		var numer, denom field.Ext
		numer.SetOne()
		denom.SetOne()

		for j := 0; j <= d; j++ {
			if j == i {
				continue
			}
			// numer *= (r − j)
			jExt := field.Lift(intToElem(j))
			var rj field.Ext
			rj.Sub(&r, &jExt)
			numer.Mul(&numer, &rj)

			// denom *= (i − j)  [integer difference, lifted to field]
			ijExt := field.Lift(intToElem(i - j))
			denom.Mul(&denom, &ijExt)
		}

		// term = P(i) · numer · denom⁻¹
		var inv field.Ext
		inv.Inverse(&denom)
		numer.Mul(&numer, &inv)
		numer.Mul(&numer, &vals[i])
		result.Add(&result, &numer)
	}

	return result
}

// intToElem converts a (possibly negative) int to a field.Element.
func intToElem(v int) field.Element {
	var e field.Element
	if v >= 0 {
		e.SetUint64(uint64(v))
	} else {
		e.SetUint64(uint64(-v))
		e.Neg(&e)
	}
	return e
}
