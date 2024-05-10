package accumulator

import (
	"strings"

	"github.com/consensys/zkevm-monorepo/prover/maths/field"
	"github.com/consensys/zkevm-monorepo/prover/protocol/column"
	"github.com/consensys/zkevm-monorepo/prover/protocol/column/verifiercol"
	"github.com/consensys/zkevm-monorepo/prover/protocol/dedicated/byte32cmp"
	"github.com/consensys/zkevm-monorepo/prover/protocol/dedicated/merkle"
	"github.com/consensys/zkevm-monorepo/prover/protocol/ifaces"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
	"github.com/consensys/zkevm-monorepo/prover/symbolic"
	"github.com/consensys/zkevm-monorepo/prover/utils"
	"github.com/consensys/zkevm-monorepo/prover/utils/types"
)

const (
	// Dimensions of the Merkle-proof verification module
	merkleTreeDepth int = 40
	maxNumProofs    int = 1 << 13

	// Column names
	ACCUMULATOR_PROOFS_NAME                ifaces.ColID = "ACCUMULATOR_PROOFS"
	ACCUMULATOR_ROOTS_NAME                 ifaces.ColID = "ACCUMULATOR_ROOTS"
	ACCUMULATOR_POSITIONS_NAME             ifaces.ColID = "ACCUMULATOR_POSITIONS"
	ACCUMULATOR_LEAVES_NAME                ifaces.ColID = "ACCUMULATOR_LEAVES"
	ACCUMULATOR_USE_NEXT_MERKLE_PROOF_NAME ifaces.ColID = "ACCUMULATOR_USE_NEXT_PROOF"
	ACCUMULATOR_IS_ACTIVE_NAME             ifaces.ColID = "ACCUMULATOR_IS_ACTIVE"
	// Column for sequentiality check
	ACCUMULATOR_COUNTER_NAME ifaces.ColID = "ACCUMULATOR_COUNTER"
	// Column for local consistency check
	ACCUMULATOR_IS_FIRST_NAME         ifaces.ColID = "ACCUMULATOR_IS_FIRST"
	ACCUMULATOR_IS_INSERT_NAME        ifaces.ColID = "ACCUMULATOR_IS_INSERT"
	ACCUMULATOR_IS_DELETE_NAME        ifaces.ColID = "ACCUMULATOR_IS_DELETE"
	ACCUMULATOR_IS_UPDATE_NAME        ifaces.ColID = "ACCUMULATOR_IS_UPDATE"
	ACCUMULATOR_IS_READ_ZERO_NAME     ifaces.ColID = "ACCUMULATOR_IS_READ_ZERO"
	ACCUMULATOR_IS_READ_NON_ZERO_NAME ifaces.ColID = "ACCUMULATOR_IS_READ_NON_ZERO"
	// Columns for sandwich check
	ACCUMULATOR_HKEY       ifaces.ColID = "ACCUMULATOR_HKEY"
	ACCUMULATOR_HKEY_MINUS ifaces.ColID = "ACCUMULATOR_HKEY_MINUS"
	ACCUMULATOR_HKEY_PLUS  ifaces.ColID = "ACCUMULATOR_HKEY_PLUS"
	// Columns for pointer check
	ACCUMULATOR_LEAF_MINUS_INDEX   ifaces.ColID = "ACCUMULATOR_LEAF_MINUS_INDEX"
	ACCUMULATOR_LEAF_MINUS_NEXT    ifaces.ColID = "ACCUMULATOR_LEAF_MINUS_NEXT"
	ACCUMULATOR_LEAF_PLUS_INDEX    ifaces.ColID = "ACCUMULATOR_LEAF_PLUS_INDEX"
	ACCUMULATOR_LEAF_PLUS_PREV     ifaces.ColID = "ACCUMULATOR_LEAF_PLUS_PREV"
	ACCUMULATOR_LEAF_DELETED_INDEX ifaces.ColID = "ACCUMULATOR_LEAF_DELETED_INDEX"
	ACCUMULATOR_LEAF_DELETED_PREV  ifaces.ColID = "ACCUMULATOR_LEAF_DELETED_PREV"
	ACCUMULATOR_LEAF_DELETED_NEXT  ifaces.ColID = "ACCUMULATOR_LEAF_DELETED_NEXT"
	// Columns for leaf hash check (some more columns for this are declared below)
	ACCUMULATOR_LEAF_HASHES   ifaces.ColID = "ACCUMULATOR_LEAF_HASHES"
	ACCUMULATOR_IS_EMPTY_LEAF ifaces.ColID = "ACCUMULATOR_IS_EMPTY_LEAF"
	// Columns for NextFreeNode consistency check
	ACCUMULATOR_NEXT_FREE_NODE ifaces.ColID = "ACCUMULATOR_NEXT_FREE_NODE"
	ACCUMULATOR_INSERTION_PATH ifaces.ColID = "ACCUMULATOR_INSERTION_PATH"
	ACCUMULATOR_IS_INSERT_ROW3 ifaces.ColID = "ACCUMULATOR_IS_INSERT_ROW3"
)

