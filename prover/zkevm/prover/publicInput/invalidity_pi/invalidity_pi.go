package invalidity

import (
	"fmt"
	"slices"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/accessors"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/limbs"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	sym "github.com/consensys/linea-monorepo/prover/symbolic"
	zkevmcommon "github.com/consensys/linea-monorepo/prover/zkevm/prover/common"
	commonconstraints "github.com/consensys/linea-monorepo/prover/zkevm/prover/common/common_constraints"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/ecdsa"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/publicInput/logs"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/statemanager/statesummary"
	"github.com/ethereum/go-ethereum/common"
)

// Public input names for invalidity proofs - these must match what BadPrecompileCircuit expects

const (
	// Metadata key for storing the extractor in CompiledIOP
	InvalidityPIExtractorMetadata = "InvalidityPIExtractor"

	StateRootHashName    = "StateRootHash" // 8 KoalaBear elements
	TxHashName           = "TxHash"        // 16 limbs in BE order (MSB first), converted from ECDSA LE
	FromName             = "From"          // 10 limbs in BE order (MSB first), converted from ECDSA LE
	HasBadPrecompileName = "HasBadPrecompile"
	NbL2LogsName         = "NbL2Logs"
)

// InvalidityPI is the module responsible for extracting public inputs
// needed for invalidity proofs (BadPrecompile and TooManyLogs circuits).
type InvalidityPI struct {
	InputColumns InputColumns

	// Output columns - these become wizard public inputs
	TxHash limbs.Uint256Be
	From   limbs.EthAddress

	HasBadPrecompile ifaces.Column // backward accumulator of the badPrecompileCol column
	NbL2Logs         ifaces.Column // backward accumulator of the filterFetched column
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
	FilterFetchedL2L1    ifaces.Column                         // Input columns from logs - used to bubble up NbL2Logs to the public input
	RootHashFetcherFirst [zkevmcommon.NbLimbU128]ifaces.Column // Input columns from root hash fetcher - used to bubble up StateRootHash to the public input
}

// InvalidityPIExtractor holds wizard.PublicInput for invalidity public inputs
type InvalidityPIExtractor struct {
	StateRootHash [zkevmcommon.NbLimbU128]wizard.PublicInput
	TxHash        [zkevmcommon.NbLimbU256]wizard.PublicInput
	FromAddress   [zkevmcommon.NbLimbEthAddress]wizard.PublicInput

	HasBadPrecompile wizard.PublicInput // the first row of the HasBadPrecompile column (a non-zero value means a bad precompile was detected)
	NbL2Logs         wizard.PublicInput // the first row of the backward accumulator of the FilterFetched column
}

