package distributed

import (
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/column/verifiercol"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/sirupsen/logrus"
)

var (
	// moduleWitnessKey is the key used to store the witness of a module
	// in the [wizard.ProverRuntime.State]
	moduleWitnessKey = "MODULE_WITNESS"
)

// ModuleSegmentationBlueprint is a blueprint for the segmentation of a
// module. It contains the informations of ModuleGL or ModuleLPP that are
// relevant to performing the module segmentation. The raison d'être of this
// structure is to avoid having to deal with the the complete [ModuleGL] or
// [ModuleLPP] structure which are massive compared to what is actually required
// to perform the segmentation.
type ModuleSegmentationBlueprint struct {
	// ModuleName indicates the name of the module
	ModuleNames []ModuleName
	// ReceivedValuesGlobalRoots stores the list of the root column for the
	// [ModuleGL.ReceivedValuesGlobalAccs] for each received value.
	ReceivedValuesGlobalAccsRoots []ifaces.ColID
	// ReceivedValuesGlobalPosition stores the list of the position of the
	// [ModuleGL.ReceivedValuesGlobalAccs] for each received value.
	ReceivedValuesGlobalAccsPositions []int
	// NextN0SelectorRoots stores the list of the selector columns ID for the
	// Horner queries.
	NextN0SelectorRoots [][]ifaces.ColID
	// NextN0SelectorIsConst stores a list of boolean indicating if the the
	// selector is constant.
	NextN0SelectorIsConsts [][]bool
	// NextN0SelectorConsts stores the list of the constants for the Horner
	// queries.
	NextN0SelectorConsts [][]field.Element
	// NextN0SelectorConstSize lists the size of the constants column that are
	// used for the Horner queries.
	NextN0SelectorConstSizes [][]int
	// LPPColumnSets stores the list of the columns that are used for the
	// LPP segments.
	LPPColumnSets [][]ifaces.ColID
}

// ModuleWitnessGL is a structure collecting the witness of a module. And
// stores all the informations that are necessary to build the witness.
type ModuleWitnessGL struct {
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
}

// ModuleWitnessLPP is a structure collecting the witness of a module. The
// difference with a ModuleWitnessGL is that the witness is that the witness
// can be for a group of modules.
type ModuleWitnessLPP struct {
	// ModuleName indicates the name of the module
	ModuleName []ModuleName
	// ModuleIndex indicates the vertical split of the current module
	ModuleIndex int
	// InitialFiatShamirState is the initial FiatShamir state to set at
	// round 1.
	InitialFiatShamirState field.Element
	// N0 values are the parameters to the Horner queries in the same order
	// as in the [FilteredModuleInputs.HornerArgs]
	N0Values []int
	// Columns maps the column id to their witness values
	Columns map[ifaces.ColID]smartvectors.SmartVector
}

// SegmentRuntime scans a [wizard.ProverRuntime] and returns a list of
// [ModuleWitness] that contains the witness for each segment of each
// module.
func SegmentRuntime(
	runtime *wizard.ProverRuntime,
	disc ModuleDiscoverer,
	blueprintGLs, blueprintLPPs []ModuleSegmentationBlueprint,
) (
	witnessesGL []*ModuleWitnessGL,
	witnessesLPP []*ModuleWitnessLPP,
) {

	for i := range blueprintGLs {
		wGL := SegmentModuleGL(runtime, disc, &blueprintGLs[i])
		witnessesGL = append(witnessesGL, wGL...)
	}

	for i := range blueprintLPPs {
		wLPP := SegmentModuleLPP(runtime, disc, &blueprintLPPs[i])
		witnessesLPP = append(witnessesLPP, wLPP...)
	}

	return witnessesGL, witnessesLPP
}

// SegmentModule produces the list of the [ModuleWitness] for a given module
func SegmentModuleGL(runtime *wizard.ProverRuntime, disc ModuleDiscoverer, blueprintGL *ModuleSegmentationBlueprint) (witnessesGL []*ModuleWitnessGL) {

	var (
		moduleName           = blueprintGL.ModuleNames[0]
		cols                 = runtime.Spec.Columns.AllKeys()
		nbSegmentModule      = NbSegmentOfModule(runtime, disc, []ModuleName{moduleName})
		receivedValuesGlobal = make([]field.Element, len(blueprintGL.ReceivedValuesGlobalAccsRoots))
	)

	witnessesGL = make([]*ModuleWitnessGL, nbSegmentModule)

	for moduleIndex := range witnessesGL {

		moduleWitnessGL := &ModuleWitnessGL{
			ModuleName:           moduleName,
			ModuleIndex:          moduleIndex,
			IsFirst:              moduleIndex == 0,
			IsLast:               moduleIndex == nbSegmentModule-1,
			Columns:              make(map[ifaces.ColID]smartvectors.SmartVector),
			ReceivedValuesGlobal: receivedValuesGlobal,
		}

		for _, col := range cols {

			col := runtime.Spec.Columns.GetHandle(col)

			if ModuleOfColumn(disc, col) != moduleName {
				continue
			}

			segment := SegmentOfColumn(runtime, disc, col, moduleIndex, nbSegmentModule)
			moduleWitnessGL.Columns[col.GetColID()] = segment
		}

		witnessesGL[moduleIndex] = moduleWitnessGL
		receivedValuesGlobal = moduleWitnessGL.NextReceivedValuesGlobal(blueprintGL)
	}

	return witnessesGL
}

