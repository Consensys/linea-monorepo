package packing

import (
	"math/big"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	iszero "github.com/consensys/linea-monorepo/prover/protocol/dedicated"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	sym "github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/common"
	commonconstraints "github.com/consensys/linea-monorepo/prover/zkevm/prover/common/common_constraints"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/generic"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/packing/dedicated"
)

// It stores the inputs for [newDecomposition] function.
type decompositionInputs struct {
	// parameters for decomposition,
	// the are used to determine the number of slices and their lengths.
	param       generic.HashingUsecase
	cleaningCtx cleaningCtx
	Name        string
}

// Decomposition struct stores all the intermediate columns required to constraint correct decomposition.
// Decomposition module is responsible for decomposing the cleanLimbs into
// slices of different sizes.
// The max bytes pushed into a slice is
//
//	min (laneSizeBytes, remaining bytes from the limb)
type decomposition struct {
	Inputs *decompositionInputs
	// slices of cleanLimbs
	decomposedLimbs []ifaces.Column
	//  the length associated with decomposedLimbs
	decomposedLen []ifaces.Column
	// decomposedLenPowers = 2^(8*decomposedLen)
	decomposedLenPowers []ifaces.Column
	// prover action for lengthConsistency;
	// it checks that decomposedLimb is of length decomposedLen.
	pa wizard.ProverAction
	// it indicates the active part of the decomposition module
	isActive ifaces.Column
	// The filter is obtained from decomposedLen.
	// filter = 1 iff decomposedLen != 0.
	filter []ifaces.Column
	//  the result and  the ProverAction for IsZero().
	resIsZero []ifaces.Column
	paIsZero  []wizard.ProverAction
	// size of the module
	size int
	// number of slices from the decomposition;
	// it equals the number of  columns in decomposedLimbs.
	nbSlices int
	// max length in the decomposition
	maxLen int
}

/*
newDecomposition defines the columns and constraints asserting to the following facts:

 1. decomposedLimbs is the decomposition of cleanLimbs

 2. decomposedLimbs is of length decomposedLen
*/
func newDecomposition(comp *wizard.CompiledIOP, inp decompositionInputs) decomposition {

	var (
		size  = inp.cleaningCtx.CleanLimb.Size()
		nbCld = maxLanesFromLimbs(inp.param.LaneSizeBytes())
	)

	decomposed := decomposition{
		Inputs:   &inp,
		size:     size,
		nbSlices: nbCld,
		maxLen:   inp.param.LaneSizeBytes(),
		// the next assignment guarantees that isActive is from the Activation form.
		isActive: inp.cleaningCtx.Inputs.imported.IsActive,
	}

	// Declare the columns
	decomposed.insertCommit(comp)

	for j := 0; j < decomposed.nbSlices; j++ {
		// since they are later used for building the  decomposed.filter.
		commonconstraints.MustZeroWhenInactive(comp, decomposed.isActive, decomposed.decomposedLen[j])
		// this guarantees that filter and decompodedLimbs full fill the same constrains.
	}

	// Declare the constraints
	decomposed.csFilter(comp)
	decomposed.csDecomposLen(comp, inp.cleaningCtx.Inputs.imported)
	decomposed.csDecomposition(comp, inp.cleaningCtx.CleanLimb)

	// check the length consistency between decomposedLimbs and decomposedLen
	lcInputs := dedicated.LcInputs{
		Table:    decomposed.decomposedLimbs,
		TableLen: decomposed.decomposedLen,
		MaxLen:   inp.param.LaneSizeBytes(),
		Name:     inp.Name,
	}
	decomposed.pa = dedicated.LengthConsistency(comp, lcInputs)

	return decomposed
}

// declare the native columns
func (decomposed *decomposition) insertCommit(comp *wizard.CompiledIOP) {

	createCol := common.CreateColFn(comp, DECOMPOSITION+"_"+decomposed.Inputs.Name, decomposed.size)
	for x := 0; x < decomposed.nbSlices; x++ {
		decomposed.decomposedLimbs = append(decomposed.decomposedLimbs, createCol("Decomposed_Limbs", x))
		decomposed.decomposedLen = append(decomposed.decomposedLen, createCol("Decomposed_Len_%v", x))
		decomposed.decomposedLenPowers = append(decomposed.decomposedLenPowers, createCol("Decomposed_Len_Powers_%v", x))
	}

	decomposed.paIsZero = make([]wizard.ProverAction, decomposed.nbSlices)
	decomposed.resIsZero = make([]ifaces.Column, decomposed.nbSlices)
	decomposed.filter = make([]ifaces.Column, decomposed.nbSlices)
	for j := 0; j < decomposed.nbSlices; j++ {
		decomposed.filter[j] = createCol("Filter_%v", j)
	}

}

