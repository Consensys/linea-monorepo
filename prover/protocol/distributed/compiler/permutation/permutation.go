package dist_permutation

import (
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	modulediscoverer "github.com/consensys/linea-monorepo/prover/protocol/distributed/module_discoverer"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/protocol/wizardutils"
	"github.com/consensys/linea-monorepo/prover/symbolic"
)

// Used for deriving names of queries and coins
const grandProductStr = "GRAND_PRODUCT"

/*
The below function aims to process all the permuation queries specific to a target module
into a grand product query. We store the randomised symbolic products of A and B of permuation
queries combinedly into the Numerators and the Denominators of the GrandProduct query
*/
type PermutationIntoGrandProductCtx struct {
	// stores the expressions Ai + \beta_i for the ith permutation query, Ai is a linear combination with alpha_i
	// in case of multi-column
	Numerators []*symbolic.Expression
	// stores the expressions Bi + \beta_i for the ith permutation query, Bi is a linear combination with alpha_i
	// in case of multi-column
	Denominators []*symbolic.Expression
	// stores the field element obtained by collapsing the expression
	// \prod(Numerators)/\prod(Denominators)
	ParamY field.Element
	// The query id specific to the target module
	QueryId ifaces.QueryID
	// The module name for which we are processing the grand product query
	TargetModuleName string
}

// Returns a new PermutationIntoGrandProductCtx
func NewPermutationIntoGrandProductCtx(s Settings) *PermutationIntoGrandProductCtx {
	permCtx := PermutationIntoGrandProductCtx{}
	permCtx.Numerators = make([]*symbolic.Expression, 0, s.MaxNumOfQueryPerModule)
	permCtx.Denominators = make([]*symbolic.Expression, 0, s.MaxNumOfQueryPerModule)
	return &permCtx
}

// AddGdProductQuery processes all permutation queries specific to a target module
// into a grand product query. It stores the randomised symbolic products of A and B
// of permutation queries combinedly into the Numerators and the Denominators of the
// GrandProduct query.
//
// Parameters:
// - initialComp: The initial compiledIOP
// - moduleComp: The compiledIOP for the target module
// - targetModuleName: The name of the module for which we are processing the grand product query
// - run: The prover runtime
//
// Returns:
// - An instance of the grand product query (ifaces.Query)
func (p *PermutationIntoGrandProductCtx) AddGdProductQuery(initialComp, moduleComp *wizard.CompiledIOP,
	targetModuleName modulediscoverer.ModuleName, run *wizard.ProverRuntime) ifaces.Query {
	numRounds := initialComp.NumRounds()
	// Initialise the period separating module discoverer
	disc := modulediscoverer.PeriodSeperatingModuleDiscoverer{}
	disc.Analyze(initialComp)
	qId := deriveName[ifaces.QueryID](ifaces.QueryID(targetModuleName))
	p.QueryId = qId
	p.TargetModuleName = string(targetModuleName)
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
						p.push(moduleComp, &q_, i, j, true, false)
					} else if moduleNameA != targetModuleName && moduleNameB == targetModuleName {
						p.push(moduleComp, &q_, i, j, false, false)
					} else if moduleNameA == targetModuleName && moduleNameB == targetModuleName {
						p.push(moduleComp, &q_, i, j, true, true)
					} else {
						continue
					}
				}
			default:
				continue
			}
		}
	}
	// We register the grand product query in round one because
	// alphas, betas, and the query param are assigned in round one
	G := moduleComp.InsertGrandProduct(1, qId, p.Numerators, p.Denominators)
	// The below prover action is responsible for computing the query parameter
	// and assign it in round one
	moduleComp.RegisterProverAction(1, p.AssignParam(run, qId))
	return G
}

// push processes a permutation query and adds its symbolic factors to the Numerators or Denominators
// based on the provided flags. It also inserts coins for alpha and beta for each permutation query.
//
// Parameters:
// - comp: The compiled IOP for the target module
// - q: The permutation query to be processed
// - round: The round number of the permutation query
// - queryInRound: The index of the permutation query within the round
// - isNumerator: A flag indicating whether to add the symbolic factor to the Numerators
// - isBoth: A flag indicating whether we need to add the symbolic factor to both Numerators and Denominators
func (p *PermutationIntoGrandProductCtx) push(comp *wizard.CompiledIOP, q *query.Permutation, round, queryInRound int, isNumerator, isBoth bool) {
	var (
		isMultiColumn = len(q.A[0]) > 1
		alpha         coin.Info
		beta          coin.Info
	)
	if isMultiColumn {
		// alpha has to be different for different queries for a particular round for the soundness of z-packing
		alpha = comp.InsertCoin(1, deriveName[coin.Name](ifaces.QueryID(p.TargetModuleName), "ALPHA", round, queryInRound), coin.Field)
	}
	// beta has to be different for different queries for a particular round for the soundness of z-packing
	beta = comp.InsertCoin(1, deriveName[coin.Name](ifaces.QueryID(p.TargetModuleName), "BETA", round, queryInRound), coin.Field)
	if isNumerator && !isBoth {
		// Take only the numerator
		factor := computeFactor(q.A, isMultiColumn, &alpha, &beta)
		p.Numerators = append(p.Numerators, factor)
	} else if !isNumerator && !isBoth {
		// Take only the denominator
		factor := computeFactor(q.B, isMultiColumn, &alpha, &beta)
		p.Denominators = append(p.Denominators, factor)
	} else if isNumerator && isBoth {
		// Take both the numerator and the denominator
		numFactor := computeFactor(q.A, isMultiColumn, &alpha, &beta)
		denFactor := computeFactor(q.B, isMultiColumn, &alpha, &beta)
		p.Numerators = append(p.Numerators, numFactor)
		p.Denominators = append(p.Denominators, denFactor)
	} else if !isNumerator && isBoth {
		panic("Invalid case")
	}
}