// Accumulator module
type Accumulator struct {
	// The compiled IOP
	comp *wizard.CompiledIOP
	// Round of the module
	Round        int
	LeaveSize    int
	ProofSize    int
	Name         string
	MaxNumProofs int
	Cols         struct {
		Leaves    ifaces.Column
		Roots     ifaces.Column
		Positions ifaces.Column
		Proofs    ifaces.Column
		// Column to verify reuse of Merkle proofs in INSERT, DELETE, and UPDATE operations
		UseNextMerkleProof ifaces.Column
		// Column denoting the active area of the accumulator module
		IsActiveAccumulator ifaces.Column
		// Column to check sequentiality of the accumulator module with Merkle module
		AccumulatorCounter ifaces.Column
		// Column to verify the two equalities of intermediateRoot1 and intermediateRoot3, and empty
		// leafs for INSERT and DELETE operation and one equality of root in IsReadZero operation
		IsFirst ifaces.Column
		// Column indicating an INSERT operation
		IsInsert ifaces.Column
		// Column indicating an DELETE operation
		IsDelete ifaces.Column
		// Column indicating an UPDATE operation
		IsUpdate ifaces.Column
		// Column indicating an READ-ZERO operation
		IsReadZero ifaces.Column
		// Column indicating an READ-NONZERO operation
		IsReadNonZero ifaces.Column

		// Columns for the sandwitch check
		// Column storing the hash of the key of the trace
		HKey ifaces.Column
		// Column storing the hash of the key of the previous leaf
		HKeyMinus ifaces.Column
		// Column storing the hash of the key of the next leaf
		HKeyPlus ifaces.Column

		// Columns for the pointer check
		// Column storing the index of the minus leaf
		LeafMinusIndex ifaces.Column
		// Column storing the index of the next leaf of the minus leaf
		LeafMinusNext ifaces.Column
		// Column storing the index of the plus leaf
		LeafPlusIndex ifaces.Column
		// Column storing the index of the previous leaf of the plus leaf
		LeafPlusPrev ifaces.Column
		// Column storing the index of the deleted leaf
		LeafDeletedIndex ifaces.Column
		// Column storing the index of the previous leaf of the deleted leaf
		LeafDeletedPrev ifaces.Column
		// Column storing the index of the next leaf of the deleted leaf
		LeafDeletedNext ifaces.Column

		// Columns for leaf hashing check
		// LeafOpening contains four columns corresponding to HKey, HVal, Prev, and Next
		LeafOpenings []ifaces.Column
		// Interm contains the three intermediate states corresponding to the MiMC block computation
		Interm []ifaces.Column
		// Zero contains the column with zero value, used in the MiMc query
		Zero ifaces.Column
		// LeafHash contains the leafHashes (the final MiMC block), equals with Leaves, except when it is empty leaf
		LeafHashes ifaces.Column
		// IsEmptyLeaf is one when Leaves contains empty leaf and does not match with LeafHash
		IsEmptyLeaf ifaces.Column

		// Columns to check NextFreeNode consistency
		// NextFreeNode stores the nextFreeNode for each row of every operation
		NextFreeNode ifaces.Column
		// InsertionPath stores the index of the newly inserted leaf by INSERT
		InsertionPath ifaces.Column
		// IsInsertRow3 is one for row 3 of INSERT operation
		IsInsertRow3 ifaces.Column
	}
}

