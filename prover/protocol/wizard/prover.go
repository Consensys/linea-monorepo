package wizard

import (
	"encoding/csv"
	"fmt"
	"os"
	"path"

	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/gnark-crypto/field/koalabear/poseidon2"
	"github.com/consensys/gnark-crypto/hash"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"

	"strconv"
	"sync"
	"time"

	"github.com/consensys/linea-monorepo/prover/config"
	"github.com/consensys/linea-monorepo/prover/crypto/fiatshamir"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/distributed/pragmas"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/zk"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/collection"
	"github.com/consensys/linea-monorepo/prover/utils/profiling"
	"github.com/sirupsen/logrus"
)

// ProverRuntimeOption represents options that can be provided to
// methods of the [wizard.ProverRuntime] struct. These are used to
// enable/disable some of the optimization that are done by the prover
// internally.
type ProverRuntimeOption uint64

const (
	// DisableAssignmentSizeReduction is used to disable the
	// the routine that tries to reduce the space taken by a column.
	DisableAssignmentSizeReduction ProverRuntimeOption = 1 << iota
)

// This is a compilation check to ensure that the [wizard.ProverRuntime]
// implements the [wizard.Runtime] interface.
// var _ Runtime = &ProverRuntime{}

// MainProverStep represents an operation to be performed by the prover of a
// wizard protocol. It can be provided by the user or by an internal compiled
// to the protocol specification [CompiledIOP] by appending it to the field
// [CompiledIOP.SubProvers].
//
//	CompiledIOP.SubProvers.AppendToInner(round, proverStep)
//
// The MainProverStep function may interact with the prover runtime to resolve
// the values of an already assigned item: ([ifaces.Colssignment], coin,
// [ifaces.QueryParams], ...).
//
// The MainProverStep function that we pass as the `highLevelProver` argument of
// [Prove] function has the particularity that it is allowed to span
// over multiple interaction-rounds between the prover and the verifier. This
// is a behavior that we intend to deprecate and it should not be used by the
// prover as this tends to create convolutions in the runtime of the prover.

type MainProverStep[T zk.Element] func(assi *ProverRuntime[T])

// ProverRuntime collects the assignment of all the items with which the prover
// interacts by the prover of the protocol. This includes the prover's
// messages, items that are computed solely by the prover, the witness but also
// the random coins that are sampled by the verifier. The object is implicitly
// constructed by the [Prove] function and it should not be explicitly
// constructed by the user.
//
// Instead, the user should interact with the prover runtime within a
// [MainProverStep] function that he provides to the CompiledIOP that he is
// building. Example:
//
//	// Function that the user provide to specify his protocol
//	func myDefineFunction(builder wizard.Builder) {
//
//		// Registers a column "A" as a column to commit to
//		a := build.RegisterCommit("A", 16)
//
//		// Potentially add constraints over the column
//		...
//	}
//
//	// The above define function specifies a protocol involving a column
//	// named "A". If we want to concretely run our protocol, we also need
//	// to provide a way to assign concrete values to the witness of the
//	// protocol.
//	func myProverFunction(run wizard.ProverRuntime[T]) {
//		a := smartvector.ForTest(1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16)
//		run.AssignColumn("A", a)
//	}
//
// ProverRuntime also bears the logic to track the current round of interaction
// between the prover and the verifier.
type ProverRuntime[T zk.Element] struct {

	// Spec is the underlying [CompiledIOP] of the underlying protocol the prover
	// is running.
	Spec *CompiledIOP[T]

	// Columns stores all the column's ([ifaces.Column]) witnesses assigned so
	// far by the ProverRuntime. Columns that are assigned using
	// [ProverRuntime.AssignColumn] method are stored there. For most use-cases,
	// it is preferable to use[ifaces.Columns.GetColAssignment] instead of
	// fetching the assignmentdirectly from the ProverRuntime. The reason is
	// that, the column the caller is trying to fetch may be a "derivative
	// column" or another type of special column whose assignment is not directly
	// available within the prover's runtime.
	//
	// Please consider that this field could become a private field.
	Columns collection.Mapping[ifaces.ColID, ifaces.ColAssignment]

	// QueriesParams accumulates all the query parameters of the queries assigned so far. See
	// [ifaces.QueryParams]. The query parameters that are stored there
	// corresponds to the queries stored in [ProverRuntime.Spec.QueriesParams]
	QueriesParams collection.Mapping[ifaces.QueryID, ifaces.QueryParams]

	// Coins stores all the values of all random Coins that are generated internally
	// as the ProverRuntime unfolds the prover steps round after rounds.
	//
	// The user should not directly access this field and fall back to using the
	// dedicated methods [ProverRuntime.GetRandomCoinField] or
	// [ProverRuntime.GetRandomCoinIntegerVec].
	Coins collection.Mapping[coin.Name, interface{}]

	// State serves as an "any-purpose" data-storage for stateful proving. It allows
	// MainProverSteps to persist data that can be accessed in later prover steps
	// without having to store it in a column. For convenience, the user should
	// take care of deleting the entry to free memory when he knows that the
	// field will not be accessed again while proving.
	//
	// The State is used internally by the [github.com/consensys/linea-monorepo/prover/protocol/compiler/vortex] and the
	// [github.com/consensys/linea-monorepo/prover/protocol/compiler/selfrecursion] compilers as a communication channel.
	State collection.Mapping[string, interface{}]

	// currRound indicates the current round the prover is processing and it is incremented
	// every time that the [ProverRuntime.goNextRound] method is called by
	// the [Prove] function.
	currRound int

	// FS stores the Fiat-Shamir State, you probably don't want to use it
	// directly unless you know what you are doing. Just know that if you use
	// it to update the FS hash, this can potentially result in the prover and
	// the verifer end up having different state or the same message being
	// included a second time. Use it externally at your own risks.
	FS hash.StateStorer

	// lock is global lock so that the assignment maps are thread safes
	lock *sync.Mutex

	// FiatShamirHistory tracks the fiat-shamir state at the beginning of every
	// round. The first entry is the initial state, the final entry is the final
	// state.
	FiatShamirHistory [][2][]field.Element

	PerformanceMonitor *config.PerformanceMonitor

	// PerformanceLogs stores performance metrics for each major operation
	PerformanceLogs []*profiling.PerformanceLog

	// High-level prover function
	HighLevelProver MainProverStep[T]
}

