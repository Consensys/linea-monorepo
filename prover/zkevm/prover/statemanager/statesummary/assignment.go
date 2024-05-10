package statesummary

import (
	"io"

	"github.com/consensys/zkevm-monorepo/prover/backend/execution/statemanager"
	"github.com/consensys/zkevm-monorepo/prover/crypto/mimc"
	"github.com/consensys/zkevm-monorepo/prover/maths/field"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
	"github.com/consensys/zkevm-monorepo/prover/utils/parallel"
	"github.com/consensys/zkevm-monorepo/prover/utils/types"
)

// stateSummaryAssignmentBuilder is the struct holding the logic for assigning
// the columns.
type stateSummaryAssignmentBuilder struct {
	accountAddress                   *vectorBuilder
	accountAddressShiftInverseOrZero *vectorBuilder
	accountAddressShiftIsZero        *vectorBuilder
	isActive                         *vectorBuilder
	isEndOfAccountSegment            *vectorBuilder
	isBeginningOfAccountSegment      *vectorBuilder
	isInitialDeployment              *vectorBuilder
	isFinalDeployment                *vectorBuilder
	isStorage                        *vectorBuilder
	batchNumber                      *vectorBuilder
	worldStateRoot                   *vectorBuilder
	initialAccount, finalAccount     accountPeekAssignmentBuilder
	storagePeek                      storagePeekAssignmentBuilder
	accountHashing                   accountHashingAssignmentBuilder
	storageHashing                   storageHashingAssignmentBuilder
	accumulatorStatement             accumulatorStatementAssignmentBuilder
}

func (ss *StateSummary) Assign(run *wizard.ProverRuntime, traces [][]statemanager.DecodedTrace) {
	assignmentBuilder := newStateSummaryAssignmentBuilder(ss)
	for batchNumber, ts := range traces {
		assignmentBuilder.pushBlockTraces(batchNumber, ts)
	}
	assignmentBuilder.finalize(run)
}

// accountSegmentWitness represents a collection of traces representing
// an account segment. It can have either one or two subSegments. Any other
// value is invalid.
type accountSegmentWitness []accountSubSegmentWitness

// accountSubSegment hold a sequence of shomei traces relating to an account
// sub-segment.
type accountSubSegmentWitness struct {
	worldStateTrace statemanager.DecodedTrace
	storageTraces   []statemanager.DecodedTrace
}

// newStateSummaryAssignmentBuilder constructs a new
// stateSummaryAssignmentBuilder and initializes all the involved column builders.
func newStateSummaryAssignmentBuilder(ss *StateSummary) *stateSummaryAssignmentBuilder {
	return &stateSummaryAssignmentBuilder{
		accountAddress:                   newVectorBuilder(ss.AccountAddress),
		accountAddressShiftInverseOrZero: newVectorBuilder(ss.LookBackDeltaAddress.inverseOrZero),
		accountAddressShiftIsZero:        newVectorBuilder(ss.LookBackDeltaAddress.IsZero),
		isActive:                         newVectorBuilder(ss.IsActive),
		isEndOfAccountSegment:            newVectorBuilder(ss.IsEndOfAccountSegment),
		isBeginningOfAccountSegment:      newVectorBuilder(ss.IsBeginningOfAccountSegment),
		isInitialDeployment:              newVectorBuilder(ss.IsInitialDeployment),
		isFinalDeployment:                newVectorBuilder(ss.IsFinalDeployment),
		isStorage:                        newVectorBuilder(ss.IsStorage),
		batchNumber:                      newVectorBuilder(ss.BatchNumber),
		worldStateRoot:                   newVectorBuilder(ss.WorldStateRoot),
		initialAccount:                   newAccountPeekAssignmentBuilder(&ss.InitialAccount),
		finalAccount:                     newAccountPeekAssignmentBuilder(&ss.FinalAccount),
		storagePeek:                      newStoragePeekAssignmentBuilder(&ss.StoragePeek),
		accountHashing:                   newAccountHashingAssignmentBuilder(ss.AccountHashing),
		storageHashing:                   newStorageHashingAssignmentBuilder(ss.StorageHashing),
		accumulatorStatement:             newAccumulatorStatementAssignmentBuilder(&ss.AccumulatorStatement),
	}
}

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

		// This tests whether trace is a world-state trace
		if trace.Location == statemanager.WS_LOCATION {
			subSegment.worldStateTrace = trace
			curSegment = append(curSegment, subSegment)
			subSegment = accountSubSegmentWitness{}
			continue
		}

		newAddress, err := trace.GetRelatedAccount()
		if err != nil {
			panic(err)
		}

		if newAddress != curAddress {
			ss.pushAccountSegment(batchNumber, curSegment)
			curSegment = accountSegmentWitness{}
		}

		subSegment.storageTraces = append(subSegment.storageTraces, trace)
	}

	ss.pushAccountSegment(batchNumber, curSegment)
}

