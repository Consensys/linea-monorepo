package statesummary

import (
	"github.com/consensys/zkevm-monorepo/prover/protocol/column"
	"github.com/consensys/zkevm-monorepo/prover/protocol/ifaces"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
	sym "github.com/consensys/zkevm-monorepo/prover/symbolic"
	"github.com/consensys/zkevm-monorepo/prover/utils"
)

// StateSummary represents the state-summary module. It defines all the columns
// constraints and assigment methods for this module. The state-summary module
// is tasked with:
//
//   - “commanding” the accumulator module and providing a sequential
//     view of all the shomei operations.
//   - Hashing the accounts, account-addresses, storage-keys and storage-slots
//   - Ensuring the sequentiality of the root hashes
//   - There is only a single trace-segment for each block and account.
//   - The segments consists in a succession of
//
// The state-summary module **is not** tasked with:
//
//   - Ensuring the correctness of the state-access patterns (the accumulator's
//     properties are that any defect of the traces is either irrelevant or
//     detected). This is only true for what goes beside the above-required
//     points.
//
// The main heart-beat of the state-summary module is 1 row == 1 accumulator
// trace issuance.
type StateSummary struct {

	// AccountAddress represents which account is being peeked by the module.
	// It is assigned by providing
	AccountAddress ifaces.Column

	// LookBackDeltaAddress is an isZeroCtx assessing whether the current value
	// of AccountAddress is equal to the previous one. It is handled via a
	// dedicated sub-context.
	LookBackDeltaAddress lookBackDeltaAddressIsZeroCtx

	// IsActive is a module-wide flag indicating whether the current row
	// corresponds to a state operation or a padding row in which nothing is
	// happening. Most of the constraints are cancelled for the padding rows.
	IsActive ifaces.Column

	// IsEndOfAccountSegment is a binary columns indicating the end of an
	// account segment.
	IsEndOfAccountSegment, IsBeginningOfAccountSegment ifaces.Column

	// IsInitialDeployment and IsFinalDeployment are columns assigned to boolean
	// values whose row indicates whether the current segment relates to the
	// initial or the final deployment of an account. This is needed essentially
	// to support contract redeployment. For "normal" account segments, this is
	// always assigned to 1.
	IsInitialDeployment, IsFinalDeployment ifaces.Column

	// IsStorage indicates that the current row is peaking at a storage slot
	// of the account. The negative case means we are looking at the account
	// in the world-state trie and are terminating a segment.
	IsStorage ifaces.Column

	// BatchNumber represents the index of a block as part of the conflation.
	// It contrasts with the block-number in the sense that it always starts from
	// 0 for the first block of the conflation and then increases.
	BatchNumber ifaces.Column

	// WorldStateRoot stores the state-root hashes.
	WorldStateRoot ifaces.Column

	// InitialAccount and FinalAccount represents the values stored in the
	// account at the beginning and the end of the trace segment.
	InitialAccount, FinalAccount AccountPeek

	// StoragePeek and FinalStorage represent the initial and final value
	// of a storage slot currently being peeked at.
	StoragePeek StoragePeek

	// AccountHashing is a compilation sub-routines responsible for proving the
	// hashing of the initial and final account peek.
	AccountHashing AccountHashing

	// StorageHashing is a compilation sub-routines responsible for proving the
	// hashing of the initial and final account peek.
	StorageHashing StorageHashing

	// AccumulatorStatement contains the statement values associated with the
	// accumulator.
	AccumulatorStatement AccumulatorStatement
}

