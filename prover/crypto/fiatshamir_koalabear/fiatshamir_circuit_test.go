package fiatshamir_koalabear

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/gnark/constraint"
	"github.com/consensys/gnark/constraint/solver"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	"github.com/consensys/linea-monorepo/prover/crypto/poseidon2_koalabear"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/maths/field/koalagnark"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/sha3"
)

// wideCommitBuilder wraps a U32 builder to implement frontend.WideCommitter,
// which is required for multi-instance GKR on small fields.
type wideCommitBuilder struct {
	frontend.Builder[constraint.U32]
}

func wideCommitWrapper(newBuilder frontend.NewBuilderU32) frontend.NewBuilderU32 {
	return func(field *big.Int, config frontend.CompileConfig) (frontend.Builder[constraint.U32], error) {
		b, err := newBuilder(field, config)
		if err != nil {
			return nil, err
		}
		return &wideCommitBuilder{b}, nil
	}
}

func (w *wideCommitBuilder) WideCommit(width int, toCommit ...frontend.Variable) ([]frontend.Variable, error) {
	return w.NewHint(wideCommitHint, width, toCommit...)
}

func wideCommitHint(m *big.Int, inputs []*big.Int, outputs []*big.Int) error {
	nb := (m.BitLen() + 7) / 8
	buf := make([]byte, nb)
	hasher := sha3.NewCShake128(nil, []byte("gnark test engine"))
	for _, in := range inputs {
		bs := in.FillBytes(buf)
		hasher.Write(bs)
	}
	for i := range outputs {
		hasher.Read(buf)
		outputs[i].SetBytes(buf)
		outputs[i].Mod(outputs[i], m)
	}
	return nil
}

func init() {
	solver.RegisterHint(wideCommitHint)
	if err := poseidon2_koalabear.RegisterGates(); err != nil {
		panic(err)
	}
}

// Test for UpdateExt with field extension elements
type UpdateExtCircuit struct {
	ExtInputs [3]koalagnark.Ext
	Output    poseidon2_koalabear.GnarkOctuplet
}

func (c *UpdateExtCircuit) Define(api frontend.API) error {
	fs := NewGnarkFSWV(api)

	fs.UpdateExt(c.ExtInputs[:]...)
	res := fs.RandomField()
	for i := 0; i < 8; i++ {
		api.AssertIsEqual(res[i].Native(), c.Output[i])
	}
	return nil
}

func TestUpdateExt(t *testing.T) {
	fs := NewFS()

	var extInputs [3]fext.Element
	for i := 0; i < 3; i++ {
		extInputs[i].SetRandom()
	}

	fs.UpdateExt(extInputs[:]...)
	output := fs.RandomField()

	var circuit, witness UpdateExtCircuit
	for i := 0; i < 3; i++ {
		witness.ExtInputs[i] = koalagnark.NewExt(extInputs[i])
	}
	for i := 0; i < 8; i++ {
		witness.Output[i] = output[i]
	}

	ccs, err := frontend.CompileU32(koalabear.Modulus(), wideCommitWrapper(scs.NewBuilder), &circuit)
	assert.NoError(t, err)

	fullWitness, err := frontend.NewWitness(&witness, koalabear.Modulus())
	assert.NoError(t, err)
	err = ccs.IsSolved(fullWitness)
	assert.NoError(t, err)
}

// Test for RandomFieldExt
type RandomFieldExtCircuit struct {
	Input     [2]koalagnark.Element
	OutputExt koalagnark.Ext
}

func (c *RandomFieldExtCircuit) Define(api frontend.API) error {
	fs := NewGnarkFSWV(api)

	fs.Update(c.Input[:]...)
	res := fs.RandomFieldExt()

	koalaAPI := koalagnark.NewAPI(api)
	koalaAPI.AssertIsEqualExt(res, c.OutputExt)
	return nil
}

