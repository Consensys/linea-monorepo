package merkle

import (
	"strings"

	"github.com/consensys/accelerated-crypto-monorepo/crypto/mimc"
	"github.com/consensys/accelerated-crypto-monorepo/maths/common/smartvectors"
	"github.com/consensys/accelerated-crypto-monorepo/maths/common/vector"
	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/column"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/ifaces"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/variables"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/wizard"
	"github.com/consensys/accelerated-crypto-monorepo/symbolic"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
	"github.com/consensys/accelerated-crypto-monorepo/utils/parallel"
)

// ComputeMod defines by the modules whose responsibility is
// to recompute the Merkle root from the Merkle proofs.
type ComputeMod struct {

	// The compiled IOP
	comp *wizard.CompiledIOP

	// Number of rows in the module
	NumRows int
	// Number of proofs
	NumProofs int
	// Depth of the proof
	Depth int
	// Name if the name of parent context joined with a specifier.
	Name string
	// Round of the module
	Round int

	// Columns of the modules
	Cols struct {
		// IsInactive is a flags used to detect when a column
		// is not used
		IsInactive ifaces.Column
		// NewProof is a flag indicating that a new proof is
		// being verified
		NewProof ifaces.Column
		// IsEndOfProof is a flag indicating that this is the
		// last row to verify of a proof and that the NodeHash
		// field contains the computed root.
		IsEndOfProof ifaces.Column
		// Root that contains the leaf of the current proof
		Root ifaces.Column
		// Curr contains the current node to be hashes
		Curr ifaces.Column
		// Columns containing the Merkle proof
		Proof ifaces.Column
		// PosBit, indicates whether the current nodes is left
		// or right
		PosBit ifaces.Column
		// PosAcc recomputes the leaf position from the pos-bits
		PosAcc ifaces.Column
		// Zero is a dummy column containing the constant zero
		Zero ifaces.Column
		// Left contains the leftmost node of the current level
		Left ifaces.Column
		// Right contains the rightmost node of the current level
		Right ifaces.Column
		// Interm contains the intermediate hasher state after
		// hashing Left and before hashing Right.
		Interm ifaces.Column
		// NodeHash contains the hash of the parent node.
		NodeHash ifaces.Column
	}
	// Queries of the module
	SugarVar struct {
		// NotNewProof = 1 - NewProof
		NotNewProof *symbolic.Expression
		// IsActive = 1 - IsInactive
		IsActive *symbolic.Expression
		// Root
		Root, RootNext, Curr, PosBit, PosAcc, PosAccPrev *symbolic.Expression
		Proof, Left, Right, NodeHashNext, EndOfProof     *symbolic.Expression
		NewProof, IsInactive, NodeHash, NotEndOfProof    *symbolic.Expression
	}
}

// Declare all the columns of the module. Assumes that Proof as
// been assigned to the module. Also registers all the constraints
func (cm *ComputeMod) Define(comp *wizard.CompiledIOP, round int, name string, numProofs, depth int) {

	// Sanity-check that proof has been assigned
	if cm.Cols.Proof == nil {
		panic("proof must be assigned")
	}

	// Sanity-check the value of NumRows
	numRows := cm.Cols.Proof.Size()
	if numRows != utils.NextPowerOfTwo(numProofs*depth) {
		utils.Panic("numRows %v but numProofs %v and depth %v", numRows, numProofs, depth)
	}

	// Assigns the depth and the number of proofs
	cm.comp = comp
	cm.NumRows = numRows
	cm.NumProofs = numProofs
	cm.Depth = depth
	cm.Round = round
	cm.Name = strings.Join([]string{"MERKLE", "COMPUTEMOD", name}, "_")

	// Declare all the columns
	cm.defineIsInactive()
	cm.defineNewProof()
	cm.defineIsProofEnd()
	cm.defineZero()

	cm.Cols.Root = comp.InsertCommit(cm.Round, cm.colname("ROOT"), cm.NumRows)
	cm.Cols.Curr = comp.InsertCommit(cm.Round, cm.colname("CURR"), cm.NumRows)
	cm.Cols.PosBit = comp.InsertCommit(cm.Round, cm.colname("POSBIT"), cm.NumRows)
	cm.Cols.PosAcc = comp.InsertCommit(cm.Round, cm.colname("POSACC"), cm.NumRows)
	cm.Cols.Left = comp.InsertCommit(cm.Round, cm.colname("LEFT"), cm.NumRows)
	cm.Cols.Interm = comp.InsertCommit(cm.Round, cm.colname("INTERM_STATE"), cm.NumRows)
	cm.Cols.Right = comp.InsertCommit(cm.Round, cm.colname("RIGHT"), cm.NumRows)
	cm.Cols.NodeHash = comp.InsertCommit(cm.Round, cm.colname("NODE_HASH"), cm.NumRows)

	// Initializes all the sugar, we will use for the
	// module constraining
	cm.createSugarVar()

	// And registers all the queries
	cm.rootConsistency()
	cm.rootIsLastNodeHash()
	cm.posbitBoolean()
	cm.posAccConstraint()
	cm.selectLeft()
	cm.selectRight()
	cm.currIsNextNodeHash()
	cm.checkMiMCCompressions()

	// optional constraints
	if !cm.isFullyActive() {
		cm.proofCancelWhenInactive()
	}

}

