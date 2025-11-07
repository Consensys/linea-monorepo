package iokeccakf

import (
	"github.com/consensys/linea-monorepo/prover/crypto/keccak"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/common/vector"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	sym "github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/common"
	commonconstraints "github.com/consensys/linea-monorepo/prover/zkevm/prover/common/common_constraints"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/generic"
	baseconversion "github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/keccak/keccakf_koalabear/base_conversion"
)

const (
	NbOfRowsPerLane = 4
	NumSlices       = 8
)

type IOKeccakFInputs struct {
	Lane                 ifaces.Column // from packing
	IsBeginningOfNewHash ifaces.Column
	IsLaneActive         ifaces.Column
	KeccakfSize          int // the size of keccakf module.
}

type IOKeccakF struct {
	Inputs           IOKeccakFInputs
	isFromFirstBlock ifaces.Column // built from isBeginningOfNewHash and rowsPerBlock of keccakf
	isFromBlockBaseB ifaces.Column
	blocks           [][NumSlices]ifaces.Column // the blocks of the message to be absorbed. first blocks of messages are located in positions 0 mod 24 and are represented in base clean 12, other blocks of message are located in positions 23 mod 24 and are represented in base clean 11. otherwise the blocks are zero.
	isBlock          ifaces.Column              // indicates whether the row corresponds to a block
	// prover action for base conversion
	bc            *baseconversion.ToBaseX
	isFirstBlock  ifaces.Column
	isBlockBaseB  ifaces.Column
	isBlockActive ifaces.Column // active part of the blocks (technicaly it is the active part of the keccakf module).
}

// it first applies to-basex to get laneX, then a projection query to map lanex to blocks
func NewIOKeccakF(comp *wizard.CompiledIOP, inputs IOKeccakFInputs) *IOKeccakF {

	var (
		laneSize          = inputs.Lane.Size()
		params            = generic.KeccakUsecase
		nbOfLanesPerBlock = params.NbOfLanesPerBlock()
		nbOfRowsPerBlock  = nbOfLanesPerBlock * NbOfRowsPerLane
		allBlocks         = []ifaces.Column{}

		io = &IOKeccakF{
			Inputs: inputs,
		}
	)

	io.isFromFirstBlock = comp.InsertCommit(0, ifaces.ColIDf("IsFromFirstBlock"), laneSize)
	io.isFromBlockBaseB = comp.InsertCommit(0, ifaces.ColIDf("IsFromBlockBaseB"), laneSize)
	io.isBlock = comp.InsertCommit(0, ifaces.ColIDf("IsBlock"), inputs.KeccakfSize)
	io.isFirstBlock = comp.InsertCommit(0, ifaces.ColIDf("IsFirstBlock"), inputs.KeccakfSize)
	io.isBlockBaseB = comp.InsertCommit(0, ifaces.ColIDf("IsBlockBaseB"), inputs.KeccakfSize)
	io.isBlockActive = comp.InsertCommit(0, ifaces.ColIDf("IsActive"), inputs.KeccakfSize)
	colRound := comp.InsertPrecomputed(ifaces.ColIDf("KeccakFRound"),
		smartvectors.NewRegular(vector.PeriodicOne(keccak.NumRound, inputs.KeccakfSize)))

	io.blocks = make([][NumSlices]ifaces.Column, nbOfLanesPerBlock)
	for i := range io.blocks {
		for j := 0; j < NumSlices; j++ {
			io.blocks[i][j] = comp.InsertCommit(0,
				ifaces.ColIDf("KeccakFBlock_%d_%d", i, j),
				inputs.KeccakfSize)
		}
	}

	// check isFirstBlock and isBlockBaseBaseB are correctly built
	// isFromFirstBlock is well formed
	// isFromFirstBlock = sum_j Shift(l.isFirstLaneFromNewHash,-j) for j:=0,...,
	s := sym.NewConstant(0)
	for j := 0; j < nbOfRowsPerBlock; j++ {
		s = sym.Add(
			s, column.Shift(io.Inputs.IsBeginningOfNewHash, -j),
		)
	}
	comp.InsertGlobal(0, ifaces.QueryIDf("IsFromFirstBlock"),
		sym.Sub(s, io.isFromFirstBlock))

	commonconstraints.MustBeMutuallyExclusiveBinaryFlags(comp,
		io.Inputs.IsLaneActive,
		[]ifaces.Column{io.isFromFirstBlock, io.isFromBlockBaseB},
	)

	commonconstraints.MustBeMutuallyExclusiveBinaryFlags(comp,
		io.isBlock,
		[]ifaces.Column{io.isFirstBlock, io.isBlockBaseB},
	)

	commonconstraints.MustZeroWhenInactive(comp, io.isBlock, io.isBlockActive)

	comp.InsertGlobal(0, ifaces.QueryIDf("BLOCKS_POSITIONS_CHECK"),
		sym.Mul(io.isBlockActive,
			sym.Sub(
				sym.Add(io.isFirstBlock, column.Shift(io.isBlockBaseB, -1)),
				colRound,
			),
		),
	)

	commonconstraints.MustBeActivationColumns(comp, io.isBlockActive)

	for i := 0; i < nbOfLanesPerBlock; i++ {
		allBlocks = append(allBlocks, io.blocks[i][:]...)
	}

	// apply to-base-x to get lanes in bases required for keccakf
	io.bc = baseconversion.NewToBaseX(comp, baseconversion.ToBaseXInputs{
		Lane:           inputs.Lane,
		IsLaneActive:   inputs.IsLaneActive,
		BaseX:          []int{4, 11},
		NbBitsPerBaseX: 8,
		IsBaseX:        []ifaces.Column{io.isFromFirstBlock, io.isFromBlockBaseB},
	})

	// to flatten lane columns over blocks columns, we use projection query
	columnsA := make([][]ifaces.Column, len(io.bc.LaneX))
	filterA := make([]ifaces.Column, len(io.bc.LaneX))
	for i, col := range io.bc.LaneX {
		columnsA[i] = []ifaces.Column{col}
		filterA[i] = io.Inputs.IsLaneActive
	}
	columnsB := make([][]ifaces.Column, len(allBlocks))
	filterB := make([]ifaces.Column, len(allBlocks))
	for i := range allBlocks {
		columnsB[i] = []ifaces.Column{allBlocks[i]}
		filterB[i] = io.isBlock
	}

	// projection query to get blocks from lanes
	comp.InsertProjection(ifaces.QueryIDf("KeccakFBlocksFromLanes"), query.ProjectionMultiAryInput{
		ColumnsA: columnsA,
		ColumnsB: columnsB,
		FiltersA: filterA,
		FiltersB: filterB,
	})

	return io

}

