package v0

import (
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	fr381 "github.com/consensys/gnark-crypto/ecc/bls12-381/fr"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	"github.com/consensys/gnark/std/math/emulated"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/crypto/mimc/gkrmimc"
)

type hashAnyTestCircuit struct {
	A   frontend.Variable `gnark:",secret"`
	B   frontend.Variable `gnark:",secret"`
	C   emElement         `gnark:",secret"`
	D   emElement         `gnark:",secret"`
	Out frontend.Variable `gnark:",secret"`
}

func (h *hashAnyTestCircuit) Define(api frontend.API) error {
	hf := gkrmimc.NewHasherFactory(api)
	x := mimcHashAnyGnark(api, hf, h.A, h.B, h.C, h.D)
	api.AssertIsEqual(x, h.Out)
	return nil
}

func assignHashAnyTestCircuit(a, b fr.Element, c, d fr381.Element) *hashAnyTestCircuit {
	out := mimcHashAny(a, b, c, d)
	return &hashAnyTestCircuit{
		A:   a,
		B:   b,
		C:   emulated.ValueOf[emFr](c),
		D:   emulated.ValueOf[emFr](d),
		Out: out,
	}
}

func TestHashGnarkAny(t *testing.T) {

	c := &hashAnyTestCircuit{}
	a := assignHashAnyTestCircuit(
		fr.NewElement(32),
		fr.NewElement(234),
		fr381.NewElement(9089),
		fr381.NewElement(3446),
	)

	ccs, err := frontend.Compile(ecc.BLS12_377.ScalarField(), scs.NewBuilder, c)
	if err != nil {
		t.Fatalf("could not compile the circuit: %v", err)
	}

	wit, err := frontend.NewWitness(a, ecc.BLS12_377.ScalarField())
	if err != nil {
		t.Fatalf("could not get the witness: %v", err)
	}

	err = ccs.IsSolved(wit)
	if err != nil {
		t.Fatalf("circuit not solved: %v", err.Error())
	}
}