// Funtion registering and committing all the columns and queries in the Accumulator module
func (am *Accumulator) Define(comp *wizard.CompiledIOP, name string) {

	// All the columns and queries from the state-manager are for the round 0
	am.Round = 0
	am.comp = comp
	am.MaxNumProofs = maxNumProofs
	am.Name = strings.Join([]string{"ACCUMULATOR", "COMPUTEMOD", name}, "_")

	// Computes the size of the modules
	am.LeaveSize = utils.NextPowerOfTwo(maxNumProofs)
	am.ProofSize = utils.NextPowerOfTwo(maxNumProofs * merkleTreeDepth)

	// Initializes the columns
	am.Cols.Leaves = comp.InsertCommit(am.Round, ACCUMULATOR_LEAVES_NAME, am.LeaveSize)
	am.Cols.Roots = comp.InsertCommit(am.Round, ACCUMULATOR_ROOTS_NAME, am.LeaveSize)
	am.Cols.Positions = comp.InsertCommit(am.Round, ACCUMULATOR_POSITIONS_NAME, am.LeaveSize)
	am.Cols.Proofs = comp.InsertCommit(am.Round, ACCUMULATOR_PROOFS_NAME, am.ProofSize)
	am.Cols.UseNextMerkleProof = comp.InsertCommit(am.Round, ACCUMULATOR_USE_NEXT_MERKLE_PROOF_NAME, am.LeaveSize)
	am.Cols.IsActiveAccumulator = comp.InsertCommit(am.Round, ACCUMULATOR_IS_ACTIVE_NAME, am.LeaveSize)
	am.Cols.AccumulatorCounter = comp.InsertCommit(am.Round, ACCUMULATOR_COUNTER_NAME, am.LeaveSize)
	am.Cols.IsFirst = comp.InsertCommit(am.Round, ACCUMULATOR_IS_FIRST_NAME, am.LeaveSize)
	am.Cols.IsInsert = comp.InsertCommit(am.Round, ACCUMULATOR_IS_INSERT_NAME, am.LeaveSize)
	am.Cols.IsDelete = comp.InsertCommit(am.Round, ACCUMULATOR_IS_DELETE_NAME, am.LeaveSize)
	am.Cols.IsUpdate = comp.InsertCommit(am.Round, ACCUMULATOR_IS_UPDATE_NAME, am.LeaveSize)
	am.Cols.IsReadZero = comp.InsertCommit(am.Round, ACCUMULATOR_IS_READ_ZERO_NAME, am.LeaveSize)
	am.Cols.IsReadNonZero = comp.InsertCommit(am.Round, ACCUMULATOR_IS_READ_NON_ZERO_NAME, am.LeaveSize)

	// columns for the sandwitch check
	am.Cols.HKey = comp.InsertCommit(am.Round, ACCUMULATOR_HKEY, am.LeaveSize)
	am.Cols.HKeyMinus = comp.InsertCommit(am.Round, ACCUMULATOR_HKEY_MINUS, am.LeaveSize)
	am.Cols.HKeyPlus = comp.InsertCommit(am.Round, ACCUMULATOR_HKEY_PLUS, am.LeaveSize)
	// columns for the pointer check
	am.Cols.LeafMinusIndex = comp.InsertCommit(am.Round, ACCUMULATOR_LEAF_MINUS_INDEX, am.LeaveSize)
	am.Cols.LeafMinusNext = comp.InsertCommit(am.Round, ACCUMULATOR_LEAF_MINUS_NEXT, am.LeaveSize)
	am.Cols.LeafPlusIndex = comp.InsertCommit(am.Round, ACCUMULATOR_LEAF_PLUS_INDEX, am.LeaveSize)
	am.Cols.LeafPlusPrev = comp.InsertCommit(am.Round, ACCUMULATOR_LEAF_PLUS_PREV, am.LeaveSize)
	am.Cols.LeafDeletedIndex = comp.InsertCommit(am.Round, ACCUMULATOR_LEAF_DELETED_INDEX, am.LeaveSize)
	am.Cols.LeafDeletedPrev = comp.InsertCommit(am.Round, ACCUMULATOR_LEAF_DELETED_PREV, am.LeaveSize)
	am.Cols.LeafDeletedNext = comp.InsertCommit(am.Round, ACCUMULATOR_LEAF_DELETED_NEXT, am.LeaveSize)
	// Leaf hashing columns commitments
	am.commitLeafHashingCols()
	// NextFreeNode check
	am.Cols.NextFreeNode = comp.InsertCommit(am.Round, ACCUMULATOR_NEXT_FREE_NODE, am.LeaveSize)
	am.Cols.InsertionPath = comp.InsertCommit(am.Round, ACCUMULATOR_INSERTION_PATH, am.LeaveSize)
	am.Cols.IsInsertRow3 = comp.InsertCommit(am.Round, ACCUMULATOR_IS_INSERT_ROW3, am.LeaveSize)

	// Declare constraints
	// Checks the two Root equalities for Insert,
	// also checks the booleanity of the column IsInsertRow1
	am.checkInsert()

	// Checks the two Root equalities for Delete,
	// also checks the booleanity of the column IsDeleteRow1
	am.checkDelete()

	// Checks the root equality for the ReadZero operation, also checks the
	// booleanity of the IsReadZeroRow1 column
	am.checkReadZero()

	// Booleanity check on IsActitiveAccumulator and UseNextMerkleProof,
	// and row wise increment check for AccumulatorCounter
	am.checkConsistency()

	// check that Leaf[i+3] and Leaf[i+4] are empty leaves when there is a
	// INSERT or a DELETE operation respectively.
	am.checkEmptyLeaf()

	// Sandwitch check for INSERT and READ-ZERO operations
	am.checkSandwitch()

	// Pointer check for INSERT, READ-ZERO, and DELETE operations
	am.checkPointer()

	// Check leaf hashes
	am.checkLeafHashes()

	// Check NextFreeNode is constant through a segment unless there is an INSERT operation
	am.checkNextFreeNode()
	// Send the columns to the Merkle gadget for the rest of verification (along with the reuse of Merkle proofs) in the Merkle gadget
	merkle.MerkleProofCheckWithReuse(
		comp,
		"ACCUMULATOR_MERKLE_PROOFS",
		merkleTreeDepth, maxNumProofs,
		am.Cols.Proofs, am.Cols.Roots, am.Cols.Leaves, am.Cols.Positions, am.Cols.UseNextMerkleProof, am.Cols.IsActiveAccumulator, am.Cols.AccumulatorCounter,
	)
}