func TestRandomFieldExt(t *testing.T) {
	fs := NewFS()

	var input [2]field.Element
	input[0].SetRandom()
	input[1].SetRandom()

	fs.Update(input[:]...)
	output := fs.RandomFext()

	var circuit, witness RandomFieldExtCircuit
	witness.Input[0] = koalagnark.NewElement(input[0].String())
	witness.Input[1] = koalagnark.NewElement(input[1].String())
	witness.OutputExt = koalagnark.NewExt(output)

	ccs, err := frontend.CompileU32(koalabear.Modulus(), wideCommitWrapper(scs.NewBuilder), &circuit)
	assert.NoError(t, err)

	fullWitness, err := frontend.NewWitness(&witness, koalabear.Modulus())
	assert.NoError(t, err)
	err = ccs.IsSolved(fullWitness)
	assert.NoError(t, err)
}

// Test for UpdateVec with multiple vectors
type UpdateVecCircuit struct {
	Vec1   [3]koalagnark.Element
	Vec2   [4]koalagnark.Element
	Output poseidon2_koalabear.GnarkOctuplet
}

func (c *UpdateVecCircuit) Define(api frontend.API) error {
	fs := NewGnarkFSWV(api)

	fs.UpdateVec(c.Vec1[:], c.Vec2[:])
	res := fs.RandomField()
	for i := 0; i < 8; i++ {
		api.AssertIsEqual(res[i].Native(), c.Output[i])
	}
	return nil
}

func TestUpdateVec(t *testing.T) {
	fs := NewFS()

	var vec1 [3]field.Element
	var vec2 [4]field.Element
	for i := 0; i < 3; i++ {
		vec1[i].SetRandom()
	}
	for i := 0; i < 4; i++ {
		vec2[i].SetRandom()
	}

	fs.UpdateVec(vec1[:], vec2[:])
	output := fs.RandomField()

	var circuit, witness UpdateVecCircuit
	for i := 0; i < 3; i++ {
		witness.Vec1[i] = koalagnark.NewElement(vec1[i].String())
	}
	for i := 0; i < 4; i++ {
		witness.Vec2[i] = koalagnark.NewElement(vec2[i].String())
	}
	for i := 0; i < 8; i++ {
		witness.Output[i] = output[i]
	}

	ccs, err := frontend.CompileU32(koalabear.Modulus(), wideCommitWrapper(scs.NewBuilder), &circuit)
	assert.NoError(t, err)

	fullWitness, err := frontend.NewWitness(&witness, koalabear.Modulus())
	assert.NoError(t, err)
	err = ccs.IsSolved(fullWitness)
	assert.NoError(t, err)
}

// Test for RandomManyIntegers with different bounds
type RandomManyIntegersCircuit struct {
	Input  [5]koalagnark.Element
	Output []koalagnark.Element
	n      int
	bound  int
}

func (c *RandomManyIntegersCircuit) Define(api frontend.API) error {
	fs := NewGnarkFSWV(api)

	fs.Update(c.Input[:]...)
	res := fs.RandomManyIntegers(c.n, c.bound)
	for i := 0; i < len(res); i++ {
		api.AssertIsEqual(res[i], c.Output[i].Native())
	}
	return nil
}

func TestRandomManyIntegersVariousBounds(t *testing.T) {
	testCases := []struct {
		name  string
		n     int
		bound int
	}{
		{"small_bound", 3, 4},
		{"medium_bound", 5, 16},
		{"large_bound", 10, 128},
		{"power_of_two_bound", 8, 32},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fs := NewFS()

			var input [5]field.Element
			for i := 0; i < 5; i++ {
				input[i].SetRandom()
			}

			fs.Update(input[:]...)
			output := fs.RandomManyIntegers(tc.n, tc.bound)

			var circuit, witness RandomManyIntegersCircuit
			circuit.n = tc.n
			circuit.bound = tc.bound
			circuit.Output = make([]koalagnark.Element, tc.n)
			witness.Output = make([]koalagnark.Element, tc.n)

			for i := 0; i < 5; i++ {
				witness.Input[i] = koalagnark.NewElement(input[i].String())
			}
			for i := 0; i < tc.n; i++ {
				witness.Output[i] = koalagnark.NewElement(output[i])
			}

			ccs, err := frontend.CompileU32(koalabear.Modulus(), wideCommitWrapper(scs.NewBuilder), &circuit)
			assert.NoError(t, err)

			fullWitness, err := frontend.NewWitness(&witness, koalabear.Modulus())
			assert.NoError(t, err)
			err = ccs.IsSolved(fullWitness)
			assert.NoError(t, err)
		})
	}
}

