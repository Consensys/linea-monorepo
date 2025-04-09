package logs

import (
	"github.com/consensys/linea-monorepo/prover/crypto/mimc"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	sym "github.com/consensys/linea-monorepo/prover/symbolic"
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
	// the hash value after each step, as explained in the description of LogHasher
	hashFirst, hashSecond ifaces.Column
	// L2L1 logs: inter is a shifted version of hashSecond, necessary due to how the MiMC constraints operate
	inter ifaces.Column
	// the relevant value of the hash (the last value when isActive ends)
	HashFinal ifaces.Column
}

// NewLogHasher returns a new LogHasher with initialized columns that are not constrained.
func NewLogHasher(comp *wizard.CompiledIOP, size int, name string) LogHasher {
	return LogHasher{
		hashFirst:  util.CreateCol(name, "HASH_FIRST", size, comp),
		hashSecond: util.CreateCol(name, "HASH_SECOND", size, comp),
		inter:      util.CreateCol(name, "INTER", size, comp),
		HashFinal:  util.CreateCol(name, "HASH_FINAL", size, comp),
	}
}

// DefineHasher specifies the constraints of the LogHasher with respect to the ExtractedData fetched from the arithmetization
func DefineHasher(comp *wizard.CompiledIOP, hasher LogHasher, name string, fetched ExtractedData) {

	// MiMC constraints
	comp.InsertMiMC(0, ifaces.QueryIDf("%s_%s", name, "MIMC_CONSTRAINT"), fetched.Hi, hasher.inter, hasher.hashFirst, nil)
	comp.InsertMiMC(0, ifaces.QueryIDf("%s_%s", name, "MIMC_CONSTRAINT_SECOND"), fetched.Lo, hasher.hashFirst, hasher.hashSecond, nil)

	// intermediary state integrity
	comp.InsertGlobal(0, ifaces.QueryIDf("%s_%s", name, "CONSISTENCY_INTER_AND_HASH_LAST"), // LAST is either hashSecond or hashThird
		sym.Sub(hasher.hashSecond,
			column.Shift(hasher.inter, 1),
		),
	)

	// inter, the old state column, is initially zero
	comp.InsertLocal(0, ifaces.QueryIDf("%s_%s", name, "INTER_LOCAL"), ifaces.ColumnAsVariable(hasher.inter))

	// constrain HashFinal
	commonconstraints.MustBeConstant(comp, hasher.HashFinal)
	util.CheckLastELemConsistency(comp, fetched.filterFetched, hasher.hashSecond, hasher.HashFinal, name)
}

// AssignHasher assigns the data in the LogHasher using the ExtractedData fetched from the arithmetization
func AssignHasher(run *wizard.ProverRuntime, hasher LogHasher, fetched ExtractedData) {
	size := fetched.Hi.Size()
	hashFirst := make([]field.Element, size)
	hashSecond := make([]field.Element, size)
	inter := make([]field.Element, size)

	var (
		hashFinal field.Element
	)

	state := field.Zero() // the initial state is zero
	for i := 0; i < len(hashFirst); i++ {
		// first, hash the HI part of the fetched log message
		state = mimc.BlockCompression(state, fetched.Hi.GetColAssignmentAt(run, i))
		hashFirst[i].Set(&state)

		// secondly, hash the Lo part of the fetched log message
		state = mimc.BlockCompression(state, fetched.Lo.GetColAssignmentAt(run, i))
		hashSecond[i].Set(&state)

		// the data in hashSecond is used to initialize the next initial state, stored in the inter column
		if i+1 < len(hashFirst) {
			inter[i+1] = hashSecond[i]
		}

		isActive := fetched.filterFetched.GetColAssignmentAt(run, i)
		// continuously update HashFinal
		if isActive.IsOne() {
			hashFinal.Set(&hashSecond[i])
		}
	}

	// assign the hasher columns
	run.AssignColumn(hasher.hashFirst.GetColID(), smartvectors.NewRegular(hashFirst))
	run.AssignColumn(hasher.hashSecond.GetColID(), smartvectors.NewRegular(hashSecond))
	run.AssignColumn(hasher.inter.GetColID(), smartvectors.NewRegular(inter))
	run.AssignColumn(hasher.HashFinal.GetColID(), smartvectors.NewConstant(hashFinal, size))
}