func (am *Accumulator) commitLeafHashingCols() {
	ACCUMULATOR_LEAF_OPENING := make([]ifaces.ColID, 4)
	ACCUMULATOR_INTERM := make([]ifaces.ColID, 3)
	ACCUMULATOR_LEAF_OPENING[0] = "ACCUMULATOR_LEAF_OPENING_PREV"
	ACCUMULATOR_LEAF_OPENING[1] = "ACCUMULATOR_LEAF_OPENING_NEXT"
	ACCUMULATOR_LEAF_OPENING[2] = "ACCUMULATOR_LEAF_OPENING_HKEY"
	ACCUMULATOR_LEAF_OPENING[3] = "ACCUMULATOR_LEAF_OPENING_HVAL"
	ACCUMULATOR_INTERM[0] = "ACCUMULATOR_INTERM_PREV"
	ACCUMULATOR_INTERM[1] = "ACCUMULATOR_INTERM_NEXT"
	ACCUMULATOR_INTERM[2] = "ACCUMULATOR_INTERM_HKEY"
	am.Cols.LeafOpenings = make([]ifaces.Column, 4)
	am.Cols.Interm = make([]ifaces.Column, 3)
	am.Cols.Zero = verifiercol.NewConstantCol(field.Zero(), am.LeaveSize)
	for i := 0; i < len(am.Cols.LeafOpenings); i++ {
		am.Cols.LeafOpenings[i] = am.comp.InsertCommit(am.Round, ACCUMULATOR_LEAF_OPENING[i], am.LeaveSize)
	}
	for i := 0; i < len(am.Cols.Interm); i++ {
		am.Cols.Interm[i] = am.comp.InsertCommit(am.Round, ACCUMULATOR_INTERM[i], am.LeaveSize)
	}
	am.Cols.LeafHashes = am.comp.InsertCommit(am.Round, ACCUMULATOR_LEAF_HASHES, am.LeaveSize)
	am.Cols.IsEmptyLeaf = am.comp.InsertCommit(am.Round, ACCUMULATOR_IS_EMPTY_LEAF, am.LeaveSize)

}

