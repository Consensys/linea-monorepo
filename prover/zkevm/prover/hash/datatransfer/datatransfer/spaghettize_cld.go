package datatransfer

import (
	"github.com/consensys/zkevm-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/zkevm-monorepo/prover/maths/field"
	"github.com/consensys/zkevm-monorepo/prover/protocol/column"
	"github.com/consensys/zkevm-monorepo/prover/protocol/ifaces"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
	"github.com/consensys/zkevm-monorepo/prover/symbolic"
	"github.com/consensys/zkevm-monorepo/prover/zkevm/prover/hash/datatransfer/dedicated"
)

// The spaghettizedCLD module implements the utility to spaghetize the CLD columns.
type spaghettizedCLD struct {
	// the size of spaghettized columns
	spaghettiSize int

	// the spaghetti versions of cld, cldLen,cldPower
	cldSpaghetti          ifaces.Column
	cldLenSpaghetti       ifaces.Column
	cldLenPowersSpaghetti ifaces.Column

	// It accumulate the length of all the slices pushed into the same lanes (supposed to be 8).
	accCldLenSpaghetti ifaces.Column

	// It indicate the first slice from the new hash over the spaghetti
	isFirstSliceOfNewHash ifaces.Column

	// The matrix form of isFirstSliceOfNewHash
	isNewHashMatrix []ifaces.Column

	// It is 1 when a lane in complete.
	isLaneComplete ifaces.Column

	// The spaghetti version of cldLenBinary, used to indicate the effective part of the spaghetties.
	isActive ifaces.Column
}

// NewSpaghetti declare the new columns specific to the module,
// and also the constraints asserting to the correct form of the spaghetti.
// It facilitate the task of pushing limbs to the lanes.
// It insures that 8 bytes are engaged to be pushed in each lane.
func (s *spaghettizedCLD) newSpaghetti(
	comp *wizard.CompiledIOP,
	round int,
	iPadd importAndPadd,
	cld cleanLimbDecomposition,
	maxNumRows int,
) {
	s.spaghettiSize = maxNumRows
	// Declare the columns
	s.insertCommit(comp, round, cld, maxNumRows)

	// Declare the constraints
	// Constraints over the spaghetti forms
	dedicated.InsertIsSpaghetti(comp, round, ifaces.QueryIDf("Spaghetti-CLD"),
		[][]ifaces.Column{cld.cld[:], cld.cldLen[:],
			cld.cldLenPowers[:], cld.cldLenBinary[:], s.isNewHashMatrix[:]},
		cld.cldLenBinary[:],
		[]ifaces.Column{s.cldSpaghetti, s.cldLenSpaghetti,
			s.cldLenPowersSpaghetti, s.isActive, s.isFirstSliceOfNewHash},
		maxNumRows,
	)

	// Constraint over isLaneComplete
	// The number of bytes to be pushed into the same lane is 8.
	// accCldLenSpaghetti = 8 iff isLaneComplete = 1
	dedicated.InsertIsTargetValue(comp, round, ifaces.QueryIDf("LanesAreComplete"),
		field.NewElement(numBytesInLane),
		s.accCldLenSpaghetti,
		s.isLaneComplete,
	)
	//  isLaneComplete is binary
	comp.InsertGlobal(round, ifaces.QueryIDf("IsLaneComplete_IsBinary"),
		symbolic.Mul(s.isLaneComplete, symbolic.Sub(1, s.isLaneComplete)))

	// Constraints over the accumulator of cldLenSpaghtti
	// accCldLenSpaghtti[0] = accCldLenSpaghtti[0]
	comp.InsertLocal(round, ifaces.QueryIDf("AccCLDLenSpaghetti_Loc"),
		symbolic.Sub(s.accCldLenSpaghetti, s.cldLenSpaghetti))

	// accCldLenSpaghtti[i] = accCldLenSpaghtti[i-1]*(1-isLaneComplete[i-1]) + cldLenSpaghtti[i]
	res := symbolic.Sub(1, column.Shift(s.isLaneComplete, -1)) // 1-isLaneComplete[i-1]
	expr := symbolic.Sub(symbolic.Add(symbolic.Mul(column.Shift(s.accCldLenSpaghetti, -1), res),
		s.cldLenSpaghetti), s.accCldLenSpaghetti)
	comp.InsertGlobal(round, ifaces.QueryIDf("AccCLDLenSpaghetti_Glob"),
		expr)

	// constraints over the form of isNewHashOverSpaghetti.
	// Considering cldSpaghetti, for
	// the first byte from the new hash, isNewHashOverSpaghetti = 1.
	// define matrices a , b  as follows where cld-matrix is from cld module
	// a[0] =cld[0]
	// a[j+1] = cld[j+1]-cld[j] for j=1,...,15
	// b[j] = a[j] * iPadd.isNewHash for any j.
	// Thus the constraint is equivalent with;
	// spaghetti(b) == isNewHashOverSpaghetti
	s.csIsNewHash(comp, round, iPadd, cld)
}

