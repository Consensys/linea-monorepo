/*
Descriptions:
context for Byte32cmp, checklimbs
verifiercol for fieldModulus
*/
package byte32cmp

import (
	"fmt"
	"math/big"

	"github.com/consensys/zkevm-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/zkevm-monorepo/prover/maths/field"
	"github.com/consensys/zkevm-monorepo/prover/protocol/column/verifiercol"
	"github.com/consensys/zkevm-monorepo/prover/protocol/dedicated/bigrange"
	"github.com/consensys/zkevm-monorepo/prover/protocol/ifaces"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
	"github.com/consensys/zkevm-monorepo/prover/symbolic"
	"github.com/consensys/zkevm-monorepo/prover/utils"
)

// byte32cmp context
type BytesCmpCtx struct {
	// The compiled IOP
	comp *wizard.CompiledIOP
	// Number of limbs needed to represent a Byte32 value
	numLimbs int
	// Number of bits per each limb
	bitPerLimbs int
	// Name if the name of parent context joined with a specifier.
	name string
	// Round of the module
	round int
	// size of ColumnA and ColumnB
	size int
	// ColumnA, allegedly the column with larger values
	columnA ifaces.Column
	// ColumnB, allegedly the column with smaller values
	columnB ifaces.Column
	// Column containing the limbs of ColumnA
	columnAlimbs []ifaces.Column
	// Column containing the limbs of ColumnB
	columnBlimbs []ifaces.Column
	// Column containing the limbs of the field Modulus
	modulusLimbs []ifaces.Column
	// Flags when comparing modulus with ColumnA
	gCmpModulusColA []ifaces.Column
	lCmpModulusColA []ifaces.Column
	eCmpModulusColA []ifaces.Column
	// Flags when comparing ColumnA with ColumnB
	gCmpColAColB []ifaces.Column
	lCmpColAColB []ifaces.Column
	eCmpColAColB []ifaces.Column
	// activeRow works as a filter
	activeRow *symbolic.Expression
}

// Defines all the columns of Byte32cmpCtx and registers all the required constraint
func (bcp *BytesCmpCtx) Define(comp *wizard.CompiledIOP, numLimbs, bitPerLimbs int, name string) {
	if bcp.columnA.Size() != bcp.columnB.Size() {
		utils.Panic("The size of columnA and columnB are different, %v vs %v", bcp.columnA.Size(), bcp.columnB.Size())
	}
	bcp.size = bcp.columnA.Size()
	bcp.comp = comp
	bcp.numLimbs = numLimbs
	bcp.bitPerLimbs = bitPerLimbs
	bcp.name = name
	// Specify the sizes of columns
	bcp.columnAlimbs = make([]ifaces.Column, numLimbs)
	bcp.columnBlimbs = make([]ifaces.Column, numLimbs)
	bcp.gCmpModulusColA = make([]ifaces.Column, numLimbs)
	bcp.lCmpModulusColA = make([]ifaces.Column, numLimbs)
	bcp.eCmpModulusColA = make([]ifaces.Column, numLimbs)
	bcp.gCmpColAColB = make([]ifaces.Column, numLimbs)
	bcp.lCmpColAColB = make([]ifaces.Column, numLimbs)
	bcp.eCmpColAColB = make([]ifaces.Column, numLimbs)
	bcp.modulusLimbs = make([]ifaces.Column, numLimbs)

	// Define the columns and insert range constraints
	bcp.defineColumns()
	// Define the flags
	bcp.defineFlags()
	// Queries for the comparison of modulus limbs with that of ColumnA
	// This proves that modulus is strictly greater than each element of ColumnA
	bcp.cmpLimbs(true)
	// Queries for the comparison of limbs of ColumnA and ColumnB
	// This proves that all the elements of ColumnA are strictly greater than that of ColumnB
	bcp.cmpLimbs(false)
}

