package invalidityPI

import (
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/accessors"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/ecdsa"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/publicInput/logs"
)

// Public input names for invalidity proofs - these must match what BadPrecompileCircuit expects
const (
	StateRootHash     = "StateRootHash"
	TxHashHi          = "TxHash_Hi"
	TxHashLo          = "TxHash_Lo"
	FromAddress       = "FromAddress"
	HashBadPrecompile = "HashBadPrecompile"
	NbL2Logs          = "NbL2Logs"
)

// InvalidityPI is the module responsible for extracting public inputs
// needed for invalidity proofs (BadPrecompile and TooManyLogs circuits).
type InvalidityPI struct {
	StateRootHash wizard.PublicInput
	// Output columns (size 1) - these become wizard public inputs
	TxHashHi          ifaces.Column
	TxHashLo          ifaces.Column
	FromAddress       ifaces.Column
	HashBadPrecompile ifaces.Column
	NbL2Logs          ifaces.Column

	// Input columns from arithmetization (stored for use in Assign)
	badPrecompileCol ifaces.Column
	addresses        *ecdsa.Addresses

	// Input columns from ECDSA TxSignature module
	txSignature *ecdsa.TxSignature

	// Input columns from logs module
	fetchedL2L1 *logs.ExtractedData

	// Extractor holds the LocalOpening queries for the public inputs
	Extractor InvalidityPIExtractor
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

	// Create output columns of size 1 (single values)
	txHashHi := comp.InsertCommit(0, ifaces.ColIDf("%s_TX_HASH_HI", name), 1)
	txHashLo := comp.InsertCommit(0, ifaces.ColIDf("%s_TX_HASH_LO", name), 1)
	fromAddress := comp.InsertCommit(0, ifaces.ColIDf("%s_FROM_ADDRESS", name), 1)
	hashBadPrecompile := comp.InsertCommit(0, ifaces.ColIDf("%s_HASH_BAD_PRECOMPILE", name), 1)
	nbL2Logs := comp.InsertCommit(0, ifaces.ColIDf("%s_NB_L2_LOGS", name), 1)

	pi := &InvalidityPI{
		StateRootHash: comp.GetPublicInput(StateRootHash),
		// Output columns
		TxHashHi:          txHashHi,
		TxHashLo:          txHashLo,
		FromAddress:       fromAddress,
		HashBadPrecompile: hashBadPrecompile,
		NbL2Logs:          nbL2Logs,

		// Input columns from arithmetization
		badPrecompileCol: comp.Columns.GetHandle("hub.PROVER_ILLEGAL_TRANSACTION_DETECTED"),
		fetchedL2L1:      fetchedL2L1,
		txSignature:      ecdsa.Ant.TxSignature,
		addresses:        ecdsa.Ant.Addresses,
	}

	// Define constraints over the columns of pi.
	pi.defineConstraints(comp)

	// Create extractor and register public inputs
	pi.generateExtractor(comp, name)

	return pi
}

// defineConstraints defines constraints over the columns of pi.
func (pi *InvalidityPI) defineConstraints(comp *wizard.CompiledIOP) {

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
		HashBadPrecompile: createNewLocalOpening(pi.HashBadPrecompile),
		NbL2Logs:          createNewLocalOpening(pi.NbL2Logs),
	}

	// Register as wizard public inputs with names matching BadPrecompileCircuit expectations
	comp.PublicInputs = append(comp.PublicInputs,
		wizard.PublicInput{Name: TxHashHi, Acc: accessors.NewLocalOpeningAccessor(pi.Extractor.TxHashHi, 0)},
		wizard.PublicInput{Name: TxHashLo, Acc: accessors.NewLocalOpeningAccessor(pi.Extractor.TxHashLo, 0)},
		wizard.PublicInput{Name: FromAddress, Acc: accessors.NewLocalOpeningAccessor(pi.Extractor.FromAddress, 0)},
		wizard.PublicInput{Name: HashBadPrecompile, Acc: accessors.NewLocalOpeningAccessor(pi.Extractor.HashBadPrecompile, 0)},
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
	isFromCol := pi.addresses.IsAddressFromTxnData.GetColAssignment(run)
	sizeEcdsa := isFromCol.Len()
	for i := 0; i < sizeEcdsa; i++ {
		source := isFromCol.Get(i)
		if source.IsOne() {
			fromAddressHi := pi.addresses.AddressHi.GetColAssignmentAt(run, i)
			fromAddressLo := pi.addresses.AddressLo.GetColAssignmentAt(run, i)
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

	// 4. Extract TxHash from ECDSA module - find the row where IsTxHash = 1
	// Note: For invalidity proofs, there should be only a single tx in the trace
	isTxHashCol := pi.txSignature.IsTxHash.GetColAssignment(run)
	ecdsaSize := isTxHashCol.Len()

	for i := 0; i < ecdsaSize; i++ {
		isTxHash := isTxHashCol.Get(i)
		if isTxHash.IsOne() {
			txHashHi = pi.txSignature.TxHashHi.GetColAssignmentAt(run, i)
			txHashLo = pi.txSignature.TxHashLo.GetColAssignmentAt(run, i)
			break
		}
	}

	// 5. Extract NbL2Logs from
	filterFetched := pi.fetchedL2L1.FilterFetched.GetColAssignment(run)
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

	// Assign local openings
	run.AssignLocalPoint(pi.Extractor.TxHashHi.ID, txHashHi)
	run.AssignLocalPoint(pi.Extractor.TxHashLo.ID, txHashLo)
	run.AssignLocalPoint(pi.Extractor.FromAddress.ID, fromAddress)
	run.AssignLocalPoint(pi.Extractor.HashBadPrecompile.ID, hashBadPrecompile)
	run.AssignLocalPoint(pi.Extractor.NbL2Logs.ID, field.NewElement(nbL2Logs))
}
