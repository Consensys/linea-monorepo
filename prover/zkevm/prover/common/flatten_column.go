package common

import (
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

const (
	// NbFlattenColLimbs defines the default number of columns allocated for storing the limbs.
	NbFlattenColLimbs = 8
)

// FlattenColumn flattens multiple limb columns and an accompanying mask into single columns,
// provides consistency checks via a precomputed projection mask.
type FlattenColumn struct {
	// Limbs is the row-wise concatenation of all limb columns.
	Limbs ifaces.Column
	// Mask is the row-wise concatenation of the original mask column.
	Mask ifaces.Column
	// auxProjectionMask selects flattenLimbs's positions to validate flattening consistency.
	auxProjectionMask ifaces.Column
	// originalLimbs holds the original limb columns to flatten.
	originalLimbs []ifaces.Column
	// originalMask holds the original mask column that selects elements for gnark circuit.
	originalMask ifaces.Column
	// onesColumn selects elements from the original limbs. This is always a column of 1s.
	onesColumn  ifaces.Column
	module      string
	circuit     string
	nbLimbsCols int
	// isDuplicated indicates if this FlattenColumn is already registered by other circuit,
	// so we don't need to commit to a new one.
	isDuplicated bool
}

// NewFlattenColumn initializes a FlattenColumn with:
//   - size: length of the original limbs columns
//   - nbLimbsCols: number of limb columns to flatten
//   - module: prefix for column identifiers
//
// It commits placeholders for flattened limbs and mask, and precomputes the projection mask.
func NewFlattenColumn(comp *wizard.CompiledIOP, size, nbLimbsCols int, module, circuit string) *FlattenColumn {
	flattenLimbsID := ifaces.ColIDf("%s.FLATTEN_LIMBS", module)
	auxProjectionMaskID := ifaces.ColIDf("%s.FLATTEN_PROJECTION_MASK", module)
	onesColumnID := ifaces.ColIDf("%s.FLATTEN_ORIG_LIMBS_MASK", module)

	flattenSize := size * nbLimbsCols

	// If the column already exists, we assume it is already registered by another circuit.
	var isDuplicated bool
	var flattenLimbs, auxProjectionMask, onesColumn ifaces.Column
	if comp.Columns.Exists(flattenLimbsID) {
		isDuplicated = true

		flattenLimbs = comp.Columns.GetHandle(flattenLimbsID)
		auxProjectionMask = comp.Columns.GetHandle(auxProjectionMaskID)
		onesColumn = comp.Columns.GetHandle(onesColumnID)
	} else {
		flattenLimbs = comp.InsertCommit(0, flattenLimbsID, flattenSize)
		auxProjectionMask = comp.InsertPrecomputed(auxProjectionMaskID,
			precomputeAuxProjectionMask(flattenSize, nbLimbsCols))
		onesColumn = comp.InsertPrecomputed(onesColumnID,
			precomputeAuxProjectionMask(size, 1))
	}

	return &FlattenColumn{
		Limbs:             flattenLimbs,
		Mask:              comp.InsertCommit(0, ifaces.ColIDf("%s.%s_FLATTEN_MASK", module, circuit), flattenSize),
		auxProjectionMask: auxProjectionMask,
		nbLimbsCols:       nbLimbsCols,
		onesColumn:        onesColumn,
		module:            module,
		circuit:           circuit,
		isDuplicated:      isDuplicated,
	}
}

// CsFlattenProjection adds a single batched projection constraint that enforces
// the “flattened” limbs and mask columns exactly match the row‐wise concatenation
// of the original limb columns and their mask.
//
// It works by:
//  1. Shifting the committed flattened limbs and mask by 0…nbLimbsCols−1 rows.
//  2. Duplicating the original limbs and mask into nbLimbsCols slots.
//  3. Using auxProjectionMask (a sparse selector with a 1 at each block start)
//     and a onesColumn to feed into the projection gadget.
//  4. Requiring at each selected position that
//     shiftedFlattenLimbs[i] == originalLimbs[i]
//     shiftedFlattenMask[i]  == originalMask
//     for all limb indices i and row positions.
//
// Suppose nbLimbsCols = 3, size = 4:
// original limbs (rows × limbs):
//
//	r\i   0    1    2
//	 0  [a0,  b0,  c0]
//	 1  [a1,  b1,  c1]
//	 2  [a2,  b2,  c2]
//	 3  [a3,  b3,  c3]
//
// flattened (size*3 = 12):
//
//	[a0, b0, c0,  a1, b1, c1,  a2, b2, c2,  a3, b3, c3]
//
// shift0 (i=0):
//
//	[a0, b0, c0,  a1, b1, c1,  a2, b2, c2,  a3, b3, c3]
//
// shift1 (i=1):
//
//	[b0, c0, a1,  b1, c1, a2,  b2, c2, a3,  b3, c3,  0 ]
//
// shift2 (i=2):
//
//	[c0, a1, b1,  c1, a2, b2,  c2, a3, b3,  c3,  0,   0 ]
//
// auxProjectionMask   = [1,0,0, 1,0,0, 1,0,0, 1,0,0]
//
// At each ‘1’ at idx = r*3, for shift i, enforce:
//
//	shift_i[idx] == original[r][i]
//
// CsFlattenProjection batches all these equalities into one projection check.
func (l *FlattenColumn) CsFlattenProjection(comp *wizard.CompiledIOP, limbs []ifaces.Column, mask ifaces.Column) {
	masks := make([]ifaces.Column, l.nbLimbsCols)
	shiftedFlattenLimbs := make([]ifaces.Column, l.nbLimbsCols)
	shiftedFlattenMask := make([]ifaces.Column, l.nbLimbsCols)

	for i := 0; i < l.nbLimbsCols; i++ {
		masks[i] = mask
		shiftedFlattenLimbs[i] = column.Shift(l.Limbs, i)
		shiftedFlattenMask[i] = column.Shift(l.Mask, i)
	}

	comp.InsertProjection(ifaces.QueryIDf("%v_%s_FLATTEN_PROJECTION", l.module, l.circuit),
		query.ProjectionInput{
			ColumnA: append(shiftedFlattenLimbs[:], shiftedFlattenMask[:]...),
			ColumnB: append(limbs[:], masks[:]...),
			FilterA: l.auxProjectionMask,
			FilterB: l.onesColumn,
		},
	)

	l.originalMask = mask
	l.originalLimbs = limbs
}

// Assign maps trace limb columns and mask into the flattened columns.
func (l *FlattenColumn) Assign(run *wizard.ProverRuntime) {
	l.assignMask(run)

	if !l.isDuplicated {
		l.assignLimbs(run)
	}
}

func (l *FlattenColumn) assignMask(run *wizard.ProverRuntime) {
	maskCol := l.originalMask.GetColAssignment(run).IntoRegVecSaveAlloc()

	flattenMask := NewVectorBuilder(l.Mask)
	for i := 0; i < l.originalMask.Size(); i++ {
		for j := 0; j < l.nbLimbsCols; j++ {
			flattenMask.PushField(maskCol[i])
		}
	}

	flattenMask.PadAndAssign(run, field.Zero())
}

func (l *FlattenColumn) assignLimbs(run *wizard.ProverRuntime) {
	limbsCols := make([][]field.Element, l.nbLimbsCols)
	for i, limb := range l.originalLimbs {
		limbsCols[i] = limb.GetColAssignment(run).IntoRegVecSaveAlloc()
	}

	flattenLimbs := NewVectorBuilder(l.Limbs)
	for i := 0; i < l.originalMask.Size(); i++ {
		for j := 0; j < l.nbLimbsCols; j++ {
			flattenLimbs.PushField(limbsCols[j][i])
		}
	}

	flattenLimbs.PadAndAssign(run, field.Zero())
}

// precomputeAuxProjectionMask creates a SmartVector with total size `size`,
// where `nbMasked` positions are periodically set to one.
func precomputeAuxProjectionMask(size, period int) smartvectors.SmartVector {
	resSlice := make([]field.Element, size)

	for i := 0; i < size; i += period {
		resSlice[i].SetOne()
	}

	return smartvectors.NewRegular(resSlice)
}
