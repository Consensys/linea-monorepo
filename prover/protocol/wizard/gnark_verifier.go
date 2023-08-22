package wizard

import (
	"github.com/consensys/accelerated-crypto-monorepo/crypto/fiatshamir"
	"github.com/consensys/accelerated-crypto-monorepo/crypto/mimc/gkrmimc"
	"github.com/consensys/accelerated-crypto-monorepo/maths/common/smartvectors"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/coin"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/column"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/ifaces"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/query"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
	"github.com/consensys/accelerated-crypto-monorepo/utils/collection"
	"github.com/consensys/accelerated-crypto-monorepo/utils/gnarkutil"
	"github.com/consensys/gnark/frontend"
	"github.com/sirupsen/logrus"
)

/*
Gnark array, mirrorring the VerifierRuntime
*/
type WizardVerifierCircuit struct {
	Spec *CompiledIOP `gnark:"-"`
	// Maps a messages name to a position in the array
	MessagesIDs collection.Mapping[ifaces.ColID, int] `gnark:"-"`

	// Columns are arrays of frontend variables
	Columns [][]frontend.Variable `gnark:",secret"`

	// Maps a query's name to a position in the arrays below
	UnivariateParamsIDs collection.Mapping[ifaces.QueryID, int] `gnark:"-"`
	UnivariateParams    []query.GnarkUnivariateEvalParams       `gnark:",secret"`

	InnerProductIDs    collection.Mapping[ifaces.QueryID, int] `gnark:"-"`
	InnerProductParams []query.GnarkInnerProductParams         `gnark:",secret"`

	LocalOpeningIDs    collection.Mapping[ifaces.QueryID, int] `gnark:"-"`
	LocalOpeningParams []query.GnarkLocalOpeningParams         `gnark:",secret"`

	// FiatShamir state
	FS *fiatshamir.GnarkFiatShamir `gnark:"-"`

	// Also the coins
	Coins collection.Mapping[coin.Name, interface{}] `gnark:"-"`

	// Also the MiMC factory
	HasherFactory *gkrmimc.HasherFactory `gnark:"-"`
}

/*
Functions that can be registered in the CompiledIOP by the successive
compilation steps. They correspond to "precompiled" verification steps.
*/
type GnarkVerifierStep func(frontend.API, *WizardVerifierCircuit)

/*
Allocate the inner-slices of the verifier struct from a precompiled
IOP.
*/
func AllocateWizardCircuit(comp *CompiledIOP) (*WizardVerifierCircuit, error) {

	res := newWizardVerifierCircuit()

	for i, colName := range comp.Columns.AllKeys() {
		// filter the columns by status
		status := comp.Columns.Status(colName)
		if !status.IsPublic() {
			// the column is not public so it is not part of the proof
			continue

		}

		if status == column.VerifyingKey {
			// this are constant columns
			continue
		}

		logrus.Tracef("VERIFIER CIRCUIT : registering column %v (as %v) in circuit (#%v)", colName, status.String(), i)
		// deactivate the flag guarding against empty circuits
		size := comp.Columns.GetSize(colName)

		// Allocates the column in the circuit and indexes it
		colID := len(res.Columns)
		res.Columns = append(res.Columns, gnarkutil.AllocateSlice(size))
		res.MessagesIDs.InsertNew(colName, colID)
	}

	/*
		Allocate the queries params also. Note that AllKeys does give a
		deterministic order iteration and that's why we do not iterate
		on the map directly.
	*/
	for _, qName := range comp.QueriesParams.AllKeys() {

		/*
			Note that we do not filter out the "already compiled" queries
			here.
		*/
		qInfoIface := comp.QueriesParams.Data(qName)

		switch qInfo := qInfoIface.(type) {
		case query.UnivariateEval:
			// Note that nil is the default value for frontend.Variable
			res.UnivariateParamsIDs.InsertNew(qName, len(res.UnivariateParams))
			res.UnivariateParams = append(res.UnivariateParams, qInfo.GnarkAllocate())
		case query.InnerProduct:
			// Note that nil is the default value for frontend.Variable
			res.InnerProductIDs.InsertNew(qName, len(res.InnerProductParams))
			res.InnerProductParams = append(res.InnerProductParams, qInfo.GnarkAllocate())
		case query.LocalOpening:
			// Note that nil is the default value for frontend.Variable
			res.LocalOpeningIDs.InsertNew(qName, len(res.LocalOpeningParams))
			res.LocalOpeningParams = append(res.LocalOpeningParams, query.GnarkLocalOpeningParams{})
		}
	}

	res.Spec = comp
	return res, nil
}

