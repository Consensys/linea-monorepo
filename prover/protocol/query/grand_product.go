package query

import (
	"fmt"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
)

// The GrandProduct query enables splitting of the Permutation query between sub-provers
// by splitting the grand product Z = \prod(A_i+\beta)/(B_i+\beta) itself.
type GrandProduct struct {
	ID           ifaces.QueryID
	Numerators   []*symbolic.Expression // stores A as multi-column
	Denominators []*symbolic.Expression // stores B as multi-column
	Round        int
}

func NewGrandProduct(round int, id ifaces.QueryID, numerators, denominators []*symbolic.Expression) GrandProduct {
	return GrandProduct{
		ID:           id,
		Numerators:   numerators,
		Denominators: denominators,
		Round:        round,
	}
}

func (g GrandProduct) Name() ifaces.QueryID {
	return g.ID
}

func (g GrandProduct) Check(run ifaces.Runtime) error {
	var (
		numNumerators   = len(g.Numerators)
		numDenominators = len(g.Denominators)
		numProd         = symbolic.NewConstant(1)
		denProd         = symbolic.NewConstant(1)
	)

	for i := 0; i < numNumerators; i++ {
		numProd = symbolic.Mul(numProd, g.Numerators[i])
	}
	for j := 0; j < numDenominators; j++ {
		denProd = symbolic.Mul(denProd, g.Denominators[j])
	}
	if numProd != denProd {
		return fmt.Errorf("the grand product query %v is not satisfied, numProd = %v, denProd = %v", g.ID, numProd, denProd)
	}

	return nil
}

func (g GrandProduct) CheckGnark(api frontend.API, run ifaces.GnarkRuntime) {
	utils.Panic("Unimplemented")
}