// NewInvalidityPI creates a new InvalidityPI module from the arithmetization, ECDSA, and logs module
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
func newInvalidityPIFromFetcher(comp *wizard.CompiledIOP, ecdsa *ecdsa.EcdsaZkEvm, filteredFetchedL2L1 ifaces.Column, rootHashFetcherFirst [zkevmcommon.NbLimbU128]ifaces.Column) *InvalidityPI {

	var (
		name              = "INVALIDITY_PI"
		ecdsaSize         = ecdsa.Ant.Size
		badPrecompileCol  = comp.Columns.GetHandle("hub.PROVER_ILLEGAL_TRANSACTION_DETECTED")
		badPrecompileSize = badPrecompileCol.Size()
		filterFetchedSize = filteredFetchedL2L1.Size()
	)

	// Create output columns using limbs types for clear endianness
	txHash := limbs.NewUint256Be(comp, ifaces.ColID(name+"_TX_HASH"), ecdsaSize)
	from := limbs.NewEthAddress(comp, ifaces.ColID(name+"_FROM"), ecdsaSize)

	hashBadPrecompile := comp.InsertCommit(0, ifaces.ColIDf("%s_HAS_BAD_PRECOMPILE", name), badPrecompileSize, true)
	nbL2Logs := comp.InsertCommit(0, ifaces.ColIDf("%s_NB_L2_LOGS", name), filterFetchedSize, true)

	pi := &InvalidityPI{
		// Output columns
		TxHash:           txHash,
		From:             from,
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

	// Define constraints over the columns of pi.
	pi.defineConstraints(comp)

	// Create extractor and register public inputs
	pi.generateExtractor(comp, name)

	return pi
}

// defineConstraints defines constraints over the columns of pi.
func (pi *InvalidityPI) defineConstraints(comp *wizard.CompiledIOP) {

	// the idea is that we propagate the target value over the rows to bring it to the first row
	// each limb of TxHash must be constant,
	commonconstraints.LimbsMustBeConstant(comp, pi.TxHash.GetLimbs())

	// each limb of From must be constant
	commonconstraints.LimbsMustBeConstant(comp, pi.From.GetLimbs())

	// HasBadPrecompile must be an accumulator backward of the badPrecompileCol column
	commonconstraints.MustBeAccumulatorBackward(comp, pi.HasBadPrecompile, pi.InputColumns.badPrecompileCol)

	// NbL2Logs is the backward accumulator of the filterFetched column
	commonconstraints.MustBeAccumulatorBackward(comp, pi.NbL2Logs, pi.InputColumns.FilterFetchedL2L1)

	// when IsAddressFromTxnData = 1, From equals Address (using the same layout as ecdsa.Addresses())
	limbs.NewGlobal(comp, ifaces.QueryIDf("%s_FROM_EQUALS_ADDRESS", "INVALIDITY_PI"),
		sym.Mul(pi.InputColumns.addresses.IsAddressFromTxnData,
			sym.Sub(pi.From.AsDynSize(), pi.InputColumns.addresses.Addresses().ToBigEndianLimbs()),
		))

	// when IsTxHash = 1, TxHash must equal txSignature.TxHash
	limbs.NewGlobal(comp, ifaces.QueryIDf("%s_TX_HASH_EQUALS_TX_HASH", "INVALIDITY_PI"),
		sym.Mul(pi.InputColumns.txSignature.IsTxHash,
			sym.Sub(pi.TxHash.AsDynSize(), pi.InputColumns.txSignature.TxHash.ToBigEndianLimbs()),
		))
}

// generateExtractor registers public inputs
func (pi *InvalidityPI) generateExtractor(comp *wizard.CompiledIOP, name string) {
	// Helper: creates LocalOpening query + PublicInput for a single column
	newLoPublicInput := func(col ifaces.Column, name string) wizard.PublicInput {
		q := comp.InsertLocalOpening(0, ifaces.QueryIDf("%s_%s", "PUBLIC_INPUT_LOCAL_OPENING", name), col)
		return comp.InsertPublicInput(name, accessors.NewLocalOpeningAccessor(q, 0))
	}

	// Helper: creates LocalOpening queries + PublicInputs for an array of columns
	newLoPublicInputs := func(cols []ifaces.Column, baseName string) []wizard.PublicInput {
		pis := make([]wizard.PublicInput, len(cols))
		for i, col := range cols {
			pis[i] = newLoPublicInput(col, fmt.Sprintf("%s_%d", baseName, i))
		}
		return pis
	}

	// Register StateRootHash public inputs (8 KoalaBear elements)
	copy(pi.Extractor.StateRootHash[:], newLoPublicInputs(pi.InputColumns.RootHashFetcherFirst[:], StateRootHashName))

	// Register TxHash public inputs
	copy(pi.Extractor.TxHash[:], newLoPublicInputs(pi.TxHash.GetLimbs(), TxHashName))

	// Register FromAddress public inputs
	copy(pi.Extractor.FromAddress[:], newLoPublicInputs(pi.From.GetLimbs(), FromName))

	// Register scalar public inputs
	pi.Extractor.HasBadPrecompile = newLoPublicInput(pi.HasBadPrecompile, HasBadPrecompileName)
	pi.Extractor.NbL2Logs = newLoPublicInput(pi.NbL2Logs, NbL2LogsName)

	// Store extractor in CompiledIOP metadata for easy circuit access
	// This follows the same pattern as execution circuit's input_extractor.go:171
	comp.ExtraData[InvalidityPIExtractorMetadata] = &pi.Extractor
}

// assignFromFetcher assigns values to the InvalidityPI columns
func (pi *InvalidityPI) assignFromFetcher(run *wizard.ProverRuntime) {

	// 2. Extract FromAddress from addresses module using Addresses()
	// Find the row where IsAddressFromTxnData = 1 and extract limb values
	isFromCol := pi.InputColumns.addresses.IsAddressFromTxnData.GetColAssignment(run)
	sizeEcdsa := isFromCol.Len()
	var fromLimbValues [zkevmcommon.NbLimbEthAddress]field.Element
	isFromCount := 0
	for i := 0; i < sizeEcdsa; i++ {
		source := isFromCol.Get(i)
		if source.IsOne() {
			isFromCount++
			addressRow := pi.InputColumns.addresses.Addresses().GetRow(run, i)
			for j := 0; j < zkevmcommon.NbLimbEthAddress; j++ {
				fromLimbValues[j] = addressRow.T[j]
			}
			break
		}
	}
	if isFromCount < 1 {
		panic(fmt.Sprintf("InvalidityPI.Assign: expected at least one row with IsAddressFromTxnData = 1, got %d", isFromCount))
	}

	// 3. Extract TxHash from ECDSA module - find the row where IsTxHash = 1
	isTxHashCol := pi.InputColumns.txSignature.IsTxHash.GetColAssignment(run)
	ecdsaSize := isTxHashCol.Len()
	var txHashLimbValues [zkevmcommon.NbLimbU256]field.Element
	isTxHashCount := 0
	for i := 0; i < ecdsaSize; i++ {
		isTxHash := isTxHashCol.Get(i)
		if isTxHash.IsOne() {
			isTxHashCount++
			txHashRow := pi.InputColumns.txSignature.TxHash.GetRow(run, i)
			for j := 0; j < zkevmcommon.NbLimbU256; j++ {
				txHashLimbValues[j] = txHashRow.T[j]
			}
			break
		}
	}
	if isTxHashCount < 1 {
		panic(fmt.Sprintf("InvalidityPI.Assign: expected at least one row with IsTxHash = 1, got %d", isTxHashCount))
	}

	// 5. Assign output columns
	// Assign TxHash limbs (constant columns)
	slices.Reverse(txHashLimbValues[:]) // reverse the txHashLimbValues to be in BE order
	for i, col := range pi.TxHash.GetLimbs() {
		run.AssignColumn(col.GetColID(), smartvectors.NewConstant(txHashLimbValues[i], col.Size()))
	}

	// Assign From limbs (constant columns)
	slices.Reverse(fromLimbValues[:]) // reverse the fromLimbValues to be in BE order
	for i, col := range pi.From.GetLimbs() {
		run.AssignColumn(col.GetColID(), smartvectors.NewConstant(fromLimbValues[i], col.Size()))
	}

	// Assign NbL2Logs (backward accumulator of filterFetched)
	filterFetched := pi.InputColumns.FilterFetchedL2L1.GetColAssignment(run)
	sizeFetched := filterFetched.Len()
	accNbL2LogsValues := make([]field.Element, sizeFetched)
	accNbL2LogsValues[sizeFetched-1] = filterFetched.Get(sizeFetched - 1)
	for i := sizeFetched - 2; i >= 0; i-- {
		val := filterFetched.Get(i)
		accNbL2LogsValues[i].Add(&val, &accNbL2LogsValues[i+1])
	}
	run.AssignColumn(pi.NbL2Logs.GetColID(), smartvectors.NewRegular(accNbL2LogsValues))

	// Assign HasBadPrecompile (backward accumulator of badPrecompileCol)
	badPrecompileCol := pi.InputColumns.badPrecompileCol.GetColAssignment(run)
	size := badPrecompileCol.Len()
	accBadPrecompileValues := make([]field.Element, size)
	accBadPrecompileValues[size-1] = badPrecompileCol.Get(size - 1)
	for i := size - 2; i >= 0; i-- {
		val := badPrecompileCol.Get(i)
		accBadPrecompileValues[i].Add(&val, &accBadPrecompileValues[i+1])
	}
	run.AssignColumn(pi.HasBadPrecompile.GetColID(), smartvectors.NewRegular(accBadPrecompileValues))

	// 6. Assign local openings from the extractor's public inputs
	assignLO := func(pi wizard.PublicInput, value field.Element) {
		q, ok := pi.Acc.(*accessors.FromLocalOpeningYAccessor)
		if !ok {
			panic("pi.Acc is not a FromLocalOpeningYAccessor")
		}
		run.AssignLocalPoint(q.Q.ID, value)
	}

	// StateRootHash limbs
	for i := range pi.Extractor.StateRootHash {
		assignLO(pi.Extractor.StateRootHash[i], pi.InputColumns.RootHashFetcherFirst[i].GetColAssignmentAt(run, 0))
	}
	// TxHash limbs
	for i := range pi.Extractor.TxHash {
		assignLO(pi.Extractor.TxHash[i], txHashLimbValues[i])
	}
	// From limbs
	for i := range pi.Extractor.FromAddress {
		assignLO(pi.Extractor.FromAddress[i], fromLimbValues[i])
	}
	assignLO(pi.Extractor.HasBadPrecompile, accBadPrecompileValues[0])
	assignLO(pi.Extractor.NbL2Logs, accNbL2LogsValues[0])
}
