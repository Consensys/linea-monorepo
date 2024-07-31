package datatransfer

import (
	"math/big"

	"github.com/consensys/zkevm-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/zkevm-monorepo/prover/maths/field"
	"github.com/consensys/zkevm-monorepo/prover/protocol/ifaces"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
	"github.com/consensys/zkevm-monorepo/prover/symbolic"
	"github.com/consensys/zkevm-monorepo/prover/zkevm/prover/hash/datatransfer/dedicated"
)

// CleanLimbDecomposition (CLD) module is responsible for cleaning the limbs,
// and decomposing them to slices.
type cleanLimbDecomposition struct {
	/*
		To reach length 16 bytes, Limbs are padded with zeroes (as LSB).
		 The column "nbZeroes" stands for these redundant zeroes of limbs.
		 To clean the limb we have; cleanLimb := limb*2^(-nbZeroes * 8)
		 The column "powersNbZeros" is equivalent with 2^(8 * nbZeroes) over each element.
		 The equivalence of "powersNbZeros" and 2^(nbZeroes * 8) is checked by lookup tables.
		 Putting together we have; limb := cleanLimb*(powersNbZeros)
	*/
	nbZeros ifaces.Column
	// powersNbZeros represent powers of nbZeroes;  powersNbZeros = 2^(8 * nbZeroes)
	powersNbZeros ifaces.Column
	// CleanLimbs' decomposition  in slices (of different sizes)
	cld []ifaces.Column
	// It is the length of the slices from the decomposition
	cldLen []ifaces.Column
	// It is the binary counterpart of cldLen.
	// Namely, it is zero iff cldLen is zero, otherwise it is one.
	cldLenBinary []ifaces.Column
	// cldLenPowers = 2^(8*cldLen)
	cldLenPowers []ifaces.Column

	// decomposition of cld to cldSlices (each slice is a single byte)
	cldSlices [][]ifaces.Column
	// length of cldSlices
	lenCldSlices [][]ifaces.Column

	nbCld, nbCldSlices int
}

/*
NewCLD creates a new CLD module, defining the columns and constraints asserting to the following facts:

 1. cld columns are the decomposition of clean limbs

 2. cldLen is 1 iff cld is not empty
*/
func (cld *cleanLimbDecomposition) newCLD(
	comp *wizard.CompiledIOP,
	round int,
	lu lookUpTables,
	iPadd importAndPadd,
	maxRows, hashType int,
) {
	switch hashType {
	case Keccak:
		{
			cld.nbCld = maxLanesFromLimb
			cld.nbCldSlices = numBytesInLane
		}
	case Sha2:
		{
			cld.nbCld = maxLanesFromLimbSha2
			cld.nbCldSlices = numBytesInLaneSha2
		}
	}

	// Declare the columns
	cld.insertCommit(comp, round, maxRows)

	// Declare the constraints

	// Constraints over the equivalence of "powersNbZeros" with "2^(8 * NbZeros)"
	cld.csNbZeros(comp, round, lu, iPadd)

	//  Constraints over the form of cldLen;
	//  -  each row should be of the form (1,...,1,O,...,0),
	//  -  the number of ones in each row is NBytes
	cld.csDecomposLen(comp, round, iPadd)

	cld.csDecomposeCLDToSlices(comp, round, lu, iPadd)
}
func (cld *cleanLimbDecomposition) insertCommit(
	comp *wizard.CompiledIOP, round, maxRows int) {

	cld.cldSlices = make([][]ifaces.Column, cld.nbCld)
	cld.lenCldSlices = make([][]ifaces.Column, cld.nbCld)

	for x := 0; x < cld.nbCld; x++ {
		cld.cld = append(cld.cld, comp.InsertCommit(round, ifaces.ColIDf("CLD_%v", x), maxRows))
		cld.cldLen = append(cld.cldLen, comp.InsertCommit(round, ifaces.ColIDf("CLD_Len_%v", x), maxRows))
		cld.cldLenBinary = append(cld.cldLenBinary, comp.InsertCommit(round, ifaces.ColIDf("CLD_Len_Binary_%v", x), maxRows))
		cld.cldLenPowers = append(cld.cldLenPowers, comp.InsertCommit(round, ifaces.ColIDf("CLD_Len_Powers_%v", x), maxRows))

		for k := 0; k < cld.nbCldSlices; k++ {
			cld.cldSlices[x] = append(cld.cldSlices[x], comp.InsertCommit(round, ifaces.ColIDf("CLD_Slice_%v_%v", x, k), maxRows))
			cld.lenCldSlices[x] = append(cld.lenCldSlices[x], comp.InsertCommit(round, ifaces.ColIDf("Len_CLD_Slice_%v_%v", x, k), maxRows))
		}
	}
	cld.nbZeros = comp.InsertCommit(round, ifaces.ColIDf("NbZeros"), maxRows)
	cld.powersNbZeros = comp.InsertCommit(round, ifaces.ColIDf("PowersNbZeros"), maxRows)

}