// Declare and assigns the IsInactive Column
func (cm *ComputeMod) defineIsInactive() {

	// check for potential optimization
	if cm.isFullyActive() {
		// we can just skip this column because all rows are used
		// no need to cancel anything.
		return
	}

	activeSize := cm.Depth * cm.NumProofs

	cm.Cols.IsInactive = cm.comp.InsertPrecomputed(
		cm.colname("IS_INACTIVE"),
		smartvectors.RightPadded(
			vector.Repeat(field.Zero(), activeSize),
			field.One(),
			cm.NumRows,
		),
	)
}

// Declare the NewProof column
func (cm *ComputeMod) defineNewProof() {

	// Compute the active part of the assignment
	window := make([]field.Element, cm.Depth*cm.NumProofs)
	for i := range window {
		if i%cm.Depth == cm.Depth-1 {
			window[i].SetOne()
		}
	}

	assignment := smartvectors.RightZeroPadded(window, cm.NumRows)

	// And register it as a precomputed column
	cm.Cols.NewProof = cm.comp.InsertPrecomputed(
		cm.colname("NEW_PROOF"),
		assignment,
	)
}

// Declare the IsProofEnd column
func (cm *ComputeMod) defineIsProofEnd() {

	// We can fully replace the variable by a periodic sample
	if cm.isFullyActive() {
		return
	}

	// Compute the active part of the assignment
	window := make([]field.Element, cm.Depth*cm.NumProofs)
	for i := range window {
		if i%cm.Depth == 0 {
			window[i].SetOne()
		}
	}

	// registers the window as the full column
	assignment := smartvectors.RightZeroPadded(window, cm.NumRows)

	// And register it as a precomputed column
	cm.Cols.IsEndOfProof = cm.comp.InsertPrecomputed(
		cm.colname("IS_PROOF_END"),
		assignment,
	)
}

// Defines the precomputed column ZERO (always zero)
func (cm *ComputeMod) defineZero() {
	cm.Cols.Zero = cm.comp.InsertPrecomputed(
		cm.colname("ZERO"),
		smartvectors.NewConstant(field.Zero(), cm.NumRows),
	)
}

// Defines all the variables that we will need for the constraints
// of the module.
func (cm *ComputeMod) createSugarVar() {
	sug := &cm.SugarVar
	cols := cm.Cols
	sug.Curr = ifaces.ColumnAsVariable(cols.Curr)
	sug.Root = ifaces.ColumnAsVariable(cols.Root)
	sug.RootNext = ifaces.ColumnAsVariable(column.Shift(cols.Root, 1))
	sug.PosBit = ifaces.ColumnAsVariable(cols.PosBit)
	sug.PosAcc = ifaces.ColumnAsVariable(cols.PosAcc)
	sug.PosAccPrev = ifaces.ColumnAsVariable(column.Shift(cols.PosAcc, -1))
	sug.Proof = ifaces.ColumnAsVariable(cols.Proof)
	sug.Left = ifaces.ColumnAsVariable(cols.Left)
	sug.Right = ifaces.ColumnAsVariable(cols.Right)
	sug.NodeHashNext = ifaces.ColumnAsVariable(column.Shift(cols.NodeHash, 1))
	sug.NewProof = ifaces.ColumnAsVariable(cols.NewProof)
	sug.NodeHash = ifaces.ColumnAsVariable(cols.NodeHash)

	switch cm.isFullyActive() {
	case true:
		// the columns are replaced directly by some variables
		sug.EndOfProof = variables.NewPeriodicSample(cm.Depth, 0)
		sug.IsInactive = symbolic.NewConstant(0)
	case false:
		sug.EndOfProof = ifaces.ColumnAsVariable(cols.IsEndOfProof)
		sug.IsInactive = ifaces.ColumnAsVariable(cols.IsInactive)
	}

	sug.IsActive = symbolic.NewConstant(1).Sub(sug.IsInactive)
	sug.NotNewProof = symbolic.NewConstant(1).Sub(sug.NewProof)
	sug.NotEndOfProof = symbolic.NewConstant(1).Sub(sug.EndOfProof)
}