func (am *Accumulator) checkInsert() {
	cols := am.Cols

	// (Root[i+1] - Root[i+2]) * IsActiveAccumulator[i] * IsFirst[i] * IsInsert[i] = 0, The (i+1)th and (i+2)th roots are equal when there is an INSERT operation and the accumulator is active.
	expr1 := symbolic.Sub(ifaces.ColumnAsVariable(column.Shift(cols.Roots, 1)), ifaces.ColumnAsVariable(column.Shift(cols.Roots, 2)))
	expr1 = symbolic.Mul(cols.IsActiveAccumulator, cols.IsFirst, cols.IsInsert, expr1)
	am.comp.InsertGlobal(am.Round, am.qname("ROOT_EQUALITY_INSERT_1"), expr1)

	// (Root[i+3] - Root[i+4]) * IsActiveAccumulator[i] * IsFirst[i] * IsInsert[i] = 0, The (i+3)th and (i+4)th roots are equal when there is an INSERT operation and the accumulator is active.
	expr2 := symbolic.Sub(ifaces.ColumnAsVariable(column.Shift(cols.Roots, 3)), ifaces.ColumnAsVariable(column.Shift(cols.Roots, 4)))
	expr2 = symbolic.Mul(cols.IsActiveAccumulator, cols.IsFirst, cols.IsInsert, expr2)
	am.comp.InsertGlobal(am.Round, am.qname("ROOT_EQUALITY_INSERT_2"), expr2)

	// Booleanity of IsFirst
	expr3 := symbolic.Sub(
		symbolic.Mul(symbolic.Square(cols.IsFirst), cols.IsActiveAccumulator),
		cols.IsFirst)
	am.comp.InsertGlobal(am.Round, am.qname("IS_FIRST_BOOLEAN"), expr3)

	// Booleanity of IsInsert
	expr4 := symbolic.Sub(
		symbolic.Mul(symbolic.Square(cols.IsInsert), cols.IsActiveAccumulator),
		cols.IsInsert)
	am.comp.InsertGlobal(am.Round, am.qname("IS_INSERT_BOOLEAN"), expr4)
}

func (am *Accumulator) checkDelete() {
	cols := am.Cols

	// (Root[i+1] - Root[i+2]) * IsActiveAccumulator[i] * IsFirst[i] * IsDelete[i] = 0, The (i+1)th and (i+2)th roots are equal when there is a DELETE operation and the accumulator is active.
	expr1 := symbolic.Sub(ifaces.ColumnAsVariable(column.Shift(cols.Roots, 1)), ifaces.ColumnAsVariable(column.Shift(cols.Roots, 2)))
	expr1 = symbolic.Mul(cols.IsActiveAccumulator, cols.IsFirst, cols.IsDelete, expr1)
	am.comp.InsertGlobal(am.Round, am.qname("ROOT_EQUALITY_DELETE_1"), expr1)

	// (Root[i+3] - Root[i+4]) * IsActiveAccumulator[i] * IsFirst[i] * IsDelete[i] = 0, The (i+3)th and (i+4)th roots are equal when there is a DELETE operation and the accumulator is active.
	expr2 := symbolic.Sub(ifaces.ColumnAsVariable(column.Shift(cols.Roots, 3)), ifaces.ColumnAsVariable(column.Shift(cols.Roots, 4)))
	expr2 = symbolic.Mul(cols.IsActiveAccumulator, cols.IsFirst, cols.IsDelete, expr2)
	am.comp.InsertGlobal(am.Round, am.qname("ROOT_EQUALITY_DELETE_2"), expr2)

	// Booleanity of IsDelete
	expr3 := symbolic.Sub(
		symbolic.Mul(symbolic.Square(cols.IsDelete), cols.IsActiveAccumulator),
		cols.IsDelete)
	am.comp.InsertGlobal(am.Round, am.qname("IS_DELETE_BOOLEAN"), expr3)
}