// csNbZeros imposes the constraints between nbZero and powersNbZeros;
// -  powersNbZeros = 2^(nbZeros * 8)
//
// -  nbZeros = 16 - nByte
func (cld *cleanLimbDecomposition) csNbZeros(
	comp *wizard.CompiledIOP,
	round int,
	lookUp lookUpTables,
	iPadd importAndPadd,
) {
	// Equivalence of "PowersNbZeros" with "2^(NbZeros * 8)"
	comp.InsertInclusion(round,
		ifaces.QueryIDf("NumToPowers"), []ifaces.Column{lookUp.colNumber, lookUp.colPowers},
		[]ifaces.Column{cld.nbZeros, cld.powersNbZeros})

	//  The constraint for nbZeros = (16 - NByte)* isActive
	maxNByte := symbolic.NewConstant(field.NewElement(uint64(maxNByte)))
	nbZeros := maxNByte.Sub(ifaces.ColumnAsVariable(iPadd.nByte))
	comp.InsertGlobal(round,
		ifaces.QueryIDf("NbZeros"), symbolic.Mul(symbolic.Sub(nbZeros, cld.nbZeros), iPadd.isActive))

	// Equivalence of "cldLenPowers" with "2^(cldLen * 8)"
	for j := range cld.cld {
		comp.InsertInclusion(round,
			ifaces.QueryIDf("CldLenPowers_%v", j), []ifaces.Column{lookUp.colNumber, lookUp.colPowers},
			[]ifaces.Column{cld.cldLen[j], cld.cldLenPowers[j]})
	}
}

// /  Constraints over the form of cldLenBinary and cldLen (similarly on lenCdSlices);
//   - cldLenBinary is binary
//   - each row should be of the form (0,...,0,1,...,1),
//   - cldLen over a row adds up to NBytes
//   - cldLenBinary = 1 iff cldLen != 0
func (cld *cleanLimbDecomposition) csDecomposLen(
	comp *wizard.CompiledIOP,
	round int,
	iPadd importAndPadd,
) {
	// The rows of cldLen adds up to NByte; \sum_i cldLen[i]=NByte
	s := symbolic.NewConstant(0)
	for j := range cld.cld {
		s = symbolic.Add(s, cld.cldLen[j])

		// cldLenBinary is binary
		comp.InsertGlobal(round, ifaces.QueryIDf("cldLenBinary_IsBinary_%v", j),
			symbolic.Mul(cld.cldLenBinary[j], symbolic.Sub(1, cld.cldLenBinary[j])))

		// cldLenBinary = 1 iff cldLen !=0
		dedicated.InsertIsTargetValue(comp, round, ifaces.QueryIDf("IsOne_IFF_IsNonZero_%v", j),
			field.Zero(), cld.cldLen[j], symbolic.Sub(1, cld.cldLenBinary[j]))

		if j < len(cld.cld)-1 {
			// a should be binary
			a := symbolic.Sub(cld.cldLenBinary[j+1], cld.cldLenBinary[j])
			comp.InsertGlobal(round, ifaces.QueryIDf("FirstZeros_ThenOnes_%v", j),
				symbolic.Mul(a, symbolic.Sub(1, a)))
		}
	}
	// \sum_i cldLen[i]=NByte
	comp.InsertGlobal(round, ifaces.QueryIDf("cldLen_IsNByte"), symbolic.Sub(s, iPadd.nByte))

	// constraints over lenCldSlices,
	//   - lenCldSlices is binary
	//   - each row should be of the form (0,...,0,1,...,1),
	//   - lenCldSlices over a row adds up to cldLen
	for j := range cld.cld {
		sum := symbolic.NewConstant(0)
		for k := 0; k < cld.nbCldSlices; k++ {
			sum = symbolic.Add(sum, cld.lenCldSlices[j][k])

			// lenCldSlices is binary
			comp.InsertGlobal(round, ifaces.QueryIDf("lenCldSlices_IsBinary_%v_%v", j, k),
				symbolic.Mul(cld.lenCldSlices[j][k], symbolic.Sub(1, cld.lenCldSlices[j][k])))

			if k < cld.nbCldSlices-1 {
				// a should be binary
				a := symbolic.Sub(cld.lenCldSlices[j][k+1], cld.lenCldSlices[j][k])
				comp.InsertGlobal(round, ifaces.QueryIDf("FirstZeros_ThenOnes_lenCldSlices_%v_%v", j, k),
					symbolic.Mul(a, symbolic.Sub(1, a)))
			}

		}

	}

}