// Prove is the top-level function that runs the Prover on the user's side. It
// is responsible for instantiating a fresh and new ProverRuntime and running
// the user's and compiler's [MainProverStep] in order and calling the Fiat-Shamir
// state to generate the randomness between every rounds.
//
// The caller can specify a `highLevelProver` function that implements the
// allocation of the columns and parameters defined in the [Compile]
// via the `define` parameter of the [Compile] function used to construct
// the provided [CompiledIOP] object `c`. In this case, and only in this case,
// the `highLevelProver` function is allowed to span over multiple interaction
// rounds between the prover and the verifier of the protocol. When this
// happens, the underlying [ProverRuntime] object is able to automatically
// follow and detect when the `highLevelProver` function is entering in a new
// round of the protocol.
//
// However, we plan to deprecate this behavior and plan to require the user to
// concretely break down the high-level prover round-by-round as this
// auto-detection adds little value and adds a lot of convolution especially
// when the specified protocol is complicated and involves multiple multi-rounds
// sub-protocols that runs independently.
func Prove[T zk.Element](c *CompiledIOP[T], highLevelprover MainProverStep[T]) Proof[T] {
	run := RunProver(c, highLevelprover)

	// Write the performance logs to the csv file is the performance monitor is active
	if run.PerformanceMonitor.Active {
		if err := run.writePerformanceLogsToCSV(); err != nil {
			utils.Panic("error writing performance logs to CSV: %v", err.Error())
		}
	}

	return run.ExtractProof()
}

// RunProver initializes a [ProverRuntime], runs the prover and returns the final
// runtime. It does not returns the [Proof] however.
func RunProver[T zk.Element](c *CompiledIOP[T], highLevelprover MainProverStep[T]) *ProverRuntime[T] {
	return RunProverUntilRound(c, highLevelprover, c.NumRounds())
}

// RunProverUntilRound runs the prover until the specified round
// We wrap highLevelProver with a struct that implements the prover action interface
func RunProverUntilRound[T zk.Element](c *CompiledIOP[T], highLevelProver MainProverStep[T], round int) *ProverRuntime[T] {
	runtime := c.createProver()
	runtime.HighLevelProver = highLevelProver

	// Execute the high-level prover as a ProverAction
	if runtime.HighLevelProver != nil {
		runtime.exec("high-level-prover", mainProverStepWrapper[T]{step: highLevelProver})
	}

	// Run sub-prover steps for the initial round
	runtime.exec(fmt.Sprintf("prover-steps-round%d", runtime.currRound), runtime.runProverSteps)

	for runtime.currRound+1 < round {
		runtime.exec(fmt.Sprintf("next-after-round-%d", runtime.currRound), runtime.goNextRound)
		runtime.exec(fmt.Sprintf("prover-steps-round-%d", runtime.currRound), runtime.runProverSteps)
	}

	return &runtime
}

// ExtractProof extracts the proof from a [ProverRuntime]. If the runtime has
// been obtained via a [RunProverUntilRound], then it may be the case that
// some columns have not been assigned at all. Those won't be included in the
// returned proof.
func (run *ProverRuntime[T]) ExtractProof() Proof[T] {
	messages := collection.NewMapping[ifaces.ColID, ifaces.ColAssignment]()
	for _, name := range run.Spec.Columns.AllKeysProof() {
		cols := run.Spec.Columns.GetHandle(name)
		if run.currRound < cols.Round() {
			continue
		}
		messageValue := run.Columns.MustGet(name)
		messages.InsertNew(name, messageValue)
	}

	queriesParams := collection.NewMapping[ifaces.QueryID, ifaces.QueryParams]()
	for round := 0; round <= run.currRound; round++ {
		for _, name := range run.Spec.QueriesParams.AllKeysAt(round) {
			queriesParams.InsertNew(name, run.QueriesParams.MustGet(name))
		}
	}

	return Proof[T]{
		Messages:      messages,
		QueriesParams: queriesParams,
	}
}

// NumRounds returns the total number of rounds in the corresponding WizardIOP.
//
// Deprecated: this method does not bring anything useful as its already easy
// to get this value from the Spec
func (run *ProverRuntime[T]) NumRounds() int {
	/*
		Getting it from the spec is the safest as it is already
		tested. We could fit more assertions here nonetheless.
	*/
	return run.Spec.NumRounds()
}

// createProver is the internal function that is used by the [Prove]
// function to instantiate and fresh and new [ProverRuntime].
func (c *CompiledIOP[T]) createProver() ProverRuntime[T] {

	// Create a new fresh FS state and bootstrap it
	fs := poseidon2.NewMerkleDamgardHasher()
	fiatshamir.Update(fs, c.FiatShamirSetup)

	// Instantiates an empty Assignment (but link it to the CompiledIOP[T])
	runtime := ProverRuntime[T]{
		Spec:               c,
		Columns:            collection.NewMapping[ifaces.ColID, ifaces.ColAssignment](),
		QueriesParams:      collection.NewMapping[ifaces.QueryID, ifaces.QueryParams](),
		Coins:              collection.NewMapping[coin.Name, interface{}](),
		State:              collection.NewMapping[string, interface{}](),
		FS:                 fs,
		currRound:          0,
		lock:               &sync.Mutex{},
		FiatShamirHistory:  make([][2][]field.Element, c.NumRounds()),
		PerformanceMonitor: profiling.GetMonitorParams(),
	}

	stateBytes := fs.State()
	var state koalabear.Element
	state.SetBytes(stateBytes)
	runtime.FiatShamirHistory[0] = [2][]field.Element{
		{state},
		{state},
	}

	// Pass the precomputed polynomials
	for key, val := range c.Precomputed.GetInnerMap() {
		runtime.Columns.InsertNew(key, val)
	}

	return runtime
}

