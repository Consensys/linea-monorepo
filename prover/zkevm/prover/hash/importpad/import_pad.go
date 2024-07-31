package importpad

import (
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/dedicated/projection"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	sym "github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/common"
	cs "github.com/consensys/linea-monorepo/prover/zkevm/prover/common/common_constraints"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/generic"
)

// ImportAndPadInputs collect the inputs of the [ImportAndPad] function.
type ImportAndPadInputs struct {
	// Name is a string identifier used to derive a unique name of each generated
	// column/constraint.
	Name string
	// Src is the list of [generic.GenericByteModule] to import
	Src generic.GenericByteModule
	// PaddingStrategy allows the caller specifying for which use-case import
	// and pad is used. This tells the [ImportAndPad] function which padding
	// strategy to use.
	PaddingStrategy generic.HashingUsecase
}

// importation stores the wizard compilation context to instantiate the
// functionality of [ImportAndPad]: e.g. it stores all the intermediate columns
// and constraints and implements the [wizard.ProverAction] interface. The
// [importation.Run] function is responsible for assigning all the generated
// columns.
type importation struct {

	// Inputs tracks the input structure used for instantiating this [importation]
	Inputs         ImportAndPadInputs
	HashNum        ifaces.Column // identifier for the hash the current limb belongs to
	Limbs          ifaces.Column // limbs declared by the current row
	NBytes         ifaces.Column // number of bytes in the current limbs
	Index          ifaces.Column // identifier for the limbs within the current hash
	IsInserted     ifaces.Column // indicates whether the current limbs was imported
	IsPadded       ifaces.Column // indicates whether the current limbs is from padding.
	AccPaddedBytes ifaces.Column // counts the number of padded bytes. This is then looked up to ensure that we don't pad more than 1 block.
	IsNewHash      ifaces.Column // indicates that a new hash stars at the current row

	// indicates whether the current row is active (or if it's just a filling row)
	IsActive ifaces.Column

	// Padder stores the padding-strategy-specific
	padder padder
}

// importationAssignmentBuilder is a utility struct used to build an assignment
// for the importation module. It is an internal of the package and is called
// in the [importation.Run] function.
type importationAssignmentBuilder struct {
	HashNum        *common.VectorBuilder
	Limbs          *common.VectorBuilder
	NBytes         *common.VectorBuilder
	Index          *common.VectorBuilder
	IsInserted     *common.VectorBuilder
	IsPadded       *common.VectorBuilder
	AccPaddedBytes *common.VectorBuilder
	IsActive       *common.VectorBuilder
	IsNewHash      *common.VectorBuilder
	Padder         padderAssignmentBuilder
}

type padder interface {
	pushPaddingRows(byteStringSize int, ab *importationAssignmentBuilder)
}

type padderAssignmentBuilder interface {
	pushInsertingRow(nbBytes int, isNewHash bool)
	padAndAssign(run *wizard.ProverRuntime)
}

