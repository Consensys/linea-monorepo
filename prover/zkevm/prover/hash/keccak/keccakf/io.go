package keccakf

import (
	"github.com/consensys/zkevm-monorepo/prover/crypto/keccak"
	"github.com/consensys/zkevm-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/zkevm-monorepo/prover/maths/common/vector"
	"github.com/consensys/zkevm-monorepo/prover/maths/field"
	"github.com/consensys/zkevm-monorepo/prover/protocol/column"
	"github.com/consensys/zkevm-monorepo/prover/protocol/ifaces"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
	"github.com/consensys/zkevm-monorepo/prover/symbolic"
	"github.com/consensys/zkevm-monorepo/prover/utils"
)

const (
	// number of slices in keccakf for base conversion
	numSlices = 16

	// number of lanes in the output of hash
	// digest is 256 bits that is 4 lanes of 64bits
	numLanesInHashOutPut = 4
)

type InputOutput struct {

	// it is 1 iff BlockBaseB is inserted
	IsBlockBaseB ifaces.Column

	// it is 1 iff firstBlock is inserted
	IsFirstBlock ifaces.Column

	//Sum of IsBlockBaseB and IsFirstBlock
	IsBlcok ifaces.Column

	// shifted version of (the effective part of) isFirstBlock.
	IsHashOutPut ifaces.Column

	PiChiIota piChiIota

	HashOutputSlicesBaseB [numLanesInHashOutPut][numSlices]ifaces.Column
}

/*
	The method newInput declares the columns specific to the submodule.

It also declares the constraints over the input/output of keccakF.

1. any permutation should be relevant to firstBlock or BlockBaseB (but not to both).

2. for an ongoing permutation the output of aIota is fed back to the state

3. for the permutations from the same hash the output of aIota is fed back to the state

Note: The initial state (of our implementation) is supposed to be the first Block.
No constraint in keccakF module is asserting to this fact.
Instead, it is guaranteed via a projection query between output of data-transfer module and keccakF

This is tanks to the fact that;     initial state (of original keccakF) = 0
And so,  FirstBlock XOR (original) initial state = our initial state.
Therefore we can directly project from the output of dataTransfer module to the keccakF state.
*/
func (io *InputOutput) newInput(comp *wizard.CompiledIOP, round, maxNumKeccakF int,
	mod Module,
) {
	lu := mod.lookups
	io.PiChiIota = mod.piChiIota

	// declare the columns
	io.declareColumnsInput(comp, round, maxNumKeccakF)

	// declare the constraints

	// For a new permutation or it is FirstBlock or it is fed by BlockBaseB.
	// The constraint also check that the columns are binary
	io.csIsFirstBlockIsBlockBaseB(comp, round, mod.isActive, lu)

	// IsBlockMassage = isFirstBlock + isBlockBaseB
	comp.InsertGlobal(round, ifaces.QueryIDf("IsBlockMassage"),
		symbolic.Sub(io.IsBlcok, symbolic.Add(io.IsFirstBlock, io.IsBlockBaseB)))

}

func (io *InputOutput) newOutput(comp *wizard.CompiledIOP, round, maxNumKeccakF int,
	mod Module,
) {
	lu := mod.lookups
	input := mod.state
	io.PiChiIota = mod.piChiIota

	// declare the columns
	io.declareColumnsOutput(comp, round, maxNumKeccakF)

	// declare the constraints

	// Constraints over isActive; it starts with ones, ends with zeroes
	io.csIsActive(comp, round, mod.isActive)

	// The constrain over the next sate;
	//  - for an ongoing permutation, the next state equals with the last aIota
	//
	//  - for the next permutation from the same hash, the state equals with the last aIota
	io.csNextState(comp, round, io.PiChiIota, lu, input)

	// constraints over hash output
	io.csHashOutput(comp, round)
}

// It declares the columns specific to the submodule.
func (io *InputOutput) declareColumnsInput(comp *wizard.CompiledIOP, round, maxNumKeccakF int) {
	size := numRows(maxNumKeccakF)
	io.IsHashOutPut = comp.InsertCommit(round, deriveName("IS_FIRST_BLOCK_SHIFTED"), size)

	io.IsFirstBlock = comp.InsertCommit(round, deriveName("IS_FIRST_BLOCK"), size)
	io.IsBlockBaseB = comp.InsertCommit(round, deriveName("IS_BLOCK_BaseB"), size)
	io.IsBlcok = comp.InsertCommit(round, ifaces.ColIDf("IS_BLOCK_MESSAGE"), size)
}

// It declares the columns specific to the submodule.
func (io *InputOutput) declareColumnsOutput(comp *wizard.CompiledIOP, round, maxNumKeccakF int) {
	hashOutputSize := utils.NextPowerOfTwo(maxNumKeccakF)
	for j := range io.HashOutputSlicesBaseB {
		for k := range io.HashOutputSlicesBaseB[0] {
			io.HashOutputSlicesBaseB[j][k] = comp.InsertCommit(round,
				ifaces.ColIDf("HashOutPut_SlicesBaseB_%v_%v", j, k), hashOutputSize)
		}
	}
}

