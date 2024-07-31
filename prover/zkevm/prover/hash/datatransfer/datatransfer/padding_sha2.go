package datatransfer

import (
	"github.com/consensys/zkevm-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/zkevm-monorepo/prover/maths/field"
	"github.com/consensys/zkevm-monorepo/prover/protocol/column"
	"github.com/consensys/zkevm-monorepo/prover/protocol/ifaces"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
	"github.com/consensys/zkevm-monorepo/prover/symbolic"
	sym "github.com/consensys/zkevm-monorepo/prover/symbolic"
	"github.com/consensys/zkevm-monorepo/prover/utils"
)

// InsertPadding asserts if the padding over the given (gbm) columns is done correctly.
/*
padding is expected to be done via inserting new limbs;
 -  limb 0x80 for the first padded limb, 0 bytes for the middle limbs, and len<<=3 for the last limb
 _  last limb should be 8 bytes
 -  the number of zeroes should not be more than block-size
    (this prevents attacks where a full block of zeroes is padded,
	the cases where less than required zeroes are padded is prevented later, by the constraint that blocks should be full)
*/
func (iPadd importAndPadd) insertPaddingSha2(comp *wizard.CompiledIOP, round int, lookUpSha2Pad ifaces.Column) {

	//  commit to the column streamLen,
	//  it keeps (lenStream<<=3) in front of the last limb.
	streamLen := comp.InsertCommit(round, ifaces.ColIDf("StreamLen"), iPadd.isActive.Size())

	comp.SubProvers.AppendToInner(round, func(run *wizard.ProverRuntime) {
		witSize := smartvectors.Density(iPadd.isActive.GetColAssignment(run))
		nByte := iPadd.nByte.GetColAssignment(run).IntoRegVecSaveAlloc()
		isInserted := iPadd.isInserted.GetColAssignment(run).IntoRegVecSaveAlloc()
		isPadded := iPadd.isPadded.GetColAssignment(run).IntoRegVecSaveAlloc()

		streamLenWit := make([]field.Element, witSize)
		s := uint64(0)
		for i := 0; i < witSize; i++ {
			if isPadded[i].IsZero() {
				s = s + nByte[i].Uint64()
				streamLenWit[i] = field.NewElement(s)
			} else {
				streamLenWit[i] = field.NewElement(s)
				if isInserted[i+1].IsOne() || i == witSize-1 {
					s <<= 3
					streamLenWit[i] = field.NewElement(s)
					s = 0
				}
			}
		}

		run.AssignColumn(streamLen.GetColID(),
			smartvectors.RightZeroPadded(streamLenWit, iPadd.isActive.Size()))
	})

	// streamLen[0] = nByte[0]
	comp.InsertLocal(round, ifaces.QueryIDf("streamLen_Local"), sym.Sub(streamLen, iPadd.nByte))

	// accumulate nBytes of the original stream (thus ignoring the padded ones)
	// for each new stream restart the accumulator
	isNotRestarted := sym.Sub(1, sym.Mul(column.Shift(iPadd.isPadded, -1), iPadd.isInserted))
	a := sym.Add(sym.Mul(column.Shift(streamLen, -1), isNotRestarted),
		sym.Mul(iPadd.nByte, iPadd.isInserted)) // ignore the padded ones

	// shift the streamLen by 3 over the last limb
	isLastLimb := sym.Add(sym.Mul(iPadd.isPadded, column.Shift(iPadd.isInserted, 1)),
		sym.Mul(iPadd.isActive, sym.Sub(1, column.Shift(iPadd.isActive, 1))))

	b := sym.Add(sym.Mul(a, sym.Sub(1, isLastLimb)), sym.Mul(a, 8, isLastLimb))
	comp.InsertGlobal(round, ifaces.QueryIDf("StreamLen_Glob"), sym.Mul(sym.Sub(b, streamLen), iPadd.isActive))

	iPadd.csZeroPadding(comp, round)
	/* 	if isPadded[i-1] = 0, isPadded[i] = 1, isPadded[i+1] =1 ----> limb = 0x80 and nByte = 1
		if isPadded[i-1] = 1, isPadded[i] = 1, isPadded[i+1] =0 ----> limb = streamLen and nByte = 8
		if isPadded[i-1] = 1, isPadded[i] = 1, isPadded[i+1] =1 ----> limb = 0
	 the constraints over NBytes also guarantees the correct number of  padded zeroes.*/

	dsv := symbolic.NewConstant(128) // domain separator value, for padding

	isPaddedMinus := column.Shift(iPadd.isPadded, -1)
	isPaddedPlus := column.Shift(iPadd.isPadded, 1)

	firstPad := sym.Mul(sym.Sub(1, isPaddedMinus), iPadd.isPadded,
		isPaddedPlus, sym.Sub(iPadd.limb, dsv))

	firstPadLen := sym.Mul(sym.Sub(1, isPaddedMinus), iPadd.isPadded,
		isPaddedPlus, sym.Sub(iPadd.nByte, 1))

	lastPad := sym.Mul(isPaddedMinus, iPadd.isPadded, sym.Sub(1, isPaddedPlus), sym.Sub(iPadd.limb, streamLen))

	lastPadLen := sym.Mul(isPaddedMinus, iPadd.isPadded, sym.Sub(1, isPaddedPlus), sym.Sub(iPadd.nByte, 8))

	middlePad := sym.Mul(isPaddedMinus, iPadd.isPadded, isPaddedPlus, iPadd.limb)

	comp.InsertGlobal(round, ifaces.QueryIDf("FirstPad"), firstPad)
	comp.InsertGlobal(round, ifaces.QueryIDf("FirstPadLen"), firstPadLen)
	comp.InsertGlobal(round, ifaces.QueryIDf("MiddlePads"), middlePad)
	comp.InsertGlobal(round, ifaces.QueryIDf("lastPad"), lastPad)
	comp.InsertGlobal(round, ifaces.QueryIDf("lastPadLen"), lastPadLen)

	comp.InsertInclusionConditionalOnIncluded(round, ifaces.QueryIDf("LOOKUP_NB_ZeroPadded"),
		[]ifaces.Column{lookUpSha2Pad}, []ifaces.Column{iPadd.accPaddedBytes}, iPadd.isPadded)

}