// ImportAndPad defines and constrains the importation and the padding of a
// group of generic byte module following a prespecified padding strategy.
func ImportAndPad(comp *wizard.CompiledIOP, inp ImportAndPadInputs, numRows int) *importation {

	var (
		res = &importation{
			Inputs:         inp,
			HashNum:        comp.InsertCommit(0, ifaces.ColIDf("%v_IMPORT_PAD_HASH_NUM", inp.Name), numRows),
			Limbs:          comp.InsertCommit(0, ifaces.ColIDf("%v_IMPORT_PAD_LIMBS", inp.Name), numRows),
			NBytes:         comp.InsertCommit(0, ifaces.ColIDf("%v_IMPORT_PAD_NBYTES", inp.Name), numRows),
			Index:          comp.InsertCommit(0, ifaces.ColIDf("%v_IMPORT_PAD_INDEX", inp.Name), numRows),
			IsInserted:     comp.InsertCommit(0, ifaces.ColIDf("%v_IMPORT_PAD_IS_INSERTED", inp.Name), numRows),
			IsPadded:       comp.InsertCommit(0, ifaces.ColIDf("%v_IMPORT_PAD_IS_PADDED", inp.Name), numRows),
			AccPaddedBytes: comp.InsertCommit(0, ifaces.ColIDf("%v_IMPORT_PAD_ACC_PADDED_BYTES", inp.Name), numRows),
			IsActive:       comp.InsertCommit(0, ifaces.ColIDf("%v_IMPORT_PAD_IS_ACTIVE", inp.Name), numRows),
			IsNewHash:      comp.InsertCommit(0, ifaces.ColIDf("%v_IMPORT_PAD_IS_NEW_HASH", inp.Name), numRows),
		}
	)

	switch {
	case inp.PaddingStrategy == generic.KeccakUsecase:
		res.padder = res.newKeccakPadder(comp)
	case inp.PaddingStrategy == generic.Sha2Usecase:
		res.padder = res.newSha2Padder(comp)
	case inp.PaddingStrategy == generic.MiMCUsecase:
		res.padder = res.newMimcPadder(comp)
	default:
		panic("unknown strategy")
	}

	cs.MustBeBinary(comp, res.IsNewHash)
	cs.MustBeActivationColumns(comp, res.IsActive)
	cs.MustBeMutuallyExclusiveBinaryFlags(comp, res.IsActive, []ifaces.Column{
		res.IsInserted, res.IsPadded,
	})

	cs.MustZeroWhenInactive(comp, res.IsActive,
		res.HashNum,
		res.Limbs,
		res.NBytes,
		res.Index,
		res.AccPaddedBytes,
		res.IsNewHash,
	)

	// When the index flag must keep increasing during the padding zone
	comp.InsertGlobal(0,
		ifaces.QueryIDf("%v_INDEX_INCREASES", inp.Name),
		sym.Mul(
			res.IsPadded,
			sym.Sub(res.Index, column.Shift(res.Index, -1), 1),
		),
	)

	// IsNewHash is correctly set after the IsPadded IsInserted structure:
	// IsInserted[i] * IsPadded[i-1] == IsNewHash
	comp.InsertGlobal(0,
		ifaces.QueryIDf("%v_IS_NEW_HASH_WELL_SET", inp.Name),
		sym.Sub(res.IsNewHash, sym.Mul(res.IsInserted, column.Shift(res.IsPadded, -1))),
	)

	// This addresses the boundary case of the above column
	comp.InsertLocal(
		0,
		ifaces.QueryIDf("%v_IS_NEW_HASH_WELL_SET_BEGIN", inp.Name),
		sym.Sub(res.IsNewHash, 1),
	)

	// When IsNewHash is 1, then the index must be zero
	comp.InsertGlobal(
		0,
		ifaces.QueryIDf("%v_INDEX_MUST_RESTART_AT_0", inp.Name),
		sym.Mul(res.IsNewHash, res.Index),
	)

	// Ensures that AccPaddedBytes is well-set
	comp.InsertGlobal(
		0,
		ifaces.QueryIDf("%v_ACC_PADDED_BYTES_CORRECTNESS", inp.Name),
		sym.Sub(
			res.AccPaddedBytes,
			sym.Mul(res.IsPadded, sym.Add(column.Shift(res.AccPaddedBytes, -1), res.NBytes)),
		),
	)

	// Ensures the padding bytes are capped at 16 and cannot be zero.
	comp.InsertInclusionConditionalOnIncluded(0,
		ifaces.QueryIDf("%v_PADDING_AT_MOST_16_BYTES", inp.Name),
		[]ifaces.Column{getLookupForSize(comp, 16)},
		[]ifaces.Column{res.NBytes},
		res.IsPadded,
	)

	projection.InsertProjection(
		comp,
		ifaces.QueryIDf("%v_IMPORT_PAD_PROJECTION", inp.Name),
		[]ifaces.Column{inp.Src.Data.HashNum, inp.Src.Data.Limb, inp.Src.Data.NBytes, inp.Src.Data.Index},
		[]ifaces.Column{res.HashNum, res.Limbs, res.NBytes, res.Index},
		inp.Src.Data.TO_HASH,
		res.IsInserted,
	)

	return res
}