// InsertCommit declare the columns specific to the module.
func (s *spaghettizedCLD) insertCommit(comp *wizard.CompiledIOP, round int, cld cleanLimbDecomposition, maxNumRows int) {
	s.isNewHashMatrix = make([]ifaces.Column, len(cld.cld))
	// declare the columns
	s.cldSpaghetti = comp.InsertCommit(round, ifaces.ColIDf("CLD_Spaghetti"), maxNumRows)
	s.cldLenSpaghetti = comp.InsertCommit(round, ifaces.ColIDf("CLD_Len_Spaghetti"), maxNumRows)
	s.accCldLenSpaghetti = comp.InsertCommit(round, ifaces.ColIDf("AccCLDLen_Spaghetti"), maxNumRows)
	s.cldLenPowersSpaghetti = comp.InsertCommit(round, ifaces.ColIDf("CLD_LenPowers_Spaghetti"), maxNumRows)
	s.isLaneComplete = comp.InsertCommit(round, ifaces.ColIDf("IsLaneComplete"), maxNumRows)
	s.isActive = comp.InsertCommit(round, ifaces.ColIDf("IsActive_Spaghetti"), maxNumRows)
	s.isFirstSliceOfNewHash = comp.InsertCommit(round, ifaces.ColIDf("IsNewHash_Spaghetti"), maxNumRows)
	for j := range s.isNewHashMatrix {
		s.isNewHashMatrix[j] = comp.InsertCommit(round, ifaces.ColIDf("IsNewHash_Matrix_%v", j), cld.cld[0].Size())

	}
}

// Considering cldSpaghetti, for
// the first byte from the new hash, isFirstByteOfNewHash = 1.
func (s *spaghettizedCLD) csIsNewHash(comp *wizard.CompiledIOP,
	round int,
	iPadd importAndPadd,
	cld cleanLimbDecomposition,
) {
	// a[0] =cld[0]
	// a[j+1] = cldLenBinary[j+1]-cldLenBinary[j] for j=0,...,14
	// b[j] = a[j] * isNewHash for any j.
	// Thus the constraint is equivalent with;
	// spaghetti(b) == isNewHashOverSpaghetti
	var a, b [maxLanesFromLimb]*symbolic.Expression
	a[0] = ifaces.ColumnAsVariable(cld.cldLenBinary[0])
	for j := 0; j < maxLanesFromLimb-1; j++ {
		a[j+1] = symbolic.Sub(cld.cldLenBinary[j+1], cld.cldLenBinary[j])
	}
	for j := range cld.cld {
		b[j] = symbolic.Mul(a[j], iPadd.isNewHash)
	}
	// the matrix form of b
	for j := range cld.cld {
		comp.InsertGlobal(round, ifaces.QueryIDf("Matrix_IsNewHash_%v", j),
			symbolic.Sub(b[j], s.isNewHashMatrix[j]))
	}
	// note: by the way that b is built, we d'ont need to check isFirstByteOfNewHash is binary
}

