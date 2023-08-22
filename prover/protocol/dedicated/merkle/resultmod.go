package merkle

import (
	"strings"

	"github.com/consensys/accelerated-crypto-monorepo/maths/common/smartvectors"
	"github.com/consensys/accelerated-crypto-monorepo/maths/common/vector"
	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/ifaces"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/wizard"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
)

// Module summarizing the Merkle proof claims
type ResultMod struct {

	// the compiled IOP
	comp *wizard.CompiledIOP

	// Number of rows in the module
	NumRows int
	// Number of proofs
	NumProofs int
	// Name if the name of parent context joined with a specifier.
	Name string
	// Round of the module
	Round int

	// Leaf contains the alleged leaves
	Leaf ifaces.Column
	// Roots contains the Merkle roots
	Roots ifaces.Column
	// Pos contains the positions of the alleged leaves
	Pos ifaces.Column
	// One is a dummy column containing only ones
	// Use for looking up and selecting only the
	// the columns containing the root in the ComputeMod
	IsActive ifaces.Column
}

// Registers all the columns also assumes that Leaf, Roots and Pos have been
// passed already.
func (rm *ResultMod) Define(comp *wizard.CompiledIOP, round int, name string, numProofs int) {

	// Sanity check that the columns have been passed
	if rm.Roots == nil || rm.Pos == nil || rm.Leaf == nil {
		panic("please set all the required columns before calling define")
	}

	// Sanity check that they all have the same size
	if rm.Roots.Size() != rm.Pos.Size() || rm.Roots.Size() != rm.Leaf.Size() {
		utils.Panic("the sizes of the passed columns should be consistent %v, %v, and %v", rm.Roots.Size(), rm.Pos.Size(), rm.Leaf.Size())
	}

	// Sanity-check the value of NumRows
	numRows := rm.Roots.Size()
	if numRows != utils.NextPowerOfTwo(numProofs) {
		utils.Panic("numRows %v but numProofs %v", numRows, numProofs)
	}

	// Assigns the depth and the number of proofs
	rm.comp = comp
	rm.NumRows = numRows
	rm.NumProofs = numProofs
	rm.Round = round
	rm.Name = strings.Join([]string{"MERKLE", "RESULTMOD", name}, "_")

	// Defines the zero-th column
	rm.IsActive = rm.comp.InsertPrecomputed(
		rm.colname("ISACTIVE"),
		smartvectors.RightZeroPadded(
			vector.Repeat(field.One(), rm.NumProofs),
			rm.NumRows,
		),
	)

}

func (rm *ResultMod) colname(name string, args ...any) ifaces.ColID {
	return ifaces.ColIDf("MERKLE_%v_%v", rm.Name, rm.comp.SelfRecursionCount) + "_" + ifaces.ColIDf(name, args...)
}