// /  Constraints over the form of decomposedLen and decomposedLenPowers;
//   - decomposedLen over a row adds up to NBytes
//   - decomposedLenPowers = 2^(8*decomposedLen)
func (decomposed *decomposition) csDecomposLen(
	comp *wizard.CompiledIOP,
	imported Importation,
) {

	lu := decomposed.Inputs.cleaningCtx.Inputs.lookup
	// The rows of decomposedLen adds up to NByte; \sum_i decomposedLen[i]=NByte
	s := sym.NewConstant(0)
	for j := range decomposed.decomposedLimbs {
		s = sym.Add(s, decomposed.decomposedLen[j])

		// Equivalence of "decomposedLenPowers" with "2^(decomposedLen * 8)"
		comp.InsertInclusion(0,
			ifaces.QueryIDf("%v_Decomposed_Len_Powers_%v", decomposed.Inputs.Name, j), []ifaces.Column{lu.colNumber, lu.colPowers},
			[]ifaces.Column{decomposed.decomposedLen[j], decomposed.decomposedLenPowers[j]})
	}
	// \sum_i decomposedLen[i]=NByte
	comp.InsertGlobal(0, ifaces.QueryIDf("%v_DecomposedLen_IsNByte", decomposed.Inputs.Name), sym.Sub(s, imported.NByte))

}

// decomposedLimbs is the the decomposition of cleanLimbs
func (decomposed *decomposition) csDecomposition(
	comp *wizard.CompiledIOP, cleanLimbs ifaces.Column) {

	// recomposition of decomposedLimbs into cleanLimbs.
	cleanLimb := ifaces.ColumnAsVariable(decomposed.decomposedLimbs[0])
	for k := 1; k < decomposed.nbSlices; k++ {
		cleanLimb = sym.Add(sym.Mul(cleanLimb, decomposed.decomposedLenPowers[k]), decomposed.decomposedLimbs[k])
	}

	comp.InsertGlobal(0, ifaces.QueryIDf("Decompose_CleanLimbs_%v", decomposed.Inputs.Name), sym.Sub(cleanLimb, cleanLimbs))
}

// /  Constraints over the form of filter and decomposedLen;
//   - filter = 1 iff decomposedLen != 0
func (decomposed decomposition) csFilter(comp *wizard.CompiledIOP) {
	// filtre = 1 iff decomposedLen !=0
	for j := 0; j < decomposed.nbSlices; j++ {
		// s.resIsZero = 1 iff decomposedLen = 0
		decomposed.resIsZero[j], decomposed.paIsZero[j] = iszero.IsZero(comp, decomposed.decomposedLen[j])
		// s.filter = (1 - s.resIsZero), this enforces filters to be binary.
		comp.InsertGlobal(0, ifaces.QueryIDf("%v_%v_%v", decomposed.Inputs.Name, "IS_NON_ZERO", j),
			sym.Sub(decomposed.filter[j],
				sym.Sub(1, decomposed.resIsZero[j])),
		)
	}

	// filter[0] = 1 over is Active.
	// this ensures that the first slice of the limb falls in the first column.
	comp.InsertGlobal(0, ifaces.QueryIDf("%v_FIRST_SLICE_IN_FIRST_COLUMN", decomposed.Inputs.Name),
		sym.Sub(
			decomposed.filter[0], decomposed.isActive),
	)

}

// assign the columns specific to the module.
func (decomposed *decomposition) Assign(run *wizard.ProverRuntime) {
	decomposed.assignMainColumns(run)

	// assign s.filter
	for j := 0; j < decomposed.nbSlices; j++ {
		decomposed.paIsZero[j].Run(run)

		var (
			filter        = decomposed.filter[j]
			compactFilter = make([]field.Element, 0)
			a             = decomposed.resIsZero[j].GetColAssignment(run)
			one           = field.One()
		)

		for ai := range a.IterateCompact() {
			var f field.Element
			f.Sub(&one, &ai)
			compactFilter = append(compactFilter, f)
		}

		run.AssignColumn(filter.GetColID(), smartvectors.FromCompactWithShape(a, compactFilter))
	}

	// assign Iszero()
	decomposed.pa.Run(run)
}

// get number of slices for the decomposition
func maxLanesFromLimbs(laneBytes int) int {
	if laneBytes > MAXNBYTE {
		return 2
	} else {
		return (MAXNBYTE / laneBytes) + 1
	}
}

// it builds the inputs for [newDecomposition]
func getDecompositionInputs(cleaning cleaningCtx, pckParam PackingInput) decompositionInputs {
	decInp := decompositionInputs{
		cleaningCtx: cleaning,
		param:       pckParam.PackingParam,
		Name:        pckParam.Name,
	}
	return decInp
}

