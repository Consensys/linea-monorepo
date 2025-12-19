package hasher_factory_test

import (
	"testing"

	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	hf "github.com/consensys/linea-monorepo/prover/crypto/hasher_factory"
	"github.com/consensys/linea-monorepo/prover/maths/field"
)

type MsetOfSingletonGnarkTestCircuit struct {
	X [24]frontend.Variable
	R [hf.MSetHashSize]frontend.Variable
}

func (circuit *MsetOfSingletonGnarkTestCircuit) Define(api frontend.API) error {
	r := hf.MsetOfSingletonGnark(api, nil, circuit.X[:]...)
	r.AssertEqualRaw(api, circuit.R[:])
	return nil
}

func TestMSetHash(t *testing.T) {

	var (
		circuit  = &MsetOfSingletonGnarkTestCircuit{}
		assigned = &MsetOfSingletonGnarkTestCircuit{}
		msg      = [24]int{
			1, 2, 3, 4, 5, 6, 7, 8,
			9, 10, 11, 12, 13, 14, 15, 16,
			17, 18, 19, 20, 21, 22, 23, 24,
		}
		msgField = []field.Element{}
		mset     = hf.MSetHash{}
	)

	for i := range msg {

		msgField = append(msgField, field.NewElement(uint64(msg[i])))

		assigned.X[i] = msg[i]

	}

	mset.Insert(msgField...)

	for i := range mset {
		assigned.R[i] = mset[i]
	}

	ccs, compileErr := frontend.CompileU32(koalabear.Modulus(), scs.NewBuilder, circuit)
	if compileErr != nil {
		t.Fatalf("unexpected compile error: %v", compileErr)
	}

	witness, wErr := frontend.NewWitness(assigned, koalabear.Modulus())
	if wErr != nil {
		t.Fatalf("unexpected witness error: %v", wErr)
	}

	_, solErr := ccs.Solve(witness)
	if solErr != nil {
		t.Fatalf("unexpected solution error: %v", solErr)
	}
}