// Define the query responsible for ensuring that the roots
// are consistent between themselves and with current. It does not
// require a bound cancellation because the inactive flag prevents
// boundary effects.
func (cm *ComputeMod) rootConsistency() {
	sug := cm.SugarVar
	expr := sug.IsActive.
		Mul(sug.RootNext).
		Sub(sug.Root).
		Mul(sug.NotNewProof)
	cm.comp.InsertGlobal(cm.Round, cm.qname("ROOT_CONSISTENCY"), expr, true)
}

// Ensures that the roots column equals the last nodehash of a segment
func (cm *ComputeMod) rootIsLastNodeHash() {
	sug := cm.SugarVar
	expr := sug.Root.Sub(sug.NodeHash).Mul(sug.EndOfProof)
	cm.comp.InsertGlobal(cm.Round, cm.qname("ROOT_IS_LAST_NODEHASH"), expr, true)
}

// Define the query responsible for ensuring that posbits are boolean
// zero when the inactive flag is set.
func (cm *ComputeMod) posbitBoolean() {
	sug := cm.SugarVar
	expr := sug.PosBit.
		Square().
		Mul(sug.IsActive).
		Sub(sug.PosBit)
	cm.comp.InsertGlobal(cm.Round, cm.qname("POSBIT_BOOLEAN"), expr)
}

// Defines the global constraint ensuring that posacc is well constructed
// and consistent with posbit. It does not need bound cancellation (and
// it should not be because we want the inactive flag to work on the last
// column.
func (cm *ComputeMod) posAccConstraint() {
	sug := cm.SugarVar
	expr := symbolic.NewConstant(2).
		Mul(sug.NotEndOfProof).
		Mul(sug.PosAccPrev).
		Add(sug.PosBit).
		Mul(sug.IsActive).
		Sub(sug.PosAcc)
	cm.comp.InsertGlobal(cm.Round, cm.qname("POSACC_CMPT"), expr, true)
}

// Defines the global constraint responsible for ensuring that left was
// correctly constructed.
func (cm *ComputeMod) selectLeft() {
	sug := cm.SugarVar
	expr := sug.Left.
		Sub(
			sug.PosBit.
				Mul(sug.Proof),
		).
		Sub(
			symbolic.NewConstant(1).
				Sub(sug.PosBit).
				Mul(sug.Curr),
		)
	cm.comp.InsertGlobal(cm.Round, cm.qname("SELECT_LEFT"), expr)
}

// Defines the global constraint responsible for ensuring that right was
// correctly constructed.
func (cm *ComputeMod) selectRight() {
	sug := cm.SugarVar
	expr := sug.Right.
		Sub(
			sug.PosBit.
				Mul(sug.Curr),
		).
		Sub(
			symbolic.NewConstant(1).
				Sub(sug.PosBit).
				Mul(sug.Proof),
		)
	cm.comp.InsertGlobal(cm.Round, cm.qname("SELECT_RIGHT"), expr)

}

// The proof should cancel on inactive
func (cm *ComputeMod) proofCancelWhenInactive() {
	sug := cm.SugarVar
	expr := sug.Proof.Mul(sug.IsInactive)
	cm.comp.InsertGlobal(cm.Round, cm.qname("INACTIVE_PROOF_CANCELS"), expr)
}

// Defines the global constraint responsible for ensuring that NodeHash
// is correctly reported into Curr during the computation. No bound cancel
// to ensure that curr is zero at the last row.
func (cm *ComputeMod) currIsNextNodeHash() {
	sug := cm.SugarVar
	expr := sug.IsActive.
		Mul(sug.NodeHashNext).
		Sub(sug.Curr).
		Mul(sug.NotNewProof)
	cm.comp.InsertGlobal(cm.Round, cm.qname("CURR_IS_NEXT_NODE_HASH"), expr, true)
}

// Ensures that the triplets (LEFT, ZERO, INTERM) and (RIGHT, INTERM, NODEHASH)
// are valid MiMC triplets.
func (cm *ComputeMod) checkMiMCCompressions() {
	cols := cm.Cols
	cm.comp.InsertMiMC(cm.Round, cm.qname("MIMC_LEFT"), cols.Left, cols.Zero, cols.Interm)
	cm.comp.InsertMiMC(cm.Round, cm.qname("MIMC_RIGHT"), cols.Right, cols.Interm, cols.NodeHash)
}

