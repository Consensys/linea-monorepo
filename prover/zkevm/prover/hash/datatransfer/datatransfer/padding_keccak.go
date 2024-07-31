package datatransfer

import (
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/symbolic"
	sym "github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
)

// InsertPadding asserts if the padding over the given (gbm) columns is done correctly.
/*
padding is expected to be done via inserting new limbs;
 -  limb 1 for the first padded limb, 0 bytes for the middle limbs, and 128 for the last limb
 -  limb 129 if we require only one byte of padding
*/
func (iPadd *importAndPadd) insertPaddingKeccak(comp *wizard.CompiledIOP, round int, lookUpKeccakPad ifaces.Column) {

	/* 	if isPadded[i-1] = 0, isPadded[i] = 1, isPadded[i+1] =1 ----> limb = 1, nByte = 1
		if isPadded[i-1] = 1, isPadded[i] = 1, isPadded[i+1] =0 ----> limb = 128, nByte = 1
		if isPadded[i-1] = 0, isPadded[i] = 1, isPadded[i+1] =0 ----> limb = 129 , nByte = 1
		if isPadded[i-1] = 1, isPadded[i] = 1, isPadded[i+1] =1 ----> limb = 0
	 the constraints over NBytes also guarantees the correct number of  padded zeroes.*/
	dsv := symbolic.NewConstant(1)       // domain separator value, for padding
	fpv := symbolic.NewConstant(128)     // final padding value
	fpvPlus := symbolic.NewConstant(129) // padding value for a single byte padding

	isPaddedMinus := column.Shift(iPadd.isPadded, -1)
	isPaddedPlus := column.Shift(iPadd.isPadded, 1)

	firstPad := sym.Mul(sym.Sub(1, isPaddedMinus), iPadd.isPadded, isPaddedPlus,
		sym.Sub(iPadd.limb, dsv))
	firstPadLen := sym.Mul(sym.Sub(1, isPaddedMinus), iPadd.isPadded, isPaddedPlus,
		sym.Sub(iPadd.nByte, 1))

	lastPad := sym.Mul(isPaddedMinus, iPadd.isPadded, sym.Sub(1, isPaddedPlus),
		sym.Sub(iPadd.limb, fpv))
	lastPaddLen := sym.Mul(isPaddedMinus, iPadd.isPadded, sym.Sub(1, isPaddedPlus),
		sym.Sub(iPadd.nByte, 1))

	middlePads := sym.Mul(isPaddedMinus, iPadd.isPadded, isPaddedPlus, iPadd.limb)

	singlePad := sym.Mul(sym.Sub(1, isPaddedMinus), iPadd.isPadded, sym.Sub(1, isPaddedPlus),
		sym.Sub(iPadd.limb, fpvPlus))

	comp.InsertGlobal(round, ifaces.QueryIDf("FirstPad"), firstPad)
	comp.InsertGlobal(round, ifaces.QueryIDf("FirstPadLen"), firstPadLen)
	comp.InsertGlobal(round, ifaces.QueryIDf("MiddlePads"), middlePads)
	comp.InsertGlobal(round, ifaces.QueryIDf("LastPad"), lastPad)
	comp.InsertGlobal(round, ifaces.QueryIDf("LastPadLen"), lastPaddLen)
	comp.InsertGlobal(round, ifaces.QueryIDf("SinglePad"), singlePad)

	comp.InsertInclusionConditionalOnIncluded(round, ifaces.QueryIDf("LOOKUP_NB_ZeroPadded"),
		[]ifaces.Column{lookUpKeccakPad}, []ifaces.Column{iPadd.accPaddedBytes}, iPadd.isPadded)
}

// InsertPaddingRows receives the number of existing bytes in the block and complete the block by padding.
func insertPaddingRowsKeccak(n, max int, hashNum field.Element, lastIndex uint64) (
	limb, nbyte, index, zeroes, ones, repeatHashNum []field.Element,
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
		lastIndex++
		index = append(index, field.NewElement(lastIndex))
		a := (remain - 2) % maxNByte
		b := (remain - 2) / maxNByte

		// zero pad on the right
		for i := 0; i < b; i++ {
			limb = append(limb, zero)
			nbyte = append(nbyte, maxNByteFr)
			zeroes = append(zeroes, zero)
			ones = append(ones, one)
			repeatHashNum = append(repeatHashNum, hashNum)
			lastIndex++
			index = append(index, field.NewElement(lastIndex))

		}
		if a != 0 {
			limb = append(limb, zero)
			nbyte = append(nbyte, field.NewElement(uint64(a)))
			zeroes = append(zeroes, zero)
			ones = append(ones, one)
			repeatHashNum = append(repeatHashNum, hashNum)
			lastIndex++
			index = append(index, field.NewElement(lastIndex))
		}
		// padding on the right with 0X80
		limb = append(limb, field.NewElement(128))
		nbyte = append(nbyte, one)
		zeroes = append(zeroes, zero)
		ones = append(ones, one)
		repeatHashNum = append(repeatHashNum, hashNum)
		lastIndex++
		index = append(index, field.NewElement(lastIndex))
	}
	// padding with 0x81, for padding with a single byte
	if remain == 1 {
		limb = append(limb, field.NewElement(129))
		nbyte = append(nbyte, one)
		zeroes = append(zeroes, zero)
		ones = append(ones, one)
		repeatHashNum = append(repeatHashNum, hashNum)
		lastIndex++
		index = append(index, field.NewElement(lastIndex))
	}
	// sanity check
	if len(zeroes) != len(repeatHashNum) {
		utils.Panic(" they should have the same length")
	}
	return limb, nbyte, index, zeroes, ones, repeatHashNum
}