func DefineStateSummaryModule(comp *wizard.CompiledIOP, size int) StateSummary {

	if !utils.IsPowerOfTwo(size) {
		utils.Panic("size must be power of two, got %v", size)
	}

	// createCol is function to quickly create a column
	createCol := func(name string) ifaces.Column {
		return comp.InsertCommit(0, ifaces.ColIDf("STATE_SUMMARY_%v", name), size)
	}

	res := StateSummary{
		AccountAddress:              createCol("ACCOUNT_ADDRESS"),
		IsActive:                    createCol("IS_ACTIVE"),
		IsEndOfAccountSegment:       createCol("IS_END_OF_ACCOUNT_SEGMENT"),
		IsBeginningOfAccountSegment: createCol("IS_BEGINNING_OF_ACCOUNT_SEGMENT"),
		IsInitialDeployment:         createCol("IS_INITIAL_DEPLOYMENT"),
		IsFinalDeployment:           createCol("IS_FINAL_DEPLOYMENT"),
		IsStorage:                   createCol("IS_STORAGE"),
		BatchNumber:                 createCol("BATCH_NUMBER"),
		WorldStateRoot:              createCol("WORLD_STATE_ROOT"),
		InitialAccount:              newAccountPeek(comp, size, "INITIAL_ACCOUNT"),
		FinalAccount:                newAccountPeek(comp, size, "FINAL_ACCOUNT"),
		StoragePeek:                 newStoragePeek(comp, size, "STORAGE_PEEK"),
		AccountHashing:              newAccountHashing(comp, size, "ACCOUNT_HASHING"),
		StorageHashing:              newStorageHashing(comp, size, "STORAGE_HASHING"),
		AccumulatorStatement:        newAccumulatorStatement(comp, size, "ACCUMULATOR_STATEMENT"),
	}

	res.csLookBackDeltaAddress(comp)
	res.csIsActive(comp)
	res.csIsEndOfAccountSegment(comp)
	res.csIsBeginningOfAccountSegment(comp)
	res.csInitialFinalDeployment(comp)
	res.csIsStorage(comp)
	res.csBatchNumberCanOnlyIncrement(comp)
	res.csAccountNew(comp)
	res.csAccountOld(comp)
	res.csPickHKeyAndHVal(comp)
	res.csAccountOrStorageValZeroized(comp)
	res.csWorldStateRootSequentiality(comp)
	res.csStorageRootHashSequentiality(comp)
	res.AccountHashing.csIntermediateHashesAreWellComputed(comp, &res.InitialAccount, &res.FinalAccount, res.AccountAddress)
	res.StorageHashing.csHashCorrectness(comp, &res.StoragePeek)
	res.AccumulatorStatement.csAccumulatorStatementFlags(comp, res.IsActive)

	return res

}

// csLookBackDeltaAddress constructs the associated IsZeroCtx
func (ss *StateSummary) csLookBackDeltaAddress(comp *wizard.CompiledIOP) {
	ss.LookBackDeltaAddress = LookBackDeltaAddressIsZeroCtx(
		comp, ss.AccountAddress, ss.IsActive,
	)
}

// csIsActive constrains the IsActiveFlag so that:
//   - It is a binary column
//   - It cannot transition from 0 to 1
//   - It can only transition when ss.AccountDeltaLookBack.IsZero is zero
func (ss *StateSummary) csIsActive(comp *wizard.CompiledIOP) {

	comp.InsertGlobal(
		0,
		"STATE_SUMMARY_IS_ACTIVE_IS_BINARY",
		sym.Mul(ss.IsActive, sym.Sub(ss.IsActive, 1)),
	)

	comp.InsertGlobal(
		0,
		"STATE_SUMMARY_IS_ACTIVE_NO_0_TO_1",
		sym.Sub(ss.IsActive, sym.Mul(column.Shift(ss.IsActive, -1), ss.IsActive)),
	)

	comp.InsertGlobal(
		0,
		"STATE_SUMMARY_IS_ACTIVE_TRANSITION_AFTER_ACCOUNT_SEGMENT",
		sym.Mul(
			ss.LookBackDeltaAddress.IsZero,
			sym.Sub(ss.IsActive, column.Shift(ss.IsActive, -1)),
		),
	)
}

// IsBeginningOfAccountSegment is a virtual binary column indicating whether
//
//	AccountAddress[i] != AccountAddress[i-1]. It is virtual in the sense that
//
// it is not a column that we commit to but an expression derived from already
// committed columns.
//
// When the inactive flag is set to 0, the column returns also zero.
func (ss *StateSummary) csIsBeginningOfAccountSegment(comp *wizard.CompiledIOP) {
	mustBeBinary(comp, ss.IsBeginningOfAccountSegment)
	isZeroWhenInactive(comp, ss.IsBeginningOfAccountSegment, ss.IsActive)

	// IsBeginningOfAccountSegment being one implies that either the batchNumber
	// bumped or that we jumped to another account and reciprocally.
	comp.InsertGlobal(
		0,
		ifaces.QueryIDf("STATE_SUMMARY_IBOAS_IS_ONE_IFF_EITHER_NEW_BATCH_OR_NEW_ADDR"),
		sym.Add(
			sym.Mul(
				ss.IsActive,
				ss.IsBeginningOfAccountSegment,
				sym.Sub(ss.BatchNumber, column.Shift(ss.BatchNumber, -1), 1),
				ss.LookBackDeltaAddress.IsZero,
			),
			sym.Mul(
				ss.IsActive,
				sym.Sub(1, ss.IsBeginningOfAccountSegment),
				sym.Add(
					sym.Sub(ss.BatchNumber, column.Shift(ss.BatchNumber, -1)),
					sym.Sub(1, ss.LookBackDeltaAddress.IsZero),
				),
			),
		),
	)

	comp.InsertLocal(
		0,
		ifaces.QueryIDf("STATE_SUMMARY_IBOA_STARTS_AT_ONE"),
		sym.Sub(1, ss.IsBeginningOfAccountSegment),
	)
}

