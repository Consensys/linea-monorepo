package wizard

import (
	// "reflect"

	"github.com/consensys/accelerated-crypto-monorepo/maths/common/smartvectors"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/coin"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/column"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/ifaces"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/query"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/variables"
	"github.com/consensys/accelerated-crypto-monorepo/symbolic"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
	"github.com/consensys/accelerated-crypto-monorepo/utils/collection"
)

// Carries the description of the IOP protocol
type CompiledIOP struct {

	// All the registered Columns (ie: messages for the oracle)
	Columns column.Store
	// All QueriesParams that have been registered
	QueriesParams ByRoundRegister[ifaces.QueryID, ifaces.Query]
	// All queries without parameters
	QueriesNoParams ByRoundRegister[ifaces.QueryID, ifaces.Query]
	// All the Coins that have been issued by the protocol
	Coins ByRoundRegister[coin.Name, coin.Info]

	// Prover functions, for all the subprotocols that starts at the given round
	SubProvers collection.VecVec[ProverStep]

	// Verifier functions, for all the subprotocol that starts at the given.
	// GnarkSubVerifiers mirror the SubVerifiers in a gnark circuit.
	subVerifiers      collection.VecVec[VerifierStep]
	gnarkSubVerifiers collection.VecVec[GnarkVerifierStep]

	// List of precomputed polynomials
	Precomputed collection.Mapping[ifaces.ColID, ifaces.ColAssignment]

	// Cryptographic commitment compiler
	CryptographicCompilerCtx any
	// Dummy compiled : adhoc flag used to indicate
	// the verifier should not be compiled into a circuit
	DummyCompiled bool
	// Count the number of self-recursions induced in the
	// protocol. Used to derive unique names for when the
	// self-recursion is called several time
	SelfRecursionCount int
}

// Returns the number of rounds in the IOP
func (c *CompiledIOP) NumRounds() int {
	// If there are no coins, we should still return 1 (at least)
	return utils.Max(1, c.Coins.NumRounds())
}

// Returns a list of all the registered commitments so far
func (c *CompiledIOP) ListCommitments() []ifaces.ColID {
	return c.Columns.AllKeys()
}

/*
Registers a new column in the protocol at a given round.
Returns a struct summarizing the metadata of the column.
*/
func (c *CompiledIOP) InsertCommit(round int, name ifaces.ColID, size int) ifaces.Column {
	return c.InsertColumn(round, name, size, column.Committed)
}

/*
Registers a new column in the protocol at a given round.
Returns a struct summarizing the metadata of the column.
*/
func (c *CompiledIOP) InsertColumn(round int, name ifaces.ColID, size int, status column.Status) ifaces.Column {
	c.assertConsistentRound(round)

	if len(name) == 0 {
		panic("Column with an empty name")
	}

	// This performs all the checks
	return c.Columns.AddToRound(round, name, size, status)
}

/*
Registers a new coin at a given rounds. Returns a coin.Info object.

* For normal coins, pass

```
_ = c.InsertCoin(<round of sampling>, <stringID of the coin>, coin.Field)
```

* For IntegerVec coins, pass

```
_ = c.InsertCoin(<round of sampling>, <stringID of the coin>, coin.IntegerVec, <#Size of the vec>, <#Bound on the integers>)
```
*/
func (c *CompiledIOP) InsertCoin(round int, name coin.Name, type_ coin.Type, size ...int) coin.Info {
	// Short-hand to access the compiled object
	info := coin.NewInfo(name, type_, round, size...)
	c.Coins.AddToRound(round, name, info)
	return info
}

/*
Insert global constraint
*/
func (c *CompiledIOP) InsertGlobal(round int, name ifaces.QueryID, expr *symbolic.Expression, noBoundCancel ...bool) query.GlobalConstraint {

	c.assertConsistentRound(round)

	/*
		The constructor of the global constraint is assumed to
		perform all the well-formation checks
	*/
	cs := query.NewGlobalConstraint(name, expr, noBoundCancel...)
	boarded := cs.Board()
	metadatas := boarded.ListVariableMetadata()

	// Test the existence of all variable in the instance
	for _, metadataInterface := range metadatas {
		switch metadata := metadataInterface.(type) {
		case ifaces.Column:
			// The handle mecanism prevents this.
		case coin.Info:
			c.Coins.MustExists(metadata.Name)
		case variables.X, variables.PeriodicSample, *ifaces.Accessor:
			// Pass
		default:
			utils.Panic("Not a variable type %T in query %v", metadataInterface, cs.ID)
		}
	}

	// Finally registers the query
	c.QueriesNoParams.AddToRound(round, name, cs)

	return cs
}

// Insert local
func (c *CompiledIOP) InsertLocal(round int, name ifaces.QueryID, cs_ *symbolic.Expression) query.LocalConstraint {

	c.assertConsistentRound(round)

	cs := query.NewLocalConstraint(name, cs_)
	boarded := cs.Board()
	metadatas := boarded.ListVariableMetadata()

	// Test the existence of all variable in the instance
	for _, metadataInterface := range metadatas {
		switch metadata := metadataInterface.(type) {
		case ifaces.Column:
			// Existence is guaranteed already
		case coin.Info:
			c.Coins.MustExists(metadata.Name)
		case variables.X, variables.PeriodicSample, *ifaces.Accessor:
			// Pass
		default:
			utils.Panic("Not a variable type %T in query %v", metadataInterface, cs.ID)
		}
	}

	// Finally registers the query
	c.QueriesNoParams.AddToRound(round, name, cs)

	return cs
}

/*
Insert a permutation
*/
func (c *CompiledIOP) InsertPermutation(round int, name ifaces.QueryID, a, b []ifaces.Column) query.Permutation {
	c.assertConsistentRound(round)
	query_ := query.NewPermutation(name, a, b)
	c.QueriesNoParams.AddToRound(round, name, query_)
	return query_
}

