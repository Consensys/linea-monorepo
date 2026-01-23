package invalidityPI

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/accessors"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/ecdsa"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/publicInput/logs"
)

// mockInvalidityPIInputs holds mock columns for testing InvalidityPI
type mockInvalidityPIInputs struct {
	// Mock badPrecompile column
	BadPrecompileCol ifaces.Column

	// Mock Addresses columns
	AddressHi            ifaces.Column
	AddressLo            ifaces.Column
	IsAddressFromTxnData ifaces.Column

	// Mock TxSignature columns
	TxHashHi ifaces.Column
	TxHashLo ifaces.Column
	IsTxHash ifaces.Column

	// Mock ExtractedData columns
	FilterFetched ifaces.Column
}

var (
	// Test values
	testAddressHi     = field.NewElement(0xDEAD)
	testAddressLo     = field.NewElement(0xBEEF)
	testTxHashHi      = field.NewElement(0x1234)
	testTxHashLo      = field.NewElement(0x5678)
	testStateRootHash = field.NewElement(0x1234567890abcdef)
	colSize           = 16          // power of 2
	filterFetchedSize = colSize / 2 // FilterFetched can have different size
)

// createMockInputs creates mock columns for testing
func createMockInputs(comp *wizard.CompiledIOP, size int) *mockInvalidityPIInputs {
	// add StateRootHash to the publicInput of comp
	comp.InsertPublicInput("StateRootHash", accessors.NewConstant(testStateRootHash))

	return &mockInvalidityPIInputs{
		BadPrecompileCol:     comp.InsertCommit(0, "hub.PROVER_ILLEGAL_TRANSACTION_DETECTED", size),
		AddressHi:            comp.InsertCommit(0, "MOCK_ADDRESS_HI", size),
		AddressLo:            comp.InsertCommit(0, "MOCK_ADDRESS_LO", size),
		IsAddressFromTxnData: comp.InsertCommit(0, "MOCK_IS_ADDRESS_FROM_TXNDATA", size),
		TxHashHi:             comp.InsertCommit(0, "MOCK_TX_HASH_HI", size),
		TxHashLo:             comp.InsertCommit(0, "MOCK_TX_HASH_LO", size),
		IsTxHash:             comp.InsertCommit(0, "MOCK_IS_TX_HASH", size),
		FilterFetched:        comp.InsertCommit(0, "MOCK_FILTER_FETCHED", filterFetchedSize),
	}
}

// testCase defines a test case for InvalidityPI
type testCase struct {
	name                  string
	hasBadPrecompile      bool
	numL2Logs             int
	expectedBadPrecompile field.Element
}

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

// TestInvalidityPIAssign tests  that public inputs are accessible and return correct values
func TestInvalidityPIAssign(t *testing.T) {

	testCases := []testCase{
		{
			name:                  "no_bad_precompile_no_logs",
			hasBadPrecompile:      false,
			numL2Logs:             0,
			expectedBadPrecompile: field.Zero(),
		},
		{
			name:                  "has_bad_precompile_no_logs",
			hasBadPrecompile:      true,
			numL2Logs:             0,
			expectedBadPrecompile: field.One(),
		},
		{
			name:                  "no_bad_precompile_with_logs",
			hasBadPrecompile:      false,
			numL2Logs:             3,
			expectedBadPrecompile: field.Zero(),
		},
		{
			name:                  "has_bad_precompile_with_logs",
			hasBadPrecompile:      true,
			numL2Logs:             5,
			expectedBadPrecompile: field.One(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			var (
				mockInputs *mockInvalidityPIInputs
				pi         *InvalidityPI
			)

			define := func(b *wizard.Builder) {
				comp := b.CompiledIOP

				// Create mock input columns
				mockInputs = createMockInputs(comp, colSize)

				pi = NewInvalidityPIZkEvm(comp,
					&logs.ExtractedData{FilterFetched: mockInputs.FilterFetched},
					&ecdsa.EcdsaZkEvm{
						Ant: &ecdsa.Antichamber{
							Size: colSize, // Must set Size for column creation
							TxSignature: &ecdsa.TxSignature{
								IsTxHash: mockInputs.IsTxHash,
								TxHashHi: mockInputs.TxHashHi,
								TxHashLo: mockInputs.TxHashLo,
							},
							Addresses: &ecdsa.Addresses{
								AddressHi:            mockInputs.AddressHi,
								AddressLo:            mockInputs.AddressLo,
								IsAddressFromTxnData: mockInputs.IsAddressFromTxnData,
							},
						},
					},
				)
			}

			prove := func(run *wizard.ProverRuntime) {
				// First assign the mock input columns
				assignMockInputs(run, mockInputs, tc)

				// Run the InvalidityPI Assign
				pi.Assign(run)

				checkPublicInputsAreAccessible(run, pi, tc)

			}

			comp := wizard.Compile(define, dummy.Compile)
			proof := wizard.Prove(comp, prove)
			err := wizard.Verify(comp, proof)

			if err != nil {
				t.Fatalf("verification failed: %v", err)
			}
		})
	}
}

