package hasherfactory_koalabear

import (
	"math/big"
	"testing"

	cs "github.com/consensys/gnark/constraint/koalabear"

	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/gnark-crypto/field/koalabear/vortex"
	"github.com/consensys/gnark/constraint/solver"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/crypto/poseidon2_koalabear"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/plonkinternal/plonkbuilder"
	"golang.org/x/crypto/sha3"
)

func init() {
	if err := poseidon2_koalabear.RegisterGates(); err != nil {
		panic(err)
	}
	solver.RegisterHint(testWideCommitHint)
}

func testWideCommitHint(m *big.Int, inputs []*big.Int, outputs []*big.Int) error {
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

// externalPoseidon2FactoryTestLinear is used to test the external Poseidon2 factory
// and is a gnark circuit implementing a linear hash.
type externalPoseidon2FactoryTestLinear struct {
	Inp [16]frontend.Variable
}

// Define implements the gnark frontend.Circuit interface.
// It is a test circuit to compare the ExternalHasherFactory with the BasicHasherFactory.
// It takes 16 inputs and compute the poseidon2 hash of the inputs using both factories.
// The two results are then compared to ensure they are equal.
func (circuit *externalPoseidon2FactoryTestLinear) Define(api frontend.API) error {

	var (
		factory      = &ExternalHasherFactory{Api: api}
		factoryBasic = &BasicHasherFactory{Api: api}
		hasher       = factory.NewHasher()
		hasherBasic  = factoryBasic.NewHasher()
	)

	hasher.Write(circuit.Inp[:]...)
	hasherBasic.Write(circuit.Inp[:]...)
	hsum := hasher.Sum()
	hsumBasic := hasherBasic.Sum()
	for i := 0; i < poseidon2_koalabear.BlockSize; i++ {
		api.AssertIsEqual(hsum[i], hsumBasic[i])
	}

	return nil
}

func TestPoseidon2Factories(t *testing.T) {

	var (
		circuit            = &externalPoseidon2FactoryTestLinear{}
		builder, hshGetter = NewExternalHasherBuilder(true)
		koalaField         = koalabear.Modulus()
		assignment         = &externalPoseidon2FactoryTestLinear{Inp: [16]frontend.Variable{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}}
	)

	solver.RegisterHint(Poseidon2Hintfunc)

	ccs_, compileErr := frontend.CompileU32(koalaField, builder, circuit)

	if compileErr != nil {
		t.Fatalf("unexpected compile error: %v", compileErr)
	}

	ccs := ccs_.(*cs.SparseR1CS)

	witness, wErr := frontend.NewWitness(assignment, koalaField)

	if wErr != nil {
		t.Fatalf("unexpected witness error: %v", wErr)
	}

	sol_, solErr := ccs.Solve(witness,
		solver.OverrideHint(solver.GetHintID(plonkbuilder.PlaceholderWideCommitHint), testWideCommitHint),
	)

	if solErr != nil {
		t.Fatalf("unexpected solution error: %v", solErr)
	}

	var (
		sol        = sol_.(*cs.SparseR1CSSolution)
		hshWires   = hshGetter()
		getFromLRO = func(csID, colID int) field.Element {

			if colID == 0 {
				return sol.L[csID]
			}

			if colID == 1 {
				return sol.R[csID]
			}
			return sol.O[csID]
		}
	)
	var (
		oldState [poseidon2_koalabear.BlockSize]field.Element
		block    [poseidon2_koalabear.BlockSize]field.Element
		newState [poseidon2_koalabear.BlockSize]field.Element
	)
	for i, triplet := range hshWires {

		oldState[i%poseidon2_koalabear.BlockSize] = getFromLRO(triplet[0][0], triplet[0][1])
		block[i%poseidon2_koalabear.BlockSize] = getFromLRO(triplet[1][0], triplet[1][1])
		newState[i%poseidon2_koalabear.BlockSize] = getFromLRO(triplet[2][0], triplet[2][1])

		if i%poseidon2_koalabear.BlockSize == poseidon2_koalabear.BlockSize-1 {
			newState_ := vortex.CompressPoseidon2(oldState, block)
			if newState != newState_ {
				t.Errorf("expected %v, got %v", newState, newState_)
			}
		}

	}
}