// IsEndOfAccountSegment is a virtual binary column indicating whether:
//
//	ss.LookBackDeltaAddress.IsZero[i+1] == 0
//		OR
//	isActive[i] == 1 && isActive[i+1] == 0
//
// It must be zero when inactive is set to 0.
//
// At the last row, there is a boundary condition specifying that the column
// must be equal to IsInactive.
func (ss *StateSummary) csIsEndOfAccountSegment(comp *wizard.CompiledIOP) {

	mustBeBinary(comp, ss.IsEndOfAccountSegment)
	isZeroWhenInactive(comp, ss.IsEndOfAccountSegment, ss.IsActive)

	// IsEndOfSegment == 1 <==>
	// 		IBOAS[i=1] == 1 || (isActive[i+1] + 1 == isActive[i])
	comp.InsertGlobal(
		0,
		"STATE_SUMMARY_IS_END_OF_ACCOUNT_SEGMENT_PROPERLY_SET",
		sym.Add(
			sym.Mul(
				ss.IsEndOfAccountSegment,
				sym.Sub(1, column.Shift(ss.IsBeginningOfAccountSegment, 1)),
				sym.Sub(ss.IsActive, column.Shift(ss.IsActive, 1), 1),
			),
			sym.Mul(
				sym.Sub(1, ss.IsEndOfAccountSegment),
				sym.Add(
					sym.Sub(ss.IsActive, column.Shift(ss.IsActive, 1)),
					column.Shift(ss.IsBeginningOfAccountSegment, 1),
				),
			),
		),
	)

	comp.InsertLocal(
		0,
		"STATE_SUMMARY_IS_END_OF_ACCOUNT_SEGMENT_BOUNDARY_CONDITION",
		sym.Sub(
			column.Shift(ss.IsActive, -1),
			column.Shift(ss.IsEndOfAccountSegment, -1),
		),
	)
}

// csInitialFinalDeployment ensures that the columns ss.IsInitialDeployment and
// ss.IsFinalDeployment are both binary columns such that at least one of the
// two columns is set to 1. This is done with 3 booleanity checks:
//
//   - One for IsInitialDeployment
//   - One for IsFinalDeployment
//   - One for IsInitial + IsFinalDeployment - IsActive
//   - A check that (IsInitial + IsFinal) * (1 - IsActive) == 0
//
// Additionally, the function constrains that:
//
//   - At the beginning of an account segment: IsInitialDeployment	== 1
//   - At the end of an account segment: 		IsFinalDeployment 	== 1
//   - They must remain constant unless IsStorage is false in the previous row.
func (ss *StateSummary) csInitialFinalDeployment(comp *wizard.CompiledIOP) {

	comp.InsertGlobal(
		0,
		"STATE_SUMMARY_IS_INITIAL_DEPLOYMENT_BINARY",
		sym.Mul(ss.IsInitialDeployment, sym.Sub(ss.IsInitialDeployment, 1)),
	)

	comp.InsertGlobal(
		0,
		"STATE_SUMMARY_IS_FINAL_DEPLOYMENT_BINARY",
		sym.Mul(ss.IsFinalDeployment, sym.Sub(ss.IsFinalDeployment, 1)),
	)

	// mustBeBinary is expected to be 0 when one of the flag is set and 1 when
	// both flags are set.
	mustBeBinary := sym.Add(ss.IsInitialDeployment, ss.IsFinalDeployment, sym.Neg(ss.IsActive))

	comp.InsertGlobal(
		0,
		"STATE_SUMMARY_IS_INITIAL_FINAL_DEPLOYMENT_ONE_FLAG_AT_LEAST_IS_SET",
		sym.Mul(mustBeBinary, sym.Sub(mustBeBinary, 1)),
	)

	comp.InsertGlobal(
		0,
		"STATE_SUMMARY_IS_INITIAL_FINAL_DEPLOYMENT_ARE_ZERO_WHEN_INACTIVE",
		sym.Mul(
			sym.Add(ss.IsInitialDeployment, ss.IsFinalDeployment),
			sym.Sub(1, ss.IsActive),
		),
	)

	comp.InsertGlobal(
		0,
		"STATE_SUMMARY_IS_INITIAL_DEPLOYMENT_AT_THE_BEGINNING_OF_AN_ACCOUNT_SEGMENT",
		sym.Mul(
			ss.IsBeginningOfAccountSegment,
			sym.Sub(1, ss.IsInitialDeployment),
		),
	)

	comp.InsertGlobal(
		0,
		"STATE_SUMMARY_IS_FINAL_DEPLOYMENT_AT_THE_END_OF_AN_ACCOUNT_SEGMENT",
		sym.Mul(
			ss.IsEndOfAccountSegment,
			sym.Sub(1, ss.IsFinalDeployment),
		),
	)

	comp.InsertGlobal(
		0,
		"STATE_SUMMARY_INITIAL_FINAL_DEPLOYMENT_STAY_CONSTANT_UNLESS_IS_STORAGE_WAS_FALSE",
		sym.Mul(
			ss.IsActive,
			column.Shift(ss.IsStorage, -1),
			sym.Sub(
				sym.Sub(ss.IsFinalDeployment, ss.IsInitialDeployment),
				sym.Sub(
					column.Shift(ss.IsFinalDeployment, -1),
					column.Shift(ss.IsInitialDeployment, -1),
				),
			),
		),
	)
}