func (am *Accumulator) checkReadZero() {
	cols := am.Cols

	// (Root[i] - Root[i+1]) * IsActiveAccumulator[i] * IsFirst[i] * IsReadZero[i] = 0, The ith and (i+1)th roots are equal when there is a READ-ZERO operation and the accumulator is active.
	expr1 := symbolic.Sub(cols.Roots, ifaces.ColumnAsVariable(column.Shift(cols.Roots, 1)))
	expr1 = symbolic.Mul(cols.IsActiveAccumulator, cols.IsFirst, cols.IsReadZero, expr1)
	am.comp.InsertGlobal(am.Round, am.qname("ROOT_EQUALITY_READ_ZERO"), expr1)

	// Booleanity of IsReadZero
	expr2 := symbolic.Sub(
		symbolic.Mul(symbolic.Square(cols.IsReadZero), cols.IsActiveAccumulator),
		cols.IsReadZero)
	am.comp.InsertGlobal(am.Round, am.qname("IS_READ_ZERO_BOOLEAN"), expr2)
}

func (am *Accumulator) checkConsistency() {
	cols := am.Cols

	// Booleanity of IsUpdate
	expr1 := symbolic.Sub(
		symbolic.Mul(symbolic.Square(cols.IsUpdate), cols.IsActiveAccumulator),
		cols.IsUpdate)
	am.comp.InsertGlobal(am.Round, am.qname("IS_UPDATE_BOOLEAN"), expr1)

	// Booleanity of IsReadNonZero
	expr2 := symbolic.Sub(
		symbolic.Mul(symbolic.Square(cols.IsReadNonZero), cols.IsActiveAccumulator),
		cols.IsReadNonZero)
	am.comp.InsertGlobal(am.Round, am.qname("IS_READ_NON_ZERO_BOOLEAN"), expr2)

	// Booleanity of IsActiveAccumulator
	expr3 := symbolic.Sub(
		symbolic.Square(cols.IsActiveAccumulator),
		cols.IsActiveAccumulator)
	am.comp.InsertGlobal(am.Round, am.qname("IS_IS_ACTIVE_ACCUMULATOR_BOOLEAN"), expr3)

	// Booleanity of UseNextMerkleProof
	expr4 := symbolic.Sub(
		symbolic.Mul(symbolic.Square(cols.UseNextMerkleProof), cols.IsActiveAccumulator),
		cols.UseNextMerkleProof)
	am.comp.InsertGlobal(am.Round, am.qname("USE_NEXT_MERKLE_PROOF_BOOLEAN"), expr4)

	// Row-wise increment of AccumulatorCounter
	// IsActiveAccumulator[i+1] * (AccumulatorCounter[i+1] - AccumulatorCounter[i] - 1)
	expr5 := symbolic.Mul(ifaces.ColumnAsVariable(column.Shift(cols.IsActiveAccumulator, 1)),
		symbolic.Sub(ifaces.ColumnAsVariable(column.Shift(cols.AccumulatorCounter, 1)),
			cols.AccumulatorCounter, symbolic.NewConstant(1)))
	am.comp.InsertGlobal(am.Round, am.qname("COUNTER_INCREMENT"), expr5)
	// Local constraint that the counter starts at zero
	am.comp.InsertLocal(am.Round, am.qname("COUNTER_LOCAL"), symbolic.Sub(cols.AccumulatorCounter, 0))

}

func (am *Accumulator) checkEmptyLeaf() {
	// Creating the emptyLeaf column
	emptyLeafBytes32 := types.Bytes32{}
	emptyLeafBytes := emptyLeafBytes32[:]
	var emptyLeafField field.Element
	if err := emptyLeafField.SetBytesCanonical(emptyLeafBytes[:]); err != nil {
		panic(err)
	}
	emptyLeaf := verifiercol.NewConstantCol(emptyLeafField, am.LeaveSize)
	cols := am.Cols

	// (Leaf[i+2] - emptyLeaf) * IsActiveAccumulator[i] * IsFirst[i] * IsInsert[i]
	expr1 := symbolic.Sub(ifaces.ColumnAsVariable(column.Shift(cols.Leaves, 2)), emptyLeaf)
	expr1 = symbolic.Mul(cols.IsActiveAccumulator, cols.IsFirst, cols.IsInsert, expr1)
	am.comp.InsertGlobal(am.Round, am.qname("EMPTY_LEAVES_FOR_INSERT"), expr1)

	// (Leaf[i+3] - emptyLeaf) * IsActiveAccumulator[i] * IsFirst[i] * IsDelete[i]
	expr2 := symbolic.Sub(ifaces.ColumnAsVariable(column.Shift(cols.Leaves, 3)), emptyLeaf)
	expr2 = symbolic.Mul(cols.IsActiveAccumulator, cols.IsFirst, cols.IsDelete, expr2)
	am.comp.InsertGlobal(am.Round, am.qname("EMPTY_LEAVES_FOR_DELETE"), expr2)
}

