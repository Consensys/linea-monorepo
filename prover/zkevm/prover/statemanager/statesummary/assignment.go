package statesummary

import (
	"io"
	"sync"

	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/common"

	"github.com/consensys/linea-monorepo/prover/backend/execution/statemanager"
	"github.com/consensys/linea-monorepo/prover/crypto/mimc"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils/types"
)

// stateSummaryAssignmentBuilder is the struct holding the logic for assigning
// the columns.
type stateSummaryAssignmentBuilder struct {
	StateSummary                *Module
	isActive                    *common.VectorBuilder
	isEndOfAccountSegment       *common.VectorBuilder
	isBeginningOfAccountSegment *common.VectorBuilder
	isInitialDeployment         *common.VectorBuilder
	isFinalDeployment           *common.VectorBuilder
	IsDeleteSegment             *common.VectorBuilder
	isStorage                   *common.VectorBuilder
	batchNumber                 *common.VectorBuilder
	worldStateRoot              [common.NbLimbU256]*common.VectorBuilder
	account                     accountPeekAssignmentBuilder
	storage                     storagePeekAssignmentBuilder
	accumulatorStatement        AccumulatorStatementAssignmentBuilder
	arithmetizationStorage      *ArithmetizationStorageParser
}

// Assign assigns all the columns of the StateSummary module.
func (ss *Module) Assign(run *wizard.ProverRuntime, traces [][]statemanager.DecodedTrace) {

	assignmentBuilder := newStateSummaryAssignmentBuilder(ss, run)

	if ss.ArithmetizationLink != nil {
		assignmentBuilder.arithmetizationStorage.Process()
	}

	for batchNumber, ts := range traces {
		// +1 is to harmonize with the HUB block numbering, which starts from 1
		assignmentBuilder.pushBlockTraces(batchNumber+1, ts)
	}

	assignmentBuilder.finalize(run)

	if ss.ArithmetizationLink != nil {
		ss.assignArithmetizationLink(run)
	}
}

// accountSegmentWitness represents a collection of traces representing
// an account segment. It can have either one or two subSegments. Any other
// value is invalid.
type accountSegmentWitness []accountSubSegmentWitness

// accountSubSegment hold a sequence of shomei traces relating to an account
// sub-segment.
type accountSubSegmentWitness struct {
	// worldStateTrace is the shomei trace relating to the account we are
	// accessing
	worldStateTrace statemanager.DecodedTrace
	// storageTraces is a list of shomei traces relating to the storage of the
	// account we are currently accessing for the same deployment.
	storageTraces []statemanager.DecodedTrace
}

// newStateSummaryAssignmentBuilder constructs a new
// stateSummaryAssignmentBuilder and initializes all the involved column builders.
func newStateSummaryAssignmentBuilder(ss *Module, run *wizard.ProverRuntime) *stateSummaryAssignmentBuilder {
	res := &stateSummaryAssignmentBuilder{

		StateSummary: ss,

		isActive:                    common.NewVectorBuilder(ss.IsActive),
		isEndOfAccountSegment:       common.NewVectorBuilder(ss.IsEndOfAccountSegment),
		isBeginningOfAccountSegment: common.NewVectorBuilder(ss.IsBeginningOfAccountSegment),
		isInitialDeployment:         common.NewVectorBuilder(ss.IsInitialDeployment),
		isFinalDeployment:           common.NewVectorBuilder(ss.IsFinalDeployment),
		IsDeleteSegment:             common.NewVectorBuilder(ss.IsDeleteSegment),
		isStorage:                   common.NewVectorBuilder(ss.IsStorage),
		batchNumber:                 common.NewVectorBuilder(ss.BatchNumber),
		storage:                     newStoragePeekAssignmentBuilder(&ss.Storage),
		account:                     newAccountPeekAssignmentBuilder(&ss.Account),
		accumulatorStatement:        newAccumulatorStatementAssignmentBuilder(&ss.AccumulatorStatement),
		arithmetizationStorage:      newArithmetizationStorageParser(ss, run),
	}

	for i := range common.NbLimbU256 {
		res.worldStateRoot[i] = common.NewVectorBuilder(ss.WorldStateRoot[i])
	}

	return res
}