// csIsStorageSequentiality adds the relevant constraints to ensure that the
// IsStorageFlag is properly set.
//
//   - It is a binary flag
//   - It must be zero if this is the end of an account segment
//   - It must be zero if the (initialDeployment-finalDeployment) transition from
//     1 to -1.
//   - It must be zero in all other situations.
func (ss *StateSummary) csIsStorage(comp *wizard.CompiledIOP) {

	comp.InsertGlobal(
		0,
		"STATE_SUMMARY_IS_STORAGE_IS_BOOLEAN",
		sym.Mul(
			ss.IsStorage,
			sym.Sub(ss.IsStorage, 1),
		),
	)

	comp.InsertGlobal(
		0,
		"STATE_SUMMARY_IS_STORAGE_IS_ONE_WHEN_END_OF_ACCOUNT_SEGMENT",
		sym.Sub(
			ss.IsEndOfAccountSegment,
			sym.Mul(
				ss.IsEndOfAccountSegment,
				sym.Sub(1, ss.IsStorage),
			),
		),
	)

	// In the middle of an account segment
	// !isStorage && !isEndOfAccountSegment <=> isInitialDeployment 1->0
	comp.InsertGlobal(
		0,
		"STATE_SUMMARY_IS_STORAGE_IN_THE_MIDDLE_OF_AN_ACCOUNT_SEGMENT",
		sym.Mul(
			ss.IsActive,
			sym.Sub(1, column.Shift(ss.IsStorage, -1)),
			sym.Sub(1, column.Shift(ss.IsEndOfAccountSegment, -1)),
			sym.Add(
				sym.Sub(1, column.Shift(ss.IsFinalDeployment, -1)),
				column.Shift(ss.IsInitialDeployment, -1),
				ss.IsFinalDeployment,
				sym.Sub(1, ss.IsInitialDeployment),
				-4,
			),
		),
	)

}

// csBatchNumberCanOnlyIncrement ensures that the BatchNumber column can only
// increment by one. And it can only do so if the
func (ss *StateSummary) csBatchNumberCanOnlyIncrement(comp *wizard.CompiledIOP) {

	comp.InsertGlobal(
		0,
		"STATE_SUMMARY_BATCH_NUMBER_CAN_ONLY_INCREMENT",
		sym.Mul(
			ss.IsActive,
			sym.Sub(ss.BatchNumber, column.Shift(ss.BatchNumber, -1)),
			sym.Sub(ss.BatchNumber, column.Shift(ss.BatchNumber, -1), 1),
		),
	)

	comp.InsertGlobal(
		0,
		"STATE_SUMMARY_BATCH_NUMBER_CAN_ONLY_CHANGE_ON_A_NEW_SEGMENT",
		sym.Mul(
			ss.IsActive,
			sym.Sub(ss.BatchNumber, column.Shift(ss.BatchNumber, -1)),
			sym.Sub(1, ss.IsBeginningOfAccountSegment),
		),
	)

	comp.InsertLocal(
		0,
		"STATE_SUMMARY_BATCH_NUMER_START_FROM_ZERO",
		sym.NewVariable(ss.BatchNumber),
	)
}

