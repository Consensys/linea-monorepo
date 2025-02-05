package distributed

import (
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
)

// SegmentModuleInputs stores the inputs for both
// vertical and horizontal splitting of a [wizard.CompiledIOP] object.
type SegmentModuleInputs struct {
	// InitialComp subject to the splitting
	InitialComp *wizard.CompiledIOP
	// inputs for horizontal splitting
	Disc       ModuleDiscoverer
	ModuleName ModuleName
	// inputs for vertical splitting
	NumSegmentsInModule int
}

// GetFreshSegmentModuleComp returns a [wizard.DefineFunc] that creates
// a [wizard.CompiledIOP] object including only the columns relevant to the module.
// It splits the columns to the segments and assign them to the relevant CompiledIOP.
// It also contains the prover steps for assigning the module column.
// For all the segments from the same module, compiledIOP object is the same.
func GetFreshSegmentModuleComp(in SegmentModuleInputs) *wizard.CompiledIOP {

	var (
		// initialize the moduleComp
		segModComp  = wizard.NewCompiledIOP()
		initialComp = in.InitialComp
	)

	for round := 0; round < initialComp.NumRounds(); round++ {
		var columnsInRound []ifaces.Column
		// get the columns per round
		for _, colName := range initialComp.Columns.AllKeysAt(round) {

			col := initialComp.Columns.GetHandle(colName)
			if !in.Disc.ColumnIsInModule(col, in.ModuleName) {
				continue
			}

			segModComp.InsertCommit(col.Round(), col.GetColID(), col.Size()/in.NumSegmentsInModule)
			columnsInRound = append(columnsInRound, col)
		}

		// create a new  moduleProver
		segModuleProver := segmentModuleProver{
			cols:        columnsInRound,
			round:       round,
			numSegments: in.NumSegmentsInModule,
		}

		// register Prover action for the segment-module to assign columns per round
		segModComp.RegisterProverAction(round, segModuleProver)
	}

	return segModComp
}

// it stores the input for the module prover
type segmentModuleProver struct {
	round int
	// columns for a specific round
	cols        []ifaces.Column
	numSegments int
}

// It implements [wizard.ProverAction] for the module prover.
func (p segmentModuleProver) Run(run *wizard.ProverRuntime) {

	if run.ParentRuntime == nil {
		utils.Panic("invalid call: the runtime does not have a [ParentRuntime]")
	}
	if run.ProverID > p.numSegments {
		panic("proverID can not be larger than number of segments")
	}

	for _, col := range p.cols {
		// get the witness from the initialProver
		colWitness := run.ParentRuntime.GetColumn(col.GetColID())
		colSegWitness := getSegmentFromWitness(colWitness, p.numSegments, run.ProverID)
		// assign it in the module in the round col was declared
		run.AssignColumn(col.GetColID(), colSegWitness, col.Round())
	}
}

func getSegmentFromWitness(wit ifaces.ColAssignment, numSegs, segID int) ifaces.ColAssignment {
	segSize := wit.Len() / numSegs
	return wit.SubVector(segSize*segID, segSize*segID+segSize)
}

func GetFreshCompGL(in SegmentModuleInputs) *wizard.CompiledIOP {

	var (
		// initialize the segment CompiledIOP
		segComp           = wizard.NewCompiledIOP()
		initialComp       = in.InitialComp
		glColumns         = extractGLColumns(initialComp)
		glColumnsInModule = []ifaces.Column{}
	)

	// get the GL columns
	for _, colName := range initialComp.Columns.AllKeysAt(0) {

		col := initialComp.Columns.GetHandle(colName)
		if !in.Disc.ColumnIsInModule(col, in.ModuleName) {
			continue
		}

		if isGLColumn(col, glColumns) {
			// TBD: register at round 1, to create a separate commitment over GL.
			segComp.InsertCommit(0, col.GetColID(), col.Size()/in.NumSegmentsInModule)
			glColumnsInModule = append(glColumnsInModule, col)
		}

	}

	// register provider and receiver
	provider := segComp.InsertCommit(0, "PROVIDER", getSizeForProviderReceiver(initialComp))
	receiver := segComp.InsertCommit(0, "RECEIVER", getSizeForProviderReceiver(initialComp))

	// create a new  moduleProver
	glProver := glProver{
		glCols:      glColumnsInModule,
		numSegments: in.NumSegmentsInModule,
		provider:    provider,
		receiver:    receiver,
	}

	// register Prover action for the segment-module to assign columns per round
	segComp.RegisterProverAction(0, glProver)
	return segComp

}

type glProver struct {
	glCols      []ifaces.Column
	numSegments int
	provider    ifaces.Column
	receiver    ifaces.Column
}