// pushBlockTraces pushes a list of rows corresponding to a block of execution
// on L2.
func (ss *stateSummaryAssignmentBuilder) pushBlockTraces(batchNumber int, traces []statemanager.DecodedTrace) {

	var (
		subSegment = accountSubSegmentWitness{}
		curSegment = accountSegmentWitness{}
		curAddress = types.EthAddress{}
		err        error
	)

	for i, trace := range traces {

		if i == 0 {
			curAddress, err = trace.GetRelatedAccount()
			if err != nil {
				panic(err)
			}
		}

		newAddress, err := trace.GetRelatedAccount()
		if err != nil {
			panic(err)
		}

		if newAddress != curAddress {
			// This addresses the case where the segment is Read|ReadZero. In
			// that situation, the account trace is at the beginning of the
			// segment. When that happens, we want to be sure that the
			// storage rows and the account segment arise in the same position.
			if actualUnskippedLength(subSegment.storageTraces) > 0 {
				curSegment[len(curSegment)-1].storageTraces = subSegment.storageTraces
				subSegment = accountSubSegmentWitness{}
			}

			ss.pushAccountSegment(batchNumber, curSegment)
			curSegment = accountSegmentWitness{}
			curAddress = newAddress
		}

		// This tests whether trace is a world-state trace
		if trace.Location == statemanager.WS_LOCATION {
			subSegment.worldStateTrace = trace
			curSegment = append(curSegment, subSegment)
			subSegment = accountSubSegmentWitness{}
			continue
		}

		subSegment.storageTraces = append(subSegment.storageTraces, trace)
	}

	if actualUnskippedLength(subSegment.storageTraces) > 0 {
		curSegment[len(curSegment)-1].storageTraces = subSegment.storageTraces
	}

	ss.pushAccountSegment(batchNumber, curSegment)
}

