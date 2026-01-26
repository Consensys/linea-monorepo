package invalidityPI

import (
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/accessors"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/ecdsa"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/publicInput/logs"
)

// MockInvalidityPIInputs holds mock columns for testing InvalidityPI
type MockInvalidityPIInputs struct {
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

// FixedInputs are the inputs that are fixed for the test
type FixedInputs struct {
	StateRootHash field.Element
	TxHashHi      field.Element
	TxHashLo      field.Element
	AddressHi     field.Element
	AddressLo     field.Element
	FromAddress   field.Element
	ColSize       int
}

// CaseInputs are the inputs that are different for each test case
type CaseInputs struct {
	HasBadPrecompile bool
	NumL2Logs        int
}
type Inputs struct {
	FixedInputs
	CaseInputs
}

// CreateMockInputs creates mock columns for testing
func CreateMockInputs(comp *wizard.CompiledIOP, size int, testStateRootHash field.Element) *MockInvalidityPIInputs {
	// add StateRootHash to the publicInput of comp
	comp.InsertPublicInput("StateRootHash", accessors.NewConstant(testStateRootHash))

	return &MockInvalidityPIInputs{
		BadPrecompileCol:     comp.InsertCommit(0, "hub.PROVER_ILLEGAL_TRANSACTION_DETECTED", size/2),
		AddressHi:            comp.InsertCommit(0, "MOCK_ADDRESS_HI", size),
		AddressLo:            comp.InsertCommit(0, "MOCK_ADDRESS_LO", size),
		IsAddressFromTxnData: comp.InsertCommit(0, "MOCK_IS_ADDRESS_FROM_TXNDATA", size),
		TxHashHi:             comp.InsertCommit(0, "MOCK_TX_HASH_HI", size),
		TxHashLo:             comp.InsertCommit(0, "MOCK_TX_HASH_LO", size),
		IsTxHash:             comp.InsertCommit(0, "MOCK_IS_TX_HASH", size),
		FilterFetched:        comp.InsertCommit(0, "MOCK_FILTER_FETCHED", size*2),
	}
}

func AssignMockInputs(run *wizard.ProverRuntime, colSize int, mockInputs *MockInvalidityPIInputs, in Inputs) {
	// Assign mock badPrecompile column
	badPrecompileVec := make([]field.Element, colSize/2)
	if in.CaseInputs.HasBadPrecompile {
		badPrecompileVec[2] = field.One() // Set a non-zero value at index 2
		badPrecompileVec[4] = field.One() // edge-case: more than one bad precompile
	}
	run.AssignColumn(mockInputs.BadPrecompileCol.GetColID(), smartvectors.NewRegular(badPrecompileVec))

	// Assign mock address columns
	addressHiVec := make([]field.Element, colSize)
	addressLoVec := make([]field.Element, colSize)
	isAddressFromTxnDataVec := make([]field.Element, colSize)
	addressHiVec[3] = in.AddressHi
	addressLoVec[3] = in.AddressLo
	isAddressFromTxnDataVec[3] = field.One() // Mark row 3 as having the address
	run.AssignColumn(mockInputs.AddressHi.GetColID(), smartvectors.NewRegular(addressHiVec))
	run.AssignColumn(mockInputs.AddressLo.GetColID(), smartvectors.NewRegular(addressLoVec))
	run.AssignColumn(mockInputs.IsAddressFromTxnData.GetColID(), smartvectors.NewRegular(isAddressFromTxnDataVec))

	// Assign mock TxSignature columns
	txHashHiVec := make([]field.Element, colSize)
	txHashLoVec := make([]field.Element, colSize)
	isTxHashVec := make([]field.Element, colSize)
	txHashHiVec[5] = in.TxHashHi
	txHashLoVec[5] = in.TxHashLo
	isTxHashVec[5] = field.One() // Mark row 5 as having the tx hash
	run.AssignColumn(mockInputs.TxHashHi.GetColID(), smartvectors.NewRegular(txHashHiVec))
	run.AssignColumn(mockInputs.TxHashLo.GetColID(), smartvectors.NewRegular(txHashLoVec))
	run.AssignColumn(mockInputs.IsTxHash.GetColID(), smartvectors.NewRegular(isTxHashVec))

	// Assign mock FilterFetched column (number of L2L1 logs) - uses different size
	filterFetchedSize := colSize * 2
	filterFetchedVec := make([]field.Element, filterFetchedSize)
	for i := 0; i < in.NumL2Logs && i < filterFetchedSize; i++ {
		filterFetchedVec[i] = field.One()
	}
	run.AssignColumn(mockInputs.FilterFetched.GetColID(), smartvectors.NewRegular(filterFetchedVec))
}

// MockZkevmArithCols creates mock the arithmetization columns that are used to create the public inputs
func MockZkevmArithCols(in Inputs) (*wizard.CompiledIOP, wizard.Proof) {
	var (
		mockInputs *MockInvalidityPIInputs
		pi         *InvalidityPI
	)

	define := func(b *wizard.Builder) {
		comp := b.CompiledIOP

		// Create mock input columns
		mockInputs = CreateMockInputs(comp, in.ColSize, in.StateRootHash)

		pi = NewInvalidityPIZkEvm(comp,
			&logs.ExtractedData{FilterFetched: mockInputs.FilterFetched},
			&ecdsa.EcdsaZkEvm{
				Ant: &ecdsa.Antichamber{
					Size: in.ColSize,
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
		AssignMockInputs(run, in.ColSize, mockInputs, in)

		// Run the InvalidityPI Assign
		pi.Assign(run)
	}

	comp := wizard.Compile(define, dummy.Compile)
	proof := wizard.Prove(comp, prove)

	return comp, proof
}
