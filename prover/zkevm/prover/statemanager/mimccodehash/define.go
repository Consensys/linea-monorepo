package mimccodehash

import (
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/symbolic"
)

const (
	// Column names
	MIMC_CODE_HASH_IS_ACTIVE_NAME          ifaces.ColID = "MIMC_CODE_HASH_IS_ACTIVE"
	MIMC_CODE_HASH_CFI_NAME                ifaces.ColID = "MIMC_CODE_HASH_CFI"
	MIMC_CODE_HASH_LIMB_NAME               ifaces.ColID = "MIMC_CODE_HASH_LIMB"
	MIMC_CODE_HASH_IS_NEW_HASH_NAME        ifaces.ColID = "MIMC_CODE_HASH_IS_NEW_HASH"
	MIMC_CODE_HASH_IS_HASH_END_NAME        ifaces.ColID = "MIMC_CODE_HASH_IS_HASH_END"
	MIMC_CODE_HASH_PREV_STATE_NAME         ifaces.ColID = "MIMC_CODE_HASH_PREV_STATE"
	MIMC_CODE_HASH_NEW_STATE_NAME          ifaces.ColID = "MIMC_CODE_HASH_NEW_STATE"
	MIMC_CODE_HASH_CODE_SIZE_NAME          ifaces.ColID = "MIMC_CODE_HASH_CODE_SIZE"
	MIMC_CODE_HASH_KECCAK_CODEHASH_HI_NAME ifaces.ColID = "MIMC_CODE_HASH_KECCAK_CODEHASH_HI"
	MIMC_CODE_HASH_KECCAK_CODEHASH_LO_NAME ifaces.ColID = "MIMC_CODE_HASH_KECCAK_CODEHASH_LO"
)

// Inputs stores the caller's parameters to NewMiMCCodeHash
type Inputs struct {
	Round int
	Name  string
	Size  int
}

// Module stores all the columns responsible for computing the MiMC
// codehash of every contract occuring during the EVM computation.
type Module struct {
	// inputs are the parameteress provided by the user of the struct
	inputs Inputs

	// All the columns characterizing the module
	IsActive   ifaces.Column
	CFI        ifaces.Column
	Limb       ifaces.Column
	CodeHashHi ifaces.Column
	CodeHashLo ifaces.Column
	CodeSize   ifaces.Column
	IsNewHash  ifaces.Column
	IsHashEnd  ifaces.Column
	PrevState  ifaces.Column

	// Contains the MiMC code hash when IsHashEnd is 1
	NewState ifaces.Column

	// inputModule stores the modules connected the present Module (e.g. Rom/RomLex)
	// when they are not omitted.
	inputModules *inputModules
}

// NewModule registers and committing all the columns and queries in the mimc_code_hash module
func NewModule(comp *wizard.CompiledIOP, inputs Inputs) (mh Module) {

	mh = Module{
		inputs:     inputs,
		IsActive:   comp.InsertCommit(inputs.Round, MIMC_CODE_HASH_IS_ACTIVE_NAME, inputs.Size),
		CFI:        comp.InsertCommit(inputs.Round, MIMC_CODE_HASH_CFI_NAME, inputs.Size),
		Limb:       comp.InsertCommit(inputs.Round, MIMC_CODE_HASH_LIMB_NAME, inputs.Size),
		CodeHashHi: comp.InsertCommit(inputs.Round, MIMC_CODE_HASH_KECCAK_CODEHASH_HI_NAME, inputs.Size),
		CodeHashLo: comp.InsertCommit(inputs.Round, MIMC_CODE_HASH_KECCAK_CODEHASH_LO_NAME, inputs.Size),
		CodeSize:   comp.InsertCommit(inputs.Round, MIMC_CODE_HASH_CODE_SIZE_NAME, inputs.Size),
		IsNewHash:  comp.InsertCommit(inputs.Round, MIMC_CODE_HASH_IS_NEW_HASH_NAME, inputs.Size),
		IsHashEnd:  comp.InsertCommit(inputs.Round, MIMC_CODE_HASH_IS_HASH_END_NAME, inputs.Size),
		PrevState:  comp.InsertCommit(inputs.Round, MIMC_CODE_HASH_PREV_STATE_NAME, inputs.Size),
		NewState:   comp.InsertCommit(inputs.Round, MIMC_CODE_HASH_NEW_STATE_NAME, inputs.Size),
	}

	mh.checkConsistency(comp)
	return mh
}

