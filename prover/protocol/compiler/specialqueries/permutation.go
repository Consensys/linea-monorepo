package specialqueries

import (
	sv "github.com/consensys/accelerated-crypto-monorepo/maths/common/smartvectors"
	"github.com/consensys/accelerated-crypto-monorepo/maths/common/vector"
	"github.com/consensys/accelerated-crypto-monorepo/maths/fft"
	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/coin"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/column"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/ifaces"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/query"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/wizard"
	"github.com/consensys/accelerated-crypto-monorepo/symbolic"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
)

const (
	/*
		Prefix to indicate an identifier is related to permutation
	*/
	PERMUTATION string = "PERMUTATION"
	/*
		Names for the intermediates commitments generated in the
		permutation compiler
	*/
	PERMUTATION_A_SUFFIX string = "A"
	PERMUTATION_B_SUFFIX string = "B"
	PERMUTATION_Z_SUFFIX string = "Z"
	/*
		Suffixes for the coins
	*/
	PERMUTATION_ALPHA_SUFFIX string = "ALPHA"
	PERMUTATION_BETA_SUFFIX  string = "BETA"
	/*
		The queries
			(a) Z(1) = 1
			(b) Z(g^n) = 1
			(c) Z(gX) * (beta + B(X)) = Z(X) * (beta + A(X))

	*/
	PERMUTATION_LINCOMB_A string = "LINCOMB_A"
	PERMUTATION_LINCOMB_B string = "LINCOMB_B"
	PERMUTATION_QA        string = "QA"
	PERMUTATION_QB        string = "QB"
	PERMUTATION_QC        string = "QC"
)

/*
Utility context whose goal is to gather all the parameters
generated during a specific permutation compilation. Should
not be exported. Is useful for code-factorization purpose.
*/
type permutationCtx struct {
	// Name of the query to be reduced
	name ifaces.QueryID
	// The query being reduced
	q query.Permutation
	// Names of the commitment created during the protocol
	A_NAME, B_NAME, Z_NAME ifaces.ColID
	// Names of the commitment created during the protocol
	A, B, Z ifaces.Column
	// Names of the coin created duing the protocol
	ALPHA, BETA coin.Name
	// Names of the queries created by the protocol
	QA, QB, QC, LINCOMB_A, LINCOMB_B ifaces.QueryID
	// Number of rows in both permutation tables
	N int
	// gValue is a (n+1)-root of unity. gN is its inverse.
	gValue, gNValue field.Element
	// Stores the coin info of the generated polynomials
	alpha, beta coin.Info
	// Has a single column
	hasSingleCol bool
	// round at which the inclusion was declared in the underlying protocol
	round int
}

/*
Reduce a permutation query. Follows the grand product argument
TODO: find the ref
*/
func reducePermutation(comp *wizard.CompiledIOP, meta *metaCtx, q query.Permutation, round int) {
	/*
		Sanity checks : Mark the query as compiled and make sure that
		it was not previously compiled.
	*/
	comp.QueriesNoParams.MarkAsIgnored(q.ID)

	/*
		Derives the identifiers name
	*/
	p := createPermutationCtx(comp, q, round)

	/*
		If necessary, collapses all the columns of A and B in a
		single column each through a random linear combination.
	*/
	p.compilerOptionalStep(comp)
	p.compilerCommitZ(comp)
	p.compilerQueries(comp)

	if !p.hasSingleCol {
		/*
			Add a round to compute the linear combinations in round n+1.
			Prover does no work at round n, because we need to sample alpha.
		*/
		meta.provers.AppendToInner(round+1, p.proverOptionalStep())
		meta.provers.AppendToInner(round+2, p.proverAssignZ())
	} else {
		meta.provers.AppendToInner(round+1, p.proverAssignZ())
	}
}