/*
Insert a fixedpermutation
*/
func (c *CompiledIOP) InsertFixedPermutation(round int, name ifaces.QueryID, p []ifaces.ColAssignment, a, b []ifaces.Column) query.FixedPermutation {
	query_ := query.NewFixedPermutation(name, p, a, b)
	c.QueriesNoParams.AddToRound(round, name, query_)
	return query_
}

/*
Creates an inclusion query. Here, `included` and `including` are viewed
as a arrays and the query asserts that `included` contains only rows
that are contained within `includings`, regardless of the multiplicity.
*/
func (c *CompiledIOP) InsertInclusion(round int, name ifaces.QueryID, including, included []ifaces.Column) {
	c.assertConsistentRound(round)
	query := query.NewInclusion(name, included, including)
	c.QueriesNoParams.AddToRound(round, name, query)
}

/*
Registers a proof message and specifies static informations regarding it
*/
func (c *CompiledIOP) InsertPrecomputed(name ifaces.ColID, v smartvectors.SmartVector) (msg ifaces.Column) {
	// Common : No zero length
	if v.Len() == 0 {
		utils.Panic("when registering %v, VecType with length zero", name)
	}

	// Circuit-breaker : if the precomputed poly had already been inserted we
	// can simply return it.
	if c.Columns.Exists(name) {
		return c.Columns.GetHandle(name)
	}

	c.Precomputed.InsertNew(name, v)
	return c.Columns.AddToRound(0, name, v.Len(), column.Precomputed)
}

/*
Registers a proof message and specifies static informations regarding it
*/
func (c *CompiledIOP) InsertProof(round int, name ifaces.ColID, size int) (msg ifaces.Column) {
	c.assertConsistentRound(round)

	// Common : No zero length
	if size == 0 {
		utils.Panic("when registering %v, VecType with length zero", name)
	}

	return c.Columns.AddToRound(round, name, size, column.Proof)
}

/*
Registers a public input column, and specifies static information regarding it
*/
func (c *CompiledIOP) InsertPublicInput(round int, name ifaces.ColID, size int) (msg ifaces.Column) {
	c.assertConsistentRound(round)

	// Common : No zero length
	if size == 0 {
		utils.Panic("when registering %v, VecType with length zero", name)
	}

	return c.Columns.AddToRound(round, name, size, column.PublicInput)
}

/*
Add a precompiled verifier step

The user must pass at the same time, the "native" VerifierStep but also the
GnarkVerifierStep mirroring it on a GnarkVerifierCircuit.
*/
func (c *CompiledIOP) InsertVerifier(round int, ver VerifierStep, gnarkVer GnarkVerifierStep) {
	c.assertConsistentRound(round)
	c.gnarkSubVerifiers.AppendToInner(round, gnarkVer)
	c.subVerifiers.AppendToInner(round, ver)
}

/*
Creates a range query
*/
func (c *CompiledIOP) InsertRange(round int, name ifaces.QueryID, h ifaces.Column, max int) {

	// sanity-check the bound should be larger than 0
	if max == 0 {
		panic("max is zero : perhaps an overflow")
	}

	c.assertConsistentRound(round)
	/*
		In case the range is applied over a composite handle.
		We apply the range over each natural component of the handle.
	*/
	query := query.NewRange(name, h, max)
	c.QueriesNoParams.AddToRound(round, name, query)
}

// Insert an inner-product query
func (c *CompiledIOP) InsertInnerProduct(round int, name ifaces.QueryID, a ifaces.Column, bs []ifaces.Column) query.InnerProduct {
	c.assertConsistentRound(round)

	// Also ensures that the query round does not predates the columns rounds
	maxComRound := a.Round()
	for _, b := range bs {
		maxComRound = utils.Max(maxComRound, b.Round())
	}

	if maxComRound > round {
		utils.Panic("The query is declared for round %v, but at least one column is declared for round %v", round, maxComRound)
	}

	query := query.NewInnerProduct(name, a, bs...)
	c.QueriesParams.AddToRound(round, name, query)
	return query
}

// Get an Inner-product query
func (run *CompiledIOP) GetInnerProduct(name ifaces.QueryID) query.InnerProduct {
	return run.QueriesParams.Data(name).(query.InnerProduct)
}

/*
Insert univariate
*/
func (c *CompiledIOP) InsertUnivariate(round int, name ifaces.QueryID, pols []ifaces.Column) query.UnivariateEval {
	c.assertConsistentRound(round)
	q := query.NewUnivariateEval(name, pols...)
	// Finally registers the query
	c.QueriesParams.AddToRound(round, name, q)
	return q
}

/*
Insert univariate evaluation with a fixed point
*/
func (c *CompiledIOP) InsertLocalOpening(round int, name ifaces.QueryID, pol ifaces.Column) query.LocalOpening {
	c.assertConsistentRound(round)
	q := query.NewLocalOpening(name, pol)
	// Finally registers the query
	c.QueriesParams.AddToRound(round, name, q)
	return q
}

// compare the round passed as an argument and panic if it greater than `coin.Round`. This
// helps ensuring that we do not have "useless" rounds.
func (c *CompiledIOP) assertConsistentRound(round int) {
	if round > c.Coins.NumRounds() {
		utils.Panic("Inserted at round %v, but the max should be %v", round, c.Coins.NumRounds())
	}
}

// Declares a MiMC constraints query
func (c *CompiledIOP) InsertMiMC(round int, id ifaces.QueryID, block, old, new ifaces.Column) query.MiMC {
	c.assertConsistentRound(round)
	q := query.NewMiMC(id, block, old, new)
	c.QueriesNoParams.AddToRound(round, id, q)
	return q
}
