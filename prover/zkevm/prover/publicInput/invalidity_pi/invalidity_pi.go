package invalidity

import (
	"fmt"
	"math/big"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/accessors"
	"github.com/consensys/linea-monorepo/prover/protocol/dedicated"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	sym "github.com/consensys/linea-monorepo/prover/symbolic"
	commonconstraints "github.com/consensys/linea-monorepo/prover/zkevm/prover/common/common_constraints"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/ecdsa"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/publicInput/logs"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/statemanager/statesummary"
	"github.com/ethereum/go-ethereum/common"
)

// Public input names for invalidity proofs - these must match what BadPrecompileCircuit expects
const (
	StateRootHash    = "StateRootHash"
	TxHashHi         = "TxHash_Hi"
	TxHashLo         = "TxHash_Lo"
	FromAddress      = "FromAddress"
	HasBadPrecompile = "HasBadPrecompile"
	NbL2Logs         = "NbL2Logs"
)

// InvalidityPI is the module responsible for extracting public inputs
// needed for invalidity proofs (BadPrecompile and TooManyLogs circuits).
type InvalidityPI struct {
	InputColumns InputColumns
	Aux          AuxiliaryColumns

	// Output columns - these become wizard public inputs
	TxHashHi         ifaces.Column
	TxHashLo         ifaces.Column
	FromAddress      ifaces.Column
	HasBadPrecompile ifaces.Column
	NbL2Logs         ifaces.Column
	// Extractor holds the LocalOpening queries for the public inputs
	// used to register public inputs via LocalOpening queries
	Extractor InvalidityPIExtractor

	// PublicInputFetcher fetches the public inputs from the logs and root hash fetcher, used to facilitate the assignment of the public inputs
	PublicInputFetcher *PublicInputFetcher
}

// InputColumns collects the input columns from the arithmetization, ECDSA, and logs module
type InputColumns struct {
	// Input columns from arithmetization indicate if a bad precompile was detected
	badPrecompileCol ifaces.Column

	// Input columns from ECDSA
	addresses   *ecdsa.Addresses   // used to bubble up FromAddress to the public input
	txSignature *ecdsa.TxSignature // used to bubble up TxHash to the public input

	// Input columns from logs and root hash fetcher
	FilterFetchedL2L1    ifaces.Column // Input columns from logs - used to bubble up NbL2Logs to the public input
	RootHashFetcherFirst ifaces.Column // Input columns from root hash fetcher - used to bubble up StateRootHash to the public input
}

// AuxiliaryColumns collects the intermediate columns used to constrain the public inputs
type AuxiliaryColumns struct {
	accBadPrecompile ifaces.Column // backward accumulator of badPrecompileCol
	accIsZero        ifaces.Column
	pa               wizard.ProverAction
	FromHi           ifaces.Column
	FromLo           ifaces.Column
}

// InvalidityPIExtractor holds LocalOpening queries for invalidity public inputs
type InvalidityPIExtractor struct {
	StateRootHash     query.LocalOpening
	TxHashHi          query.LocalOpening
	TxHashLo          query.LocalOpening
	FromAddress       query.LocalOpening
	HashBadPrecompile query.LocalOpening
	NbL2Logs          query.LocalOpening
}

func NewInvalidityPI(comp *wizard.CompiledIOP, ecdsa *ecdsa.EcdsaZkEvm, ss *statesummary.Module, logCols logs.LogColumns) *InvalidityPI {
	fetcher := NewPublicInputFetcher(comp, ss, logCols)
	pi := newInvalidityPIFromFetcher(comp,
		ecdsa,
		fetcher.FetchedL2L1.FilterFetched,
		fetcher.RootHashFetcher.First)
	pi.PublicInputFetcher = &fetcher
	return pi
}

func (pi *InvalidityPI) Assign(run *wizard.ProverRuntime, l2BridgeAddress common.Address) {
	pi.PublicInputFetcher.Assign(run, l2BridgeAddress)
	pi.assignFromFetcher(run)
}