// GetColumn implements `ifaces.Runtime`. Returns a column witness, that has been
// previously stored. It is a deep-copy operation. And thus, it guarantees that
// the stored witness cannot be accidentally mutated by the caller as a side
// effect.
//
// Something to note however, is that the function will panic if the
// the provided name does not exists explictly in the [ProverRuntime.Columns]
// database and this will be the case if the attempts to recover a column such
// as a [column.Shifted] or any other type of derivative columns. While theses
// columns are absolutely legal they are not stored explicitly in the runtime
// and they must be accessed through the [ifaces.Column.GetColAssignment]
// method instead which will work for any type of column. The user should use
// the latter as a go-to way to access an assigned column. The reason this
// function is exported is to make it accessible to the other functions of the
// [github.com/consensys/linea-monorepo/prover/protocol/column] package.
//
// Namely, the function will panic if:
//   - `name` relates to a column that does not exists or to a column that is
//     not explictly an assigned column.
//   - `name` relates to a column that does exists but whose assignment is
//     not readily available when the function is called.
func (run ProverRuntime[T]) GetColumn(name ifaces.ColID) ifaces.ColAssignment {

	// global prover's lock before accessing the witnesses
	run.lock.Lock()
	defer run.lock.Unlock()

	/*
		Make sure the column is registered. If the name is the one specified
		does not correcpond to a natural column, this will panic. And this is
		expected behaviour.
	*/
	run.Spec.Columns.MustHaveName(name)
	res := run.Columns.MustGet(name)
	return res
}

// HasColumn returns whether the column is assigned. The function panics if the
// provided column name does not exists
func (run ProverRuntime[T]) HasColumn(name ifaces.ColID) bool {

	// global prover's lock before accessing the witnesses
	run.lock.Lock()
	defer run.lock.Unlock()

	/*
		Make sure the column is registered. If the name is the one specified
		does not correcpond to a natural column, this will panic. And this is
		expected behaviour.
	*/
	run.Spec.Columns.MustHaveName(name)
	return run.Columns.Exists(name)
}

// CopyColumnInto implements `column.GetWitness`. Copies the witness into a slice
// Deprecated: this is deadcode
func (run ProverRuntime[T]) CopyColumnInto(name ifaces.ColID, buff *ifaces.ColAssignment) {

	// global prover's lock before accessing the witnesses
	run.lock.Lock()
	defer run.lock.Unlock()

	/*
		Make sure the column is registered. If the name is the one specified
		does not correcpond to a natural column, this will panic. And this is
		expected behaviour.
	*/
	run.Spec.Columns.MustHaveName(name)
	toCopy := run.Columns.MustGet(name)

	if toCopy.Len() != (*buff).Len() {
		utils.Panic("buffer has the wrong length %v, witness has length %v", (*buff).Len(), toCopy.Len())
	}

	smartvectors.Copy(buff, toCopy)
}

// GetColumnAt does the same as [GetColumn] but only returns a single position
// instead of returning the whole vector; i.e. it returns the assignment of
// an explictly assigned column at a requested position.
//
// The same cautiousness as for [ProverRuntime.AssignColumn] applies to this
// function. Namely, this function will only work if the requested column is
// explicitly an assigned column (meaning not a derive column).
func (run ProverRuntime[T]) GetColumnAt(name ifaces.ColID, pos int) field.Element {

	// global prover's lock before accessing the witnesses
	run.lock.Lock()
	defer run.lock.Unlock()

	/*
		Make sure the column is registered. If the name is the one specified
		does not correcpond to a natural column, this will panic. And this is
		expected behaviour.
	*/
	run.Spec.Columns.MustHaveName(name)
	wit := run.Columns.MustGet(name)

	if pos >= wit.Len() || pos < 0 {
		utils.Panic("asked pos %v for vector of size %v", pos, wit)
	}

	return wit.Get(pos)
}

func (run ProverRuntime[T]) GetColumnAtBase(name ifaces.ColID, pos int) (field.Element, error) {
	// global prover's lock before accessing the witnesses
	run.lock.Lock()
	defer run.lock.Unlock()

	/*
		Make sure the column is registered. If the name is the one specified
		does not correcpond to a natural column, this will panic. And this is
		expected behaviour.
	*/
	run.Spec.Columns.MustHaveName(name)
	wit := run.Columns.MustGet(name)

	if _, err := wit.GetBase(0); err == nil {
		if pos >= wit.Len() || pos < 0 {
			utils.Panic("asked pos %v for vector of size %v", pos, wit)
		}
		result, _ := wit.GetBase(pos)
		return result, nil
	} else {
		return field.Zero(), err
	}

}

func (run ProverRuntime[T]) GetColumnAtExt(name ifaces.ColID, pos int) fext.Element {
	// global prover's lock before accessing the witnesses
	run.lock.Lock()
	defer run.lock.Unlock()

	/*
		Make sure the column is registered. If the name is the one specified
		does not correcpond to a natural column, this will panic. And this is
		expected behaviour.
	*/
	run.Spec.Columns.MustHaveName(name)
	wit := run.Columns.MustGet(name)
	return wit.GetExt(pos)
}

// GetRandomCoinField returns a field element random. The coin should be issued
// at the same round as it was registered. The same coin can't be retrieved more
// than once. The coin should also have been registered as a field element
// before doing this call. Will also trigger the "goNextRound" logic if
// appropriate.
//
// The coin must also be of type [coin.FieldFromSeed] or [coin.Field].
func (run *ProverRuntime[T]) GetRandomCoinField(name coin.Name) field.Element {
	mycoin := run.Spec.Coins.Data(name)
	if mycoin.Type != coin.Field && mycoin.Type != coin.FieldFromSeed {
		utils.Panic("coin %v is not a field randomness", name)
	}
	return run.getRandomCoinGeneric(name, mycoin.Type).(field.Element)
}

// GetRandomCoinFieldExt returns a field extension randomness. The coin should
// be isseued at the same round as it was registered. The same coin can't be
// retrieved more than once. The coin should also have been registered as a
// field extension randomness before doing this call. Will also trigger the
// "goNextRound" logic if appropriate.
//
// The type must also be of type [coin.FieldExtFromSeed] or [coin.FieldExt].
func (run *ProverRuntime[T]) GetRandomCoinFieldExt(name coin.Name) fext.Element {
	mycoin := run.Spec.Coins.Data(name)
	if mycoin.Type != coin.FieldExt && mycoin.Type != coin.FieldExtFromSeed {
		utils.Panic("coin %v is not a field extension randomness", name)
	}
	return run.getRandomCoinGeneric(name, mycoin.Type).(fext.Element)
}

