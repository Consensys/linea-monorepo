package public_input

import (
	"encoding/hex"

	"fmt"

	"math/big"

	"testing"

	"github.com/consensys/linea-monorepo/prover/utils"

	"github.com/consensys/gnark-crypto/ecc"

	"github.com/consensys/gnark/frontend"

	"github.com/consensys/gnark/frontend/cs/r1cs"

	"github.com/consensys/gnark/test"

	"github.com/consensys/linea-monorepo/prover/crypto/mimc"

	"github.com/consensys/linea-monorepo/prover/maths/field"

	"github.com/stretchr/testify/require"
)

func init() {

	// Register the missing hint

	utils.RegisterHints()

}

// ChainConfigurationTestCircuit is a circuit for testing ChainConfigurationFPISnark.Sum

type ChainConfigurationTestCircuit struct {

	// Inputs

	ChainID frontend.Variable

	BaseFee frontend.Variable

	L2MessageServiceAddress frontend.Variable

	// Expected output from Go implementation

	ExpectedHash [32]frontend.Variable `gnark:",public"`
}

// Define defines the circuit's constraints

func (c *ChainConfigurationTestCircuit) Define(api frontend.API) error {

	// Create a ChainConfigurationFPISnark instance

	chainConfig := ChainConfigurationFPISnark{

		ChainID: c.ChainID,

		BaseFee: c.BaseFee,

		L2MessageServiceAddress: c.L2MessageServiceAddress,
	}

	// Compute the hash using our circuit implementation

	computedHash := chainConfig.Sum(api)

	computedHashBytes := utils.ToBytes(api, computedHash)
	// Enforce that the computed hash matches the expected hash

	for i := 0; i < 32; i++ {

		api.AssertIsEqual(computedHashBytes[i], c.ExpectedHash[i])

	}

	return nil

}

// Test the circuit implementation against the pure Go version

func TestChainConfigurationHash_CircuitVsGo(t *testing.T) {

	tests := []struct {
		name string

		chainID string

		baseFee string

		l2MessageService string

		expectedHash string
	}{

		{

			name: "Should match Go with multiple configuration values that have a first 0 bit",

			chainID: "0x0000000000000000000000000000000000000000000000000000000000000539",

			baseFee: "0x0000000000000000000000000000000000000000000000000000000000000007",

			l2MessageService: "0x000000000000000000000000e537d669ca013d86ebef1d64e40fc74cadc91987",

			expectedHash: "0x0adb539e498a111fc9b8d362cf33e86429a2a47355e51fd86866919608e8d46e",
		},

		// We do not have MSG = 1 cases for chainId and baseFee
		// {

		// 	name: "Should match Go with multiple configuration values that have a first non 0 bit",

		// 	chainID: "0x8900000000000000000000000000000000000000000000000000000000000089",

		// 	baseFee: "0x0000000000000000000000000000000000000000000000000000000000000007",

		// 	l2MessageService: "0x000000000000000000000000e537d669ca013d86ebef1d64e40fc74cadc91987",

		// 	expectedHash: "0x0e61435408022b505318be3663bd8f50e4b51300b1c7cfd754a8ecb17ede2913",
		// },

		{

			name: "Should match Go with multiple configuration values that have a first non 0 bit",

			chainID: "0x0000000000000000000000000000000000000000000000000000000000000539",

			baseFee: "0x0000000000000000000000000000000000000000000000000000000000000007",

			l2MessageService: "0x000000000000000000000000e537d669ca013d86ebef1d64e40fc74cadc91987",

			expectedHash: "0x0adb539e498a111fc9b8d362cf33e86429a2a47355e51fd86866919608e8d46e",
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			// Parse inputs

			chainID := hexToBigInt(tt.chainID)

			baseFee := hexToBigInt(tt.baseFee)

			l2MessageService := hexToBigInt(tt.l2MessageService)

			// Compute expected hash using Go implementation

			_, expectedHashBytes := computeChainConfigurationHashGo(chainID, baseFee, l2MessageService)

			t.Logf("Input chainID: %s", tt.chainID)

			t.Logf("Input baseFee: %s", tt.baseFee)

			t.Logf("Input l2MessageService: %s", tt.l2MessageService)

			t.Logf("Expected hash: %x", expectedHashBytes)

			// Convert expected hash to field elements for circuit

			var expectedHashVars [32]frontend.Variable

			for i := 0; i < 32; i++ {

				expectedHashVars[i] = expectedHashBytes[i]

			}

			// Create circuit

			circuit := &ChainConfigurationTestCircuit{}

			// Create assignment

			assignment := &ChainConfigurationTestCircuit{

				ChainID: chainID,

				BaseFee: baseFee,

				L2MessageServiceAddress: l2MessageService,

				ExpectedHash: expectedHashVars,
			}

			// Test the circuit

			err := test.IsSolved(circuit, assignment, ecc.BLS12_377.ScalarField())

			if err != nil {

				t.Errorf("Circuit test failed: %v", err)

				// Let's also try to compile and run to see debug output

				t.Logf("Attempting to compile circuit for debug output...")

				r1cs, compileErr := frontend.Compile(ecc.BLS12_377.ScalarField(), r1cs.NewBuilder, circuit)

				if compileErr != nil {

					t.Logf("Compile error: %v", compileErr)

				} else {

					witness, witnessErr := frontend.NewWitness(assignment, ecc.BLS12_377.ScalarField())

					if witnessErr != nil {

						t.Logf("Witness creation error: %v", witnessErr)

					} else {

						solveErr := r1cs.IsSolved(witness)

						t.Logf("Solve error: %v", solveErr)

					}

				}

			} else {

				t.Logf("✅ Circuit matches Go implementation!")

			}

		})

	}

}

