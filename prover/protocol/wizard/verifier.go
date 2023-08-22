package wizard

import (
	"fmt"

	"github.com/consensys/accelerated-crypto-monorepo/crypto/fiatshamir"
	"github.com/consensys/accelerated-crypto-monorepo/maths/common/smartvectors"
	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/coin"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/ifaces"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/query"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
	"github.com/consensys/accelerated-crypto-monorepo/utils/collection"
	"github.com/sirupsen/logrus"
)

// Represents generically a proof.
type Proof struct {
	// Collection of the prover's message sent to the verifier.
	// In short, it is the proof.
	Messages collection.Mapping[ifaces.ColID, ifaces.ColAssignment]

	// Stores all the query parameters (i.e) the messages of the
	// oracle to the verifier.
	QueriesParams collection.Mapping[ifaces.QueryID, ifaces.QueryParams]
}

// Describes a single step of verifier for a single subprotocol.
type VerifierStep func(a *VerifierRuntime) error

// VerifierRuntime runtime gathers all data for the verifier
type VerifierRuntime struct {
	/*
		A static description of the protocol
	*/
	Spec *CompiledIOP
	/*
		Collection of the prover's message sent to the verifier.
		In short, it is the proof.
	*/
	Columns collection.Mapping[ifaces.ColID, ifaces.ColAssignment]
	/*
		Stores all the random coins issued during the protocol
	*/
	Coins collection.Mapping[coin.Name, interface{}]
	/*
		Stores all the query parameters (i.e) the messages of the
		oracle to the verifier.
	*/
	QueriesParams collection.Mapping[ifaces.QueryID, ifaces.QueryParams]
	/*
		It is supposed to be an internal and private attribute.

		Fiat-Shamir State, you probably don't want to use it directly unless
		you know what you are doing. Just know that if you use it to update
		the FS hash, this can potentially result in the prover and the verifer
		end up having different state or the same message being included a second
		time.
	*/
	FS *fiatshamir.State
}

// Top-level function to pass an assignment
func Verify(c *CompiledIOP, proof Proof) error {

	runtime := c.CreateVerifier(proof)

	/*
		Pre-emptively generates the random coins. As the entire set of prover
		messages is available at once. We can do it upfront, as opposed to the
		prover's implementation.
	*/
	runtime.generateAllRandomCoins()

	/*
		And run all the precompiled rounds. Collecting the errors if there are any
	*/
	anyError := false
	errMsg := ""
	for _, roundSteps := range runtime.Spec.subVerifiers.Inner() {
		for _, step := range roundSteps {
			if err := step(&runtime); err != nil {
				errMsg += fmt.Sprintf("\t%v\n", err)
				anyError = true
			}
		}
	}

	if anyError {
		return fmt.Errorf("verifier failed : \n%v", errMsg)
	}

	return nil
}

/*
Construct a new empty verifier runtime and populates its verifier
steps from the `VerifierFuncGen` in the `c`. The user also passes
the list of prover messages (which can also be pronounced "proof")
*/
func (c *CompiledIOP) CreateVerifier(proof Proof) VerifierRuntime {

	/*
		Instantiate an empty assigment for the verifier
	*/
	runtime := VerifierRuntime{
		Spec:          c,
		Coins:         collection.NewMapping[coin.Name, interface{}](),
		Columns:       proof.Messages,
		QueriesParams: proof.QueriesParams,
		FS:            fiatshamir.NewMiMCFiatShamir(),
	}

	/*
		Insert the verifying key into the messages
	*/
	for _, name := range c.Columns.AllVerifyingKey() {
		val := c.Precomputed.MustGet(name)
		runtime.Columns.InsertNew(name, val)
	}

	return runtime
}

