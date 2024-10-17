package wizard

import (
	"testing"

	"github.com/consensys/zkevm-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/zkevm-monorepo/prover/maths/field"
	"github.com/consensys/zkevm-monorepo/prover/utils"
	"github.com/consensys/zkevm-monorepo/prover/utils/collection"
	"github.com/stretchr/testify/assert"
)

// DefineFunc is a function mutating an [API] object. They are meant as the
// initial definer of the protocol.
type DefineFunc = func(*API)

// DefineFunc is a function mutating an [API] object. They are meant to
// transmute an already initialized [API] object to another by applying query
// reduction.
type Compiler = func(*API)

// API is a struture recording the specification of a wizard protocol. In
// addition to the specification of the wizard protocol.
type API struct {
	// currScope holds context information about which part of the protocol is
	// being built. The informations attached to the scope are propagated to
	// all the protocol object declared under this scope.
	currScope scope
	// The item counter is incremented everytime a new object/item is declared
	// within the protocol. This is used to give a unique id to every object
	// of the protocol.
	itemCounter *int
	*CompiledIOP
}

// CompiledIOP stores the representation of a protocol. The struct essentially
// acts as a data structure providing a static description of the protocoL.
type CompiledIOP struct {
	columns                *byRoundRegister[ColNatural]
	coins                  *byRoundRegister[Coin]
	queries                *byRoundRegister[Query]
	runtimeProverActions   *byRoundRegister[runtimeProverAction]
	runtimeVerifierActions *byRoundRegister[runtimeVerifierAction]
	precomputeds           collection.Mapping[id, smartvectors.SmartVector]
	protocolHash           field.Element
}

// newAPI initializes an empty API object. The returned API object is set to
// have no scope information.
func NewAPI(defineFunc DefineFunc) *API {
	var (
		itemCounter = int(0)
		api         = &API{
			currScope: scope{},
			CompiledIOP: &CompiledIOP{
				columns:                newRegister[ColNatural](),
				coins:                  newRegister[Coin](),
				queries:                newRegister[Query](),
				runtimeProverActions:   newRegister[runtimeProverAction](),
				runtimeVerifierActions: newRegister[runtimeVerifierAction](),
			},
			itemCounter: &itemCounter,
		}
	)
	defineFunc(api)
	return api
}

func (api *API) newID() id {
	*api.itemCounter++
	return id(*api.itemCounter)
}

// CurrScope returns a string representation of the scope stored in the API.
func (api *API) CurrScope() string {
	return api.currScope.getFullScope()
}

// NumRounds returns the number of interaction rounds registered as part of the
// protocol.
func (comp *CompiledIOP) NumRounds() int {

	var (
		numRounds = []int{
			comp.coins.numRounds(),
			comp.columns.numRounds(),
			comp.queries.numRounds(),
			comp.runtimeProverActions.numRounds(),
			comp.runtimeVerifierActions.numRounds(),
		}
		maxNumRounds = utils.Max(numRounds...)
	)

	// This equalizes the number of rounds in every stores
	comp.coins.reserveFor(maxNumRounds)
	comp.columns.reserveFor(maxNumRounds)
	comp.queries.reserveFor(maxNumRounds)
	comp.runtimeProverActions.reserveFor(maxNumRounds)
	comp.runtimeVerifierActions.reserveFor(maxNumRounds)

	return maxNumRounds
}

func (api *API) Compile(compilers ...Compiler) *API {
	for i := range compilers {
		compilers[i](api)
	}
	return api
}

func CompileTest(t *testing.T, define DefineFunc, prover func(*RuntimeProver), compilers ...Compiler) (comp *CompiledIOP, rp *RuntimeProver) {
	comp = NewAPI(define).Compile(compilers...).CompiledIOP
	rp = comp.NewRuntimeProver(prover).Run()
	proof := rp.Proof()
	valid := comp.Verify(proof)
	assert.NoError(t, valid)
	return comp, rp
}
