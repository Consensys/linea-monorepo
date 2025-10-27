package protocols

import (
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	sym "github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/keccak/keccakf"
)

// LinearCombination represents the linear combination of the given columns over the powers of the given scalar.
// it is technically a polynomial evaluation at the given scalar point.
type LinearCombination struct {
	// scalar for linear combination
	scalar int
	// input columns
	cols []ifaces.Column
	// output column
	CombinationRes ifaces.Column
	// size of columns
	size int
}

// LinearCombination is similar to polynomial evaluation, implementing [wizard.ProverAction].
func NewLinearCombination(comp *wizard.CompiledIOP, name string, r []ifaces.Column, base int) *LinearCombination {
	var (
		size = r[0].Size()
		col  = comp.InsertCommit(0, ifaces.ColIDf("%v", name), size)
	)
	// .. using the Horner method
	s := sym.NewConstant(0)
	for i := len(r) - 1; i >= 0; i-- {
		s = sym.Mul(s, base)
		s = sym.Add(s, r[i])
	}

	comp.InsertGlobal(0, ifaces.QueryIDf("%v_QUERY", name),
		sym.Sub(col, s),
	)

	return &LinearCombination{
		scalar:         base,
		cols:           r,
		CombinationRes: col,
		size:           size,
	}
}

// Run  assign the values to the linear combination result column.
func (bc *LinearCombination) Run(run *wizard.ProverRuntime) {

	var (
		colValues = make([][]field.Element, len(bc.cols))

		ss     = make([]field.Element, bc.size)
		baseFr = field.NewElement(uint64(bc.scalar))
	)
	for i := 0; i < len(bc.cols); i++ {
		colValues[i] = bc.cols[i].GetColAssignment(run).IntoRegVecSaveAlloc()
	}
	for j := 0; j < bc.size; j++ {
		var t []field.Element
		for i := 0; i < len(bc.cols); i++ {
			t = append(t, colValues[i][j])
		}

		ss[j] = keccakf.BaseRecompose(t, &baseFr)
		res := DecomposeForTesting(ss[j].Uint64(), bc.scalar, len(bc.cols))

		for i := range res {
			if res[i] != t[i].Uint64() {
				utils.Panic("linear combination assignment failed at row %v: expected %v, got %v", j, t[i].Uint64(), res[i])
			}
		}

	}

	run.AssignColumn(bc.CombinationRes.GetColID(), smartvectors.NewRegular(ss))
}

func DecomposeForTesting(r uint64, base int, nb int) (res []uint64) {
	// It will essentially be used for chunk to slice decomposition
	res = make([]uint64, 0, nb)
	base64 := uint64(base)
	curr := r
	for curr > 0 {
		limb := curr % base64
		res = append(res, limb)
		curr /= base64
	}

	if len(res) > nb {
		utils.Panic("expected %v limbs, but got %v", nb, len(res))
	}

	if len(res) < nb {
		// Complete with zeroes
		for len(res) < nb {
			res = append(res, 0)
		}
	}
	return res
}