func (am *Accumulator) checkSandwitch() {
	cols := am.Cols
	// We want sandwitch check only at row 1 of INSERT and READ-ZERO
	activeRow := symbolic.Add(symbolic.Mul(cols.IsFirst, cols.IsInsert), symbolic.Mul(cols.IsFirst, cols.IsReadZero))
	byte32cmp.Bytes32Cmp(am.comp, 16, 16, string(am.qname("CMP_HKEY_HKEY_MINUS")), am.Cols.HKey, am.Cols.HKeyMinus, activeRow)
	byte32cmp.Bytes32Cmp(am.comp, 16, 16, string(am.qname("CMP_HKEY_PLUS_HKEY")), am.Cols.HKeyPlus, am.Cols.HKey, activeRow)
}

func (am *Accumulator) checkPointer() {
	cols := am.Cols
	// Check #1 for INSERT: IsFirst[i] * IsInsert[i] * (LeafMinusNext[i] - LeafPlusIndex[i])
	expr1 := symbolic.Mul(cols.IsFirst, cols.IsInsert, symbolic.Sub(cols.LeafMinusNext, cols.LeafPlusIndex))
	am.comp.InsertGlobal(am.Round, am.qname("INSERT_POINTER_1"), expr1)

	// Check #2 for INSERT: IsFirst[i] * IsInsert[i] *(LeafPlusPrev[i] - LeafMinusIndex[i])
	expr2 := symbolic.Mul(cols.IsFirst, cols.IsInsert, symbolic.Sub(cols.LeafPlusPrev, cols.LeafMinusIndex))
	am.comp.InsertGlobal(am.Round, am.qname("INSERT_POINTER_2"), expr2)

	// Check #1 for DELETE: IsFirst[i] * IsDelete[i] * (LeafMinusNext[i] - LeafDeletedIndex[i])
	expr3 := symbolic.Mul(cols.IsFirst, cols.IsDelete, symbolic.Sub(cols.LeafMinusNext, cols.LeafDeletedIndex))
	am.comp.InsertGlobal(am.Round, am.qname("DELETE_POINTER_1"), expr3)

	// Check #2 for DELETE: IsFirst[i] * IsDelete[i] * (LeafDeletedPrev[i] - LeafMinusIndex[i])
	expr4 := symbolic.Mul(cols.IsFirst, cols.IsDelete, symbolic.Sub(cols.LeafDeletedPrev, cols.LeafMinusIndex))
	am.comp.InsertGlobal(am.Round, am.qname("DELETE_POINTER_2"), expr4)

	// Check #3 for DELETE: IsFirst[i] * IsDelete[i] * (LeafDeletedNext[i] - LeafPlusIndex[i])
	expr5 := symbolic.Mul(cols.IsFirst, cols.IsDelete, symbolic.Sub(cols.LeafDeletedNext, cols.LeafPlusIndex))
	am.comp.InsertGlobal(am.Round, am.qname("DELETE_POINTER_3"), expr5)

	// Check #4 for DELETE: IsFirst[i] * IsDelete[i] * (LeafPlusPrev[i] - LeafDeletedIndex[i])
	expr6 := symbolic.Mul(cols.IsFirst, cols.IsDelete, symbolic.Sub(cols.LeafPlusPrev, cols.LeafDeletedIndex))
	am.comp.InsertGlobal(am.Round, am.qname("DELETE_POINTER_4"), expr6)

	// Check #1 for READ-ZERO: IsFirst[i] * IsReadZero[i] * (LeafMinusNext[i] - LeafPlusIndex[i])
	expr7 := symbolic.Mul(cols.IsFirst, cols.IsReadZero, symbolic.Sub(cols.LeafMinusNext, cols.LeafPlusIndex))
	am.comp.InsertGlobal(am.Round, am.qname("READ_ZERO_POINTER_1"), expr7)

	// Check #2 for READ-ZERO: IsFirst[i] * IsReadZero[i] * (LeafPlusPrev[i] - LeafMinusIndex[i])
	expr8 := symbolic.Mul(cols.IsFirst, cols.IsReadZero, symbolic.Sub(cols.LeafPlusPrev, cols.LeafMinusIndex))
	am.comp.InsertGlobal(am.Round, am.qname("READ_ZERO_POINTER_2"), expr8)
}