// pushAccountSegment pushes a list of rows corresponding to an account within
// a block.
func (ss *stateSummaryAssignmentBuilder) pushAccountSegment(batchNumber int, segment accountSegmentWitness) {

	for segID, seg := range segment {

		var (
			accountAddress, errAddr      = seg.worldStateTrace.GetRelatedAccount()
			initialAccount, finalAccount = getOldAndNewAccount(seg.worldStateTrace.Underlying)
			initWsRoot, finalWsRoot      = getOldAndNewTopRoot(seg.worldStateTrace.Underlying)
			_, isDeleteSegment           = seg.worldStateTrace.Underlying.(statemanager.DeletionTraceWS)
		)

		accountAddressLimbs := common.SplitBytes(accountAddress[:])

		if errAddr != nil {
			panic("could not get the account address")
		}

		noOfSkippedStorageTraces := 0
		for i := range seg.storageTraces {

			var (
				stoTrace         = seg.storageTraces[i].Underlying
				oldRoot, newRoot = getOldAndNewTopRoot(stoTrace)
			)

			_ = oldRoot // TODO golangci-lint thinks oldRoot is otherwise unused, even though it's clearly used in the switch case

			if !seg.storageTraces[i].IsSkipped {
				// the storage trace is to be kept, and not skipped
				ss.batchNumber.PushInt(batchNumber)

				for j := range common.NbLimbEthAddress {
					ss.account.address[j].PushBytes(common.LeftPadToFrBytes(accountAddressLimbs[j]))
				}

				ss.isInitialDeployment.PushBoolean(segID == 0)
				ss.isFinalDeployment.PushBoolean(segID == len(segment)-1)
				ss.IsDeleteSegment.PushBoolean(isDeleteSegment)
				ss.isActive.PushOne()
				ss.isStorage.PushOne()
				ss.isEndOfAccountSegment.PushZero()
				ss.isBeginningOfAccountSegment.PushBoolean(
					segID == 0 && i == firstUnskippedIndex(seg.storageTraces),
				)
				ss.account.initial.pushAll(initialAccount)
				ss.account.final.pushOverrideStorageRoot(finalAccount, common.SplitBytes(newRoot[:]))

				for j, initWsRootLimbs := range common.SplitBytes(initWsRoot[:]) {
					ss.worldStateRoot[j].PushBytes(common.LeftPadToFrBytes(initWsRootLimbs))
				}

				oldRootLimbs := common.SplitBytes(oldRoot[:])
				newRootLimbs := common.SplitBytes(newRoot[:])

				switch t := stoTrace.(type) {
				case statemanager.ReadZeroTraceST:
					if isDeleteSegment {
						/*
							Special case: the Shomei compactification process automatically sets storage values to zero if the account later gets deleted
							which might not be the case in the arithmetization
							in this particular case, for the consistency lookups to work,
							we fetch and use the last corresponding storage value/block from the arithmetization columns using
							an ArithmetizationStorageParser
						*/
						x := *(&field.Element{}).SetBytes(accountAddress[:])
						keysAndBlock := KeysAndBlock{
							address:    x.Bytes(),
							storageKey: t.Key,
							block:      batchNumber,
						}
						arithStorage := ss.arithmetizationStorage.Values[keysAndBlock]

						ss.storage.push(t.Key, types.FullBytes32{}, arithStorage)

						keyH := hash(t.Key)
						keyHashBytesLimbs := common.SplitBytes(keyH[:])

						ss.accumulatorStatement.PushReadZero(oldRootLimbs, keyHashBytesLimbs)
					} else {
						keyH := hash(t.Key)
						keyHashBytesLimbs := common.SplitBytes(keyH[:])

						ss.storage.pushOnlyKey(t.Key)
						ss.accumulatorStatement.PushReadZero(oldRootLimbs, keyHashBytesLimbs)
					}
				case statemanager.ReadNonZeroTraceST:
					if isDeleteSegment {
						/*
							Special case, same motivation and fix as in the case of ReadZeroTraceST
						*/
						x := *(&field.Element{}).SetBytes(accountAddress[:])
						keysAndBlock := KeysAndBlock{
							address:    x.Bytes(),
							storageKey: t.Key,
							block:      batchNumber,
						}
						arithStorage := ss.arithmetizationStorage.Values[keysAndBlock]

						keyH := hash(t.Key)
						keyHashBytesLimbs := common.SplitBytes(keyH[:])

						valueH := hash(t.Value)
						valueHashBytesLimbs := common.SplitBytes(valueH[:])

						ss.storage.push(t.Key, t.Value, arithStorage)
						ss.accumulatorStatement.PushReadNonZero(oldRootLimbs, keyHashBytesLimbs, valueHashBytesLimbs)

					} else {
						keyH := hash(t.Key)
						keyHashBytesLimbs := common.SplitBytes(keyH[:])

						valueH := hash(t.Value)
						valueHashBytesLimbs := common.SplitBytes(valueH[:])

						ss.storage.push(t.Key, t.Value, t.Value)
						ss.accumulatorStatement.PushReadNonZero(oldRootLimbs, keyHashBytesLimbs, valueHashBytesLimbs)
					}

				case statemanager.InsertionTraceST:
					keyH := hash(t.Key)
					keyHashBytesLimbs := common.SplitBytes(keyH[:])

					valueH := hash(t.Val)
					valueHashBytesLimbs := common.SplitBytes(valueH[:])

					ss.storage.pushOnlyNew(t.Key, t.Val)
					ss.accumulatorStatement.PushInsert(oldRootLimbs, newRootLimbs, keyHashBytesLimbs, valueHashBytesLimbs)

				case statemanager.UpdateTraceST:
					keyH := hash(t.Key)
					keyHashBytesLimbs := common.SplitBytes(keyH[:])

					oldValueH := hash(t.OldValue)
					oldValueHashBytesLimbs := common.SplitBytes(oldValueH[:])

					newValueH := hash(t.NewValue)
					newValueHashBytesLimbs := common.SplitBytes(newValueH[:])

					ss.storage.push(t.Key, t.OldValue, t.NewValue)
					ss.accumulatorStatement.PushUpdate(oldRootLimbs, newRootLimbs, keyHashBytesLimbs, oldValueHashBytesLimbs, newValueHashBytesLimbs)

				case statemanager.DeletionTraceST:
					keyH := hash(t.Key)
					keyHashBytesLimbs := common.SplitBytes(keyH[:])

					delValueH := hash(t.DeletedValue)
					delValueHashBytesLimbs := common.SplitBytes(delValueH[:])

					ss.storage.pushOnlyOld(t.Key, t.DeletedValue)
					ss.accumulatorStatement.PushDelete(oldRootLimbs, newRootLimbs, keyHashBytesLimbs, delValueHashBytesLimbs)
				default:
					panic("unknown trace type")
				}
			} else {
				// the storage trace is skipped
				noOfSkippedStorageTraces++
			}
		}

		ss.batchNumber.PushInt(batchNumber)

		for j := range common.NbLimbEthAddress {
			ss.account.address[j].PushBytes(common.LeftPadToFrBytes(accountAddressLimbs[j]))
		}

		ss.isInitialDeployment.PushBoolean(segID == 0)
		ss.isFinalDeployment.PushBoolean(segID == len(segment)-1)
		ss.IsDeleteSegment.PushBoolean(isDeleteSegment)
		ss.isActive.PushOne()
		ss.isStorage.PushZero()
		ss.isEndOfAccountSegment.PushBoolean(segID == len(segment)-1)
		ss.isBeginningOfAccountSegment.PushBoolean(segID == 0 && actualUnskippedLength(seg.storageTraces) == 0)
		ss.account.initial.pushAll(initialAccount)
		ss.account.final.pushAll(finalAccount)

		for j, finalWsRootLimbs := range common.SplitBytes(finalWsRoot[:]) {
			ss.worldStateRoot[j].PushBytes(common.LeftPadToFrBytes(finalWsRootLimbs))
		}

		ss.storage.pushAllZeroes()

		initWsRootLimbs := common.SplitBytes(initWsRoot[:])
		finalWsRootLimbs := common.SplitBytes(finalWsRoot[:])

		switch t := seg.worldStateTrace.Underlying.(type) {
		case statemanager.ReadZeroTraceWS:
			keyH := hash(t.Key)
			keyHashBytesLimbs := common.SplitBytes(keyH[:])

			ss.accumulatorStatement.PushReadZero(initWsRootLimbs, keyHashBytesLimbs)
		case statemanager.ReadNonZeroTraceWS:
			keyH := hash(t.Key)
			keyHashBytesLimbs := common.SplitBytes(keyH[:])

			valueH := hash(t.Value)
			valueHashBytesLimbs := common.SplitBytes(valueH[:])

			ss.accumulatorStatement.PushReadNonZero(initWsRootLimbs, keyHashBytesLimbs, valueHashBytesLimbs)
		case statemanager.InsertionTraceWS:
			keyH := hash(t.Key)
			keyHashBytesLimbs := common.SplitBytes(keyH[:])

			valueH := hash(t.Val)
			valueHashBytesLimbs := common.SplitBytes(valueH[:])

			ss.accumulatorStatement.PushInsert(initWsRootLimbs, finalWsRootLimbs, keyHashBytesLimbs, valueHashBytesLimbs)
		case statemanager.UpdateTraceWS:
			keyH := hash(t.Key)
			keyHashBytesLimbs := common.SplitBytes(keyH[:])

			oldValueH := hash(t.OldValue)
			oldValueHashBytesLimbs := common.SplitBytes(oldValueH[:])

			newValueH := hash(t.NewValue)
			newValueHashBytesLimbs := common.SplitBytes(newValueH[:])

			ss.accumulatorStatement.PushUpdate(initWsRootLimbs, finalWsRootLimbs, keyHashBytesLimbs, oldValueHashBytesLimbs, newValueHashBytesLimbs)
		case statemanager.DeletionTraceWS:
			keyH := hash(t.Key)
			keyHashBytesLimbs := common.SplitBytes(keyH[:])

			deletedValueH := hash(t.DeletedValue)
			deletedValueHashBytesLimbs := common.SplitBytes(deletedValueH[:])

			ss.accumulatorStatement.PushDelete(initWsRootLimbs, finalWsRootLimbs, keyHashBytesLimbs, deletedValueHashBytesLimbs)
		default:
			panic("unknown trace type")
		}
	}
}