// Compute Modulus Limbs
func (bcp *BytesCmpCtx) computeModulusLimbs() []field.Element {
	moduluslimbsWitness := make([]field.Element, bcp.numLimbs)
	for i := 0; i < bcp.numLimbs; i++ {
		l := uint64(0)
		for k := i * bcp.bitPerLimbs; k < (i+1)*bcp.bitPerLimbs; k++ {
			extractedBit := field.Modulus().Bit(k)
			l |= uint64(extractedBit) << (k % bcp.bitPerLimbs)
		}
		moduluslimbsWitness[i].SetUint64(l)
	}
	return moduluslimbsWitness
}

func (bcp *BytesCmpCtx) defineColumns() {
	for i := 0; i < bcp.numLimbs; i++ {
		// Declare the limbs for columnA
		bcp.columnAlimbs[i] = bcp.comp.InsertCommit(
			bcp.round,
			ifaces.ColIDf("BYTE32CMP_COLUMN_A_%v_LIMB_%v", bcp.name, i),
			bcp.size,
		)
		// Enforces the range over the limbs of columnA
		bcp.comp.InsertRange(
			bcp.round,
			ifaces.QueryIDf("BYTE32CMP_LIMB_RANGE_COLUMN_A_%v_LIMB_%v", bcp.name, i),
			bcp.columnAlimbs[i],
			1<<bcp.bitPerLimbs,
		)
		// Declare the limbs for ColumnB
		bcp.columnBlimbs[i] = bcp.comp.InsertCommit(
			bcp.round,
			ifaces.ColIDf("BYTE32CMP_COLUMN_B_%v_LIMB_%v", bcp.name, i),
			bcp.size,
		)
		// Enforces the range over the limbs of columnB
		bcp.comp.InsertRange(
			bcp.round,
			ifaces.QueryIDf("BYTE32CMP_LIMB_RANGE_COLUMN_B_%v_LIMB_%v", bcp.name, i),
			bcp.columnBlimbs[i],
			1<<bcp.bitPerLimbs,
		)
		// Accessing the Modulus limbs
		moduluslimbsWitness := bcp.computeModulusLimbs()
		// Assign the modulus limbs
		bcp.modulusLimbs[i] = verifiercol.NewConstantCol(moduluslimbsWitness[i], bcp.size)
	}

	// Build the linear combination with powers of 2^bitPerLimbs.
	// The limbs are in "little-endian" order. Namely, the first
	// limb encodes the least significant bits first.
	pow2 := symbolic.NewConstant(1 << bcp.bitPerLimbs)
	accA := ifaces.ColumnAsVariable(bcp.columnAlimbs[bcp.numLimbs-1])
	accB := ifaces.ColumnAsVariable(bcp.columnBlimbs[bcp.numLimbs-1])
	for i := bcp.numLimbs - 2; i >= 0; i-- {
		accA = symbolic.Mul(accA, pow2)
		accA = symbolic.Add(accA, bcp.columnAlimbs[i])
		accB = symbolic.Mul(accB, pow2)
		accB = symbolic.Add(accB, bcp.columnBlimbs[i])
	}

	// Declare the global constraint for columnA and columnB
	bcp.comp.InsertGlobal(bcp.round, ifaces.QueryIDf("GLOBAL_BYTE32CMP_ACCUMULATION_COLUMN_A_%v", bcp.name), symbolic.Sub(accA, bcp.columnA))
	bcp.comp.InsertGlobal(bcp.round, ifaces.QueryIDf("GLOBAL_BYTE32CMP_ACCUMULATION_COLUMN_B_%v", bcp.name), symbolic.Sub(accB, bcp.columnB))
}