/*
Initializes all the static variable occuring during the protocol
The commitment.Info / coin.Info are describe later in the compilation
process. (Typically when they are defined).
*/
func createPermutationCtx(comp *wizard.CompiledIOP, q query.Permutation, round int) permutationCtx {
	res := permutationCtx{
		name:         q.ID,
		q:            q,
		Z_NAME:       deriveName[ifaces.ColID](PERMUTATION, q.ID, PERMUTATION_Z_SUFFIX),
		ALPHA:        deriveName[coin.Name](PERMUTATION, q.ID, PERMUTATION_ALPHA_SUFFIX),
		BETA:         deriveName[coin.Name](PERMUTATION, q.ID, PERMUTATION_BETA_SUFFIX),
		QA:           deriveName[ifaces.QueryID](PERMUTATION, q.ID, PERMUTATION_QA),
		QB:           deriveName[ifaces.QueryID](PERMUTATION, q.ID, PERMUTATION_QB),
		QC:           deriveName[ifaces.QueryID](PERMUTATION, q.ID, PERMUTATION_QC),
		LINCOMB_A:    deriveName[ifaces.QueryID](PERMUTATION, q.ID, PERMUTATION_LINCOMB_A),
		LINCOMB_B:    deriveName[ifaces.QueryID](PERMUTATION, q.ID, PERMUTATION_LINCOMB_B),
		N:            q.A[0].Size(),
		hasSingleCol: len(q.A) == 1,
		round:        round,
	}

	res.gValue = fft.GetOmega(res.N)
	res.gNValue.Inverse(&res.gValue) // Recall that g is a n+1 root of unity

	return res
}

/*
Run the first "optional step" of the compiler
*/
func (ctx *permutationCtx) compilerOptionalStep(comp *wizard.CompiledIOP) {

	if ctx.hasSingleCol {
		/*
			Then just assign `S` to be the `included` and `T` to be including.
			We do this, by changing the names. This works only because no commitments
			named with the old values of S and T have been registered so-far.
		*/
		ctx.A = ctx.q.A[0]
		ctx.B = ctx.q.B[0]
	} else {
		/*
			Else S and T are respectively random linear combinations of included
			and including (by `a`). The commitments are added in the same round as
			the query. Their respective info will be the same as the commitments
			they. The fact that this is a random linear combination is not registered
			as query, we will verify it on the fly for simplicity.

			TO BE DONE : find a way to do it without adding a commitment.
		*/
		if field.USING_GOLDILOCKS {
			utils.Panic("Not supported yet : can't do only one linear combination in this context")
		}

		ctx.alpha = comp.InsertCoin(ctx.round+1, ctx.ALPHA, coin.Field)

		ctx.A_NAME = deriveName[ifaces.ColID](PERMUTATION, ctx.q.ID, PERMUTATION_A_SUFFIX)
		ctx.B_NAME = deriveName[ifaces.ColID](PERMUTATION, ctx.q.ID, PERMUTATION_B_SUFFIX)
		ctx.A = comp.InsertCommit(ctx.round+1, ctx.A_NAME, ctx.N)
		ctx.B = comp.InsertCommit(ctx.round+1, ctx.B_NAME, ctx.N)

		lastColID := len(ctx.q.A) - 1

		/*
			Also create the global constraint that assess that "f" and "t" are indeed
			linear combinations of the columns of included and including. We do so using
			the Horner method.
		*/
		aExp := ifaces.ColumnAsVariable(ctx.q.A[lastColID])
		bExp := ifaces.ColumnAsVariable(ctx.q.B[lastColID])
		alpha := ctx.alpha.AsVariable()

		for i := lastColID - 1; i >= 0; i-- {
			tmpA := ifaces.ColumnAsVariable(ctx.q.A[i])
			aExp = aExp.Mul(alpha).Add(tmpA)

			tmpB := ifaces.ColumnAsVariable(ctx.q.B[i])
			bExp = bExp.Mul(alpha).Add(tmpB)
		}

		aExp = aExp.Sub(ifaces.ColumnAsVariable(ctx.A))
		bExp = bExp.Sub(ifaces.ColumnAsVariable(ctx.B))

		comp.InsertGlobal(ctx.round+1, ctx.LINCOMB_A, aExp)
		comp.InsertGlobal(ctx.round+1, ctx.LINCOMB_B, bExp)
	}
}