func (ss *stateSummaryAssignmentBuilder) pushAccountSegment(batchNumber int, segment accountSegmentWitness) {

	for segID, seg := range segment {

		var (
			accountAddress, errAddr      = seg.worldStateTrace.GetRelatedAccount()
			initialAccount, finalAccount = getOldAndNewAccount(seg.worldStateTrace.Underlying)
			initWsRoot, finalWsRoot      = getOldAndNewTopRoot(seg.worldStateTrace.Underlying)
		)

		if errAddr != nil {
			panic("could not get the account address")
		}

		for i := range seg.storageTraces {

			var (
				stoTrace         = seg.storageTraces[i].Underlying
				oldRoot, newRoot = getOldAndNewTopRoot(stoTrace)
			)

			ss.batchNumber.PushInt(batchNumber)
			ss.accountAddress.PushAddr(accountAddress)
			ss.isInitialDeployment.PushBoolean(segID == 0)
			ss.isFinalDeployment.PushBoolean(segID == len(segment)-1)
			ss.isActive.PushOne()
			ss.isStorage.PushOne()
			ss.isEndOfAccountSegment.PushZero()
			ss.isBeginningOfAccountSegment.PushBoolean(segID == 0 && i == 0)
			ss.initialAccount.pushAll(initialAccount)
			ss.finalAccount.pushOverrideStorageRoot(initialAccount, newRoot)
			ss.worldStateRoot.PushBytes32(initWsRoot)

			switch t := stoTrace.(type) {
			case statemanager.ReadZeroTraceST:
				ss.storagePeek.pushOnlyKey(t.Key)
				ss.accumulatorStatement.pushReadZero(oldRoot, hash(t.Key))
			case statemanager.ReadNonZeroTraceST:
				ss.storagePeek.push(t.Key, t.Value, t.Value)
				ss.accumulatorStatement.pushReadNonZero(oldRoot, hash(t.Key), hash(t.Value))
			case statemanager.InsertionTraceST:
				ss.storagePeek.pushOnlyNew(t.Key, t.Val)
				ss.accumulatorStatement.pushInsert(oldRoot, newRoot, hash(t.Key), hash(t.Val))
			case statemanager.UpdateTraceST:
				ss.storagePeek.push(t.Key, t.OldValue, t.NewValue)
				ss.accumulatorStatement.pushUpdate(oldRoot, newRoot, hash(t.Key), hash(t.OldValue), hash(t.NewValue))
			case statemanager.DeletionTraceST:
				ss.storagePeek.pushOnlyOld(t.Key, t.DeletedValue)
				ss.accumulatorStatement.pushDelete(oldRoot, newRoot, hash(t.Key), hash(t.DeletedValue))
			default:
				panic("unknown trace type")
			}
		}

		ss.batchNumber.PushInt(batchNumber)
		ss.accountAddress.PushAddr(accountAddress)
		ss.isInitialDeployment.PushBoolean(segID == 0)
		ss.isFinalDeployment.PushBoolean(segID == len(segment)-1)
		ss.isActive.PushOne()
		ss.isStorage.PushZero()
		ss.isEndOfAccountSegment.PushBoolean(segID == len(segment)-1)
		ss.isBeginningOfAccountSegment.PushBoolean(segID == 0 && len(seg.storageTraces) == 0)
		ss.initialAccount.pushAll(initialAccount)
		ss.finalAccount.pushAll(finalAccount)
		ss.worldStateRoot.PushBytes32(finalWsRoot)
		ss.storagePeek.pushAllZeroes()

		switch t := seg.worldStateTrace.Underlying.(type) {
		case statemanager.ReadZeroTraceWS:
			ss.accumulatorStatement.pushReadZero(initWsRoot, hash(t.Key))
		case statemanager.ReadNonZeroTraceWS:
			ss.accumulatorStatement.pushReadNonZero(initWsRoot, hash(t.Key), hash(t.Value))
		case statemanager.InsertionTraceWS:
			ss.accumulatorStatement.pushInsert(initWsRoot, finalWsRoot, hash(t.Key), hash(t.Val))
		case statemanager.UpdateTraceWS:
			ss.accumulatorStatement.pushUpdate(initWsRoot, finalWsRoot, hash(t.Key), hash(t.OldValue), hash(t.NewValue))
		case statemanager.DeletionTraceWS:
			ss.accumulatorStatement.pushDelete(initWsRoot, finalWsRoot, hash(t.Key), hash(t.DeletedValue))
		default:
			panic("unknown trace type")
		}
	}
}