// finalize pads all the columns and ensure they are all assigned within `run`
func (ss *stateSummaryAssignmentBuilder) finalize(run *wizard.ProverRuntime) {

	ss.isActive.PadAndAssign(run)
	ss.isEndOfAccountSegment.PadAndAssign(run)
	ss.isBeginningOfAccountSegment.PadAndAssign(run)
	ss.isInitialDeployment.PadAndAssign(run)
	ss.isFinalDeployment.PadAndAssign(run)
	ss.IsDeleteSegment.PadAndAssign(run)
	ss.isStorage.PadAndAssign(run)
	ss.batchNumber.PadAndAssign(run)

	for i := range common.NbLimbU256 {
		ss.worldStateRoot[i].PadAndAssign(run)
	}

	ss.account.initial.PadAndAssign(run)
	ss.account.final.PadAndAssign(run)

	for i := range common.NbLimbEthAddress {
		ss.account.address[i].PadAndAssign(run)
	}

	ss.storage.padAssign(run)
	ss.accumulatorStatement.PadAndAssign(run)

	runConcurrent := func(pas []wizard.ProverAction) {
		wg := &sync.WaitGroup{}
		for _, pa := range pas {
			wg.Add(1)
			go func(pa wizard.ProverAction) {
				pa.Run(run)
				wg.Done()
			}(pa)
		}

		wg.Wait()
	}

	summaryAccountActions := [][]wizard.ProverAction{
		ss.StateSummary.Account.Initial.CptHasEmptyCodeHash[:],
		ss.StateSummary.Account.Final.CptHasEmptyCodeHash[:],
		{
			ss.StateSummary.Account.ComputeAddressHash,
			ss.StateSummary.Account.ComputeHashFinal,
			ss.StateSummary.Account.ComputeHashInitial,
			ss.StateSummary.Storage.ComputeKeyHash,
			ss.StateSummary.Storage.ComputeOldValueHash,
			ss.StateSummary.Storage.ComputeNewValueHash,
			ss.StateSummary.AccumulatorStatement.CptSameTypeAsBefore,
		},
	}

	var accountActions []wizard.ProverAction
	for _, action := range summaryAccountActions {
		accountActions = append(accountActions, action...)
	}

	runConcurrent(accountActions)

	runConcurrent(
		append(
			ss.StateSummary.Account.ComputeAddressLimbs[:],
			ss.StateSummary.Storage.ComputeKeyLimbs[:]...,
		),
	)

	summaryStorageActions := [][]wizard.ProverAction{
		ss.StateSummary.Account.ComputeInitialAndFinalAreSame[:],
		{ss.StateSummary.Account.ComputeAddressComparison},
		ss.StateSummary.Storage.ComputeOldValueIsZero[:],
		ss.StateSummary.Storage.ComputeNewValueIsZero[:],
		{ss.StateSummary.Storage.ComputeKeyIncreased},
		ss.StateSummary.Storage.ComputeOldAndNewValuesAreEqual[:],
	}

	var storageActions []wizard.ProverAction
	for _, action := range summaryStorageActions {
		storageActions = append(storageActions, action...)
	}

	runConcurrent(storageActions)

	runConcurrent(append(
		ss.StateSummary.AccumulatorStatement.ComputeInitialAndFinalHValEqual[:],
		ss.StateSummary.AccumulatorStatement.ComputeFinalHValIsZero[:]...,
	))

	//for i := range common.NbLimbU256 {
	//	v := ss.StateSummary.Account.HashFinal[i].GetColAssignmentAt(run, 0)
	//	print(v.Text(16), "")
	//}
	//println()
}

