package wizard

import (
	"github.com/consensys/zkevm-monorepo/prover/utils/collection"
)

/*
In a nutshell, an item is an abstract type that
accounts for the fact that CompiledProtocol
registers various things for different rounds
*/
type byRoundRegister[DATA any] struct {
	// All the IDs for a given round
	byRounds collection.VecVec[DATA]
}

/*
Construct a new round register
*/
func newRegister[DATA any]() *byRoundRegister[DATA] {
	return &byRoundRegister[DATA]{
		byRounds: collection.NewVecVec[DATA](),
	}
}

/*
Insert for a given round. Will panic if an item
with the same ID has been registered first
*/
func (r *byRoundRegister[DATA]) addToRound(round int, data DATA) {
	r.byRounds.AppendToInner(round, data)
}

/*
Returns the list of all the keys ever. The result is returned in
Deterministic order.
*/
func (r *byRoundRegister[DATA]) all() []DATA {
	res := []DATA{}
	for roundID := 0; roundID < r.numRounds(); roundID++ {
		ids := r.allAt(roundID)
		res = append(res, ids...)
	}
	return res
}

/*
Returns the list of all keys for a given round. Result has deterministic
order (order of insertion)
*/
func (r *byRoundRegister[DATA]) allAt(round int) []DATA {
	// Reserve up to the desired length just in case.
	// It is absolutely legitimate to query "too far"
	// this can happens for queries for instance.
	// However, it should not happen for coins.
	r.byRounds.Reserve(round + 1)
	return r.byRounds.MustGet(round)
}

/*
Returns the number of rounds
*/
func (r *byRoundRegister[DATA]) numRounds() int {
	return r.byRounds.Len()
}

/*
Make sure enough rounds are allocated up to the given length
No-op if enough rounds have been allocated, otherwise, will
reserve as many as necessary.
*/
func (r *byRoundRegister[DATA]) reserveFor(newNumRounds int) {
	if r.byRounds.Len() < newNumRounds {
		r.byRounds.Reserve(newNumRounds)
	}
}
