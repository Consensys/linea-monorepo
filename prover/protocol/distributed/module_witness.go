package experiment

import (
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/column/verifiercol"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
)

var (
	// moduleWitnessKey is the key used to store the witness of a module
	// in the [wizard.ProverRuntime.State]
	moduleWitnessKey = "MODULE_WITNESS"
)

// ModuleWitness is a structure collecting the witness of a module. And
// stores all the informations that are necessary to build the witness.
type ModuleWitness struct {
	// ModuleName indicates the name of the module
	ModuleName ModuleName
	// IsLPP indicates if the current instance of [ModuleWitness] is for
	// an LPP segment. In the contrary case, it is understood to be for
	// a GL segment.
	IsLPP bool
	// ModuleIndex indicates the vertical split of the current module
	ModuleIndex int
	// IsFirst, IsLast indicates if the module is the first or the last
	// segment of the module. When [ModuleIndex] == 0, [IsFirst] is true.
	IsFirst, IsLast bool
	// Columns maps the column id to their witness values
	Columns map[ifaces.ColID]smartvectors.SmartVector
	// ReceivedValuesGlobal stores the received values (for the global
	// constraints) of the current segment.
	ReceivedValuesGlobal []field.Element
	// N0 values are the parameters to the Horner queries in the same order
	// as in the [FilteredModuleInputs.HornerArgs]
	N0Values []int
	// InitialFiatShamirState is the initial FiatShamir state to set at
	// round 1.
	InitialFiatShamirState field.Element
}

// SegmentRuntime scans a [wizard.ProverRuntime] and returns a list of
// [ModuleWitness] that contains the witness for each segment of each
// module.
func SegmentRuntime(runtime *wizard.ProverRuntime, distributedWizard *DistributedWizard) (witnessesGL, witnessesLPP []*ModuleWitness) {

	for i := range distributedWizard.ModuleNames {
		wGL, wLPP := SegmentModule(runtime, distributedWizard.GLs[i], distributedWizard.LPPs[i])
		witnessesGL = append(witnessesGL, wGL...)
		witnessesLPP = append(witnessesLPP, wLPP...)
	}

	return witnessesGL, witnessesLPP
}

// SegmentModule produces the list of the [ModuleWitness] for a given module
func SegmentModule(runtime *wizard.ProverRuntime, moduleGL *ModuleGL, moduleLPP *ModuleLPP) (witnessesGL, witnessesLPP []*ModuleWitness) {

	var (
		fmi                  = moduleGL.definitionInput
		cols                 = runtime.Spec.Columns.AllKeys()
		nbSegmentModule      = NbSegmentOfModule(runtime, fmi.Disc, fmi.ModuleName)
		n0                   = make([]int, len(fmi.HornerArgs))
		receivedValuesGlobal = make([]field.Element, len(moduleGL.ReceivedValuesGlobalAccs))
	)

	witnessesLPP = make([]*ModuleWitness, nbSegmentModule)
	witnessesGL = make([]*ModuleWitness, nbSegmentModule)

	for moduleIndex := range witnessesLPP {

		moduleWitnessGL := &ModuleWitness{
			ModuleName:           fmi.ModuleName,
			ModuleIndex:          moduleIndex,
			IsFirst:              moduleIndex == 0,
			IsLast:               moduleIndex == nbSegmentModule-1,
			Columns:              make(map[ifaces.ColID]smartvectors.SmartVector),
			ReceivedValuesGlobal: receivedValuesGlobal,
		}

		moduleWitnessLPP := &ModuleWitness{
			ModuleName:  fmi.ModuleName,
			ModuleIndex: moduleIndex,
			IsFirst:     moduleIndex == 0,
			IsLast:      moduleIndex == nbSegmentModule-1,
			IsLPP:       true,
			Columns:     make(map[ifaces.ColID]smartvectors.SmartVector),
			N0Values:    n0,
		}

		for _, col := range cols {

			col := runtime.Spec.Columns.GetHandle(col)

			if ModuleOfColumn(fmi.Disc, col) != fmi.ModuleName {
				continue
			}

			segment := SegmentOfColumn(runtime, fmi.Disc, col, moduleIndex, nbSegmentModule)
			moduleWitnessGL.Columns[col.GetColID()] = segment

			if _, ok := fmi.ColumnsLPPSet[col.GetColID()]; ok {
				moduleWitnessLPP.Columns[col.GetColID()] = segment
			}
		}

		witnessesGL[moduleIndex] = moduleWitnessGL
		witnessesLPP[moduleIndex] = moduleWitnessLPP

		n0 = moduleWitnessLPP.NextN0s(moduleLPP)
		receivedValuesGlobal = moduleWitnessGL.NextReceivedValuesGlobal(moduleGL)
	}

	return witnessesGL, witnessesLPP
}

