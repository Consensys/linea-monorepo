package query

import (
	"fmt"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/crypto/fiatshamir"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
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

type GrandProductParams struct {
	Y field.Element
}

func NewGrandProduct(round int, id ifaces.QueryID, numerators, denominators []*symbolic.Expression) *GrandProduct {
	return &GrandProduct{
		ID:           id,
		Numerators:   numerators,
		Denominators: denominators,
		Round:        round,
	}
}

// Constructor for grand product query parameters
func NewGrandProductParams(y field.Element) GrandProductParams {
	return GrandProductParams{Y: y}
}

func (g GrandProduct) Name() ifaces.QueryID {
	return g.ID
}

// Updates a Fiat-Shamir state
func (gp GrandProductParams) UpdateFS(fs *fiatshamir.State) {
	fs.Update(gp.Y)
}

func (g *GrandProduct) Check(run ifaces.Runtime) error {
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
	params := run.GetParams(g.ID).(GrandProductParams)
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
	if numProdFr != *denProdFr.Mul(&denProdFr, &params.Y) {
		return fmt.Errorf("the grand product query %v is not satisfied, numProd = %v, denProd = %v, witness = %v", g.ID, numProdFr, denProdFr, params.Y)
	}

	return nil
}

func (g GrandProduct) CheckGnark(api frontend.API, run ifaces.GnarkRuntime) {
	utils.Panic("Unimplemented")
}