// csAccountNew ensures that the account new is updated consistently with the
// account segment structure.
func (ss *StateSummary) csAccountNew(comp *wizard.CompiledIOP) {

	comp.InsertGlobal(
		0,
		"STATE_SUMMARY_NEW_BALANCE_IS_CONSTANT_WHEN_IS_STORAGE",
		sym.Mul(
			ss.IsActive,
			sym.Sub(1, ss.AccumulatorStatement.IsDelete, ss.AccumulatorStatement.IsInsert),
			column.Shift(ss.IsStorage, -1),
			deltaOf(ss.FinalAccount.Balance),
		),
	)

	comp.InsertGlobal(
		0,
		"STATE_SUMMARY_NEW_NONCE_IS_CONSTANT_WHEN_IS_STORAGE",
		sym.Mul(
			ss.IsActive,
			sym.Sub(1, ss.AccumulatorStatement.IsDelete, ss.AccumulatorStatement.IsInsert),
			column.Shift(ss.IsStorage, -1),
			deltaOf(ss.FinalAccount.Nonce),
		),
	)

	comp.InsertGlobal(
		0,
		"STATE_SUMMARY_NEW_CODESIZE_IS_CONSTANT_WHEN_IS_STORAGE",
		sym.Mul(
			ss.IsActive,
			sym.Sub(1, ss.AccumulatorStatement.IsDelete, ss.AccumulatorStatement.IsInsert),
			column.Shift(ss.IsStorage, -1),
			deltaOf(ss.FinalAccount.CodeSize),
		),
	)

	comp.InsertGlobal(
		0,
		"STATE_SUMMARY_NEW_KECCAK_CODEHASH_HI_IS_CONSTANT_WHEN_IS_STORAGE",
		sym.Mul(
			ss.IsActive,
			sym.Sub(1, ss.AccumulatorStatement.IsDelete, ss.AccumulatorStatement.IsInsert),
			column.Shift(ss.IsStorage, -1),
			deltaOf(ss.FinalAccount.KeccakCodeHash.Hi),
		),
	)

	comp.InsertGlobal(
		0,
		"STATE_SUMMARY_NEW_KECCAK_CODEHASH_LO_IS_CONSTANT_WHEN_IS_STORAGE",
		sym.Mul(
			ss.IsActive,
			sym.Sub(1, ss.AccumulatorStatement.IsDelete, ss.AccumulatorStatement.IsInsert),
			column.Shift(ss.IsStorage, -1),
			deltaOf(ss.FinalAccount.KeccakCodeHash.Lo),
		),
	)

	comp.InsertGlobal(
		0,
		"STATE_SUMMARY_NEW_MIMC_CODEHASH_IS_CONSTANT_WHEN_IS_STORAGE",
		sym.Mul(
			ss.IsActive,
			sym.Sub(1, ss.AccumulatorStatement.IsDelete, ss.AccumulatorStatement.IsInsert),
			column.Shift(ss.IsStorage, -1),
			deltaOf(ss.FinalAccount.MiMCCodeHash),
		),
	)

	comp.InsertGlobal(
		0,
		"STATE_SUMMARY_NEW_STORAGE_ROOT_IS_CONSTANT_WHEN_IS_NOT_STORAGE",
		sym.Mul(
			ss.IsActive,
			sym.Sub(1, ss.AccumulatorStatement.IsDelete, ss.AccumulatorStatement.IsInsert),
			sym.Sub(1, ss.IsStorage),
			deltaOf(ss.FinalAccount.StorageRoot),
		),
	)
}

// csAccountOld ensures that the account new is updated consistently with the
// account segment structure. Namely, the initial account value may only be the
// one of the
func (ss *StateSummary) csAccountOld(comp *wizard.CompiledIOP) {

	comp.InsertGlobal(
		0,
		"STATE_SUMMARY_OLD_BALANCE_IS_CONSTANT_WHEN_NOT_BEGINNING_OF_SEGMENT",
		sym.Mul(
			column.Shift(ss.IsStorage, -1),
			deltaOf(ss.InitialAccount.Balance),
		),
	)

	comp.InsertGlobal(
		0,
		"STATE_SUMMARY_OLD_NONCE_IS_CONSTANT_WHEN_NOT_BEGINNING_OF_SEGMENT",
		sym.Mul(
			column.Shift(ss.IsStorage, -1),
			deltaOf(ss.InitialAccount.Nonce),
		),
	)

	comp.InsertGlobal(
		0,
		"STATE_SUMMARY_OLD_CODESIZE_IS_CONSTANT_WHEN_NOT_BEGINNING_OF_SEGMENT",
		sym.Mul(
			column.Shift(ss.IsStorage, -1),
			deltaOf(ss.InitialAccount.CodeSize),
		),
	)

	comp.InsertGlobal(
		0,
		"STATE_SUMMARY_OLD_KECCAK_CODEHASH_HI_IS_CONSTANT_WHEN_NOT_BEGINNING_OF_SEGMENT",
		sym.Mul(
			column.Shift(ss.IsStorage, -1),
			deltaOf(ss.InitialAccount.KeccakCodeHash.Hi),
		),
	)

	comp.InsertGlobal(
		0,
		"STATE_SUMMARY_OLD_KECCAK_CODEHASH_LO_IS_CONSTANT_WHEN_NOT_BEGINNING_OF_SEGMENT",
		sym.Mul(
			column.Shift(ss.IsStorage, -1),
			deltaOf(ss.InitialAccount.KeccakCodeHash.Lo),
		),
	)

	comp.InsertGlobal(
		0,
		"STATE_SUMMARY_OLD_MIMC_CODEHASH_IS_CONSTANT_WHEN_NOT_BEGINNING_OF_SEGMENT",
		sym.Mul(
			column.Shift(ss.IsStorage, -1),
			deltaOf(ss.InitialAccount.MiMCCodeHash),
		),
	)

	comp.InsertGlobal(
		0,
		"STATE_SUMMARY_OLD_STORAGE_ROOT_IS_CONSTANT_WHEN_NOT_BEGINNING_OF_SEGMENT",
		sym.Mul(
			column.Shift(ss.IsStorage, -1),
			deltaOf(ss.InitialAccount.StorageRoot),
		),
	)
}