func (bcp *BytesCmpCtx) defineFlags() {
	for i := 0; i < bcp.numLimbs; i++ {
		// Declare gCmpModulusColA
		bcp.gCmpModulusColA[i] = bcp.comp.InsertCommit(
			bcp.round,
			ifaces.ColIDf("BYTE32CMP_G_MOD_COL_A_%v_LIMB_%v", bcp.name, i),
			bcp.size,
		)

		// Declare lCmpModulusColA
		bcp.lCmpModulusColA[i] = bcp.comp.InsertCommit(
			bcp.round,
			ifaces.ColIDf("BYTE32CMP_L_MOD_COL_A_%v_LIMB_%v", bcp.name, i),
			bcp.size,
		)

		// Declare eCmpModulusColA
		bcp.eCmpModulusColA[i] = bcp.comp.InsertCommit(
			bcp.round,
			ifaces.ColIDf("BYTE32CMP_E_MOD_COL_A_%v_LIMB_%v", bcp.name, i),
			bcp.size,
		)

		// Declare gCmpColAColB
		bcp.gCmpColAColB[i] = bcp.comp.InsertCommit(
			bcp.round,
			ifaces.ColIDf("BYTE32CMP_G_COL_A_COL_B_%v_LIMB_%v", bcp.name, i),
			bcp.size,
		)

		// Declare lCmpColAColB
		bcp.lCmpColAColB[i] = bcp.comp.InsertCommit(
			bcp.round,
			ifaces.ColIDf("BYTE32CMP_L_COL_A_COL_B_%v_LIMB_%v", bcp.name, i),
			bcp.size,
		)

		// Declare eCmpColAColB
		bcp.eCmpColAColB[i] = bcp.comp.InsertCommit(
			bcp.round,
			ifaces.ColIDf("BYTE32CMP_COL_A_COL_B_%v_LIMB_%v", bcp.name, i),
			bcp.size,
		)
	}
}

