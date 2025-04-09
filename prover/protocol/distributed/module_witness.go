package distributed

import (
	"strings"

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
	// ModuleNames indicates the name of the module
	ModuleNames []ModuleName
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
func SegmentRuntime(runtime *wizard.ProverRuntime, distributedWizard *DistributedWizard) (witnessesGL []*ModuleWitnessGL, witnessesLPP []*ModuleWitnessLPP) {

	for i := range distributedWizard.GLs {
		wGL := SegmentModuleGL(runtime, distributedWizard.GLs[i])
		witnessesGL = append(witnessesGL, wGL...)
	}

	for i := range distributedWizard.LPPs {
		wLPP := SegmentModuleLPP(runtime, distributedWizard.LPPs[i])
		witnessesLPP = append(witnessesLPP, wLPP...)
	}

	return witnessesGL, witnessesLPP
}

// SegmentModule produces the list of the [ModuleWitness] for a given module
func SegmentModuleGL(runtime *wizard.ProverRuntime, moduleGL *ModuleGL) (witnessesGL []*ModuleWitnessGL) {

	var (
		fmi                  = moduleGL.definitionInput
		cols                 = runtime.Spec.Columns.AllKeys()
		nbSegmentModule      = NbSegmentOfModule(runtime, fmi.Disc, []ModuleName{fmi.ModuleName})
		receivedValuesGlobal = make([]field.Element, len(moduleGL.ReceivedValuesGlobalAccs))
	)

	witnessesGL = make([]*ModuleWitnessGL, nbSegmentModule)

	for moduleIndex := range witnessesGL {

		moduleWitnessGL := &ModuleWitnessGL{
			ModuleName:           fmi.ModuleName,
			ModuleIndex:          moduleIndex,
			IsFirst:              moduleIndex == 0,
			IsLast:               moduleIndex == nbSegmentModule-1,
			Columns:              make(map[ifaces.ColID]smartvectors.SmartVector),
			ReceivedValuesGlobal: receivedValuesGlobal,
		}

		for _, col := range cols {

			col := runtime.Spec.Columns.GetHandle(col)

			if ModuleOfColumn(fmi.Disc, col) != fmi.ModuleName {
				continue
			}

			segment := SegmentOfColumn(runtime, fmi.Disc, col, moduleIndex, nbSegmentModule)
			moduleWitnessGL.Columns[col.GetColID()] = segment
		}

		witnessesGL[moduleIndex] = moduleWitnessGL
		receivedValuesGlobal = moduleWitnessGL.NextReceivedValuesGlobal(moduleGL)
	}

	return witnessesGL
}

// SegmentModuleLPP produces the list of the [ModuleWitness] for a given module
func SegmentModuleLPP(runtime *wizard.ProverRuntime, moduleLPP *ModuleLPP) (witnessesLPP []*ModuleWitnessLPP) {

	var (
		fmis          = moduleLPP.definitionInputs
		cols          = runtime.Spec.Columns.AllKeys()
		_, _, hArgs   = getQueryArgs(fmis)
		n0            = make([]int, len(hArgs))
		moduleNames   = make([]ModuleName, 0, len(fmis))
		moduleNameSet = make(map[ModuleName]struct{})
		columnsLPPSet = make(map[ifaces.ColID]struct{})
	)

	for _, fmi := range moduleLPP.definitionInputs {
		moduleNames = append(moduleNames, fmi.ModuleName)
		moduleNameSet[fmi.ModuleName] = struct{}{}
		for col := range fmi.ColumnsLPPSet {
			columnsLPPSet[col] = struct{}{}
		}
	}

	var (
		nbSegmentModule = NbSegmentOfModule(runtime, moduleLPP.Disc, moduleNames)
		witnessesLPPs   = make([]*ModuleWitnessLPP, nbSegmentModule)
	)

	for moduleIndex := range witnessesLPPs {

		moduleWitnessLPP := &ModuleWitnessLPP{
			ModuleNames: moduleNames,
			ModuleIndex: moduleIndex,
			Columns:     make(map[ifaces.ColID]smartvectors.SmartVector),
			N0Values:    n0,
		}

		for _, col := range cols {

			if _, ok := columnsLPPSet[col]; !ok {
				continue
			}

			col := runtime.Spec.Columns.GetHandle(col)

			segment := SegmentOfColumn(runtime, moduleLPP.Disc, col, moduleIndex, nbSegmentModule)
			moduleWitnessLPP.Columns[col.GetColID()] = segment
		}

		witnessesLPPs[moduleIndex] = moduleWitnessLPP
		n0 = moduleWitnessLPP.NextN0s(moduleLPP)
	}

	return witnessesLPPs
}

// NbSegmentOfModule returns the number of segments for a given module
func NbSegmentOfModule(runtime *wizard.ProverRuntime, disc ModuleDiscoverer, moduleName []ModuleName) int {

	var (
		cols                  = runtime.Spec.Columns.AllKeys()
		nbSegmentModule       = -1
		colNamesWithOrientErr = []ifaces.ColID{}
		moduleSet             = map[ModuleName]struct{}{}
	)

	for _, mn := range moduleName {
		moduleSet[mn] = struct{}{}
	}

	for _, col := range cols {

		var (
			col = runtime.Spec.Columns.GetHandle(col)
			mn  = ModuleOfColumn(disc, col)
		)

		if _, ok := moduleSet[mn]; !ok {
			continue
		}

		var (
			newSize         = NewSizeOfColumn(disc, col)
			start, stop     = disc.SegmentBoundaryOf(runtime, col.(column.Natural))
			nbSegmentForCol = utils.DivExact(stop-start, newSize)
		)

		if nbSegmentForCol >= nbSegmentModule {
			nbSegmentModule = nbSegmentForCol
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

	if status := col.(column.Natural).Status(); status == column.Precomputed || status == column.VerifyingKey {
		return col.GetColAssignment(runtime)
	}

	var (
		newSize = NewSizeOfColumn(disc, col)
		// This returns the start and stop index of the segment for the current
		// standard module. But we might need to adjust for the LPP columns because
		// LPP modules group several standard modules together and they might have
		// different number of segments.
		startSeg, stopSeg = disc.SegmentBoundaryOf(runtime, col.(column.Natural))
	)

	if startSeg > 0 && (stopSeg-startSeg) < totalNbSegment*newSize {
		startSeg = stopSeg - totalNbSegment*newSize
	}

	var (
		start      = startSeg + index*newSize
		end        = start + newSize
		assignment = col.GetColAssignment(runtime)
	)

	isOOB := end > col.Size() || start < 0

	// This dirty hack is needed because sometime, the log-derivative-m columns are
	// paired with precomputed columns. It means their sizes are not affected by the
	// splitter and they will go OOB if they are used for more than 1 segment. When,
	// this happens we "pad" it on the fly with zeroes to signify that they corresponds
	// to unmatched lookup value.
	if isOOB && strings.HasSuffix(string(col.GetColID()), "_LOGDERIVATIVE_M") {
		return smartvectors.NewConstant(field.Zero(), newSize)
	}

	if isOOB {
		utils.Panic("going to overflow a column, name=%v length=%v start=%v stop=%v", col.GetColID(), col.Size(), start, end)
	}

	return assignment.SubVector(start, end)
}

// NextN0s returns the next value of N0, from the current one and the witness
// of the current module.
func (mw *ModuleWitnessLPP) NextN0s(moduleLPP *ModuleLPP) []int {

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
			utils.Panic("selector: %v is missing from witness columns for module: %v index: %v", selCol, mw.ModuleNames, mw.ModuleIndex)
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
func (mw *ModuleWitnessGL) NextReceivedValuesGlobal(moduleGL *ModuleGL) []field.Element {

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