func (ss *stateSummaryAssignmentBuilder) finalize(run *wizard.ProverRuntime) {

	var (
		numRow = len(ss.accountAddress.slice)
	)
	ss.accountAddressShiftInverseOrZero.Resize(numRow)
	ss.accountAddressShiftIsZero.Resize(numRow)
	ss.accountHashing.Resize(numRow)
	ss.storageHashing.Resize(numRow)

	parallel.Execute(numRow, func(start, stop int) {

		ss.accountHashing.assignChunk(ss.accountAddress, &ss.initialAccount, &ss.finalAccount, start, stop)
		ss.storageHashing.assignChunk(&ss.storagePeek, start, stop)

		for i := start; i < stop; i++ {
			if i == 0 {
				ss.accountAddressShiftInverseOrZero.slice[i] = field.One()
				ss.accountAddressShiftIsZero.slice[i] = field.Zero()
				continue
			}

			if i > 0 {
				if ss.accountAddress.slice[i] == ss.accountAddress.slice[i-1] {
					ss.accountAddressShiftInverseOrZero.slice[i] = field.One()
					ss.accountAddressShiftIsZero.slice[i] = field.One()
					continue
				}

				ss.accountAddressShiftInverseOrZero.slice[i].Sub(&ss.accountAddress.slice[i], &ss.accountAddress.slice[i-1])
				ss.accountAddressShiftInverseOrZero.slice[i].Inverse(&ss.accountAddressShiftInverseOrZero.slice[i])
				ss.accountAddressShiftIsZero.slice[i] = field.Zero()
			}
		}
	})

	ss.accountAddress.PadAndAssign(run)
	ss.accountAddressShiftInverseOrZero.PadAndAssign(run)
	ss.accountAddressShiftIsZero.PadAndAssign(run)
	ss.isActive.PadAndAssign(run)
	ss.isEndOfAccountSegment.PadAndAssign(run)
	ss.isBeginningOfAccountSegment.PadAndAssign(run)
	ss.isInitialDeployment.PadAndAssign(run)
	ss.isFinalDeployment.PadAndAssign(run)
	ss.isStorage.PadAndAssign(run)
	ss.batchNumber.PadAndAssign(run)
	ss.worldStateRoot.PadAndAssign(run)
	ss.initialAccount.PadAndAssign(run)
	ss.finalAccount.PadAndAssign(run)
	ss.storagePeek.padAssign(run)
	ss.accountHashing.PadAndAssign(run)
	ss.storageHashing.PadAndAssign(run)
	ss.accumulatorStatement.padAndAssign(run)

}

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
