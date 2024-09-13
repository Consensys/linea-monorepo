package mimc

import (
	"github.com/consensys/linea-monorepo/prover/crypto/mimc"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/sirupsen/logrus"
)

// Compiles the MiMC queries by instantiating a MiMC module
func CompileMiMC(comp *wizard.CompiledIOP) {

	// Scans the compiled IOP, looking for unignored MiMC queries.
	// And mark them as ignored when encountered.
	totalLen := 0
	round := 0
	mimcQueries := []query.MiMC{}

	for _, id := range comp.QueriesNoParams.AllUnignoredKeys() {

		// Fetch the query
		q := comp.QueriesNoParams.Data(id)
		qMiMC, ok := q.(query.MiMC)
		if !ok {
			// not a MiMC query, skip it
			continue
		}

		// else mark it as ignored
		comp.QueriesNoParams.MarkAsIgnored(id)

		mimcQueries = append(mimcQueries, qMiMC)
		totalLen += qMiMC.Blocks.Size()
		round = utils.Max(round, comp.QueriesNoParams.Round(id))
	}

	if len(mimcQueries) == 0 {
		// nothing to compile :
		logrus.Debug("MiMC compiler exited : no MiMC queries to compile")
		return
	}

	if len(mimcQueries) == 1 {
		// unroll the mimc check directly over the columns of the unique query
		logrus.Debug("MiMC compiler : only one MiMC query to compile, no lookup needed")
		manualCheckMiMCBlock(comp, mimcQueries[0].Blocks, mimcQueries[0].OldState, mimcQueries[0].NewState)
	}

	// Else, we conflate every query in a single module and we apply the MiMC check over it.
	totalLen = utils.NextPowerOfTwo(totalLen)

	blocks := comp.InsertCommit(round, ifaces.ColID(mimcName(comp, "ALL_BLOCKS")), totalLen)
	oldStates := comp.InsertCommit(round, ifaces.ColID(mimcName(comp, "ALL_OLD_STATES")), totalLen)
	newStates := comp.InsertCommit(round, ifaces.ColID(mimcName(comp, "ALL_NEW_STATES")), totalLen)

	// Assign these columns
	comp.SubProvers.AppendToInner(round, func(run *wizard.ProverRuntime) {

		// Preallocate all the slices
		blocksWit := make([]field.Element, 0, totalLen)
		oldStatesWit := make([]field.Element, 0, totalLen)
		newStatesWit := make([]field.Element, 0, totalLen)

		// Append all the blocks, old states and new states to the slices for all queries
		for _, q := range mimcQueries {
			blocksWit = append(blocksWit,
				smartvectors.IntoRegVec(q.Blocks.GetColAssignment(run))...,
			)
			oldStatesWit = append(oldStatesWit,
				smartvectors.IntoRegVec(q.OldState.GetColAssignment(run))...,
			)
			newStatesWit = append(newStatesWit,
				smartvectors.IntoRegVec(q.NewState.GetColAssignment(run))...,
			)
		}

		if len(blocksWit) == totalLen {
			// Just allocate the slices as is and return
			run.AssignColumn(blocks.GetColID(), smartvectors.NewRegular(blocksWit))
			run.AssignColumn(oldStates.GetColID(), smartvectors.NewRegular(oldStatesWit))
			run.AssignColumn(newStates.GetColID(), smartvectors.NewRegular(newStatesWit))
			return
		}

		// Else, we need to pad with a dummy value
		dumBlock, dumOld, dumNew := field.Zero(), field.Zero(), mimc.BlockCompression(field.Zero(), field.Zero())

		run.AssignColumn(blocks.GetColID(),
			smartvectors.RightPadded(blocksWit, dumBlock, totalLen),
		)
		run.AssignColumn(oldStates.GetColID(),
			smartvectors.RightPadded(oldStatesWit, dumOld, totalLen),
		)
		run.AssignColumn(newStates.GetColID(),
			smartvectors.RightPadded(newStatesWit, dumNew, totalLen),
		)
	})

	// Internal consistency of the new columns
	manualCheckMiMCBlock(comp, blocks, oldStates, newStates)

	// And lookupize all MiMC queries parames into the central MiMC checking module
	for _, q := range mimcQueries {
		comp.InsertInclusion(
			round,
			ifaces.QueryID(mimcName(comp, "INCLUSION", q.ID)),
			[]ifaces.Column{blocks, oldStates, newStates},
			[]ifaces.Column{q.Blocks, q.OldState, q.NewState},
		)
	}

}
