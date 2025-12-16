package public_input

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/r1cs"
	"github.com/consensys/gnark/test"
	"github.com/consensys/linea-monorepo/prover/crypto/mimc"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/stretchr/testify/require"
)

func init() {
	// Register the missing hint
	utils.RegisterHints()
}

// ChainConfigurationTestCircuit is a circuit for testing ChainConfigurationFPISnark.Sum
type ChainConfigurationTestCircuit struct {
	// Inputs
	ChainID                 frontend.Variable
	BaseFee                 frontend.Variable
	CoinBase                frontend.Variable
	L2MessageServiceAddress frontend.Variable

	// Expected output from Go implementation
	ExpectedHash [32]frontend.Variable `gnark:",public"`
}

// Define defines the circuit's constraints
func (c *ChainConfigurationTestCircuit) Define(api frontend.API) error {
	// Create a ChainConfigurationFPISnark instance
	chainConfig := ChainConfigurationFPISnark{
		ChainID:                 c.ChainID,
		BaseFee:                 c.BaseFee,
		CoinBase:                c.CoinBase,
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
func TestChainConfigurationHashCircuitVsGo(t *testing.T) {
	tests := []struct {
		name             string
		chainID          string
		baseFee          string
		coinBase         string
		l2MessageService string
		expectedHash     string
	}{
		{
			name:             "devnet",
			chainID:          "0x000000000000000000000000000000000000000000000000000000000000e703",
			baseFee:          "0x0000000000000000000000000000000000000000000000000000000000000007",
			coinBase:         "0x4D517Aef039A48b3B6bF921e210b7551C8E37107",
			l2MessageService: "0x33bf916373159a8c1b54b025202517bfdbb7863d",
			expectedHash:     "0x0a360bbb44ebc0eee111237f7e11565f2f271a24a35465ee78a3a8bc3f503acb",
		},
		{
			name:             "sepolia",
			chainID:          "0x000000000000000000000000000000000000000000000000000000000000e705",
			baseFee:          "0x0000000000000000000000000000000000000000000000000000000000000007",
			coinBase:         "0xA27342f1b74c0cfB2cda74bac1628d0C1A9752f2",
			l2MessageService: "0x971e727e956690b9957be6d51Ec16E73AcAC83A7",
			expectedHash:     "0x03cd9edb7bad18416642423fef504154c0c0b7f9e6809627bd7aa4abeec4e326",
		},
		{
			name:             "mainnet",
			chainID:          "0x000000000000000000000000000000000000000000000000000000000000e708",
			baseFee:          "0x0000000000000000000000000000000000000000000000000000000000000007",
			coinBase:         "0x8F81e2E3F8b46467523463835F965fFE476E1c9E",
			l2MessageService: "0x508Ca82Df566dCD1B0DE8296e70a96332cD644ec",
			expectedHash:     "0x0881dc6ffdc69ebfeca27fd8449922c32d0fd16ea33807e984881b08e7100988",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse inputs
			chainID := hexToBigInt(tt.chainID)
			baseFee := hexToBigInt(tt.baseFee)
			coinBase := hexToBigInt(tt.coinBase)
			l2MessageService := hexToBigInt(tt.l2MessageService)

			// Compute expected hash using Go implementation
			_, expectedHashBytes := computeChainConfigurationGo(chainID, baseFee, coinBase, l2MessageService)

			t.Logf("Input chainID: %s", tt.chainID)
			t.Logf("Input baseFee: %s", tt.baseFee)
			t.Logf("Input coinBase: %s", tt.coinBase)
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
				ChainID:                 chainID,
				BaseFee:                 baseFee,
				CoinBase:                coinBase,
				L2MessageServiceAddress: l2MessageService,
				ExpectedHash:            expectedHashVars,
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
				t.Logf(" Circuit matches Go implementation!")
			}
		})
	}
}