// InsertPaddingRows receives the number of existing bytes in the block and complete the block by padding.
func insertPaddingRowsSha2(streamLen, max int, hashNum field.Element, lastIndex uint64) (
	limb, nbyte, index, zeroes, ones, repeatHashNum []field.Element,
) {
	maxNByteFr := field.NewElement(16)
	dsv := field.NewElement(128)
	zero := field.Zero()
	one := field.One()

	n := streamLen % max

	// applies the domain separator
	limb = append(limb, dsv)
	nbyte = append(nbyte, field.One())
	zeroes = append(zeroes, zero)
	ones = append(ones, one)
	repeatHashNum = append(repeatHashNum, hashNum)
	lastIndex++
	index = append(index, field.NewElement(lastIndex))
	var a, b int
	if n < 56 {
		remain := max - n - 9 //
		a = remain % maxNByte
		b = remain / maxNByte
	} else {
		// from the current block it has remained (max - 1 - n)
		// and (max - 8) from the next block
		remain := (max - 1 - n) + (max - 8)
		a = remain % maxNByte
		b = remain / maxNByte
	}
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
	// padding on the right with len<<=3
	streamLen <<= 3
	limb = append(limb, field.NewElement(uint64(streamLen)))
	nbyte = append(nbyte, field.NewElement(8))
	zeroes = append(zeroes, zero)
	ones = append(ones, one)
	repeatHashNum = append(repeatHashNum, hashNum)
	lastIndex++
	index = append(index, field.NewElement(lastIndex))

	// sanity check
	if len(zeroes) != len(repeatHashNum) {
		utils.Panic(" they should have the same length")
	}
	return limb, nbyte, index, zeroes, ones, repeatHashNum
}
