package dist_permutation

import (
	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/permutation"
	modulediscoverer "github.com/consensys/linea-monorepo/prover/protocol/distributed/module_discoverer"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/protocol/wizardutils"
	"github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/sirupsen/logrus"
)

/*
The below function does the following:
1. For a given target module name, it finds all the relevant permutation query and combine them into a big grand product query
*/
type PermutationIntoGrandProductCtx struct {
	numerators   []*symbolic.Expression // aimed at storing the expressions Ai + \beta_i for a particular permutation query
	denominators []*symbolic.Expression // aimed at storing the expressions Bi + \beta_i for a particular permutation query
}

// Returns a new PermutationIntoGrandProductCtx
func newPermutationIntoGrandProductCtx(s Settings) *PermutationIntoGrandProductCtx {
	permCtx := PermutationIntoGrandProductCtx{}
	permCtx.numerators = make([]*symbolic.Expression, s.MaxNumOfQueryPerModule)
	permCtx.denominators = make([]*symbolic.Expression, s.MaxNumOfQueryPerModule)
	return &permCtx
}

func AddGdProductQuery(initialComp, moduleComp *wizard.CompiledIOP,
	targetModuleName modulediscoverer.ModuleName,
	s Settings, run *wizard.ProverRuntime) ifaces.Query {
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
					logrus.Printf("moduleNameA = %v, moduleNameB = %v", moduleNameA, moduleNameB)
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
	qId := ifaces.QueryIDf(string(targetModuleName) + "_GRAND_PRODUCT")
	G := moduleComp.InsertGrandProduct(0, qId, permCtx.numerators, permCtx.denominators)
	moduleComp.RegisterProverAction(1, permCtx.AssignParam(run, qId))
	return G
}

// The below function does the following:
// 1. Register beta and alpha (for the random linear combination in case A and B are multi-columns) in the compiledIop
// 2. Populates the nemerators and the denominators of the grand product query
func (p *PermutationIntoGrandProductCtx) push(comp *wizard.CompiledIOP, q *query.Permutation, round, queryInRound int, isNumerator, isBoth bool) {
	logrus.Printf("queryInRound %d", queryInRound)
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
		numFrag    = len(aOrB)
		factor     = symbolic.NewConstant(1)
		fragFactor = symbolic.NewConstant(1)
	)

	for frag := range numFrag {
		if isMultiColumn {
			fragFactor = wizardutils.RandLinCombColSymbolic(alpha, aOrB[frag])
		} else {
			fragFactor = ifaces.ColumnAsVariable(aOrB[frag][0])
		}
		fragFactor = symbolic.Add(fragFactor, beta)
		factor = symbolic.Mul(factor, fragFactor)
	}
	return factor
}

func (p *PermutationIntoGrandProductCtx) AssignParam(run *wizard.ProverRuntime, name ifaces.QueryID) *PermutationIntoGrandProductCtx {
	var (
		numNumerators   = len(p.numerators)
		numDenominators = len(p.denominators)
		numProd         = symbolic.NewConstant(1)
		denProd         = symbolic.NewConstant(1)
	)

	for i := 0; i < numNumerators; i++ {
		numProd = symbolic.Mul(numProd, p.numerators[i])
	}
	for j := 0; j < numDenominators; j++ {
		denProd = symbolic.Mul(denProd, p.denominators[j])
	}
	numProdFrVec := column.EvalExprColumn(run, numProd.Board()).IntoRegVecSaveAlloc()
	denProdFrVec := column.EvalExprColumn(run, denProd.Board()).IntoRegVecSaveAlloc()
	numProdFr := numProdFrVec[0]
	denProdFr := denProdFrVec[0]
	if len(numProdFrVec) > 1 {
		for i := 1; i < len(numProdFrVec); i++ {
			numProdFr.Mul(&numProdFr, &numProdFrVec[i])
		}
	}
	if len(numProdFrVec) > 1 {
		for j := 1; j < len(denProdFrVec); j++ {
			denProdFr.Mul(&denProdFr, &denProdFrVec[j])
		}
	}
	denProdFr.Inverse(&denProdFr)
	Y := numProdFr.Mul(&numProdFr, &denProdFr)
	run.AssignGrandProduct(name, *Y)
	return p
}

func (p *PermutationIntoGrandProductCtx) Run(run *wizard.ProverRuntime) {
	panic("unimplemented")
}
