package base_conversion

import (
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	sym "github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/common"
)

// DecompositionInputs stors the inputs for [DecomposeBE]
type DecompositionInputs struct {
	name          string
	col           ifaces.Column
	numLimbs      int
	bytesPerLimbs int
}

// decompositionCtx stores the result of the decomposition  (i.e., Limbs).
// The Limbs are represented in Big-Endian order, but are not range-checked.
//
// Decomposition without range-check is only interesting for the case where,
//
//	an InclusionCheck is applied over the Limbs via lookup tables.
//
// Note that the endian is meant for the limbs and not for the the bytes here.
type decompositionCtx struct {
	Inputs DecompositionInputs
	// Limbs stores the  result of the decomposition.
	Limbs []ifaces.Column
}

// DecomposeBE receives a column and a base, and decompose the column in Limbs.
// The Limbs are not range checked, thus the result can be used only beside an Inclusion check.
func DecomposeBE(comp *wizard.CompiledIOP, inp DecompositionInputs) *decompositionCtx {

	var (
		size  = inp.col.Size()
		limbs = make([]ifaces.Column, inp.numLimbs)
	)

	for i := range limbs {
		// Declare the limbs for the number
		limbs[i] = comp.InsertCommit(
			0,
			ifaces.ColIDf("%v_%v_%v", inp.name, "LIMB", i),
			size,
		)
	}

	// Build the linear combination with powers of 2^bitPerLimbs. The limbs are
	// in "Big-endian" order. Namely, the first limb encodes the least
	// significant bits first.
	base := sym.NewConstant(1 << (inp.bytesPerLimbs * 8))
	// BaseRecomposeSliceHandles (de-)composes the slices in the given base and
	// returns the corresponding expression.
	res := sym.NewConstant(0)
	for k := range limbs {
		res = sym.Add(
			sym.Mul(base, res),
			limbs[k])
	}

	// Declare the global constraint
	comp.InsertGlobal(
		0,
		ifaces.QueryIDf("LIMB_RECOMPOSION"),
		sym.Sub(inp.col, res),
	)

	return &decompositionCtx{
		Inputs: inp,
		Limbs:  limbs,
	}

}

// Run implements the [wizard.ProverAction] interface
func (d *decompositionCtx) Run(run *wizard.ProverRuntime) {

	var (
		numLimbs      = d.Inputs.numLimbs
		bytesPerLimbs = d.Inputs.bytesPerLimbs
		totalNumBytes = numLimbs * bytesPerLimbs
		col           = d.Inputs.col.GetColAssignment(run).IntoRegVecSaveAlloc()
		limbs         = make([]*common.VectorBuilder, numLimbs)
		size          = d.Inputs.col.Size()
	)

	for j := range limbs {
		limbs[j] = common.NewVectorBuilder(d.Limbs[j])
	}

	for row := 0; row < size; row++ {
		b := col[row].Bytes()
		res := b[32-totalNumBytes:]
		for j := range limbs {
			a := res[j*bytesPerLimbs : j*bytesPerLimbs+bytesPerLimbs]
			limbs[j].PushField(*new(field.Element).SetBytes(a))
		}
	}

	// Then assigns the limbs
	for j := range limbs {
		limbs[j].PadAndAssign(run, field.Zero())
	}

}
