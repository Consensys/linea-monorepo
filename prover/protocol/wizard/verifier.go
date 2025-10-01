package wizard

import (
	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/gnark-crypto/field/koalabear/poseidon2"
	"github.com/consensys/gnark-crypto/hash"
	"github.com/consensys/linea-monorepo/prover/crypto/fiatshamir"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/protocol/accessors"
	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/zk"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/collection"
)

// Proof generically represents a proof obtained from the wizard. This object does not
// implement any logic and only serves as a registry for all the prover messages
// that are assigned by the prover runtime and that are necessary to run
// the verifier. It includes the assignment of all the columns that are visible
// to the verifier; meaning all columns bearing the tag [column.Proof] and the
// query parameters [ifaces.QueryParams] provided by the prover runtime.
//
// The proof can be constructed using the [Prove] function and can be
// used as an input to the [Verify] function. It can also be used to
// assign a [VerifierCircuit] in order to recursively compose
// the proof within a gnark circuit.
//
// The struct does not implement any serialization logic.
type Proof[T zk.Element] struct {
	// Messages collection of the prover's message sent to the verifier.
	Messages collection.Mapping[ifaces.ColID, ifaces.ColAssignment]

	// QueriesParams stores all the query parameters (i.e) the messages of the
	// oracle to the verifier.
	QueriesParams collection.Mapping[ifaces.QueryID, ifaces.QueryParams]
}

// Runtime is a generic interface extending the [ifaces.Runtime] interface
// with all methods of [wizard.VerifierRuntime[T]]. This is used to allow the
// writing of adapters for the verifier runtime.
type Runtime[T zk.Element] interface {
	ifaces.Runtime
	GetSpec() *CompiledIOP[T]
	GetPublicInput(name string) field.Element
	GetGrandProductParams(name ifaces.QueryID) query.GrandProductParams
	GetHornerParams(name ifaces.QueryID) query.HornerParams[T]
	GetLogDerivSumParams(name ifaces.QueryID) query.LogDerivSumParams[T]
	GetLocalPointEvalParams(name ifaces.QueryID) query.LocalOpeningParams[T]
	GetInnerProductParams(name ifaces.QueryID) query.InnerProductParams[T]
	GetUnivariateEval(name ifaces.QueryID) query.UnivariateEval[T]
	GetUnivariateParams(name ifaces.QueryID) query.UnivariateEvalParams[T]
	GetQuery(name ifaces.QueryID) ifaces.Query[T]
	Fs() hash.StateStorer
	InsertCoin(name coin.Name, value any)
	GetState(name string) (any, bool)
	SetState(name string, value any)
}

// VerifierStep specifies a single step of verifier for a single subprotocol.
// This can be used to specify verifier checks involving user-provided
// columns for relations that cannot be automatically enforced via a
// [ifaces.Query[T]]
type VerifierStep[T zk.Element] func(a Runtime[T]) error

// VerifierRuntime[T] runtime collects all data that visible or computed by the
// verifier of the wizard protocol. This includes the prover's messages, the
// [column.VerifyingKey] tagged columns.
//
// The struct is not intended to be constructed by the user and is internally
// constructed by the [Verify] function. The user should instead
// restricts its usage of the function within [VerifierStep] functions that are
// provided to either the [CompiledIOP[T]] or the [Verify] function.
type VerifierRuntime[T zk.Element] struct {

	// Spec points to the static description of the underlying protocol
	Spec *CompiledIOP[T]

	// Collection of the prover's message sent to the verifier.
	Columns collection.Mapping[ifaces.ColID, ifaces.ColAssignment]

	// Coins stores all the random coins issued during the protocol
	Coins collection.Mapping[coin.Name, interface{}]

	// Stores all the query parameters (i.e) the messages of the oracle to the
	// verifier.
	QueriesParams collection.Mapping[ifaces.QueryID, ifaces.QueryParams]

	// FS stores the Fiat-Shamir State, you probably don't want to use it
	// directly unless you know what you are doing. Just know that if you use
	// it to update the FS hash, this can potentially result in the prover and
	// the verifer end up having different state or the same message being
	// included a second time. Use it externally at your own risks.
	FS hash.StateStorer

	// State stores arbitrary data that can be used by the verifier. This
	// can be used to communicate values between verifier states.
	State map[string]interface{}
}