/*
Generates all the random coins at the beginning
*/
func (run *VerifierRuntime) generateAllRandomCoins() {

	for currRound := 0; currRound < run.Spec.NumRounds(); currRound++ {
		if currRound > 0 {
			/*
				Sanity-check : Make sure all issued random coin have been
				"consumed" by all the verifiers steps, in the round we are
				"closing"
			*/
			toBeConsumed := run.Spec.Coins.AllKeysAt(currRound - 1)
			run.Coins.Exists(toBeConsumed...)

			/*
				Make sure that all messages have been written and use them
				to update the FS state.  Note that we do not need to update
				FS using the last round of the prover because he is always
				the last one to "talk" in the protocol.
			*/
			msgsToFS := run.Spec.Columns.AllKeysProofAt(currRound - 1)
			for _, msgName := range msgsToFS {
				instance := run.GetColumn(msgName)
				logrus.Tracef("VERIFIER : Update fiat-shamir with proof message %v", msgName)
				run.FS.UpdateSV(instance)
			}

			msgsToFS = run.Spec.Columns.AllKeysPublicInputAt(currRound - 1)
			for _, msgName := range msgsToFS {
				instance := run.GetColumn(msgName)
				logrus.Tracef("VERIFIER : Update fiat-shamir with public input %v", msgName)
				run.FS.UpdateSV(instance)
			}

			/*
				Also include the prover's allegations for all evaluations
			*/
			queries := run.Spec.QueriesParams.AllKeysAt(currRound - 1)
			for _, qName := range queries {
				// Implicitly, this will panic whenever we start supporting
				// a new type of query params
				logrus.Tracef("VERIFIER : Update fiat-shamir with query parameters %v", qName)
				params := run.QueriesParams.MustGet(qName)
				params.UpdateFS(run.FS)
			}
		}

		/*
			Then assigns the coins for the new round. As the round incrementation
			is made lazily, we expect that there is a next round.
		*/
		toCompute := run.Spec.Coins.AllKeysAt(currRound)
		for _, coin := range toCompute {
			logrus.Tracef("VERIFIER : Generate coin %v", coin)
			info := run.Spec.Coins.Data(coin)
			value := info.Sample(run.FS)
			run.Coins.InsertNew(coin, value)
		}
	}

}

/*
Returns a field element random a preassigned random coin as field element.
The implementation implicitly checks that the field element is of the right typ
*/
func (run *VerifierRuntime) GetRandomCoinField(name coin.Name) field.Element {
	/*
		Early check, ensures the coin has been registered at all
		and that it has the correct type
	*/
	infos := run.Spec.Coins.Data(name)
	if infos.Type != coin.Field {
		utils.Panic("Coin was registered as %v but got %v", infos.Type, coin.Field)
	}
	// If this panics, it means we generates the coins wrongly
	return run.Coins.MustGet(name).(field.Element)
}

/*
Returns a pre-sampled integer vec random coin.
*/
func (run *VerifierRuntime) GetRandomCoinIntegerVec(name coin.Name) []int {
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

/*
Returns the parameters of a univariate evaluation (i.e: x, the evaluation point)
and y, the alleged polynomial opening.
*/
func (run *VerifierRuntime) GetUnivariateParams(name ifaces.QueryID) query.UnivariateEvalParams {
	return run.QueriesParams.MustGet(name).(query.UnivariateEvalParams)
}

/*
Returns the number of rounds in the assignment.
*/
func (run *VerifierRuntime) NumRounds() int {
	/*
		Getting it from the spec is the safest as it is already
		tested. We could fit more assertions here nonetheless.
	*/
	return run.Spec.NumRounds()
}

/*
Get univariate eval metadata. Panic if not found.
*/
func (run *VerifierRuntime) GetUnivariateEval(name ifaces.QueryID) query.UnivariateEval {
	return run.Spec.QueriesParams.Data(name).(query.UnivariateEval)
}

/*
Returns a column by name. The status of the columns should be either proof or public input
*/
func (run *VerifierRuntime) GetColumn(name ifaces.ColID) ifaces.ColAssignment {

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
		utils.Panic("bad dimension %v, spec expected %v", msgIFace.Len(), expectedSize)
	}

	return msgIFace
}

// Returns pre-assigned parameters for the current query
func (run *VerifierRuntime) GetInnerProductParams(name ifaces.QueryID) query.InnerProductParams {
	return run.QueriesParams.MustGet(name).(query.InnerProductParams)
}

/*
Returns the parameters of a univariate evaluation (i.e: x, the evaluation point)
and y, the alleged polynomial opening.
*/
func (run *VerifierRuntime) GetLocalPointEvalParams(name ifaces.QueryID) query.LocalOpeningParams {
	return run.QueriesParams.MustGet(name).(query.LocalOpeningParams)
}

/*
This implements `column.GetWitness`
Copies the witness into a slice
*/
func (run VerifierRuntime) CopyColumnInto(name ifaces.ColID, buff *ifaces.ColAssignment) {
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
func (run VerifierRuntime) GetColumnAt(name ifaces.ColID, pos int) field.Element {
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

// Generic function to extract the parameters of a query. Will panic if no
// parameters are found
func (run *VerifierRuntime) GetParams(name ifaces.QueryID) ifaces.QueryParams {
	return run.QueriesParams.MustGet(name)
}