func (am *Accumulator) checkLeafHashes() {
	cols := am.Cols
	am.comp.InsertMiMC(am.Round, am.qname("MIMC_PREV"), cols.LeafOpenings[0], cols.Zero, cols.Interm[0])
	am.comp.InsertMiMC(am.Round, am.qname("MIMC_NEXT"), cols.LeafOpenings[1], cols.Interm[0], cols.Interm[1])
	am.comp.InsertMiMC(am.Round, am.qname("MIMC_HKEY"), cols.LeafOpenings[2], cols.Interm[1], cols.Interm[2])
	am.comp.InsertMiMC(am.Round, am.qname("MIMC_HVAL_LEAF"), cols.LeafOpenings[3], cols.Interm[2], cols.LeafHashes)
	// Global: IsActive[i] * (1 - IsEmptyLeaf[i]) * (Leaves[i] - LeafHashes[i])
	expr1 := symbolic.Sub(cols.Leaves, cols.LeafHashes)
	expr2 := symbolic.Sub(symbolic.NewConstant(1), cols.IsEmptyLeaf)
	expr3 := symbolic.Mul(cols.IsActiveAccumulator, expr1, expr2)
	am.comp.InsertGlobal(am.Round, am.qname("LEAF_HASH_EQUALITY"), expr3)
	// Booleaninty of IsEmptyLeaf: IsActive[i] * (IsEmptyLeaf^2[i] - IsEmptyLeaf[i])
	expr4 := symbolic.Sub(symbolic.Square(cols.IsEmptyLeaf), cols.IsEmptyLeaf)
	expr4 = symbolic.Mul(expr4, cols.IsActiveAccumulator)
	am.comp.InsertGlobal(am.Round, am.qname("IS_EMPTY_LEAF_BOOLEANITY"), expr4)
}

func (am *Accumulator) checkNextFreeNode() {
	cols := am.Cols
	/*
		IsActive[i] * (1 - IsFirst[i]) * (
		IsInsertRow3[i] * (NextFreeNode[i] - NextFreeNode[i-1] - 1)
		+ (1- IsInsertRow3[i]) * (NextFreeNode[i] - NextFreeNode[i-1])
		)
	*/
	expr1 := symbolic.Mul(cols.IsInsertRow3,
		symbolic.Sub(cols.NextFreeNode, ifaces.ColumnAsVariable(column.Shift(cols.NextFreeNode, -1)), symbolic.NewConstant(1)))
	expr2 := symbolic.Mul(symbolic.Sub(symbolic.NewConstant(1), cols.IsInsertRow3),
		symbolic.Sub(cols.NextFreeNode, ifaces.ColumnAsVariable(column.Shift(cols.NextFreeNode, -1))))
	expr3 := symbolic.Mul(cols.IsActiveAccumulator,
		symbolic.Sub(symbolic.NewConstant(1), cols.IsFirst),
		symbolic.Add(expr1, expr2))
	am.comp.InsertGlobal(am.Round, am.qname("NEXT_FREE_NODE_CONSISTENCY_1"), expr3)

	// IsActive[i] * (1 - IsFirst[i]) * IsInsertRow3[i] * (NextFreeNode[i] - InsertionPath[i] - 1)
	expr4 := symbolic.Mul(cols.IsActiveAccumulator,
		symbolic.Sub(symbolic.NewConstant(1), cols.IsFirst),
		cols.IsInsertRow3,
		symbolic.Sub(cols.NextFreeNode, cols.InsertionPath, symbolic.NewConstant(1)))
	am.comp.InsertGlobal(am.Round, am.qname("NEXT_FREE_NODE_CONSISTENCY_2"), expr4)
}

// Function returning a query name
func (am *Accumulator) qname(name string, args ...any) ifaces.QueryID {
	return ifaces.QueryIDf("%v_%v", am.Name, am.comp.SelfRecursionCount) + "_" + ifaces.QueryIDf(name, args...)
}