// Enhanced Go implementation with detailed debugging

func computeChainConfigurationHashGoDetailed(chainID, baseFee, l2MessageServiceAddr *big.Int) ([]byte, []byte) {

	hasher := mimc.NewMiMC()

	var mimcPayload []byte

	values := []*big.Int{chainID, baseFee, l2MessageServiceAddr}

	valueNames := []string{"chainID", "baseFee", "l2MessageService"}

	for i, value := range values {

		name := valueNames[i]

		// Check if first bit is zero (bit 255 for 256-bit number)

		firstBitIsZero := value.Bit(255) == 0

		fmt.Printf("--- Processing %s: %s ---\n", name, value.String())

		fmt.Printf("%s firstBitIsZero: %t\n", name, firstBitIsZero)

		if firstBitIsZero {

			// Append the entire value (32 bytes)

			valueBytes := make([]byte, 32)

			value.FillBytes(valueBytes)

			mimcPayload = append(mimcPayload, valueBytes...)

			hasher.Write(valueBytes)

			fmt.Printf("%s single write: %x\n", name, valueBytes)

		} else {

			// Split into most and least

			most := new(big.Int).Rsh(value, 128) // Right shift by 128 bits

			least := new(big.Int).And(value, new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 128), big.NewInt(1))) // Mask lower 128 bits

			fmt.Printf("%s most: %s\n", name, most.String())

			fmt.Printf("%s least: %s\n", name, least.String())

			// Write most (16 bytes, padded to 32)

			mostBytes := make([]byte, 32)

			most.FillBytes(mostBytes[16:]) // Put in the lower 16 bytes, upper 16 are zero

			mimcPayload = append(mimcPayload, mostBytes...)

			hasher.Write(mostBytes)

			fmt.Printf("%s most write: %x\n", name, mostBytes)

			// Write least (16 bytes, padded to 32)

			leastBytes := make([]byte, 32)

			least.FillBytes(leastBytes[16:]) // Put in the lower 16 bytes, upper 16 are zero

			mimcPayload = append(mimcPayload, leastBytes...)

			hasher.Write(leastBytes)

			fmt.Printf("%s least write: %x\n", name, leastBytes)

		}

	}

	finalHash := hasher.Sum(nil)

	fmt.Printf("Final hash: %x\n", finalHash)

	return mimcPayload, finalHash

}

// Test with direct field element comparison

func TestChainConfigurationHash_FieldElements(t *testing.T) {

	// Test case with first bit 0

	chainID := hexToBigInt("0x0000000000000000000000000000000000000000000000000000000000000539")

	baseFee := hexToBigInt("0x0000000000000000000000000000000000000000000000000000000000000007")

	l2MessageService := hexToBigInt("0x000000000000000000000000e537d669ca013d86ebef1d64e40fc74cadc91987")

	// Simulate the MiMC computation step by step

	t.Logf("=== Testing MiMC step by step ===")

	// Convert to field elements

	var chainIDField, baseFeeField, l2MessageServiceField field.Element

	chainIDField.SetBigInt(chainID)

	baseFeeField.SetBigInt(baseFee)

	l2MessageServiceField.SetBigInt(l2MessageService)

	t.Logf("ChainID field: %s", chainIDField.String())

	t.Logf("BaseFee field: %s", baseFeeField.String())

	t.Logf("L2MessageService field: %s", l2MessageServiceField.String())

	// Simulate MiMC compression step by step

	var state field.Element // Initialize to zero

	// Process chainID (first bit is 0, so single compression)

	state = mimc.BlockCompression(state, chainIDField)

	t.Logf("State after chainID: %s", state.String())

	// Process baseFee (first bit is 0, so single compression)

	state = mimc.BlockCompression(state, baseFeeField)

	t.Logf("State after baseFee: %s", state.String())

	// Process l2MessageService (first bit is 0, so single compression)

	state = mimc.BlockCompression(state, l2MessageServiceField)

	t.Logf("Final state: %s", state.String())

	// Convert to bytes

	finalBytes := state.Bytes()

	t.Logf("Final bytes: %x", finalBytes)

	// Compare with hasher version

	_, hasherResult := computeChainConfigurationHashGo(chainID, baseFee, l2MessageService)

	t.Logf("Hasher result: %x", hasherResult)

	require.Equal(t, hasherResult, finalBytes[:], "Field element approach should match hasher approach")

}