// Run performs the assignment of the importation module.
func (imp *importation) Run(run *wizard.ProverRuntime) {

	var (
		srcData = imp.Inputs.Src.Data
		hashNum = srcData.HashNum.GetColAssignment(run).IntoRegVecSaveAlloc()
		limbs   = srcData.Limb.GetColAssignment(run).IntoRegVecSaveAlloc()
		nBytes  = srcData.NBytes.GetColAssignment(run).IntoRegVecSaveAlloc()
		index   = srcData.Index.GetColAssignment(run).IntoRegVecSaveAlloc()
		toHash  = srcData.TO_HASH.GetColAssignment(run).IntoRegVecSaveAlloc()

		iab = importationAssignmentBuilder{
			HashNum:        common.NewVectorBuilder(imp.HashNum),
			Limbs:          common.NewVectorBuilder(imp.Limbs),
			NBytes:         common.NewVectorBuilder(imp.NBytes),
			Index:          common.NewVectorBuilder(imp.Index),
			IsInserted:     common.NewVectorBuilder(imp.IsInserted),
			IsPadded:       common.NewVectorBuilder(imp.IsPadded),
			AccPaddedBytes: common.NewVectorBuilder(imp.AccPaddedBytes),
			IsActive:       common.NewVectorBuilder(imp.IsActive),
			IsNewHash:      common.NewVectorBuilder(imp.IsNewHash),
		}

		currByteSize = 0
		currHashNum  field.Element
	)

	switch {
	case imp.Inputs.PaddingStrategy == generic.KeccakUsecase:
		iab.Padder = keccakPadderAssignmentBuilder{}
	case imp.Inputs.PaddingStrategy == generic.Sha2Usecase:
		iab.Padder = &sha2PaddingAssignmentBuilder{
			AccInsertedBytes: common.NewVectorBuilder(imp.padder.(*sha2Padder).AccInsertedBytes),
		}
	case imp.Inputs.PaddingStrategy == generic.MiMCUsecase:
		iab.Padder = &mimcPadderAssignmentBuilder{}
	default:
		panic("unknown strategy")
	}

	for i := range hashNum {

		if toHash[i].IsZero() {
			if i == len(hashNum)-1 {
				imp.padder.pushPaddingRows(currByteSize, &iab)
			}

			continue
		}

		if i > 0 && currHashNum != hashNum[i] && !currHashNum.IsZero() {
			imp.padder.pushPaddingRows(currByteSize, &iab)
		}

		if currHashNum != hashNum[i] {
			currHashNum = hashNum[i]
			currByteSize = 0
			iab.IsNewHash.PushOne()
		} else {
			iab.IsNewHash.PushZero()
		}

		var (
			indexInt  = int(index[i].Uint64())
			nBytesInt = int(nBytes[i].Uint64())
		)

		currByteSize += nBytesInt

		iab.pushInsertion(hashNum[i], limbs[i], nBytesInt, indexInt)

		if i == len(hashNum)-1 {
			imp.padder.pushPaddingRows(currByteSize, &iab)
		}
	}

	iab.HashNum.PadAndAssign(run, field.Zero())
	iab.Limbs.PadAndAssign(run, field.Zero())
	iab.NBytes.PadAndAssign(run, field.Zero())
	iab.Index.PadAndAssign(run, field.Zero())
	iab.IsInserted.PadAndAssign(run, field.Zero())
	iab.IsPadded.PadAndAssign(run, field.Zero())
	iab.AccPaddedBytes.PadAndAssign(run, field.Zero())
	iab.IsActive.PadAndAssign(run, field.Zero())
	iab.IsNewHash.PadAndAssign(run, field.Zero())
	iab.Padder.padAndAssign(run)
}

// pushPaddingCommonColumns push an insertion row corresponding to the first
// row of a new hash.
func (iab *importationAssignmentBuilder) pushInsertion(hashNum field.Element, limb field.Element, nbBytes int, index int) {

	iab.HashNum.PushField(hashNum)
	iab.Limbs.PushField(limb)
	iab.NBytes.PushInt(nbBytes)
	iab.Index.PushInt(index)
	iab.AccPaddedBytes.PushZero()
	iab.IsActive.PushOne()
	iab.IsInserted.PushOne()
	iab.IsPadded.PushZero()
	iab.Padder.pushInsertingRow(nbBytes, index == 0)
}

// pushPaddingCommonColumns adds pushes a padding rows for the columns that are
// independant of the padding strategy
func (ipad *importationAssignmentBuilder) pushPaddingCommonColumns() {

	ipad.HashNum.RepushLast()
	ipad.Index.PushInc()
	ipad.IsActive.PushOne()
	ipad.IsInserted.PushZero()
	ipad.IsPadded.PushOne()
	ipad.IsNewHash.PushZero()
}
