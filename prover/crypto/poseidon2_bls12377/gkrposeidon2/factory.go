package gkrposeidon2

import (
	"errors"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/hash"
	"github.com/consensys/gnark/std/permutation/poseidon2"
)

var cache []*HasherFactory

// NewHasherFactory initializes a new [HasherFactory] object. Ideally it should be
// called only once per circuit.
func NewHasherFactory(api frontend.API) *HasherFactory {
	for _, hf := range cache {
		if hf.api == api {
			return hf
		}
	}

	res := &HasherFactory{api: api}
	api.Compiler().Defer(res.finalize)
	cache = append(cache, res)
	return res
}

// HasherFactory builds Poseidon2 (t=2) Merkle-Damgard hashers whose compression
// claims are verified in-circuit via a deferred GKR proof.
type HasherFactory struct {
	api  frontend.API
	ins1 []frontend.Variable
	ins2 []frontend.Variable
	outs []frontend.Variable
}

func (f *HasherFactory) pushTriplets(a, b, out frontend.Variable) {
	f.ins1 = append(f.ins1, a)
	f.ins2 = append(f.ins2, b)
	f.outs = append(f.outs, out)
}

// Compile-time check.
var _ hash.FieldHasher = &Hasher{}

// Hasher is a Poseidon2 Merkle-Damgard hasher that defers verification of its
// compression function to the factory.
type Hasher struct {
	data    []frontend.Variable
	state   frontend.Variable
	factory *HasherFactory
}

func (f *HasherFactory) NewCompresser() *Hasher {
	return &Hasher{factory: f, state: frontend.Variable(0)}
}

func (h *Hasher) Write(data ...frontend.Variable) {
	h.data = append(h.data, data...)
}

func (h *Hasher) Reset() {
	h.data = nil
	h.state = 0
}

func (h *Hasher) Sum() frontend.Variable {
	curr := h.state
	for _, stream := range h.data {
		curr = h.Compress(curr, stream)
	}
	h.data = nil
	h.state = curr
	return curr
}

func (h *Hasher) SetState(newState []frontend.Variable) error {
	if len(h.data) > 0 {
		return errors.New("the hasher is not in an initial state")
	}
	if len(newState) != 1 {
		return errors.New("the Poseidon2 hasher expects a single field element to represent the state")
	}
	h.state = newState[0]
	return nil
}

func (h *Hasher) State() []frontend.Variable {
	_ = h.Sum()
	return []frontend.Variable{h.state}
}

func (h *Hasher) Compress(state, block frontend.Variable) frontend.Variable {
	newState, err := h.factory.api.Compiler().NewHint(permuteHint, 1, state, block)
	if err != nil {
		panic(err)
	}
	h.factory.pushTriplets(state, block, newState[0])
	return newState[0]
}

func (f *HasherFactory) finalize(api frontend.API) error {
	if f.api != api {
		panic("unexpected API")
	}

	if len(f.outs) == 0 {
		return nil
	}

	// Small-instance fast path: avoid GKR overhead.
	if len(f.outs) == 1 {
		perm, err := poseidon2.NewPoseidon2FromParameters(api, 2, 6, 26)
		if err != nil {
			return err
		}
		api.AssertIsEqual(f.outs[0], perm.Compress(f.ins1[0], f.ins2[0]))
		return nil
	}

	// Register gates for the curve.
	registerGkrSolverOptions(api)

	// Pad inputs to power-of-two number of instances for GKR.
	target := ecc.NextPowerOfTwo(uint64(len(f.ins1)))
	ins1Padded := make([]frontend.Variable, target)
	ins2Padded := make([]frontend.Variable, target)
	copy(ins1Padded, f.ins1)
	copy(ins2Padded, f.ins2)
	for i := len(f.ins1); i < len(ins1Padded); i++ {
		ins1Padded[i] = 0
		ins2Padded[i] = 0
	}

	gkrApi, y, err := defineCircuit(ins1Padded, ins2Padded)
	if err != nil {
		return err
	}

	solution, err := gkrApi.Solve(api)
	if err != nil {
		return err
	}

	yVals := solution.Export(y)
	for i := range f.outs {
		api.AssertIsEqual(yVals[i], f.outs[i])
	}

	allVals := make([]frontend.Variable, 0, 3*len(f.ins1))
	allVals = append(allVals, f.ins1...)
	allVals = append(allVals, f.ins2...)
	allVals = append(allVals, f.outs...)

	committer, ok := f.api.(frontend.Committer)
	if !ok {
		return errors.New("frontend.API does not implement frontend.Committer")
	}
	challenge, err := committer.Commit(allVals...)
	if err != nil {
		return err
	}
	return solution.Verify(hash.POSEIDON2.String(), challenge)
}