/*
Generates the constraints to assess that the correctness of a wizard
transcript.
*/
func (c *WizardVerifierCircuit) Verify(api frontend.API) {
	c.HasherFactory = gkrmimc.NewHasherFactory(api)
	c.FS = fiatshamir.NewGnarkFiatShamir(api, c.HasherFactory)
	c.generateAllRandomCoins(api)

	logrus.Tracef("Generated the coins")

	for _, roundSteps := range c.Spec.gnarkSubVerifiers.Inner() {
		for _, step := range roundSteps {
			step(api, c)
		}
	}
}

func (c *WizardVerifierCircuit) generateAllRandomCoins(_ frontend.API) {

	for currRound := 0; currRound < c.Spec.NumRounds(); currRound++ {
		if currRound > 0 {
			/*
				Sanity-check : Make sure all issued random coin have been
				"consumed" by all the verifiers steps, in the round we are
				"closing"
			*/
			toBeConsumed := c.Spec.Coins.AllKeysAt(currRound - 1)
			c.Coins.Exists(toBeConsumed...)

			// Make sure that all messages have been written and use them
			// to update the FS state. Note that we do not need to update
			// FS using the last round of the prover because he is always
			// the last one to "talk" in the protocol.
			toUpdateFS := c.Spec.Columns.AllKeysProofAt(currRound - 1)
			for _, msg := range toUpdateFS {

				msgID := c.MessagesIDs.MustGet(msg)
				msgContent := c.Columns[msgID]

				logrus.Tracef("VERIFIER CIRCUIT : Updating the FS oracle with a message - %v", msg)
				c.FS.UpdateVec(msgContent)
			}

			toUpdateFS = c.Spec.Columns.AllKeysPublicInputAt(currRound - 1)
			for _, msg := range toUpdateFS {

				msgID := c.MessagesIDs.MustGet(msg)
				msgContent := c.Columns[msgID]

				logrus.Tracef("VERIFIER CIRCUIT : Updating the FS oracle with public input - %v", msg)
				c.FS.UpdateVec(msgContent)
			}

			/*
				Also include the prover's allegations for all evaluations
			*/
			queries := c.Spec.QueriesParams.AllKeysAt(currRound - 1)
			for _, qName := range queries {
				// Implicitly, this will panic whenever we start supporting
				// a new type of query params
				params := c.GetParams(qName)
				params.UpdateFS(c.FS)
			}
		}

		/*
			Then assigns the coins for the new round.
		*/
		toCompute := c.Spec.Coins.AllKeysAt(currRound)
		for _, coinName := range toCompute {
			info := c.Spec.Coins.Data(coinName)
			logrus.Tracef("VERIFIER CIRCUIT : Generate a random coin - %v", coinName)
			switch info.Type {
			case coin.Field:
				value := c.FS.RandomField()
				c.Coins.InsertNew(coinName, value)
			case coin.IntegerVec:
				value := c.FS.RandomManyIntegers(info.Size, info.UpperBound)
				c.Coins.InsertNew(coinName, value)
			}
		}
	}
}