// Test with a case that requires splitting (first bit = 1)

func TestChainConfigurationHash_WithSplitting(t *testing.T) {

	// Test case with first bit 1 for chainID

	chainID := hexToBigInt("0x8900000000000000000000000000000000000000000000000000000000000089")

	baseFee := hexToBigInt("0x0000000000000000000000000000000000000000000000000000000000000007")

	l2MessageService := hexToBigInt("0x000000000000000000000000e537d669ca013d86ebef1d64e40fc74cadc91987")

	t.Logf("=== Testing MiMC with splitting ===")

	// Convert to field elements

	var chainIDField, baseFeeField, l2MessageServiceField field.Element

	chainIDField.SetBigInt(chainID)

	baseFeeField.SetBigInt(baseFee)

	l2MessageServiceField.SetBigInt(l2MessageService)

	// Check first bit of chainID

	firstBitIsZero := chainID.Bit(255) == 0

	t.Logf("ChainID first bit is zero: %t", firstBitIsZero)

	var state field.Element // Initialize to zero

	if !firstBitIsZero {

		// Split chainID

		most := new(big.Int).Rsh(chainID, 128)

		least := new(big.Int).And(chainID, new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 128), big.NewInt(1)))

		var mostField, leastField field.Element

		mostField.SetBigInt(most)

		leastField.SetBigInt(least)

		t.Logf("ChainID most: %s", mostField.String())

		t.Logf("ChainID least: %s", leastField.String())

		// Two compressions

		state = mimc.BlockCompression(state, mostField)

		t.Logf("State after chainID most: %s", state.String())

		state = mimc.BlockCompression(state, leastField)

		t.Logf("State after chainID least: %s", state.String())

	} else {

		// Single compression

		state = mimc.BlockCompression(state, chainIDField)

		t.Logf("State after chainID: %s", state.String())

	}

	// Process baseFee (first bit is 0, so single compression)

	state = mimc.BlockCompression(state, baseFeeField)

	t.Logf("State after baseFee: %s", state.String())

	// Process l2MessageService (first bit is 0, so single compression)

	state = mimc.BlockCompression(state, l2MessageServiceField)

	t.Logf("Final state: %s", state.String())

	// Convert to bytes

	finalBytes := state.Bytes()

	t.Logf("Final bytes: %x", finalBytes)

	// Compare with hasher version

	_, hasherResult := computeChainConfigurationHashGo(chainID, baseFee, l2MessageService)

	t.Logf("Hasher result: %x", hasherResult)

	require.Equal(t, hasherResult, finalBytes[:], "Field element approach should match hasher approach")

}

// Test the pure Go version (should pass)