// NewInvalidityPIZkEvm creates a new InvalidityPI module
func newInvalidityPIFromFetcher(comp *wizard.CompiledIOP, ecdsa *ecdsa.EcdsaZkEvm, filteredFetchedL2L1 ifaces.Column, rootHashFetcherFirst ifaces.Column) *InvalidityPI {

	var (
		name              = "INVALIDITY_PI"
		ecdsaSize         = ecdsa.Ant.Size
		badPrecompileCol  = comp.Columns.GetHandle("hub.PROVER_ILLEGAL_TRANSACTION_DETECTED")
		badPrecompileSize = badPrecompileCol.Size()
		filterFetchedSize = filteredFetchedL2L1.Size()
	)

	// Create output columns
	txHashHi := comp.InsertCommit(0, ifaces.ColIDf("%s_TX_HASH_HI", name), ecdsaSize)
	txHashLo := comp.InsertCommit(0, ifaces.ColIDf("%s_TX_HASH_LO", name), ecdsaSize)
	fromAddress := comp.InsertCommit(0, ifaces.ColIDf("%s_FROM_ADDRESS", name), ecdsaSize)
	hashBadPrecompile := comp.InsertCommit(0, ifaces.ColIDf("%s_HAS_BAD_PRECOMPILE", name), badPrecompileSize)
	nbL2Logs := comp.InsertCommit(0, ifaces.ColIDf("%s_NB_L2_LOGS", name), filterFetchedSize)

	pi := &InvalidityPI{
		// Output columns
		TxHashHi:         txHashHi,
		TxHashLo:         txHashLo,
		FromAddress:      fromAddress,
		HasBadPrecompile: hashBadPrecompile,
		NbL2Logs:         nbL2Logs,

		// Input columns
		InputColumns: InputColumns{
			badPrecompileCol:     badPrecompileCol,
			txSignature:          ecdsa.Ant.TxSignature,
			addresses:            ecdsa.Ant.Addresses,
			FilterFetchedL2L1:    filteredFetchedL2L1,
			RootHashFetcherFirst: rootHashFetcherFirst,
		},
	}

	// Create auxiliary columns
	pi.Aux = createAuxiliaryColumns(comp,
		pi.InputColumns.badPrecompileCol,
		filteredFetchedL2L1, ecdsaSize,
	)

	// Define constraints over the columns of pi.
	pi.defineConstraints(comp)

	// Create extractor and register public inputs
	pi.generateExtractor(comp, name)

	return pi
}

