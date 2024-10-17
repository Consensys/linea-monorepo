package wizard

import (
	"fmt"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/zkevm-monorepo/prover/maths/field"
	"github.com/consensys/zkevm-monorepo/prover/utils"
)

// Enforces a range constraint, that all elements in the handle must
// be within range [0, B)
//
// Where B is a power of two
type QueryRange struct {
	// Maybe we should enforce that the handle is a natural one here
	Col Column
	// Upper-bound
	B        int
	metadata *metadata
	*subQuery
}

// Constructor for range constraints also makes the input validation
func (api *API) NewQueryRange(col Column, b int) *QueryRange {
	res := &QueryRange{
		B:        b,
		Col:      col,
		metadata: api.newMetadata(),
		subQuery: &subQuery{
			round: col.Round(),
		},
	}
	api.queries.addToRound(col.Round(), res)
	return res
}

/*
Test that the range checks hold
*/
func (r QueryRange) Check(run Runtime) error {

	b := field.NewElement(uint64(r.B))

	if run == nil {
		panic("got a nil runtime")
	}

	if r.Col == nil {
		utils.Panic("handle was poisoned")
	}

	wit := r.Col.GetAssignment(run)
	for i := 0; i < wit.Len(); i++ {
		v := wit.Get(i)
		if v.Cmp(&b) >= 0 {
			return fmt.Errorf("range check failed %v (bound %v on %v)", r.String(), r.B, r.Col.String())
		}
	}

	return nil
}

// CheckGnark will panic in this construction because we do not have a good way
// to check the query within a circuit
func (r QueryRange) CheckGnark(api frontend.API, run RuntimeGnark) {
	panic("UNSUPPORTED : can't check an inclusion query directly into the circuit")
}