// Verify verifies a wizard proof. The caller specifies a [CompiledIOP[T]] that
// describes the protocol to run and a proof to verify. The function returns
// `nil` to indicate that the proof passed and an error to indicate the proof
// was invalid.
func Verify[T zk.Element](c *CompiledIOP[T], proof Proof[T]) error {
	_, err := VerifyWithRuntime(c, proof)
	return err
}

// VerifyWithRuntime runs the verifier of the protocol and returns the result
// and the runtime of the verifier.
func VerifyWithRuntime[T zk.Element](c *CompiledIOP[T], proof Proof[T]) (*VerifierRuntime[T], error) {
	return verifyWithRuntimeUntilRound(c, proof, c.NumRounds())
}

// VerifyUntilRound runs the verifier up to a specified round
func VerifyUntilRound[T zk.Element](c *CompiledIOP[T], proof Proof[T], stopRound int) error {
	_, err := verifyWithRuntimeUntilRound(c, proof, stopRound)
	return err
}

// verifyWithRuntimeUntilRound runs the verifier of 'comp' up to (and excluding)
// the provided round "stopRound". By "excluding", we mean that the function
// won't run the round "stopRound". If stopRound is higher than the number of
// rounds in comp, the function runs the whole protocol.
func verifyWithRuntimeUntilRound[T zk.Element](comp *CompiledIOP[T], proof Proof[T], stopRound int) (run *VerifierRuntime[T], err error) {

	var (
		runtime = comp.createVerifier(proof)
		errs    = []error{}
	)

	stopRound = min(stopRound, comp.NumRounds())

	for round := 0; round < stopRound; round++ {

		runtime.GenerateCoinsFromRound(round)

		verifierSteps := runtime.Spec.SubVerifiers.MustGet(round)
		for _, step := range verifierSteps {
			if err := step.Run(&runtime); err != nil {
				errs = append(errs, err)
			}
		}
	}

	if len(errs) > 0 {
		return &runtime, utils.WrapErrsAlphabetically(errs)
	}

	return &runtime, nil
}

// createVerifier is an internal constructor for a new empty [VerifierRuntime[T]] runtime. It
// prepopulates the [VerifierRuntime[T].Columns] and [VerifierRuntime[T].QueriesParams]
// with the one that are in the proof. It also populates its verifier steps from
// the `VerifierFuncGen` in the `c`. The user also passes the list of prover
// messages. It is internally called by the [Verify] function.
func (c *CompiledIOP[T]) createVerifier(proof Proof[T]) VerifierRuntime[T] {

	/*
		Instantiate an empty assigment for the verifier
	*/
	runtime := VerifierRuntime[T]{
		Spec:          c,
		Coins:         collection.NewMapping[coin.Name, interface{}](),
		Columns:       proof.Messages,
		QueriesParams: proof.QueriesParams,
		FS:            poseidon2.NewMerkleDamgardHasher(),
		State:         make(map[string]interface{}),
	}

	fiatshamir.Update(runtime.FS, c.FiatShamirSetup)

	/*
		Insert the verifying key into the messages
	*/
	for _, name := range c.Columns.AllVerifyingKey() {
		val := c.Precomputed.MustGet(name)
		runtime.Columns.InsertNew(name, val)
	}

	return runtime
}

// GetPublicInput extracts the value of a public input from the proof.
func (proof Proof[T]) GetPublicInput(comp *CompiledIOP[T], name string) field.Element {

	publicInputsAccessor := comp.GetPublicInputAccessor(name)

	switch a := publicInputsAccessor.(type) {
	case *accessors.FromConstAccessor[T]:
		if a.IsBase() {
			return a.Base
		} else {
			panic("Requested a base element from a public input that is a field extension")
		}
	case *accessors.FromPublicColumn[T]:
		if a.Col.Status() == column.Proof {
			return proof.Messages.MustGet(a.Col.ID).Get(a.Pos)
		}
	case *accessors.FromLocalOpeningYAccessor[T]:
		return proof.QueriesParams.MustGet(a.Q.ID).(query.LocalOpeningParams[T]).BaseY
	}

	// This generically returns the value of a public input by extracting
	// it from the runtime of the verifier. This is inefficient because it
	// needs to run the verifier to extract the value. So this behaviour
	// should be used only for types of [ifaces.Accessor] who need it.
	//
	// These are not directly visible from the proof. Thus we need to
	// run the verifier and extract them from the runtime.
	verifierRuntime, _ := VerifyWithRuntime(comp, proof)
	return verifierRuntime.GetPublicInput(name)
}