func (io *IOKeccakF) Run(run *wizard.ProverRuntime) {
	var (
		laneSize             = io.Inputs.Lane.Size()
		isBeginningOfNewHash = io.Inputs.IsBeginningOfNewHash.GetColAssignment(run).IntoRegVecSaveAlloc()
		param                = generic.KeccakUsecase
		nbOfRowsPerLane      = param.LaneSizeBytes() / MAXNBYTE
		numRowsPerBlock      = param.NbOfLanesPerBlock() * nbOfRowsPerLane
		isActive             = io.Inputs.IsLaneActive.GetColAssignment(run).IntoRegVecSaveAlloc()
		colIsFromFirstBlock  = common.NewVectorBuilder(io.isFromFirstBlock)
		colIsFromOtherBlocks = common.NewVectorBuilder(io.isFromBlockBaseB)
		ones                 = vector.Repeat(field.One(), numRowsPerBlock)
	)

	// assign the columns isFromFirstBlock, isFromBlockBaseB
	for j := 0; j < laneSize; j++ {
		if isBeginningOfNewHash[j].IsOne() {
			colIsFromFirstBlock.PushSliceF(ones)
			j = j + (numRowsPerBlock - 1)
		} else {
			colIsFromFirstBlock.PushInt(0)
		}
	}

	isNotFirstBlock := make([]field.Element, laneSize)
	vector.Sub(isNotFirstBlock, isActive, colIsFromFirstBlock.Slice())
	colIsFromOtherBlocks.PushSliceF(isNotFirstBlock)

	colIsFromFirstBlock.PadAndAssign(run, field.Zero())
	colIsFromOtherBlocks.PadAndAssign(run, field.Zero())

	// run base conversion to get laneX from lane
	io.bc.Run(run)
	// assign the blocks and flags
	io.assignBlocks(run)

}

