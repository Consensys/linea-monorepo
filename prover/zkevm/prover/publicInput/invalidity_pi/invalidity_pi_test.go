package invalidity

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	zkevmcommon "github.com/consensys/linea-monorepo/prover/zkevm/prover/common"
)

var (
	// Test values - these are limb values (16-bit each)
	testTxHashLimbs    [zkevmcommon.NbLimbU256]field.Element
	testFromLimbs      [zkevmcommon.NbLimbEthAddress]field.Element
	testStateRootLimbs [zkevmcommon.NbLimbU128]field.Element
	colSize            = 16 // power of 2
)

func init() {
	// Initialize test limb values
	for i := range testTxHashLimbs {
		testTxHashLimbs[i] = field.NewElement(uint64(0x1000 + i))
	}
	for i := range testFromLimbs {
		testFromLimbs[i] = field.NewElement(uint64(0x2000 + i))
	}
	for i := range testStateRootLimbs {
		testStateRootLimbs[i] = field.NewElement(uint64(0x3000 + i))
	}
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
		TxHashLimbs:    testTxHashLimbs,
		FromLimbs:      testFromLimbs,
		StateRootLimbs: testStateRootLimbs,
		ColSize:        colSize,
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
// are registered and have correct values using the extractor pattern
func checkPublicInputsFromProof(t *testing.T, comp *wizard.CompiledIOP, proof wizard.Proof, tc testCase) {

	// Retrieve the extractor from ExtraData
	extraData, found := comp.ExtraData[InvalidityPIExtractorMetadata]
	if !found {
		t.Fatal("InvalidityPIExtractor not found in ExtraData")
	}
	extractor, ok := extraData.(*InvalidityPIExtractor)
	if !ok {
		t.Fatalf("ExtraData[InvalidityPIExtractorMetadata] has wrong type: %T", extraData)
	}

	// Helper function to get public input value from proof
	getPI := func(pi wizard.PublicInput) field.Element {
		val := proof.GetPublicInput(comp, pi.Name, false)
		if !val.GetIsBase() {
			t.Fatalf("Public input %s is not a base value", pi.Name)
		}
		fieldVal, err := val.GetBase()
		if err != nil {
			t.Fatalf("Failed to get base value for %s: %v", pi.Name, err)
		}
		return fieldVal
	}

	// Check StateRootHash limbs (8 limbs)
	for i := 0; i < zkevmcommon.NbLimbU128; i++ {
		gotField := getPI(extractor.StateRootHash[i])
		expected := testStateRootLimbs[i]
		if !gotField.Equal(&expected) {
			t.Errorf("StateRootHash_%d mismatch: got %v, want %v", i, gotField, expected)
		}
	}

	// Check TxHash limbs (16 limbs)
	// Public inputs are in BE order, so we need to reverse the expected values
	for i := 0; i < zkevmcommon.NbLimbU256; i++ {
		gotField := getPI(extractor.TxHash[i])
		expectedIdx := zkevmcommon.NbLimbU256 - 1 - i // Reverse index for BE
		expected := testTxHashLimbs[expectedIdx]
		if !gotField.Equal(&expected) {
			t.Errorf("TxHash_%d mismatch: got %v, want %v", i, gotField, expected)
		}
	}

	// Check From limbs (10 limbs)
	// Public inputs are now in BE order, so we need to reverse the expected values
	for i := 0; i < zkevmcommon.NbLimbEthAddress; i++ {
		gotField := getPI(extractor.FromAddress[i])
		expectedIdx := zkevmcommon.NbLimbEthAddress - 1 - i // Reverse index for BE
		expected := testFromLimbs[expectedIdx]
		if !gotField.Equal(&expected) {
			t.Errorf("From_%d mismatch: got %v, want %v", i, gotField, expected)
		}
	}

	// HasBadPrecompile
	hasBadPrecompileVal := proof.GetPublicInput(comp, extractor.HasBadPrecompile.Name, false)
	if tc.hasBadPrecompile && !hasBadPrecompileVal.IsOne() {
		t.Errorf("HasBadPrecompile mismatch: got %v, want 1", hasBadPrecompileVal)
	}

	// NbL2Logs
	expectedNbL2Logs := field.NewElement(uint64(tc.numL2Logs))
	gotNbL2Logs := getPI(extractor.NbL2Logs)
	if !gotNbL2Logs.Equal(&expectedNbL2Logs) {
		t.Errorf("NbL2Logs mismatch: got %v, want %v", gotNbL2Logs, expectedNbL2Logs)
	}
}