// Constraints over the blocks of the message;
// any permutation should be relevant to the firstBlock or to the blockBaseB.
// The flag columns related to firstBlock or to blockBaseB (i.e., isFirstBlock and isBlockBaseB)
// are binary. Also, they are inactive on the inactive part of the module.
func (io InputOutput) csIsFirstBlockIsBlockBaseB(comp *wizard.CompiledIOP, round int, isActive ifaces.Column, lu lookUpTables) {
	//  for a new permutation or it is firstBlock or it is BlockBaseB
	//  newPerm[i] = 1 ---> isFirstBlock[i] =1 or isBlockBaseB[i-1] = 1 (but not both)
	//  thus, newPerm[i] * (isFirstBlock[i] -1) * (isBlockBaseB[i-1] - 1) =0
	//  and isFirstBlock * isBlockBaseB =0
	isNotFirstBlock := symbolic.Sub(1, io.IsFirstBlock)                          // 1- isFirtBlock[i]
	isNotBlockBaseBShifted := symbolic.Sub(1, column.Shift(io.IsBlockBaseB, -1)) // 1- isBlockBaseB[i-1]
	isNewPerm := symbolic.Sub(1, lu.UsePrevAIota)                                // isNewPerm
	expr := symbolic.Mul(symbolic.Mul(isNewPerm, isNotFirstBlock), isNotBlockBaseBShifted)

	comp.InsertGlobal(round, ifaces.QueryIDf("ISFIRST_OR_ISBLOCKBASEB_1"), symbolic.Mul(expr, isActive))
	comp.InsertGlobal(round, ifaces.QueryIDf("ISFIRST_OR_ISBLOCKBASEB_2"),
		symbolic.Mul(io.IsFirstBlock, io.IsBlockBaseB))

	// The columns are binary
	comp.InsertGlobal(round, ifaces.QueryIDf("IsFisrtBlock_IsBinary"),
		symbolic.Mul(io.IsFirstBlock, symbolic.Sub(1, io.IsFirstBlock)))

	comp.InsertGlobal(round, ifaces.QueryIDf("IsBlockBaseB_IsBinary"),
		symbolic.Mul(io.IsBlockBaseB, symbolic.Sub(1, io.IsBlockBaseB)))

	// The columns are zero when isActive is zero
	// This constraint is important for data-transferring.
	// The prover may respect the number of  blocks to be imported to keccakF,
	// but it may not insert them in the active zone.
	comp.InsertGlobal(round, ifaces.QueryIDf("IsFirstBlock_IsNotActive"),
		symbolic.Mul(io.IsFirstBlock, symbolic.Sub(1, isActive)))

	comp.InsertGlobal(round, ifaces.QueryIDf("IsBlockBaseB_IsNotActive"),
		symbolic.Mul(io.IsBlockBaseB, symbolic.Sub(1, isActive)))

	// constraints over isFirstBlockShifted
	comp.InsertGlobal(round, ifaces.QueryIDf("IsFirstBlockShifted"),
		symbolic.Mul(symbolic.Sub(column.Shift(io.IsFirstBlock, 1),
			io.IsHashOutPut), column.Shift(isActive, 1)))

	// the last active cell of isFirstBlockShifted is 1
	// if isActive[i]=  1 and isActive[i+1] = 0 ---> isFirstBlockShifted[i] =1
	expr = symbolic.Mul(symbolic.Mul(isActive,
		symbolic.Sub(1, column.Shift(isActive, 1)),
		symbolic.Sub(1, io.IsHashOutPut)))
	comp.InsertGlobal(round, ifaces.QueryIDf("Last_Active_Cell_IsOne"), expr)

}

// Constraints over isActive, it starts with ones, and ends with zeroes
//
//  1. a := (isActive[i]-isActive[i+1]) is binary
//  2. isActive is binary
//
// Note: we don't have any constraint over the number of ones.
// This would be later guaranteed by the projection query between datatransfer-outputs and keccakf-inputs.
func (io *InputOutput) csIsActive(comp *wizard.CompiledIOP, round int, isActive ifaces.Column) {
	a := symbolic.Sub(isActive, column.Shift(isActive, 1))

	comp.InsertGlobal(round, ifaces.QueryIDf("OnesThenZeroes_Keccakf"), symbolic.Mul(a, symbolic.Sub(1, a)))
	comp.InsertGlobal(round, ifaces.QueryIDf("IsActive_IsBinary_Keccakf"), symbolic.Mul(isActive, symbolic.Sub(1, isActive)))
}

