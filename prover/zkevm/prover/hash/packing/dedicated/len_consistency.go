package dedicated

import (
	"slices"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/dedicated/byte32cmp"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	sym "github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/common"
	commonconstraints "github.com/consensys/linea-monorepo/prover/zkevm/prover/common/common_constraints"
)

// LcInputs stores the inputs for [LengthConsistency] function.
type LcInputs struct {
	// input tables
	Table []ifaces.Column
	// the claimed lengths for the elements embedded inside Table.
	TableLen []ifaces.Column
	// max length in bytes.
	MaxLen int
	// name of the table
	Name string
}

// lengthConsistency stores the intermediate columns for [LengthConsistency] function.
type lengthConsistency struct {
	inp LcInputs
	// decomposition of Table into bytes
	bytes []byte32cmp.LimbColumns
	// the ProverAction for decomposition of Tables into bytes.
	pa []wizard.ProverAction
	// a binary column indicating if the byte is non-empty
	bytesLen [][]ifaces.Column
	// size of the columns
	size int
}

// LengthConsistency receives a table and the associated lengths,
//
//	and assert that the elements in the table of are the given lengths.
func LengthConsistency(comp *wizard.CompiledIOP, inp LcInputs) *lengthConsistency {

	var (
		name      = inp.Name
		numCol    = len(inp.Table)
		size      = inp.Table[0].Size()
		numBytes  = inp.MaxLen
		createCol = common.CreateColFn(comp, "LENGTH_CONSISTENCY_"+name, size)
	)

	res := &lengthConsistency{
		bytesLen: make([][]ifaces.Column, numCol),
		inp:      inp,
		size:     size,
	}

	for j := 0; j < numCol; j++ {
		res.bytesLen[j] = make([]ifaces.Column, numBytes)
		for k := range res.bytesLen[0] {
			res.bytesLen[j][k] = createCol("BYTE_LEN_%v_%v", j, k)
		}
	}

	// 1. decompose each column of the table to bytes
	// 2. check that number of bytes == tableLen
	res.bytes = make([]byte32cmp.LimbColumns, numCol)
	res.pa = make([]wizard.ProverAction, numCol)
	for j := range inp.Table {
		// 	// constraint asserting to the correct decomposition of table to bytes
		res.bytes[j], res.pa[j] = byte32cmp.Decompose(comp, inp.Table[j], numBytes, 8, res.bytesLen[j]...)
	}

	// claimed lengths for the table are correct;
	//   - bytesLen is binary
	//   - bytesLen over a row adds up to tableLen
	for j := range inp.Table {
		sum := sym.NewConstant(0)
		for k := 0; k < numBytes; k++ {
			sum = sym.Add(sum, res.bytesLen[j][k])

			// bytesLen is binary
			commonconstraints.MustBeBinary(comp, res.bytesLen[j][k])
		}
		comp.InsertGlobal(0, ifaces.QueryIDf("%v_CLDLen_%v", name, j), sym.Sub(sum, inp.TableLen[j]))
	}
	return res
}

func (lc *lengthConsistency) Run(run *wizard.ProverRuntime) {
	var (
		numCol   = len(lc.inp.Table)
		numBytes = lc.inp.MaxLen
	)

	for j := range lc.inp.Table {
		lc.pa[j].Run(run)
	}

	tableLen := make([]smartvectors.SmartVector, numCol)
	bytesLen := make([][]*common.VectorBuilder, numCol)

	// allocate bytesLen
	for j := 0; j < numCol; j++ {
		bytesLen[j] = make([]*common.VectorBuilder, numBytes)
		for k := range bytesLen[0] {
			bytesLen[j][k] = common.NewVectorBuilder(lc.bytesLen[j][k])
		}
	}

	// populate bytesLen
	for j := 0; j < numCol; j++ {
		tableLen[j] = lc.inp.TableLen[j].GetColAssignment(run)

		for row := 0; row < tableLen[0].Len(); row++ {
			dec := getZeroOnes(tableLen[j].Get(row), numBytes)
			//  this is used in bytes32cmp.Decompose() which needs little-endian
			slices.Reverse(dec)

			for k := range dec {
				bytesLen[j][k].PushField(dec[k])
			}
		}
	}
	for j := range tableLen {
		for k := range bytesLen[0] {
			bytesLen[j][k].PadAndAssign(run)
		}
	}
}

// getZeroOnes receives n  and outputs the pattern  (0,..0,1,..,1) such that there are n elements 1.
func getZeroOnes(n field.Element, max int) (a []field.Element) {
	_n := field.ToInt(&n)
	if _n > max {
		utils.Panic("%v should be smaller than %v", _n, max)
	}
	for j := 0; j < max-_n; j++ {
		a = append(a, field.Zero())

	}
	for i := max - _n; i < max; i++ {
		a = append(a, field.One())

	}

	return a

}