// computeFactor computes the symbolic factor for a permutation query based on the given parameters.
// It iterates through the fragments of the query, computes the linear combination of columns with alpha
// (if multi-column) or directly uses the column as a variable, adds the beta value, and multiplies it with
// the current factor. The final computed factor is returned.
//
// Parameters:
// - aOrB: A 2D slice of Column interfaces representing the fragments of the permutation query.
// - isMultiColumn: A boolean indicating whether the permutation query is multi-column.
// - alpha: A pointer to a CoinInfo struct representing the alpha coin for the permutation query.
// - beta: A pointer to a CoinInfo struct representing the beta coin for the permutation query.
//
// Returns:
// - A pointer to a symbolic.Expression representing the computed factor for the permutation query.
func computeFactor(aOrB [][]ifaces.Column, isMultiColumn bool, alpha, beta *coin.Info) *symbolic.Expression {
    var (
        numFrag    = len(aOrB)
        factor     = symbolic.NewConstant(1)
        fragFactor = symbolic.NewConstant(1)
    )

    for frag := range numFrag {
        if isMultiColumn {
            fragFactor = wizardutils.RandLinCombColSymbolic(*alpha, aOrB[frag])
        } else {
            fragFactor = ifaces.ColumnAsVariable(aOrB[frag][0])
        }
        fragFactor = symbolic.Add(fragFactor, *beta)
        factor = symbolic.Mul(factor, fragFactor)
    }
    return factor
}

// AssignParam computes the query parameter for the grand product query and assigns it in round one.
// It multiplies the products of the Numerators and Denominators, evaluates the resulting symbolic expressions,
// and assigns the result to the field element ParamY.
//
// Parameters:
// - run: The prover runtime.
// - name: The query ID specific to the target module.
//
// Returns:
// - A pointer to the PermutationIntoGrandProductCtx instance with the updated ParamY field.
func (p *PermutationIntoGrandProductCtx) AssignParam(run *wizard.ProverRuntime, name ifaces.QueryID) *PermutationIntoGrandProductCtx {
    var (
        numNumerators   = len(p.Numerators)
        numDenominators = len(p.Denominators)
        numProd         = symbolic.NewConstant(1)
        denProd         = symbolic.NewConstant(1)
    )
    // Multiply all Numerators
    for i := 0; i < numNumerators; i++ {
        numProd = symbolic.Mul(numProd, p.Numerators[i])
    }
    // Multiply all Denominators
    for j := 0; j < numDenominators; j++ {
        denProd = symbolic.Mul(denProd, p.Denominators[j])
    }
    // Evaluate the symbolic expressions for Numerator and Denominator products
    numProdFrVec := column.EvalExprColumn(run, numProd.Board()).IntoRegVecSaveAlloc()
    denProdFrVec := column.EvalExprColumn(run, denProd.Board()).IntoRegVecSaveAlloc()
    numProdFr := numProdFrVec[0]
    denProdFr := denProdFrVec[0]
    // Multiply all field elements in the Numerator product vector
    if len(numProdFrVec) > 1 {
        for i := 1; i < len(numProdFrVec); i++ {
            numProdFr.Mul(&numProdFr, &numProdFrVec[i])
        }
    }
    // Multiply all field elements in the Denominator product vector
    if len(denProdFrVec) > 1 {
        for j := 1; j < len(denProdFrVec); j++ {
            denProdFr.Mul(&denProdFr, &denProdFrVec[j])
        }
    }
    // Invert the Denominator product field element
    denProdFr.Inverse(&denProdFr)
    // Compute the final query parameter Y
    Y := numProdFr.Mul(&numProdFr, &denProdFr)
    p.ParamY = *Y
    return p
}

// Run executes the grand product query by assigning the computed parameter Y to the prover runtime.
//
// Parameters:
// - run: The prover runtime where the grand product query will be executed.
//
// The function does not return any value. It directly assigns the computed parameter Y to the prover runtime
// using the AssignGrandProduct method of the runtime.
func (p *PermutationIntoGrandProductCtx) Run(run *wizard.ProverRuntime) {
    run.AssignGrandProduct(p.QueryId, p.ParamY)
}

// deriveName constructs a name for the PermutationIntoGrandProduct context
func deriveName[R ~string](q ifaces.QueryID, ss ...any) R {
	ss = append([]any{grandProductStr, q}, ss...)
	return wizardutils.DeriveName[R](ss...)
}