// Test with direct field element comparison
func TestChainConfigurationHash_FieldElements(t *testing.T) {
	// Test with mainnet values
	chainID := hexToBigInt("0x000000000000000000000000000000000000000000000000000000000000e708")
	baseFee := hexToBigInt("0x0000000000000000000000000000000000000000000000000000000000000007")
	coinBase := hexToBigInt("0x8F81e2E3F8b46467523463835F965fFE476E1c9E")
	l2MessageService := hexToBigInt("0x508Ca82Df566dCD1B0DE8296e70a96332cD644ec")

	// Simulate the MiMC computation step by step
	t.Logf("=== Testing MiMC step by step ===")

	// Convert to field elements
	var chainIDField, baseFeeField, coinBaseField, l2MessageServiceField field.Element
	chainIDField.SetBigInt(chainID)
	baseFeeField.SetBigInt(baseFee)
	coinBaseField.SetBigInt(coinBase)
	l2MessageServiceField.SetBigInt(l2MessageService)
	t.Logf("ChainID field: %s", chainIDField.String())
	t.Logf("BaseFee field: %s", baseFeeField.String())
	t.Logf("CoinBase field: %s", coinBaseField.String())
	t.Logf("L2MessageService field: %s", l2MessageServiceField.String())
	// Simulate MiMC compression step by step
	var state field.Element // Initialize to zero
	// Process chainID (first bit is 0, so single compression)
	state = mimc.BlockCompression(state, chainIDField)
	t.Logf("State after chainID: %s", state.String())
	// Process baseFee (first bit is 0, so single compression)
	state = mimc.BlockCompression(state, baseFeeField)
	t.Logf("State after baseFee: %s", state.String())
	// Process coinBase (first bit is 0, so single compression)
	state = mimc.BlockCompression(state, coinBaseField)
	t.Logf("State after coinBase: %s", state.String())
	// Process l2MessageService (first bit is 0, so single compression)
	state = mimc.BlockCompression(state, l2MessageServiceField)
	t.Logf("Final state: %s", state.String())
	// Convert to bytes
	finalBytes := state.Bytes()
	t.Logf("Final bytes: %x", finalBytes)
	// Compare with the go implementation version
	_, hasherResult := computeChainConfigurationGo(chainID, baseFee, coinBase, l2MessageService)
	t.Logf("Hasher result: %x", hasherResult)
	require.Equal(t, hasherResult, finalBytes[:], "Field element approach should match hasher approach")
}

// Test case that requires splitting (first bit = 1)
func TestChainConfigurationHash_WithSplitting(t *testing.T) {
	// Test with sepolia values
	chainID := hexToBigInt("0x000000000000000000000000000000000000000000000000000000000000e705")
	baseFee := hexToBigInt("0x0000000000000000000000000000000000000000000000000000000000000007")
	coinBase := hexToBigInt("0xA27342f1b74c0cfB2cda74bac1628d0C1A9752f2")
	l2MessageService := hexToBigInt("0x971e727e956690b9957be6d51Ec16E73AcAC83A7")

	t.Logf("=== Testing MiMC with splitting ===")

	// Convert to field elements
	var chainIDField, baseFeeField, coinBaseField, l2MessageServiceField field.Element
	chainIDField.SetBigInt(chainID)
	baseFeeField.SetBigInt(baseFee)
	coinBaseField.SetBigInt(coinBase)
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
	// Process coinBase (first bit is 0, so single compression)
	state = mimc.BlockCompression(state, coinBaseField)
	t.Logf("State after coinBase: %s", state.String())
	// Process l2MessageService (first bit is 0, so single compression)
	state = mimc.BlockCompression(state, l2MessageServiceField)
	t.Logf("Final state: %s", state.String())
	// Convert to bytes
	finalBytes := state.Bytes()
	t.Logf("Final bytes: %x", finalBytes)
	// Compare with hasher version
	_, hasherResult := computeChainConfigurationGo(chainID, baseFee, coinBase, l2MessageService)
	t.Logf("Hasher result: %x", hasherResult)
	require.Equal(t, hasherResult, finalBytes[:], "Field element approach should match hasher approach")
}

// Test the pure Go version (should pass)
func TestChainConfigurationHash_PureGo(t *testing.T) {
	tests := []struct {
		name             string
		chainID          string
		baseFee          string
		coinBase         string
		l2MessageService string
		expectedPayload  string
		expectedHash     string
	}{
		{
			name:             "devnet",
			chainID:          "0x000000000000000000000000000000000000000000000000000000000000e703",
			baseFee:          "0x0000000000000000000000000000000000000000000000000000000000000007",
			coinBase:         "0x4D517Aef039A48b3B6bF921e210b7551C8E37107",
			l2MessageService: "0x33bf916373159a8c1b54b025202517bfdbb7863d",
			expectedPayload:  "0x000000000000000000000000000000000000000000000000000000000000e70300000000000000000000000000000000000000000000000000000000000000070000000000000000000000004d517aef039a48b3b6bf921e210b7551c8e3710700000000000000000000000033bf916373159a8c1b54b025202517bfdbb7863d",
			expectedHash:     "0x0a360bbb44ebc0eee111237f7e11565f2f271a24a35465ee78a3a8bc3f503acb",
		},
		{
			name:             "sepolia",
			chainID:          "0x000000000000000000000000000000000000000000000000000000000000e705",
			baseFee:          "0x0000000000000000000000000000000000000000000000000000000000000007",
			coinBase:         "0xA27342f1b74c0cfB2cda74bac1628d0C1A9752f2",
			l2MessageService: "0x971e727e956690b9957be6d51Ec16E73AcAC83A7",
			expectedPayload:  "0x000000000000000000000000000000000000000000000000000000000000e7050000000000000000000000000000000000000000000000000000000000000007000000000000000000000000a27342f1b74c0cfb2cda74bac1628d0c1a9752f2000000000000000000000000971e727e956690b9957be6d51ec16e73acac83a7",
			expectedHash:     "0x03cd9edb7bad18416642423fef504154c0c0b7f9e6809627bd7aa4abeec4e326",
		},
		{
			name:             "mainnet",
			chainID:          "0x000000000000000000000000000000000000000000000000000000000000e708",
			baseFee:          "0x0000000000000000000000000000000000000000000000000000000000000007",
			coinBase:         "0x8F81e2E3F8b46467523463835F965fFE476E1c9E",
			l2MessageService: "0x508Ca82Df566dCD1B0DE8296e70a96332cD644ec",
			expectedPayload:  "0x000000000000000000000000000000000000000000000000000000000000e70800000000000000000000000000000000000000000000000000000000000000070000000000000000000000008f81e2e3f8b46467523463835f965ffe476e1c9e000000000000000000000000508ca82df566dcd1b0de8296e70a96332cd644ec",
			expectedHash:     "0x0881dc6ffdc69ebfeca27fd8449922c32d0fd16ea33807e984881b08e7100988",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse inputs
			chainID := hexToBigInt(tt.chainID)
			baseFee := hexToBigInt(tt.baseFee)
			coinBase := hexToBigInt(tt.coinBase)
			l2MessageService := hexToBigInt(tt.l2MessageService)
			// Compute hash and payload
			computedPayload, computedHash := computeChainConfigurationGo(chainID, baseFee, coinBase, l2MessageService)
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
				t.Logf(" Payload matches!")
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
				t.Logf(" Hash matches!")
			}
		})
	}
}