// GetRandomCoinIntegerVec returns a pre-sampled integer vec random coin. The
// coin should be issued at the same round as it was registered. The same coin
// can't be retrieved more than once. The coin should also have been registered
// as an integer vec before doing this call. Will also trigger the
// "goNextRound" logic if appropriate.
func (run *ProverRuntime[T]) GetRandomCoinIntegerVec(name coin.Name) []int {
	return run.getRandomCoinGeneric(name, coin.IntegerVec).([]int)
}

// AssignColumn assigns a value to a column specified in the underlying
// CompiledIOP. For an external user, it should be used only on columns
// explicitly created via the [Builder.RegisterCommit] or
// [CompiledIOP.InsertColumn], [CompiledIOP.InsertCommit] or
// [CompiledIOP.InsertProof] or even [CompiledIOP.InsertPublicInput].
//
// The function will panic if
//   - an empty column name is provided
//   - the column is not explictly registered in the CompiledIOP (e.g. if it is
//     a derive column or the underlying type is found in
//     [github.com/consensys/linea-monorepo/prover/protocol/column/verifiercol] for instance).
//   - the assignment does not have the correct size
//   - the column assignment occurs at the wrong round. If this error happens,
//     it is likely that the [ifaces.Column] was created in the wrong round to
//     begin with.
func (run *ProverRuntime[T]) AssignColumn(name ifaces.ColID, witness ifaces.ColAssignment, opts ...ProverRuntimeOption) {

	var opts_ ProverRuntimeOption
	for i := range opts {
		opts_ |= opts[i]
	}

	// global prover's lock before accessing the witnesses. This makes the
	// function thread-safe
	run.lock.Lock()
	defer run.lock.Unlock()

	// Sanity-check : the handle should not be empty
	if len(name) == 0 {
		panic("given an empty name")
	}

	// @alex: this check is redundant with the following check since GetHandle
	// would panic as it won't be able to find the column if it is
	// "VerifierDefined". The advantage of this check is that it provides a
	// more accurate error message if this occurs.
	if run.Spec.Columns.Status(name) == column.VerifierDefined {
		utils.Panic("tried assign a verifier defined column : %v", name)
	}

	// Sanity-check: Make sure, it is done at the right round
	handle := run.Spec.Columns.GetHandle(name).(column.Natural[T])

	// if round is empty, we expect it to assign the column at the current round,
	// otherwise it assigns it in the round the column was declared.
	// This is useful when we have for loop over rounds.
	ifaces.MustBeInRound(handle, run.currRound)

	// This sanity-check here is to ensure that if we declare a column as "IsBase"
	// , then it's assignment should be done on the base field. If the provides
	// a field extension witness, then the function will try to cast it into
	// a [smartvectors.Regular] and will panic if that is not possible.
	if handle.IsBase() && !smartvectors.IsBase(witness) {
		w_, err := smartvectors.IntoBase(witness)
		if err != nil {
			utils.Panic("could not convert witness into base smartvector: %v", err)
		}
		witness = w_
	}

	if witness.Len() != handle.Size() {
		utils.Panic("Bad length for %v, expected %v got %v\n", handle, handle.Size(), witness.Len())
	}

	// This is supposed to be redundant with the checks we make when
	// registering a column. So, if it fails here it likely should have failed
	// earlier. Thus, the check is there purely for defensive purposes.
	if !utils.IsPowerOfTwo(witness.Len()) {
		utils.Panic("Witness with non-power of two sizes, should have been caught earlier")
	}

	// If the column is generated after the first round, there is no need
	// optimizing the assignment because it is likely created by wizard
	// compilation and its representation is already optimized.
	if run.currRound > 0 {
		// Adds it to the assignments
		run.Columns.InsertNew(handle.GetColID(), witness)
		return
	}

	start, stop := smartvectors.CoWindowRange(witness)

	var (
		hasRightPaddedRange  = stop < witness.Len()
		hasLeftPaddedRange   = start > 0
		hasRightPaddedPragma = pragmas.IsRightPadded(handle)
		hasLeftPaddedPragma  = pragmas.IsLeftPadded(handle)
		hasFullColumnPragma  = pragmas.IsFullColumn(handle)
	)

	switch {
	case hasLeftPaddedPragma:

		if !hasLeftPaddedRange {
			// logrus.Warnf("Left-padded column with non-left-padded witness: %v, start: %v, stop: %v", name, start, stop)
			// This conversion to regular ensures that the witness won't be
			// stored as a right-padded column. The size reduction might later
			// find a padding opportunity in the right direction. The conversion
			// is ineffective in case the column is a regular column (which
			// might be caught by the condition)
			witness = smartvectors.NewRegular(witness.IntoRegVecSaveAlloc())
		}

		// This reduction is a trade-off between runtime and memory. It costs CPU
		// but can save a significant amount of memory.
		if opts_&DisableAssignmentSizeReduction == 0 {
			witness, _ = smartvectors.TryReduceSizeLeft(witness)
		}

	case hasRightPaddedPragma:

		if !hasRightPaddedRange {
			// logrus.Warnf("Right-padded column with non-right-padded witness: %v, start: %v, stop: %v", name, start, stop)
			// This conversion to regular ensures that the witness won't be
			// stored as a left-padded column. The size reduction might later
			// find a padding opportunity in the right direction. The conversion
			// is ineffective in case the column is a regular column (which
			// might be caught by the condition)
			witness = smartvectors.NewRegular(witness.IntoRegVecSaveAlloc())
		}

		// This reduction is a trade-off between runtime and memory. It costs CPU
		// but can save a significant amount of memory.
		if opts_&DisableAssignmentSizeReduction == 0 {
			witness, _ = smartvectors.TryReduceSizeRight(witness)
		}

	case hasFullColumnPragma:

		if hasLeftPaddedRange || hasRightPaddedRange {
			logrus.Errorf("Full column with non-full witness: %v, start: %v, stop: %v", name, start, stop)
		}

		witness = smartvectors.NewRegular(witness.IntoRegVecSaveAlloc())

	default:

		// This reduction is a trade-off between runtime and memory. It costs CPU
		// but can save a significant amount of memory.
		if opts_&DisableAssignmentSizeReduction == 0 {
			witness, _ = smartvectors.TryReduceSizeRight(witness)
		}
	}

	// Adds it to the assignments
	run.Columns.InsertNew(handle.GetColID(), witness)
}