// SegmentModuleLPP produces the list of the [ModuleWitness] for a given module
func SegmentModuleLPP(runtime *wizard.ProverRuntime, disc ModuleDiscoverer, moduleLPP *ModuleSegmentationBlueprint) (witnessesLPP []*ModuleWitnessLPP) {

	var (
		cols          = runtime.Spec.Columns.AllKeys()
		n0            = make([]int, len(moduleLPP.NextN0SelectorRoots))
		moduleNameSet = make(map[ModuleName]struct{})
		columnsLPPSet = make(map[ifaces.ColID]struct{})
	)

	for moduleIndex := range moduleLPP.LPPColumnSets {
		moduleNameSet[moduleLPP.ModuleNames[moduleIndex]] = struct{}{}
		for _, col := range moduleLPP.LPPColumnSets[moduleIndex] {
			columnsLPPSet[col] = struct{}{}
		}
	}

	var (
		nbSegmentModule = NbSegmentOfModule(runtime, disc, moduleLPP.ModuleNames)
		witnessesLPPs   = make([]*ModuleWitnessLPP, nbSegmentModule)
	)

	for moduleIndex := range witnessesLPPs {

		moduleWitnessLPP := &ModuleWitnessLPP{
			ModuleName:  moduleLPP.ModuleNames,
			ModuleIndex: moduleIndex,
			Columns:     make(map[ifaces.ColID]smartvectors.SmartVector),
			N0Values:    n0,
		}

		for _, col := range cols {

			if _, ok := columnsLPPSet[col]; !ok {
				continue
			}

			col := runtime.Spec.Columns.GetHandle(col)
			segment := SegmentOfColumn(runtime, disc, col, moduleIndex, nbSegmentModule)
			moduleWitnessLPP.Columns[col.GetColID()] = segment
		}

		witnessesLPPs[moduleIndex] = moduleWitnessLPP
		n0 = moduleWitnessLPP.NextN0s(moduleLPP)
	}

	return witnessesLPPs
}

// NbSegmentOfModule returns the number of segments for a given module
func NbSegmentOfModule(runtime *wizard.ProverRuntime, disc ModuleDiscoverer, moduleName []ModuleName) (nbSegment int) {

	var (
		cols            = runtime.Spec.Columns.AllKeys()
		nbSegmentModule = -1
		moduleSet       = map[ModuleName]struct{}{}
		largeQbmSet     = map[ModuleName]struct{}{}
	)

	for _, mn := range moduleName {
		moduleSet[mn] = struct{}{}
	}

	for _, col := range cols {

		var (
			col = runtime.Spec.Columns.GetHandle(col)
			mn  = disc.ModuleOf(col.(column.Natural))
		)

		if len(mn) == 0 {
			disc := disc.(*StandardModuleDiscoverer)
			utils.Panic("one column does not belong to any module: %v, disc: %v, mn: %v", col.GetColID(), disc.ColumnsToModule, mn)
		}

		if _, ok := moduleSet[mn]; !ok {
			continue
		}

		var (
			newSize         = NewSizeOfColumn(disc, col)
			start, stop, _  = disc.SegmentBoundaryOf(runtime, col.(column.Natural))
			nbSegmentForCol = utils.DivExact(stop-start, newSize)
		)

		if nbSegmentForCol >= nbSegmentModule {

			// If the number of segment is large, we are very likely meeting a
			// bottlenecking QBM and it might be worth considering increasing
			// the new-size of that module.
			if nbSegmentForCol >= 2 {
				qbm, _ := disc.(*StandardModuleDiscoverer).QbmOf(col.(column.Natural))
				if _, ok := largeQbmSet[qbm.ModuleName]; !ok {
					largeQbmSet[qbm.ModuleName] = struct{}{}
					logrus.Warnf(
						"[large number of segment] column=%v newSize=%v start=%v stop=%v nbSegment=%v, qbm=%v, module=%v",
						col.GetColID(), newSize, start, stop, nbSegmentForCol, qbm.ModuleName, mn,
					)
				}
			}

			nbSegmentModule = nbSegmentForCol
		}
	}

	if nbSegmentModule == -1 {
		utils.Panic("could not resolve the number of segment for module %v", moduleName)
	}

	return nbSegmentModule
}