// defineConstraints defines constraints over the columns of pi.
func (pi *InvalidityPI) defineConstraints(comp *wizard.CompiledIOP) {

	// TxHashHi and TxHashLo must be constant
	commonconstraints.MustBeConstant(comp, pi.TxHashHi)
	commonconstraints.MustBeConstant(comp, pi.TxHashLo)

	// FromHi and FromLo must be constant
	commonconstraints.MustBeConstant(comp, pi.Aux.FromHi)
	commonconstraints.MustBeConstant(comp, pi.Aux.FromLo)

	// accBadPrecompile must be an accumulator backward of the badPrecompileCol column
	commonconstraints.MustBeAccumulatorBackward(comp, pi.Aux.accBadPrecompile, pi.InputColumns.badPrecompileCol)

	// HasBadPrecompile[0] must be 1 iff accBadPrecompile is non-zero at the first row
	commonconstraints.MustBeBinary(comp, pi.HasBadPrecompile)

	pi.Aux.accIsZero, pi.Aux.pa = dedicated.IsZero(comp, pi.Aux.accBadPrecompile).GetColumnAndProverAction()

	// if isZero[0] = 1 then HasBadPrecompile[0] = 0, otherwise HasBadPrecompile[0] = 1
	comp.InsertLocal(0, ifaces.QueryIDf("%s_HAS_BAD_PRECOMPILE_FIRST_ROW", "INVALIDITY_PI"),
		sym.Sub(pi.HasBadPrecompile,
			sym.Sub(1, pi.Aux.accIsZero),
		),
	)

	// NbL2Logs is the backward accumulator of the filterFetched column
	commonconstraints.MustBeAccumulatorBackward(comp, pi.NbL2Logs, pi.InputColumns.FilterFetchedL2L1)

	// when IsAddressFromTxnData = 1, FromHi and FromLo are equal to AddressHi and AddressLo columns
	comp.InsertGlobal(0, ifaces.QueryIDf("%s_FROM_HI_EQUALS_ADDRESS_HI_WHEN_IS_ADDRESS_FROM_TXNDATA", "INVALIDITY_PI"),
		sym.Mul(pi.InputColumns.addresses.IsAddressFromTxnData,
			sym.Sub(pi.Aux.FromHi, pi.InputColumns.addresses.AddressHi),
		))

	comp.InsertGlobal(0, ifaces.QueryIDf("%s_FROM_LO_EQUALS_ADDRESS_LO_WHEN_IS_ADDRESS_FROM_TXNDATA", "INVALIDITY_PI"),
		sym.Mul(pi.InputColumns.addresses.IsAddressFromTxnData,
			sym.Sub(pi.Aux.FromLo, pi.InputColumns.addresses.AddressLo),
		))

	// when IsTxHash = 1, TxHashHi must equal txSignature.TxHashHi
	comp.InsertGlobal(0, ifaces.QueryIDf("%s_TX_HASH_HI_EQUALS_TX_HASH_HI_WHEN_IS_TX_HASH", "INVALIDITY_PI"),
		sym.Mul(pi.InputColumns.txSignature.IsTxHash,
			sym.Sub(pi.TxHashHi, pi.InputColumns.txSignature.TxHashHi),
		))

	// when IsTxHash = 1, TxHashLo must equal txSignature.TxHashLo
	comp.InsertGlobal(0, ifaces.QueryIDf("%s_TX_HASH_LO_EQUALS_TX_HASH_LO_WHEN_IS_TX_HASH", "INVALIDITY_PI"),
		sym.Mul(pi.InputColumns.txSignature.IsTxHash,
			sym.Sub(pi.TxHashLo, pi.InputColumns.txSignature.TxHashLo),
		))

	// fromAddress = FromLo + FromHi * 2^(16*8) = FromLo + FromHi * 2^128
	pow2_128 := new(big.Int).Lsh(big.NewInt(1), 128) // 2^128
	fromAddressExpr := sym.Add(pi.Aux.FromLo, sym.Mul(sym.NewConstant(pow2_128), pi.Aux.FromHi))
	comp.InsertGlobal(0, ifaces.QueryIDf("%s_FROM_ADDRESS_COMPOSITION", "INVALIDITY_PI"),
		sym.Sub(pi.FromAddress, fromAddressExpr))
}

// generateExtractor creates LocalOpening queries and registers public inputs
func (pi *InvalidityPI) generateExtractor(comp *wizard.CompiledIOP, name string) {
	createNewLocalOpening := func(col ifaces.Column) query.LocalOpening {
		return comp.InsertLocalOpening(0, ifaces.QueryIDf("%s_LOCAL_OPENING_%s", name, col.GetColID()), col)
	}

	pi.Extractor = InvalidityPIExtractor{
		TxHashHi:          createNewLocalOpening(pi.TxHashHi),
		TxHashLo:          createNewLocalOpening(pi.TxHashLo),
		FromAddress:       createNewLocalOpening(pi.FromAddress),
		HashBadPrecompile: createNewLocalOpening(pi.HasBadPrecompile),
		NbL2Logs:          createNewLocalOpening(pi.NbL2Logs),
		StateRootHash:     createNewLocalOpening(pi.InputColumns.RootHashFetcherFirst),
	}

	// Register as wizard public inputs with names matching BadPrecompileCircuit expectations
	comp.PublicInputs = append(comp.PublicInputs,
		wizard.PublicInput{Name: StateRootHash, Acc: accessors.NewLocalOpeningAccessor(pi.Extractor.StateRootHash, 0)},
		wizard.PublicInput{Name: TxHashHi, Acc: accessors.NewLocalOpeningAccessor(pi.Extractor.TxHashHi, 0)},
		wizard.PublicInput{Name: TxHashLo, Acc: accessors.NewLocalOpeningAccessor(pi.Extractor.TxHashLo, 0)},
		wizard.PublicInput{Name: FromAddress, Acc: accessors.NewLocalOpeningAccessor(pi.Extractor.FromAddress, 0)},
		wizard.PublicInput{Name: HasBadPrecompile, Acc: accessors.NewLocalOpeningAccessor(pi.Extractor.HashBadPrecompile, 0)},
		wizard.PublicInput{Name: NbL2Logs, Acc: accessors.NewLocalOpeningAccessor(pi.Extractor.NbL2Logs, 0)},
	)
}