// getRandomCoinGeneric is an internal utility function that we use when
// resolving the value of a random coin. When called in the context of the
// `highLevelProver` argument function of [Prove], the function
// encompasses the necessary logic for understanding when to move to the next
// round.
//
// We plan on removing this "go-to-next-round" auto-detection logic in the
// future as this convolutes the code for little benefits.
func (run *ProverRuntime[T]) getRandomCoinGeneric(name coin.Name, requestedType coin.Type) interface{} {
	/*
		Early check, ensures the coin has been registered at all
		and that it has the correct type
	*/
	infos := run.Spec.Coins.Data(name)
	if infos.Type != requestedType {
		utils.Panic("Coin was registered as %v but got %v", infos.Type, requestedType)
	}

	// Check if maybe we need to go to the next round
	foundAtRound := run.Spec.Coins.Round(name)

	switch {
	case foundAtRound <= run.currRound:
		/*
			Regular case : we are fetching the value of a past or current round
			and the coin is already available
		*/
	case foundAtRound == run.currRound+1:
		/*
			The user wants a coin for the next round. This signals that the user is
			moving to the next round. At this stage the "compiled steps" of the current
			rounds still have not been run. This needs to be done before we leave the
			current round. After that, we can generate the coins for the new round.
		*/
		run.runProverSteps()
		/*
			Then, we can try to transitionate to the next rounds. We assert all witnesses,
			params, messages etc.. from the past round have been set. Then, we can generate
			the coins the user is requesting.
		*/
		run.goNextRound()

	case foundAtRound > run.currRound+1:
		/*
			Certainly a bug
		*/
		utils.Panic("Requested coin %v (registered at round %v) but we are at round %v", name, foundAtRound, run.currRound)
	}

	/*
		Note, if we are at round 0. We are guaranteed that this will
		panic (and this is expected, but it does not make sense to query)
		a random coin ex-nihilo.
	*/
	return run.Coins.MustGet(name)
}

// Internal function handling the transition to the next round. This function
// is called when the prover is entering into the next rounds. It proceeds by
// first ensuring that all the columns that were defined in the current round
// have been assigned, otherwise `goNextRound` will panic. Then, it passes all
// the columns visible to the verifier into the Fiat-Shamir state of the
// protocol (i.e. all the columns with either the tag [column.Proof] or
// [column.PublicInput] or the [ifaces.QueryParams] object that have been
// assigned during the current round. Then the function proceeds by entering in
// the next round and sampling all the random coins for the new round with the
// Fiat-Shamir state that we just obtained by appending all the columns and
// parameters. This makes all the new coins available in the prover runtime.
func (run *ProverRuntime[T]) goNextRound() {

	initialStateBytes := run.FS.State()
	var initialState koalabear.Element
	initialState.SetBytes(initialStateBytes)

	if !run.Spec.DummyCompiled {

		/*
			Make sure that all messages have been written and use them
			to update the FS state.  Note that we do not need to update
			FS using the last round of the prover because he is always
			the last one to "talk" in the protocol.
		*/
		msgsToFS := run.Spec.Columns.AllKeysInProverTranscript(run.currRound)
		for _, msgName := range msgsToFS {

			if run.Spec.Columns.IsExplicitlyExcludedFromProverFS(msgName) {
				continue
			}

			if run.Spec.Precomputed.Exists(msgName) {
				continue
			}

			instance := run.GetMessage(msgName)
			fiatshamir.UpdateSV(run.FS, instance)
		}

		/*
			Also include the prover's allegations for all evaluations
		*/
		paramsToFS := run.Spec.QueriesParams.AllKeysAt(run.currRound)
		for _, qName := range paramsToFS {
			if run.Spec.QueriesParams.IsSkippedFromProverTranscript(qName) {
				continue
			}

			// Implicitly, this will panic whenever we start supporting
			// a new type of query params
			params := run.QueriesParams.MustGet(qName)
			params.UpdateFS(run.FS)
		}
	}

	// Increment the number of rounds
	run.currRound++

	if run.Spec.FiatShamirHooksPreSampling.Len() > run.currRound {
		fsHooks := run.Spec.FiatShamirHooksPreSampling.MustGet(run.currRound)
		for i := range fsHooks {
			// if fsHooks[i].IsSkipped() {
			// 	continue
			// }

			fsHooks[i].Run(run)
		}
	}

	seedByte := run.FS.State()
	var seed koalabear.Element
	seed.SetBytes(seedByte)

	// Then assigns the coins for the new round. As the round
	// incrementation is made lazily, we expect that there is
	// a next round.
	toCompute := run.Spec.Coins.AllKeysAt(run.currRound)

	for _, myCoin := range toCompute {
		if run.Spec.Coins.IsSkippedFromProverTranscript(myCoin) {
			continue
		}

		info := run.Spec.Coins.Data(myCoin)
		value := info.Sample(run.FS, seed)
		run.Coins.InsertNew(myCoin, value)
	}

	finalStateBytes := run.FS.State()
	var finalState koalabear.Element
	finalState.SetBytes(finalStateBytes)

	run.FiatShamirHistory[run.currRound] = [2][]field.Element{
		[]koalabear.Element{initialState},
		[]koalabear.Element{finalState},
	}
}

// runProverSteps runs all the [ProverStep] specified in the underlying
// [CompiledIOP] object for the current round.
func (run *ProverRuntime[T]) runProverSteps() {
	// Run all the assigners
	subProverSteps := run.Spec.SubProvers.MustGet(run.currRound)
	for idx, step := range subProverSteps {

		// Profile individual prover steps
		namePrefix := fmt.Sprintf("prover-round%d-step%d", run.currRound, idx)
		run.exec(namePrefix, step)
	}
}

