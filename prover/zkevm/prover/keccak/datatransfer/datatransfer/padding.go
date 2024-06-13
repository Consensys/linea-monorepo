package datatransfer

import (
	permTrace "github.com/consensys/zkevm-monorepo/prover/crypto/keccak"
	"github.com/consensys/zkevm-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/zkevm-monorepo/prover/maths/common/vector"
	"github.com/consensys/zkevm-monorepo/prover/maths/field"
	"github.com/consensys/zkevm-monorepo/prover/protocol/column"
	"github.com/consensys/zkevm-monorepo/prover/protocol/ifaces"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
	"github.com/consensys/zkevm-monorepo/prover/symbolic"
	"github.com/consensys/zkevm-monorepo/prover/utils"
	"github.com/consensys/zkevm-monorepo/prover/zkevm/prover/keccak/generic"
)

// InsertPadding asserts if the padding over the given (gbm) columns is done correctly.
/*
padding is expected to be done via inserting new limbs;
 -  limb 1 for the first padded limb, 0 bytes for the middle limbs, and 128 for the last limb
 -  limb 129 if we require only one byte of padding
*/
func (iPadd *importAndPadd) insertPadding(comp *wizard.CompiledIOP, round int) {

	/* 	if isPadded[i-1] = 0, isPadded[i] = 1, isPadded[i+1] =1 ----> limb = 1
		if isPadded[i-1] = 1, isPadded[i] = 1, isPadded[i+1] =0 ----> limb = 128
		if isPadded[i-1] = 0, isPadded[i] = 1, isPadded[i+1] =0 ----> limb = 129
		if isPadded[i-1] = 1, isPadded[i] = 1, isPadded[i+1] =1 ----> limb = 0
	 the constraints over NBytes also guarantees the correct number of  padded zeroes.*/
	one := symbolic.NewConstant(1)       // domain separator value, for padding
	fpv := symbolic.NewConstant(128)     // final padding value
	fpvPlus := symbolic.NewConstant(129) // padding value for a single byte padding

	isPadded := ifaces.ColumnAsVariable(iPadd.isPadded)
	isPaddedMinus := ifaces.ColumnAsVariable(column.Shift(iPadd.isPadded, -1))
	isPaddedPlus := ifaces.ColumnAsVariable(column.Shift(iPadd.isPadded, 1))

	expr1 := (one.Sub(isPaddedMinus)).Mul(isPadded).Mul(isPaddedPlus).
		Mul(ifaces.ColumnAsVariable(iPadd.limb).Sub(one))
	expr2 := (isPaddedMinus).Mul(isPadded).Mul(one.Sub(isPaddedPlus)).
		Mul(ifaces.ColumnAsVariable(iPadd.limb).Sub(fpv))
	expr3 := (one.Sub(isPaddedMinus)).Mul(isPadded).Mul(one.Sub(isPaddedPlus)).
		Mul(ifaces.ColumnAsVariable(iPadd.limb).Sub(fpvPlus))
	expr4 := (isPaddedMinus).Mul(isPadded).Mul(isPaddedPlus).
		Mul(ifaces.ColumnAsVariable(iPadd.limb))

	comp.InsertGlobal(round, ifaces.QueryIDf("Padding1"), expr1)
	comp.InsertGlobal(round, ifaces.QueryIDf("Padding2"), expr2)
	comp.InsertGlobal(round, ifaces.QueryIDf("Padding3"), expr3)
	comp.InsertGlobal(round, ifaces.QueryIDf("Padding4"), expr4)

}