func (bcp *BytesCmpCtx) cmpLimbs(isCmpWithModulus bool) {
	var (
		g         = make([]*symbolic.Expression, bcp.numLimbs)
		l         = make([]*symbolic.Expression, bcp.numLimbs)
		e         = make([]*symbolic.Expression, bcp.numLimbs)
		gMinusl   = make([]*symbolic.Expression, bcp.numLimbs)
		colALimbs = make([]*symbolic.Expression, bcp.numLimbs)
		colBLimbs = make([]*symbolic.Expression, bcp.numLimbs)
		name_     string
	)
	// If comparing with the modulus, then colA is modulus and colB is bcp.columnA
	// If not comparing with the modulus, then colA is bcp.columnA and colB is bcp.columnB
	if isCmpWithModulus {
		name_ = "MOD_COL_A"
		for i := 0; i < bcp.numLimbs; i++ {
			g[i] = ifaces.ColumnAsVariable(bcp.gCmpModulusColA[i])
			l[i] = ifaces.ColumnAsVariable(bcp.lCmpModulusColA[i])
			e[i] = ifaces.ColumnAsVariable(bcp.eCmpModulusColA[i])
			colALimbs[i] = ifaces.ColumnAsVariable(bcp.modulusLimbs[i])
			colBLimbs[i] = ifaces.ColumnAsVariable(bcp.columnAlimbs[i])
		}
	} else {
		name_ = "COL_A_COL_B"
		for i := 0; i < bcp.numLimbs; i++ {
			g[i] = ifaces.ColumnAsVariable(bcp.gCmpColAColB[i])
			l[i] = ifaces.ColumnAsVariable(bcp.lCmpColAColB[i])
			e[i] = ifaces.ColumnAsVariable(bcp.eCmpColAColB[i])
			colALimbs[i] = ifaces.ColumnAsVariable(bcp.columnAlimbs[i])
			colBLimbs[i] = ifaces.ColumnAsVariable(bcp.columnBlimbs[i])
		}
	}

	// declare gMinusL
	for i := 0; i < bcp.numLimbs; i++ {
		gMinusl[i] = symbolic.Sub(g[i], l[i])
	}

	// Declare the sequential limb check query (considering g, l, e follow little endian order)
	// If there are three limbs (numLimbs = 3) for each element of colA and colB, this will look like
	// expr = activeRow[i]*{1 - ((g1[i]-l1[i])+e1[i]*((g2[i]-l2[i])+e2[i]*(g3[i]-l3[i]-e3[i])))}
	// (here g1 is MSB and g3 is LSB)
	// To give an example: suppose MSB of colA is more than MSB of colB. Then we have g1 = 1, l1 = 0, e1 = 0
	// (the rest of the values are irrelevant). The expr boils down to zero as expected.
	acc := gMinusl[0]
	acc = symbolic.Sub(acc, e[0])
	for i := 1; i < bcp.numLimbs; i++ {
		acc = symbolic.Mul(acc, e[i])
		acc = symbolic.Add(acc, gMinusl[i])
	}
	// acc = 1 - acc
	acc = symbolic.Sub(symbolic.NewConstant(1), acc)

	// Filtering by the activeRows
	acc = symbolic.Mul(acc, bcp.activeRow)

	// Declare the global constraint
	bcp.comp.InsertGlobal(bcp.round, ifaces.QueryIDf("GLOBAL_BYTE32CMP_SEQUENTIAL_LIMB_CHECK_%v_%v", name_, bcp.name), acc)

	for i := 1; i < bcp.numLimbs; i++ {
		// Range query on (g[i](colALimbs[i]-colBLimbs[i]) + l[i](colBLimbs[i]-colALimbs[i]))
		summand1 := symbolic.Mul(g[i], symbolic.Sub(colALimbs[i], colBLimbs[i]))
		summand2 := symbolic.Mul(l[i], symbolic.Sub(colBLimbs[i], colALimbs[i]))
		expr1 := symbolic.Add(summand1, summand2)
		// Filtering by the activeRows
		expr1 = symbolic.Mul(expr1, bcp.activeRow)
		name2 := fmt.Sprintf("GLOBAL_BYTE32CMP_%v_BIGRANGE_%v_%v_", name_, bcp.name, i)
		// As we compare modulus limbs, we need total number of bits 256
		bigrange.BigRange(bcp.comp, expr1, 16, 16, name2)

		// Sanity of g, l, and e, when active they should sum up to 1 (law of tricotomy)
		expr2 := symbolic.Add(g[i], l[i], e[i])
		expr2 = symbolic.Sub(symbolic.NewConstant(1), expr2)
		expr2 = symbolic.Mul(expr2, bcp.activeRow)
		bcp.comp.InsertGlobal(bcp.round, ifaces.QueryIDf("GLOBAL_BYTE32CMP_GLE_SUM_%v_%v_%v", name_, bcp.name, i), expr2)
	}

}

