package codegen

import (
	"github.com/consensys/linea-monorepo/prover-ray/wiop"
	"github.com/consensys/linea-monorepo/prover-ray/wiop/compilers/logderivativesum"
	"github.com/consensys/linea-monorepo/prover-ray/wiop/compilers/lookuptologderivsum"
)

// LogDerivSystem is the compiled metadata for every LogDerivativeSum query in a
// wiop.System, in the form the Zig logderivativesum sub-verifier consumes.
//
// The Z-recurrence and the L_0 initial condition are ordinary vanishing
// constraints (registered by the logderivativesum compiler), so they are
// discharged by the vanishing sub-verifier. What remains for this sub-verifier
// is the boundary identity the compiler's verifier action enforces:
//
//	Σ_entries Z[n-1] == Result        (and, for lookups, Result == 0)
//
// All operands are cell openings, so no expression evaluation is needed.
type LogDerivSystem struct {
	SourceName string
	Queries    []LogDerivQuery
}

// LogDerivQuery is one reduced LogDerivativeSum query: the endpoint openings of
// its Z columns and the claimed Result cell.
type LogDerivQuery struct {
	SourceName string
	// ZFinals are the openings of Z[n-1] for each Z column of the query.
	ZFinals []ScalarRef
	// Result is the query's claimed aggregated value.
	Result ScalarRef
	// ResultIsZero is set for lookup-reduced queries, whose Result must be 0.
	// Populated by the lookuptologderivsum verifier action (see Phase 3).
	ResultIsZero bool
}

// BuildLogDerivSystem extracts the LogDerivativeSum verifier actions registered
// on sys into a LogDerivSystem. Queries are collected in round/registration
// order so the output is deterministic.
func BuildLogDerivSystem(sys *wiop.System) (LogDerivSystem, error) {
	out := LogDerivSystem{SourceName: sys.Context.Path()}

	// First pass: collect the LogDerivativeSum queries that a lookup reduction
	// requires to be zero (lookuptologderivsum registers a ResultIsZero action
	// alongside the logderivativesum reduction).
	resultMustBeZero := map[*wiop.LogDerivativeSum]bool{}
	for _, round := range sys.Rounds {
		for _, action := range round.VerifierActions {
			if la, ok := action.(*lookuptologderivsum.ResultIsZeroVerifierAction); ok {
				resultMustBeZero[la.Ld] = true
			}
		}
	}

	for _, round := range sys.Rounds {
		for _, action := range round.VerifierActions {
			va, ok := action.(*logderivativesum.VerifierAction)
			if !ok {
				continue
			}
			query := LogDerivQuery{
				SourceName:   va.Ld.Context().Path(),
				Result:       cellScalarRef(va.Ld.Result),
				ZFinals:      make([]ScalarRef, len(va.Entries)),
				ResultIsZero: resultMustBeZero[va.Ld],
			}
			for i, e := range va.Entries {
				query.ZFinals[i] = cellScalarRef(e.ZFinal)
			}
			out.Queries = append(out.Queries, query)
		}
	}

	return out, nil
}

// cellScalarRef converts a wiop.Cell opening into the (round, index) reference
// the Zig verifier reads from ctx.rounds[round].cells[index]. Mirrors the
// *wiop.Cell case of appendExpr in vanishing.go.
func cellScalarRef(c *wiop.Cell) ScalarRef {
	return ScalarRef{
		Round:      c.Context.ID.Slot(),
		Index:      c.Context.ID.Position(),
		SourceName: c.Context.Label,
	}
}
