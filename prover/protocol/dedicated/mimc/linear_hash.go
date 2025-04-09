package mimc

import (
	"fmt"

	"github.com/consensys/linea-monorepo/prover/crypto/mimc"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/common/vector"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/dedicated/selector"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/variables"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/parallel"
)

/*
Holds intermediate values for the linear hash wizard
*/
type linearHashCtx struct {
	// The compiled IOP
	comp *wizard.CompiledIOP

	name string

	// Output column, which containing the result each
	// individual hash.
	ExpectedHash ifaces.Column
	// Column describing the hashing of each chunk
	ToHash, OldState, NewState, NewStateClean ifaces.Column

	// Name of the optional columns

	// IsFullyActive is a boolean flag indicating whether
	// the columns are entirely used in for hashing.
	// Name of the created columns
	IsFullyActive bool
	// Optional precomputed column indicating
	// if a row is active or not. And when a
	// hash is finished
	IsActiveLarge, isActiveExpected, IsEndOfHash, Ones ifaces.Column

	// Period of the linear hash
	Period, NumHash int
	// Round of declaration of the linear hash
	Round int
}

/*
Check a linear hashby chunk of columns

  - tohash : (a, b, c, d) || (e, f, g, h)

  - period : 4 // indicates that we should hash chunks of 4 entries

  - expectedhash : (Hash(a, b, c, d) || Hash(e, f, g, h))
*/
func CheckLinearHash(
	comp *wizard.CompiledIOP,
	name string,
	tohash ifaces.Column,
	period int, numHash int,
	expectedHashes ifaces.Column,
) {

	// Initialize the context
	ctx := linearHashCtx{
		comp:         comp,
		name:         name,
		ToHash:       tohash,
		Period:       period,
		NumHash:      numHash,
		ExpectedHash: expectedHashes,
	}

	// Get the rounds
	ctx.Round = utils.Max(tohash.Round(), expectedHashes.Round())
	ctx.IsFullyActive = utils.IsPowerOfTwo(numHash * period)

	ctx.HashingCols()

	if ctx.IsFullyActive {
		selector.CheckSubsample(
			comp,
			prefixWithLinearHash(comp, name, "RES_EXTRACTION"),
			[]ifaces.Column{ctx.NewStateClean},
			[]ifaces.Column{ctx.ExpectedHash},
			period-1,
		)
	} else {
		ctx.comp.InsertInclusion(
			ctx.Round,
			ifaces.QueryID(prefixWithLinearHash(comp, name, "RESULT_CHECK_%v", tohash.GetColID())),
			[]ifaces.Column{ctx.IsEndOfHash, ctx.NewStateClean},
			[]ifaces.Column{ctx.IsActiveExpected(), ctx.ExpectedHash},
		)
	}

}

func prefixWithLinearHash(comp *wizard.CompiledIOP, name, msg string, args ...any) string {
	args = append([]any{name, comp.SelfRecursionCount}, args...)
	return fmt.Sprintf("%v.LINEAR_HASH_%v_"+msg, args...)
}