// NbSegmentOfModule returns the number of segments for a given module
func NbSegmentOfModule(runtime *wizard.ProverRuntime, disc ModuleDiscoverer, moduleName ModuleName) int {

	var (
		cols                  = runtime.Spec.Columns.AllKeys()
		nbSegmentModule       = -1
		colNamesWithOrientErr = []ifaces.ColID{}
	)

	for _, col := range cols {

		col := runtime.Spec.Columns.GetHandle(col)

		if ModuleOfColumn(disc, col) != moduleName {
			continue
		}

		var (
			newSize      = NewSizeOfColumn(disc, col)
			assignment   = col.GetColAssignment(runtime)
			density      = smartvectors.Density(assignment)
			_, orientErr = smartvectors.PaddingOrientationOf(assignment)
			nbSegmentCol = utils.DivCeil(density, newSize)
		)

		if orientErr != nil && density > 0 {
			// the column cannot be taken into account for the segmentation
			colNamesWithOrientErr = append(colNamesWithOrientErr, col.GetColID())
			continue
		}

		// We ignore the columns where the density if full because most of the time
		// these columns only exist by lack of optimization. (e.g.) use of a regular
		// smart-vector while a full vector could be used.
		if density < col.Size() {
			nbSegmentModule = max(nbSegmentModule, nbSegmentCol)
		}
	}

	if nbSegmentModule == -1 {
		utils.Panic("could not resolve the number of segment for module %v. columns with ambiguous orientation error: %v", moduleName, colNamesWithOrientErr)
	}

	return nbSegmentModule
}

// SegmentColumn returns the segment of a given column for given index. The
// function also takes a maxNbSegment value which is useful in case
func SegmentOfColumn(runtime *wizard.ProverRuntime, disc ModuleDiscoverer,
	col ifaces.Column, index, totalNbSegment int) smartvectors.SmartVector {

	var (
		newSize                 = NewSizeOfColumn(disc, col)
		assignment              = col.GetColAssignment(runtime)
		orientiation, orientErr = smartvectors.PaddingOrientationOf(assignment)
		start                   = index * newSize
		end                     = start + newSize
	)

	if orientErr != nil {
		// If a column is assigned to a plain-vector, then it is assumed to
		// be right-padded. The reason for this assumption is that the
		// columns from the arithmetization are systematically padded on the
		// left while the columns from the prover are all right-padded and the
		// sometime they (suboptimally) assigned to plain-vectors.
		orientiation = 1
	}

	if orientiation == -1 {
		start += assignment.Len() - totalNbSegment*newSize
		end += assignment.Len() - totalNbSegment*newSize
	}

	return assignment.SubVector(start, end)
}

// NextN0s returns the next value of N0, from the current one and the witness
// of the current module.
func (mw *ModuleWitness) NextN0s(moduleLPP *ModuleLPP) []int {

	newN0s := append([]int{}, mw.N0Values...)
	args := moduleLPP.Horner.Parts

	for i := range newN0s {

		// Note: the selector might be a non-natural column. Possibly a const-col.
		selCol := args[i].Selector

		if constCol, isConstCol := selCol.(verifiercol.ConstCol); isConstCol {

			if constCol.F.IsZero() {
				continue
			}

			if constCol.F.IsOne() {
				newN0s[i] += constCol.Size()
				continue
			}

			utils.Panic("the selector column has non-zero values: %v", constCol.F.String())
		}

		// Expectedly, at this point. The column must be a natural column. We can't support
		// shifted selector columns.
		_ = selCol.(column.Natural)

		selSV, ok := mw.Columns[selCol.GetColID()]
		if !ok {
			utils.Panic("selector: %v is missing from witness columns for module: %v index: %v", selCol, mw.ModuleName, mw.ModuleIndex)
		}

		sel := selSV.IntoRegVecSaveAlloc()

		for j := range sel {
			if sel[j].IsOne() {
				newN0s[i]++
			}
		}
	}

	return newN0s
}

// NextReceivedValuesGlobal returns the next value of ReceivedValuesGlobal, from
// the witness of the current module.
func (mw *ModuleWitness) NextReceivedValuesGlobal(moduleGL *ModuleGL) []field.Element {

	newReceivedValuesGlobal := make([]field.Element, len(mw.ReceivedValuesGlobal))

	for i, loc := range moduleGL.SentValuesGlobal {

		var (
			col      = column.RootParents(loc.Pol)
			pos      = column.StackOffsets(loc.Pol)
			colName  = col.GetColID()
			smartvec = mw.Columns[colName]
		)

		pos = utils.PositiveMod(pos, col.Size())
		newReceivedValuesGlobal[i] = smartvec.Get(pos)
	}

	return newReceivedValuesGlobal
}