// getOldAndNewAccount traces a world-state trace and return the old and the
// new value of the account.
func getOldAndNewAccount(trace any) (old, new types.Account) {
	switch wst := trace.(type) {
	case statemanager.ReadNonZeroTraceWS:
		return wst.Value, wst.Value
	case statemanager.ReadZeroTraceWS:
		return types.Account{}, types.Account{}
	case statemanager.InsertionTraceWS:
		return types.Account{}, wst.Val
	case statemanager.UpdateTraceWS:
		return wst.OldValue, wst.NewValue
	case statemanager.DeletionTraceWS:
		return wst.DeletedValue, types.Account{}
	default:
		panic("unknown trace")
	}
}

// getOldAndNewTopRoot returns the accumulator root transition for a shomei
// trace.
func getOldAndNewTopRoot(trace any) (old, new types.Bytes32) {

	getTopRoot := func(subRoot types.Bytes32, nextFreeNode int64) types.Bytes32 {
		hasher := mimc.NewMiMC()
		types.WriteInt64On32Bytes(hasher, nextFreeNode)
		subRoot.WriteTo(hasher)
		b32 := hasher.Sum(nil)
		return types.AsBytes32(b32)
	}

	switch wst := trace.(type) {
	case statemanager.ReadNonZeroTraceWS:
		res := getTopRoot(wst.SubRoot, int64(wst.NextFreeNode))
		return res, res
	case statemanager.ReadZeroTraceWS:
		res := getTopRoot(wst.SubRoot, int64(wst.NextFreeNode))
		return res, res
	case statemanager.InsertionTraceWS:
		var (
			old = getTopRoot(wst.OldSubRoot, int64(wst.NewNextFreeNode-1))
			new = getTopRoot(wst.NewSubRoot, int64(wst.NewNextFreeNode))
		)
		return old, new
	case statemanager.UpdateTraceWS:
		var (
			old = getTopRoot(wst.OldSubRoot, int64(wst.NewNextFreeNode))
			new = getTopRoot(wst.NewSubRoot, int64(wst.NewNextFreeNode))
		)
		return old, new
	case statemanager.DeletionTraceWS:
		var (
			old = getTopRoot(wst.OldSubRoot, int64(wst.NewNextFreeNode))
			new = getTopRoot(wst.NewSubRoot, int64(wst.NewNextFreeNode))
		)
		return old, new
	case statemanager.ReadNonZeroTraceST:
		res := getTopRoot(wst.SubRoot, int64(wst.NextFreeNode))
		return res, res
	case statemanager.ReadZeroTraceST:
		res := getTopRoot(wst.SubRoot, int64(wst.NextFreeNode))
		return res, res
	case statemanager.InsertionTraceST:
		var (
			old = getTopRoot(wst.OldSubRoot, int64(wst.NewNextFreeNode-1))
			new = getTopRoot(wst.NewSubRoot, int64(wst.NewNextFreeNode))
		)
		return old, new
	case statemanager.UpdateTraceST:
		var (
			old = getTopRoot(wst.OldSubRoot, int64(wst.NewNextFreeNode))
			new = getTopRoot(wst.NewSubRoot, int64(wst.NewNextFreeNode))
		)
		return old, new
	case statemanager.DeletionTraceST:
		var (
			old = getTopRoot(wst.OldSubRoot, int64(wst.NewNextFreeNode))
			new = getTopRoot(wst.NewSubRoot, int64(wst.NewNextFreeNode))
		)
		return old, new
	default:
		panic("unknown trace")
	}
}

func hash(x io.WriterTo) types.Bytes32 {
	hasher := mimc.NewMiMC()
	x.WriteTo(hasher)
	return types.AsBytes32(hasher.Sum(nil))
}

// actualUnskippedLength computes the actual number of traces that form the segments
// meaning it adds up only the unskipped traces
func actualUnskippedLength(traces []statemanager.DecodedTrace) int {
	res := 0
	for _, trace := range traces {
		if !trace.IsSkipped {
			res++
		}
	}
	return res
}

// firstUnskippedIndex returns the index of the first unskipped storage trace.
func firstUnskippedIndex(traces []statemanager.DecodedTrace) int {
	for i, trace := range traces {
		if !trace.IsSkipped {
			return i
		}
	}
	panic("There are no unskipped storage traces, but that is out of Shomei's expected specifications")
}
