package gkrmimc

import (
	"math/big"

	"github.com/consensys/gnark/constraint"
	cs "github.com/consensys/gnark/constraint/bls12-377"
	"github.com/consensys/gnark/constraint/solver"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/gkr"
	"github.com/consensys/gnark/std/hash"
	"github.com/consensys/linea-monorepo/prover/crypto/mimc"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/sirupsen/logrus"
)

var cache []*HasherFactory

// NewHasherFactory initializes a new [HasherFactory] object. Ideally, it should
// be called only once per circuit.
func NewHasherFactory(api frontend.API) *HasherFactory {
	for _, hf := range cache {
		if hf.api == api {
			return hf
		}
	}

	res := &HasherFactory{
		initStates: []frontend.Variable{},
		blocks:     []frontend.Variable{},
		newStates:  []frontend.Variable{},
		api:        api,
	}

	api.Compiler().Defer(res.finalize)
	cache = append(cache, res)

	return res
}

// HasherFactory is an object that can construct hashers satisfying the
// [hash.FieldHasher] interface and which can be used to perform MiMC hashing
// in a gnark circuit. All hashing operations performed by these hashers are
// bare claims whose truthfullness is backed by the verification of a GKR proof
// in the same circuit. This deferred GKR verification is hidden from the user.
type HasherFactory struct {
	initStates []frontend.Variable
	blocks     []frontend.Variable
	newStates  []frontend.Variable
	api        frontend.API
}

// pushTriplets pushes a claim-triplet of frontend.Variable to the list of triplet
// to be verified by the GKR circuit at the end of the circuit.
func (f *HasherFactory) pushTriplets(initstate, block, newstate frontend.Variable) {
	f.initStates = append(f.initStates, initstate)
	f.blocks = append(f.blocks, block)
	f.newStates = append(f.newStates, newstate)
}

// Compile check to ensure that the hasher does implement the field hasher interface
var _ hash.FieldHasher = &Hasher{}

// Hasher that defers the hash verification to the factori
type Hasher struct {
	data    []frontend.Variable
	state   frontend.Variable
	factory *HasherFactory
}

// NewHasher spawns a hasher that will defer the hash verification to the
// factory. It is safe to be called multiple times and the returned Hasher can
// be used exactly in the same way as [github.com/consensys/gnark/std/hash/mimc.NewMiMC]
// and will provide the same results for the same usage.
//
// However, the hasher should not be used in deferred gnark circuit execution.
func (f *HasherFactory) NewHasher() Hasher {
	return Hasher{factory: f, state: frontend.Variable(0)}
}

// Writes fields elements into the hasher; implements [hash.FieldHasher]
func (h *Hasher) Write(data ...frontend.Variable) {
	// sanity-check : it is a common bug that we may be []frontend.Variable
	// as a frontend.Variable
	for i := range data {
		if _, ok := data[i].([]frontend.Variable); ok {
			utils.Panic("bug in define, got a []frontend.Variable")
		}
	}
	h.data = append(h.data, data...)
}

// Reinitialize the state of the hasher; implements [hash.FieldHasher]
func (h *Hasher) Reset() {
	h.data = nil
	h.state = 0
}

// Sum returns the hash of what was appended to the hasher so far. Calling it
// multiple time without updating returns the same result. This function
// implements [hash.FieldHasher] interfae
func (h *Hasher) Sum() frontend.Variable {
	// 1 - Call the compression function in a loop
	curr := h.state
	for _, stream := range h.data {
		curr = h.compress(curr, stream)
	}
	// flush the data already hashed
	h.data = nil
	h.state = curr
	return curr
}

// compress calls returns a frontend.Variable holding the result of applying
// the compression function of MiMC over state and block. The alleged returned
// result is pushed on the stack of all the claims to verify.
func (h *Hasher) compress(state, block frontend.Variable) frontend.Variable {

	newState, err := h.factory.api.Compiler().NewHint(mimcHintfunc, 1, state, block)
	if err != nil {
		panic(err)
	}

	h.factory.pushTriplets(state, block, newState[0])
	return newState[0]
}