// SegmentColumn returns the segment of a given column for given index. The
// function also takes a maxNbSegment value which is useful in case
func SegmentOfColumn(runtime *wizard.ProverRuntime, disc ModuleDiscoverer, col ifaces.Column, index, totalNbSegment int) smartvectors.SmartVector {

	if status := col.(column.Natural).Status(); status == column.Precomputed || status == column.VerifyingKey {
		return col.GetColAssignment(runtime)
	}

	var (
		newSize = NewSizeOfColumn(disc, col)
		// This returns the start and stop index of the segment for the current
		// standard module. But we might need to adjust for the LPP columns because
		// LPP modules group several standard modules together and they might have
		// different number of segments.
		startSeg, stopSeg, paddingInfo = disc.SegmentBoundaryOf(runtime, col.(column.Natural))
	)

	if paddingInfo == leftPaddingInformation && (stopSeg-startSeg) < totalNbSegment*newSize {
		startSeg = stopSeg - totalNbSegment*newSize
	}

	var (
		start      = startSeg + index*newSize
		end        = start + newSize
		assignment = col.GetColAssignment(runtime)
	)

	// This switch case corresponds to a dirty-hack where the original column.
	// It is unexpected to have start < 0 and end > 0 due to the fact that the
	// columns and segment size are all power of two. And same observation for
	// stop > size and start < size.
	switch {

	// This is the regular case and there is no hack going on here.
	case start >= 0 && end <= col.Size():
		return assignment.SubVector(start, end)

	case start < 0 && end <= 0:
		// Otherwise, the padding technique is completely fine.
		if startSeg == 0 {
			logrus.Warnf("[ModuleWitnessOverflow] start and end are both negative, "+
				"name=%v length=%v start=%v stop=%v sub-module-segment=[%v - %v]. "+
				"Going to use the first value of the vector as a constant but this might fail.",
				col.GetColID(), col.Size(), start, end, startSeg, stopSeg)
		}
		// At this point, we are sure that the correct padding value is the
		// first value.
		padding := assignment.Get(0)
		return smartvectors.NewConstant(padding, newSize)

	case start >= col.Size() && end > col.Size():
		// Otherwise, the padding technique is completely fine.
		if stopSeg == col.Size() {
			logrus.Warnf("[ModuleWitnessOverflow] start and end are both greater than the length of the vector, "+
				"name=%v length=%v start=%v stop=%v sub-module-segment=[%v - %v]. "+
				"Going to use the last value of the vector as a constant but this might fail.",
				col.GetColID(), col.Size(), start, end, startSeg, stopSeg)
		}
		// At this point, we are sure that the correct padding value is the
		// last value (otherwise, we would have no way of guessing it).
		padding := assignment.Get(col.Size() - 1)
		return smartvectors.NewConstant(padding, newSize)

	default:
		utils.Panic(
			"unexpected case, col=%v, start=%v, end=%v, size=%v, startSeg=%v, stopSeg=%v",
			col, start, end, col.Size(), startSeg, stopSeg,
		)
		return nil
	}
}

// Blueprint returns the blueprint for the current module.
func (moduleGL *ModuleGL) Blueprint() ModuleSegmentationBlueprint {

	blueprintGL := ModuleSegmentationBlueprint{
		ModuleNames:                       []ModuleName{moduleGL.DefinitionInput.ModuleName},
		ReceivedValuesGlobalAccsRoots:     make([]ifaces.ColID, len(moduleGL.SentValuesGlobal)),
		ReceivedValuesGlobalAccsPositions: make([]int, len(moduleGL.SentValuesGlobal)),
	}

	for i, loc := range moduleGL.SentValuesGlobal {

		var (
			col     = column.RootParents(loc.Pol)
			pos     = column.StackOffsets(loc.Pol)
			colName = col.GetColID()
		)

		blueprintGL.ReceivedValuesGlobalAccsRoots[i] = colName
		blueprintGL.ReceivedValuesGlobalAccsPositions[i] = pos
	}

	return blueprintGL
}

