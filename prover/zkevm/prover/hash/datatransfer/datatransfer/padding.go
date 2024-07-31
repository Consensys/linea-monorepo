package datatransfer

import (
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/common/vector"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	sym "github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/generic"
)

// it check that the number of padded zeroes for hash is not larger than the block size. This prevents the attack where the prover appends zero blocks.
func (iPadd importAndPadd) csZeroPadding(comp *wizard.CompiledIOP, round int) {

	// accPaddedBytes[0] = nByte[0] * iPadded[0]
	comp.InsertLocal(round, ifaces.QueryIDf("AccPaddedBytes_Loc"),
		sym.Sub(iPadd.accPaddedBytes, sym.Mul(iPadd.nByte, iPadd.isPadded)))

	// accPaddedBytes[i] = (accPaddedBytes[i-1] + nByte[i]) * iPadded[i]
	// for i index of the row
	comp.InsertGlobal(round, ifaces.QueryIDf("AccPaddedBytes_Glob"),
		sym.Sub(
			iPadd.accPaddedBytes,
			sym.Mul(
				sym.Add(column.Shift(iPadd.accPaddedBytes, -1), iPadd.nByte),
				iPadd.isPadded),
		))

}

// It assign the native columns specific to the module,
// the columns are extended by padding.
func (iPadd *importAndPadd) assignImportAndPadd(
	run *wizard.ProverRuntime,
	gt generic.GenTrace,
	maxRows, hashType int) {
	isSelected := gt.TO_HASH
	one := field.One()
	var hashNum, limb, nByte, index, cleanLimb []field.Element
	for i := range isSelected {
		if isSelected[i].Cmp(&one) == 0 {
			hashNum = append(hashNum, gt.HashNum[i])
			limb = append(limb, gt.Limb[i])
			nByte = append(nByte, gt.NByte[i])
			index = append(index, gt.Index[i])
			cleanLimb = append(cleanLimb, gt.CleanLimb[i])
		}
	}

	// extend the columns to include the padded limbs
	eLimb, eNbyte, eHashNum, eIndex, eCleanLimb, isNewHash, isInserted, isPadded :=
		extendWithPadding(limb, nByte, hashNum, index, cleanLimb, hashType)

	// sanity check
	if len(eHashNum) != len(isNewHash) {
		utils.Panic("HashNum and  isNewHash have different sizes  %v, %v ",
			len(eHashNum), len(isNewHash))
	}

	// assign the columns
	run.AssignColumn(iPadd.isNewHash.GetColID(), smartvectors.RightZeroPadded(isNewHash, maxRows))
	run.AssignColumn(iPadd.limb.GetColID(), smartvectors.RightZeroPadded(eLimb, maxRows))
	run.AssignColumn(iPadd.nByte.GetColID(), smartvectors.RightZeroPadded(eNbyte, maxRows))
	run.AssignColumn(iPadd.isActive.GetColID(), smartvectors.RightZeroPadded(vector.Repeat(one, len(eLimb)), maxRows))
	run.AssignColumn(iPadd.isInserted.GetColID(), smartvectors.RightZeroPadded(isInserted, maxRows))
	run.AssignColumn(iPadd.isPadded.GetColID(), smartvectors.RightZeroPadded(isPadded, maxRows))
	run.AssignColumn(iPadd.hashNum.GetColID(), smartvectors.RightZeroPadded(eHashNum, maxRows))
	run.AssignColumn(iPadd.index.GetColID(), smartvectors.RightZeroPadded(eIndex, maxRows))
	run.AssignColumn(iPadd.cleanLimb.GetColID(), smartvectors.RightZeroPadded(eCleanLimb, maxRows))
	run.AssignColumn(iPadd.oneCol.GetColID(), smartvectors.RightZeroPadded(vector.Repeat(field.One(), len(eLimb)), maxRows))

	accPaddedBytes := make([]field.Element, len(eNbyte))
	if len(eNbyte) != 0 {
		accPaddedBytes[0].Mul(&eNbyte[0], &isPadded[0])
	}
	for i := 1; i < len(eNbyte); i++ {
		accPaddedBytes[i].Add(&accPaddedBytes[i-1], &eNbyte[i])
		accPaddedBytes[i].Mul(&accPaddedBytes[i], &isPadded[i])
	}

	run.AssignColumn(iPadd.accPaddedBytes.GetColID(), smartvectors.RightZeroPadded(accPaddedBytes, maxRows))
}