// GenerateCoinsFromRound generates all the random coins for the given round.
// It does so by updating the FS with all the prover messages from round-1
// and then generating all the coins for the current round.
func (run *VerifierRuntime[T]) GenerateCoinsFromRound(currRound int) {

	if currRound > 0 {

		if !run.Spec.DummyCompiled {

			/*
				Make sure that all messages have been written and use them
				to update the FS state.  Note that we do not need to update
				FS using the last round of the prover because he is always
				the last one to "talk" in the protocol.
			*/
			msgsToFS := run.Spec.Columns.AllKeysProofAt(currRound - 1)
			for _, msgName := range msgsToFS {

				if run.Spec.Columns.IsExplicitlyExcludedFromProverFS(msgName) {
					continue
				}

				instance := run.GetColumn(msgName)
				fiatshamir.UpdateSV(run.FS, instance)
			}

			/*
				Also include the prover's allegations for all evaluations
			*/
			queries := run.Spec.QueriesParams.AllKeysAt(currRound - 1)
			for _, qName := range queries {
				if run.Spec.QueriesParams.IsSkippedFromVerifierTranscript(qName) {
					continue
				}

				params := run.QueriesParams.MustGet(qName)
				params.UpdateFS(run.FS)
			}
		}
	}

	if run.Spec.FiatShamirHooksPreSampling.Len() > currRound {
		fsHooks := run.Spec.FiatShamirHooksPreSampling.MustGet(currRound)
		for i := range fsHooks {
			// if fsHooks[i].IsSkipped() {
			// 	continue
			// }

			fsHooks[i].Run(run)
		}
	}

	seed := run.FS.State()
	var kSeed koalabear.Element
	kSeed.SetBytes(seed[:1]) // TODO @thomas why take only the first byte ?

	/*
		Then assigns the coins for the new round. As the round incrementation
		is made lazily, we expect that there is a next round.
	*/
	toCompute := run.Spec.Coins.AllKeysAt(currRound)
	for _, myCoin := range toCompute {
		if run.Spec.Coins.IsSkippedFromVerifierTranscript(myCoin) {
			continue
		}

		info := run.Spec.Coins.Data(myCoin)
		value := info.Sample(run.FS, kSeed)
		run.Coins.InsertNew(myCoin, value)
	}
}

// GetRandomCoinField returns a field element random. The coin should be issued
// at the same round as it was registered. The same coin can't be retrieved more
// than once. The coin should also have been registered as a field element
// before doing this call. Will also trigger the "goNextRound" logic if
// appropriate.
func (run *VerifierRuntime[T]) GetRandomCoinField(name coin.Name) field.Element {
	/*
		Early check, ensures the coin has been registered at all
		and that it has the correct type
	*/
	infos := run.Spec.Coins.Data(name)
	if infos.Type != coin.Field && infos.Type != coin.FieldFromSeed && infos.Type != coin.FieldExt {
		utils.Panic("Coin %v was registered with type %v but got %v", name, infos.Type, coin.Field)
	}
	// If this panics, it means we generates the coins wrongly
	return run.Coins.MustGet(name).(field.Element)
}

func (run *VerifierRuntime[T]) GetRandomCoinFieldExt(name coin.Name) fext.Element {
	infos := run.Spec.Coins.Data(name)
	if infos.Type != coin.FieldExt {
		utils.Panic("Coin was registered as %v but got %v", infos.Type, coin.FieldExt)
	}
	// If this panics, it means we generates the coins wrongly
	return run.Coins.MustGet(name).(fext.Element)
}

// GetRandomCoinFromSeed returns a field element random based on the seed.
func (run *VerifierRuntime[T]) GetRandomCoinFromSeed(name coin.Name) field.Element {
	/*
		Early check, ensures the coin has been registered at all
		and that it has the correct type
	*/
	infos := run.Spec.Coins.Data(name)
	if infos.Type != coin.FieldFromSeed {
		utils.Panic("Coin was registered as %v but expected %v", infos.Type, coin.FieldFromSeed)
	}
	// If this panics, it means we generates the coins wrongly
	return run.Coins.MustGet(name).(field.Element)
}

