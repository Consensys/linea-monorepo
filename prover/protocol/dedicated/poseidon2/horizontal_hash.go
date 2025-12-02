package poseidon2

import (
	"strconv"

	"github.com/consensys/gnark-crypto/field/koalabear/vortex"
	"github.com/consensys/linea-monorepo/prover/crypto/poseidon2"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/column/verifiercol"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/parallel"
)

const BlockSize = poseidon2.BlockSize

// HashingCtx is the wizard context responsible for hashing a table row-wise.
// It implements the [wizard.ProverAction] interface and is used to assign the
// InputColumns field.
type HashingCtx struct {
	// InputCols is the list of columns forming a table for which the current
	// context is computing the rows.
	InputCols [][BlockSize]ifaces.Column
	// IntermediateHashes stores the intermediate values of the hasher.
	IntermediateHashes [][BlockSize]ifaces.Column
}

// MaxOctRound round of declaration for a list of commitment
func MaxOctRound(handles ...[BlockSize]ifaces.Column) int {
	res := 0
	for _, handle := range handles {
		res = utils.Max(res, handle[0].Round())
	}
	return res
}

// AssertSameLength is a utility function comparing the Size of all the columns
// in `list` , panicking if the lengths are not all the same and returning the
// shared length otherwise.
func AssertOctSameLength(list ...[BlockSize]ifaces.Column) int {
	if len(list) == 0 {
		panic("passed an empty leaf")
	}

	res := list[0][0].Size()
	for i := range list {
		if list[i][0].Size() != res {
			utils.Panic("the column %v (size %v) does not have the same size as column 0 (size %v)", i, list[i][0].Size(), res)
		}
	}

	return res
}

// HashOf returns an [ifaces.Column] object containing the hash of the inputs
// columns and a [wizard.ProverAction] object responsible for assigning all
// the column taking part in justifying the returned column as well as the
// returned column itself.
func HashOf(comp *wizard.CompiledIOP, inputCols [][BlockSize]ifaces.Column) *HashingCtx {

	var (
		ctx = &HashingCtx{
			InputCols:          inputCols,
			IntermediateHashes: make([][BlockSize]ifaces.Column, len(inputCols)),
		}

		round     = MaxOctRound(inputCols...)
		ctxID     = len(comp.ListCommitments())
		numRows   = AssertOctSameLength(inputCols...)
		prevState [BlockSize]ifaces.Column
	)

	for i := 0; i < BlockSize; i++ {
		prevState[i] = verifiercol.NewConstantCol(field.Zero(), numRows, strconv.Itoa(ctxID))
	}

	for i := range ctx.IntermediateHashes {
		for j := 0; j < BlockSize; j++ {
			ctx.IntermediateHashes[i][j] = comp.InsertCommit(
				round,
				ifaces.ColIDf("HASHING_%v_%v_%v", ctxID, i, j),
				numRows,
				true,
			)
		}

		comp.InsertPoseidon2(
			round,
			ifaces.QueryIDf("HASHING_%v_%v", ctxID, i),
			inputCols[i], prevState, ctx.IntermediateHashes[i],
			nil,
		)

		prevState = ctx.IntermediateHashes[i]
	}

	return ctx
}

// Result returns the column containing the result of the hashing.
func (ctx *HashingCtx) Result() [BlockSize]ifaces.Column {
	return ctx.IntermediateHashes[len(ctx.IntermediateHashes)-1]
}

// Run implements the [wizard.ProverAction] interface
func (ctx *HashingCtx) Run(run *wizard.ProverRuntime) {

	var (
		numRow = ctx.InputCols[0][0].Size()
		numCol = len(ctx.InputCols)
		inputs [BlockSize][]smartvectors.SmartVector
		interm = make([][]field.Octuplet, numCol)
	)

	for i := 0; i < BlockSize; i++ {
		inputs[i] = make([]smartvectors.SmartVector, numCol)
	}

	for i := range interm {
		for j := 0; j < BlockSize; j++ {
			inputs[j][i] = ctx.InputCols[i][j].GetColAssignment(run)
		}
	}

	rangeStart, rangeStop := smartvectors.CoCompactRange(inputs[0]...)

	for i := range interm {
		interm[i] = make([]field.Octuplet, rangeStop-rangeStart)
	}

	parallel.Execute(rangeStop-rangeStart, func(t0, t1 int) {

		var (
			prevState = make([]field.Octuplet, t1-t0)
			start     = rangeStart + t0
			stop      = rangeStart + t1
		)

		for i := range interm {
			for k := start; k < stop; k++ {
				var block [BlockSize]field.Element
				for j := 0; j < BlockSize; j++ {
					block[j] = inputs[j][i].Get(k)
				}
				interm[i][k] = vortex.CompressPoseidon2(prevState[k-start], block)
			}
			prevState = interm[i][start:stop]
		}
	})

	for i := range interm {
		for j := 0; j < BlockSize; j++ {
			intermSlice := make([]field.Element, len(interm[i]))
			for k := range interm[i] {
				intermSlice[k] = interm[i][k][j]
			}
			run.AssignColumn(
				ctx.IntermediateHashes[i][j].GetColID(),
				smartvectors.FromCompactWithRange(intermSlice, rangeStart, rangeStop, numRow),
			)
		}
	}
}

// SplitBy splits the input slice into subarrays of size poseidon2.BlockSize. If the input is not divisible by the size, it appends constant verifier columns to the input to make it divisible by the size.
func SplitBy(input []ifaces.Column) [][poseidon2.BlockSize]ifaces.Column {

	var (
		n = len(input) % poseidon2.BlockSize
	)

	if n != 0 {

		verifCol := verifiercol.NewConstantCol(field.Zero(), input[0].Size(), "CONSTANT_COLUMN")

		// append constant verifier columns to the input to make it divisible by size
		for i := 0; i < poseidon2.BlockSize-n; i++ {
			input = append(input, verifCol)
		}
	}
	m := len(input) / poseidon2.BlockSize

	result := make([][poseidon2.BlockSize]ifaces.Column, m)
	for i := 0; i < n; i++ {
		for j := 0; j < poseidon2.BlockSize; j++ {
			result[i][j] = input[i*poseidon2.BlockSize+j]
		}
	}
	return result
}
