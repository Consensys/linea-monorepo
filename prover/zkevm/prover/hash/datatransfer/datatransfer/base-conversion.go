package datatransfer

import (
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/common/vector"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/keccak/keccakf"
)

// The submodule baseConversion implements the base conversion over the lanes, in order to export them to the keccakf.
// The lanes from the first block of hash are in baseA and others are in baseB.
type baseConversion struct {
	// It is 1 when the lane is from the first block of the hash
	IsFromFirstBlock ifaces.Column

	// IsFromBlockBaseB := 1-isFromFirstBlock
	IsFromBlockBaseB ifaces.Column

	// Decomposition of lanes into slices of 16bits
	laneSlices [4]ifaces.Column

	// Slices in baseX; the one from first blocks are in baseA and others are in baseB.
	laneSlicesX [4]ifaces.Column

	// lanes from first block in baseA, others in baseB
	LaneX ifaces.Column
}

func (b *baseConversion) newBaseConversionOfLanes(
	comp *wizard.CompiledIOP,
	round, maxNumRows int,
	lane lane,
	lu lookUpTables,
) {
	// declare the columns
	b.insertCommit(comp, round, maxNumRows)

	// declare the constraints
	// 0. isFromFirstBlock is well formed
	// 1. base conversion via lookups
	b.csIsFromFirstBlock(comp, round, lane)
	b.csBaseConversion(comp, round, lane, lu)
}
func (b *baseConversion) insertCommit(comp *wizard.CompiledIOP, round, maxNumRows int) {
	for j := range b.laneSlices {
		b.laneSlices[j] = comp.InsertCommit(round, ifaces.ColIDf("SlicesUint16_%v", j), maxNumRows)
		b.laneSlicesX[j] = comp.InsertCommit(round, ifaces.ColIDf("SlicesX_%v", j), maxNumRows)
	}
	b.LaneX = comp.InsertCommit(round, ifaces.ColIDf("LaneX"), maxNumRows)
	b.IsFromFirstBlock = comp.InsertCommit(round, ifaces.ColIDf("IsFromFirstBlock"), maxNumRows)
	b.IsFromBlockBaseB = comp.InsertCommit(round, ifaces.ColIDf("IsFromBlockBaseB"), maxNumRows)
}

// assign the columns specific to the submodule
func (b *baseConversion) assignBaseConversion(run *wizard.ProverRuntime, l lane, maxNumRows int) {
	b.assignIsFromFirstBlock(run, l, maxNumRows)
	b.assignSlicesLaneX(run, l, maxNumRows)
}

func (b *baseConversion) csIsFromFirstBlock(comp *wizard.CompiledIOP, round int, l lane) {
	// isFromFirstBlock = sum_j Shift(l.isFirstLaneFromNewHash,-j) for j:=0,...,16
	s := symbolic.NewConstant(0)
	for j := 0; j < numLanesInBlock; j++ {
		s = symbolic.Add(s, column.Shift(l.isFirstLaneOfNewHash, -j))
	}
	comp.InsertGlobal(round, ifaces.QueryIDf("IsFromFirstBlock"),
		symbolic.Sub(s, b.IsFromFirstBlock))

	// isFromBlockBaseB = (1- isFromFirstBlock) * isLaneActive
	comp.InsertGlobal(round, ifaces.QueryIDf("isNotFirstBlock"),
		symbolic.Mul(symbolic.Sub(b.IsFromBlockBaseB,
			symbolic.Sub(1, b.IsFromFirstBlock)), l.isLaneActive))
}

func (b *baseConversion) csBaseConversion(comp *wizard.CompiledIOP, round int, l lane, lu lookUpTables) {
	// if isFromFirstBlock = 1  ---> convert to keccak.BaseA
	// otherwise convert to keccak.BaseB

	// fist decompose to slice of size 16bits and then convert via lookUps
	res := keccakf.BaseRecomposeHandles(b.laneSlices[:], power16)
	comp.InsertGlobal(round, ifaces.QueryIDf("RecomposeLaneFromUint16"), symbolic.Sub(res, l.lane))

	// base conversion slice by slice and via lookups
	for j := range b.laneSlices {
		comp.InsertInclusionConditionalOnIncluded(round,
			ifaces.QueryIDf("BaseConversionIntoBaseA_%v", j),
			[]ifaces.Column{lu.colUint16, lu.colBaseA},
			[]ifaces.Column{b.laneSlices[j], b.laneSlicesX[j]},
			b.IsFromFirstBlock)

		comp.InsertInclusionConditionalOnIncluded(round,
			ifaces.QueryIDf("BaseConversionIntoBaseB_%v", j),
			[]ifaces.Column{lu.colUint16, lu.colBaseB},
			[]ifaces.Column{b.laneSlices[j], b.laneSlicesX[j]},
			b.IsFromBlockBaseB)
	}
	// recomposition of slicesX into blockX
	base1Power16 := keccakf.BaseAPow4 * keccakf.BaseAPow4 * keccakf.BaseAPow4 * keccakf.BaseAPow4 // no overflow
	base2Power16 := keccakf.BaseBPow4 * keccakf.BaseBPow4 * keccakf.BaseBPow4 * keccakf.BaseBPow4
	base := symbolic.Add(symbolic.Mul(b.IsFromFirstBlock, base1Power16),
		symbolic.Mul(b.IsFromBlockBaseB, base2Power16))

	laneX := baseRecomposeHandles(b.laneSlicesX[:], base)
	comp.InsertGlobal(round, ifaces.QueryIDf("RecomposeLaneFromBaseX"), symbolic.Sub(laneX, b.LaneX))
}

