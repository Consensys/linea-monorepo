package packing

import (
	"math/big"

	"github.com/consensys/zkevm-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/zkevm-monorepo/prover/maths/field"
	iszero "github.com/consensys/zkevm-monorepo/prover/protocol/dedicated"
	"github.com/consensys/zkevm-monorepo/prover/protocol/ifaces"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
	sym "github.com/consensys/zkevm-monorepo/prover/symbolic"
	"github.com/consensys/zkevm-monorepo/prover/utils"
	"github.com/consensys/zkevm-monorepo/prover/zkevm/prover/common"
	commonconstraints "github.com/consensys/zkevm-monorepo/prover/zkevm/prover/common/common_constraints"
	"github.com/consensys/zkevm-monorepo/prover/zkevm/prover/hash/generic"
	"github.com/consensys/zkevm-monorepo/prover/zkevm/prover/packing/dedicated"
)

// It stores the inputs for [newDecomposition] function.
type decompositionInputs struct {
	// parameters for decomposition,
	// the are used to determine the number slices and their lengths.
	param       generic.HashingUsecase
	cleaningCtx cleaningCtx
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
	// it checks that decomposedLimbs is of length decomposedLen.
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
	}
	decomposed.pa = dedicated.LengthConsistency(comp, lcInputs)

	return decomposed
}

// declare the native columns
func (decomposed *decomposition) insertCommit(comp *wizard.CompiledIOP) {

	createCol := common.CreateColFn(comp, DECOMPOSITION, decomposed.size)
	for x := 0; x < decomposed.nbSlices; x++ {
		decomposed.decomposedLimbs = append(decomposed.decomposedLimbs, createCol("Decomposed_Limbs", x))
		decomposed.decomposedLen = append(decomposed.decomposedLen, createCol("Decomposed_Len", x))
		decomposed.decomposedLenPowers = append(decomposed.decomposedLenPowers, createCol("Decomposed_Len_Powers", x))
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
			ifaces.QueryIDf("Decomposed_Len_Powers_%v", j), []ifaces.Column{lu.colNumber, lu.colPowers},
			[]ifaces.Column{decomposed.decomposedLen[j], decomposed.decomposedLenPowers[j]})
	}
	// \sum_i decomposedLen[i]=NByte
	comp.InsertGlobal(0, ifaces.QueryIDf("DecomposedLen_IsNByte"), sym.Sub(s, imported.NByte))

}

// decomposedLimbs cis the the decomposition of cleanLimbs
func (decomposed *decomposition) csDecomposition(
	comp *wizard.CompiledIOP, cleanLimbs ifaces.Column) {

	// recomposition of decomposedLimbs into cleanLimbs.
	cleanLimb := ifaces.ColumnAsVariable(decomposed.decomposedLimbs[0])
	for k := 1; k < decomposed.nbSlices; k++ {
		cleanLimb = sym.Add(sym.Mul(cleanLimb, decomposed.decomposedLenPowers[k]), decomposed.decomposedLimbs[k])
	}

	comp.InsertGlobal(0, ifaces.QueryIDf("Decompose_CleanLimbs"), sym.Sub(cleanLimb, cleanLimbs))
}

// /  Constraints over the form of filter and decomposedLen;
//   - filter = 1 iff decomposedLen != 0
func (decomposed decomposition) csFilter(comp *wizard.CompiledIOP) {
	// filtre = 1 iff decomposedLen !=0
	for j := 0; j < decomposed.nbSlices; j++ {
		// s.resIsZero = 1 iff decomposedLen = 0
		decomposed.resIsZero[j], decomposed.paIsZero[j] = iszero.IsZero(comp, decomposed.decomposedLen[j])
		// s.filter = (1 - s.resIsZero), this enforces filters to be binary.
		comp.InsertGlobal(0, ifaces.QueryIDf("%v_%v", "IS_NON_ZERO", j),
			sym.Sub(decomposed.filter[j],
				sym.Sub(1, decomposed.resIsZero[j])),
		)
	}

	// filter[0] = 1 over is Active.
	// this ensures that the first slice of the limb falls in the first column.
	comp.InsertGlobal(0, "FIRST_SLICE_IN_FIRST_COLUMN",
		sym.Sub(
			decomposed.filter[0], decomposed.isActive),
	)

}

