package common

import (
	"github.com/consensys/linea-monorepo/prover/crypto/mimc"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/column/verifiercol"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/protocol/wizardutils"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/parallel"
)

// HashingCtx is the wizard context responsible for hashing a table row-wise.
// It implements the [wizard.ProverAction] interface and is used to assign the
// InputColumns field.
type hashingCtx struct {
	// InputCols is the list of columns forming a table for which the current
	// context is computing the rows.
	InputCols []ifaces.Column
	// IntermediateHashes stores the intermediate values of the hasher.
	IntermediateHashes []ifaces.Column
}

// HashOf returns an [ifaces.Column] object containing the hash of the inputs
// columns and a [wizard.ProverAction] object responsible for assigning all
// the column taking part in justifying the returned column as well as the
// returned column itself.
func HashOf(comp *wizard.CompiledIOP, inputCols []ifaces.Column) (ifaces.Column, wizard.ProverAction) {

	var (
		ctx = &hashingCtx{
			InputCols:          inputCols,
			IntermediateHashes: make([]ifaces.Column, len(inputCols)),
		}
		round     = wizardutils.MaxRound(inputCols...)
		ctxID     = len(comp.ListCommitments())
		numRows   = ifaces.AssertSameLength(inputCols...)
		prevState = verifiercol.NewConstantCol(field.Zero(), numRows)
	)

	for i := range ctx.IntermediateHashes {
		ctx.IntermediateHashes[i] = comp.InsertCommit(
			round,
			ifaces.ColIDf("HASHING_%v_%v", ctxID, i),
			numRows,
		)

		comp.InsertMiMC(
			round,
			ifaces.QueryIDf("HASHING_%v_%v", ctxID, i),
			inputCols[i], prevState, ctx.IntermediateHashes[i],
		)

		prevState = ctx.IntermediateHashes[i]
	}

	return ctx.IntermediateHashes[len(inputCols)-1], ctx
}

// Run implements the [wizard.ProverAction] interface
func (ctx *hashingCtx) Run(run *wizard.ProverRuntime) {

	var (
		numRow = ctx.InputCols[0].Size()
		numCol = len(ctx.InputCols)
		inputs = make([][]field.Element, numCol)
		interm = make([][]field.Element, numCol)
	)

	for i := range interm {
		inputs[i] = ctx.InputCols[i].GetColAssignment(run).IntoRegVecSaveAlloc()
		interm[i] = make([]field.Element, numRow)
	}

	parallel.Execute(numRow, func(start, stop int) {
		prevState := make([]field.Element, stop-start)
		for i := range interm {
			mimcVecCompression(prevState, inputs[i][start:stop], interm[i][start:stop])
			prevState = interm[i][start:stop]
		}
	})

	for i := range interm {
		run.AssignColumn(
			ctx.IntermediateHashes[i].GetColID(),
			smartvectors.NewRegular(interm[i]),
		)
	}
}

func mimcVecCompression(oldState, block, newState []field.Element) {

	if len(oldState) != len(block) || len(block) != len(newState) {
		utils.Panic("the lengths are inconsistent: %v %v %v", len(oldState), len(block), len(newState))
	}

	if len(oldState) == 0 {
		return
	}

	for i := range oldState {
		newState[i] = mimc.BlockCompression(oldState[i], block[i])
	}
}
