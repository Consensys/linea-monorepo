package specialqueries

import (
	sv "github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/common/vector"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils/profiling"
)

const (
	/*
		Prefix to indicate an identifier is related to fixedpermutation
	*/
	FIXED_PERMUTATION      string = "FIXED_PERMUTATION"
	PERMUTATION_COLLAPSE_A string = "PERMUTATION_COLLAPSE_A"
	PERMUTATION_COLLAPSE_B string = "PERMUTATION_COLLAPSE_B"
)

/*
Utility context whose goal is to gather all the parameters
generated during a specific fixedPermutation compilation. Should
not be exported. Is useful for code-factorization purpose.
*/
type fixedPermutationCtx struct {
	q query.FixedPermutation
	// Names of the commitment created during the protocol
	Acollapse_NAME, Bcollapse_NAME, Z_NAME ifaces.ColID
	// Names of the commitment created during the protocol
	//all the splittings A of fixedPermutation are collapsed to the single column Acollapse, similarly for B
	Acollapse, Bcollapse, Z ifaces.Column
	// Names of the coin created duing the protocol
	ALPHA, BETA coin.Name
	// Names of the queries created by the protocol (queries of the form local and Global constraints)
	QA, QB, COLLAPSE_A, COLLAPSE_B ifaces.QueryID
	// Stores the coin info of the generated polynomials
	alpha, beta coin.Info
	//splittings of the permutation as handles
	S_id, S          []ifaces.Column
	Sid_NAME, S_NAME []ifaces.ColID
	// the identity polynomials for fix permutaion
	SidWit []ifaces.ColAssignment
	//size of the columns
	N     int
	round int
}

/*
Reduce a permutation query. Follows the grand product argument
from PLONK paper: extended permutation part
*/
func reduceFixedPermutation(comp *wizard.CompiledIOP, q query.FixedPermutation, round int) {
	/*
		Sanity checks : Mark the query as compiled and make sure that
		it was not previously compiled.
	*/
	if comp.QueriesNoParams.MarkAsIgnored(q.ID) {
		panic("did not expect that a query no param could be ignored at this stage")
	}

	/*
		Derives the identifiers name
	*/
	p := createFixedPermutationCtx(q, round)

	/*
		 collapses all the columns of A and B to the
		single columns each.
	*/

	p.compilerColapsStep(comp)
	p.compilerPolyZ(comp)

	comp.SubProvers.AppendToInner(round, p.proverAssignS())
	comp.SubProvers.AppendToInner(round+1, p.proverColapsStep())
	comp.SubProvers.AppendToInner(round+1, p.proverAssignExtendedZ())

}

/*
Initializes all the static variable occuring during the protocol
The commitment.Info / coin.Info are describe later in the compilation
process. (Typically when they are defined).
*/
func createFixedPermutationCtx(q query.FixedPermutation, round int) fixedPermutationCtx {
	n := q.S[0].Len()

	//name for S,S_id
	name := make([]ifaces.ColID, len(q.S))
	nameID := make([]ifaces.ColID, len(q.S))
	sid := make([]ifaces.Column, len(q.S))
	s := make([]ifaces.Column, len(q.S))
	for i := range q.S {
		nameID[i] = deriveNamePerm("IDENTITY_PERM", q.ID, i)
		name[i] = deriveNamePerm("PERM", q.ID, i)

	}
	//assigning identity polynomials s_{id}
	sidWit := make([]ifaces.ColAssignment, len(q.S))
	for j := range q.S {
		identity := make([]field.Element, n)
		for i := 0; i < n; i++ {
			identity[i] = field.NewElement(uint64(n*j + i))
		}
		sidWit[j] = sv.NewRegular(identity)
	}
	res := fixedPermutationCtx{
		Acollapse_NAME: deriveName[ifaces.ColID](FIXED_PERMUTATION, q.ID, "A"),
		Bcollapse_NAME: deriveName[ifaces.ColID](FIXED_PERMUTATION, q.ID, "B"),
		Z_NAME:         deriveName[ifaces.ColID](FIXED_PERMUTATION, q.ID, "Z"),
		ALPHA:          deriveName[coin.Name](FIXED_PERMUTATION, q.ID, "ALPHA"),
		BETA:           deriveName[coin.Name](FIXED_PERMUTATION, q.ID, "BETA"),
		QA:             deriveName[ifaces.QueryID](FIXED_PERMUTATION, q.ID, "QA"),
		QB:             deriveName[ifaces.QueryID](FIXED_PERMUTATION, q.ID, "QB"),
		COLLAPSE_A:     deriveName[ifaces.QueryID](FIXED_PERMUTATION, q.ID, "COLLAPSE_A"),
		COLLAPSE_B:     deriveName[ifaces.QueryID](FIXED_PERMUTATION, q.ID, "COLLAPSE_B"),
		// for S_id,S
		Sid_NAME: nameID,
		S_NAME:   name,
		SidWit:   sidWit,
		round:    round,
		N:        n,
		q:        q,
		S_id:     sid,
		S:        s,
	}

	return res
}

