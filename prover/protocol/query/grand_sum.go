package query

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/crypto/fiatshamir"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/utils"
)

// GrandSum is a query for computing the grand-sum of several vector expressions. The
// query returns a unique field element result.
type GrandSum struct {
	Round int
	ID    ifaces.QueryID
}

type GrandSumParams struct {
	Y field.Element
}

// NewGrandSum creates a new instance of a GrandSum query.
func NewGrandSum(round int, id ifaces.QueryID) GrandSum {

	return GrandSum{
		Round: round,
		ID:    id,
	}
}

// Constructor for grand product query parameters
func NewGrandSumParams(y field.Element) GrandSumParams {
	return GrandSumParams{Y: y}
}

// Name returns the unique identifier of the GrandSum query.
func (g GrandSum) Name() ifaces.QueryID {
	return g.ID
}

// Updates a Fiat-Shamir state
func (gp GrandSumParams) UpdateFS(fs *fiatshamir.State) {
	fs.Update(gp.Y)
}

// Check verifies the satisfaction of the GrandSum query using the provided runtime.
func (g GrandSum) Check(run ifaces.Runtime) error {
	utils.Panic("Unimplemented")
	return nil
}

func (g GrandSum) CheckGnark(api frontend.API, run ifaces.GnarkRuntime) {
	utils.Panic("Unimplemented")
}