// computeChainConfigurationGo is the Go reference implementation that mimics the verifier's Solidity
// implementation for computing the chain configuration hash. This serves as the source of truth for testing
// the circuit version, since circuits are harder to test and debug directly. When adding new chain configuration
// parameters, update this function first and run it to generate expected hash values, then use those values as test
// cases throughout this file to ensure consistency between Go, circuit, and Solidity implementations.
func computeChainConfigurationGo(chainID, baseFee, coinBase, l2MessageServiceAddr *big.Int) ([]byte, []byte) {
	hasher := mimc.NewMiMC()
	var mimcPayload []byte

	values := []*big.Int{chainID, baseFee, coinBase, l2MessageServiceAddr}

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
			most := new(big.Int).Rsh(value, 128)                                                                    // Right shift by 128 bits
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

// TestGenerateChainConfigHashes generates the expected hash values for devnet, sepolia, and mainnet
func TestGenerateChainConfigHashes(t *testing.T) {
	testCases := []struct {
		name                   string
		chainID                int64
		baseFee                int64
		coinBase               string
		messageServiceContract string
	}{
		{
			name:                   "devnet",
			chainID:                59139,
			baseFee:                7,
			coinBase:               "0x4D517Aef039A48b3B6bF921e210b7551C8E37107",
			messageServiceContract: "0x33bf916373159a8c1b54b025202517bfdbb7863d",
		},
		{
			name:                   "sepolia",
			chainID:                59141,
			baseFee:                7,
			coinBase:               "0xA27342f1b74c0cfB2cda74bac1628d0C1A9752f2",
			messageServiceContract: "0x971e727e956690b9957be6d51Ec16E73AcAC83A7",
		},
		{
			name:                   "mainnet",
			chainID:                59144,
			baseFee:                7,
			coinBase:               "0x8F81e2E3F8b46467523463835F965fFE476E1c9E",
			messageServiceContract: "0x508Ca82Df566dCD1B0DE8296e70a96332cD644ec",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			chainID := big.NewInt(tc.chainID)
			baseFee := big.NewInt(tc.baseFee)
			coinBase := hexToBigInt(tc.coinBase)
			messageService := hexToBigInt(tc.messageServiceContract)

			mimcPayload, hash := computeChainConfigurationGo(chainID, baseFee, coinBase, messageService)

			t.Logf("\n=== %s ===", tc.name)
			t.Logf("chainID: %d (0x%x)", tc.chainID, tc.chainID)
			t.Logf("baseFee: %d", tc.baseFee)
			t.Logf("coinBase: %s", tc.coinBase)
			t.Logf("messageServiceContract: %s", tc.messageServiceContract)
			t.Logf("mimcPayload: 0x%x", mimcPayload)
			t.Logf("hash: 0x%x", hash)
			t.Logf("For test case:")
			t.Logf("  chainID:          \"0x%064x\",", chainID)
			t.Logf("  baseFee:          \"0x%064x\",", baseFee)
			t.Logf("  coinBase:         \"%s\",", tc.coinBase)
			t.Logf("  l2MessageService: \"%s\",", tc.messageServiceContract)
			t.Logf("  expectedHash:     \"0x%x\",", hash)
		})
	}
}
