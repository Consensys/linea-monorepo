package query

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
)

// The GrandProduct query enables splitting of the Permutation query between sub-provers
// by splitting the grand product Z = \prod(A_i+\beta)/(B_i+\beta) itself.
type GrandProduct struct {
	ID          ifaces.QueryID
	Numerator   *symbolic.Expression // stores A as multi-column
	Denominator *symbolic.Expression // stores B as multi-column
	Round       int
}

func NewGrandProduct(round int, id ifaces.QueryID, numerator, denominator *symbolic.Expression) GrandProduct {
	return GrandProduct{
		ID:          id,
		Numerator:   numerator,
		Denominator: denominator,
		Round:       round,
	}
}

func (g GrandProduct) Name() ifaces.QueryID {
	return g.ID
}

func (g GrandProduct) Check(run ifaces.Runtime) error {
	utils.Panic("Unimplemented")
	return nil
}

func (g GrandProduct) CheckGnark(api frontend.API, run ifaces.GnarkRuntime) {
	utils.Panic("Unimplemented")
}
