package datatransfer

import (
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/dedicated/projection"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/datatransfer/dedicated"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/generic"
)

const (
	Keccak int = iota
	Sha2
)

// The struct importAndPadd presents the columns initiated by the subModule.
// It consists of the counterpart columns for the arithmetization columns, extended via padding.
// The data is extracted from the arithmetization columns, it is then padded if required.
// The result is embedded (preserving the order) in the columns of the module.
type importAndPadd struct {
	// counterparts for the arithmetization columns, extended via padding
	hashNum, limb, nByte, index, cleanLimb ifaces.Column
	// Indicates where the imported rows are inserted in the module
	isInserted ifaces.Column
	// Indicates where the padding values are added
	isPadded ifaces.Column
	// Indicated the active Rows of the module
	isActive ifaces.Column
	// It is 1 when a new hash is launched, otherwise it is zero
	isNewHash ifaces.Column
	// a column of all 1
	oneCol ifaces.Column

	// accumulator for padded nBytes
	accPaddedBytes ifaces.Column
}

/*
NewImportAndPadd builds an instance of importAndPadd.
It commits to the columns specific to the submodule and defines the constrains asserting to the following facts.

-  the correct extraction of the data from the arithmetization columns.

-  the correct padding of the limbs.

-  the correct insertion of the data to the columns of the module.

-  the correct form of the columns, for instance the binary constraints.
// for Keccak it uses hashType = 0 , for Sha2 hashType = 1.
*/
func (iPadd *importAndPadd) newImportAndPadd(comp *wizard.CompiledIOP,
	round, maxRows int,
	gbm generic.GenericByteModule, // arithmetization columns
	lu lookUpTables,
	hashType int,
) {
	// Declare the columns
	iPadd.insertCommit(comp, round, maxRows)

	// Declare the constraints

	// padding over the arithmetization columns (gbm columns) is done correctly
	switch hashType {
	case Keccak:
		iPadd.insertPaddingKeccak(comp, round, lu.colKeccakMaxPadding)
	case Sha2:
		iPadd.insertPaddingSha2(comp, round, lu.colSha2MaxPadding)
	default:
		utils.Panic("The hash type is not supported")
	}

	// projection query between gbm columns and module column;
	// asserting the rows of arithmetization columns are correctly projected over the module columns.
	data := gbm.Data
	projection.InsertProjection(comp, ifaces.QueryIDf("HashNum_OrderPreserving"),
		[]ifaces.Column{data.HashNum, data.Limb, data.NBytes},
		[]ifaces.Column{iPadd.hashNum, iPadd.limb, iPadd.nByte}, data.TO_HASH, iPadd.isInserted)

	// constraints on flag columns; isInserted,isPadded, isNewHash, isActive
	/*
		1.  they are all binary
		2.  isInserted and isPadded are partition of isActive
		.
		.
		.
	*/
	iPadd.csBinaryColumns(comp, round)
}

// InsertCommitToImportAndPadd commits to the columns initiated by the ImportAndPadd submodule.
func (iPadd *importAndPadd) insertCommit(comp *wizard.CompiledIOP, round, maxRows int) {
	iPadd.hashNum = comp.InsertCommit(round, deriveName("HashNum"), maxRows)
	iPadd.limb = comp.InsertCommit(round, deriveName("Limb"), maxRows)
	iPadd.cleanLimb = comp.InsertCommit(round, deriveName("CleanLimb"), maxRows)
	iPadd.nByte = comp.InsertCommit(round, deriveName("NByte"), maxRows)
	iPadd.index = comp.InsertCommit(round, deriveName("Index"), maxRows)
	iPadd.isInserted = comp.InsertCommit(round, deriveName("IsInserted"), maxRows)
	iPadd.isPadded = comp.InsertCommit(round, deriveName("IsPadded"), maxRows)
	iPadd.isNewHash = comp.InsertCommit(round, deriveName("IsNewHash"), maxRows)
	iPadd.isActive = comp.InsertCommit(round, deriveName("IsActive"), maxRows)
	iPadd.oneCol = comp.InsertCommit(round, deriveName("OneCol"), maxRows)
	iPadd.accPaddedBytes = comp.InsertCommit(round, ifaces.ColID("AccPaddedBytes"), maxRows)
}