//  AssignImportAndPadd assigns the part of the columns corresponding to the padding limbs.
/*
padding is done via inserting new limbs;
 -  limb 1 for the first limb, 0 bytes for the middle limbs, and 128 for the last limb
 -  limb 129 if we require only one byte of padding
*/
func (iPadd *importAndPadd) assignImportAndPadd(
	run *wizard.ProverRuntime,
	permTrace permTrace.PermTraces,
	gt generic.GenTrace, maxRows int) {
	isSelected := gt.TO_HASH
	one := field.One()
	var hashNum, limb, nByte, cleanLimb []field.Element
	for i := range isSelected {
		if isSelected[i].Cmp(&one) == 0 {
			hashNum = append(hashNum, gt.HashNum[i])
			limb = append(limb, gt.Limb[i])
			nByte = append(nByte, gt.NByte[i])
			cleanLimb = append(cleanLimb, gt.CleanLimb[i])
		}
	}

	// extend the columns to include the padded limbs
	eLimb, eNbyte, eHashNum, eCleanLimb, isNewHash, isInserted, isPadded :=
		extendWithPadding(limb, nByte, hashNum, cleanLimb)

	// sanity check
	if len(eHashNum) != len(isNewHash) {
		utils.Panic("HashNum and  isNewHash have different sizes  %v, %v ",
			len(eHashNum), len(isNewHash))
	}

	iPadd.assignIsBlockComplete(run, eNbyte, isNewHash, eCleanLimb, permTrace, maxRows)

	// assign the columns
	run.AssignColumn(iPadd.isNewHash.GetColID(), smartvectors.RightZeroPadded(isNewHash, maxRows))
	run.AssignColumn(iPadd.limb.GetColID(), smartvectors.RightZeroPadded(eLimb, maxRows))
	run.AssignColumn(iPadd.nByte.GetColID(), smartvectors.RightZeroPadded(eNbyte, maxRows))
	run.AssignColumn(iPadd.isActive.GetColID(), smartvectors.RightZeroPadded(vector.Repeat(one, len(eLimb)), maxRows))
	run.AssignColumn(iPadd.isInserted.GetColID(), smartvectors.RightZeroPadded(isInserted, maxRows))
	run.AssignColumn(iPadd.isPadded.GetColID(), smartvectors.RightZeroPadded(isPadded, maxRows))
	run.AssignColumn(iPadd.hashNum.GetColID(), smartvectors.RightZeroPadded(eHashNum, maxRows))
	run.AssignColumn(iPadd.cleanLimb.GetColID(), smartvectors.RightZeroPadded(eCleanLimb, maxRows))
	run.AssignColumn(iPadd.oneCol.GetColID(), smartvectors.RightZeroPadded(vector.Repeat(field.One(), len(eLimb)), maxRows))
}

// ExtendWithPadding extends the columns by adding rows to include the padding limbs.
func extendWithPadding(limb, nByte, hashNum, cleanLimb []field.Element) (
	extendedLimb, extendedNbyte []field.Element,
	extendedHashNum, extendedCleanLimb []field.Element,
	isNewHash, isInserted, isPadded []field.Element,
) {
	one := field.One()
	zero := field.Zero()
	lenLimb := len(limb)
	s := 0
	var paddingLimb, paddingNbyte, zeroes, ones, repeatHashNum []field.Element
	var nextMinusOne, a field.Element

	isNewHash = append(isNewHash, one)
	for j := 0; j < lenLimb; j++ {
		extendedLimb = append(extendedLimb, limb[j])
		extendedNbyte = append(extendedNbyte, nByte[j])
		extendedHashNum = append(extendedHashNum, hashNum[j])
		extendedCleanLimb = append(extendedCleanLimb, cleanLimb[j])
		isInserted = append(isInserted, one)
		isPadded = append(isPadded, zero)
		if j > 0 {
			a.Sub(&hashNum[j], &hashNum[j-1])
			isNewHash = append(isNewHash, a)

		}

		s = s + int(nByte[j].Uint64())
		if j != lenLimb-1 {
			// if a new hash is launched, pad the last block
			nextMinusOne.Sub(&hashNum[j+1], &one)
			if hashNum[j].Cmp(&nextMinusOne) == 0 {
				// insert new rows to include the padding limbs
				paddingLimb, paddingNbyte, zeroes, ones, repeatHashNum =
					insertPaddingRows(s%maxBlockSize, maxBlockSize, hashNum[j])
				extendedLimb = append(extendedLimb, paddingLimb...)
				extendedNbyte = append(extendedNbyte, paddingNbyte...)
				extendedCleanLimb = append(extendedCleanLimb, paddingLimb...)
				extendedHashNum = append(extendedHashNum, repeatHashNum...)
				isNewHash = append(isNewHash, zeroes...)
				isInserted = append(isInserted, zeroes...)
				isPadded = append(isPadded, ones...)
				s = 0
			}
		} else {
			// if it is the last limb in the column, pad the last block
			paddingLimb, paddingNbyte, zeroes, ones, repeatHashNum =
				insertPaddingRows(s%maxBlockSize, maxBlockSize, hashNum[j])
			extendedLimb = append(extendedLimb, paddingLimb...)
			extendedNbyte = append(extendedNbyte, paddingNbyte...)
			extendedHashNum = append(extendedHashNum, repeatHashNum...)
			extendedCleanLimb = append(extendedCleanLimb, paddingLimb...)
			isNewHash = append(isNewHash, zeroes...)
			isInserted = append(isInserted, zeroes...)
			isPadded = append(isPadded, ones...)
		}

	}
	return extendedLimb, extendedNbyte, extendedHashNum, extendedCleanLimb, isNewHash, isInserted, isPadded
}