// AssignSpaghetti assigns the columns specific to the module.
func (s *spaghettizedCLD) assignSpaghetti(
	run *wizard.ProverRuntime,
	iPadd importAndPadd,
	cld cleanLimbDecomposition,
	maxNumRows int) {
	// populate filter
	filter := cld.cldLenBinary
	witSize := smartvectors.Density(filter[0].GetColAssignment(run))

	// fetch the columns
	filterWit := make([][]field.Element, len(filter))
	cldLenWit := make([][]field.Element, len(filter))
	cldLenPowersWit := make([][]field.Element, len(filter))
	cldWit := make([][]field.Element, len(filter))
	cldLenBinaryWit := make([][]field.Element, len(filter))

	for i := range filter {
		filterWit[i] = make([]field.Element, witSize)
		cldWit[i] = cld.cld[i].GetColAssignment(run).IntoRegVecSaveAlloc()[:witSize]
		cldLenWit[i] = cld.cldLen[i].GetColAssignment(run).IntoRegVecSaveAlloc()[:witSize]
		cldLenBinaryWit[i] = cld.cldLenBinary[i].GetColAssignment(run).IntoRegVecSaveAlloc()[:witSize]
		cldLenPowersWit[i] = cld.cldLenPowers[i].GetColAssignment(run).IntoRegVecSaveAlloc()[:witSize]
		filterWit[i] = cldLenBinaryWit[i]
	}

	// make the spaghetti version of the fetched columns.
	spaghetti := makeSpaghetti(filterWit, cldWit, cldLenWit, cldLenBinaryWit, cldLenPowersWit)

	// assign the columns
	run.AssignColumn(s.cldSpaghetti.GetColID(), smartvectors.RightZeroPadded(spaghetti[0], maxNumRows))
	run.AssignColumn(s.cldLenSpaghetti.GetColID(), smartvectors.RightZeroPadded(spaghetti[1], maxNumRows))
	run.AssignColumn(s.isActive.GetColID(), smartvectors.RightZeroPadded(spaghetti[2], maxNumRows))
	run.AssignColumn(s.cldLenPowersSpaghetti.GetColID(), smartvectors.RightZeroPadded(spaghetti[3], maxNumRows))

	// populate isLaneComplete
	cldLenSpaghetti := spaghetti[1]
	isLaneComplete := AccReachedTargetValue(cldLenSpaghetti, numBytesInLane)

	// populate accumulator
	accCldLenWit := make([]field.Element, len(spaghetti[0]))
	if len(cldLenSpaghetti) != 0 {
		accCldLenWit[0] = cldLenSpaghetti[0]
	}
	var res field.Element
	one := field.One()
	for i := 1; i < len(spaghetti[0]); i++ {
		res.Sub(&one, &isLaneComplete[i-1])
		res.Mul(&accCldLenWit[i-1], &res)
		accCldLenWit[i].Add(&res, &cldLenSpaghetti[i])
	}

	// assign the accumulator of cldLen and isLaneComplete.
	run.AssignColumn(s.accCldLenSpaghetti.GetColID(), smartvectors.RightZeroPadded(accCldLenWit, maxNumRows))
	run.AssignColumn(s.isLaneComplete.GetColID(), smartvectors.RightZeroPadded(isLaneComplete, maxNumRows))

	// assign isNewHashMatrix and isFirstByteOfNewHash.
	s.assignIsNewHash(run, iPadd, cld, maxNumRows)
}

func (s *spaghettizedCLD) assignIsNewHash(
	run *wizard.ProverRuntime,
	iPadd importAndPadd,
	cld cleanLimbDecomposition,
	maxNumRows int) {
	witSize := smartvectors.Density(iPadd.isNewHash.GetColAssignment(run))
	isNewHash := iPadd.isNewHash.GetColAssignment(run).IntoRegVecSaveAlloc()[:witSize]
	nByte := iPadd.nByte.GetColAssignment(run).IntoRegVecSaveAlloc()[:witSize]

	sum := 0
	var t []int
	for j := range nByte {
		if isNewHash[j] == field.One() {
			// the length of the stream when we reach a newHash
			t = append(t, sum)
		}
		sum = sum + int(nByte[j].Uint64())
	}
	// isActive is equivalent with cldLenBinarySpaghetti
	witSizeSpaghetti := smartvectors.Density(s.isActive.GetColAssignment(run))
	cldLenSpaghetti := s.cldLenSpaghetti.GetColAssignment(run).IntoRegVecSaveAlloc()[:witSizeSpaghetti]

	ctr := 0
	sumS := 0
	var col []field.Element
	for j := range cldLenSpaghetti {
		if ctr < len(t) && t[ctr] == sumS {
			col = append(col, field.One())
			ctr++
		} else {
			col = append(col, field.Zero())
		}
		sumS = sumS + int(cldLenSpaghetti[j].Uint64())
	}

	//assign the columns
	run.AssignColumn(s.isFirstSliceOfNewHash.GetColID(), smartvectors.RightZeroPadded(col, maxNumRows))

	// populate the isNewHashMatrix
	cldLenBinary := make([][]field.Element, maxLanesFromLimb)
	for j := range cld.cld {
		cldLenBinary[j] = make([]field.Element, witSize)
		cldLenBinary[j] = cld.cldLenBinary[j].GetColAssignment(run).IntoRegVecSaveAlloc()[:witSize]
	}
	matrix := makeMatrix(cldLenBinary[:], col)

	for j := range matrix {
		run.AssignColumn(s.isNewHashMatrix[j].GetColID(), smartvectors.RightZeroPadded(matrix[j], cld.cld[0].Size()))
	}
}
