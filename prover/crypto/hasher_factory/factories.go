package hasher_factory

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/consensys/gnark-crypto/field/koalabear/vortex"
	"github.com/consensys/gnark/constraint"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	"github.com/consensys/linea-monorepo/prover/crypto/poseidon2_koalabear"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/plonkinternal/plonkbuilder"
	"github.com/consensys/linea-monorepo/prover/utils"
)

type Octuplet [8]frontend.Variable

// HasherFactory is an interface implemented by structures that can construct a
// Poseidon2 hasher in a gnark circuit. Some implementation may leverage the GKR
// protocol as in [gkrmimc.HasherFactory] or may trigger specific behaviors
// of Plonk in Wizard.
type HasherFactory interface {
	NewHasher() poseidon2_koalabear.GnarkMDHasher
}

// BasicHasherFactory is a simple implementation of HasherFactory that returns
// the standard Poseidon2 hasher.
type BasicHasherFactory struct {
	Api frontend.API
}

// ExternalHasherFactory is an implementation of the HasherFactory interface
// that tags the variables happening in a Poseidon2 hasher claim.
type ExternalHasherFactory struct {
	Api frontend.API
}

// ExternalHasher is an implementation of the [ghash.StateStorer] interface
// that tags the variables happening in a MiMC hasher claim.
type ExternalHasher struct {
	api   frontend.API
	data  []frontend.Variable
	state [poseidon2_koalabear.BlockSize]frontend.Variable
}

// externalHasherBuilder is an implementation of the [frontend.Builder]
// interface.
type ExternalHasherBuilder struct {
	plonkbuilder.Builder
	// claimTriplets stores the tripled [oldState, block, newState]
	claimTriplets [][3]frontend.Variable
	// rcCols is a channel used to pass back the position of the wires
	// corresponding to the claims.
	rcCols chan [][3][2]int
	// addGateForHashCheck indicates whether the factory should add a gate
	// to check the claims when it holds over the sum of two gates.
	addGateForHashCheck bool
}

// externalHashBuilderIFace is an interface implemented by [externalHasherBuilder]
// and potential struct wrappers.
type externalHashBuilderIFace interface {
	CheckHashExternally(oldState, block, newState frontend.Variable)
}

// NewHasher returns the standard Poseidon2 hasher.
func (f *BasicHasherFactory) NewHasher() poseidon2_koalabear.GnarkMDHasher {
	h, _ := poseidon2_koalabear.NewGnarkMDHasher(f.Api)
	return h
}

// NewHasher returns an external Poseidon2 hasher.
func (f *ExternalHasherFactory) NewHasher() ExternalHasher {
	initState := [poseidon2_koalabear.BlockSize]frontend.Variable{}
	for i := 0; i < poseidon2_koalabear.BlockSize; i++ {
		initState[i] = frontend.Variable(0)
	}
	return ExternalHasher{api: f.Api, state: initState}

}

