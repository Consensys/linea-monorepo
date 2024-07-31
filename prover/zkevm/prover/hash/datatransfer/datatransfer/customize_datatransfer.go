package datatransfer

import (
	"github.com/consensys/linea-monorepo/prover/crypto/keccak"
	permTrace "github.com/consensys/linea-monorepo/prover/crypto/keccak"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
)

// NewCustomizedDataTransfer initializes a submodule for a special case where the lanes are provided directly.
// Currently used for the public inputs interconnection proofs when aggregating.
// warning: the submodule does not prove the relation among columns (lane,isFirstLaneFrmNewHash,isLaneActive)
// These relations should be handled before calling CostumDataTransfer.
func (mod *Module) NewCustomizedDataTransfer(comp *wizard.CompiledIOP, round, maxNumKeccakF int) {
	mod.MaxNumKeccakF = maxNumKeccakF
	maxNumRowsForLane := utils.NextPowerOfTwo(numLanesInBlock * maxNumKeccakF)

	// Declare lookup columns
	mod.LookUps = newLookupTables(comp)

	// declare columns
	// warning : these columns are not native and are not constrained here.
	mod.lane.isLaneActive = comp.InsertCommit(round, "IsLaneActive", maxNumRowsForLane)
	mod.lane.isFirstLaneOfNewHash = comp.InsertCommit(round, "IsFirstLaneOfNewHash", maxNumRowsForLane)
	mod.lane.lane = comp.InsertCommit(round, "Lane", maxNumRowsForLane)

	// base conversion over lanes, it converts lanes from uint64 into BaseA/BaseB of keccakf
	mod.BaseConversion.newBaseConversionOfLanes(comp, round, maxNumRowsForLane, mod.lane, mod.LookUps)
	// hashOutput
	mod.HashOutput.newHashOutput(comp, round, maxNumKeccakF)
}

// it assigns the columns
func (mod *Module) AssignCustomizedDataTransfer(
	run *wizard.ProverRuntime,
	permTrace permTrace.PermTraces,
) {
	maxNumRowsForLane := utils.NextPowerOfTwo(numLanesInBlock * mod.MaxNumKeccakF)

	// assign lane info
	mod.lane.assignLaneInfo(run, permTrace, maxNumRowsForLane)

	mod.BaseConversion.assignBaseConversion(run, mod.lane, maxNumRowsForLane)
	mod.HashOutput.AssignHashOutPut(run, permTrace)
}

// it assigns the lane-info (lane, isLaneActive, isFirstLaneFromNewHash)
func (l *lane) assignLaneInfo(run *wizard.ProverRuntime, trace keccak.PermTraces, maxNumRows int) {

	// populate the lane-info via the trace
	blocks := trace.Blocks
	var laneFr, isLaneFromNewHash, isActive []field.Element
	for j := range blocks {
		for i := range blocks[0] {
			laneFr = append(laneFr, field.NewElement(blocks[j][i]))
			isActive = append(isActive, field.One())
			if trace.IsNewHash[j] {
				if i == 0 {
					isLaneFromNewHash = append(isLaneFromNewHash, field.One())
				} else {
					isLaneFromNewHash = append(isLaneFromNewHash, field.Zero())
				}
			} else {
				isLaneFromNewHash = append(isLaneFromNewHash, field.Zero())
			}
		}
	}

	// assign lane-info
	run.AssignColumn(l.lane.GetColID(), smartvectors.RightZeroPadded(laneFr, maxNumRows))

	run.AssignColumn(l.isFirstLaneOfNewHash.GetColID(),
		smartvectors.RightZeroPadded(isLaneFromNewHash, maxNumRows))

	run.AssignColumn(l.isLaneActive.GetColID(), smartvectors.RightZeroPadded(isActive, maxNumRows))

}
