package wizard

import (
	"sync"

	"github.com/consensys/accelerated-crypto-monorepo/crypto/fiatshamir"
	"github.com/consensys/accelerated-crypto-monorepo/maths/common/smartvectors"
	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/coin"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/column"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/ifaces"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/query"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
	"github.com/consensys/accelerated-crypto-monorepo/utils/collection"
	"github.com/sirupsen/logrus"
)

/*
Function that specifies which are the values to pass to the assignment.
*/
type ProverStep func(assi *ProverRuntime)

// ProverRuntime gathers all the runtime data of a prover instance.
type ProverRuntime struct {
	// The "static" description of the protocol
	Spec *CompiledIOP

	// Accumulates, all column's witness assigned so far
	Columns collection.Mapping[ifaces.ColID, ifaces.ColAssignment]
	// Accumulates all the query parameters of the queries assigned so far
	QueriesParams collection.Mapping[ifaces.QueryID, ifaces.QueryParams]
	// Stores all the values of all random Coins
	Coins collection.Mapping[coin.Name, interface{}]

	// State store : "any purpose" data-storage for stateful proving.
	// It allows ProverSteps to persist data that can be accessed in
	// later prover states.
	State collection.Mapping[string, interface{}]

	// Indicate the current round the prover is processing
	currRound int

	// Fiat-Shamir State, you probably don't want to use it directly unless
	// you know what you are doing. Just know that if you use it to update
	// the FS hash, this can potentially result in the prover and the verifer
	// end up having different state or the same message being included a second
	// time.
	FS *fiatshamir.State

	// Global lock so that the assignment maps are thread safes
	lock *sync.Mutex
}

// Top-level function to pass an assignment
func Prove(c *CompiledIOP, highLevelprover ProverStep) Proof {
	runtime := c.CreateProver()
	/*
		Run the user provided assignment function. We can't expect it
		to run all the rounds, because the compilation could have added
		extra-rounds.
	*/
	highLevelprover(&runtime)

	/*
		Then, run the compiled prover steps
	*/
	runtime.runProverSteps()
	for runtime.currRound+1 < runtime.NumRounds() {
		runtime.goNextRound()
		runtime.runProverSteps()
	}

	/*
		Pass all the prover message columns as part of the proof
	*/
	messages := collection.NewMapping[ifaces.ColID, ifaces.ColAssignment]()

	for _, name := range runtime.Spec.Columns.AllKeysProof() {
		messageValue := runtime.Columns.MustGet(name)
		messages.InsertNew(name, messageValue)
	}

	// And also the public inputs
	for _, name := range runtime.Spec.Columns.AllKeysPublicInput() {
		messageValue := runtime.Columns.MustGet(name)
		messages.InsertNew(name, messageValue)
	}

	return Proof{
		Messages:      messages,
		QueriesParams: runtime.QueriesParams,
	}
}

/*
Returns the number of rounds in the assignment.
*/
func (run *ProverRuntime) NumRounds() int {
	/*
		Getting it from the spec is the safest as it is already
		tested. We could fit more assertions here nonetheless.
	*/
	return run.Spec.NumRounds()
}

/*
Constructor for a new an prover runtime for the compiled IOP.
*/
func (c *CompiledIOP) CreateProver() ProverRuntime {

	// Create a new fresh FS state
	fs := fiatshamir.NewMiMCFiatShamir()

	// Instantiates an empty Assignment (but link it to the CompiledIOP)
	runtime := ProverRuntime{
		Spec:          c,
		Columns:       collection.NewMapping[ifaces.ColID, ifaces.ColAssignment](),
		QueriesParams: collection.NewMapping[ifaces.QueryID, ifaces.QueryParams](),
		Coins:         collection.NewMapping[coin.Name, interface{}](),
		State:         collection.NewMapping[string, interface{}](),
		FS:            fs,
		currRound:     0,
		lock:          &sync.Mutex{},
	}

	// Pass the precomputed polynomials
	for key, val := range c.Precomputed.InnerMap() {
		runtime.Columns.InsertNew(key, val)
	}

	return runtime
}

/*
This implements `ifaces.Runtime`. Returns a column witness, that
has been previously stored. It is a deep-copy operation.
*/
func (run ProverRuntime) GetColumn(name ifaces.ColID) ifaces.ColAssignment {

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
	// The deep-copy here protects against side-effects
	return res.DeepCopy()
}