// Blueprint returns the blueprint for the current module.
func (moduleLPP *ModuleLPP) Blueprint() ModuleSegmentationBlueprint {

	hornerParts := moduleLPP.Horner.Parts
	numHornerPart := len(moduleLPP.Horner.Parts)
	numSubmodule := len(moduleLPP.ModuleNames())

	res := ModuleSegmentationBlueprint{
		ModuleNames:              moduleLPP.ModuleNames(),
		NextN0SelectorRoots:      make([][]ifaces.ColID, numHornerPart),
		NextN0SelectorIsConsts:   make([][]bool, numHornerPart),
		NextN0SelectorConsts:     make([][]field.Element, numHornerPart),
		NextN0SelectorConstSizes: make([][]int, numHornerPart),
		LPPColumnSets:            make([][]ifaces.ColID, numSubmodule),
	}

	for i, di := range moduleLPP.DefinitionInputs {
		res.LPPColumnSets[i] = make([]ifaces.ColID, len(di.ColumnsLPP))
		for j := range di.ColumnsLPP {
			res.LPPColumnSets[i][j] = di.ColumnsLPP[j].GetColID()
		}
	}

	for i := range hornerParts {

		numParts := len(hornerParts[i].Selectors)
		res.NextN0SelectorConstSizes[i] = make([]int, numParts)
		res.NextN0SelectorRoots[i] = make([]ifaces.ColID, numParts)
		res.NextN0SelectorIsConsts[i] = make([]bool, numParts)
		res.NextN0SelectorConsts[i] = make([]field.Element, numParts)

		for k := range hornerParts[i].Selectors {

			// Note: the selector might be a non-natural column. Possibly a const-col.
			selCol := hornerParts[i].Selectors[k]
			res.NextN0SelectorRoots[i][k] = selCol.GetColID()

			if constCol, isConstCol := selCol.(verifiercol.ConstCol); isConstCol {

				if !constCol.F.IsZero() || constCol.F.IsOne() {
					utils.Panic("the selector column has non-binary values: %v", constCol.F.String())
				}

				res.NextN0SelectorConsts[i][k] = constCol.F
				res.NextN0SelectorIsConsts[i][k] = true
				res.NextN0SelectorConstSizes[i][k] = constCol.Size()

				continue
			}

			// Expectedly, at this point. The column must be a natural column. We can't support
			// shifted selector columns.
			_ = selCol.(column.Natural)
		}
	}

	return res
}

// NextN0s returns the next value of N0, from the current one and the witness
// of the current module.
func (mw *ModuleWitnessLPP) NextN0s(blueprintLPP *ModuleSegmentationBlueprint) []int {

	newN0s := append([]int{}, mw.N0Values...)

	for i := range blueprintLPP.NextN0SelectorRoots {
		for k := range blueprintLPP.NextN0SelectorRoots[i] {

			var (
				selColID        = blueprintLPP.NextN0SelectorRoots[i][k]
				selColIsConst   = blueprintLPP.NextN0SelectorIsConsts[i][k]
				selColConst     = blueprintLPP.NextN0SelectorConsts[i][k]
				selColConstSize = blueprintLPP.NextN0SelectorConstSizes[i][k]
			)

			if selColIsConst {

				if selColConst.IsZero() {
					continue
				}

				if selColConst.IsOne() {
					newN0s[i] += selColConstSize
					continue
				}

				utils.Panic("the selector column has non-zero values: %v", selColConst.String())
			}

			selSV, ok := mw.Columns[selColID]
			if !ok {
				utils.Panic("selector: %v is missing from witness columns for module: %v index: %v", selColID, mw.ModuleName, mw.ModuleIndex)
			}

			sel := selSV.IntoRegVecSaveAlloc()

			for j := range sel[k] {
				if sel[j].IsOne() {
					newN0s[i]++
				}
			}
		}
	}

	return newN0s
}

// NextReceivedValuesGlobal returns the next value of ReceivedValuesGlobal, from
// the witness of the current module.
func (mw *ModuleWitnessGL) NextReceivedValuesGlobal(blueprintGL *ModuleSegmentationBlueprint) []field.Element {

	newReceivedValuesGlobal := make([]field.Element, len(mw.ReceivedValuesGlobal))

	for i := range blueprintGL.ReceivedValuesGlobalAccsRoots {

		var (
			rootName = blueprintGL.ReceivedValuesGlobalAccsRoots[i]
			loc      = blueprintGL.ReceivedValuesGlobalAccsPositions[i]
			smartvec = mw.Columns[rootName]
		)

		loc = utils.PositiveMod(loc, smartvec.Len())
		newReceivedValuesGlobal[i] = smartvec.Get(loc)
	}

	return newReceivedValuesGlobal
}