// Test for SetState and State round-trip
type StateRoundTripCircuit struct {
	InitialState poseidon2_koalabear.GnarkOctuplet
	FinalState   poseidon2_koalabear.GnarkOctuplet
}

func (c *StateRoundTripCircuit) Define(api frontend.API) error {
	fs := NewGnarkFSWV(api)

	var oct koalagnark.Octuplet
	for i := range oct {
		oct[i] = koalagnark.WrapFrontendVariable(c.InitialState[i])
	}
	fs.SetState(oct)
	state := fs.State()

	for i := 0; i < 8; i++ {
		api.AssertIsEqual(state[i].Native(), c.FinalState[i])
	}
	return nil
}

func TestStateRoundTrip(t *testing.T) {
	fs := NewFS()

	var initialState field.Octuplet
	for i := 0; i < 8; i++ {
		initialState[i].SetRandom()
	}

	fs.SetState(initialState)
	finalState := fs.State()

	var circuit, witness StateRoundTripCircuit
	for i := 0; i < 8; i++ {
		witness.InitialState[i] = initialState[i]
		witness.FinalState[i] = finalState[i]
	}

	ccs, err := frontend.CompileU32(koalabear.Modulus(), wideCommitWrapper(scs.NewBuilder), &circuit)
	assert.NoError(t, err)

	fullWitness, err := frontend.NewWitness(&witness, koalabear.Modulus())
	assert.NoError(t, err)
	err = ccs.IsSolved(fullWitness)
	assert.NoError(t, err)
}

// Test for multiple sequential RandomField calls
type MultipleRandomFieldCircuit struct {
	Input   [4]koalagnark.Element
	Output1 poseidon2_koalabear.GnarkOctuplet
	Output2 poseidon2_koalabear.GnarkOctuplet
}

func (c *MultipleRandomFieldCircuit) Define(api frontend.API) error {
	fs := NewGnarkFSWV(api)

	fs.Update(c.Input[:]...)
	res1 := fs.RandomField()
	res2 := fs.RandomField()

	for i := 0; i < 8; i++ {
		api.AssertIsEqual(res1[i].Native(), c.Output1[i])
		api.AssertIsEqual(res2[i].Native(), c.Output2[i])
	}
	return nil
}

func TestMultipleRandomFieldCalls(t *testing.T) {
	fs := NewFS()

	var input [4]field.Element
	for i := 0; i < 4; i++ {
		input[i].SetRandom()
	}

	fs.Update(input[:]...)
	output1 := fs.RandomField()
	output2 := fs.RandomField()

	var circuit, witness MultipleRandomFieldCircuit
	for i := 0; i < 4; i++ {
		witness.Input[i] = koalagnark.NewElement(input[i].String())
	}
	for i := 0; i < 8; i++ {
		witness.Output1[i] = output1[i]
		witness.Output2[i] = output2[i]
	}

	ccs, err := frontend.CompileU32(koalabear.Modulus(), wideCommitWrapper(scs.NewBuilder), &circuit)
	assert.NoError(t, err)

	fullWitness, err := frontend.NewWitness(&witness, koalabear.Modulus())
	assert.NoError(t, err)
	err = ccs.IsSolved(fullWitness)
	assert.NoError(t, err)
}

// Test edge case with zero values
type ZeroValuesCircuit struct {
	Input  [4]koalagnark.Element
	Output poseidon2_koalabear.GnarkOctuplet
}