/*
Returns a field element random a preassigned random coin as field element.
The implementation implicitly checks that the field element is of the right typ
*/
func (c *WizardVerifierCircuit) GetRandomCoinField(name coin.Name) frontend.Variable {
	/*
		Early check, ensures the coin has been registered at all
		and that it has the correct type
	*/
	infos := c.Spec.Coins.Data(name)
	if infos.Type != coin.Field {
		utils.Panic("Coin was registered as %v but got %v", infos.Type, coin.Field)
	}
	// If this panics, it means we generates the coins wrongly
	return c.Coins.MustGet(name).(frontend.Variable)
}

/*
Returns a pre-sampled integer vec random coin.
*/
func (c *WizardVerifierCircuit) GetRandomCoinIntegerVec(name coin.Name) []frontend.Variable {
	/*
		Early check, ensures the coin has been registered at all
		and that it has the correct type
	*/
	infos := c.Spec.Coins.Data(name)
	if infos.Type != coin.IntegerVec {
		utils.Panic("Coin was registered as %v but got %v", infos.Type, coin.IntegerVec)
	}
	// If this panics, it means we generates the coins wrongly
	return c.Coins.MustGet(name).([]frontend.Variable)
}

/*
Returns the parameters of a univariate evaluation (i.e: x, the evaluation point)
and y, the alleged polynomial opening.
*/
func (c *WizardVerifierCircuit) GetUnivariateParams(name ifaces.QueryID) query.GnarkUnivariateEvalParams {
	qID := c.UnivariateParamsIDs.MustGet(name)
	params := c.UnivariateParams[qID]

	// Sanity-checks
	info := c.GetUnivariateEval(name)
	if len(info.Pols) != len(params.Ys) {
		utils.Panic("(for %v) inconsistent lengths %v %v", name, len(info.Pols), len(params.Ys))
	}
	return params
}

/*
Get univariate eval metadata. Panic if not found.
*/
func (c *WizardVerifierCircuit) GetUnivariateEval(name ifaces.QueryID) query.UnivariateEval {
	return c.Spec.QueriesParams.Data(name).(query.UnivariateEval)
}

// Returns pre-assigned parameters for the current query
func (c *WizardVerifierCircuit) GetInnerProductParams(name ifaces.QueryID) query.GnarkInnerProductParams {
	qID := c.InnerProductIDs.MustGet(name)
	params := c.InnerProductParams[qID]

	// Sanity-checks
	info := c.Spec.QueriesParams.Data(name).(query.InnerProduct)
	if len(info.Bs) != len(params.Ys) {
		utils.Panic("(for %v) inconsistent lengths %v %v", name, len(info.Bs), len(params.Ys))
	}
	return params
}

/*
Returns the parameters of a univariate evaluation (i.e: x, the evaluation point)
and y, the alleged polynomial opening.
*/
func (c *WizardVerifierCircuit) GetLocalPointEvalParams(name ifaces.QueryID) query.GnarkLocalOpeningParams {
	qID := c.LocalOpeningIDs.MustGet(name)
	return c.LocalOpeningParams[qID]
}

/*
Returns a message by name
*/
func (c *WizardVerifierCircuit) GetColumn(name ifaces.ColID) []frontend.Variable {

	// case where the column is part of the verification key
	if c.Spec.Columns.Status(name) == column.VerifyingKey {
		val := smartvectors.IntoRegVec(c.Spec.Precomputed.MustGet(name))
		res := gnarkutil.AllocateSlice(len(val))
		// Return the column as an array of constants
		for i := range val {
			res[i] = val[i]
		}
		return res
	}

	msgID := c.MessagesIDs.MustGet(name)
	wrappedMsg := c.Columns[msgID]

	size := c.Spec.Columns.GetSize(name)
	if len(wrappedMsg) != size {
		utils.Panic("bad dimension %v, spec expected %v", len(wrappedMsg), size)
	}

	return wrappedMsg
}

/*
Returns a position in a message by name
*/
func (c *WizardVerifierCircuit) GetColumnAt(name ifaces.ColID, pos int) frontend.Variable {
	return c.GetColumn(name)[pos]
}