// it assigns the main columns (not generated via an inner ProverAction)
func (decomposed *decomposition) assignMainColumns(run *wizard.ProverRuntime) {
	var (
		imported   = decomposed.Inputs.cleaningCtx.Inputs.imported
		cleanLimbs = decomposed.Inputs.cleaningCtx.CleanLimb.GetColAssignment(run)
		nByte      = imported.NByte.GetColAssignment(run)

		// Assign the columns decomposedLimbs and decomposedLen
		decomposedLen   [][]field.Element
		decomposedLimbs = make([][]field.Element, decomposed.nbSlices)

		// These are needed for sanity-checking the implementation which
		// crucially relies on the fact that the input vectors are post-padded.
		orientCleamLimbs, eOCL = smartvectors.PaddingOrientationOf(cleanLimbs)
		orientNByte, eONB      = smartvectors.PaddingOrientationOf(nByte)
	)

	if orientCleamLimbs <= 0 || orientNByte <= 0 || eOCL != nil || eONB != nil {
		panic("the implementation relies on the fact that the columns are post-padded, but they are pre-padded")
	}

	// assign row-by-row
	decomposedLen = cutUpToMax(nByte, decomposed.nbSlices, decomposed.maxLen)

	for j := range decomposedLimbs {
		decomposedLimbs[j] = make([]field.Element, len(decomposedLen[0]))
	}

	for i := 0; i < len(decomposedLen[0]); i++ {

		cleanLimb := cleanLimbs.Get(i)
		nByte := nByte.Get(i)

		// i-th row of DecomposedLen
		var lenRow []int
		for j := 0; j < decomposed.nbSlices; j++ {
			lenRow = append(lenRow, utils.ToInt(decomposedLen[j][i].Uint64()))
		}

		// populate DecomposedLimb
		decomposedLimb := decomposeByLength(cleanLimb, field.ToInt(&nByte), lenRow)

		for j := 0; j < decomposed.nbSlices; j++ {
			decomposedLimbs[j][i] = decomposedLimb[j]
		}
	}

	for j := 0; j < decomposed.nbSlices; j++ {
		run.AssignColumn(decomposed.decomposedLimbs[j].GetColID(), smartvectors.RightZeroPadded(decomposedLimbs[j], decomposed.size))
		run.AssignColumn(decomposed.decomposedLen[j].GetColID(), smartvectors.RightZeroPadded(decomposedLen[j], decomposed.size))
	}

	// powersOf256 stores the successive powers of 256. This is used to compute
	// the decomposedLenPowers.
	powersOf256 := make([]field.Element, decomposed.maxLen+1)
	for i := range powersOf256 {
		powersOf256[i].Exp(field.NewElement(256), big.NewInt(int64(i)))
	}

	// assign decomposedLenPowers
	for j := range decomposedLen {
		decomposedLenPowers := make([]field.Element, len(decomposedLen[j]))
		for i := range decomposedLen[0] {
			decomLen := field.ToInt(&decomposedLen[j][i])
			decomposedLenPowers[i] = powersOf256[decomLen]
		}
		run.AssignColumn(decomposed.decomposedLenPowers[j].GetColID(), smartvectors.RightPadded(decomposedLenPowers, field.One(), decomposed.size))
	}
}

// It receives an (set of) integer number N and outputs (n_1,...,n_nbChunk) such that
// N = \sum_{i < nbChunk} n_i
// n_i = min ( max, N - \sum_{j<i} n_j)
func cutUpToMax(nByte smartvectors.SmartVector, nbChunk, max int) [][]field.Element {

	var (
		missing = uint64(max)
		b       = make([][]field.Element, nbChunk)
	)

	for curr := range nByte.IterateSkipPadding() {

		// a is filled with nbChunk elements. Theses elements are the decomposition
		// of 'curr' into of integers less or equal to "max". For instance, curr = 10
		// and max=3 would give a and nChunk=5 would give a = [3, 3, 3, 1, 0].
		//
		// missing is a sort of carry that propagate the remainder of the decomposition
		// to the next line. With our example, the missing value would be 2 at the end
		// of the decomposition. And the next line would start with a 2 and then keep
		// going with 3s.
		var a = make([]uint64, 0, nbChunk)
		curr := curr.Uint64()
		for curr != 0 {
			if curr >= missing {
				a = append(a, missing)
				curr = curr - missing
				missing = uint64(max)
			} else {
				a = append(a, curr)
				missing = missing - curr
				curr = 0
			}
		}

		for len(a) < nbChunk {
			a = append(a, 0)
		}

		for j := 0; j < nbChunk; j++ {
			b[j] = append(b[j], field.NewElement(a[j]))
		}
	}

	return b
}

// It receives the length of the slices and decompose the element to the slices with the given lengths.
func decomposeByLength(cleanLimb field.Element, nBytes int, givenLen []int) (slices []field.Element) {

	//sanity check
	s := 0
	for i := range givenLen {
		s = s + givenLen[i]
	}
	if s != nBytes {
		utils.Panic("input can not be decomposed to the given lengths s=%v nBytes=%v", s, nBytes)
	}

	b := cleanLimb.Bytes()
	bytes := b[32-nBytes:]
	slices = make([]field.Element, len(givenLen))
	for i := range givenLen {
		if givenLen[i] == 0 {
			slices[i] = field.Zero()
		} else {
			b := bytes[:givenLen[i]]
			slices[i].SetBytes(b)
			bytes = bytes[givenLen[i]:]
		}
	}

	return slices

}
