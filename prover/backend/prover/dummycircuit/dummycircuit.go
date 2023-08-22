package dummycircuit

import (
	"math/big"
	"path"

	"github.com/consensys/accelerated-crypto-monorepo/backend/prover/plonkutil"
	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/kzg"
	"github.com/consensys/gnark/backend/plonk"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
)

// Circuit, that is verified by only one public input
// This relies on the fact that x^5 is a permutation
// in Fr.
type Circuit struct {
	X  frontend.Variable `gnark:",public"`
	X5 frontend.Variable `gnark:",secret"`
}

func (c *Circuit) Define(api frontend.API) error {
	x5 := api.Mul(c.X, c.X, c.X, c.X, c.X)
	committer, _ := api.(frontend.Committer)
	_, err := committer.Commit(c.X)
	if err != nil {
		panic(err)
	}
	api.AssertIsEqual(x5, c.X5)
	return nil
}

// Generates a deterministic (and unsafe) setup
func GenPublicParamsUnsafe() (pp plonkutil.Setup) {
	circuit := &Circuit{}

	scs, err := frontend.Compile(fr.Modulus(), scs.NewBuilder, circuit)
	if err != nil {
		panic(err)
	}

	// Deterministic (and unsafe) SRS generation
	srs, err := kzg.NewSRS(16, big.NewInt(42))
	if err != nil {
		panic(err)
	}

	pk, vk, err := plonk.Setup(scs, srs)
	if err != nil {
		panic(err)
	}

	return plonkutil.Setup{PK: pk, VK: vk, SCS: scs}
}

func assign(x field.Element) *Circuit {
	var x5 fr.Element
	x5.Exp(x, big.NewInt(5))
	return &Circuit{X: x, X5: x5}
}

func MakeProof(pp plonkutil.Setup, x fr.Element) string {

	assignment := assign(x)

	witness, err := frontend.NewWitness(assignment, ecc.BN254.ScalarField())
	if err != nil {
		panic(err)
	}

	proof, err := plonk.Prove(pp.SCS, pp.PK, witness)
	if err != nil {
		panic(err)
	}

	// Sanity-check : the proof must pass
	{
		pubwitness, err := frontend.NewWitness(assignment, ecc.BN254.ScalarField(), frontend.PublicOnly())
		if err != nil {
			panic(err)
		}

		err = plonk.Verify(proof, pp.VK, pubwitness)
		if err != nil {
			panic(err)
		}
	}

	// Write the serialized proof
	return plonkutil.SerializeProof(proof)
}

// Generate the light setup
func GenSetupLight(folderOut string) {
	pp := GenPublicParamsUnsafe()
	plonkutil.ExportToFile(pp, path.Join(folderOut, "light"))
}