// ExtendWithPadding extends the columns by adding rows to include the padding limbs.
// for Keccak it uses hashType = 0, for Sha2 hashType = 1.
func extendWithPadding(limb, nByte, hashNum, index, cleanLimb []field.Element, hashType int) (
	extendedLimb, extendedNbyte []field.Element,
	extendedHashNum, extendedIndex []field.Element,
	extendedCleanLimb []field.Element,
	isNewHash, isInserted, isPadded []field.Element,
) {
	one := field.One()
	zero := field.Zero()
	lenLimb := len(limb)
	s := 0
	var paddingLimb, paddingNbyte, paddingIndex, zeroes, ones, repeatHashNum []field.Element

	for j := 0; j < lenLimb; j++ {
		extendedLimb = append(extendedLimb, limb[j])
		extendedNbyte = append(extendedNbyte, nByte[j])
		extendedHashNum = append(extendedHashNum, hashNum[j])
		extendedIndex = append(extendedIndex, index[j])
		extendedCleanLimb = append(extendedCleanLimb, cleanLimb[j])
		isInserted = append(isInserted, one)
		isPadded = append(isPadded, zero)
		if index[j].Uint64() == 0 {
			isNewHash = append(isNewHash, field.One())
		} else {
			isNewHash = append(isNewHash, field.Zero())
		}
		s = s + int(nByte[j].Uint64())
		if j != lenLimb-1 {
			// if a new hash is about to launched, pad the last block
			if index[j+1].Uint64() == 0 {
				// insert new rows to include the padding limbs
				switch hashType {
				case Keccak:
					paddingLimb, paddingNbyte, paddingIndex, zeroes, ones, repeatHashNum =
						insertPaddingRowsKeccak(s%maxBlockSize, maxBlockSize, hashNum[j], index[j].Uint64())
				case Sha2:
					paddingLimb, paddingNbyte, paddingIndex, zeroes, ones, repeatHashNum =
						insertPaddingRowsSha2(s, maxBlockSizeSha2, hashNum[j], index[j].Uint64())
				default:
					utils.Panic("The hashType is not supported")
				}
				extendedLimb = append(extendedLimb, paddingLimb...)
				extendedNbyte = append(extendedNbyte, paddingNbyte...)
				extendedIndex = append(extendedIndex, paddingIndex...)
				extendedCleanLimb = append(extendedCleanLimb, paddingLimb...)
				extendedHashNum = append(extendedHashNum, repeatHashNum...)
				isNewHash = append(isNewHash, zeroes...)
				isInserted = append(isInserted, zeroes...)
				isPadded = append(isPadded, ones...)
				s = 0
			}
		} else {
			// if it is the last limb in the column, pad the last block
			switch hashType {
			case Keccak:
				paddingLimb, paddingNbyte, paddingIndex, zeroes, ones, repeatHashNum =
					insertPaddingRowsKeccak(s%maxBlockSize, maxBlockSize, hashNum[j], index[j].Uint64())
			case Sha2:
				paddingLimb, paddingNbyte, paddingIndex, zeroes, ones, repeatHashNum =
					insertPaddingRowsSha2(s, maxBlockSizeSha2, hashNum[j], index[j].Uint64())
			default:
				utils.Panic("The hashType is not supported")
			}
			extendedLimb = append(extendedLimb, paddingLimb...)
			extendedNbyte = append(extendedNbyte, paddingNbyte...)
			extendedHashNum = append(extendedHashNum, repeatHashNum...)
			extendedIndex = append(extendedIndex, paddingIndex...)
			extendedCleanLimb = append(extendedCleanLimb, paddingLimb...)
			isNewHash = append(isNewHash, zeroes...)
			isInserted = append(isInserted, zeroes...)
			isPadded = append(isPadded, ones...)
		}

	}
	return extendedLimb, extendedNbyte, extendedHashNum, extendedIndex, extendedCleanLimb, isNewHash, isInserted, isPadded
}