/*
Creates an empty wizard verifier circuit. Initializes the underlying
structs and collections.
*/
func newWizardVerifierCircuit() *WizardVerifierCircuit {
	res := &WizardVerifierCircuit{}
	res.MessagesIDs = collection.NewMapping[ifaces.ColID, int]()
	res.UnivariateParamsIDs = collection.NewMapping[ifaces.QueryID, int]()
	res.LocalOpeningIDs = collection.NewMapping[ifaces.QueryID, int]()
	res.InnerProductIDs = collection.NewMapping[ifaces.QueryID, int]()
	res.Columns = [][]frontend.Variable{}
	res.UnivariateParams = make([]query.GnarkUnivariateEvalParams, 0)
	res.InnerProductParams = make([]query.GnarkInnerProductParams, 0)
	res.LocalOpeningParams = make([]query.GnarkLocalOpeningParams, 0)
	res.Coins = collection.NewMapping[coin.Name, interface{}]()
	return res
}

/*
Assigns values to the wizard verifier circuit from a proof
*/
func GetWizardVerifierCircuitAssignment(comp *CompiledIOP, proof Proof) *WizardVerifierCircuit {

	res := newWizardVerifierCircuit()

	/*
		Assigns the messages. Note that the iteration order is made
		consistent with `AllocateWizardCircuit`
	*/
	for i, colName := range comp.Columns.AllKeys() {

		// filter the columns by status
		status := comp.Columns.Status(colName)
		if !status.IsPublic() {
			// the column is not public so it is not part of the proof
			continue
		}

		if status == column.VerifyingKey {
			// this are constant columns
			continue
		}

		logrus.Tracef("VERIFIER CIRCUIT : registering column %v (as %v) in circuit (#%v)", colName, status.String(), i)
		msgDataIFace := proof.Messages.MustGet(colName)
		msgData := msgDataIFace

		// Perform the conversion to frontend.Variable, element by element
		assignedMsg := smartvectors.IntoGnarkAssignment(msgData)
		res.Columns = append(res.Columns, assignedMsg)

		// Also add the index
		res.MessagesIDs.InsertNew(colName, i)
	}

	/*
		Assigns the query parameters. Note that the iteration order is
		made deterministic to match the iteration order of the
	*/
	for _, qName := range comp.QueriesParams.AllKeys() {
		/*
			Note that we do not filter out the "already compiled" queries
			here.
		*/
		paramsIface := proof.QueriesParams.MustGet(qName)

		switch params := paramsIface.(type) {

		case query.UnivariateEvalParams:
			res.UnivariateParamsIDs.InsertNew(qName, len(res.UnivariateParams))
			res.UnivariateParams = append(res.UnivariateParams, params.GnarkAssign())

		case query.InnerProductParams:
			res.InnerProductIDs.InsertNew(qName, len(res.InnerProductParams))
			res.InnerProductParams = append(res.InnerProductParams, params.GnarkAssign())

		case query.LocalOpeningParams:
			res.LocalOpeningIDs.InsertNew(qName, len(res.LocalOpeningParams))
			res.LocalOpeningParams = append(res.LocalOpeningParams, params.GnarkAssign())

		default:
			utils.Panic("unknow type %T", params)
		}
	}

	res.Spec = comp
	return res
}

// Returns a query parameters as a generic interface
func (c *WizardVerifierCircuit) GetParams(id ifaces.QueryID) ifaces.GnarkQueryParams {
	switch t := c.Spec.QueriesParams.Data(id).(type) {
	case query.UnivariateEval:
		return c.GetUnivariateParams(id)
	case query.LocalOpening:
		return c.GetLocalPointEvalParams(id)
	case query.InnerProduct:
		return c.GetInnerProductParams(id)
	default:
		utils.Panic("unexpected type : %T", t)
	}
	panic("unreachable")
}
