package invalidityPI

import (
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/accessors"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	commonconstraints "github.com/consensys/linea-monorepo/prover/zkevm/prover/common/common_constraints"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/ecdsa"
)

// Public input names for invalidity proofs
var (
	HasBadPrecompile = "HasBadPrecompile"
	FromAddress      = "FromAddress"
)

// InvalidityPI is the module responsible for extracting public inputs
// needed for invalidity proofs, specifically detecting illegal precompile calls
// (RIPEMD-160 or BLAKE2f) and capturing the sender address.
type InvalidityPI struct {
	// Output columns (size 1)
	// HasBadPrecompile is 1 if any illegal precompile was called, 0 otherwise
	HasBadPrecompile ifaces.Column
	// FromAddress is the sender address of the transaction that called the illegal precompile
	FromAddress ifaces.Column

	// Input columns from arithmetization (stored for use in Assign)
	badPrecompileCol ifaces.Column
	txnData          *ecdsa.TxnData

	// Extractor holds the LocalOpening queries for the public inputs
	Extractor InvalidityPIExtractor
}

// InvalidityPIExtractor holds LocalOpening queries for invalidity public inputs
type InvalidityPIExtractor struct {
	HasBadPrecompile query.LocalOpening
	FromAddress      query.LocalOpening
}

// NewInvalidityPIZkEvm creates a new InvalidityPI module using columns from the arithmetization
func NewInvalidityPIZkEvm(comp *wizard.CompiledIOP) *InvalidityPI {
	panic("implementation is not complete - this is just pseudocode")

	name := "INVALIDITY_PI"

	// Create output columns of size 1 (single values)
	hasBadPrecompile := comp.InsertCommit(0, ifaces.ColIDf("%s_HAS_BAD_PRECOMPILE", name), 1)
	fromAddress := comp.InsertCommit(0, ifaces.ColIDf("%s_FROM_ADDRESS", name), 1)

	pi := &InvalidityPI{
		// Output columns
		HasBadPrecompile: hasBadPrecompile,
		FromAddress:      fromAddress,
		// Input columns from arithmetization
		badPrecompileCol: comp.Columns.GetHandle("hub.PROVER_ILLEGAL_TRANSACTION_DETECTED"),
		txnData:          ecdsa.GetTxnDataArithmetization(comp),
	}

	// Define constraints
	defineInvalidityPI(comp, pi, name)

	// Create extractor
	pi.generateExtractor(comp, name)

	return pi
}

// defineInvalidityPI defines constraints for the InvalidityPI module
func defineInvalidityPI(comp *wizard.CompiledIOP, pi *InvalidityPI, name string) {
	panic("implementation is not complete - this is just pseudocode")
	// Both columns must be constant since they are single-value columns
	commonconstraints.MustBeConstant(comp, pi.HasBadPrecompile)
	commonconstraints.MustBeConstant(comp, pi.FromAddress)
}

// generateExtractor creates LocalOpening queries and registers public inputs
func (pi *InvalidityPI) generateExtractor(comp *wizard.CompiledIOP, name string) {
	panic("implementation is not complete - this is just pseudocode")
	createNewLocalOpening := func(col ifaces.Column) query.LocalOpening {
		return comp.InsertLocalOpening(0, ifaces.QueryIDf("%s_LOCAL_OPENING_%s", name, col.GetColID()), col)
	}

	pi.Extractor = InvalidityPIExtractor{
		HasBadPrecompile: createNewLocalOpening(pi.HasBadPrecompile),
		FromAddress:      createNewLocalOpening(pi.FromAddress),
	}

	// Register as wizard public inputs
	comp.PublicInputs = append(comp.PublicInputs,
		wizard.PublicInput{Name: HasBadPrecompile, Acc: accessors.NewLocalOpeningAccessor(pi.Extractor.HasBadPrecompile, 0)},
		wizard.PublicInput{Name: FromAddress, Acc: accessors.NewLocalOpeningAccessor(pi.Extractor.FromAddress, 0)},
	)
}

// Assign assigns values to the InvalidityPI columns using data from the runtime
func (pi *InvalidityPI) Assign(run *wizard.ProverRuntime) {
	panic("implementation is not complete - this is just pseudocode")
	var (
		hasBadPrecompile = field.Zero()
		fromAddress      = field.Zero()
	)

	// Scan the badPrecompile column to find if any value is non-zero
	badPrecompileCol := pi.badPrecompileCol.GetColAssignment(run)
	size := badPrecompileCol.Len()

	for i := 0; i < size; i++ {
		val := badPrecompileCol.Get(i)
		if !val.IsZero() {
			hasBadPrecompile.SetOne()
			break
		}
	}

	// If we found a bad precompile, extract the from address from TxnData
	if !hasBadPrecompile.IsZero() {
		// Find the first user transaction's from address
		// Look for the row where USER=1 and CT=0 (first row of transaction data)
		userCol := pi.txnData.User.GetColAssignment(run)
		ctCol := pi.txnData.Ct.GetColAssignment(run)
		fromLoCol := pi.txnData.FromLo.GetColAssignment(run)
		txnSize := userCol.Len()

		for i := 0; i < txnSize; i++ {
			user := userCol.Get(i)
			ct := ctCol.Get(i)
			if !user.IsZero() && ct.IsZero() {
				// Found a user transaction at CT=0, get the from address
				fromAddress = fromLoCol.Get(i)
				break
			}
		}
	}

	// Assign the columns
	run.AssignColumn(pi.HasBadPrecompile.GetColID(), smartvectors.NewConstant(hasBadPrecompile, 1))
	run.AssignColumn(pi.FromAddress.GetColID(), smartvectors.NewConstant(fromAddress, 1))

	// Assign local openings
	run.AssignLocalPoint(pi.Extractor.HasBadPrecompile.ID, hasBadPrecompile)
	run.AssignLocalPoint(pi.Extractor.FromAddress.ID, fromAddress)
}