func (c *ZeroValuesCircuit) Define(api frontend.API) error {
	fs := NewGnarkFSWV(api)

	fs.Update(c.Input[:]...)
	res := fs.RandomField()
	for i := 0; i < 8; i++ {
		api.AssertIsEqual(res[i].Native(), c.Output[i])
	}
	return nil
}

func TestZeroValues(t *testing.T) {
	fs := NewFS()

	var input [4]field.Element
	// All zeros
	for i := 0; i < 4; i++ {
		input[i].SetZero()
	}

	fs.Update(input[:]...)
	output := fs.RandomField()

	var circuit, witness ZeroValuesCircuit
	for i := 0; i < 4; i++ {
		witness.Input[i] = koalagnark.NewElement(input[i].String())
	}
	for i := 0; i < 8; i++ {
		witness.Output[i] = output[i]
	}

	ccs, err := frontend.CompileU32(koalabear.Modulus(), wideCommitWrapper(scs.NewBuilder), &circuit)
	assert.NoError(t, err)

	fullWitness, err := frontend.NewWitness(&witness, koalabear.Modulus())
	assert.NoError(t, err)
	err = ccs.IsSolved(fullWitness)
	assert.NoError(t, err)
}

// Test edge case with max values
type MaxValuesCircuit struct {
	Input  [4]koalagnark.Element
	Output poseidon2_koalabear.GnarkOctuplet
}

func (c *MaxValuesCircuit) Define(api frontend.API) error {
	fs := NewGnarkFSWV(api)

	fs.Update(c.Input[:]...)
	res := fs.RandomField()
	for i := 0; i < 8; i++ {
		api.AssertIsEqual(res[i].Native(), c.Output[i])
	}
	return nil
}

func TestMaxValues(t *testing.T) {
	fs := NewFS()

	var input [4]field.Element
	// All max values (p-1)
	for i := 0; i < 4; i++ {
		input[i] = *field.MaxVal
	}

	fs.Update(input[:]...)
	output := fs.RandomField()

	var circuit, witness MaxValuesCircuit
	for i := 0; i < 4; i++ {
		witness.Input[i] = koalagnark.NewElement(input[i].String())
	}
	for i := 0; i < 8; i++ {
		witness.Output[i] = output[i]
	}

	ccs, err := frontend.CompileU32(koalabear.Modulus(), wideCommitWrapper(scs.NewBuilder), &circuit)
	assert.NoError(t, err)

	fullWitness, err := frontend.NewWitness(&witness, koalabear.Modulus())
	assert.NoError(t, err)
	err = ccs.IsSolved(fullWitness)
	assert.NoError(t, err)
}

// Flush444Circuit writes 444 KoalaBear elements via Update + RandomField,
// exercising the native KoalaBear Poseidon2 GKR path.
type Flush444Circuit struct {
	Input  [444]koalagnark.Element
	Output poseidon2_koalabear.GnarkOctuplet
}

func (c *Flush444Circuit) Define(api frontend.API) error {
	fs := NewGnarkFSWV(api)

	fs.Update(c.Input[:]...)
	res := fs.RandomField()
	for i := 0; i < 8; i++ {
		api.AssertIsEqual(res[i].Native(), c.Output[i])
	}
	return nil
}

func TestFlush444KoalaConstraints(t *testing.T) {
	fs := NewFS()

	var input [444]field.Element
	for i := 0; i < 444; i++ {
		input[i].SetRandom()
	}

	fs.Update(input[:]...)
	output := fs.RandomField()

	var circuit, witness Flush444Circuit
	for i := 0; i < 444; i++ {
		witness.Input[i] = koalagnark.NewElement(input[i].String())
	}
	for i := 0; i < 8; i++ {
		witness.Output[i] = output[i]
	}

	ccs, err := frontend.CompileU32(koalabear.Modulus(), wideCommitWrapper(scs.NewBuilder), &circuit)
	assert.NoError(t, err)
	t.Logf("KoalaBear native 444 elements: %d constraints", ccs.GetNbConstraints())

	fullWitness, err := frontend.NewWitness(&witness, koalabear.Modulus())
	assert.NoError(t, err)
	err = ccs.IsSolved(fullWitness)
	assert.NoError(t, err)
}

