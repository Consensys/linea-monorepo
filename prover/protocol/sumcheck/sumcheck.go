package sumcheck

import "github.com/consensys/linea-monorepo/prover/maths/field/fext"

// Claim is one input to a batched sumcheck: a multilinear polynomial P_i is
// asserted to evaluate to Eval at Point. All claims fed to ProveBatched/
// VerifyBatched must share the same number of variables (= len(Point)).
type Claim struct {
	Point []fext.Element
	Eval  fext.Element
}

// BatchedProof is the transcript of a batched sumcheck reducing N input
// claims (P_i, r_i, y_i) to a single residual point c, plus the per-poly
// evaluations P_i(c). The verifier samples lambda once at the start to form
// the random linear combination of claims, then n round challenges driving
// the sumcheck.
//
// Layout of RoundPolys: one entry per variable, each storing the three
// evaluations of the round polynomial at X ∈ {0, 1, 2}. The round polynomial
// has degree 2 because each summand eq(r_i, X) * P_i(X) is degree 1 in eq
// and degree 1 in P_i along the bound variable, so degree 2 total.
type BatchedProof struct {
	RoundPolys [][3]fext.Element
	FinalEvals []fext.Element // P_i(c) for each input claim, in order
}

// labels for the transcript. Centralised so prover and verifier stay in sync.
const (
	labelClaimEval     = "claim-eval"
	labelClaimPoint    = "claim-point"
	labelLambda        = "lambda"
	labelRoundPoly     = "round-poly"
	labelRoundChallenge = "round-challenge"
	labelFinalEvals    = "final-evals"
)