func (cm *ComputeMod) colname(name string, args ...any) ifaces.ColID {
	return ifaces.ColIDf("%v_%v", cm.Name, cm.comp.SelfRecursionCount) + "_" + ifaces.ColIDf(name, args...)
}

func (cm *ComputeMod) qname(name string, args ...any) ifaces.QueryID {
	return ifaces.QueryIDf("%v_%v", cm.Name, cm.comp.SelfRecursionCount) + "_" + ifaces.QueryIDf(name, args...)
}

// Assigns the module from an assignment to the inputs of the
// roots and leaves
func (cm *ComputeMod) assign(
	run *wizard.ProverRuntime,
	leaves, pos smartvectors.SmartVector,
) {

	// Function responsible for post-padding with zeroes
	pad := func(vec []field.Element, padVal_ ...field.Element) smartvectors.SmartVector {
		// padding value
		padVal := field.Zero()
		if len(padVal_) > 0 {
			padVal = padVal_[0]
		}
		return smartvectors.RightPadded(vec, padVal, cm.NumRows)
	}

	// Number of active rows
	numActiveRows := cm.Depth * cm.NumProofs

	// List of columns to assign
	var (
		roots    = make([]field.Element, numActiveRows)
		curr     = make([]field.Element, numActiveRows)
		proof    = cm.Cols.Proof.GetColAssignment(run)
		posbit   = make([]field.Element, numActiveRows)
		posacc   = make([]field.Element, numActiveRows)
		left     = make([]field.Element, numActiveRows)
		right    = make([]field.Element, numActiveRows)
		interm   = make([]field.Element, numActiveRows)
		nodehash = make([]field.Element, numActiveRows)
	)

	bitAt := func(x field.Element, i int) uint64 {
		xint := x.Uint64()
		return (xint >> i) & 1
	}

	// Assigns everything in parallel proof per proof
	parallel.Execute(cm.NumProofs, func(start, stop int) {
		for proofNo := start; proofNo < stop; proofNo++ {

			// placeholder for the root of the current proof
			var root field.Element

			// recall that we fill the trace bottom up for every proof
			for level := 0; level < cm.Depth; level++ {
				row := (proofNo+1)*cm.Depth - level - 1

				// assign curr
				if level == 0 {
					curr[row] = leaves.Get(proofNo)
				} else {
					curr[row] = nodehash[row+1]
				}

				// Assign posbit
				posbitUint := bitAt(pos.Get(proofNo), level)
				posbit[row].SetUint64(posbitUint)

				// Assign left, right
				switch posbitUint {
				case 0:
					left[row] = curr[row]
					right[row] = proof.Get(row)
				case 1:
					left[row] = proof.Get(row)
					right[row] = curr[row]
				default:
					utils.Panic("not a bit")
				}

				// And run the mimc compression function
				interm[row] = mimc.BlockCompression(field.Zero(), left[row])
				nodehash[row] = mimc.BlockCompression(interm[row], right[row])
			}

			root = nodehash[proofNo*cm.Depth]

			// Then computes the posacc and root, topdown
			for i := 0; i < cm.Depth; i++ {
				row := proofNo*cm.Depth + i
				if i == 0 {
					posacc[row] = posbit[row]
				} else {
					posacc[row].Double(&posacc[row-1]).Add(&posacc[row], &posbit[row])
				}
				roots[row] = root
			}
		}
	})

	intermPadding := mimc.BlockCompression(field.Zero(), field.Zero())
	nodeHashPadding := mimc.BlockCompression(intermPadding, field.Zero())

	// and assign the freshly computed columns
	cols := cm.Cols
	run.AssignColumn(cols.Root.GetColID(), pad(roots))
	run.AssignColumn(cols.Curr.GetColID(), pad(curr))
	run.AssignColumn(cols.PosBit.GetColID(), pad(posbit))
	run.AssignColumn(cols.PosAcc.GetColID(), pad(posacc))
	run.AssignColumn(cols.Left.GetColID(), pad(left))
	run.AssignColumn(cols.Right.GetColID(), pad(right))
	run.AssignColumn(cols.Interm.GetColID(), pad(interm, intermPadding))
	run.AssignColumn(cols.NodeHash.GetColID(), pad(nodehash, nodeHashPadding))
}

func (cm *ComputeMod) isFullyActive() bool {
	return cm.NumRows == cm.Depth*cm.NumProofs
}