// GetRandomCoinIntegerVec returns a pre-sampled integer vec random coin. The
// coin should be issued at the same round as it was registered. The same coin
// can't be retrieved more than once. The coin should also have been registered
// as an integer vec before doing this call. Will also trigger the
// "goNextRound" logic if appropriate.
func (run *VerifierRuntime[T]) GetRandomCoinIntegerVec(name coin.Name) []int {
	/*
		Early check, ensures the coin has been registered at all
		and that it has the correct type
	*/
	infos := run.Spec.Coins.Data(name)
	if infos.Type != coin.IntegerVec {
		utils.Panic("Coin was registered as %v but got %v", infos.Type, coin.IntegerVec)
	}
	// If this panics, it means we generates the coins wrongly
	return run.Coins.MustGet(name).([]int)
}

// GetUnivariateParams returns the parameters of a univariate evaluation (i.e:
// x, the evaluation point) and y, the alleged polynomial opening. This is
// intended to resolve parameters that have been provided by the proof.
func (run *VerifierRuntime[T]) GetUnivariateParams(name ifaces.QueryID) query.UnivariateEvalParams[T] {
	return run.QueriesParams.MustGet(name).(query.UnivariateEvalParams[T])
}

/*
Returns the number of rounds in the assignment.
Deprecated: get it from the CompiledIOP[T] instead
*/
func (run *VerifierRuntime[T]) NumRounds() int {
	/*
		Getting it from the spec is the safest as it is already
		tested. We could fit more assertions here nonetheless.
	*/
	return run.Spec.NumRounds()
}

/*
GetUnivariateEval returns a registered [query.UnivariateEval[T]]. Panic if not found.
Deprecated: get it from the CompiledIOP[T] instead
*/
func (run *VerifierRuntime[T]) GetUnivariateEval(name ifaces.QueryID) query.UnivariateEval[T] {
	return run.Spec.QueriesParams.Data(name).(query.UnivariateEval[T])
}

// GetColumn returns a column by name. The status of the columns must be
// either proof or public input and the column must be visible to the verifier
// and consequently be available in the proof.
func (run *VerifierRuntime[T]) GetColumn(name ifaces.ColID) ifaces.ColAssignment {

	msgStatus := run.Spec.Columns.Status(name)

	// Sanity-check : the verifier may only access public columns.
	// In case it was not it would be caught by the MustGet below
	// but it's cleaner if the panic happens before.
	if !msgStatus.IsPublic() {
		utils.Panic("the verifier attempted to get message : %v (status %v)", name, msgStatus.String())
	}

	msgIFace := run.Columns.MustGet(name)

	// Just a sanity-check to ensure the message has the right size
	expectedSize := run.Spec.Columns.GetSize(name)
	if msgIFace.Len() != expectedSize {
		utils.Panic("bad dimension %v, spec expected %v, column-name %v", msgIFace.Len(), expectedSize, name)
	}

	return msgIFace
}

// GetInnerProductParams returns the parameters of an inner-product query
// [query.InnerProduct] provided by the proof. The function will panic if the
// query does not exist or if the parameters are not available in the proof.
func (run *VerifierRuntime[T]) GetInnerProductParams(name ifaces.QueryID) query.InnerProductParams[T] {
	return run.QueriesParams.MustGet(name).(query.InnerProductParams[T])
}

// GetLocalPointEvalParams returns the parameters of a [query.LocalOpening]
// query  (i.e: y, the alleged opening of the query's column at the first
// position.
func (run *VerifierRuntime[T]) GetLocalPointEvalParams(name ifaces.QueryID) query.LocalOpeningParams[T] {
	return run.QueriesParams.MustGet(name).(query.LocalOpeningParams[T])
}

// GetLogDerivSumParams returns the parameters of a [query.LogDerivativeSum]
func (run *VerifierRuntime[T]) GetLogDerivSumParams(name ifaces.QueryID) query.LogDerivSumParams[T] {
	return run.QueriesParams.MustGet(name).(query.LogDerivSumParams[T])
}

// GetGrandProductParams returns the parameters of a [query.GrandProduct]
func (run *VerifierRuntime[T]) GetGrandProductParams(name ifaces.QueryID) query.GrandProductParams {
	return run.QueriesParams.MustGet(name).(query.GrandProductParams)
}

