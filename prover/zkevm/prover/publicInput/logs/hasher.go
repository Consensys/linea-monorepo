package logs

import (
	"fmt"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/distributed/pragmas"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	sym "github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/common"
	commonconstraints "github.com/consensys/linea-monorepo/prover/zkevm/prover/common/common_constraints"
	util "github.com/consensys/linea-monorepo/prover/zkevm/prover/publicInput/utilities"
)

// LogHasher is used to MiMC-hash the data in LogMessages. Using a zero initial stata,
// the data in L2L1 logs must be hashed as follows:
// msg1HashHi, msg1HashLo, msg2HashHi, msg2HashLo, and so on.
// The hashing proceeds row by row. For row i, after hashing msgHashHi[i], the result is put into hashFirst[i].
// hashFirst[i] is then used as the state to hash msgHashLo[i], and the result is put into hashSecond[i]
// Finally, hashSecond[i] is used as the initial state for hashing msgHashHi[i+1]
// Here is a diagram of how the values are computed
// state:          0            x1           x1'       x2
// block:       msg1Hi         msg1Lo      msg2Hi    msg2Lo     ....
// MimcOutput     x1            x1'          x2        x2'
// x1, x2 will correspond to the values stored in hashFirst
// x1', x2' will correspond to values stored in hashSecond
// The final value of the chained hash can be retrieved as ---> hashSecond[ctMax[any index]]
type LogHasher struct {
	// the Hash value after each step, as explained in the description of LogHasher
	Hash [common.NbLimbU256]ifaces.Column
	// L2L1 logs: Inter is a shifted version of hashSecond, necessary due to how the MiMC constraints operate
	Inter [common.NbLimbU256]ifaces.Column
	// the relevant value of the hash (the last value when isActive ends)
	HashFinal [common.NbLimbU256]ifaces.Column
}

// NewLogHasher returns a new LogHasher with initialized columns that are not constrained.
func NewLogHasher(comp *wizard.CompiledIOP, size int, name string) LogHasher {
	var res LogHasher

	for i := range res.Hash {
		res.Hash[i] = util.CreateCol(name, fmt.Sprintf("HASH_%d", i), size, comp)
		res.Inter[i] = util.CreateCol(name, fmt.Sprintf("INTER_%d", i), size, comp)
		res.HashFinal[i] = util.CreateCol(name, fmt.Sprintf("HASH_FINAL_%d", i), size, comp)
	}

	return res
}

// DefineHasher specifies the constraints of the LogHasher with respect to the ExtractedData fetched from the arithmetization
func DefineHasher(comp *wizard.CompiledIOP, hasher LogHasher, name string, fetched ExtractedData) {

	// Needed for the limitless prover to understand that the columns are
	// not just empty with just padding and suboptimal representation.
	pragmas.MarkRightPadded(hasher.Inter)
	pragmas.MarkRightPadded(hasher.HashFirst)
	pragmas.MarkRightPadded(hasher.HashSecond)

	panic("replace by poseidon query")

	// MiMC constraints
	comp.InsertMiMC(0, ifaces.QueryIDf("%s_%s", name, "MIMC_CONSTRAINT"), fetched.Hi, hasher.Inter, hasher.HashFirst, fetched.FilterFetched)
	comp.InsertMiMC(0, ifaces.QueryIDf("%s_%s", name, "MIMC_CONSTRAINT_SECOND"), fetched.Lo, hasher.HashFirst, hasher.HashSecond, fetched.FilterFetched)

	for i := range hasher.Hash {
		// intermediary state integrity
		comp.InsertGlobal(0, ifaces.QueryIDf("%s_%s", name, "CONSISTENCY_INTER_AND_HASH_LAST"), // LAST is either hashSecond or hashThird
			sym.Mul(
				sym.Sub(column.Shift(hasher.HashSecond, -1), hasher.Inter),
				fetched.FilterFetched,
			),
		)

		// inter, the old state column, is initially zero
		comp.InsertLocal(0, ifaces.QueryIDf("%s_INTER_LOCAL_%d", name, i), ifaces.ColumnAsVariable(hasher.Inter[i]))

		// constrain HashFinal
		commonconstraints.MustBeConstant(comp, hasher.HashFinal[i])
		util.CheckLastELemConsistency(comp, fetched.FilterFetched, hasher.Hash[i], hasher.HashFinal[i], name)
	}
}

// AssignHasher assigns the data in the LogHasher using the ExtractedData fetched from the arithmetization
func AssignHasher(run *wizard.ProverRuntime, hasher LogHasher, fetched ExtractedData) {
	size := fetched.Data[0].Size()
	var hash, inter [common.NbLimbU256][]field.Element
	for i := range hash {
		hash[i] = make([]field.Element, size)
		inter[i] = make([]field.Element, size)
	}

	var fetchedData, hashFinal [common.NbLimbU256]field.Element
	// the initial state is zero
	state := make([]field.Element, common.NbLimbU256)
	for i := range state {
		state[i].SetZero()
	}

	for i := 0; i < size; i++ {
		for j := range fetchedData {
			fetchedData[j] = fetched.Data[j].GetColAssignmentAt(run, i)
		}

		state = common.BlockCompression(state, fetchedData[:])
		for j := range hash {
			hash[j][i] = state[j]
		}

		// the data in hashSecond is used to initialize the next initial state, stored in the inter column
		if i+1 < size {
			for j := range inter {
				inter[j][i+1] = state[j]
			}
		}

		isActive := fetched.FilterFetched.GetColAssignmentAt(run, i)
		// continuously update HashFinal
		if isActive.IsOne() {
			for j := range hashFinal {
				hashFinal[j] = state[j]
			}
		}
	}

	// assign the hasher columns
	for i := range hasher.Hash {
		run.AssignColumn(hasher.Hash[i].GetColID(), smartvectors.NewRegular(hash[i]))
		run.AssignColumn(hasher.Inter[i].GetColID(), smartvectors.NewRegular(inter[i]))
		run.AssignColumn(hasher.HashFinal[i].GetColID(), smartvectors.NewConstant(hashFinal[i], size))
	}
}
