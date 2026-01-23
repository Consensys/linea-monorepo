package invalidityPI

import (
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
	// StateRootHash is already a public input in the CompiledIOP, we just bubble it up as a public input for invalidity.
	StateRootHash wizard.PublicInput

	// Output columns - these become wizard public inputs
	TxHashHi         ifaces.Column
	TxHashLo         ifaces.Column
	FromAddress      ifaces.Column
	HasBadPrecompile ifaces.Column
	NbL2Logs         ifaces.Column
	// Extractor holds the LocalOpening queries for the public inputs
	Extractor InvalidityPIExtractor // used to register public inputs via LocalOpening queries
}

// InputColumns collects the input columns from the arithmetization, ECDSA, and logs module
type InputColumns struct {
	// Input columns from arithmetization
	badPrecompileCol ifaces.Column

	// Input columns from ECDSA
	addresses   *ecdsa.Addresses   // used to bubble up FromAddress to the public input
	txSignature *ecdsa.TxSignature // used to bubble up TxHash to the public input

	// Input columns from logs module
	fetchedL2L1 *logs.ExtractedData // used to bubble up NbL2Logs to the public input
}

// AuxiliaryColumns collects the intermediate columns used to constrain the public inputs
type AuxiliaryColumns struct {
	accBadPrecompile ifaces.Column // backward accumulator of badPrecompileCol
	accNbL2Logs      ifaces.Column // backward accumulator of filterFetched column
	accIsZero        ifaces.Column
	pa               wizard.ProverAction
	FromHi           ifaces.Column
	FromLo           ifaces.Column
}

// InvalidityPIExtractor holds LocalOpening queries for invalidity public inputs
type InvalidityPIExtractor struct {
	TxHashHi          query.LocalOpening
	TxHashLo          query.LocalOpening
	FromAddress       query.LocalOpening
	HashBadPrecompile query.LocalOpening
	NbL2Logs          query.LocalOpening
}

// NewInvalidityPIZkEvm creates a new InvalidityPI module using columns from the arithmetization
func NewInvalidityPIZkEvm(comp *wizard.CompiledIOP, fetchedL2L1 *logs.ExtractedData, ecdsa *ecdsa.EcdsaZkEvm) *InvalidityPI {
	name := "INVALIDITY_PI"
	size := ecdsa.Ant.Size
	filterFetchedSize := fetchedL2L1.FilterFetched.Size()
	badPrecompileSize := comp.Columns.GetHandle("hub.PROVER_ILLEGAL_TRANSACTION_DETECTED").Size()

	// Create output columns
	txHashHi := comp.InsertCommit(0, ifaces.ColIDf("%s_TX_HASH_HI", name), size)
	txHashLo := comp.InsertCommit(0, ifaces.ColIDf("%s_TX_HASH_LO", name), size)
	fromAddress := comp.InsertCommit(0, ifaces.ColIDf("%s_FROM_ADDRESS", name), size)
	hashBadPrecompile := comp.InsertCommit(0, ifaces.ColIDf("%s_HAS_BAD_PRECOMPILE", name), badPrecompileSize)
	nbL2Logs := comp.InsertCommit(0, ifaces.ColIDf("%s_NB_L2_LOGS", name), filterFetchedSize)

	pi := &InvalidityPI{
		StateRootHash: comp.GetPublicInput(StateRootHash),
		// Output columns
		TxHashHi:         txHashHi,
		TxHashLo:         txHashLo,
		FromAddress:      fromAddress,
		HasBadPrecompile: hashBadPrecompile,
		NbL2Logs:         nbL2Logs,

		// Input columns
		InputColumns: InputColumns{
			badPrecompileCol: comp.Columns.GetHandle("hub.PROVER_ILLEGAL_TRANSACTION_DETECTED"),
			fetchedL2L1:      fetchedL2L1,
			txSignature:      ecdsa.Ant.TxSignature,
			addresses:        ecdsa.Ant.Addresses,
		},
	}

	// Create auxiliary columns
	pi.Aux = createAuxiliaryColumns(comp, pi.InputColumns.badPrecompileCol, pi.InputColumns.fetchedL2L1.FilterFetched, pi.InputColumns.addresses.IsAddressFromTxnData.Size())

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

	// nbL2Logs is the backward accumulator of the filterFetched column
	commonconstraints.MustBeAccumulatorBackward(comp, pi.Aux.accNbL2Logs, pi.InputColumns.fetchedL2L1.FilterFetched)

	// when IsAddressFromTxnData = 1, FromHi and FromLo are equal to AddressHi and AddressLo columns
	comp.InsertGlobal(0, ifaces.QueryIDf("%s_FROM_HI_EQUALS_ADDRESS_HI_WHEN_IS_ADDRESS_FROM_TXNDATA", "INVALIDITY_PI"),
		sym.Mul(pi.InputColumns.addresses.IsAddressFromTxnData,
			sym.Sub(pi.Aux.FromHi, pi.InputColumns.addresses.AddressHi),
		))

	// the same for FromLo
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
	}

	// Register as wizard public inputs with names matching BadPrecompileCircuit expectations
	comp.PublicInputs = append(comp.PublicInputs,
		wizard.PublicInput{Name: TxHashHi, Acc: accessors.NewLocalOpeningAccessor(pi.Extractor.TxHashHi, 0)},
		wizard.PublicInput{Name: TxHashLo, Acc: accessors.NewLocalOpeningAccessor(pi.Extractor.TxHashLo, 0)},
		wizard.PublicInput{Name: FromAddress, Acc: accessors.NewLocalOpeningAccessor(pi.Extractor.FromAddress, 0)},
		wizard.PublicInput{Name: HasBadPrecompile, Acc: accessors.NewLocalOpeningAccessor(pi.Extractor.HashBadPrecompile, 0)},
		wizard.PublicInput{Name: NbL2Logs, Acc: accessors.NewLocalOpeningAccessor(pi.Extractor.NbL2Logs, 0)},
	)
}