// checkConsistency adds the constraints securing the MiMCCodeHash module.
//
//	We have the following constraints:
//
//	1. NewState = MiMC(PrevState, Limb)
//	2. If IsNewHash = 0, PrevState[i] = NextState[i-1] (in the active area)
//	3. If IsNewHash = 1, PrevState = 0 (in the active area)
//	4. If CFI incremented, IsNewHash = 1
//	5. Local constraint IsNewHash starts with 1
//	6. if CFI[i+1] - CFI[i] != 0, IsHashEnd[i] = 1
//	7. Booleanity of IsNewHash, IsHashEnd (in the active area)
//	8. Booeanity of IsActive
//	9. IsActive[i] = 0 IMPLIES IsActive[i+1] = 0
//	10. in a particular CFI segment, CodeHashHi and CodeHashLo remain constant
//	11. in a particular CFI segment, CodeSize remains constant
//	11. All columns are zero in the inactive area
func (mh *Module) checkConsistency(comp *wizard.CompiledIOP) {

	// NewState = MiMC(PrevState, Limb)
	comp.InsertMiMC(mh.inputs.Round, mh.qname("MiMC_CODE_HASH"), mh.Limb, mh.PrevState, mh.NewState)

	// If IsNewHash = 0, PrevState[i] = NewState[i-1] (in the active area), e.g.,
	// IsActive[i] * (1 - IsNewHash[i]) * (PrevState[i] - NextState[i-1]) = 0
	comp.InsertGlobal(mh.inputs.Round, mh.qname("PREV_STATE_CONSISTENCY_2"),
		symbolic.Mul(mh.IsActive,
			symbolic.Sub(1, mh.IsNewHash),
			symbolic.Sub(mh.PrevState, ifaces.ColumnAsVariable(column.Shift(mh.NewState, -1)))))

	// If IsNewHash = 1, PrevState = 0 (in the active area) e.g., IsActive[i] * IsNewHash[i] * PrevState[i] = 0
	comp.InsertGlobal(mh.inputs.Round, mh.qname("PREV_STATE_ZERO_AT_BEGINNING"),
		symbolic.Mul(mh.IsActive, mh.IsNewHash, mh.PrevState))

	// If CFI incremented, IsNewHash = 1, e.g., IsActive[i] * (CFI[i] - CFI[i-1]) * (1 - IsNewHash[i]) = 0
	comp.InsertGlobal(mh.inputs.Round, mh.qname("IS_NEW_HASH_CONSISTENCY_1"),
		symbolic.Mul(mh.IsActive,
			symbolic.Sub(mh.CFI, ifaces.ColumnAsVariable(column.Shift(mh.CFI, -1))),
			symbolic.Sub(1, mh.IsNewHash)))

	// Local constraint IsNewHash starts with 1
	comp.InsertLocal(mh.inputs.Round, mh.qname("IS_NEW_HASH_LOCAL"), symbolic.Sub(mh.IsNewHash, mh.IsActive))

	// if CFI[i+1] - CFI[i] != 0, IsHashEnd[i] = 1, e.g., IsActive[i] * (CFI[i+1] - CFI[i]) * (1 - IsHashEnd[i]) = 0
	comp.InsertGlobal(mh.inputs.Round, mh.qname("IS_HASH_END_CONSISTENCY_1"),
		symbolic.Mul(mh.IsActive,
			symbolic.Sub(ifaces.ColumnAsVariable(column.Shift(mh.CFI, 1)), mh.CFI),
			symbolic.Sub(1, mh.IsHashEnd)))

	// Booleanity of IsNewHash, IsHashEnd (in the active area)
	comp.InsertGlobal(mh.inputs.Round, mh.qname("IS_NEW_HASH_BOOLEAN"),
		symbolic.Sub(symbolic.Mul(symbolic.Square(mh.IsNewHash), mh.IsActive),
			mh.IsNewHash))

	comp.InsertGlobal(mh.inputs.Round, mh.qname("IS_HASH_END_BOOLEAN"),
		symbolic.Sub(symbolic.Mul(symbolic.Square(mh.IsHashEnd), mh.IsActive),
			mh.IsHashEnd))

	// Booeanity of IsActive
	comp.InsertGlobal(mh.inputs.Round, mh.qname("IS_ACTIVE_BOOLEAN"),
		symbolic.Sub(
			symbolic.Square(mh.IsActive),
			mh.IsActive))

	// IsActive[i] = 0 IMPLIES IsActive[i+1] = 0 e.g. IsActive[i] = IsActive[i-1] * IsActive[i]
	comp.InsertGlobal(mh.inputs.Round, mh.qname("IS_ACTIVE_ZERO_FOLLOWED_BY_ZERO"),
		symbolic.Sub(mh.IsActive,
			symbolic.Mul(ifaces.ColumnAsVariable(column.Shift(mh.IsActive, -1)),
				mh.IsActive)))

	// In a particular CFI segment, CodeHashHi and CodeHashLo remain constant,
	// e.g., IsActive[i] * (1 - IsEndHash[i]) * (CodeHashHi[i+1] - CodeHashHi[i]) = 0 and,
	// IsActive[i] * (1 - IsEndHash[i]) * (CodeHashLo[i+1] - CodeHashLo[i]) = 0
	comp.InsertGlobal(mh.inputs.Round, mh.qname("CODE_HASH_HI_SEGMENT_WISE_CONSTANT"),
		symbolic.Mul(mh.IsActive,
			symbolic.Sub(1, mh.IsHashEnd),
			symbolic.Sub(ifaces.ColumnAsVariable(column.Shift(mh.CodeHashHi, 1)), mh.CodeHashHi)))

	comp.InsertGlobal(mh.inputs.Round, mh.qname("CODE_HASH_LO_SEGMENT_WISE_CONSTANT"),
		symbolic.Mul(mh.IsActive,
			symbolic.Sub(1, mh.IsHashEnd),
			symbolic.Sub(ifaces.ColumnAsVariable(column.Shift(mh.CodeHashLo, 1)), mh.CodeHashLo)))

	// In a particular CFI segment, CodeSize remains constant,
	// e.g., IsActive[i] * (1 - IsEndHash[i]) * (CodeSize[i+1] - CodeSize[i]) = 0
	comp.InsertGlobal(mh.inputs.Round, mh.qname("CODE_SIZE_SEGMENT_WISE_CONSTANT"),
		symbolic.Mul(mh.IsActive,
			symbolic.Sub(1, mh.IsHashEnd),
			symbolic.Sub(ifaces.ColumnAsVariable(column.Shift(mh.CodeSize, 1)), mh.CodeSize)))

	// All columns are zero in the inactive area, except newState
	mh.colZeroAtInactive(comp, mh.CFI, "CFI_ZERO_IN_INACTIVE")
	mh.colZeroAtInactive(comp, mh.Limb, "LIMB_ZERO_IN_INACTIVE")
	mh.colZeroAtInactive(comp, mh.IsNewHash, "IS_NEW_HASH_ZERO_IN_INACTIVE")
	mh.colZeroAtInactive(comp, mh.IsHashEnd, "IS_HASH_END_ZERO_IN_INACTIVE")
	mh.colZeroAtInactive(comp, mh.CodeHashHi, "CODE_HASH_HI_ZERO_IN_INACTIVE")
	mh.colZeroAtInactive(comp, mh.CodeHashLo, "CODE_HASH_LO_ZERO_IN_INACTIVE")
	mh.colZeroAtInactive(comp, mh.CodeSize, "CODE_SIZE_ZERO_IN_INACTIVE")
	mh.colZeroAtInactive(comp, mh.PrevState, "PREV_STATE_ZERO_IN_INACTIVE")
}

// Function returning a query name
func (mh *Module) qname(name string, args ...any) ifaces.QueryID {
	return ifaces.QueryIDf("%v", mh.inputs.Name) + "_" + ifaces.QueryIDf(name, args...)
}

// Function inserting a query that col is zero when IsActive is zero
func (mh *Module) colZeroAtInactive(comp *wizard.CompiledIOP, col ifaces.Column, name string) {
	// col zero at inactive area, e.g., (1-IsActive[i]) * col[i] = 0
	comp.InsertGlobal(mh.inputs.Round, mh.qname(name),
		symbolic.Mul(symbolic.Sub(1, mh.IsActive), col))
}