func (cld *cleanLimbDecomposition) csDecomposeCLDToSlices(
	comp *wizard.CompiledIOP,
	round int, lu lookUpTables,
	iPadd importAndPadd,
) {
	for j := range cld.cld {
		// constraint asserting to the correct decomposition of cld to slices
		cldRec := baseRecomposeByLengthHandles(cld.cldSlices[j][:], power8, cld.lenCldSlices[j][:])
		comp.InsertGlobal(round, ifaces.QueryIDf("Decompos_CLD_%v", j), symbolic.Sub(cldRec, cld.cld[j]))
	}

	// decompositions are single bytes
	for j := range cld.cldSlices {
		for k := range cld.cldSlices[0] {
			comp.InsertInclusion(round, ifaces.QueryIDf("SingleByte-Decomposition_CLD_%v_%v", j, k),
				[]ifaces.Column{lu.colSingleByte}, []ifaces.Column{cld.cldSlices[j][k]})
		}
	}

	// recomposition of slices to limbs (equivalently, recomposition of cld to limbs)
	var slices, lenSlices []ifaces.Column
	for j := len(cld.cldSlices) - 1; j >= 0; j-- {
		slices = append(slices, cld.cldSlices[j][:]...)
		lenSlices = append(lenSlices, cld.lenCldSlices[j][:]...)
	}
	cleanLimb := baseRecomposeBinaryLen(slices, power8, lenSlices)

	res := symbolic.Mul(cld.powersNbZeros, cleanLimb)

	// the padded ones are already clean so handle them separately
	limb := symbolic.Add(symbolic.Mul(res, iPadd.isInserted), symbolic.Mul(cleanLimb, iPadd.isPadded))
	comp.InsertGlobal(round, ifaces.QueryIDf("LimbDecomposition"),
		symbolic.Sub(limb, iPadd.limb))

}

// it assign the slices for cld
func (cld *cleanLimbDecomposition) assignCLDSlices(run *wizard.ProverRuntime, maxNumRows int) {

	witSize := smartvectors.Density(cld.cld[0].GetColAssignment(run))
	cldWit := make([][]field.Element, cld.nbCld)
	cldLenWit := make([][]field.Element, cld.nbCld)
	for j := range cld.cld {
		cldWit[j] = make([]field.Element, witSize)
		cldWit[j] = cld.cld[j].GetColAssignment(run).IntoRegVecSaveAlloc()[:witSize]
		cldLenWit[j] = cld.cldLen[j].GetColAssignment(run).IntoRegVecSaveAlloc()[:witSize]
	}

	// populate lenCldSlices
	for j := range cld.cld {
		lenCldSlices := make([][]field.Element, cld.nbCldSlices)
		cldSlices := make([][]field.Element, cld.nbCldSlices)

		for i := range cldWit[0] {
			dec := getZeroOnes(cldLenWit[j][i], cld.nbCldSlices)
			a := decomposeByLengthFr(cldWit[j][i], int(cldLenWit[j][i].Uint64()), dec)
			for k := range cldSlices {
				lenCldSlices[k] = append(lenCldSlices[k], dec[k])
				cldSlices[k] = append(cldSlices[k], a[k])
			}
		}

		for k := range cld.cldSlices[0] {
			run.AssignColumn(cld.lenCldSlices[j][k].GetColID(),
				smartvectors.RightZeroPadded(lenCldSlices[k], maxNumRows))

			run.AssignColumn(cld.cldSlices[j][k].GetColID(),
				smartvectors.RightZeroPadded(cldSlices[k], maxNumRows))
		}
	}
}

