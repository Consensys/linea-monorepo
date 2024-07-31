package importpad

import (
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	sym "github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/generic"
)

// mimcPadder implements the [padder] interface and zero-pads the input streams.
// The padding is not resistant to padding-attavcks so it should be used very
// carefully: either in situations where the length of the message is encoded
// within the message itself or in situation where the length of the message is
// available through other means.
type mimcPadder struct{}

type mimcPadderAssignmentBuilder struct{}

// newMimcPadder creates the constraints ensuring that the zero-padding and
// returns an object helping with the assignment.
func (ipad *importation) newMimcPadder(comp *wizard.CompiledIOP) padder {

	// The padding values are all zero
	comp.InsertGlobal(0,
		ifaces.QueryIDf("%v_MIMC_PADDING_VALUES_ARE_ZERO", ipad.Inputs.Name),
		sym.Mul(ipad.IsPadded, ipad.Limbs),
	)

	// This check that we do not pad by more than a whole block. It does not check
	// that this does exactly the right number of padded bytes. This is OK since
	// later stages of the packing check that the padded byte string has a length
	// divisible by the blocksize.
	//
	// The +8 is to account for the last 64 bits string length. In theory it
	// should be restricted to only 9..72 but the above constraints ensure that
	// that at least the first and the last limbs are included.
	comp.InsertInclusionConditionalOnIncluded(0,
		ifaces.QueryIDf("%v_LOOKUP_NB_PADDED_BYTES", ipad.Inputs.Name),
		[]ifaces.Column{getLookupForSize(comp, generic.MiMCUsecase.BlockSizeBytes()-1)},
		[]ifaces.Column{ipad.AccPaddedBytes},
		ipad.IsPadded,
	)

	return mimcPadder{}
}

func (sp mimcPadder) pushPaddingRows(byteStringSize int, ipad *importationAssignmentBuilder) {

	var (
		blocksize      = generic.MiMCUsecase.BlockSizeBytes()
		remainToPad    = blocksize - (byteStringSize % blocksize)
		accPaddedBytes = 0
	)

	for remainToPad > 0 {
		currNbBytes := utils.Min(remainToPad, 16)
		accPaddedBytes += currNbBytes
		remainToPad -= currNbBytes

		ipad.pushPaddingCommonColumns()
		ipad.Limbs.PushZero()
		ipad.NBytes.PushInt(currNbBytes)
		ipad.AccPaddedBytes.PushInt(accPaddedBytes)
	}
}

func (kp mimcPadderAssignmentBuilder) pushInsertingRow(int, bool) {}

func (kp mimcPadderAssignmentBuilder) padAndAssign(*wizard.ProverRuntime) {}