// Writes fields elements into the hasher; implements [hash.FieldHasher]
func (h *ExternalHasher) Write(data ...frontend.Variable) {
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
func (h *ExternalHasher) Reset() {
	h.data = nil
	for i := 0; i < poseidon2_koalabear.BlockSize; i++ {
		h.state[i] = 0
	}
}

// Sum returns the hash of what was appended to the hasher so far. Calling it
// multiple time without updating returns the same result. This function
// implements [hash.FieldHasher] interface.
func (h *ExternalHasher) Sum() [poseidon2_koalabear.BlockSize]frontend.Variable {
	// 1 - Call the compression function in a loop
	curr := h.state
	var block [poseidon2_koalabear.BlockSize]frontend.Variable

	remaining := len(h.data) % poseidon2_koalabear.BlockSize
	for i, stream := range h.data {
		if i >= len(h.data)-remaining {
			// padding with zeros
			block[poseidon2_koalabear.BlockSize-i%poseidon2_koalabear.BlockSize] = stream
			for j := 0; j < poseidon2_koalabear.BlockSize-i%poseidon2_koalabear.BlockSize; j++ {
				block[j] = 0
			}
			curr = h.compress(curr, block)
			break
		}
		block[i%poseidon2_koalabear.BlockSize] = stream
		if i%poseidon2_koalabear.BlockSize == poseidon2_koalabear.BlockSize-1 {
			curr = h.compress(curr, block)
		}
	}
	// flush the data already hashed
	h.data = nil
	h.state = curr
	return curr
}

// SetState manually sets the state of the hasher to the provided value. In the
// case of MiMC only a single frontend variable is expected to represent the
// state.
func (h *ExternalHasher) SetState(newState []frontend.Variable) error {

	if len(h.data) > 0 {
		return errors.New("the hasher is not in an initial state")
	}

	if len(newState) != 1 {
		return errors.New("the MiMC hasher expects a single field element to represent the state")
	}
	for i := 0; i < poseidon2_koalabear.BlockSize; i++ {
		h.state[i] = newState[i]
	}
	return nil
}

// State returns the inner-state of the hasher. In the context of MiMC only a
// single field element is returned.
func (h *ExternalHasher) State() []frontend.Variable {
	_ = h.Sum() // to flush the hasher
	return []frontend.Variable{h.state}
}

// compress calls returns a frontend.Variable holding the result of applying
// the compression function of MiMC over state and block. The alleged returned
// result is pushed on the stack of all the claims to verify.
func (h *ExternalHasher) compress(state, block [poseidon2_koalabear.BlockSize]frontend.Variable) [poseidon2_koalabear.BlockSize]frontend.Variable {
	var input [poseidon2_koalabear.BlockSize * 2]frontend.Variable
	copy(input[0:poseidon2_koalabear.BlockSize], state[:])
	copy(input[poseidon2_koalabear.BlockSize:poseidon2_koalabear.BlockSize*2], block[:])
	newState, err := h.api.Compiler().NewHint(Poseidon2Hintfunc, 8, input[:]...)
	if err != nil {
		panic(err)
	}

	// This asserts that the builder should be compatible with the external hasher,
	// doing it by comparing the types would be too strict as it should be
	// acceptable to wrap the externalHashBuilder into another builder without
	// making this check fail.
	builder, ok := h.api.(externalHashBuilderIFace)
	if !ok {
		utils.Panic("the builder doesn't implement externalHashBuilderIFace: %T", h.api)
	}

	// Convert the slice to an array of size 8
	var newStateOct [8]frontend.Variable
	copy(newStateOct[:], newState[:8])
	for j := 0; j < poseidon2_koalabear.BlockSize; j++ {
		builder.CheckHashExternally(state[j], block[j], newStateOct[j])
	}
	return newStateOct
}

// NewExternalHasherBuilder constructs and returns a new external hasher builder
// and a function to get the position of the wires corresponding to the variables
// taking part in each claim.
func NewExternalHasherBuilder(addGateForHashCheck bool) (frontend.NewBuilderU32, func() [][3][2]int) {
	rcCols := make(chan [][3][2]int)
	return func(field *big.Int, config frontend.CompileConfig) (frontend.Builder[constraint.U32], error) {
			b, err := plonkbuilder.From(scs.NewBuilder[constraint.U32])(field, config)
			if err != nil {
				return nil, fmt.Errorf("could not create new native builder: %w", err)
			}
			scb, ok := b.(plonkbuilder.Builder)
			if !ok {
				return nil, fmt.Errorf("native builder doesn't implement committer or kvstore")
			}
			return &ExternalHasherBuilder{
				Builder:             scb,
				rcCols:              rcCols,
				addGateForHashCheck: addGateForHashCheck,
			}, nil
		}, func() [][3][2]int {
			return <-rcCols
		}
}

// CheckHashExternally tags a Poseidon2 hasher claim in the circuit
func (f *ExternalHasherBuilder) CheckHashExternally(oldState, block, newState frontend.Variable) {
	f.claimTriplets = append(f.claimTriplets, [3]frontend.Variable{oldState, block, newState})
}

// Compile processes range checked variables and then calls Compile method of
// the underlying builder.
func (builder *ExternalHasherBuilder) Compile() (constraint.ConstraintSystemU32, error) {

	// As [GetWireConstraints] requires a list of variables and can only be
	// called once, we have to pack all the claims in a single slice and unpack
	// the result.
	allCheckedVariables := make([]frontend.Variable, 3*len(builder.claimTriplets))
	for i := range builder.claimTriplets {
		allCheckedVariables[3*i] = builder.claimTriplets[i][0]
		allCheckedVariables[3*i+1] = builder.claimTriplets[i][1]
		allCheckedVariables[3*i+2] = builder.claimTriplets[i][2]
	}

	// GetWireGates may add gates if [addGateForRangeCheck] is true. Call it
	// synchronously before calling compile on the circuit.
	cols, err := builder.Builder.GetWiresConstraintExact(allCheckedVariables, builder.addGateForHashCheck)

	if err != nil {
		return nil, fmt.Errorf("get wire gates: %w", err)
	}

	packedResult := make([][3][2]int, len(builder.claimTriplets))
	for i := range packedResult {
		packedResult[i] = [3][2]int{
			cols[3*i],
			cols[3*i+1],
			cols[3*i+2],
		}
	}

	// we pass the result in a goroutine until the wizard compiler is ready to
	// receive it
	go func() {
		builder.rcCols <- packedResult
	}()

	return builder.Builder.Compile()
}

// Compiler returns the compiler of the underlying builder.
func (builder *ExternalHasherBuilder) Compiler() frontend.Compiler {
	return builder.Builder.Compiler()
}

// Poseidon2Hintfunc is a gnark hint that computes the Poseidon2 compression function, it
// is used to return the pending claims of the evaluation of the Poseidon2 compression
// function.
func Poseidon2Hintfunc(f *big.Int, inputs []*big.Int, outputs []*big.Int) error {
	if f.String() != field.Modulus().String() {
		utils.Panic("Not the Koalabear field %d != %d", f, field.Modulus())
	}

	if len(inputs) != 16 {
		utils.Panic("expected 16 inputs, [init, block], got %v", len(inputs))
	}

	if len(outputs) != 8 {
		utils.Panic("expected 8 outputs [newState] got %v", len(outputs))
	}

	var old, block [poseidon2_koalabear.BlockSize]field.Element
	inpF := fromBigInts(inputs)
	copy(old[0:8], inpF[0:8])
	copy(block[0:8], inpF[8:16])

	outF := vortex.CompressPoseidon2(old, block)
	intoBigInts(outputs, outF[:])
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
func intoBigInts(res []*big.Int, arr []field.Element) []*big.Int {

	if len(res) != len(arr) {
		utils.Panic("got %v bigints but %v field elments", len(res), len(arr))
	}

	for i := range res {
		arr[i].BigInt(res[i])
	}
	return res
}