// assignCLD assigns the columns specific to the CLD module (i.e., cld and cldLen, ...).
func (cld *cleanLimbDecomposition) assignCLD(run *wizard.ProverRuntime, iPadd importAndPadd,
	maxRows int) {
	witnessSize := smartvectors.Density(iPadd.limb.GetColAssignment(run))

	// Assign nbZeros and powersNbZeros
	var nbZeros []field.Element
	var powersNbZeros []field.Element
	fr16 := field.NewElement(16)
	var res field.Element
	var a big.Int
	nByte := run.GetColumn(iPadd.nByte.GetColID())
	for i := 0; i < witnessSize; i++ {
		b := nByte.Get(i)
		res.Sub(&fr16, &b)
		nbZeros = append(nbZeros, res)
		res.BigInt(&a)
		res.Exp(field.NewElement(power8), &a)
		powersNbZeros = append(powersNbZeros, res)

	}

	run.AssignColumn(cld.nbZeros.GetColID(), smartvectors.RightZeroPadded(nbZeros, maxRows))
	run.AssignColumn(cld.powersNbZeros.GetColID(), smartvectors.RightPadded(powersNbZeros, field.One(), maxRows))

	// Assign the columns cld and cldLen
	cldLen := make([][]field.Element, cld.nbCld)
	cldCol := make([][]field.Element, cld.nbCld)
	cldLenBinary := make([][]field.Element, cld.nbCld)

	for j := range cldCol {
		cldCol[j] = make([]field.Element, witnessSize)
		cldLen[j] = make([]field.Element, witnessSize)
		cldLenBinary[j] = make([]field.Element, witnessSize)
	}
	// assign row-by-row
	cleanLimb := run.GetColumn(iPadd.cleanLimb.GetColID()).IntoRegVecSaveAlloc()[:witnessSize]
	nByteWit := nByte.IntoRegVecSaveAlloc()
	cldLen = cutAndStitch(nByteWit[:witnessSize], cld.nbCld, cld.nbCldSlices)
	for i := 0; i < witnessSize; i++ {
		// i-th row of cldLenWit
		var cldLenRow []int
		for j := 0; j < cld.nbCld; j++ {
			cldLenRow = append(cldLenRow, int(cldLen[j][i].Uint64()))
		}

		// populate cldLenBinarys
		for j := 0; j < cld.nbCld; j++ {
			if cldLen[j][i].Uint64() != 0 {
				cldLenBinary[j][i] = field.One()
			}
		}

		// populate cldCol
		cldWit := decomposeByLength(cleanLimb[i], int(nByteWit[i].Uint64()), cldLenRow)

		for j := 0; j < cld.nbCld; j++ {
			cldCol[j][i] = cldWit[j]
		}
	}

	for j := 0; j < cld.nbCld; j++ {
		run.AssignColumn(cld.cld[j].GetColID(), smartvectors.RightZeroPadded(cldCol[j], maxRows))
		run.AssignColumn(cld.cldLen[j].GetColID(), smartvectors.RightZeroPadded(cldLen[j], maxRows))
		run.AssignColumn(cld.cldLenBinary[j].GetColID(), smartvectors.RightZeroPadded(cldLenBinary[j], maxRows))
	}

	// assign cldLenPowers
	for j := range cldLen {
		cldLenPowers := make([]field.Element, witnessSize)
		for i := range cldLen[0] {
			cldLen[j][i].BigInt(&a)
			cldLenPowers[i].Exp(field.NewElement(power8), &a)
		}
		run.AssignColumn(cld.cldLenPowers[j].GetColID(), smartvectors.RightPadded(cldLenPowers, field.One(), maxRows))
	}

	cld.assignCLDSlices(run, maxRows)
}