// The constrain over the next sate;
//
//   - for an ongoing permutation, the next state equals with the last aIota
//
//   - for the next permutation from the same hash, the state equals with the last aIota
func (io *InputOutput) csNextState(
	comp *wizard.CompiledIOP,
	round int,
	pci piChiIota,
	lu lookUpTables,
	input [5][5]ifaces.Column,
) {

	for x := 0; x < 5; x++ {
		for y := 0; y < 5; y++ {
			// recompose the slices to get aIota in BaseA
			recomposedAIota := BaseRecomposeSliceHandles(
				// Recall that the chi module performs all the steps pi-chi-iota
				// at once.
				pci.aIotaBaseASliced[x][y][:],
				BaseA,
				true,
			)
			// for an ongoing permutation or for permutations from the same hash;
			// impose that the previous aIota in base A should be equal with the state.
			usePrevIota := symbolic.Add(column.Shift(io.IsBlockBaseB, -1), lu.UsePrevAIota) // isBlockBaseB + UsePrevAIota
			expr := symbolic.Mul(symbolic.Sub(input[x][y], recomposedAIota), usePrevIota)
			name := ifaces.QueryIDf("AIOTA_TO_A_%v_%v", x, y)
			comp.InsertGlobal(round, name, expr)
		}
	}

}

// It connect the data-transfer module to the keccakf module via a projection query over the blocks.
func (io *InputOutput) csHashOutput(comp *wizard.CompiledIOP, round int) {

	// constraints over info-module (outputs
	colA := append(io.PiChiIota.AIotaBaseBSliced[0][0][:], io.PiChiIota.AIotaBaseBSliced[1][0][:]...)
	colA = append(colA, io.PiChiIota.AIotaBaseBSliced[2][0][:]...)
	colA = append(colA, io.PiChiIota.AIotaBaseBSliced[3][0][:]...)

	colB := append(io.HashOutputSlicesBaseB[0][:], io.HashOutputSlicesBaseB[1][:]...)
	colB = append(colB, io.HashOutputSlicesBaseB[2][:]...)
	colB = append(colB, io.HashOutputSlicesBaseB[3][:]...)

	comp.InsertInclusionConditionalOnIncluded(round, ifaces.QueryIDf("HashOutput_Projection"),
		colB, colA, io.IsHashOutPut)
}

// It assigns the columns specific to the submodule.
func (io *InputOutput) assignInputOutput(
	run *wizard.ProverRuntime,
	permTrace keccak.PermTraces,
) {
	witnessSize := len(permTrace.KeccakFInps) * numRounds

	var isFirstBlock, isBlockBaseB []field.Element
	zeroes := make([]field.Element, keccak.NumRound-1)
	for i := range permTrace.IsNewHash {
		if permTrace.IsNewHash[i] {
			isFirstBlock = append(isFirstBlock, field.One())
			isFirstBlock = append(isFirstBlock, zeroes...)
			// append 24 zeroes
			isBlockBaseB = append(isBlockBaseB, zeroes...)
			isBlockBaseB = append(isBlockBaseB, field.Zero())

		} else {
			isFirstBlock = append(isFirstBlock, field.Zero())
			isFirstBlock = append(isFirstBlock, zeroes...)
			// overwrite the last element
			isBlockBaseB[len(isBlockBaseB)-1] = field.One()
			// append 24 zeroes
			isBlockBaseB = append(isBlockBaseB, zeroes...)
			isBlockBaseB = append(isBlockBaseB, field.Zero())
		}
	}
	size := io.IsFirstBlock.Size()

	run.AssignColumn(io.IsFirstBlock.GetColID(), smartvectors.RightZeroPadded(isFirstBlock, size))
	run.AssignColumn(io.IsBlockBaseB.GetColID(), smartvectors.RightZeroPadded(isBlockBaseB, size))

	var shifted []field.Element
	if len(isFirstBlock) > 0 {
		shifted = append(shifted, isFirstBlock[1:]...)
		shifted = append(shifted, field.One())
	}
	run.AssignColumn(io.IsHashOutPut.GetColID(), smartvectors.RightZeroPadded(shifted, size))

	//populate and assign isBlock
	isBlock := make([]field.Element, witnessSize)
	vector.Add(isBlock, isFirstBlock, isBlockBaseB)
	run.AssignColumn(io.IsBlcok.GetColID(), smartvectors.RightZeroPadded(isBlock, size))
}

// It assigns the columns
func (io *InputOutput) assignHashOutPut(run *wizard.ProverRuntime) {
	aIota := io.PiChiIota.AIotaBaseBSliced
	isHashOutput := io.IsHashOutPut.GetColAssignment(run).IntoRegVecSaveAlloc()

	witSize := smartvectors.Density(aIota[0][0][0].GetColAssignment(run))
	sizeHashOutput := io.HashOutputSlicesBaseB[0][0].Size()

	for j := range io.HashOutputSlicesBaseB {
		for k := range io.HashOutputSlicesBaseB[0] {
			aIotaWit := aIota[j][0][k].GetColAssignment(run).IntoRegVecSaveAlloc()[:witSize]
			var hashOutput []field.Element
			for i := range aIotaWit {
				if isHashOutput[i] == field.One() {
					hashOutput = append(hashOutput, aIotaWit[i])
				}
			}
			run.AssignColumn(io.HashOutputSlicesBaseB[j][k].GetColID(), smartvectors.RightZeroPadded(hashOutput, sizeHashOutput))
		}
	}
}