/*
This implements `column.GetWitness`
Copies the witness into a slice
*/
func (run ProverRuntime) CopyColumnInto(name ifaces.ColID, buff *ifaces.ColAssignment) {

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

/*
This implements `column.GetWitness`
Returns a particular entry of a witness
*/
func (run ProverRuntime) GetColumnAt(name ifaces.ColID, pos int) field.Element {

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

/*
Returns a field element random. The coin should be issued at the same round
as it was registered. The same coin can't be retrieved more than once. The
coin should also have been registered as a field element before doing this
call. Will also trigger the `goNextRoundLogic` if appropriate.
*/
func (run *ProverRuntime) GetRandomCoinField(name coin.Name) field.Element {
	return run.getRandomCoinGeneric(name, coin.Field).(field.Element)
}

/*
Returns a pre-sampled integer vec random coin. The coin should be issued at
the same round as it was registered. The same coin can't be retrieved more
than once. The coin should also have been registered as an integer vec before
doing this call. Will also trigger the `goNextRoundLogic` if appropriate.
*/
func (run *ProverRuntime) GetRandomCoinIntegerVec(name coin.Name) []int {
	return run.getRandomCoinGeneric(name, coin.IntegerVec).([]int)
}

/*
Assigns a column to the oracle
*/
func (run *ProverRuntime) AssignColumn(name ifaces.ColID, witness ifaces.ColAssignment) {

	// global prover's lock before accessing the witnesses
	run.lock.Lock()
	defer run.lock.Unlock()

	// Sanity-check : the handle should not be empty
	if len(name) == 0 {
		panic("given an empty name")
	}

	if run.Spec.Columns.Status(name) == column.VerifierDefined {
		utils.Panic("tried assign a verifier defined column : %v", name)
	}

	// Make sure, it is done at the right round
	handle := run.Spec.Columns.GetHandle(name)
	ifaces.MustBeInRound(handle, run.currRound)

	if witness.Len() != handle.Size() {
		utils.Panic("Bad length for %v, expected %v got %v\n", handle, handle.Size(), witness.Len())
	}

	// This is supposed to be redundant with the checks we make
	// when registering a column. So, if it fails here it likely
	// should have failed ealier
	if !utils.IsPowerOfTwo(witness.Len()) {
		utils.Panic("Witness with non-power of two sizes, should have been caught earlier")
	}

	// Adds it to the assignments
	run.Columns.InsertNew(handle.GetColID(), witness)
}

/*
Internal common function for getting random coins of any types
*/
func (run *ProverRuntime) getRandomCoinGeneric(name coin.Name, requestedType coin.Type) interface{} {
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

/*
Internal function handling the transition to the next round.
*/
func (run *ProverRuntime) goNextRound() {

	/*
		Make sure all issued random coin have been "consumed" by all the prover
		steps, in the round we are closing. An error occuring here is more likely
		an error in the compiler than an error from the user because it is not
		responsible for setting the coin. Thus, this is more a sanity check.
	*/
	toBeConsumed := run.Spec.Coins.AllKeysAt(run.currRound)
	run.Coins.MustExists(toBeConsumed...)

	/*
		We do not make this check for the columns, the reason is that we delete
		the columns that we do not use anymore.
	*/

	/*
		Then, make sure all the query parameters have been set
		during the rounds we are closing
	*/
	toBeParametrized := run.Spec.QueriesParams.AllKeysAt(run.currRound)
	run.QueriesParams.MustExists(toBeParametrized...)

	// Counts the transcript size of the round and the number of field
	// element generated.
	initialTranscriptSize := run.FS.TranscriptSize
	initialNumCoinsGenerated := run.FS.NumCoinGenerated

	/*
		Make sure that all messages have been written and use them
		to update the FS state.  Note that we do not need to update
		FS using the last round of the prover because he is always
		the last one to "talk" in the protocol.
	*/
	start := run.FS.TranscriptSize
	msgsToFS := run.Spec.Columns.AllKeysProofAt(run.currRound)
	for _, msgName := range msgsToFS {
		instance := run.GetMessage(msgName)
		run.FS.UpdateSV(instance)
	}
	logrus.Infof("Fiat-shamir round %v - %v proof elements in the transcript", run.currRound, run.FS.TranscriptSize-start)

	/*
		Make sure that all messages have been written and use them
		to update the FS state.  Note that we do not need to update
		FS using the last round of the prover because he is always
		the last one to "talk" in the protocol.
	*/
	start = run.FS.TranscriptSize
	msgsToFS = run.Spec.Columns.AllKeysPublicInputAt(run.currRound)
	for _, msgName := range msgsToFS {
		instance := run.GetMessage(msgName)
		run.FS.UpdateSV(instance)
	}
	logrus.Infof("Fiat-shamir round %v - %v public inputs in the transcript", run.currRound, run.FS.TranscriptSize-start)

	/*
		Also include the prover's allegations for all evaluations
	*/
	start = run.FS.TranscriptSize
	paramsToFS := run.Spec.QueriesParams.AllKeysAt(run.currRound)
	for _, qName := range paramsToFS {
		// Implicitly, this will panic whenever we start supporting
		// a new type of query params
		params := run.QueriesParams.MustGet(qName)
		params.UpdateFS(run.FS)
	}
	logrus.Infof("Fiat-shamir round %v - %v query params in the transcript", run.currRound, run.FS.TranscriptSize-start)

	// Increment the number of rounds
	run.currRound++

	/*
		Then assigns the coins for the new round. As the round
		incrementation is made lazily, we expect that there is
		a next round.
	*/
	toCompute := run.Spec.Coins.AllKeysAt(run.currRound)
	for _, coin := range toCompute {
		info := run.Spec.Coins.Data(coin)
		value := info.Sample(run.FS)
		run.Coins.InsertNew(coin, value)
	}

	logrus.Infof("Ran Fiat-Shamir for round %v, transcript size %v (field element), generated %v field elements, total-transcript %v, total-generated %v",
		run.currRound, run.FS.TranscriptSize-initialTranscriptSize, run.FS.NumCoinGenerated-initialNumCoinsGenerated, run.FS.TranscriptSize, run.FS.NumCoinGenerated,
	)
}

/*
Run the compiled prover's steps of the current round
*/
func (run *ProverRuntime) runProverSteps() {
	// Run all the assigners
	subProverSteps := run.Spec.SubProvers.MustGet(run.currRound)
	for _, step := range subProverSteps {
		step(run)
	}
}

/*
Get a message sent to the verifier
*/
func (run *ProverRuntime) GetMessage(name ifaces.ColID) ifaces.ColAssignment {
	// Global prover locks for accessing the maps
	run.lock.Lock()
	defer run.lock.Unlock()

	// Sanity-check, this panics if the column does not exists
	return run.Columns.MustGet(name)
}

// Get an Inner-product query
func (run *ProverRuntime) GetInnerProduct(name ifaces.QueryID) query.InnerProduct {
	return run.Spec.GetInnerProduct(name)
}

// Returns pre-assigned parameters for the current query
func (run *ProverRuntime) GetInnerProductParams(name ifaces.QueryID) query.InnerProductParams {
	return run.QueriesParams.MustGet(name).(query.InnerProductParams)
}

// Assign the result of an inner-product query
func (run *ProverRuntime) AssignInnerProduct(name ifaces.QueryID, ys ...field.Element) query.InnerProductParams {
	q := run.GetInnerProduct(name)
	if len(q.Bs) != len(ys) {
		utils.Panic("Inner-product query %v has %v bs but assigned for %v", name, len(q.Bs), len(ys))
	}

	param := query.NewInnerProductParams(ys...)
	run.QueriesParams.InsertNew(name, param)
	return param
}

/*
Assign evaluation point and claimed values for a univariate evaluation
*/
func (run *ProverRuntime) AssignUnivariate(name ifaces.QueryID, x field.Element, ys ...field.Element) {

	// Global prover locks for accessing the maps
	run.lock.Lock()
	defer run.lock.Unlock()

	// Make sure, it is done at the right round
	run.Spec.QueriesParams.MustBeInRound(run.currRound, name)

	// Check the length of ys
	q := run.Spec.QueriesParams.Data(name).(query.UnivariateEval)
	if len(q.Pols) != len(ys) {
		utils.Panic("Query expected ys = %v but got %v", len(q.Pols), len(ys))
	}
	// Adds it to the assignments
	params := query.NewUnivariateEvalParams(x, ys...)
	run.QueriesParams.InsertNew(name, params)
}

/*
Get univariate eval metadata. Panic if not found.
*/
func (run *ProverRuntime) GetUnivariateEval(name ifaces.QueryID) query.UnivariateEval {
	// Global prover locks for accessing the maps
	run.lock.Lock()
	defer run.lock.Unlock()
	return run.Spec.QueriesParams.Data(name).(query.UnivariateEval)
}

// Returns the parameters of a univariate evaluation (i.e: x, the evaluation point)
// and y, the alleged polynomial opening.
func (run *ProverRuntime) GetUnivariateParams(name ifaces.QueryID) query.UnivariateEvalParams {

	// Global prover's lock for accessing params
	run.lock.Lock()
	defer run.lock.Unlock()

	return run.QueriesParams.MustGet(name).(query.UnivariateEvalParams)
}

// Assign evaluation point and claimed values for a univariate evaluation
func (run *ProverRuntime) AssignLocalPoint(name ifaces.QueryID, y field.Element) {

	// Global prover locks for accessing the maps
	run.lock.Lock()
	defer run.lock.Unlock()

	// Make sure, it is done at the right round
	run.Spec.QueriesParams.MustBeInRound(run.currRound, name)

	// Adds it to the assignments
	params := query.NewLocalOpeningParams(y)
	run.QueriesParams.InsertNew(name, params)
}

// Get univariate eval metadata. Panic if not found.
func (run *ProverRuntime) GetLocalPointEval(name ifaces.QueryID) query.LocalOpening {
	// Global prover locks for accessing the maps
	run.lock.Lock()
	defer run.lock.Unlock()
	return run.Spec.QueriesParams.Data(name).(query.LocalOpening)
}

// Returns the parameters of a univariate evaluation (i.e: x, the evaluation point)
// and y, the alleged polynomial opening.
func (run *ProverRuntime) GetLocalPointEvalParams(name ifaces.QueryID) query.LocalOpeningParams {

	// Global prover's lock for accessing params
	run.lock.Lock()
	defer run.lock.Unlock()

	return run.QueriesParams.MustGet(name).(query.LocalOpeningParams)
}

// Generic function to extract the parameters of a query. Will panic if no
// parameters are found
func (run *ProverRuntime) GetParams(name ifaces.QueryID) ifaces.QueryParams {
	return run.QueriesParams.MustGet(name)
}