// Test for mixed zero and max values
type MixedValuesCircuit struct {
	Input  [8]koalagnark.Element
	Output poseidon2_koalabear.GnarkOctuplet
}

func (c *MixedValuesCircuit) Define(api frontend.API) error {
	fs := NewGnarkFSWV(api)

	fs.Update(c.Input[:]...)
	res := fs.RandomField()
	for i := 0; i < 8; i++ {
		api.AssertIsEqual(res[i].Native(), c.Output[i])
	}
	return nil
}

func TestMixedValues(t *testing.T) {
	fs := NewFS()

	var input [8]field.Element
	for i := 0; i < 8; i++ {
		if i%2 == 0 {
			input[i].SetZero()
		} else {
			input[i] = *field.MaxVal
		}
	}

	fs.Update(input[:]...)
	output := fs.RandomField()

	var circuit, witness MixedValuesCircuit
	for i := 0; i < 8; i++ {
		witness.Input[i] = koalagnark.NewElement(input[i].String())
	}
	for i := 0; i < 8; i++ {
		witness.Output[i] = output[i]
	}

	ccs, err := frontend.CompileU32(koalabear.Modulus(), wideCommitWrapper(scs.NewBuilder), &circuit)
	assert.NoError(t, err)

	fullWitness, err := frontend.NewWitness(&witness, koalabear.Modulus())
	assert.NoError(t, err)
	err = ccs.IsSolved(fullWitness)
	assert.NoError(t, err)
}

// GKRKoalaBearHashCircuit hashes KoalaBear elements directly using native
// GKR Poseidon2 over KoalaBear (width 16, block size 8).
type GKRKoalaBearHashCircuit struct {
	Input []koalagnark.Element
}

func (c *GKRKoalaBearHashCircuit) Define(api frontend.API) error {
	hasher, err := poseidon2_koalabear.NewGnarkMDHasher(api)
	if err != nil {
		return err
	}
	for _, elem := range c.Input {
		hasher.Write(elem.Native())
	}
	_ = hasher.Sum()
	return nil
}

func TestGKRPoseidon2KoalaBearConstraints(t *testing.T) {
	sizes := []int{8, 64, 512}
	for _, n := range sizes {
		t.Run(fmt.Sprintf("n=%d", n), func(t *testing.T) {
			circuit := &GKRKoalaBearHashCircuit{
				Input: make([]koalagnark.Element, n),
			}
			ccs, err := frontend.CompileU32(koalabear.Modulus(), wideCommitWrapper(scs.NewBuilder), circuit)
			assert.NoError(t, err)
			t.Logf("GKR Poseidon2 KoalaBear (%d koala elements, %d blocks): %d constraints",
				n, n/8, ccs.GetNbConstraints())
		})
	}
}

func BenchmarkGKRPoseidon2KoalaBearSolve(b *testing.B) {
	sizes := []int{8, 64, 512}
	for _, n := range sizes {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			circuit := &GKRKoalaBearHashCircuit{
				Input: make([]koalagnark.Element, n),
			}
			assignment := &GKRKoalaBearHashCircuit{
				Input: make([]koalagnark.Element, n),
			}
			for i := range n {
				v := field.NewElement(uint64(i + 1))
				assignment.Input[i] = koalagnark.NewElement(v.String())
			}
			ccs, err := frontend.CompileU32(koalabear.Modulus(), wideCommitWrapper(scs.NewBuilder), circuit)
			if err != nil {
				b.Fatal(err)
			}
			b.Logf("constraints: %d", ccs.GetNbConstraints())
			fullWitness, err := frontend.NewWitness(assignment, koalabear.Modulus())
			if err != nil {
				b.Fatal(err)
			}
			b.ResetTimer()
			for range b.N {
				if err := ccs.IsSolved(fullWitness); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