// Assign assigns values to the InvalidityPI columns
func (pi *InvalidityPI) assignFromFetcher(run *wizard.ProverRuntime) {
	var (
		hashBadPrecompile = field.Element{}
		fromAddress       = field.Element{}
		txHashHi          = field.Element{}
		txHashLo          = field.Element{}
	)

	// 1. Scan the badPrecompile column to find if any value is non-zero
	badPrecompileCol := pi.InputColumns.badPrecompileCol.GetColAssignment(run)
	size := badPrecompileCol.Len()
	for i := 0; i < size; i++ {
		val := badPrecompileCol.Get(i)
		if !val.IsZero() {
			hashBadPrecompile = field.One()
			break
		}
	}

	// 2. Extract FromAddress from addresses module
	// Sanity check: there must be exactly one row with IsAddressFromTxnData = 1
	isFromCol := pi.InputColumns.addresses.IsAddressFromTxnData.GetColAssignment(run)
	sizeEcdsa := isFromCol.Len()
	isFromCount := 0
	for i := 0; i < sizeEcdsa; i++ {
		source := isFromCol.Get(i)
		if source.IsOne() {
			isFromCount++
			fromAddressHi := pi.InputColumns.addresses.AddressHi.GetColAssignmentAt(run, i)
			fromAddressLo := pi.InputColumns.addresses.AddressLo.GetColAssignmentAt(run, i)
			// create fromAddress from fromAddressHi and fromAddressLo (four bytes from fromAddressHi and 16 bytes from fromAddressLo)
			hiBytes := fromAddressHi.Bytes()
			loBytes := fromAddressLo.Bytes()
			var b [20]byte
			copy(b[:4], hiBytes[28:])
			copy(b[4:], loBytes[16:])
			fromAddress.SetBytes(b[:])
		}
	}
	if isFromCount != 1 {
		panic(fmt.Sprintf("InvalidityPI.Assign: expected exactly one row with IsAddressFromTxnData = 1, got %d", isFromCount))
	}

	// 3. Extract TxHash from ECDSA module - find the row where IsTxHash = 1
	// Sanity check: there must be exactly one row with IsTxHash = 1
	// Note: For invalidity proofs, there should be only a single tx in the trace
	isTxHashCol := pi.InputColumns.txSignature.IsTxHash.GetColAssignment(run)
	ecdsaSize := isTxHashCol.Len()
	isTxHashCount := 0
	for i := 0; i < ecdsaSize; i++ {
		isTxHash := isTxHashCol.Get(i)
		if isTxHash.IsOne() {
			isTxHashCount++
			txHashHi = pi.InputColumns.txSignature.TxHashHi.GetColAssignmentAt(run, i)
			txHashLo = pi.InputColumns.txSignature.TxHashLo.GetColAssignmentAt(run, i)
		}
	}
	if isTxHashCount != 1 {
		panic(fmt.Sprintf("InvalidityPI.Assign: expected exactly one row with IsTxHash = 1, got %d", isTxHashCount))
	}

	// 4. Assign auxiliary columns
	pi.assignAuxiliaryColumns(run, badPrecompileCol, isFromCol)

	// Assign the columns with their correct sizes
	run.AssignColumn(pi.TxHashHi.GetColID(), smartvectors.NewConstant(txHashHi, pi.TxHashHi.Size()))
	run.AssignColumn(pi.TxHashLo.GetColID(), smartvectors.NewConstant(txHashLo, pi.TxHashLo.Size()))
	run.AssignColumn(pi.FromAddress.GetColID(), smartvectors.NewConstant(fromAddress, pi.FromAddress.Size()))
	run.AssignColumn(pi.HasBadPrecompile.GetColID(), smartvectors.NewConstant(hashBadPrecompile, pi.HasBadPrecompile.Size()))

	// Assign accNbL2Logs (backward accumulator of filterFetched)
	filterFetched := pi.InputColumns.FilterFetchedL2L1.GetColAssignment(run)
	sizeFetched := filterFetched.Len()
	accNbL2LogsValues := make([]field.Element, sizeFetched)
	accNbL2LogsValues[sizeFetched-1] = filterFetched.Get(sizeFetched - 1)
	for i := sizeFetched - 2; i >= 0; i-- {
		val := filterFetched.Get(i)
		accNbL2LogsValues[i].Add(&val, &accNbL2LogsValues[i+1])
	}
	run.AssignColumn(pi.NbL2Logs.GetColID(), smartvectors.NewRegular(accNbL2LogsValues))

	// Assign local openings
	run.AssignLocalPoint(pi.Extractor.StateRootHash.ID, pi.InputColumns.RootHashFetcherFirst.GetColAssignmentAt(run, 0))
	run.AssignLocalPoint(pi.Extractor.TxHashHi.ID, txHashHi)
	run.AssignLocalPoint(pi.Extractor.TxHashLo.ID, txHashLo)
	run.AssignLocalPoint(pi.Extractor.FromAddress.ID, fromAddress)
	run.AssignLocalPoint(pi.Extractor.HashBadPrecompile.ID, hashBadPrecompile)
	run.AssignLocalPoint(pi.Extractor.NbL2Logs.ID, accNbL2LogsValues[0])
}

