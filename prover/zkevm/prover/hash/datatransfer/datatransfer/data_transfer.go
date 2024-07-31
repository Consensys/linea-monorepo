/*
Package datatransfer implements the utilities and the submodules for transferring the data, from
the relevant arithmetization modules to the keccak module.

It includes;
1. Data Importing from arithmetization columns
2. Data Serialization (to make well-formed blocks for the use in the keccakf module)
3. Data Exporting to the keccakf module
*/
package datatransfer

import (
	permTrace "github.com/consensys/zkevm-monorepo/prover/crypto/keccak"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
	"github.com/consensys/zkevm-monorepo/prover/utils"
	"github.com/consensys/zkevm-monorepo/prover/zkevm/prover/hash/generic"
)

const (
	maxLanesFromLimb = 3                                // maximum number of lanes that fall over the same limb
	maxNByte         = 16                               // maximum size of the limb in bytes
	numBytesInLane   = 8                                // number of bytes in the lane
	numLanesInBlock  = 17                               // number of lanes in the block
	maxBlockSize     = numBytesInLane * numLanesInBlock // size of the block

	power8        = 1 << 8
	power16       = 1 << 16
	powerMaxNByte = 1 << maxNByte

	// for sha2
	maxLanesFromLimbSha2 = 5
	numBytesInLaneSha2   = 4
	numLanesInBlockSha2  = 16
	maxBlockSizeSha2     = numBytesInLaneSha2 * numLanesInBlockSha2
)

// Module consists of all the columns and submodules used for data transition.
type Module struct {
	// size of the data transfer module
	// NextPowerOfTwo(maxBlockSize * MaxNumKeccakF)
	MaxNumKeccakF int

	// Provider Trace; the arithmetization columns relevant to keccak
	Provider generic.GenericByteModule

	// SubModules specific to DataTransfer Module
	iPadd          importAndPadd
	cld            cleanLimbDecomposition
	sCLD           spaghettizedCLD
	lane           lane
	BaseConversion baseConversion

	// Lookups specific to dataTransfer Module
	LookUps lookUpTables

	HashOutput HashOutput
}

// It Imposes the constraints per subModule.
func (mod *Module) NewDataTransfer(comp *wizard.CompiledIOP, round, maxNumKeccakF, hashType int) {
	mod.MaxNumKeccakF = maxNumKeccakF
	maxNumRows := utils.NextPowerOfTwo(maxBlockSize * maxNumKeccakF)
	maxNumRowsForLane := utils.NextPowerOfTwo(numLanesInBlock * maxNumKeccakF)
	// Declare lookup columns
	mod.LookUps = newLookupTables(comp)

	// Define the subModules
	mod.iPadd.newImportAndPadd(comp, round, maxNumRows, mod.Provider, mod.LookUps, 0)
	mod.cld.newCLD(comp, round, mod.LookUps, mod.iPadd, maxNumRows, hashType)
	mod.sCLD.newSpaghetti(comp, round, mod.iPadd, mod.cld, maxNumRows)
	mod.lane.newLane(comp, round, maxNumRows, maxNumRowsForLane, mod.sCLD)
	mod.BaseConversion.newBaseConversionOfLanes(comp, round, maxNumRowsForLane, mod.lane, mod.LookUps)

	// hashOutput
	mod.HashOutput.newHashOutput(comp, round, maxNumKeccakF)
}

// It assigns the columns per subModule.
func (mod *Module) AssignModule(
	run *wizard.ProverRuntime,
	permTrace permTrace.PermTraces,
	gt generic.GenTrace) {
	maxNumRows := utils.NextPowerOfTwo(maxBlockSize * mod.MaxNumKeccakF)
	maxNumRowsForLane := utils.NextPowerOfTwo(numLanesInBlock * mod.MaxNumKeccakF)

	mod.iPadd.assignImportAndPadd(run, gt, maxNumRows, 0)
	mod.cld.assignCLD(run, mod.iPadd, maxNumRows)
	mod.sCLD.assignSpaghetti(run, mod.iPadd, mod.cld, maxNumRows)
	mod.lane.assignLane(run, mod.iPadd,
		mod.sCLD, permTrace, maxNumRows, maxNumRowsForLane)
	mod.BaseConversion.assignBaseConversion(run, mod.lane, maxNumRowsForLane)

	mod.HashOutput.AssignHashOutPut(run, permTrace)
}