// assigns the columns of Byte32CmpCtx
func (bcp *BytesCmpCtx) assign(
	run *wizard.ProverRuntime,
	colA, colB smartvectors.SmartVector,
) {
	var (
		colALimbs    = make([][]field.Element, bcp.numLimbs)
		colBLimbs    = make([][]field.Element, bcp.numLimbs)
		gCmpModColA  = make([][]field.Element, bcp.numLimbs)
		lCmpModColA  = make([][]field.Element, bcp.numLimbs)
		eCmpModColA  = make([][]field.Element, bcp.numLimbs)
		gCmpColAColB = make([][]field.Element, bcp.numLimbs)
		lCmpColAColB = make([][]field.Element, bcp.numLimbs)
		eCmpColAColB = make([][]field.Element, bcp.numLimbs)
	)
	// Accessing the Modulus limbs
	moduluslimbsWitness := bcp.computeModulusLimbs()

	// Assigning the size of the var columns
	for i := 0; i < bcp.numLimbs; i++ {
		colALimbs[i] = make([]field.Element, bcp.size)
		colBLimbs[i] = make([]field.Element, bcp.size)
		gCmpModColA[i] = make([]field.Element, bcp.size)
		lCmpModColA[i] = make([]field.Element, bcp.size)
		eCmpModColA[i] = make([]field.Element, bcp.size)
		gCmpColAColB[i] = make([]field.Element, bcp.size)
		lCmpColAColB[i] = make([]field.Element, bcp.size)
		eCmpColAColB[i] = make([]field.Element, bcp.size)
	}

	for j := 0; j < bcp.size; j++ {
		colAValFr := colA.Get(j)
		colBValFr := colB.Get(j)
		if colAValFr == field.Zero() && colBValFr == field.Zero() {
			continue
		}
		var colAVal, colBVal big.Int
		colAValFr.BigInt(&colAVal)
		colBValFr.BigInt(&colBVal)

		for i := 0; i < bcp.numLimbs; i++ {
			limbA := uint64(0)
			limbB := uint64(0)
			for k := i * bcp.bitPerLimbs; k < (i+1)*bcp.bitPerLimbs; k++ {
				extractedBitA := colAVal.Bit(k)
				extractedBitB := colBVal.Bit(k)
				limbA |= uint64(extractedBitA) << (k % bcp.bitPerLimbs)
				limbB |= uint64(extractedBitB) << (k % bcp.bitPerLimbs)
			}
			// To assign limbs
			colALimbs[i][j].SetUint64(limbA)
			colBLimbs[i][j].SetUint64(limbB)
			// To assign flags comparing Modulus and ColumnA
			switch {
			case limbA == moduluslimbsWitness[i].Uint64():
				gCmpModColA[i][j] = field.Zero()
				lCmpModColA[i][j] = field.Zero()
				eCmpModColA[i][j] = field.One()
			case moduluslimbsWitness[i].Uint64() > limbA:
				gCmpModColA[i][j] = field.One()
				lCmpModColA[i][j] = field.Zero()
				eCmpModColA[i][j] = field.Zero()
			case moduluslimbsWitness[i].Uint64() < limbA:
				gCmpModColA[i][j] = field.Zero()
				lCmpModColA[i][j] = field.One()
				eCmpModColA[i][j] = field.Zero()
			}
			// To assign flags comparing ColumnA and ColumnB
			switch {
			case limbA == limbB:
				gCmpColAColB[i][j] = field.Zero()
				lCmpColAColB[i][j] = field.Zero()
				eCmpColAColB[i][j] = field.One()
			case limbA > limbB:
				gCmpColAColB[i][j] = field.One()
				lCmpColAColB[i][j] = field.Zero()
				eCmpColAColB[i][j] = field.Zero()
			case limbA < limbB:
				gCmpColAColB[i][j] = field.Zero()
				lCmpColAColB[i][j] = field.One()
				eCmpColAColB[i][j] = field.Zero()
			}

		}

	}
	// assign the computed columns
	for i := 0; i < bcp.numLimbs; i++ {
		run.AssignColumn(bcp.columnAlimbs[i].GetColID(), smartvectors.NewRegular(colALimbs[i]))
		run.AssignColumn(bcp.columnBlimbs[i].GetColID(), smartvectors.NewRegular(colBLimbs[i]))
		run.AssignColumn(bcp.gCmpModulusColA[i].GetColID(), smartvectors.NewRegular(gCmpModColA[i]))
		run.AssignColumn(bcp.lCmpModulusColA[i].GetColID(), smartvectors.NewRegular(lCmpModColA[i]))
		run.AssignColumn(bcp.eCmpModulusColA[i].GetColID(), smartvectors.NewRegular(eCmpModColA[i]))
		run.AssignColumn(bcp.gCmpColAColB[i].GetColID(), smartvectors.NewRegular(gCmpColAColB[i]))
		run.AssignColumn(bcp.lCmpColAColB[i].GetColID(), smartvectors.NewRegular(lCmpColAColB[i]))
		run.AssignColumn(bcp.eCmpColAColB[i].GetColID(), smartvectors.NewRegular(eCmpColAColB[i]))

	}
}