/*
Compilation step - commitment to Z. Final commitment phase
Sample beta and gamma. and commit to Z
*/
func (ctx *permutationCtx) compilerCommitZ(comp *wizard.CompiledIOP) {
	/*
		Need to account for the fact that if not "hasSingleCol" then, a round was
		introduced already.
	*/
	round_ := ctx.round + 1
	if !ctx.hasSingleCol {
		round_++
	}

	ctx.beta = comp.InsertCoin(round_, ctx.BETA, coin.Field)
	ctx.Z = comp.InsertCommit(round_, ctx.Z_NAME, ctx.N)
}

func (ctx *permutationCtx) compilerQueries(comp *wizard.CompiledIOP) {
	/*
		Detect the current round. Accounts for the fact that this lookup can be on
		one or multiple columns.
	*/
	round_ := ctx.round + 1
	if !ctx.hasSingleCol {
		round_++
	}

	a := ifaces.ColumnAsVariable(ctx.A)
	b := ifaces.ColumnAsVariable(ctx.B)
	z := ifaces.ColumnAsVariable(ctx.Z)
	z1 := ifaces.ColumnAsVariable(column.Shift(ctx.Z, 1))
	beta := ctx.beta.AsVariable()
	one := symbolic.NewConstant(1)

	/*
		Left and Right hand of the equation of QC
	*/
	cs := b.Add(beta).Mul(z1)
	right := a.Add(beta).Mul(z)
	cs = cs.Sub(right)

	comp.InsertLocal(round_, ctx.QA, z.Sub(one))
	comp.InsertGlobal(round_, ctx.QB, cs, true) // We forbid the boundary cancelling
}

/*
Prover for the optional step - compute F and T as linear combinations of the
*/
func (ctx *permutationCtx) proverOptionalStep() wizard.ProverStep {
	/*
		Optional step : if necessary collapse all queries into
		a single column. E.G create F and T. If so, this adds
		one extra round at the beginning because we can't assign
		them in the same round as the original commitments were
		assigned : we need to generate `alpha` between the two.
	*/
	return func(run *wizard.ProverRuntime) {

		var aWit sv.SmartVector = sv.NewConstant(field.Zero(), ctx.N)
		var bWit sv.SmartVector = sv.NewConstant(field.Zero(), ctx.N)

		alphaWit := run.GetRandomCoinField(ctx.ALPHA)
		alphaPow := field.One()

		for colID := range ctx.q.A {
			/*
				Recall that `A` and `B` have the same number of column left.
				Thus we can compute both linear combinations in parallels.
			*/
			a := ctx.q.A[colID]
			aColWit := a.GetColAssignment(run)
			tmpA := sv.ScalarMul(aColWit, alphaPow)
			aWit = sv.Add(tmpA, aWit)

			b := ctx.q.B[colID]
			bColWit := b.GetColAssignment(run)
			tmpF := sv.ScalarMul(bColWit, alphaPow)
			bWit = sv.Add(tmpF, bWit)

			alphaPow.Mul(&alphaPow, &alphaWit)
		}

		run.AssignColumn(ctx.A_NAME, aWit)
		run.AssignColumn(ctx.B_NAME, bWit)
	}
}

/*
Compute Z - Each of Z's
*/
func (l *permutationCtx) proverAssignZ() wizard.ProverStep {
	return func(run *wizard.ProverRuntime) {

		beta := run.GetRandomCoinField(l.BETA)
		a := l.A.GetColAssignment(run)
		b := l.B.GetColAssignment(run)

		numerator := make([]field.Element, l.N)
		denominator := make([]field.Element, l.N)

		/*
			Z is expressed as a quotient. For efficiency concerns, we compute the
			numerator and the denominator apart. Then, we use the batch inverse
			trick to compute the quotient.
		*/
		numerator[0] = field.One()
		denominator[0] = field.One()

		for i := 0; i < l.N-1; i++ {
			ai, bi := a.Get(i), b.Get(i)
			numerator[i+1].Add(&beta, &ai).Mul(&numerator[i+1], &numerator[i])
			denominator[i+1].Add(&beta, &bi).Mul(&denominator[i+1], &denominator[i])
		}

		z := numerator
		denominator = field.BatchInvert(denominator)
		vector.MulElementWise(z, z, denominator)
		run.AssignColumn(l.Z_NAME, sv.NewRegular(z))
	}

}
