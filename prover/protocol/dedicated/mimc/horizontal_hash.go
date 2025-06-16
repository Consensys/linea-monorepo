package mimc

import (
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils/parallel"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/common"
)

// HashingCtx is the wizard context responsible for hashing a table row-wise.
// It implements the [wizard.ProverAction] interface and is used to assign the
// InputColumns field.
type HashingCtx struct {
	// InputCols is the list of column limbs forming a table for which the current
	// context is computing the rows.
	InputCols [][]ifaces.Column
	// IntermediateHashes stores the intermediate values of the hasher.
	IntermediateHashes [][common.NbLimbU256]ifaces.Column
}

// HashOf returns an [ifaces.Column] object containing the hash of the inputs
// columns and a [wizard.ProverAction] object responsible for assigning all
// the column taking part in justifying the returned column as well as the
// returned column itself.
func HashOf(comp *wizard.CompiledIOP, inputCols [][]ifaces.Column) *HashingCtx {

	var (
		ctx = &HashingCtx{
			InputCols:          inputCols,
			IntermediateHashes: make([][common.NbLimbU256]ifaces.Column, len(inputCols)),
		}
		round   = column.MaxRound(inputCols[0]...)
		ctxID   = len(comp.ListCommitments())
		numRows = 0
		// prevState = verifiercol.NewConstantCol(field.Zero(), numRows, strconv.Itoa(ctxID))
	)

	for i := range inputCols {
		numRows = ifaces.AssertSameLength(inputCols[i]...)
	}

	//var prevState = verifiercol.NewConstantCol(field.Zero(), numRows)
	for i, intermHashRow := range ctx.IntermediateHashes {
		for j := range intermHashRow {
			ctx.IntermediateHashes[i][j] = comp.InsertCommit(
				round,
				ifaces.ColIDf("HASHING_%v_%v_%v", ctxID, i, j),
				numRows,
			)
		}

		panic("add insert poseidon here, instead of MiMC")

		//comp.InsertMiMC(
		//	round,
		//	ifaces.QueryIDf("HASHING_%v_%v", ctxID, i),
		//	inputCols[i], prevState, ctx.IntermediateHashes[i],
		//	nil,
		//)

		//prevState = ctx.IntermediateHashes[i]
	}

	return ctx
}

// Result returns the limb columns containing the result of the hashing.
func (ctx *HashingCtx) Result() [common.NbLimbU256]ifaces.Column {
	return ctx.IntermediateHashes[len(ctx.IntermediateHashes)-1]
}

// Run implements the [wizard.ProverAction] interface
func (ctx *HashingCtx) Run(run *wizard.ProverRuntime) {

	var (
		numRow = ctx.InputCols[0][0].Size()
		numCol = len(ctx.InputCols)
		// 1st - hashing input, 2nd - input limbs
		inputs = make([][]smartvectors.SmartVector, numCol)
		// 1st - hashing input, 2nd - input limbs, 3rd - column elements
		interm = make([][][]field.Element, numCol)
	)

	for i := range ctx.InputCols {
		inputs[i] = make([]smartvectors.SmartVector, len(ctx.InputCols[i]))
		for j := range ctx.InputCols[i] {
			inputs[i][j] = ctx.InputCols[i][j].GetColAssignment(run)
		}
	}

	rangeStart, rangeStop := smartvectors.CoCompactRange(inputs[0]...)

	for i := range interm {
		interm[i] = make([][]field.Element, rangeStop-rangeStart)
		for j := range interm[i] {
			interm[i][j] = make([]field.Element, common.NbLimbU256)
		}
	}

	parallel.Execute(rangeStop-rangeStart, func(t0, t1 int) {

		var (
			prevState = make([][]field.Element, t1-t0)
			start     = rangeStart + t0
			stop      = rangeStart + t1
		)

		for i := range prevState {
			prevState[i] = make([]field.Element, common.NbLimbU256)
		}

		for i := range interm {
			for k := start; k < stop; k++ {
				var inputLimbs []field.Element
				for _, limb := range inputs[i] {
					inputLimbs = append(inputLimbs, limb.Get(k))
				}

				limbs := common.BlockCompression(prevState[k-start][:], inputLimbs)
				for j, limb := range limbs {
					interm[i][k][j] = limb
				}
			}

			prevState = interm[i][start:stop]
		}

	})

	elements := transform3DArray(interm)
	for i := range interm {
		for j := range interm[i][0] {
			run.AssignColumn(
				ctx.IntermediateHashes[i][j].GetColID(),
				smartvectors.FromCompactWithRange(elements[i][j], rangeStart, rangeStop, numRow),
			)
		}
	}
}

// transform3DArray transposes a 3D array of field.Element type.
// Specifically, it transforms an input array with dimensions [dim1][dim2][dim3] into
// an output array with dimensions [dim1][dim3][dim2].
//
// Note: this function is used for 'moving' the limb dimension higher.
func transform3DArray(input [][][]field.Element) [][][]field.Element {
	if len(input) == 0 || len(input[0]) == 0 || len(input[0][0]) == 0 {
		return nil
	}

	dim1 := len(input)
	dim2 := len(input[0])
	dim3 := len(input[0][0])

	output := make([][][]field.Element, dim1)
	for i := range output {
		output[i] = make([][]field.Element, dim3)
		for k := range output[i] {
			output[i][k] = make([]field.Element, dim2)
		}
	}

	for i := 0; i < dim1; i++ {
		for j := 0; j < dim2; j++ {
			for k := 0; k < dim3; k++ {
				output[i][k][j] = input[i][j][k]
			}
		}
	}

	return output
}