// GetMessage gets a message sent to the verifier
// Deprecated: use [ProverRuntime.GetColumn] instead
func (run *ProverRuntime[T]) GetMessage(name ifaces.ColID) ifaces.ColAssignment {
	// Global prover locks for accessing the maps
	run.lock.Lock()
	defer run.lock.Unlock()

	// Sanity-check, this panics if the column does not exists
	return run.Columns.MustGet(name)
}

// GetInnerProduct returns an inner-product query from the underlying CompiledIOP.
// Deprecated: directly use CompiledIOP.Spec.GetInnerProduct() instead, which
// does exactly the same thing.
func (run *ProverRuntime[T]) GetInnerProduct(name ifaces.QueryID) query.InnerProduct[T] {
	return run.Spec.GetInnerProduct(name)
}

// GetInnerProductParams returns pre-assigned parameters for the specified
// [query.InnerProduct] query. The caller specifies the query by its name and
// the method returns the query's parameters. As a reminders, the parameters of
// the query means the result of the inner-products.
//
// The function will panic of the parameters are not available or if the
// parameters have the wrong type: not an [query.InnerProductParams].
func (run *ProverRuntime[T]) GetInnerProductParams(name ifaces.QueryID) query.InnerProductParams[T] {
	return run.QueriesParams.MustGet(name).(query.InnerProductParams[T])
}

// AssignInnerProduct assigns the result of an inner-product query in the
// prover runtime. The function will panic if
//   - the wrong number of `ys` value is provided. It should match the length
//     of `bs` that was provided when registering the query.
//   - no query with the name `name` are found in the [CompiledIOP] object.
//   - parameters for this query have already been assigned
//   - the assignment round is not the correct one
func (run *ProverRuntime[T]) AssignInnerProduct(name ifaces.QueryID, ys ...fext.Element) query.InnerProductParams[T] {
	q := run.GetInnerProduct(name)
	if len(q.Bs) != len(ys) {
		utils.Panic("Inner-product query %v has %v bs but assigned for %v", name, len(q.Bs), len(ys))
	}

	// Make sure, it is done at the right round
	run.Spec.QueriesParams.MustBeInRound(run.currRound, name)

	param := query.NewInnerProductParams[T](ys...)
	run.QueriesParams.InsertNew(name, param)
	return param
}

// AssignUnivariate assigns the evaluation point and the evaluation result
// and claimed values for a univariate evaluation bearing `name` as an ID.
//
// The function will panic if:
//   - the wrong number of `ys` value is provided. It should match the length
//     of `bs` that was provided when registering the query.
//   - no query with the name `name` are found in the [CompiledIOP] object.
//   - parameters for this query have already been assigned
//   - the assignment round is not the correct one
func (run *ProverRuntime[T]) AssignUnivariate(name ifaces.QueryID, x field.Element, ys ...field.Element) {

	// Global prover locks for accessing the maps
	run.lock.Lock()
	defer run.lock.Unlock()

	// Make sure, it is done at the right round
	run.Spec.QueriesParams.MustBeInRound(run.currRound, name)

	// Check the length of ys
	q := run.Spec.QueriesParams.Data(name).(query.UnivariateEval[T])
	if len(q.Pols) != len(ys) {
		utils.Panic("Query expected ys = %v but got %v", len(q.Pols), len(ys))
	}
	// Adds it to the assignments
	params := query.NewUnivariateEvalParams[T](x, ys...)
	run.QueriesParams.InsertNew(name, params)
}

func (run *ProverRuntime[T]) AssignUnivariateExt(name ifaces.QueryID, x fext.Element, ys ...fext.Element) {

	// Global prover locks for accessing the maps
	run.lock.Lock()
	defer run.lock.Unlock()

	// Make sure, it is done at the right round
	run.Spec.QueriesParams.MustBeInRound(run.currRound, name)

	// Check the length of ys
	q := run.Spec.QueriesParams.Data(name).(query.UnivariateEval[T])
	if len(q.Pols) != len(ys) {
		utils.Panic("Query expected ys = %v but got %v", len(q.Pols), len(ys))
	}
	// Adds it to the assignments
	params := query.NewUnivariateEvalParamsExt[T](x, ys...)
	run.QueriesParams.InsertNew(name, params)
}

// GetUnivariateEval get univariate eval metadata. Panic if not found.
// Deprecated: fallback to run.Spec.GetUnivariateEval instead which does exactly
// the same thing.
func (run *ProverRuntime[T]) GetUnivariateEval(name ifaces.QueryID) query.UnivariateEval[T] {
	// Global prover locks for accessing the maps
	run.lock.Lock()
	defer run.lock.Unlock()
	return run.Spec.QueriesParams.Data(name).(query.UnivariateEval[T])
}

// GetUnivariateParams returns the parameters of a univariate evaluation (i.e:
// x, the evaluation point) and y, the alleged polynomial opening. This is
// intended to resolve parameters that have been already assigned in a previous
// step of the prover runtime.
func (run *ProverRuntime[T]) GetUnivariateParams(name ifaces.QueryID) query.UnivariateEvalParams[T] {
	// Global prover's lock for accessing params
	run.lock.Lock()
	defer run.lock.Unlock()
	return run.QueriesParams.MustGet(name).(query.UnivariateEvalParams[T])
}

// AssignLocalPoint assign evaluation point and claimed values for a local point
// opening. The function will panic if:
//   - the parameters were already assigned
//   - the specified query is not registered
//   - the assignment round is incorrect
func (run *ProverRuntime[T]) AssignLocalPoint(name ifaces.QueryID, y field.Element) {

	// Global prover locks for accessing the maps
	run.lock.Lock()
	defer run.lock.Unlock()

	// Make sure, it is done at the right round
	run.Spec.QueriesParams.MustBeInRound(run.currRound, name)

	q := run.Spec.QueriesParams.Data(name).(query.LocalOpening[T])
	if !q.IsBase() {
		utils.Panic("Query %v is not a base query, you should call AssignLocalPointExt", name)
	}

	// Adds it to the assignments
	params := query.NewLocalOpeningParams[T](y)
	run.QueriesParams.InsertNew(name, params)
}

