package invalidity

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

var (
	// Test values
	testAddressHi     = field.NewElement(0xDEAD)
	testAddressLo     = field.NewElement(0xBEEF)
	testTxHashHi      = field.NewElement(0x1234)
	testTxHashLo      = field.NewElement(0x5678)
	testStateRootHash = field.NewElement(0x1234567890abcdef)
	colSize           = 16 // power of 2
)

// computeExpectedAddress computes the expected FromAddress field element
// from the test AddressHi and AddressLo values
func computeExpectedAddress() field.Element {
	// The address is built from:
	// - 4 bytes from AddressHi (bytes 28-31 of the 32-byte representation)
	// - 16 bytes from AddressLo (bytes 16-31 of the 32-byte representation)
	hiBytes := testAddressHi.Bytes()
	loBytes := testAddressLo.Bytes()
	var b [20]byte
	copy(b[:4], hiBytes[28:])
	copy(b[4:], loBytes[16:])
	var result field.Element
	result.SetBytes(b[:])
	return result
}

// testCase defines a test case for InvalidityPI
type testCase struct {
	name             string
	hasBadPrecompile bool
	numL2Logs        int
}

// TestInvalidityPIAssign tests that public inputs are accessible and return correct values
func TestInvalidityPIAssign(t *testing.T) {

	testCases := []testCase{
		{
			name:             "no_bad_precompile_no_logs",
			hasBadPrecompile: false,
			numL2Logs:        0,
		},
		{
			name:             "has_bad_precompile_no_logs",
			hasBadPrecompile: true,
			numL2Logs:        0,
		},
		{
			name:             "no_bad_precompile_with_logs",
			hasBadPrecompile: false,
			numL2Logs:        3,
		},
		{
			name:             "has_bad_precompile_with_logs",
			hasBadPrecompile: true,
			numL2Logs:        24, // > MAX_L2_LOGS (16)
		},
	}

	// Build FixedInputs from test data
	fixedInputs := FixedInputs{
		StateRootHash: testStateRootHash,
		TxHashHi:      testTxHashHi,
		TxHashLo:      testTxHashLo,
		AddressHi:     testAddressHi,
		AddressLo:     testAddressLo,
		FromAddress:   computeExpectedAddress(),
		ColSize:       colSize,
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			// Build inputs for MockZkevmArithCols
			inputs := Inputs{
				FixedInputs: fixedInputs,
				CaseInputs: CaseInputs{
					HasBadPrecompile: tc.hasBadPrecompile,
					NumL2Logs:        tc.numL2Logs,
				},
			}

			// Use MockZkevmArithCols to create the wizard and proof
			comp, proof := MockZkevmArithCols(inputs)

			// Verify the wizard proof
			err := wizard.Verify(comp, proof)
			if err != nil {
				t.Fatalf("verification failed: %v", err)
			}

			// Check public inputs are accessible and have correct values
			checkPublicInputsFromProof(t, comp, proof, tc)
		})
	}
}

// checkPublicInputsFromProof verifies that all InvalidityPI public inputs
// are registered and have correct values in the proof
func checkPublicInputsFromProof(t *testing.T, comp *wizard.CompiledIOP, proof wizard.Proof, tc testCase) {

	// StateRootHash
	stateRootHash := proof.GetPublicInput(comp, StateRootHash)
	if !stateRootHash.Equal(&testStateRootHash) {
		t.Errorf("StateRootHash mismatch: got %v, want %v", stateRootHash, testStateRootHash)
	}

	// TxHashHi
	txHashHi := proof.GetPublicInput(comp, TxHashHi)
	if !txHashHi.Equal(&testTxHashHi) {
		t.Errorf("TxHashHi mismatch: got %v, want %v", txHashHi, testTxHashHi)
	}

	// TxHashLo
	txHashLo := proof.GetPublicInput(comp, TxHashLo)
	if !txHashLo.Equal(&testTxHashLo) {
		t.Errorf("TxHashLo mismatch: got %v, want %v", txHashLo, testTxHashLo)
	}

	// FromAddress
	expectedAddress := computeExpectedAddress()
	fromAddress := proof.GetPublicInput(comp, FromAddress)
	if !fromAddress.Equal(&expectedAddress) {
		t.Errorf("FromAddress mismatch: got %v, want %v", fromAddress, expectedAddress)
	}

	// HasBadPrecompile
	hasBadPrecompile := proof.GetPublicInput(comp, HasBadPrecompile)
	if tc.hasBadPrecompile && !hasBadPrecompile.IsOne() {
		t.Errorf("HasBadPrecompile mismatch: got %v, want %v", hasBadPrecompile, field.One())
	}

	// NbL2Logs
	expectedNbL2Logs := field.NewElement(uint64(tc.numL2Logs))
	nbL2Logs := proof.GetPublicInput(comp, NbL2Logs)
	if !nbL2Logs.Equal(&expectedNbL2Logs) {
		t.Errorf("NbL2Logs mismatch: got %v, want %v", nbL2Logs, expectedNbL2Logs)
	}
}
