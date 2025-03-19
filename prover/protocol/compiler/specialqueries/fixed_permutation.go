package specialqueries

import (
	sv "github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

// Reduce the fixed permutations into a permutation query
func CompileFixedPermutations(comp *wizard.CompiledIOP) {

	numRounds := comp.NumRounds()

	for i := 0; i < numRounds; i++ {
		queries := comp.QueriesNoParams.AllKeysAt(i)
		for _, qName := range queries {
			// Skip if it was already compiled to avoid recompiling a second
			// time the same query.
			if comp.QueriesNoParams.IsIgnored(qName) {
				continue
			}

			switch q_ := comp.QueriesNoParams.Data(qName).(type) {
			case query.FixedPermutation:
				reduceFixedPermutation(comp, q_)
			}
		}
	}
}

// Reduce a permutation query. Follows the grand product argument
// from PLONK paper: extended permutation part
func reduceFixedPermutation(comp *wizard.CompiledIOP, q query.FixedPermutation) {
	/*
		Sanity checks : Mark the query as compiled and make sure that
		it was not previously compiled.
	*/
	if comp.QueriesNoParams.MarkAsIgnored(q.ID) {
		panic("did not expect that a query no param could be ignored at this stage")
	}

	var (
		round = comp.QueriesNoParams.Round(q.ID)
		n     = q.A[0].Size()
		sid   = make([]ifaces.Column, len(q.S))
		s     = make([]ifaces.Column, len(q.S))
		perm  = query.Permutation{
			ID: q.Name() + "_FIXED_PERM",
		}
	)

	for i := range s {
		sid[i] = comp.InsertPrecomputed(deriveNamePerm("SID", q.ID, i), getSiD(i, n))
		s[i] = comp.InsertPrecomputed(deriveNamePerm("S", q.ID, i), q.S[i])
		perm.A = append(perm.A, []ifaces.Column{q.A[i], sid[i]})
		perm.B = append(perm.B, []ifaces.Column{q.B[i], s[i]})
	}

	comp.InsertFragmentedPermutation(round, perm.ID, perm.A, perm.B)
}

func deriveNamePerm(r string, queryName ifaces.QueryID, i int) ifaces.ColID {
	return ifaces.ColIDf("%v_%v_%v", queryName, r, i)
}

// getSiD returns a smartvector storing the witness of an SiD column. The witness
// is defined as the arithmetic sequence nj, nj+1, nj+2, ..., nj+n-1
func getSiD(j, n int) sv.SmartVector {
	identity := make([]field.Element, n)
	for i := 0; i < n; i++ {
		identity[i] = field.NewElement(uint64(n*j + i))
	}
	return sv.NewRegular(identity)
}