func (run *ProverRuntime[T]) AssignLocalPointExt(name ifaces.QueryID, y fext.Element) {

	// Global prover locks for accessing the maps
	run.lock.Lock()
	defer run.lock.Unlock()

	// Make sure, it is done at the right round
	run.Spec.QueriesParams.MustBeInRound(run.currRound, name)

	q := run.Spec.QueriesParams.Data(name).(query.LocalOpening[T])
	if q.IsBase() {
		utils.Panic("Query %v is a base query, you should call AssignLocalPoint", name)
	}

	// Adds it to the assignments
	params := query.NewLocalOpeningParamsExt[T](y)
	run.QueriesParams.InsertNew(name, params)
}

// GetLocalPointEval gets the metadata of a [query.LocalOpening[T]] query. Panic if not found.
// Deprecated, use `comp.Spec.GetLocalPointEval` instead since it does exactly
// the same thing.
func (run *ProverRuntime[T]) GetLocalPointEval(name ifaces.QueryID) query.LocalOpening[T] {
	// Global prover locks for accessing the maps
	run.lock.Lock()
	defer run.lock.Unlock()
	return run.Spec.QueriesParams.Data(name).(query.LocalOpening[T])
}

// GetLocalPointEvalParams returns the parameters of a univariate evaluation
// (i.e: x, the evaluation point) and y, the alleged polynomial opening.
func (run *ProverRuntime[T]) GetLocalPointEvalParams(name ifaces.QueryID) query.LocalOpeningParams[T] {

	// Global prover's lock for accessing params
	run.lock.Lock()
	defer run.lock.Unlock()

	return run.QueriesParams.MustGet(name).(query.LocalOpeningParams[T])
}

// AssignLogDerivSum assign the claimed values for a logDeriveSum
// The function will panic if:
//   - the parameters were already assigned
//   - the specified query is not registered
//   - the assignment round is incorrect
func (run *ProverRuntime[T]) AssignLogDerivSum(name ifaces.QueryID, y fext.GenericFieldElem) {

	// Global prover locks for accessing the maps
	run.lock.Lock()
	defer run.lock.Unlock()

	// Make sure, it is done at the right round
	run.Spec.QueriesParams.MustBeInRound(run.currRound, name)

	// Adds it to the assignments
	params := query.NewLogDerivSumParams[T](y)
	run.QueriesParams.InsertNew(name, params)
}

// AssignGrandProduct assign the claimed values for a grand product
// The function will panic if:
//   - the parameters were already assigned
//   - the specified query is not registered
//   - the assignment round is incorrect
func (run *ProverRuntime[T]) AssignGrandProduct(name ifaces.QueryID, y field.Element) {

	// Global prover locks for accessing the maps
	run.lock.Lock()
	defer run.lock.Unlock()

	// Make sure, it is done at the right round
	run.Spec.QueriesParams.MustBeInRound(run.currRound, name)

	// Adds it to the assignments
	params := query.NewGrandProductParams(y)
	run.QueriesParams.InsertNew(name, params)
}

func (run *ProverRuntime[T]) AssignGrandProductExt(name ifaces.QueryID, y fext.Element) {

	// Global prover locks for accessing the maps
	run.lock.Lock()
	defer run.lock.Unlock()

	// Make sure, it is done at the right round
	run.Spec.QueriesParams.MustBeInRound(run.currRound, name)

	// Adds it to the assignments
	params := query.NewGrandProductParamsExt(y)
	run.QueriesParams.InsertNew(name, params)
}

// GetLogDeriveSum gets the metadata of a [query.LogDerivativeSum] query. Panic if not found.
func (run *ProverRuntime[T]) GetLogDeriveSum(name ifaces.QueryID) query.LogDerivativeSum[T] {
	// Global prover locks for accessing the maps
	run.lock.Lock()
	defer run.lock.Unlock()
	return run.Spec.QueriesParams.Data(name).(query.LogDerivativeSum[T])
}

// GetLogDerivSumParams returns the parameters of [query.LogDerivativeSum]
func (run *ProverRuntime[T]) GetLogDerivSumParams(name ifaces.QueryID) query.LogDerivSumParams[T] {

	// Global prover's lock for accessing params
	run.lock.Lock()
	defer run.lock.Unlock()

	return run.QueriesParams.MustGet(name).(query.LogDerivSumParams[T])
}

// GetGrandProductParams returns the parameters of a [query.Honer] query.
func (run *ProverRuntime[T]) GetGrandProductParams(name ifaces.QueryID) query.GrandProductParams {
	return run.QueriesParams.MustGet(name).(query.GrandProductParams)
}

// GetParams generically extracts the parameters of a query. Will panic if no
// parameters are found
func (run *ProverRuntime[T]) GetParams(name ifaces.QueryID) ifaces.QueryParams {
	return run.QueriesParams.MustGet(name)
}

// AssignHornerParams assignes the parameters of a [query.Honer] query.
func (run *ProverRuntime[T]) AssignHornerParams(name ifaces.QueryID, params query.HornerParams[T]) {
	// Global prover locks for accessing the maps
	run.lock.Lock()
	defer run.lock.Unlock()

	// Make sure, it is done at the right round
	run.Spec.QueriesParams.MustBeInRound(run.currRound, name)

	// Adds it to the assignments
	run.QueriesParams.InsertNew(name, params)
}

// GetHornerParams returns the parameters of a [query.Honer] query.
func (run *ProverRuntime[T]) GetHornerParams(name ifaces.QueryID) query.HornerParams[T] {
	// Global prover's lock for accessing params
	run.lock.Lock()
	defer run.lock.Unlock()
	return run.QueriesParams.MustGet(name).(query.HornerParams[T])
}

// Fs returns the Fiat-Shamir state
func (run *ProverRuntime[T]) Fs() hash.StateStorer {
	return run.FS
}

// FsHistory returns the Fiat-Shamir state history
func (run *ProverRuntime[T]) FsHistory() [][2][]field.Element {
	return run.FiatShamirHistory
}

// GetPublicInputs return the value of a public-input from its name
func (run *ProverRuntime[T]) GetPublicInput(name string) field.Element {
	allPubs := run.Spec.PublicInputs
	for i := range allPubs {
		if allPubs[i].Name == name {
			return allPubs[i].Acc.GetVal(run)
		}
	}
	utils.Panic("could not find public input nb %v", name)
	return field.Element{}
}

