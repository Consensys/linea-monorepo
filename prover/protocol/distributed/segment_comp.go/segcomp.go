package segcomp

import (
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/distributed"
	discoverer "github.com/consensys/linea-monorepo/prover/protocol/distributed/namebaseddiscoverer"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
)

// SegmentInputs stores the inputs for both
// vertical and horizontal splitting of a [wizard.CompiledIOP] object.
type SegmentInputs struct {
	// InitialComp subject to the splitting
	InitialComp *wizard.CompiledIOP
	// inputs for horizontal splitting
	Disc discoverer.QueryBasedDiscoverer
	// module name relevant to the segment
	ModuleName distributed.ModuleName
	// inputs for vertical splitting
	NumSegmentsInModule int
	SegID               int
}

// GetFreshGLComp creates a coimpiledIOP relevant to the GL queries/sub-provers.
func GetFreshGLComp(in SegmentInputs) *wizard.CompiledIOP {

	// If it has no GL column return an empty compiledIOP
	if !in.Disc.GLColumns.Exists(in.ModuleName) {
		return wizard.NewCompiledIOP()
	}

	var (
		initialComp = in.InitialComp
		// initialize the compiledIOP.
		segComp = wizard.NewCompiledIOP()
		// extract glColumns of the module
		glCols = in.Disc.GLColumns.MustGet(in.ModuleName)
	)

	// get the segment ID via a ProverAction
	segComp.RegisterProverAction(0, segIDProvider{segID: in.SegID})

	// insert the split columns into the segment
	insertColumnIntoSegment(in, segComp, glCols)

	// register provider and receiver,  the placeholders for boundaries.
	provider := segComp.InsertCommit(0, "PROVIDER", getSizeForProviderReceiver(initialComp))
	receiver := segComp.InsertCommit(0, "RECEIVER", getSizeForProviderReceiver(initialComp))

	// create a new  sub-prover for the segment.
	glProver := glProver{
		// all the glColumns in the module
		segProver: segmentModuleProver{
			cols:        glCols,
			numSegments: in.NumSegmentsInModule,
		},
		// placeholders for boundaries
		provider: provider,
		receiver: receiver,
	}

	// register Prover action to assign columns
	segComp.RegisterProverAction(0, glProver)
	return segComp

}

// GetFreshLPPComp generates compiledIOP for compilation of LPP queries.
func GetFreshLPPComp(in SegmentInputs) *wizard.CompiledIOP {

	// If it has no LPP column return an empty compiledIOP
	if !in.Disc.LPPColumns.Exists(in.ModuleName) {
		return wizard.NewCompiledIOP()
	}

	var (
		// get LPP columns via Discoverer
		lppCols = in.Disc.LPPColumns.MustGet(in.ModuleName)
		// initialize CompiledIOP of the segment
		segComp = wizard.NewCompiledIOP()
	)

	// get the segmentID via a ProverAction
	segComp.RegisterProverAction(0, segIDProvider{})

	// insert the split columns into the segment
	insertColumnIntoSegment(in, segComp, lppCols)

	// create a new subProver for the segment.
	segProver := segmentModuleProver{
		cols:        lppCols,
		numSegments: in.NumSegmentsInModule,
	}

	// register Prover action for the segment to assign columns
	segComp.RegisterProverAction(0, segProver)

	return segComp
}

type segIDProvider struct {
	segID int
}

func (s segIDProvider) Run(run *wizard.ProverRuntime) {
	s.segID = run.ProverID
}

// it stores the input for the module prover
type segmentModuleProver struct {
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

		if distributed.IsVerifierColumn(col) {
			continue
		}

		status := run.Spec.Columns.Status(col.GetColID())
		// verifiercol and precomputed are already assigned, so assign the other columns.
		if status == column.Committed || status == column.Proof {
			// get the witness from the initialProver
			colWitness := run.ParentRuntime.GetColumn(col.GetColID())
			colSegWitness := getSegmentFromWitness(colWitness, p.numSegments, run.ProverID)
			// assign it in the module in the round 0
			run.AssignColumn(col.GetColID(), colSegWitness, 0)
		}
	}
}

func getSegmentFromWitness(wit ifaces.ColAssignment, numSegs, segID int) ifaces.ColAssignment {
	segSize := wit.Len() / numSegs
	return wit.SubVector(segSize*segID, segSize*segID+segSize)
}

