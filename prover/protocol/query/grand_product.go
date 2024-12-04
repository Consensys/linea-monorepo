package query

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
)

type GrandProduct struct {
	ID          ifaces.QueryID
	Numerator   [][]ifaces.Column // stores A as multi-column
	Denominator [][]ifaces.Column // stores B as multi-column
	Alpha       coin.Info         // randomness for random linear combination in case of multi-column, to be provided by the
	// randomness beacon
	Beta coin.Info // randomness for the grand product accumulation, to be provided by the
	// randomness beacon
	Z     symbolic.Expression // aimed at storing the expressions (Ai + \beta_i)/(Bi + \beta_i)
	Round int
}

func NewGrandProduct(id ifaces.QueryID, numerator, denominator [][]ifaces.Column, alpha, beta coin.Info, round int) GrandProduct {
	return GrandProduct{
		ID:          id,
		Numerator:   numerator,
		Denominator: denominator,
		Alpha:       alpha,
		Beta:        beta,
		Round:       round}

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