// createAuxiliaryColumns creates the auxiliary columns needed for constraints
func createAuxiliaryColumns(comp *wizard.CompiledIOP, badPrecompileCol ifaces.Column, filterFetched ifaces.Column, addressesSize int) AuxiliaryColumns {
	name := "INVALIDITY_PI_AUX"
	size := badPrecompileCol.Size()

	// accBadPrecompile is a backward accumulator of badPrecompileCol
	accBadPrecompile := comp.InsertCommit(0, ifaces.ColIDf("%s_ACC_BAD_PRECOMPILE", name), size)

	// FromHi and FromLo are intermediate columns for address constraints
	fromHi := comp.InsertCommit(0, ifaces.ColIDf("%s_FROM_HI", name), addressesSize)
	fromLo := comp.InsertCommit(0, ifaces.ColIDf("%s_FROM_LO", name), addressesSize)

	return AuxiliaryColumns{
		accBadPrecompile: accBadPrecompile,
		FromHi:           fromHi,
		FromLo:           fromLo,
	}
}

func (pi *InvalidityPI) assignAuxiliaryColumns(run *wizard.ProverRuntime, badPrecompileCol, isFromCol smartvectors.SmartVector) {
	size := badPrecompileCol.Len()
	sizeEcdsa := isFromCol.Len()

	// Assign accBadPrecompile (backward accumulator of badPrecompileCol)
	accBadPrecompileValues := make([]field.Element, size)
	accBadPrecompileValues[size-1] = badPrecompileCol.Get(size - 1)
	for i := size - 2; i >= 0; i-- {
		val := badPrecompileCol.Get(i)
		accBadPrecompileValues[i].Add(&val, &accBadPrecompileValues[i+1])
	}
	run.AssignColumn(pi.Aux.accBadPrecompile.GetColID(), smartvectors.NewRegular(accBadPrecompileValues))

	// Run the IsZero prover action to assign accIsZero
	pi.Aux.pa.Run(run)

	// Assign FromHi and FromLo (constant columns with the address Hi/Lo values)
	var fromHi, fromLo field.Element
	for i := 0; i < sizeEcdsa; i++ {
		source := isFromCol.Get(i)
		if source.IsOne() {
			fromHi = pi.InputColumns.addresses.AddressHi.GetColAssignmentAt(run, i)
			fromLo = pi.InputColumns.addresses.AddressLo.GetColAssignmentAt(run, i)
			break
		}
	}
	run.AssignColumn(pi.Aux.FromHi.GetColID(), smartvectors.NewConstant(fromHi, sizeEcdsa))
	run.AssignColumn(pi.Aux.FromLo.GetColID(), smartvectors.NewConstant(fromLo, sizeEcdsa))
}
