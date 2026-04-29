package sumcheck

import (
	"fmt"

	"github.com/consensys/linea-monorepo/prover-ray/maths/koalabear/field"
)

// Verify checks a sumcheck proof against the initial claimed sum.
//
// degree is the expected gate degree; each round polynomial must have exactly
// degree evaluations (the Gruen compressed format: {P(0), P(2), …, P(degree)}).
//
// challengeFn is called once per round with the current round polynomial; the
// caller must update its transcript and return the next verifier challenge.
//
// Returns:
//   - finalClaim: the last updated claim value (to be checked externally as
//     finalClaim == eq(challenges, qPrime) · Gate(tableClaims…))
//   - challenges: the verifier challenges, one per round
//   - err: non-nil if any consistency check fails
func Verify(
	claim field.Ext,
	proof []RoundPoly,
	degree int,
	challengeFn func(RoundPoly) field.Ext,
) (finalClaim field.Ext, challenges []field.Ext, err error) {

	challenges = make([]field.Ext, len(proof))

	for i, rp := range proof {
		if len(rp) != degree {
			return field.Ext{}, nil, fmt.Errorf(
				"sumcheck: Verify: proof[%d] has %d evaluations, want %d (degree)",
				i, len(rp), degree,
			)
		}

		// Gruen compressed format: P(1) = claim − P(0) is implicit.
		// The consistency check P(0) + P(1) = claim holds by construction.
		// Sample a challenge and advance the claim to P(r).
		r := challengeFn(rp)
		challenges[i] = r
		claim = rp.EvalAt(r, claim)
	}

	return claim, challenges, nil
}