// GetQuery returns a query from its name
func (run *ProverRuntime[T]) GetQuery(name ifaces.QueryID) ifaces.Query[T] {

	if run.Spec.QueriesParams.Exists(name) {
		return run.Spec.QueriesParams.Data(name)
	}

	if run.Spec.QueriesNoParams.Exists(name) {
		return run.Spec.QueriesNoParams.Data(name)
	}

	utils.Panic("could not find query nb %v", name)
	return nil
}

// GetSpec returns the underlying compiled IOP
func (run *ProverRuntime[T]) GetSpec() *CompiledIOP[T] {
	return run.Spec
}

// GetState returns an arbitrary value stored in the runtime
func (run *ProverRuntime[T]) GetState(name string) (any, bool) {
	res, ok := run.State.TryGet(name)
	return res, ok
}

// SetState sets an arbitrary value in the runtime
func (run *ProverRuntime[T]) SetState(name string, value any) {
	run.State.InsertNew(name, value)
}

// InsertCoin is there so that [ProverRuntime] implements the [ifaces.Runtime]
// but the function panicks if called.
func (run *ProverRuntime[T]) InsertCoin(name coin.Name, value any) {
	utils.Panic("InsertCoin is not implemented")
}

// exec: executes the `action` with the performance monitor if active
func (runtime *ProverRuntime[T]) exec(name string, action any) {

	// Define helper excute function
	execute := func() {
		switch a := action.(type) {
		case func():
			a()
		case ProverAction[T]:
			a.Run(runtime)
		default:
			utils.Panic("wizard.exec: unsupported action type: got %T; expected one of: func(), ProverAction", action)
		}
	}

	// If PerformanceMonitor is inactive, just execute the action and return
	if !runtime.PerformanceMonitor.Active {
		execute()
		return
	}

	// Determine if profiling is needed based on action type and profile setting
	shouldProfile := false
	switch runtime.PerformanceMonitor.Profile {
	case "all":
		shouldProfile = true
	case "prover-rounds":
		shouldProfile = actionIsPlainFunc(action)
	case "prover-steps":
		shouldProfile = actionIsProverAction[T](action)
	}

	if shouldProfile {
		runtime.profileAction(name, execute)
	} else {
		execute()
	}
}

// profileAction profiles the given action and logs the performance metrics.
func (runtime *ProverRuntime[T]) profileAction(name string, action func()) {
	profilingPath := path.Join(runtime.PerformanceMonitor.ProfileDir, name)
	monitor, err := profiling.StartPerformanceMonitor(name, runtime.PerformanceMonitor.SampleDuration, profilingPath)
	if err != nil {
		panic("error setting up performance monitor for " + name)
	}

	action()

	perfLog, err := monitor.Stop()
	if err != nil {
		logrus.Panicf("error:%s encountered while retrieving performance log for:%s", err.Error(), name)
	}

	// perfLog.PrintMetrics()
	runtime.PerformanceLogs = append(runtime.PerformanceLogs, perfLog)
}

// writePerformanceLogsToCSV: Dumps all the performance logs inside prover runtime
// to the csv file located at the specified path
func (runtime *ProverRuntime[T]) writePerformanceLogsToCSV() error {
	csvFilePath := path.Join(runtime.PerformanceMonitor.ProfileDir, "runtime_performance_logs.csv")
	file, err := os.Create(csvFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	startTime := time.Now()
	logrus.Infof("Writing the runtime performance logs to csv file located at path%s", csvFilePath)

	// Define CSV headers
	headers := []string{
		"Description", "Runtime (s)",
		"CPU_Usage_Min", "CPU_Usage_Avg", "CPU_Usage_Max",
		"Mem_Allocated_Min (GiB)", "Mem_Allocated_Avg (GiB)", "Mem_Allocated_Max (GiB)",
		"Mem_InUse_Min (GiB)", "Mem_InUse_Avg (GiB)", "Mem_InUse_Max (GiB)",
		"Mem_GC_NotDeallocated_Min (GiB)", "Mem_GC_NotDeallocated_Avg (GiB)", "Mem_GC_NotDeallocated_Max (GiB)",
	}
	writer.Write(headers)

	// Write performance logs to CSV
	for _, log := range runtime.PerformanceLogs {
		record := []string{
			log.Description,
			strconv.FormatFloat(log.StopTime.Sub(log.StartTime).Seconds(), 'f', -1, 64),
			strconv.FormatFloat(log.CpuUsageStats[0], 'f', 2, 64),
			strconv.FormatFloat(log.CpuUsageStats[1], 'f', 2, 64),
			strconv.FormatFloat(log.CpuUsageStats[2], 'f', 2, 64),
			strconv.FormatFloat(log.MemoryAllocatedStatsGiB[0], 'f', 2, 64),
			strconv.FormatFloat(log.MemoryAllocatedStatsGiB[1], 'f', 2, 64),
			strconv.FormatFloat(log.MemoryAllocatedStatsGiB[2], 'f', 2, 64),
			strconv.FormatFloat(log.MemoryInUseStatsGiB[0], 'f', 2, 64),
			strconv.FormatFloat(log.MemoryInUseStatsGiB[1], 'f', 2, 64),
			strconv.FormatFloat(log.MemoryInUseStatsGiB[2], 'f', 2, 64),
			strconv.FormatFloat(log.MemoryGCNotDeallocatedStatsGiB[0], 'f', 2, 64),
			strconv.FormatFloat(log.MemoryGCNotDeallocatedStatsGiB[1], 'f', 2, 64),
			strconv.FormatFloat(log.MemoryGCNotDeallocatedStatsGiB[2], 'f', 2, 64),
		}
		writer.Write(record)
	}

	logrus.Infof("Finished writing to the csv file. Took %s", time.Since(startTime).String())
	return nil
}

// actionIsProverRound checks if the action is a plain function such as nextRound or runProverSteps
func actionIsPlainFunc(action any) bool {
	_, ok := action.(func())
	return ok
}

// actionIsProverAction checks if the action is an individual ProverStep in a specific round or highlevelProver.
func actionIsProverAction[T zk.Element](action any) bool {
	_, ok := action.(ProverAction[T])
	return ok
}