// csWorldStateRootSequentiality ensures that the WorldStateRoot column is
// properly set w.r.t. the accumulator statement.
//
// The root must be constant, unless a state-access happen on the world-state
// trie. In this case, the old and new values of the world-state must be
// consistent with the value of the accumulator. More in details:
//
// if IsStorage == 1, then:
//
//	WorldStateRoot[i-1]	== WorldStateRoot[i]
//
// if IsStorage == 0, then:
//
//	WorldStateRoot[i-1]	== AccumulatorStatement.InitialRoot
//	WorldStateRoot[i]	== AccumulatorStatement.FinalRoot
//
// This constraint has an edge-case in the initial row because of the following: If this row is
// storage access, this is "fine" but if this is a world-state access, we have
// that the first root of the column is **NOT** the initial state-root hash.
// This has to be taken into account when extracting the initial state-root
// hash from the module. In this case, the initial state-root hash is found in
// the first row of the AccumulatorStatement.InitialRoot column.
func (ss *StateSummary) csWorldStateRootSequentiality(comp *wizard.CompiledIOP) {

	var (
		prevRoot   = column.Shift(ss.WorldStateRoot, -1)
		currRoot   = ss.WorldStateRoot
		isWsAccess = sym.Mul(sym.Sub(1, ss.IsStorage), ss.IsActive)
		accOldRoot = ss.AccumulatorStatement.InitialRoot
		accNewRoot = ss.AccumulatorStatement.FinalRoot
	)

	comp.InsertGlobal(
		0,
		"STATE_SUMMARY_WORLD_STATE_ROOT_SEQUENTIALITY_WHEN_NOT_WS_ACCESS",
		sym.Mul(ss.IsStorage, sym.Sub(prevRoot, currRoot)),
	)

	comp.InsertGlobal(
		0,
		"STATE_SUMMARY_WORLD_STATE_ROOT_SEQUENTIALITY_WHEN_WS_ACCESS_OLD_ROOT",
		sym.Mul(isWsAccess, sym.Sub(prevRoot, accOldRoot)),
	)

	comp.InsertGlobal(
		0,
		"STATE_SUMMARY_WORLD_STATE_ROOT_SEQUENTIALITY_WHEN_WS_ACCESS_NEW_ROOT",
		sym.Mul(isWsAccess, sym.Sub(currRoot, accNewRoot)),
	)
}

// csStorageRootHashSequentiality constrains the InitialAccount and FinalAccount
// to have StorageRootHash that is consistent with the statement of the
// accumulator when the IsStorage flag is up.
//
// For that to work, we discriminate two cases:
//  1. The current row corresponds to the beginning of an account segment
//  2. The current row does not correspond to the beginning of an account segment
//
// For case 1. the old root hash is in InitAccount.StorageRootHash
// For case 2. the old root hash is in FinalAccount.StorageRootHash[i-1]
//
// We discriminate between the two cases using the IsStorage by noticing that
// a transition 0 => 1 happens only at the beginning of an account segment
// in which the storage is being peeked at. This will not reveal the cases where
// the storage is not touched within an account segment (for instance the
// modification of an EOA). But this case is not relevant in the present
// situation. The other relevant case is the transition 1 => 1.
//
// Either way, the final root hash in the accumulator statement should match the
// FinalAccount.StorageRootHash in both cases.
func (ss *StateSummary) csStorageRootHashSequentiality(comp *wizard.CompiledIOP) {

	// IsBeginningOfAccountSegment && IsStorage

	comp.InsertGlobal(
		0,
		"STATE_SUMMARY_STORAGE_ROOT_HASH_SEQUENTIALITY_BEGINNING_OF_SUBSEGMENT",
		sym.Mul(
			ss.IsStorage,
			sym.Sub(1, column.Shift(ss.IsStorage, -1)),
			sym.Add(
				sym.Mul(
					ss.InitialAccount.Exists,
					sym.Sub(ss.InitialAccount.StorageRoot, ss.AccumulatorStatement.InitialRoot),
				),
				sym.Mul(
					sym.Sub(1, ss.InitialAccount.Exists),
					sym.Sub(emptyStorageRoot, ss.AccumulatorStatement.InitialRoot),
				),
			),
		),
	)

	comp.InsertLocal(
		0,
		"STATE_SUMMARY_STORAGE_ROOT_HASH_SEQUENTIALITY_FIRST_ROW",
		sym.Mul(
			ss.IsStorage,
			sym.Add(
				sym.Mul(
					ss.InitialAccount.Exists,
					sym.Sub(ss.InitialAccount.StorageRoot, ss.AccumulatorStatement.InitialRoot),
				),
				sym.Mul(
					sym.Sub(1, ss.InitialAccount.Exists),
					sym.Sub(emptyStorageRoot, ss.AccumulatorStatement.InitialRoot),
				),
			),
		),
	)

	comp.InsertGlobal(
		0,
		"STATE_SUMMARY_ACCOUNT_ROOT_HASH_SEQUENTIALITY_MIDDLE_OF_SUBSEGMENT",
		sym.Mul(
			ss.IsStorage,
			column.Shift(ss.IsStorage, -1),
			sym.Sub(
				column.Shift(ss.FinalAccount.StorageRoot, -1),
				ss.AccumulatorStatement.InitialRoot,
			),
		),
	)

	comp.InsertGlobal(
		0,
		"STATE_SUMMARY_STORAGE_ROOT_HASH_SEQUENTIALITY_MIDDLE_OF_SEGMENT",
		sym.Mul(
			ss.IsStorage,
			sym.Sub(
				ss.FinalAccount.StorageRoot,
				ss.AccumulatorStatement.FinalRoot,
			),
		),
	)
}