// Assign assigns values to the InvalidityPI columns using data from the runtime
func (pi *InvalidityPI) Assign(run *wizard.ProverRuntime) {
	var (
		hashBadPrecompile = field.Element{}
		fromAddress       = field.Element{}
		nbL2Logs          uint64
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
	isFromCol := pi.InputColumns.addresses.IsAddressFromTxnData.GetColAssignment(run)
	sizeEcdsa := isFromCol.Len()
	for i := 0; i < sizeEcdsa; i++ {
		source := isFromCol.Get(i)
		if source.IsOne() {
			fromAddressHi := pi.InputColumns.addresses.AddressHi.GetColAssignmentAt(run, i)
			fromAddressLo := pi.InputColumns.addresses.AddressLo.GetColAssignmentAt(run, i)
			// create fromAddress from fromAddressHi and fromAddressLo (four bytes from fromAddressHi and 16 bytes from fromAddressLo)
			hiBytes := fromAddressHi.Bytes()
			loBytes := fromAddressLo.Bytes()
			var b [20]byte
			copy(b[:4], hiBytes[28:])
			copy(b[4:], loBytes[16:])
			fromAddress.SetBytes(b[:])
			break
		}
	}

	// 3. Extract TxHash from ECDSA module - find the row where IsTxHash = 1
	// Note: For invalidity proofs, there should be only a single tx in the trace
	isTxHashCol := pi.InputColumns.txSignature.IsTxHash.GetColAssignment(run)
	ecdsaSize := isTxHashCol.Len()

	for i := 0; i < ecdsaSize; i++ {
		isTxHash := isTxHashCol.Get(i)
		if isTxHash.IsOne() {
			txHashHi = pi.InputColumns.txSignature.TxHashHi.GetColAssignmentAt(run, i)
			txHashLo = pi.InputColumns.txSignature.TxHashLo.GetColAssignmentAt(run, i)
			break
		}
	}

	// 4. Extract NbL2Logs from logs module
	filterFetched := pi.InputColumns.fetchedL2L1.FilterFetched.GetColAssignment(run)
	sizeFetched := filterFetched.Len()
	for i := 0; i < sizeFetched; i++ {
		filter := filterFetched.Get(i)
		if filter.IsOne() {
			nbL2Logs++
		}
	}

	// 5. Assign auxiliary columns
	pi.assignAuxiliaryColumns(run, badPrecompileCol, filterFetched, isFromCol)

	// Assign the columns with their correct sizes
	run.AssignColumn(pi.TxHashHi.GetColID(), smartvectors.NewConstant(txHashHi, pi.TxHashHi.Size()))
	run.AssignColumn(pi.TxHashLo.GetColID(), smartvectors.NewConstant(txHashLo, pi.TxHashLo.Size()))
	run.AssignColumn(pi.FromAddress.GetColID(), smartvectors.NewConstant(fromAddress, pi.FromAddress.Size()))
	run.AssignColumn(pi.HasBadPrecompile.GetColID(), smartvectors.NewConstant(hashBadPrecompile, pi.HasBadPrecompile.Size()))
	run.AssignColumn(pi.NbL2Logs.GetColID(), smartvectors.NewConstant(field.NewElement(nbL2Logs), pi.NbL2Logs.Size()))

	// Assign local openings
	run.AssignLocalPoint(pi.Extractor.TxHashHi.ID, txHashHi)
	run.AssignLocalPoint(pi.Extractor.TxHashLo.ID, txHashLo)
	run.AssignLocalPoint(pi.Extractor.FromAddress.ID, fromAddress)
	run.AssignLocalPoint(pi.Extractor.HashBadPrecompile.ID, hashBadPrecompile)
	run.AssignLocalPoint(pi.Extractor.NbL2Logs.ID, field.NewElement(nbL2Logs))
}

// createAuxiliaryColumns creates the auxiliary columns needed for constraints
func createAuxiliaryColumns(comp *wizard.CompiledIOP, badPrecompileCol ifaces.Column, filterFetched ifaces.Column, addressesSize int) AuxiliaryColumns {
	name := "INVALIDITY_PI_AUX"
	size := badPrecompileCol.Size()
	filterFetchedSize := filterFetched.Size()

	// accBadPrecompile is a backward accumulator of badPrecompileCol
	accBadPrecompile := comp.InsertCommit(0, ifaces.ColIDf("%s_ACC_BAD_PRECOMPILE", name), size)

	// accNbL2Logs is a backward accumulator of filterFetched column
	accNbL2Logs := comp.InsertCommit(0, ifaces.ColIDf("%s_ACC_NB_L2_LOGS", name), filterFetchedSize)

	// FromHi and FromLo are intermediate columns for address constraints
	fromHi := comp.InsertCommit(0, ifaces.ColIDf("%s_FROM_HI", name), addressesSize)
	fromLo := comp.InsertCommit(0, ifaces.ColIDf("%s_FROM_LO", name), addressesSize)

	return AuxiliaryColumns{
		accBadPrecompile: accBadPrecompile,
		accNbL2Logs:      accNbL2Logs,
		FromHi:           fromHi,
		FromLo:           fromLo,
	}
}

func (pi *InvalidityPI) assignAuxiliaryColumns(run *wizard.ProverRuntime, badPrecompileCol, filterFetched, isFromCol smartvectors.SmartVector) {
	size := badPrecompileCol.Len()
	sizeFetched := filterFetched.Len()
	sizeEcdsa := isFromCol.Len()

	// Assign accBadPrecompile (backward accumulator of badPrecompileCol)
	accBadPrecompileValues := make([]field.Element, size)
	accBadPrecompileValues[size-1] = badPrecompileCol.Get(size - 1)
	for i := size - 2; i >= 0; i-- {
		val := badPrecompileCol.Get(i)
		accBadPrecompileValues[i].Add(&val, &accBadPrecompileValues[i+1])
	}
	run.AssignColumn(pi.Aux.accBadPrecompile.GetColID(), smartvectors.NewRegular(accBadPrecompileValues))

	// Assign accNbL2Logs (backward accumulator of filterFetched)
	accNbL2LogsValues := make([]field.Element, sizeFetched)
	accNbL2LogsValues[sizeFetched-1] = filterFetched.Get(sizeFetched - 1)
	for i := sizeFetched - 2; i >= 0; i-- {
		val := filterFetched.Get(i)
		accNbL2LogsValues[i].Add(&val, &accNbL2LogsValues[i+1])
	}
	run.AssignColumn(pi.Aux.accNbL2Logs.GetColID(), smartvectors.NewRegular(accNbL2LogsValues))

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
	run.AssignColumn(pi.Aux.FromHi.GetColID(), smartvectors.NewConstant(fromHi, pi.Aux.FromHi.Size()))
	run.AssignColumn(pi.Aux.FromLo.GetColID(), smartvectors.NewConstant(fromLo, pi.Aux.FromLo.Size()))
}