// assign the columns specific to the module.
func (decomposed *decomposition) Assign(run *wizard.ProverRuntime) {
	decomposed.assignMainColumns(run)
	// assign s.filter
	var (
		filter = make([]*common.VectorBuilder, decomposed.nbSlices)
		one    = field.One()
	)
	for j := 0; j < decomposed.nbSlices; j++ {
		decomposed.paIsZero[j].Run(run)
		filter[j] = common.NewVectorBuilder(decomposed.filter[j])
		a := decomposed.resIsZero[j].GetColAssignment(run).IntoRegVecSaveAlloc()
		for i := range a {
			f := *new(field.Element).Sub(&one, &a[i])
			filter[j].PushField(f)
		}
		filter[j].PadAndAssign(run, field.Zero())
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
	}
	return decInp
}

// it assigns the main columns (not generated via an inner ProverAction)
func (decomposed *decomposition) assignMainColumns(run *wizard.ProverRuntime) {
	var (
		imported   = decomposed.Inputs.cleaningCtx.Inputs.imported
		cleanLimbs = decomposed.Inputs.cleaningCtx.CleanLimb.GetColAssignment(run).IntoRegVecSaveAlloc()
		nByte      = imported.NByte.GetColAssignment(run).IntoRegVecSaveAlloc()
		size       = decomposed.size

		// Assign the columns decomposedLimbs and decomposedLen
		decomposedLen   = make([][]field.Element, decomposed.nbSlices)
		decomposedLimbs = make([][]field.Element, decomposed.nbSlices)
	)

	for j := range decomposedLimbs {
		decomposedLimbs[j] = make([]field.Element, size)
		decomposedLen[j] = make([]field.Element, size)
	}

	// assign row-by-row
	decomposedLen = cutUpToMax(nByte, decomposed.nbSlices, decomposed.maxLen)
	for i := 0; i < size; i++ {
		// i-th row of DecomposedLen
		var lenRow []int
		for j := 0; j < decomposed.nbSlices; j++ {
			lenRow = append(lenRow, int(decomposedLen[j][i].Uint64()))
		}

		// populate DecomposedLimb
		decomposedLimb := decomposeByLength(cleanLimbs[i], int(nByte[i].Uint64()), lenRow)

		for j := 0; j < decomposed.nbSlices; j++ {
			decomposedLimbs[j][i] = decomposedLimb[j]
		}

	}

	for j := 0; j < decomposed.nbSlices; j++ {
		run.AssignColumn(decomposed.decomposedLimbs[j].GetColID(), smartvectors.RightZeroPadded(decomposedLimbs[j], decomposed.size))
		run.AssignColumn(decomposed.decomposedLen[j].GetColID(), smartvectors.RightZeroPadded(decomposedLen[j], decomposed.size))
	}

	// assign decomposedLenPowers
	var a big.Int
	for j := range decomposedLen {
		decomposedLenPowers := make([]field.Element, size)
		for i := range decomposedLen[0] {
			decomposedLen[j][i].BigInt(&a)
			decomposedLenPowers[i].Exp(field.NewElement(POWER8), &a)
		}
		run.AssignColumn(decomposed.decomposedLenPowers[j].GetColID(), smartvectors.RightPadded(decomposedLenPowers, field.One(), decomposed.size))
	}
}

// It receives an (set of) integer number N and outputs (n_1,...,n_nbChunk) such that
// N = \sum_{i < nbChunk} n_i
// n_i = min ( max, N - \sum_{j<i} n_j)
func cutUpToMax(nByte []field.Element, nbChunk, max int) (b [][]field.Element) {

	missing := uint64(max)
	b = make([][]field.Element, nbChunk)
	for i := range nByte {
		var a []field.Element
		curr := nByte[i].Uint64()
		for curr != 0 {
			if curr >= missing {
				a = append(a, field.NewElement(missing))
				curr = curr - missing
				missing = uint64(max)
			} else {
				a = append(a, field.NewElement(curr))
				missing = missing - curr
				curr = 0
			}
		}
		for len(a) < nbChunk {
			a = append(a, field.Zero())
		}
		s := 0
		for j := 0; j < nbChunk; j++ {
			s = s + int(a[j].Uint64())
			b[j] = append(b[j], a[j])
		}

		if s != int(nByte[i].Uint64()) {
			utils.Panic("decomposition of nByte is not correct; nByte %v, s %v", nByte[i].Uint64(), s)
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
		utils.Panic("input can not be decomposed to the given lengths")
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