// It assigns the columns specific to the submodule.
func (io *IOKeccakF) assignBlocks(
	run *wizard.ProverRuntime) {
	var (
		isFirstBlock         = common.NewVectorBuilder(io.isFirstBlock)
		isBlockBaseB         = common.NewVectorBuilder(io.isBlockBaseB)
		isBlock              = common.NewVectorBuilder(io.isBlock)
		isBeginningOfNewHash = io.Inputs.IsBeginningOfNewHash.GetColAssignment(run).IntoRegVecSaveAlloc()
		laneX                = make([][]field.Element, len(io.bc.LaneX))
		blocks               = make([][NumSlices]*common.VectorBuilder, len(io.blocks))
		numRowsPerBlock      = generic.KeccakUsecase.NbOfLanesPerBlock() * NbOfRowsPerLane
		isLaneActive         = io.Inputs.IsLaneActive.GetColAssignment(run).IntoRegVecSaveAlloc()
		isBlockActive        = *common.NewVectorBuilder(io.isBlockActive)
	)

	for i := range io.blocks {
		for j := 0; j < NumSlices; j++ {
			blocks[i][j] = common.NewVectorBuilder(io.blocks[i][j])
		}
	}

	blockBuilder := &blockBuilder{blocks: blocks}
	for j := range io.bc.LaneX {
		laneX[j] = run.GetColumn(io.bc.LaneX[j].GetColID()).IntoRegVecSaveAlloc()
	}

	// assign isFirstBlock, isBlockBaseB, isBlock
	zeroes := make([]field.Element, keccak.NumRound-1)
	laneActivePart := 0
	ctr := 0
	for i, isNewHash := range isBeginningOfNewHash {

		if isLaneActive[i].IsZero() {
			break
		}
		if i%numRowsPerBlock != 0 {
			continue
		}
		laneActivePart++

		if isNewHash.IsOne() {
			isFirstBlock.PushOne()
			isFirstBlock.PushSliceF(zeroes)
			// append 24 zeroes
			isBlockBaseB.PushSliceF(zeroes)
			isBlockBaseB.PushInt(0)
			// populate IsBlock
			isBlock.PushOne()
			isBlock.PushSliceF(zeroes)
			blockBuilder.pushLaneToBlock(laneX, ctr, ctr+numRowsPerBlock)
			blockBuilder.pushZeroSliceToBlock(keccak.NumRound - 1)
			ctr += numRowsPerBlock

		} else {
			isFirstBlock.PushZero()
			isFirstBlock.PushSliceF(zeroes)

			isBlockBaseB.OverWriteInt(1)
			// append 24 zeroes
			isBlockBaseB.PushSliceF(zeroes)
			isBlockBaseB.PushZero()

			//populate IsBlock
			isBlock.OverWriteInt(1)
			isBlock.PushSliceF(zeroes)
			isBlock.PushZero()
			blockBuilder.overWriteBlock(laneX, ctr, ctr+numRowsPerBlock)
			blockBuilder.pushZeroSliceToBlock(keccak.NumRound)
			ctr += numRowsPerBlock
		}
	}

	blockActivePart := ((laneActivePart + 1) / numRowsPerBlock) * keccak.NumRound
	ones := vector.Repeat(field.One(), blockActivePart)
	isBlockActive.PushSliceF(ones)
	isBlockActive.PadAndAssign(run)
	isBlock.PadAndAssign(run)
	isFirstBlock.PadAndAssign(run)
	isBlockBaseB.PadAndAssign(run)

	for i := range io.blocks {
		for j := 0; j < NumSlices; j++ {
			blockBuilder.blocks[i][j].PadAndAssign(run, field.Zero())
		}
	}

}

type blockBuilder struct {
	blocks [][NumSlices]*common.VectorBuilder
}

func (b *blockBuilder) pushLaneToBlock(
	laneX [][]field.Element,
	rowStart int,
	rowStop int,
) {
	if (rowStop-rowStart)*len(laneX) != len(b.blocks)*NumSlices {
		panic("invalid size for pushLaneToBlock")
	}

	iIndex := 0
	jIndex := 0
	for row := rowStart; row < rowStop; row++ {
		for i := range laneX {
			b.blocks[iIndex][jIndex].PushField(laneX[i][row])
			jIndex++
			if jIndex == NumSlices {
				jIndex = 0
				iIndex++
			}
		}
	}
}

func (b *blockBuilder) pushZeroSliceToBlock(n int) {
	zeros := vector.Zero(n)
	for i := range b.blocks {
		for j := 0; j < NumSlices; j++ {
			b.blocks[i][j].PushSliceF(zeros)
		}
	}
}

func (b *blockBuilder) overWriteBlock(laneX [][]field.Element,
	rowStart int,
	rowStop int) {
	iIndex := 0
	jIndex := 0
	for row := rowStart; row < rowStop; row++ {
		for i := range laneX {
			b.blocks[iIndex][jIndex].OverWriteInt(int(laneX[i][row].Uint64()))
			jIndex++
			if jIndex == NumSlices {
				jIndex = 0
				iIndex++
			}
		}
	}

}
func (io *IOKeccakF) Blocks() [][NumSlices]ifaces.Column {
	return io.blocks
}