// Declares assign and constraints the columns OldStates and NewStates
func (ctx *linearHashCtx) HashingCols() {

	// Registers the old states columns
	ctx.OldState = ctx.comp.InsertCommit(
		ctx.Round,
		ifaces.ColID(prefixWithLinearHash(ctx.comp, ctx.name, "OLD_STATE_%v", ctx.ToHash.GetColID())),
		ctx.ToHash.Size(),
	)

	ctx.NewState = ctx.comp.InsertCommit(
		ctx.Round,
		ifaces.ColID(prefixWithLinearHash(ctx.comp, ctx.name, "NEW_STATE_%v", ctx.ToHash.GetColID())),
		ctx.ToHash.Size(),
	)

	ctx.NewStateClean = ctx.comp.InsertCommit(
		ctx.Round,
		ifaces.ColIDf(prefixWithLinearHash(ctx.comp, ctx.name, "NEW_STATE_CLEAN_%v", ctx.ToHash.GetColID())),
		ctx.ToHash.Size(),
	)

	ctx.comp.SubProvers.AppendToInner(
		ctx.Round,
		func(run *wizard.ProverRuntime) {
			// Extract the blocks
			blocksWit := ctx.ToHash.GetColAssignment(run)

			olds := make([]field.Element, ctx.Period*ctx.NumHash)
			news := make([]field.Element, ctx.Period*ctx.NumHash)

			// Assign the hashes in parallel
			parallel.Execute(ctx.NumHash, func(start, stop int) {
				for hashID := start; hashID < stop; hashID++ {
					// each hash start from zero
					old := field.Zero()
					for i := 0; i < ctx.Period; i++ {
						pos := hashID*ctx.Period + i
						currentBlock := blocksWit.Get(pos)
						new := mimc.BlockCompression(old, currentBlock)
						olds[pos] = old
						news[pos] = new
						old = new
					}
				}
			})

			padNew := mimc.BlockCompression(field.Zero(), field.Zero())
			oldSV := smartvectors.RightZeroPadded(olds, ctx.ToHash.Size())
			newSV := smartvectors.RightPadded(news, padNew, ctx.ToHash.Size())
			newCleanSV := smartvectors.RightZeroPadded(vector.DeepCopy(news), ctx.ToHash.Size())

			// assign old state
			run.AssignColumn(ctx.OldState.GetColID(), oldSV)

			// assign new state
			run.AssignColumn(ctx.NewState.GetColID(), newSV)

			// and new clean, the same as newstate but clean
			run.AssignColumn(ctx.NewStateClean.GetColID(), newCleanSV)
		},
	)

	// And registers queries for the initial values

	//
	// Propagation of the state within the chunks
	//
	expr := ifaces.ColumnAsVariable(ctx.NewState).
		Mul(ctx.IsActiveVar()).
		Mul(ctx.IsNotEndOfHashVar()).
		Sub(ifaces.ColumnAsVariable(column.Shift(ctx.OldState, 1)))

	ctx.comp.InsertGlobal(
		ctx.Round,
		ifaces.QueryID(prefixWithLinearHash(ctx.comp, ctx.name, "STATE_PROPAGATION_%v", ctx.ToHash.GetColID())),
		expr,
		true, // no bound cancel to also enforce the first value of old state to be zero
	)

	//
	// Cleaning the new state
	//
	ctx.comp.InsertGlobal(
		ctx.Round,
		ifaces.QueryIDf(prefixWithLinearHash(ctx.comp, ctx.name, "CLEAN_NEW_STATE_%v", ctx.ToHash.GetColID())),
		ctx.IsActiveVar().
			Mul(ifaces.ColumnAsVariable(ctx.NewState)).
			Sub(ifaces.ColumnAsVariable(ctx.NewStateClean)),
	)

	//
	// Correctness of the blocks compressions
	//
	ctx.comp.InsertMiMC(
		ctx.Round,
		ifaces.QueryID(prefixWithLinearHash(ctx.comp, ctx.name, "BLOCKS_COMPRESSION_%v", ctx.ToHash.GetColID())),
		ctx.ToHash, ctx.OldState, ctx.NewState, nil,
	)

}

func (ctx *linearHashCtx) IsActiveExpected() ifaces.Column {
	if ctx.IsFullyActive {
		// Always active
		panic("asked isActiveExpected but the module is fully active")
	}

	// Lazily registers the columns
	if ctx.isActiveExpected == nil {

		var assignment smartvectors.SmartVector

		if utils.IsPowerOfTwo(ctx.NumHash) {
			assignment = smartvectors.NewConstant(field.One(), ctx.NumHash)
		} else {
			assignment = smartvectors.RightZeroPadded(
				vector.Repeat(field.One(), ctx.NumHash),
				ctx.ExpectedHash.Size(),
			)
		}

		ctx.isActiveExpected = ctx.comp.InsertPrecomputed(
			ifaces.ColIDf(prefixWithLinearHash(ctx.comp, ctx.name, "IS_ACTIVE_EXPECTED_%v", ctx.ToHash.GetColID())),
			assignment,
		)
	}

	return ctx.isActiveExpected
}

func (ctx *linearHashCtx) IsActiveVar() *symbolic.Expression {
	if ctx.IsFullyActive {
		// Always active
		return symbolic.NewConstant(1)
	}

	// Lazily registers the columns
	if ctx.IsActiveLarge == nil {
		ctx.IsActiveLarge = ctx.comp.InsertPrecomputed(
			ifaces.ColIDf(prefixWithLinearHash(ctx.comp, ctx.name, "IS_ACTIVE_%v", ctx.ToHash.GetColID())),
			smartvectors.RightZeroPadded(
				vector.Repeat(field.One(), ctx.NumHash*ctx.Period),
				ctx.ToHash.Size(),
			),
		)
	}

	return ifaces.ColumnAsVariable(ctx.IsActiveLarge)
}

func (ctx *linearHashCtx) IsEndOfHashVar() *symbolic.Expression {
	if ctx.IsFullyActive {
		// Always active
		return variables.NewPeriodicSample(ctx.Period, ctx.Period-1)
	}

	// Lazily registers the columns
	if ctx.IsEndOfHash == nil {

		window := make([]field.Element, ctx.NumHash*ctx.Period)
		for i := range window {
			if i%ctx.Period == ctx.Period-1 {
				window[i].SetOne()
			}
		}

		ctx.IsEndOfHash = ctx.comp.InsertPrecomputed(
			ifaces.ColIDf(prefixWithLinearHash(ctx.comp, ctx.name, "IS_END_OF_HASH_%v", ctx.ToHash.GetColID())),
			smartvectors.RightZeroPadded(window, ctx.ToHash.Size()),
		)
	}

	return ifaces.ColumnAsVariable(ctx.IsEndOfHash)
}

func (ctx *linearHashCtx) IsNotEndOfHashVar() *symbolic.Expression {
	return symbolic.NewConstant(1).Sub(ctx.IsEndOfHashVar())
}
