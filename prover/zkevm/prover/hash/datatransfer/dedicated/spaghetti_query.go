package dedicated

import (
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/symbolic"
)

// It checks if  spaghettiOfMatrix is a filtered spaghetti of matrix
// by spaghetti form we mean ,
// passing through matrix row by row append the elements indicated by the filters to the spaghettiOfMatrix.
//
// The filter should have a special form over rows; starting with zero ending with ones.
// The query can handle multiple matrices over the same filter.
// Note : query does not check the form of the filter. Thus, it should have been checked before calling the query
func InsertIsSpaghetti(comp *wizard.CompiledIOP, round int, queryName ifaces.QueryID,
	matrix [][]ifaces.Column, filter []ifaces.Column, spaghettiOfMatrix []ifaces.Column, spaghettiSize int) {

	size := filter[0].Size()

	// tags for the filtered element of the matrix
	tags := make([]ifaces.Column, len(filter))

	// projection of the filter over the spaghettiOfMatrix
	projectedFilter := make([]ifaces.Column, len(filter))

	// Declare the new columns
	spaghettiOfTags := comp.InsertCommit(round,
		ifaces.ColIDf("%v_%v", queryName, "TagSpaghetti"), spaghettiSize)

	for j := range filter {
		projectedFilter[j] = comp.InsertCommit(round,
			ifaces.ColIDf("%v_%v_%v", queryName, "FilterSpaghetti", j), spaghettiSize)

		tags[j] = comp.InsertCommit(round, ifaces.ColIDf("%v_%v_%v", queryName, "Tags", j), size)
	}

	// Constraints over the tags; tag increases by one over the filtered elements
	for j := 1; j < len(filter); j++ {
		// tags[j]-tags[j-1] is 1 if filter[j-1]=1
		a := symbolic.Sub(tags[j], tags[j-1])
		comp.InsertGlobal(round, ifaces.QueryIDf("%v_%v-%v", queryName, 1, j),
			symbolic.Mul(symbolic.Mul(symbolic.Sub(1, a)),
				filter[j-1]))

		// We have to go to the previous row if filter[j] = 1 and filter[j-1]=0.
		// In this case,  tags[j]- shift(tags[len(matrix)-1],-1)  should be 1.
		b := symbolic.Sub(tags[j], column.Shift(tags[len(filter)-1], -1))
		expr2 := symbolic.Mul(symbolic.Sub(b, 1), symbolic.Mul(filter[j],
			symbolic.Mul(symbolic.Sub(1, filter[j-1]))))

		comp.InsertGlobal(round, ifaces.QueryIDf("%v_%v_%v", queryName, 2, j), expr2)
	}

	// Constraints over spaghettiTags; it increases by 1.
	// spaghettiTag - shift(spaghettiTag , -1) = 1
	dif := symbolic.Sub(spaghettiOfTags, column.Shift(spaghettiOfTags, -1))
	isActive := symbolic.NewConstant(0)
	for i := range filter {
		isActive = symbolic.Add(isActive, projectedFilter[i])
	}
	comp.InsertGlobal(round, ifaces.QueryIDf("%v_%v", queryName, "SpaghettiOfTag_Increases"),
		symbolic.Mul(symbolic.Sub(dif, 1), isActive))

	comp.SubProvers.AppendToInner(round,
		func(run *wizard.ProverRuntime) {
			assignSpaghetti(run, tags, filter,
				spaghettiOfTags, projectedFilter)
		},
	)

	matrices := make([][]ifaces.Column, len(matrix[0]))
	for i := range matrix[0] {
		for j := range matrix {
			matrices[i] = append(matrices[i], matrix[j][i])
		}
		matrices[i] = append(matrices[i], tags[i])
	}

	colB := append(spaghettiOfMatrix, spaghettiOfTags)

	// project matrix and tags over the spaghetti version
	for i := range matrix[0] {
		comp.InsertInclusionDoubleConditional(round, ifaces.QueryIDf("%v_LookUp_Matrix-In-Vector_%v", queryName, i),
			colB, matrices[i], projectedFilter[i], filter[i])

		comp.InsertInclusionDoubleConditional(round, ifaces.QueryIDf("%v_LookUp_Vector-In-Matrix_%v", queryName, i),
			matrices[i], colB, filter[i], projectedFilter[i])

	}

}

func assignSpaghetti(run *wizard.ProverRuntime, tags, filter []ifaces.Column,
	spaghettiOfTags ifaces.Column, spaghettiOfFilters []ifaces.Column) {

	witSize := smartvectors.Density(filter[0].GetColAssignment(run))

	tagsWit := make([][]field.Element, len(filter))
	filtersWit := make([][]field.Element, len(filter))

	// populate filter
	for i := range filter {
		filtersWit[i] = make([]field.Element, witSize)
		filtersWit[i] = filter[i].GetColAssignment(run).IntoRegVecSaveAlloc()[:witSize]
	}

	tag := uint64(1)
	// populate tags
	for i := range filtersWit[0] {
		for j := range filtersWit {
			if filtersWit[j][i].Uint64() != 0 {
				tagsWit[j] = append(tagsWit[j], field.NewElement(tag))
				tag++

			} else {
				tagsWit[j] = append(tagsWit[j], field.Zero())
			}
		}
	}
	for j := range tagsWit {
		run.AssignColumn(tags[j].GetColID(), smartvectors.RightZeroPadded(tagsWit[j], filter[0].Size()))
	}

	var spaghettiOfTagWit []field.Element
	spaghettiOfFiltersWit := make([][]field.Element, len(spaghettiOfFilters))

	// populate spaghetties
	for i := range filtersWit[0] {
		for j := range filtersWit {
			if filtersWit[j][i].Uint64() == 1 {
				spaghettiOfTagWit = append(spaghettiOfTagWit, tagsWit[j][i])
			}
			for k := range filtersWit {
				if k != j && filtersWit[k][i].Uint64() != 0 {
					spaghettiOfFiltersWit[j] = append(spaghettiOfFiltersWit[j], field.Zero())
				} else if k == j && filtersWit[j][i].Uint64() != 0 {
					spaghettiOfFiltersWit[j] = append(spaghettiOfFiltersWit[j], filtersWit[j][i])

				}

			}
		}
	}

	// assign the columns
	run.AssignColumn(spaghettiOfTags.GetColID(), smartvectors.RightZeroPadded(spaghettiOfTagWit, spaghettiOfTags.Size()))

	for j := range filter {
		run.AssignColumn(spaghettiOfFilters[j].GetColID(), smartvectors.RightZeroPadded(spaghettiOfFiltersWit[j], spaghettiOfTags.Size()))
	}
}
