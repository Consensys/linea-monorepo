package dist_permutation

import (
	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/permutation"
	modulediscoverer "github.com/consensys/linea-monorepo/prover/protocol/distributed/module_discoverer"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/protocol/wizardutils"
	"github.com/consensys/linea-monorepo/prover/symbolic"
)

/*
The below function does the following:
1. For a given target module name, it finds all the relevant permutation query and combine them into a big grand product query
*/
type PermutationIntoGrandProductCtx struct {
	numerators   []*symbolic.Expression // aimed at storing the expressions Ai + \beta_i for a particular permutation query
	denominators []*symbolic.Expression // aimed at storing the expressions Bi + \beta_i for a particular permutation query
}

// Return a new PermutationIntoGrandProductCtx with numerator and denominator set as symobilc constant 1
func newPermutationIntoGrandProductCtx(s Settings) *PermutationIntoGrandProductCtx {
	permCtx := PermutationIntoGrandProductCtx{}
	permCtx.numerators = make([]*symbolic.Expression, s.maxNumofQueryPerModule)
	permCtx.denominators = make([]*symbolic.Expression, s.maxNumofQueryPerModule)
	return &permCtx
}

func AddGdProductQuery(initialComp, moduleComp *wizard.CompiledIOP,
	targetModuleName modulediscoverer.ModuleName,
	s Settings) {
	numRounds := initialComp.NumRounds()
	permCtx := newPermutationIntoGrandProductCtx(s)
	// Initialise the period separating module discoverer
	disc := modulediscoverer.PeriodSeperatingModuleDiscoverer{}
	disc.Analyze(initialComp)
	/*
		Handles the lookups and permutations checks
	*/
	for i := 0; i < numRounds; i++ {
		queries := initialComp.QueriesNoParams.AllKeysAt(i)
		for j, qName := range queries {
			// Skip if it was already compiled
			if initialComp.QueriesNoParams.IsIgnored(qName) {
				continue
			}

			switch q_ := initialComp.QueriesNoParams.Data(qName).(type) {
			case query.Permutation:
				{
					moduleNameA := disc.FindModule(q_.A[0][0])
					moduleNameB := disc.FindModule(q_.B[0][0])
					if moduleNameA == targetModuleName && moduleNameB != targetModuleName {
						permCtx.push(moduleComp, &q_, i, j, true, false)
					} else if moduleNameA != targetModuleName && moduleNameB == targetModuleName {
						permCtx.push(moduleComp, &q_, i, j, false, false)
					} else if moduleNameA == targetModuleName && moduleNameB == targetModuleName {
						permCtx.push(moduleComp, &q_, i, j, true, true)
					} else {
						continue
					}
				}
			default:
				continue
			}
		}
	}
	// Reduce a permutation query into a GrandProduct query
	moduleComp.InsertGrandProduct(0, ifaces.QueryIDf(string(targetModuleName)), permCtx.numerators, permCtx.denominators)
}

// The below function does the following:
// 1. Register beta and alpha (for the random linear combination in case A and B are multi-columns) in the compiledIop
// 2. Populates the nemerators and the denominators of the grand product query
func (p *PermutationIntoGrandProductCtx) push(comp *wizard.CompiledIOP, q *query.Permutation, round, queryInRound int, isNumerator, isBoth bool) {
	/*
		Sanity checks : Mark the query as compiled and make sure that
		it was not previously compiled.
	*/
	if comp.QueriesNoParams.MarkAsIgnored(q.ID) {
		panic("did not expect that a query no param could be ignored at this stage")
	}
	var (
		isMultiColumn = len(q.A[0]) > 1
		alpha         coin.Info
		// beta has to be different for different queries for a perticular round for the soundness of z-packing
		beta = comp.InsertCoin(round+1, permutation.DeriveName[coin.Name](*q, "BETA_%v", queryInRound), coin.Field)
	)

	if isMultiColumn {
		// alpha has to be different for different queries for a perticular round for the soundness of z-packing
		alpha = comp.InsertCoin(round+1, permutation.DeriveName[coin.Name](*q, "ALPHA_%v", queryInRound), coin.Field)

	}
	if isNumerator && !isBoth {
		// Take only the numerator
		factor := computeFactor(q.A, isMultiColumn, alpha, beta)
		p.numerators = append(p.numerators, factor)
	} else if !isNumerator && !isBoth {
		// Take only the denominator
		factor := computeFactor(q.B, isMultiColumn, alpha, beta)
		p.denominators = append(p.denominators, factor)
	} else if isNumerator && isBoth {
		// Take both the numerator and the denominator
		numFactor := computeFactor(q.A, isMultiColumn, alpha, beta)
		denFactor := computeFactor(q.B, isMultiColumn, alpha, beta)
		p.numerators = append(p.numerators, numFactor)
		p.denominators = append(p.denominators, denFactor)
	} else if !isNumerator && isBoth {
		panic("Invalid case")
	}

}

func computeFactor(aOrB [][]ifaces.Column, isMultiColumn bool, alpha, beta coin.Info) *symbolic.Expression {
	var (
		numFrag = len(aOrB)
		factor  = symbolic.NewConstant(1)
	)

	for frag := range numFrag {
		fragFactor := symbolic.NewVariable(aOrB[frag][0])
		if isMultiColumn {
			fragFactor = wizardutils.RandLinCombColSymbolic(alpha, aOrB[frag])
		}
		fragFactor = symbolic.Add(fragFactor, beta)
		factor = symbolic.Mul(factor, fragFactor)
	}
	return factor
}