// csPickHKeyAndHVal ensures that the hkeys and kvals provided to the
// accumulator statement are picked from the correct origin. Either from the
// StorageHashing or from the AccountHashing. This is done using binary
// selectors.
func (ss *StateSummary) csPickHKeyAndHVal(comp *wizard.CompiledIOP) {

	comp.InsertGlobal(
		0,
		ifaces.QueryIDf("STATE_SUMMARY_ACC_STATEMENT_HKEY"),
		sym.Sub(
			ss.AccumulatorStatement.HKey,
			sym.Mul(
				ss.IsActive,
				ternary(ss.IsStorage, ss.StorageHashing.KeyResult(), ss.AccountHashing.AddressResult()),
			),
		),
	)

	comp.InsertGlobal(
		0,
		ifaces.QueryIDf("STATE_SUMMARY_ACC_STATEMENT_INITIAL_HVAL"),
		sym.Sub(
			ss.AccumulatorStatement.InitialHVal,
			sym.Mul(
				sym.Sub(ss.IsActive, ss.AccumulatorStatement.IsReadZero, ss.AccumulatorStatement.IsInsert),
				ternary(ss.IsStorage, ss.StorageHashing.OldValueResult(), ss.AccountHashing.OldAccountResult()),
			),
		),
	)

	comp.InsertGlobal(
		0,
		ifaces.QueryIDf("STATE_SUMMARY_ACC_STATEMENT_FINAL_HVAL"),
		sym.Sub(
			ss.AccumulatorStatement.FinalHVal,
			sym.Mul(
				sym.Sub(ss.IsActive, ss.AccumulatorStatement.IsReadZero, ss.AccumulatorStatement.IsDelete),
				ternary(ss.IsStorage, ss.StorageHashing.NewValueResult(), ss.AccountHashing.NewAccountResult()),
			),
		),
	)
}

