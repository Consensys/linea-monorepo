package statesummary

import (
	"github.com/consensys/zkevm-monorepo/prover/protocol/column"
	"github.com/consensys/zkevm-monorepo/prover/protocol/ifaces"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizardutils"
	sym "github.com/consensys/zkevm-monorepo/prover/symbolic"
)

// lookBackDeltaAddressIsZeroCtx is a compilation routine constructing a binary
// column indicating whether an input column is zero. The object also stores the
// intermediate columns needed to fully constrain the IsZero column.
//
// The context does not ensure that the column are cancelled when the inactive
// flag is set. This is because these columns already serves as a basis to
// ensure that the inactive flag is set.
type lookBackDeltaAddressIsZeroCtx struct {
	// IsZero is the constructed column that the caller can use once it is
	// constructed and constrained.
	IsZero ifaces.Column
	// inverseOrZero is a by-product of the compilation routine it contains the
	// inverse of the input column or zero if the input column is zero.
	//
	// NB: this column should not be used outside of the scope of this function
	// because it is not actually constrained to be either the inverse or zero.
	// Only IsZero can be safely used outside.
	inverseOrZero ifaces.Column
}

// LookBackDeltaAddressIsZeroCtx creates a new context to construct a column
// from the input column (expected to be ss.AccountAddress)
// The context ensures that
//
//	IsZero == 1 <==> accAddress[i] == accAddress[i-1]
//		and
//	IsZero == 0 <==> accAddress[i] != accAddress[i-1]
//
// It also ensures that the first row is always set to zero as boundary
// condition.
func LookBackDeltaAddressIsZeroCtx(
	comp *wizard.CompiledIOP,
	accAddress ifaces.Column,
	isActive ifaces.Column,
) lookBackDeltaAddressIsZeroCtx {

	var (
		// This is a trick to convert input into an expression in case a column is
		// provided. The operation is a no-op formally speaking.
		expr   = sym.Sub(accAddress, column.Shift(accAddress, -1))
		board  = expr.Board()
		size   = wizardutils.ExprIsOnSameLengthHandles(&board)
		isZero = comp.InsertCommit(
			0,
			ifaces.ColIDf("STATE_SUMMARY_LOOK_BACK_DELTA_ADDRESS_IS_ZERO"),
			size,
		)
		inverseOrZero = comp.InsertCommit(
			0,
			ifaces.ColIDf("STATE_SUMMARY_LOOK_BACK_DELTA_ADDRESS_INVERSE_OR_ZERO"),
			size,
		)
	)

	comp.InsertGlobal(
		0,
		ifaces.QueryIDf("STATE_SUMMARY_LOOK_BACK_DELTA_ADDRESS_IS_ZERO_RES_IS_ONE_IF_INPUT_ISZERO"),
		sym.Add(isZero, sym.Mul(inverseOrZero, expr), sym.Neg(isActive)),
	)

	comp.InsertGlobal(
		0,
		ifaces.QueryIDf("STATE_SUMMARY_LOOK_BACK_DELTA_ADDRESS_IS_ZERO_RES_IS_ZERO_IF_INPUT_IS_NON_ZERO"),
		sym.Mul(expr, isZero),
	)

	comp.InsertGlobal(
		0,
		ifaces.QueryIDf("STATE_SUMMARY_LBDA_IS_ZERO_WHEN_INACTIVE"),
		sym.Sub(isZero, sym.Mul(isActive, isZero)),
	)

	comp.InsertGlobal(
		0,
		ifaces.QueryIDf("STATE_SUMMARY_LBDA_INVERSE_OR_ZERO_WHEN_INACTIVE"),
		sym.Sub(inverseOrZero, sym.Mul(isActive, inverseOrZero)),
	)

	comp.InsertLocal(
		0,
		ifaces.QueryIDf("STATE_SUMMARY_LBDA_IS_ZERO_FIRST_POSITION"),
		sym.NewVariable(isZero),
	)

	return lookBackDeltaAddressIsZeroCtx{
		IsZero:        isZero,
		inverseOrZero: inverseOrZero,
	}
}
