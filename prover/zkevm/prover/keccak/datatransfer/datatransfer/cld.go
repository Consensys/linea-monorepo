package datatransfer

import (
	"math/big"

	"github.com/consensys/zkevm-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/zkevm-monorepo/prover/maths/field"
	"github.com/consensys/zkevm-monorepo/prover/protocol/ifaces"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
	"github.com/consensys/zkevm-monorepo/prover/symbolic"
	"github.com/consensys/zkevm-monorepo/prover/zkevm/prover/keccak/datatransfer/dedicated"
)

// CleanLimbDecomposition (CLD) module is responsible for cleaning the limbs,
// and decomposing them to slices of 1 byte.
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
	cld [maxLanesFromLimb]ifaces.Column
	// It is the length of the slices from the decomposition
	cldLen [maxLanesFromLimb]ifaces.Column
	// It is the binary counterpart of cldLen.
	// Namely, it is zero iff cldLen is zero, otherwise it is one.
	cldLenBinary [maxLanesFromLimb]ifaces.Column
	// cldLenPowers = 2^(8*cldLen)
	cldLenPowers [maxLanesFromLimb]ifaces.Column

	// decomposition of cld to cldSlices (each slice is a single byte)
	cldSlices [maxLanesFromLimb][numBytesInLane]ifaces.Column
	// length of cldSlices
	lenCldSlices [maxLanesFromLimb][numBytesInLane]ifaces.Column
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
	maxRows int,
) {

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
func (cld *cleanLimbDecomposition) insertCommit(comp *wizard.CompiledIOP, round, maxRows int) {
	for x := 0; x < maxLanesFromLimb; x++ {
		cld.cld[x] = comp.InsertCommit(round, ifaces.ColIDf("CLD_%v", x), maxRows)
		cld.cldLen[x] = comp.InsertCommit(round, ifaces.ColIDf("CLD_Len_%v", x), maxRows)
		cld.cldLenBinary[x] = comp.InsertCommit(round, ifaces.ColIDf("CLD_Len_Binary_%v", x), maxRows)
		cld.cldLenPowers[x] = comp.InsertCommit(round, ifaces.ColIDf("CLD_Len_Powers_%v", x), maxRows)

		for k := 0; k < numBytesInLane; k++ {
			cld.cldSlices[x][k] = comp.InsertCommit(round, ifaces.ColIDf("CLD_Slice_%v_%v", x, k), maxRows)
			cld.lenCldSlices[x][k] = comp.InsertCommit(round, ifaces.ColIDf("Len_CLD_Slice_%v_%v", x, k), maxRows)
		}
	}
	cld.nbZeros = comp.InsertCommit(round, deriveName("NbZeros"), maxRows)
	cld.powersNbZeros = comp.InsertCommit(round, deriveName("PowersNbZeros"), maxRows)
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
	for j := 0; j < maxLanesFromLimb; j++ {
		s = symbolic.Add(s, cld.cldLen[j])

		// cldLenBinary is binary
		comp.InsertGlobal(round, ifaces.QueryIDf("cldLenBinary_IsBinary_%v", j),
			symbolic.Mul(cld.cldLenBinary[j], symbolic.Sub(1, cld.cldLenBinary[j])))

		// cldLenBinary = 1 iff cldLen !=0
		dedicated.InsertIsTargetValue(comp, round, ifaces.QueryIDf("IsOne_IFF_IsNnZero_%v", j),
			field.Zero(), cld.cldLen[j], symbolic.Sub(1, cld.cldLenBinary[j]))

		if j < maxLanesFromLimb-1 {
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
	for j := 0; j < maxLanesFromLimb; j++ {
		sum := symbolic.NewConstant(0)
		for k := 0; k < numBytesInLane; k++ {
			sum = symbolic.Add(sum, cld.lenCldSlices[j][k])

			// lenCldSlices is binary
			comp.InsertGlobal(round, ifaces.QueryIDf("lenCldSlices_IsBinary_%v_%v", j, k),
				symbolic.Mul(cld.lenCldSlices[j][k], symbolic.Sub(1, cld.lenCldSlices[j][k])))

			if k < maxLanesFromLimb-1 {
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

func (cld *cleanLimbDecomposition) assignCLDSlices(run *wizard.ProverRuntime, maxNumRows int) {

	witSize := smartvectors.Density(cld.cld[0].GetColAssignment(run))
	cldWit := make([][]field.Element, len(cld.cld))
	cldLenWit := make([][]field.Element, len(cld.cld))
	for j := range cld.cld {
		cldWit[j] = make([]field.Element, witSize)
		cldWit[j] = cld.cld[j].GetColAssignment(run).IntoRegVecSaveAlloc()[:witSize]
		cldLenWit[j] = cld.cldLen[j].GetColAssignment(run).IntoRegVecSaveAlloc()[:witSize]
	}

	// populate lenCldSlices
	for j := range cld.cld {
		lenCldSlices := make([][]field.Element, numBytesInLane)
		cldSlices := make([][]field.Element, numBytesInLane)

		for i := range cldWit[0] {
			dec := getZeroOnes(cldLenWit[j][i], numBytesInLane)
			a := decomposeByLengthFr(cldWit[j][i], int(cldLenWit[j][i].Uint64()), dec)
			for k := range cldSlices {
				lenCldSlices[k] = append(lenCldSlices[k], dec[k])
				cldSlices[k] = append(cldSlices[k], a[k])
			}
		}

		for k := 0; k < numBytesInLane; k++ {
			// fmt.Printf("cldSlices %v\n", vector.Prettify(cldSlices[k]))

			run.AssignColumn(cld.lenCldSlices[j][k].GetColID(),
				smartvectors.RightZeroPadded(lenCldSlices[k], maxNumRows))

			run.AssignColumn(cld.cldSlices[j][k].GetColID(),
				smartvectors.RightZeroPadded(cldSlices[k], maxNumRows))
		}
	}
}

// assignCLD assigns the columns specific to the CLD module (i.e., cld and cldLen, ...).
func (cld *cleanLimbDecomposition) assignCLD(run *wizard.ProverRuntime, iPadd importAndPadd, maxRows int) {
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
	var cldLen, cldCol, cldLenBinary [maxLanesFromLimb][]field.Element
	for j := range cldCol {
		cldCol[j] = make([]field.Element, witnessSize)
		cldLen[j] = make([]field.Element, witnessSize)
		cldLenBinary[j] = make([]field.Element, witnessSize)
	}
	// assign row-by-row
	cleanLimb := run.GetColumn(iPadd.cleanLimb.GetColID()).IntoRegVecSaveAlloc()[:witnessSize]
	nByteWit := nByte.IntoRegVecSaveAlloc()
	cldLen = cutAndStitchBy8(nByteWit[:witnessSize])
	for i := 0; i < witnessSize; i++ {
		// i-th row of cldLenWit
		cldLenRow := []int{int(cldLen[0][i].Uint64()),
			int(cldLen[1][i].Uint64()), int(cldLen[2][i].Uint64())}

		// populate cldLenBinarys
		for j := range cldLenRow {
			if cldLen[j][i].Uint64() != 0 {
				cldLenBinary[j][i] = field.One()
			}
		}

		// populate cldCol
		cldWit := decomposeByLength(cleanLimb[i], int(nByteWit[i].Uint64()), cldLenRow)
		cldCol[0][i] = cldWit[0]
		cldCol[1][i] = cldWit[1]
		cldCol[2][i] = cldWit[2]
	}

	for j := range cld.cld {
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
