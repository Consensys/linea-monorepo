package invalidity

import (
	"fmt"
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
// are registered and have correct values 
func checkPublicInputsFromProof(t *testing.T, comp *wizard.CompiledIOP, proof wizard.Proof, tc testCase) {

	// Check StateRootHash limbs (8 limbs)
	for i := 0; i < zkevmcommon.NbLimbU128; i++ {
		name := fmt.Sprintf("%s_%d", StateRootHashName, i)
		val := proof.GetPublicInput(comp, name, false)
		expected := testStateRootLimbs[i]
		if val.GetIsBase() {
			gotField, err := val.GetBase()
			if err != nil {
				t.Errorf("StateRootHash_%d: failed to get base value: %v", i, err)
				continue
			}
			if !gotField.Equal(&expected) {
				t.Errorf("StateRootHash_%d mismatch: got %v, want %v", i, gotField, expected)
			}
		}
	}

	// Check TxHash limbs (16 limbs)
	for i := 0; i < zkevmcommon.NbLimbU256; i++ {
		name := fmt.Sprintf("%s_%d", TxHashName, i)
		val := proof.GetPublicInput(comp, name, false)
		expected := testTxHashLimbs[i]
		if val.GetIsBase() {
			gotField, err := val.GetBase()
			if err != nil {
				t.Errorf("TxHash_%d: failed to get base value: %v", i, err)
				continue
			}
			if !gotField.Equal(&expected) {
				t.Errorf("TxHash_%d mismatch: got %v, want %v", i, gotField, expected)
			}
		}
	}

	// Check From limbs (10 limbs)
	for i := 0; i < zkevmcommon.NbLimbEthAddress; i++ {
		name := fmt.Sprintf("%s_%d", FromName, i)
		val := proof.GetPublicInput(comp, name, false)
		expected := testFromLimbs[i]
		if val.GetIsBase() {
			gotField, err := val.GetBase()
			if err != nil {
				t.Errorf("From_%d: failed to get base value: %v", i, err)
				continue
			}
			if !gotField.Equal(&expected) {
				t.Errorf("From_%d mismatch: got %v, want %v", i, gotField, expected)
			}
		}
	}

	// HasBadPrecompile
	hasBadPrecompile := proof.GetPublicInput(comp, HasBadPrecompileName, false)
	if tc.hasBadPrecompile && !hasBadPrecompile.IsOne() {
		t.Errorf("HasBadPrecompile mismatch: got %v, want 1", hasBadPrecompile)
	}

	// NbL2Logs
	expectedNbL2Logs := field.NewElement(uint64(tc.numL2Logs))
	nbL2Logs := proof.GetPublicInput(comp, NbL2LogsName, false)
	if nbL2Logs.GetIsBase() {
		gotField, err := nbL2Logs.GetBase()
		if err != nil {
			t.Errorf("NbL2Logs: failed to get base value: %v", err)
		} else if !gotField.Equal(&expectedNbL2Logs) {
			t.Errorf("NbL2Logs mismatch: got %v, want %v", gotField, expectedNbL2Logs)
		}
	}
}