type glProver struct {
	segProver segmentModuleProver
	provider  ifaces.Column
	receiver  ifaces.Column
}

func (p glProver) Run(run *wizard.ProverRuntime) {
	if run.ParentRuntime == nil {
		utils.Panic("invalid call: the runtime does not have a [ParentRuntime]")
	}
	if run.ProverID > p.segProver.numSegments {
		panic("proverID can not be larger than number of segments")
	}
	// assign the gl columns
	p.segProver.Run(run)
	// assign Provider and Receiver
	assignProvider(run, run.ProverID, p.segProver.numSegments, p.provider)
	//  for the current segment, the receiver is the provider of the previous segment.
	assignProvider(run, utils.PositiveMod(run.ProverID-1, p.segProver.numSegments),
		p.segProver.numSegments, p.receiver)

}

func getSizeForProviderReceiver(comp *wizard.CompiledIOP) int {

	numBoundaries := 0

	for _, queryID := range comp.QueriesNoParams.AllKeysAt(0) {

		if global, ok := comp.QueriesNoParams.Data(queryID).(query.GlobalConstraint); ok {

			var (
				board    = global.Board()
				metadata = board.ListVariableMetadata()
				maxShift = global.MinMaxOffset().Max
			)

			for _, m := range metadata {
				switch t := m.(type) {
				case ifaces.Column:

					if shifted, ok := t.(column.Shifted); ok {
						// number of boundaries from the current column
						numBoundaries += maxShift - column.StackOffsets(shifted)

					} else {
						numBoundaries += maxShift
					}

				}
			}
		}
	}
	return utils.NextPowerOfTwo(numBoundaries)
}

// assignProvider mainly assigns the provider
// it also can be used for the receiver assignment,
// since the receiver of segment i equal to the provider of segment (i-1).
func assignProvider(run *wizard.ProverRuntime, segID, numSegments int, col ifaces.Column) {

	var (
		parentRuntime = run.ParentRuntime
		initialComp   = parentRuntime.Spec
		allBoundaries = []field.Element{}
	)

	for _, q := range initialComp.QueriesNoParams.AllKeysAt(0) {
		if global, ok := initialComp.QueriesNoParams.Data(q).(query.GlobalConstraint); ok {

			var (
				board    = global.Board()
				metadata = board.ListVariableMetadata()
				maxShift = global.MinMaxOffset().Max
			)

			for _, m := range metadata {
				switch t := m.(type) {
				case ifaces.Column:

					var (
						segmentSize = t.Size() / numSegments
						lastRow     = (segID+1)*segmentSize - 1
						colWit      []field.Element
						// number of boundaries from the current column
						numBoundaries = 0
					)

					if shifted, ok := t.(column.Shifted); ok {
						numBoundaries = maxShift - column.StackOffsets(shifted)
						colWit = shifted.Parent.GetColAssignment(parentRuntime).IntoRegVecSaveAlloc()

					} else {
						numBoundaries = maxShift
						colWit = t.GetColAssignment(parentRuntime).IntoRegVecSaveAlloc()
					}

					for i := lastRow - numBoundaries + 1; i <= lastRow; i++ {

						allBoundaries = append(allBoundaries,
							colWit[i])

					}

				}

			}

		}
	}
	run.AssignColumn(col.GetColID(), smartvectors.RightZeroPadded(allBoundaries, col.Size()))
}

func insertColumnIntoSegment(in SegmentInputs, segComp *wizard.CompiledIOP, cols []ifaces.Column) {

	for _, col := range cols {

		if distributed.IsVerifierColumn(col) {
			continue
		}

		var (
			segSize = col.Size() / in.NumSegmentsInModule
			status  = in.InitialComp.Columns.Status(col.GetColID())
		)

		switch status {
		case column.Precomputed:

			precom := in.InitialComp.Precomputed.MustGet(col.GetColID())
			segComp.InsertPrecomputed(col.GetColID(),
				precom.SubVector(segSize*in.SegID, segSize*(in.SegID+1)))

		case column.VerifyingKey:

			precom := in.InitialComp.Precomputed.MustGet(col.GetColID())
			segComp.InsertPrecomputed(col.GetColID(),
				precom.SubVector(segSize*in.SegID, segSize*(in.SegID+1)))
			segComp.Columns.SetStatus(col.GetColID(), column.VerifyingKey)

		default:
			segComp.InsertColumn(0, col.GetColID(), segSize, status)

		}

	}
}