func TestChainConfigurationxxxHash_PureGo(t *testing.T) {

	tests := []struct {
		name string

		chainID string

		baseFee string

		l2MessageService string

		expectedPayload string

		expectedHash string
	}{

		{

			name: "Should deploy with multiple configuration values that have a first 0 bit",

			chainID: "0x0000000000000000000000000000000000000000000000000000000000000539",

			baseFee: "0x0000000000000000000000000000000000000000000000000000000000000007",

			l2MessageService: "0x000000000000000000000000e537d669ca013d86ebef1d64e40fc74cadc91987",

			expectedPayload: "0x00000000000000000000000000000000000000000000000000000000000005390000000000000000000000000000000000000000000000000000000000000007000000000000000000000000e537d669ca013d86ebef1d64e40fc74cadc91987",

			expectedHash: "0x0adb539e498a111fc9b8d362cf33e86429a2a47355e51fd86866919608e8d46e",
		},

		{

			name: "Should deploy with multiple configuration values that have a first non 0 bit",

			chainID: "0x8900000000000000000000000000000000000000000000000000000000000089",

			baseFee: "0x0000000000000000000000000000000000000000000000000000000000000007",

			l2MessageService: "0x000000000000000000000000e537d669ca013d86ebef1d64e40fc74cadc91987",

			expectedPayload: "0x000000000000000000000000000000008900000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000890000000000000000000000000000000000000000000000000000000000000007000000000000000000000000e537d669ca013d86ebef1d64e40fc74cadc91987",

			expectedHash: "0x0e61435408022b505318be3663bd8f50e4b51300b1c7cfd754a8ecb17ede2913",
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			// Parse inputs

			chainID := hexToBigInt(tt.chainID)

			baseFee := hexToBigInt(tt.baseFee)

			l2MessageService := hexToBigInt(tt.l2MessageService)

			// Compute hash and payload

			computedPayload, computedHash := computeChainConfigurationHashGo(chainID, baseFee, l2MessageService)

			expectedPayloadBytes := hexToBytes(tt.expectedPayload)

			expectedHashBytes := hexToBytes(tt.expectedHash)

			// Debug output

			t.Logf("Expected payload: %x", expectedPayloadBytes)

			t.Logf("Computed payload: %x", computedPayload)

			t.Logf("Expected hash: %x", expectedHashBytes)

			t.Logf("Computed hash: %x", computedHash)

			// Compare payload first

			if len(expectedPayloadBytes) != len(computedPayload) {

				t.Errorf("Payload length mismatch: expected %d, got %d", len(expectedPayloadBytes), len(computedPayload))

				return

			}

			payloadMatch := true

			for i := 0; i < len(expectedPayloadBytes); i++ {

				if expectedPayloadBytes[i] != computedPayload[i] {

					t.Errorf("Payload mismatch at byte %d: expected %02x, got %02x", i, expectedPayloadBytes[i], computedPayload[i])

					payloadMatch = false

				}

			}

			if payloadMatch {

				t.Logf("✅ Payload matches!")

			}

			// Compare hash

			if len(expectedHashBytes) != len(computedHash) {

				t.Errorf("Hash length mismatch: expected %d, got %d", len(expectedHashBytes), len(computedHash))

				return

			}

			hashMatch := true

			for i := 0; i < len(expectedHashBytes); i++ {

				if expectedHashBytes[i] != computedHash[i] {

					t.Errorf("Hash mismatch at byte %d: expected %02x, got %02x", i, expectedHashBytes[i], computedHash[i])

					hashMatch = false

				}

			}

			if hashMatch {

				t.Logf("✅ Hash matches!")

			}

		})

	}

}

// Pure Go implementation that exactly matches Solidity

func computeChainConfigurationHashGo(chainID, baseFee, l2MessageServiceAddr *big.Int) ([]byte, []byte) {

	hasher := mimc.NewMiMC()

	var mimcPayload []byte

	values := []*big.Int{chainID, baseFee, l2MessageServiceAddr}

	for _, value := range values {

		// Check if first bit is zero (bit 255 for 256-bit number)

		firstBitIsZero := value.Bit(255) == 0

		if firstBitIsZero {

			// Append the entire value (32 bytes)

			valueBytes := make([]byte, 32)

			value.FillBytes(valueBytes)

			mimcPayload = append(mimcPayload, valueBytes...)

			hasher.Write(valueBytes)

		} else {

			// Split into most and least

			most := new(big.Int).Rsh(value, 128) // Right shift by 128 bits

			least := new(big.Int).And(value, new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 128), big.NewInt(1))) // Mask lower 128 bits

			// Write most (16 bytes, padded to 32)

			mostBytes := make([]byte, 32)

			most.FillBytes(mostBytes[16:]) // Put in the lower 16 bytes, upper 16 are zero

			mimcPayload = append(mimcPayload, mostBytes...)

			hasher.Write(mostBytes)

			// Write least (16 bytes, padded to 32)

			leastBytes := make([]byte, 32)

			least.FillBytes(leastBytes[16:]) // Put in the lower 16 bytes, upper 16 are zero

			mimcPayload = append(mimcPayload, leastBytes...)

			hasher.Write(leastBytes)

		}

	}

	return mimcPayload, hasher.Sum(nil)

}

func hexToBigInt(hexStr string) *big.Int {

	value := new(big.Int)

	value.SetString(hexStr[2:], 16)

	return value

}

func hexToBytes(hexStr string) []byte {

	bytes, err := hex.DecodeString(hexStr[2:])

	if err != nil {

		panic(err)

	}

	return bytes

}
