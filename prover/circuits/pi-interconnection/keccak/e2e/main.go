package main

import (
	"encoding/hex"
	"fmt"
	"os"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/r1cs"
	"github.com/consensys/linea-monorepo/prover/circuits/internal"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak"
)

// TestCircuit hashes multiple blocks of data using keccak
type TestCircuit struct {
	Keccak keccak.StrictHasherCircuit

	// Inputs to hash (variable number of hashes, each with varying numbers of blocks)
	Inputs [][][32]frontend.Variable

	// Expected outputs
	Outputs [][32]frontend.Variable `gnark:",public"`
}

func (c *TestCircuit) Define(api frontend.API) error {
	hasher := c.Keccak.NewHasher(api)

	// Hash each input
	for i, input := range c.Inputs {
		out := hasher.Sum(nil, input...)
		internal.AssertSliceEquals(api, out[:], c.Outputs[i][:])
	}

	return hasher.Finalize()
}

type testCase struct {
	name   string
	data   string
	length int // length in bytes (must be multiple of 32)
}

func run(paramsComment string, params keccak.CompilationParams) {
	fmt.Printf("=== Keccak End-to-End Test (%s) ===\n", paramsComment)
	fmt.Println()

	// Define test cases
	testCases := []testCase{
		{name: "Hello, World!", data: "Hello, World!", length: 64},
		{name: "Keccak test", data: "Keccak test", length: 96},
		{name: "Short", data: "Hi", length: 32},
	}

	// Step 1: Compile the keccak hasher
	fmt.Println("Step 1: Compiling keccak hasher...")
	compiler := keccak.NewStrictHasherCompiler(0)
	for _, tc := range testCases {
		compiler.WithStrictHashLengths(tc.length)
	}
	compiled := compiler.Compile(params)
	fmt.Println("✓ Keccak hasher compiled")
	fmt.Println()

	// Step 2: Get the hasher and hash all test data
	fmt.Println("Step 2: Hashing test data...")
	hasher := compiled.GetHasher()

	inputs := make([][]byte, len(testCases))
	outputs := make([][]byte, len(testCases))

	for i, tc := range testCases {
		input := make([]byte, tc.length)
		copy(input, tc.data)
		hasher.Reset()
		_, err := hasher.Write(input)
		if err != nil {
			fmt.Printf("Error writing to hasher: %v\n", err)
			os.Exit(1)
		}
		output := hasher.Sum(nil)
		inputs[i] = input
		outputs[i] = output
		fmt.Printf("Test %d: %s (padded to %d bytes)\n", i+1, tc.name, tc.length)
		fmt.Printf("  Output: %s\n", hex.EncodeToString(output))
	}
	fmt.Println()

	// Step 3: Create the circuit assignment
	fmt.Println("Step 3: Creating circuit assignment...")
	keccak.RegisterHints()
	internal.RegisterHints()

	assignment := TestCircuit{
		Inputs:  make([][][32]frontend.Variable, len(testCases)),
		Outputs: make([][32]frontend.Variable, len(testCases)),
	}

	// Copy all inputs and outputs
	for i, tc := range testCases {
		nbBlocks := tc.length / 32
		assignment.Inputs[i] = make([][32]frontend.Variable, nbBlocks)
		for j := range nbBlocks {
			block := [32]frontend.Variable{}
			for k := range 32 {
				block[k] = inputs[i][j*32+k]
			}
			assignment.Inputs[i][j] = block
		}

		for j := range 32 {
			assignment.Outputs[i][j] = outputs[i][j]
		}
	}

	// Assign the keccak circuit
	var err error
	assignment.Keccak, err = hasher.Assign()
	if err != nil {
		fmt.Printf("Error assigning keccak circuit: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ Circuit assignment created")
	fmt.Println()

	// Step 4: Get the circuit definition
	fmt.Println("Step 4: Creating circuit definition...")
	circuit := TestCircuit{
		Inputs:  make([][][32]frontend.Variable, len(testCases)),
		Outputs: make([][32]frontend.Variable, len(testCases)),
	}
	for i, tc := range testCases {
		nbBlocks := tc.length / 32
		circuit.Inputs[i] = make([][32]frontend.Variable, nbBlocks)
		for j := range nbBlocks {
			circuit.Inputs[i][j] = [32]frontend.Variable{}
		}
	}
	circuit.Keccak, err = compiled.GetCircuit()
	if err != nil {
		fmt.Printf("Error getting circuit: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ Circuit definition created")
	fmt.Println()

	// Step 5: Compile the R1CS
	fmt.Println("Step 5: Compiling R1CS (this may take a while)...")
	ccs, err := frontend.Compile(ecc.BLS12_377.ScalarField(), r1cs.NewBuilder, &circuit)
	if err != nil {
		fmt.Printf("Error compiling R1CS: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✓ R1CS compiled (%d constraints)\n", ccs.GetNbConstraints())
	fmt.Println()

	// Step 6: Setup (Generate proving and verifying keys)
	fmt.Println("Step 6: Running trusted setup (this will take several minutes)...")
	pk, vk, err := groth16.Setup(ccs)
	if err != nil {
		fmt.Printf("Error in setup: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ Trusted setup complete")
	fmt.Println()

	// Step 7: Create witness
	fmt.Println("Step 7: Creating witness...")
	witness, err := frontend.NewWitness(&assignment, ecc.BLS12_377.ScalarField())
	if err != nil {
		fmt.Printf("Error creating witness: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ Witness created")
	fmt.Println()

	// Step 8: Generate proof
	fmt.Println("Step 8: Generating proof (this may take a while)...")
	proof, err := groth16.Prove(ccs, pk, witness)
	if err != nil {
		fmt.Printf("Error generating proof: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ Proof generated")
	fmt.Println()

	// Step 9: Verify proof
	fmt.Println("Step 9: Verifying proof...")
	publicWitness, err := witness.Public()
	if err != nil {
		fmt.Printf("Error extracting public witness: %v\n", err)
		os.Exit(1)
	}
	err = groth16.Verify(proof, vk, publicWitness)
	if err != nil {
		fmt.Printf("❌ Verification failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ Proof verified successfully!")
	fmt.Println()

	fmt.Println("=== Test Complete ===")
	fmt.Printf("Successfully verified %d keccak hashes:\n", len(testCases))
	for i, tc := range testCases {
		fmt.Printf("  %d. %s (%d bytes)\n", i+1, tc.name, tc.length)
	}
	fmt.Println("The keccak package successfully:")
	fmt.Println("  - Compiled the wizard-based keccak circuit")
	fmt.Println("  - Hashed data during assignment")
	fmt.Println("  - Generated a valid SNARK proof")
	fmt.Println("  - Verified the proof")
}

func main() {
	run("dummy", keccak.DummyCompile())
	//run("full", keccak.WizardCompilationParameters())
}
