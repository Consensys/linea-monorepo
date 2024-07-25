package wizard

import (
	"github.com/consensys/zkevm-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/zkevm-monorepo/prover/maths/field"
	"github.com/consensys/zkevm-monorepo/prover/utils"
	"github.com/consensys/zkevm-monorepo/prover/utils/collection"
)

// DefineFunc is a function mutating an [API] object. They are meant as the
// initial definer of the protocol.
type DefineFunc = func(*API)

// DefineFunc is a function mutating an [API] object. They are meant to
// transmute an already initialized [API] object to another by applying query
// reduction.
type Compiler = func(*API)

type API struct {
	currScope   scope
	itemCounter *int
	*CompiledIOP
}

type CompiledIOP struct {
	columns                *byRoundRegister[ColNatural]
	coins                  *byRoundRegister[Coin]
	queries                *byRoundRegister[Query]
	runtimeProverActions   *byRoundRegister[runtimeProverAction]
	runtimeVerifierActions *byRoundRegister[runtimeVerifierAction]
	precomputeds           collection.Mapping[id, smartvectors.SmartVector]
	protocolHash           field.Element
}

// newAPI initializes an empty API object
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

func (api *API) CurrScope() string {
	return api.currScope.getFullScope()
}

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