/*
Run the first "collapse step" of the compiler to collapse all the splittings A to a single column
*/
func (t *fixedPermutationCtx) compilerColapsStep(comp *wizard.CompiledIOP) {
	//tracker : round or 0? as it is used by PLONK it should be the same round as PLONK
	for i := range t.S {
		t.S[i] = comp.InsertCommit(t.round, t.S_NAME[i], t.N)
		t.S_id[i] = comp.InsertCommit(t.round, t.Sid_NAME[i], t.N)
	}

	t.alpha = comp.InsertCoin(t.round+1, t.ALPHA, coin.Field)
	t.beta = comp.InsertCoin(t.round+1, t.BETA, coin.Field)
	alpha := t.alpha.AsVariable()
	beta := t.beta.AsVariable()

	// building Acollapse=prod_i (A_i+alpha*S_{id,i}+beta)
	t.Acollapse = comp.InsertCommit(t.round+1, t.Acollapse_NAME, t.N)
	t.Bcollapse = comp.InsertCommit(t.round+1, t.Bcollapse_NAME, t.N)

	aExp := symbolic.NewConstant(1)
	bExp := symbolic.NewConstant(1)
	for i := range t.q.A {
		aExp = aExp.Mul(ifaces.ColumnAsVariable(t.q.A[i]).
			Add(alpha.Mul(ifaces.ColumnAsVariable(t.S_id[i]))).
			Add(beta))
		bExp = bExp.Mul(ifaces.ColumnAsVariable(t.q.B[i]).
			Add(alpha.Mul(ifaces.ColumnAsVariable(t.S[i]))).
			Add(beta))
	}

	aExp = aExp.Sub(ifaces.ColumnAsVariable(t.Acollapse))
	bExp = bExp.Sub(ifaces.ColumnAsVariable(t.Bcollapse))

	comp.InsertGlobal(t.round+1, t.COLLAPSE_A, aExp)
	comp.InsertGlobal(t.round+1, t.COLLAPSE_B, bExp)

}

/*
Compilation step - commitment to Z. Final commitment phase
*/

func (t *fixedPermutationCtx) compilerPolyZ(comp *wizard.CompiledIOP) {
	round_ := t.round + 1
	//commit to Z
	t.Z = comp.InsertCommit(round_, t.Z_NAME, t.N)

	a := ifaces.ColumnAsVariable(t.Acollapse)
	b := ifaces.ColumnAsVariable(t.Bcollapse)
	z := ifaces.ColumnAsVariable(t.Z)
	z1 := ifaces.ColumnAsVariable(column.Shift(t.Z, 1))
	one := symbolic.NewConstant(1)

	//constraints
	cs := b.Mul(z1)
	right := a.Mul(z)
	cs = cs.Sub(right)

	comp.InsertLocal(round_, t.QA, z.Sub(one))
	comp.InsertGlobal(round_, t.QB, cs, true) // We forbid the boundary cancelling
}

// Prover assignes  witnesses to S,S_id
func (t *fixedPermutationCtx) proverAssignS() wizard.MainProverStep {
	return func(run *wizard.ProverRuntime) {
		stopTimer := profiling.LogTimer("exPermutation prover - assign s %v", t.q.ID)
		for colID := range t.S {
			run.AssignColumn(t.Sid_NAME[colID], t.SidWit[colID])
			run.AssignColumn(t.S_NAME[colID], t.q.S[colID])
		}
		stopTimer()
	}
}

/*
Prover for the collapse step - compute Acollapse and Bcollapse
*/

func (t *fixedPermutationCtx) proverColapsStep() wizard.MainProverStep {
	return func(run *wizard.ProverRuntime) {
		stopTimer := profiling.LogTimer("exPermutation prover - colaps step %v", t.q.ID)
		alphaWit := run.GetRandomCoinField(t.ALPHA)
		betaWit := run.GetRandomCoinField(t.BETA)

		var betaVec sv.SmartVector = sv.NewConstant(betaWit, t.N)
		var aWit sv.SmartVector = sv.NewConstant(field.One(), t.N)
		var bWit sv.SmartVector = sv.NewConstant(field.One(), t.N)

		for colID := range t.q.A {
			/*
				Recall that `A` and `B` have the same number of column left.
				Thus we can compute both collapses in parallels.
			*/
			a := t.q.A[colID]
			aColWit := a.GetColAssignment(run)
			tmpA := sv.ScalarMul(t.SidWit[colID], alphaWit)
			u := sv.Add(tmpA, betaVec)
			v := sv.Add(u, aColWit)
			aWit = sv.Mul(v, aWit)

			b := t.q.B[colID]
			bColWit := b.GetColAssignment(run)
			tmpF := sv.ScalarMul(t.q.S[colID], alphaWit)
			u = sv.Add(tmpF, betaVec)
			v = sv.Add(u, bColWit)
			bWit = sv.Mul(v, bWit)

		}

		run.AssignColumn(t.Acollapse_NAME, aWit)
		run.AssignColumn(t.Bcollapse_NAME, bWit)

		stopTimer()
	}
}

/*
Compute Z
*/
func (t *fixedPermutationCtx) proverAssignExtendedZ() wizard.MainProverStep {
	return func(run *wizard.ProverRuntime) {

		stopTimer := profiling.LogTimer("exPermutation prover - assign extended z %v", t.q.ID)

		a := t.Acollapse.GetColAssignment(run)
		b := t.Bcollapse.GetColAssignment(run)

		numerator := make([]field.Element, t.N)
		denominator := make([]field.Element, t.N)

		/*
			Z is expressed as a quotient. For efficiency concerns, we compute the
			numerator and the denominator apart. Then, we use the batch inverse
			trick to compute the quotient.
		*/
		numerator[0] = field.One()
		denominator[0] = field.One()

		for i := 0; i < t.N-1; i++ {
			ai, bi := a.Get(i), b.Get(i)
			numerator[i+1] = ai
			denominator[i+1] = bi
			numerator[i+1].Mul(&numerator[i+1], &numerator[i])
			denominator[i+1].Mul(&denominator[i+1], &denominator[i])
		}

		z := numerator
		denominator = field.BatchInvert(denominator)
		vector.MulElementWise(z, z, denominator)

		run.AssignColumn(t.Z_NAME, sv.NewRegular(z))

		stopTimer()
	}

}
func deriveNamePerm(r string, queryName ifaces.QueryID, i int) ifaces.ColID {
	return ifaces.ColIDf("%v_%v_%v", queryName, r, i)
}