// csAccountOrStorageValZeroized ensures that either the account value or the storage
// value of the "old" side of the state operation is zeroed out when the
// accumulator operation is READ_ZERO or INSERT.
//
// Inversely, the "new" side of the state operation is zeroed out when the
// the accumulator operation is either READ_ZERO of DELETE.
//
// Additionally, in case the state operation is a read. We enforce that the
// old and the new side of the operation are equal. On top of that, the
// storage roots of the accumulator statement are also enforced to be equal.
func (ss *StateSummary) csAccountOrStorageValZeroized(comp *wizard.CompiledIOP) {

	var (
		oldAccountMustBeZero = sym.Mul(
			sym.Sub(1, ss.IsStorage),
			sym.Add(ss.AccumulatorStatement.IsReadZero, ss.AccumulatorStatement.IsInsert),
		)

		newAccountMustBeZero = sym.Mul(
			sym.Sub(1, ss.IsStorage),
			sym.Add(ss.AccumulatorStatement.IsReadZero, ss.AccumulatorStatement.IsDelete),
		)

		oldAndNewAccountMustBeEqual = sym.Mul(
			sym.Sub(1, ss.IsStorage),
			ss.AccumulatorStatement.IsReadNonZero,
		)

		oldStorageValueMustBeZero = sym.Mul(
			ss.IsStorage,
			sym.Add(ss.AccumulatorStatement.IsReadZero, ss.AccumulatorStatement.IsInsert),
		)

		newStorageValueMustBeZero = sym.Mul(
			ss.IsStorage,
			sym.Add(ss.AccumulatorStatement.IsReadZero, ss.AccumulatorStatement.IsDelete),
		)

		oldAndNewStorageMustBeEqual = sym.Mul(
			ss.IsStorage,
			ss.AccumulatorStatement.IsReadNonZero,
		)

		oldAccountColumns = []ifaces.Column{
			ss.InitialAccount.Nonce,
			ss.InitialAccount.Balance,
			ss.InitialAccount.MiMCCodeHash,
			ss.InitialAccount.CodeSize,
			ss.InitialAccount.StorageRoot,
			ss.InitialAccount.KeccakCodeHash.Hi,
			ss.InitialAccount.KeccakCodeHash.Lo,
		}

		newAccountColumns = []ifaces.Column{
			ss.FinalAccount.Nonce,
			ss.FinalAccount.Balance,
			ss.FinalAccount.MiMCCodeHash,
			ss.FinalAccount.CodeSize,
			ss.FinalAccount.StorageRoot,
			ss.FinalAccount.KeccakCodeHash.Hi,
			ss.FinalAccount.KeccakCodeHash.Lo,
		}

		oldStorageColumns = []ifaces.Column{
			ss.StoragePeek.OldValue.Hi,
			ss.StoragePeek.OldValue.Lo,
		}

		newStorageColumns = []ifaces.Column{
			ss.StoragePeek.NewValue.Hi,
			ss.StoragePeek.NewValue.Lo,
		}
	)

	for i := range oldAccountColumns {

		comp.InsertGlobal(
			0,
			ifaces.QueryIDf("STATE_SUMMARY_OLD_ACCOUNT_ZEROIZED_COL_#%v", i),
			sym.Mul(
				oldAccountMustBeZero,
				oldAccountColumns[i],
			),
		)

		comp.InsertGlobal(
			0,
			ifaces.QueryIDf("STATE_SUMMARY_NEW_ACCOUNT_ZEROIZED_COL_#%v", i),
			sym.Mul(
				newAccountMustBeZero,
				newAccountColumns[i],
			),
		)

		comp.InsertGlobal(
			0,
			ifaces.QueryIDf("STATE_SUMMARY_OLD_AND_NEW_ACCOUNT_ARE_EQUALS_%v", i),
			sym.Mul(
				oldAndNewAccountMustBeEqual,
				sym.Sub(oldAccountColumns[i], newAccountColumns[i]),
			),
		)
	}

	for i := range oldStorageColumns {

		comp.InsertGlobal(
			0,
			ifaces.QueryIDf("STATE_SUMMARY_OLD_STORAGE_VAL_ZEROIZED_COL_#%v", i),
			sym.Mul(
				oldStorageValueMustBeZero,
				oldStorageColumns[i],
			),
		)

		comp.InsertGlobal(
			0,
			ifaces.QueryIDf("STATE_SUMMARY_NEW_STORAGE_VAL_ZEROIZED_COL_#%v", i),
			sym.Mul(
				newStorageValueMustBeZero,
				newStorageColumns[i],
			),
		)

		comp.InsertGlobal(
			0,
			ifaces.QueryIDf("STATE_SUMMARY_OLD_AND_NEW_STORAGE_VAL_ARE_EQUALS_%v", i),
			sym.Mul(
				oldAndNewStorageMustBeEqual,
				sym.Sub(oldStorageColumns[i], newStorageColumns[i]),
			),
		)
	}

	comp.InsertGlobal(
		0,
		ifaces.QueryIDf("STATE_SUMMARY_ROOT_STAYS_CONSTANT_WHEN_READING"),
		sym.Mul(
			sym.Add(ss.AccumulatorStatement.IsReadZero, ss.AccumulatorStatement.IsReadNonZero),
			sym.Sub(ss.AccumulatorStatement.InitialRoot, ss.AccumulatorStatement.FinalRoot),
		),
	)

}

func ternary(cond, if1, if0 any) *sym.Expression {
	return sym.Add(
		sym.Mul(sym.Sub(1, cond), if0),
		sym.Mul(cond, if1),
	)
}

func isZeroWhenInactive(comp *wizard.CompiledIOP, c, isActive ifaces.Column) {
	comp.InsertGlobal(
		0,
		ifaces.QueryIDf("%v_IS_ZERO_WHEN_INACTIVE", c.GetColID()),
		sym.Sub(c, sym.Mul(c, isActive)),
	)
}

func deltaOf(a ifaces.Column) *sym.Expression {
	return sym.Sub(a, column.Shift(a, -1))
}

func mustBeBinary(comp *wizard.CompiledIOP, c ifaces.Column) {
	comp.InsertGlobal(
		0,
		ifaces.QueryIDf("%v_MUST_BE_BINARY", c.GetColID()),
		sym.Mul(c, sym.Sub(c, 1)),
	)
}