// assign column isFromFirstBlock
func (b *baseConversion) assignIsFromFirstBlock(run *wizard.ProverRuntime, l lane, maxNumRows int) {
	ones := vector.Repeat(field.One(), numLanesInBlock)
	var col []field.Element
	witSize := smartvectors.Density(l.isFirstLaneOfNewHash.GetColAssignment(run))
	isFirstLaneOfNewHash := l.isFirstLaneOfNewHash.GetColAssignment(run).IntoRegVecSaveAlloc()[:witSize]

	for j := 0; j < witSize; j++ {
		if isFirstLaneOfNewHash[j] == field.One() {
			col = append(col, ones...)
			j = j + (numLanesInBlock - 1)
		} else {
			col = append(col, field.Zero())
		}
	}
	oneCol := vector.Repeat(field.One(), witSize)
	isNotFirstBlock := make([]field.Element, witSize)
	vector.Sub(isNotFirstBlock, oneCol, col)

	run.AssignColumn(b.IsFromFirstBlock.GetColID(), smartvectors.RightZeroPadded(col, maxNumRows))
	run.AssignColumn(b.IsFromBlockBaseB.GetColID(), smartvectors.RightZeroPadded(isNotFirstBlock, maxNumRows))
}

// assign slices for the base conversion; Slice, SliceX, laneX
func (b *baseConversion) assignSlicesLaneX(
	run *wizard.ProverRuntime,
	l lane, maxNumRows int) {

	witSize := smartvectors.Density(b.IsFromFirstBlock.GetColAssignment(run))
	isFirstBlock := b.IsFromFirstBlock.GetColAssignment(run).IntoRegVecSaveAlloc()[:witSize]
	lane := l.lane.GetColAssignment(run).IntoRegVecSaveAlloc()[:witSize]

	// decomposition
	// populating slices
	var laneSlices [4][]field.Element
	var laneSlicesX [4][]field.Element
	for i := range lane {
		v := keccakf.DecomposeSmall(lane[i].Uint64(), power16, 4)
		for k := 0; k < 4; k++ {
			laneSlices[k] = append(laneSlices[k], v[k])
		}
	}

	// assign the slice-columns
	for j := range b.laneSlices {
		run.AssignColumn(b.laneSlices[j].GetColID(), smartvectors.RightZeroPadded(laneSlices[j], maxNumRows))
	}

	// base conversion (slice by slice)
	// populating sliceX
	for i := range lane {
		if isFirstBlock[i].Uint64() == 1 {
			// base conversion to baseA
			for k := 0; k < 4; k++ {
				laneSlicesX[k] = append(laneSlicesX[k], uInt16ToBaseX(uint16(laneSlices[k][i].Uint64()), &keccakf.BaseAFr))
			}

		} else {
			// base conversion to baseB
			for k := 0; k < 4; k++ {
				laneSlicesX[k] = append(laneSlicesX[k], uInt16ToBaseX(uint16(laneSlices[k][i].Uint64()), &keccakf.BaseBFr))
			}

		}
	}
	// assign the sliceX-columns
	for j := range b.laneSlicesX {
		run.AssignColumn(b.laneSlicesX[j].GetColID(), smartvectors.RightZeroPadded(laneSlicesX[j], maxNumRows))
	}

	// populate laneX
	var laneX []field.Element
	for j := range lane {
		if isFirstBlock[j] == field.One() {
			laneX = append(laneX, keccakf.U64ToBaseX(lane[j].Uint64(), &keccakf.BaseAFr))

		} else {
			laneX = append(laneX, keccakf.U64ToBaseX(lane[j].Uint64(), &keccakf.BaseBFr))
		}
	}

	// assign the laneX
	run.AssignColumn(b.LaneX.GetColID(), smartvectors.RightZeroPadded(laneX, maxNumRows))
}