// mimcHintfunc is a gnark hint that computes the MiMC compression function, it
// is used to return the pending claims of the evaluation of the MiMC compression
// function.
func mimcHintfunc(f *big.Int, inputs []*big.Int, outputs []*big.Int) error {

	if f.String() != field.Modulus().String() {
		utils.Panic("Not the BLS field %d != %d", f, field.Modulus())
	}

	if len(inputs) != 2 {
		utils.Panic("expected 2 inputs, [init, block], got %v", len(inputs))
	}

	if len(outputs) != 1 {
		utils.Panic("expected 1 output [newState] got %v", len(inputs))
	}

	inpF := fromBigInts(inputs)
	outF := mimc.BlockCompression(inpF[0], inpF[1])

	intoBigInts(outputs, outF)
	return nil
}

// fromBigInts converts an array of big.Integer's into an array of field.Element's
func fromBigInts(arr []*big.Int) []field.Element {
	res := make([]field.Element, len(arr))
	for i := range res {
		res[i].SetBigInt(arr[i])
	}
	return res
}

// intoBigInts converts an array of field.Element's into an array of big.Integer's
func intoBigInts(res []*big.Int, arr ...field.Element) []*big.Int {

	if len(res) != len(arr) {
		utils.Panic("got %v bigints but %v field elments", len(res), len(arr))
	}

	for i := range res {
		arr[i].BigInt(res[i])
	}
	return res
}

// finalize operates the deferred verification of the claims made by all the
// [Hasher]'s linked to the receiver. This function coordinates the GKR proving
// and the verification of the proof in-circuit and also the initial randomness
// related operations.
//
// It takes a _ frontend.API because we need it to pass it to Defer.
func (f *HasherFactory) finalize(_ frontend.API) error {

	// Edge-case the circuit does not use the factory, in that case
	// we can early return
	if len(f.blocks) == 0 {
		return nil
	}

	// Edge-case, when the number of hashes is small use mimc directly
	if len(f.blocks) == 1 {
		new := mimc.GnarkBlockCompression(f.api, f.initStates[0], f.blocks[0])
		f.api.AssertIsEqual(new, f.newStates[0])
		return nil
	}

	// Pad the alleged mimc compression evaluations with dummy values
	// (which satisfy the circuit)
	f.padToPow2()

	// defer the hash verification to the GKR API
	checkWithGkr(f.api, f.initStates, f.blocks, f.newStates)
	return nil
}

// padToPow2 pads the receiver [HasherFactory] with dummy claims to reach a
// power-of-two number of instances. Returns the padded length
func (f *HasherFactory) padToPow2() int {
	size := len(f.blocks)
	targetSize := utils.NextPowerOfTwo(size)

	zero := field.Element{}
	hashOfZero := mimc.BlockCompression(zero, zero)

	for i := size; i < targetSize; i++ {
		f.pushTriplets(0, 0, hashOfZero.String())
	}

	return targetSize
}

// SolverOpts returns the list of the [solver.Option] required to prove the
// satisfiability of a circuit using the MiMC GKR circuit (and thus the
// [HasherFactory]). It registers all the hints that are necessary to solve
// the circuit.
//
// The result of this function has to be passed to the plonk.Prove function.
func SolverOpts(scs constraint.ConstraintSystem) []solver.Option {

	// Attempts to parse it as a ccs
	spr, ok := scs.(*cs.SparseR1CS)
	if !ok {
		panic("not a sparse r1cs")
	}

	// not a circuit using GKR
	if !spr.GkrInfo.Is() {
		logrus.Warn("Not a circuit using gkr, this can happen if nothing is actually hashed")
	}

	var gkrData cs.GkrSolvingData
	opts := []solver.Option{
		solver.WithHints(
			mimcHintfunc,
			gkr.SolveHintPlaceholder,
			gkr.ProveHintPlaceholder,
		),
	}

	opts = append(opts,
		solver.OverrideHint(spr.GkrInfo.SolveHintID, cs.GkrSolveHint(spr.GkrInfo, &gkrData)),
		solver.OverrideHint(spr.GkrInfo.ProveHintID, cs.GkrProveHint(spr.GkrInfo.HashName, &gkrData)),
	)

	return opts
}
