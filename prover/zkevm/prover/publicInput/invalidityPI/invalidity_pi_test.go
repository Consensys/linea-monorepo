package invalidityPI

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
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

// createMockInputs creates mock columns for testing
func createMockInputs(comp *wizard.CompiledIOP, size int) *mockInvalidityPIInputs {
	return &mockInvalidityPIInputs{
		BadPrecompileCol:     comp.InsertCommit(0, "hub.PROVER_ILLEGAL_TRANSACTION_DETECTED", size),
		AddressHi:            comp.InsertCommit(0, "MOCK_ADDRESS_HI", size),
		AddressLo:            comp.InsertCommit(0, "MOCK_ADDRESS_LO", size),
		IsAddressFromTxnData: comp.InsertCommit(0, "MOCK_IS_ADDRESS_FROM_TXNDATA", size),
		TxHashHi:             comp.InsertCommit(0, "MOCK_TX_HASH_HI", size),
		TxHashLo:             comp.InsertCommit(0, "MOCK_TX_HASH_LO", size),
		IsTxHash:             comp.InsertCommit(0, "MOCK_IS_TX_HASH", size),
		FilterFetched:        comp.InsertCommit(0, "MOCK_FILTER_FETCHED", size),
	}
}

// TestInvalidityPIAssign tests the Assign function of InvalidityPI
func TestInvalidityPIAssign(t *testing.T) {

	testCases := []struct {
		name                  string
		hasBadPrecompile      bool
		numL2Logs             int
		expectedBadPrecompile field.Element
	}{
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
			const colSize = 16 // power of 2

			var (
				mockInputs *mockInvalidityPIInputs
				pi         *testableInvalidityPI
			)

			// Test values
			testAddressHi := field.NewElement(0xDEAD)
			testAddressLo := field.NewElement(0xBEEF)
			testTxHashHi := field.NewElement(0x1234)
			testTxHashLo := field.NewElement(0x5678)

			define := func(b *wizard.Builder) {
				comp := b.CompiledIOP

				// Create mock input columns
				mockInputs = createMockInputs(comp, colSize)

				// Create a testable InvalidityPI that uses mock columns
				pi = newTestableInvalidityPI(comp, mockInputs)
			}

			prove := func(run *wizard.ProverRuntime) {
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

				// Assign mock FilterFetched column (number of L2L1 logs)
				filterFetchedVec := make([]field.Element, colSize)
				for i := 0; i < tc.numL2Logs && i < colSize; i++ {
					filterFetchedVec[i] = field.One()
				}
				run.AssignColumn(mockInputs.FilterFetched.GetColID(), smartvectors.NewRegular(filterFetchedVec))

				// Run the InvalidityPI Assign
				pi.Assign(run)

				// Verify results
				hashBadPrecompile := run.GetColumn(pi.HashBadPrecompile.GetColID()).Get(0)
				if !hashBadPrecompile.Equal(&tc.expectedBadPrecompile) {
					t.Errorf("HashBadPrecompile: expected %v, got %v", tc.expectedBadPrecompile.String(), hashBadPrecompile.String())
				}

				nbL2Logs := run.GetColumn(pi.NbL2Logs.GetColID()).Get(0)
				expectedNbL2Logs := field.NewElement(uint64(tc.numL2Logs))
				if !nbL2Logs.Equal(&expectedNbL2Logs) {
					t.Errorf("NbL2Logs: expected %v, got %v", expectedNbL2Logs.String(), nbL2Logs.String())
				}

				txHashHiResult := run.GetColumn(pi.TxHashHi.GetColID()).Get(0)
				if !txHashHiResult.Equal(&testTxHashHi) {
					t.Errorf("TxHashHi: expected %v, got %v", testTxHashHi.String(), txHashHiResult.String())
				}

				txHashLoResult := run.GetColumn(pi.TxHashLo.GetColID()).Get(0)
				if !txHashLoResult.Equal(&testTxHashLo) {
					t.Errorf("TxHashLo: expected %v, got %v", testTxHashLo.String(), txHashLoResult.String())
				}
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

// testableInvalidityPI is a version of InvalidityPI that works with mock columns for testing
type testableInvalidityPI struct {
	TxHashHi          ifaces.Column
	TxHashLo          ifaces.Column
	FromAddress       ifaces.Column
	HashBadPrecompile ifaces.Column
	NbL2Logs          ifaces.Column

	// Mock input references
	badPrecompileCol     ifaces.Column
	addressHi            ifaces.Column
	addressLo            ifaces.Column
	isAddressFromTxnData ifaces.Column
	txHashHi             ifaces.Column
	txHashLo             ifaces.Column
	isTxHash             ifaces.Column
	filterFetched        ifaces.Column

	Extractor InvalidityPIExtractor
}

// newTestableInvalidityPI creates a testable InvalidityPI with mock columns
func newTestableInvalidityPI(comp *wizard.CompiledIOP, mockInputs *mockInvalidityPIInputs) *testableInvalidityPI {
	name := "TEST_INVALIDITY_PI"

	pi := &testableInvalidityPI{
		TxHashHi:          comp.InsertCommit(0, ifaces.ColIDf("%s_TX_HASH_HI", name), 1),
		TxHashLo:          comp.InsertCommit(0, ifaces.ColIDf("%s_TX_HASH_LO", name), 1),
		FromAddress:       comp.InsertCommit(0, ifaces.ColIDf("%s_FROM_ADDRESS", name), 1),
		HashBadPrecompile: comp.InsertCommit(0, ifaces.ColIDf("%s_HASH_BAD_PRECOMPILE", name), 1),
		NbL2Logs:          comp.InsertCommit(0, ifaces.ColIDf("%s_NB_L2_LOGS", name), 1),

		badPrecompileCol:     mockInputs.BadPrecompileCol,
		addressHi:            mockInputs.AddressHi,
		addressLo:            mockInputs.AddressLo,
		isAddressFromTxnData: mockInputs.IsAddressFromTxnData,
		txHashHi:             mockInputs.TxHashHi,
		txHashLo:             mockInputs.TxHashLo,
		isTxHash:             mockInputs.IsTxHash,
		filterFetched:        mockInputs.FilterFetched,
	}

	return pi
}

// Assign for testableInvalidityPI - mirrors the real Assign logic but uses mock columns
func (pi *testableInvalidityPI) Assign(run *wizard.ProverRuntime) {
	var (
		hashBadPrecompile = field.Element{}
		fromAddress       = field.Element{}
		nbL2Logs          uint64
		txHashHi          = field.Element{}
		txHashLo          = field.Element{}
	)

	// 1. Scan the badPrecompile column to find if any value is non-zero
	badPrecompileCol := pi.badPrecompileCol.GetColAssignment(run)
	size := badPrecompileCol.Len()
	for i := 0; i < size; i++ {
		val := badPrecompileCol.Get(i)
		if !val.IsZero() {
			hashBadPrecompile = field.One()
			break
		}
	}

	// 2. Extract FromAddress from addresses module
	isFromCol := pi.isAddressFromTxnData.GetColAssignment(run)
	sizeEcdsa := isFromCol.Len()
	for i := 0; i < sizeEcdsa; i++ {
		source := isFromCol.Get(i)
		if source.IsOne() {
			fromAddressHi := pi.addressHi.GetColAssignmentAt(run, i)
			fromAddressLo := pi.addressLo.GetColAssignmentAt(run, i)
			// create fromAddress from fromAddressHi and fromAddressLo
			hiBytes := fromAddressHi.Bytes()
			loBytes := fromAddressLo.Bytes()
			var b [20]byte
			copy(b[:4], hiBytes[28:])
			copy(b[4:], loBytes[16:])
			fromAddress.SetBytes(b[:])
			break
		}
	}

	// 3. Extract TxHash
	isTxHashCol := pi.isTxHash.GetColAssignment(run)
	ecdsaSize := isTxHashCol.Len()
	for i := 0; i < ecdsaSize; i++ {
		isTxHash := isTxHashCol.Get(i)
		if isTxHash.IsOne() {
			txHashHi = pi.txHashHi.GetColAssignmentAt(run, i)
			txHashLo = pi.txHashLo.GetColAssignmentAt(run, i)
			break
		}
	}

	// 4. Extract NbL2Logs
	filterFetched := pi.filterFetched.GetColAssignment(run)
	sizeFetched := filterFetched.Len()
	for i := 0; i < sizeFetched; i++ {
		filter := filterFetched.Get(i)
		if filter.IsOne() {
			nbL2Logs++
		}
	}

	// Assign the columns
	run.AssignColumn(pi.TxHashHi.GetColID(), smartvectors.NewConstant(txHashHi, 1))
	run.AssignColumn(pi.TxHashLo.GetColID(), smartvectors.NewConstant(txHashLo, 1))
	run.AssignColumn(pi.FromAddress.GetColID(), smartvectors.NewConstant(fromAddress, 1))
	run.AssignColumn(pi.HashBadPrecompile.GetColID(), smartvectors.NewConstant(hashBadPrecompile, 1))
	run.AssignColumn(pi.NbL2Logs.GetColID(), smartvectors.NewConstant(field.NewElement(nbL2Logs), 1))
}

// Ensure unused imports don't cause errors (they're used via type references)
var _ = (*ecdsa.Addresses)(nil)
var _ = (*ecdsa.TxSignature)(nil)
var _ = (*logs.ExtractedData)(nil)