func assignMockInputs(run *wizard.ProverRuntime, mockInputs *mockInvalidityPIInputs, tc testCase) {
	// Assign mock badPrecompile column
	badPrecompileVec := make([]field.Element, colSize)
	if tc.hasBadPrecompile {
		badPrecompileVec[2] = field.One() // Set a non-zero value at index 2
	}
	run.AssignColumn(mockInputs.BadPrecompileCol.GetColID(), smartvectors.NewRegular(badPrecompileVec))

	// Assign mock address columns
	addressHiVec := make([]field.Element, colSize)
	addressLoVec := make([]field.Element, colSize)
	isAddressFromTxnDataVec := make([]field.Element, colSize)
	addressHiVec[3] = testAddressHi
	addressLoVec[3] = testAddressLo
	isAddressFromTxnDataVec[3] = field.One() // Mark row 3 as having the address
	run.AssignColumn(mockInputs.AddressHi.GetColID(), smartvectors.NewRegular(addressHiVec))
	run.AssignColumn(mockInputs.AddressLo.GetColID(), smartvectors.NewRegular(addressLoVec))
	run.AssignColumn(mockInputs.IsAddressFromTxnData.GetColID(), smartvectors.NewRegular(isAddressFromTxnDataVec))

	// Assign mock TxSignature columns
	txHashHiVec := make([]field.Element, colSize)
	txHashLoVec := make([]field.Element, colSize)
	isTxHashVec := make([]field.Element, colSize)
	txHashHiVec[5] = testTxHashHi
	txHashLoVec[5] = testTxHashLo
	isTxHashVec[5] = field.One() // Mark row 5 as having the tx hash
	run.AssignColumn(mockInputs.TxHashHi.GetColID(), smartvectors.NewRegular(txHashHiVec))
	run.AssignColumn(mockInputs.TxHashLo.GetColID(), smartvectors.NewRegular(txHashLoVec))
	run.AssignColumn(mockInputs.IsTxHash.GetColID(), smartvectors.NewRegular(isTxHashVec))

	// Assign mock FilterFetched column (number of L2L1 logs) - uses different size
	filterFetchedVec := make([]field.Element, filterFetchedSize)
	for i := 0; i < tc.numL2Logs && i < filterFetchedSize; i++ {
		filterFetchedVec[i] = field.One()
	}
	run.AssignColumn(mockInputs.FilterFetched.GetColID(), smartvectors.NewRegular(filterFetchedVec))
}

// checkPublicInputsAreAccessible verifies that all InvalidityPI public inputs
// are registered in the CompiledIOP and accessible via GetPublicInput
func checkPublicInputsAreAccessible(run *wizard.ProverRuntime, pi *InvalidityPI, tc testCase) {

	// Verify public inputs are accessible and return correct values
	// StateRootHash
	stateRootHash := run.GetPublicInput(StateRootHash)
	if !stateRootHash.Equal(&testStateRootHash) {
		panic("StateRootHash value mismatch")
	}

	// TxHashHi
	txHashHi := run.GetPublicInput(TxHashHi)
	if !txHashHi.Equal(&testTxHashHi) {
		panic("TxHashHi value mismatch")
	}

	// TxHashLo
	txHashLo := run.GetPublicInput(TxHashLo)
	if !txHashLo.Equal(&testTxHashLo) {
		panic("TxHashLo value mismatch")
	}

	// FromAddress
	expectedAddress := computeExpectedAddress()
	fromAddress := run.GetPublicInput(FromAddress)
	if !fromAddress.Equal(&expectedAddress) {
		panic("FromAddress value mismatch")
	}

	// HashBadPrecompile
	hashBadPrecompile := run.GetPublicInput(HasBadPrecompile)
	if !hashBadPrecompile.Equal(&tc.expectedBadPrecompile) {
		panic("HashBadPrecompile value mismatch")
	}

	// NbL2Logs
	expectedNbL2Logs := field.NewElement(uint64(tc.numL2Logs))
	nbL2Logs := run.GetPublicInput(NbL2Logs)
	if !nbL2Logs.Equal(&expectedNbL2Logs) {
		panic("NbL2Logs value mismatch")
	}
}