// csBinaryColumns aims for imposing the constraints on the flag columns,
// isInserted, isImported,isPadded, isNewHash, isActive.
/*
	1.  they are all binary
	2.  isInserted and isPadded are partition of isActive
	3.  isPadded appears only before isNewHash
	4.  isNewhas has the right form
	5. isActive has the right form (starting with ones followed by zeroes, if required)
*/
func (iPadd importAndPadd) csBinaryColumns(comp *wizard.CompiledIOP, round int) {
	one := symbolic.NewConstant(1)
	isInserted := ifaces.ColumnAsVariable(iPadd.isInserted)
	isPadded := ifaces.ColumnAsVariable(iPadd.isPadded)
	isActive := ifaces.ColumnAsVariable(iPadd.isActive)
	isNewHash := ifaces.ColumnAsVariable(iPadd.isNewHash)

	// binary constraints
	comp.InsertGlobal(round, ifaces.QueryIDf("IsInserted_IsBinary"), isInserted.Mul(one.Sub(isInserted)))
	comp.InsertGlobal(round, ifaces.QueryIDf("IsPadded_IsBinary"), isPadded.Mul(one.Sub(isPadded)))
	comp.InsertGlobal(round, ifaces.QueryIDf("IsActive_IsBinary"), isActive.Mul(one.Sub(isActive)))
	comp.InsertGlobal(round, ifaces.QueryIDf("IsNewHash_IsBinary"), isNewHash.Mul(one.Sub(isNewHash)))

	// isActive is of the right form, starting with ones and all zeroes are at the end
	shiftIsActive := ifaces.ColumnAsVariable(column.Shift(iPadd.isActive, -1))
	comp.InsertGlobal(round, ifaces.QueryIDf("IsActive"), (shiftIsActive.Sub(isActive)).Mul(one.Sub(shiftIsActive.Sub(isActive))))

	// isInserted = (1- isPAdded) * isActive
	// isActive = 0 ---> isPadded = 0 , isInserted = 0
	comp.InsertGlobal(round, ifaces.QueryIDf("isInserted_isPadded"),
		isInserted.Sub((one.Sub(isPadded)).Mul(isActive)))
	comp.InsertGlobal(round, ifaces.QueryIDf("IsPadded_IsActive"), symbolic.Mul(symbolic.Sub(1, isActive), isPadded))
	comp.InsertGlobal(round, ifaces.QueryIDf("IsInserted_IsActive"), symbolic.Mul(symbolic.Sub(1, isActive), isInserted))

	// a sequence of isPadded =1 appears iff a newHash is launched.
	// if isPadded[i] = 1 ---> isPadded[i+1] = 1 or isNewHash[i+1] = 1
	// if isNewHas[i] = 1 ---> isPadded[i-1] = 1 and isPadded[i] = 0
	isPaddedNext := ifaces.ColumnAsVariable(column.Shift(iPadd.isPadded, 1))
	isPaddedPrev := ifaces.ColumnAsVariable(column.Shift(iPadd.isPadded, -1))
	isNewHashNext := ifaces.ColumnAsVariable(column.Shift(iPadd.isNewHash, 1))

	// to impose a bound for the global constraints
	isActiveShift := ifaces.ColumnAsVariable(column.Shift(iPadd.isActive, 1))
	expr1 := (isPadded.Mul(one.Sub(isPaddedNext)).Mul(one.Sub(isNewHashNext))).Mul(isActiveShift)
	expr2 := (isNewHash.Mul((one.Sub(isPaddedPrev)).Add(isPadded)))

	comp.InsertGlobal(round, ifaces.QueryIDf("isPadded_isNewHash1"), expr1)
	comp.InsertGlobal(round, ifaces.QueryIDf("isPadded_isNewHash2"), expr2)

	// constraints over isNewhash;
	// if index = 0 ---> isNewHash[i] = 1
	// otherwise ---> isNewHash = 0
	isZero := symbolic.Add(iPadd.isNewHash, symbolic.Sub(1, iPadd.isActive))
	dedicated.InsertIsTargetValue(comp, round, ifaces.QueryIDf("IsNewHash_IndexIsZero"), field.Zero(), iPadd.index, isZero)

	// if isActive = 0 ---> isNewHash = 0
	comp.InsertGlobal(round, ifaces.QueryIDf("IsNewHash_IsNotActive"),
		symbolic.Mul(symbolic.Sub(1, iPadd.isActive), iPadd.isNewHash))
}