// InsertPaddingRows receives the number of existing bytes in the block and complete the block by padding.
func insertPaddingRows(n, max int, hashNum field.Element) (
	limb, nbyte, zeroes, ones, repeatHashNum []field.Element,
) {
	zero := field.Zero()
	one := field.One()
	maxNByteFr := field.NewElement(maxNByte)
	remain := max - n

	if remain >= 2 {
		// applies the domain separator
		limb = append(limb, one)
		nbyte = append(nbyte, field.One())
		zeroes = append(zeroes, zero)
		ones = append(ones, one)
		repeatHashNum = append(repeatHashNum, hashNum)
		a := (remain - 2) % maxNByte
		b := (remain - 2) / maxNByte

		// zero pad on the right
		for i := 0; i < b; i++ {
			limb = append(limb, zero)
			nbyte = append(nbyte, maxNByteFr)
			zeroes = append(zeroes, zero)
			ones = append(ones, one)
			repeatHashNum = append(repeatHashNum, hashNum)

		}
		if a != 0 {
			limb = append(limb, zero)
			nbyte = append(nbyte, field.NewElement(uint64(a)))
			zeroes = append(zeroes, zero)
			ones = append(ones, one)
			repeatHashNum = append(repeatHashNum, hashNum)
		}
		// padding on the right with 0X80
		limb = append(limb, field.NewElement(128))
		nbyte = append(nbyte, one)
		zeroes = append(zeroes, zero)
		ones = append(ones, one)
		repeatHashNum = append(repeatHashNum, hashNum)
	}
	// padding with 0x81, for padding with a single byte
	if remain == 1 {
		limb = append(limb, field.NewElement(129))
		nbyte = append(nbyte, one)
		zeroes = append(zeroes, zero)
		ones = append(ones, one)
		repeatHashNum = append(repeatHashNum, hashNum)
	}
	// sanity check
	if len(zeroes) != len(repeatHashNum) {
		utils.Panic(" they should have the same length")
	}
	return limb, nbyte, zeroes, ones, repeatHashNum
}

// It receives the assignment for the column 'nByte' extended by padding
// and set isBlockComplete to 1 whenever the block is complete.
func (iPadd *importAndPadd) assignIsBlockComplete(
	run *wizard.ProverRuntime,
	eNByte, isNewHash, eCleanLimb []field.Element,
	permTrace permTrace.PermTraces,
	maxRows int,
) {
	witLen := len(eNByte)
	isEndOfBlock := make([]field.Element, witLen)
	ctr := 0
	s := eNByte[0].Uint64()

	var block [][]byte
	var stream []byte
	nbyte := eNByte[0].Uint64()
	limbBytes := eCleanLimb[0].Bytes()
	usefulBytes := limbBytes[32-nbyte:]
	stream = append(stream, usefulBytes[:nbyte]...)
	for j := 1; j < witLen; j++ {

		// sanity check
		if isNewHash[j] == field.One() && s != 0 {
			utils.Panic(" the last block should be complete before launching a new hash")
		}
		nbyte := eNByte[j].Uint64()
		s = s + nbyte
		limbBytes := eCleanLimb[j].Bytes()
		usefulBytes := limbBytes[32-nbyte:]
		if s > blockSize || s == blockSize {
			s = s - blockSize
			res := usefulBytes[:(nbyte - s)]
			newBlock := append(stream, res...)
			block = append(block, newBlock)
			stream = usefulBytes[(nbyte - s):nbyte]
			isEndOfBlock[j] = field.One()
			ctr++
		} else {
			stream = append(stream, usefulBytes[:nbyte]...)
		}

	}
	run.AssignColumn(iPadd.isBlockComplete.GetColID(), smartvectors.RightZeroPadded(isEndOfBlock, maxRows))
	iPadd.numBlocks = ctr

	if len(permTrace.Blocks) != iPadd.numBlocks {
		utils.Panic("the length is not correct, by permTrace %v, by iPadd %v", len(permTrace.Blocks), iPadd.numBlocks)
	}
	for j := range block {
		if len(block[j]) != blockSize {
			utils.Panic("the length of the %v-th block is %v", j, len(block[j]))
		}

		if permTrace.Blocks[j] != *bytesAsBlockPtrUnsafe(block[j]) {
			utils.Panic("traces are different for block %v", j)

		}

	}
}
