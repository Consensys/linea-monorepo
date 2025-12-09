package hasher_factory

import (
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

// externalHasherBuilder is an implementation of the [frontend.Builder]
// interface.
type ExternalHasherBuilder struct {
	plonkbuilder.Builder
	// claimTriplets stores the tripled [oldState, block, newState]
	claimTriplets [][3][poseidon2_koalabear.BlockSize]frontend.Variable
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
	CheckHashExternally(oldState, block, newState [poseidon2_koalabear.BlockSize]frontend.Variable)
}

// NewHasher returns the standard Poseidon2 hasher.
func (f *BasicHasherFactory) NewHasher() poseidon2_koalabear.GnarkMDHasher {
	h, _ := poseidon2_koalabear.NewGnarkMDHasher(f.Api)
	return h
}

// NewHasher returns an external Poseidon2 hasher.
func (f *ExternalHasherFactory) NewHasher() poseidon2_koalabear.GnarkMDHasher {

	h, _ := poseidon2_koalabear.NewGnarkMDHasher(f.Api)

	return h
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
func (f *ExternalHasherBuilder) CheckHashExternally(oldState, block, newState [poseidon2_koalabear.BlockSize]frontend.Variable) {
	f.claimTriplets = append(f.claimTriplets, [3][poseidon2_koalabear.BlockSize]frontend.Variable{oldState, block, newState})
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