// GetHornerParams returns the parameters of a [query.Honer] query.
func (run *VerifierRuntime[T]) GetHornerParams(name ifaces.QueryID) query.HornerParams[T] {
	return run.QueriesParams.MustGet(name).(query.HornerParams[T])
}

/*
CopyColumnInto implements `column.GetWitness`
Copies the witness into a slice

Deprecated: this is deadcode
*/
func (run VerifierRuntime[T]) CopyColumnInto(name ifaces.ColID, buff *ifaces.ColAssignment) {
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

// GetColumnAt returns the value of a verifier [ifaces.Column] at a specified
// position. This is needed to implement the [column.GetWitness] interface and
// it will only work if the requested column is part of the proof the verifier
// is running on.
func (run VerifierRuntime[T]) GetColumnAt(name ifaces.ColID, pos int) field.Element {
	/*
		Make sure the column is registered. If the name is the one specified
		does not correcpond to a natural column, this will panic. And this is
		expected behaviour.
	*/
	run.Spec.Columns.MustHaveName(name)
	wit := run.Columns.MustGet(name)

	if pos >= wit.Len() || pos < 0 {
		utils.Panic("asked pos %v for vector of size %v", pos, wit.Len())
	}

	return wit.Get(pos)
}

func (run *VerifierRuntime[T]) GetColumnAtBase(name ifaces.ColID, pos int) (field.Element, error) {
	run.Spec.Columns.MustHaveName(name)
	wit := run.Columns.MustGet(name)

	if pos >= wit.Len() || pos < 0 {
		utils.Panic("asked pos %v for vector of size %v", pos, wit)
	}

	if _, err := wit.GetBase(0); err == nil {
		return wit.GetBase(pos)
	} else {
		return field.Zero(), err
	}

}
func (run *VerifierRuntime[T]) GetColumnAtExt(name ifaces.ColID, pos int) fext.Element {
	run.Spec.Columns.MustHaveName(name)
	wit := run.Columns.MustGet(name)

	if pos >= wit.Len() || pos < 0 {
		utils.Panic("asked pos %v for vector of size %v", pos, wit)
	}
	return wit.GetExt(pos)
}

// GetParams extracts the parameters of a query. Will panic if no
// parameters are found
//
// Deprecated: there are already methods to return parameters with an explicit
// type.
func (run *VerifierRuntime[T]) GetParams(name ifaces.QueryID) ifaces.QueryParams {
	return run.QueriesParams.MustGet(name)
}

// GetPublicInput returns a public input from its name
func (run *VerifierRuntime[T]) GetPublicInput(name string) field.Element {
	allPubs := run.Spec.PublicInputs
	for i := range allPubs {
		if allPubs[i].Name == name {
			return allPubs[i].Acc.GetVal(run)
		}
	}
	utils.Panic("could not find public input nb %v", name)
	return field.Element{}
}

// Fs returns the Fiat-Shamir state
func (run *VerifierRuntime[T]) Fs() hash.StateStorer {
	return run.FS
}

// GetSpec returns the compiled IOP
func (run *VerifierRuntime[T]) GetSpec() *CompiledIOP[T] {
	return run.Spec
}

// InsertCoin inserts a coin into the runtime. It should not be
// used by usual verifier action but is useful when implementing
// recursion utilities.
func (run *VerifierRuntime[T]) InsertCoin(name coin.Name, value any) {
	run.Coins.InsertNew(name, value)
}

// GetState returns an arbitrary value stored in the runtime
func (run *VerifierRuntime[T]) GetState(name string) (any, bool) {
	res, ok := run.State[name]
	return res, ok
}

// SetState sets an arbitrary value in the runtime
func (run *VerifierRuntime[T]) SetState(name string, value any) {
	run.State[name] = value
}

// GetQuery returns a query from its name
func (run *VerifierRuntime[T]) GetQuery(name ifaces.QueryID) ifaces.Query[T] {

	if run.Spec.QueriesParams.Exists(name) {
		return run.Spec.QueriesParams.Data(name)
	}

	if run.Spec.QueriesNoParams.Exists(name) {
		return run.Spec.QueriesNoParams.Data(name)
	}

	utils.Panic("could not find query nb %v", name)
	return nil
}
