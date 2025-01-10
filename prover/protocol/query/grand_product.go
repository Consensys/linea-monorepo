package query

import (
	"fmt"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/crypto/fiatshamir"
	"github.com/consensys/linea-monorepo/prover/maths/common/vector"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
)

// The GrandProduct query is obtained by processing all the permuation queries specific to a target module.
// We store the randomised symbolic products of A and B of permuation queries combinedly
// into the Numerators and the Denominators of the GrandProduct query
type GrandProductInput struct {
	Size         int
	Numerators   []*symbolic.Expression // stores A as multi-column
	Denominators []*symbolic.Expression // stores B as multi-column
}

// GrandProduct is a query for computing the grand-product of several vector expressions. The
// query returns a unique field element result.
type GrandProduct struct {
	Round int
	ID    ifaces.QueryID
	// The list of the inputs of the query, grouped by sizes
	Inputs map[int]*GrandProductInput
}

type GrandProductParams struct {
	Y field.Element
}

// NewGrandProduct creates a new instance of a GrandProduct query.
//
// Parameters:
// - round: The round number of the query.
// - id: The unique identifier for the query.
// - numerators: A slice of symbolic expressions representing the numerators of the permutation queries.
// - denominators: A slice of symbolic expressions representing the denominators of the permutation queries.
//
// Returns:
// - A pointer to a new instance of GrandProduct.
func NewGrandProduct(round int, inp map[int]*GrandProductInput, id ifaces.QueryID) *GrandProduct {
	// check the length consistency
	for key := range inp {
		for i := range inp[key].Numerators {
			if err := inp[key].Numerators[i].Validate(); err != nil {
				utils.Panic(" Numerator[%v] is not a valid expression", i)
			}
			if err := inp[key].Denominators[i].Validate(); err != nil {
				utils.Panic(" Denominator[%v] is not a valid expression", i)
			}
		}
	}

	return &GrandProduct{
		Round:  round,
		Inputs: inp,
		ID:     id,
	}
}

// Constructor for grand product query parameters
func NewGrandProductParams(y field.Element) GrandProductParams {
	return GrandProductParams{Y: y}
}

// Name returns the unique identifier of the GrandProduct query.
func (g GrandProduct) Name() ifaces.QueryID {
	return g.ID
}

// Updates a Fiat-Shamir state
func (gp GrandProductParams) UpdateFS(fs *fiatshamir.State) {
	fs.Update(gp.Y)
}

// Check verifies the satisfaction of the GrandProduct query using the provided runtime.
// It calculates the product of numerators and denominators, and checks
// if prod(Numerators) == Prod(Denominators)*ParamY, and returns an error if the condition is not satisfied.
//
// Parameters:
// - run: The runtime interface providing access to the query parameter for query verification.
//
// Returns:
// - An error if the grand product query is not satisfied, or nil if it is satisfied.
func (g *GrandProduct) Check(run ifaces.Runtime) error {
	params := run.GetParams(g.ID).(GrandProductParams)
	actualProd := field.One()
	for key := range g.Inputs {
		for i, num := range g.Inputs[key].Numerators {

			var (
				numBoard          = num.Board()
				denBoard          = g.Inputs[key].Denominators[i].Board()
				numeratorMetadata = numBoard.ListVariableMetadata()
				denominator       = column.EvalExprColumn(run, denBoard).IntoRegVecSaveAlloc()
				numerator         []field.Element
				packedZ           = field.BatchInvert(denominator)
			)

			if len(numeratorMetadata) == 0 {
				numerator = vector.Repeat(field.One(), g.Inputs[key].Size)
			}

			if len(numeratorMetadata) > 0 {
				numerator = column.EvalExprColumn(run, numBoard).IntoRegVecSaveAlloc()
			}

			for k := range packedZ {
				packedZ[k].Mul(&numerator[k], &packedZ[k])
				if k > 0 {
					packedZ[k].Mul(&packedZ[k], &packedZ[k-1])
				}
			}
			actualProd.Mul(&actualProd, &packedZ[len(packedZ)-1])
		}
	}
	if actualProd != params.Y {
		return fmt.Errorf("the grand product query %v is not satisfied, actualProd = %v, param.Y = %v", g.ID, actualProd, params.Y)
	}

	return nil
}

func (g GrandProduct) CheckGnark(api frontend.API, run ifaces.GnarkRuntime) {
	utils.Panic("Unimplemented")
}
