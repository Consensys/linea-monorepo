package gkrmimc

import (
	"math/big"

	"github.com/consensys/accelerated-crypto-monorepo/crypto/mimc"
	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
	"github.com/consensys/gnark/constraint"
	cs "github.com/consensys/gnark/constraint/bn254"
	"github.com/consensys/gnark/constraint/solver"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/gkr"
	"github.com/consensys/gnark/std/hash"
	"github.com/sirupsen/logrus"
)

// Initialize hashing factory
func NewHasherFactory(api frontend.API) *HasherFactory {
	res := &HasherFactory{
		initStates: []frontend.Variable{},
		blocks:     []frontend.Variable{},
		newStates:  []frontend.Variable{},
		api:        api,
	}

	api.Compiler().Defer(res.finalize)
	return res
}

// Hasher factory that outputs uses GKR for MiMC verification
type HasherFactory struct {
	initStates []frontend.Variable
	blocks     []frontend.Variable
	newStates  []frontend.Variable
	api        frontend.API
}

// Push a triplet of variable onto the list of hashes to prover
func (f *HasherFactory) pushTriplets(initstate, block, newstate frontend.Variable) {
	f.initStates = append(f.initStates, initstate)
	f.blocks = append(f.blocks, block)
	f.newStates = append(f.newStates, newstate)
}

// Compile check to ensure that the hasher does implement the
// field hasher interface
var _ hash.FieldHasher = &Hasher{}

// Hasher that defers the hash verification to the factori
type Hasher struct {
	data    []frontend.Variable
	state   frontend.Variable
	factory *HasherFactory
}

// Spawns a hasher that will defer the hash verification to the factory
func (f *HasherFactory) NewHasher() Hasher {
	return Hasher{factory: f, state: frontend.Variable(0)}
}

// Writes fields elements into the hasher
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

// Reinitialize the state of the hasher
func (h *Hasher) Reset() {
	h.data = nil
	h.state = 0
}

// Sum hashes in circuit
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

// Calls the compression function for a hash
func (h *Hasher) compress(state, block frontend.Variable) frontend.Variable {

	new, err := h.factory.api.Compiler().NewHint(mimcHintfunc, 1, state, block)
	if err != nil {
		panic(err)
	}

	h.factory.pushTriplets(state, block, new[0])
	return new[0]
}

// Hint function used for MiMC
func mimcHintfunc(f *big.Int, inputs []*big.Int, outputs []*big.Int) error {

	if f.String() != field.Modulus().String() {
		utils.Panic("Not the bn field %d != %d", f, field.Modulus())
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

func fromBigInts(arr []*big.Int) []field.Element {
	res := make([]field.Element, len(arr))
	for i := range res {
		res[i].SetBigInt(arr[i])
	}
	return res
}

func intoBigInts(res []*big.Int, arr ...field.Element) []*big.Int {

	if len(res) != len(arr) {
		utils.Panic("got %v bigints but %v field elments", len(res), len(arr))
	}

	for i := range res {
		arr[i].BigInt(res[i])
	}
	return res
}

func (f *HasherFactory) finalize(api frontend.API) error {

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

// Pad the factory with dummy hashes to reach a power-of-two number
// of instances. Returns the padded length
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

// call this every time you want to solve the circuit
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