func (p glProver) Run(run *wizard.ProverRuntime) {
	if run.ParentRuntime == nil {
		utils.Panic("invalid call: the runtime does not have a [ParentRuntime]")
	}
	if run.ProverID > p.numSegments {
		panic("proverID can not be larger than number of segments")
	}

	for _, col := range p.glCols {
		// get the witness from the initialProver
		colWitness := run.ParentRuntime.GetColumn(col.GetColID())
		colSegWitness := getSegmentFromWitness(colWitness, p.numSegments, run.ProverID)
		// assign it in the module in the round col was declared
		run.AssignColumn(col.GetColID(), colSegWitness, col.Round())
	}
	// assign Provider and Receiver
	assignProvider(run, run.ProverID, p.numSegments, p.provider)
	//  for the current segment, the receiver is the provider of the previous segment.
	assignProvider(run, utils.PositiveMod(run.ProverID-1, p.numSegments), p.numSegments, p.receiver)

}

func extractGLColumns(comp *wizard.CompiledIOP) []ifaces.Column {

	glColumns := []ifaces.Column{}
	// extract global queries
	for _, queryID := range comp.QueriesNoParams.AllKeysAt(0) {

		if glob, ok := comp.QueriesNoParams.Data(queryID).(query.GlobalConstraint); ok {
			glColumns = append(glColumns, ListColumnsFromExpr(glob.Expression)...)
		}

		if local, ok := comp.QueriesNoParams.Data(queryID).(query.LocalConstraint); ok {
			glColumns = append(glColumns, ListColumnsFromExpr(local.Expression)...)
		}
	}

	// extract localOpenings
	return glColumns
}

// ListColumnsFromExpr returns the natural version of all the columns in the expression.
func ListColumnsFromExpr(expr *symbolic.Expression) []ifaces.Column {

	var (
		board    = expr.Board()
		metadata = board.ListVariableMetadata()
		colList  = []ifaces.Column{}
	)

	for _, m := range metadata {
		switch t := m.(type) {
		case ifaces.Column:

			if shifted, ok := t.(column.Shifted); ok {
				colList = append(colList, shifted.Parent)
			} else {
				colList = append(colList, t)
			}

		}
	}
	return colList

}

func isGLColumn(col ifaces.Column, glColumns []ifaces.Column) bool {

	for _, glCol := range glColumns {
		if col.GetColID() == glCol.GetColID() {
			return true
		}
	}
	return false

}

func getSizeForProviderReceiver(comp *wizard.CompiledIOP) int {

	numBoundaries := 0

	for _, queryID := range comp.QueriesNoParams.AllKeysAt(0) {

		if global, ok := comp.QueriesNoParams.Data(queryID).(query.GlobalConstraint); ok {

			var (
				board    = global.Board()
				metadata = board.ListVariableMetadata()
				maxShift = GetMaxShift(global.Expression)
			)

			for _, m := range metadata {
				switch t := m.(type) {
				case ifaces.Column:

					if shifted, ok := t.(column.Shifted); ok {
						// number of boundaries from the current column
						numBoundaries += maxShift - shifted.Offset

					} else {
						numBoundaries += maxShift
					}

				}
			}
		}
	}
	return utils.NextPowerOfTwo(numBoundaries)
}

func GetMaxShift(expr *symbolic.Expression) int {
	var (
		board    = expr.Board()
		metadata = board.ListVariableMetadata()
		maxshift = 0
	)
	for _, m := range metadata {
		switch t := m.(type) {
		case ifaces.Column:
			if shifted, ok := t.(column.Shifted); ok {
				maxshift = max(maxshift, shifted.Offset)
			}
		}
	}
	return maxshift
}

// assignProvider mainly assigns the provider
// it also can be used for the receiver assignment,
// since the receiver of segment i equal to the provider of segment (i-1).
func assignProvider(run *wizard.ProverRuntime, segID, numSegments int, col ifaces.Column) {

	var (
		segComp       = run.Spec
		parentRuntime = run.ParentRuntime
		initialComp   = parentRuntime.Spec
		allBoundaries = []field.Element{}
	)

	for _, q := range segComp.QueriesNoParams.AllKeysAt(0) {
		if global, ok := initialComp.QueriesNoParams.Data(q).(query.GlobalConstraint); ok {

			var (
				board    = global.Board()
				metadata = board.ListVariableMetadata()
				maxShift = GetMaxShift(global.Expression)
			)

			for _, m := range metadata {
				switch t := m.(type) {
				case ifaces.Column:

					var (
						segmentSize = t.Size() / numSegments
						lastRow     = (segID+1)*segmentSize - 1
						colWitness  = t.GetColAssignment(parentRuntime).IntoRegVecSaveAlloc()
						// number of boundaries from the current column
						numBoundaries = 0
					)

					if shifted, ok := t.(column.Shifted); ok {
						numBoundaries = maxShift - shifted.Offset

					} else {
						numBoundaries = maxShift
					}

					for i := lastRow - numBoundaries + 1; i <= lastRow; i++ {

						allBoundaries = append(allBoundaries,
							colWitness[i])

					}

				}

			}

		}
	}
	run.AssignColumn(col.GetColID(), smartvectors.RightZeroPadded(allBoundaries, col.Size()))
}
