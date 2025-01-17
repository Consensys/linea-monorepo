package distributed

import (
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
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
