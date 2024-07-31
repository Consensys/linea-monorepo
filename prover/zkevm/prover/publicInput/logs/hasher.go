package logs

import (
	"github.com/consensys/zkevm-monorepo/prover/crypto/mimc"
	"github.com/consensys/zkevm-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/zkevm-monorepo/prover/maths/field"
	"github.com/consensys/zkevm-monorepo/prover/protocol/column"
	"github.com/consensys/zkevm-monorepo/prover/protocol/ifaces"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
	sym "github.com/consensys/zkevm-monorepo/prover/symbolic"
	publicInput "github.com/consensys/zkevm-monorepo/prover/zkevm/prover/publicInput/utilities"
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
	// counter increases 1 by 1 for each row of the table
	counter ifaces.Column
	// the ctMax (counter max) column contains the maximal counter for the entire table.
	// Or alternatively, it denotes the number of extracted data pairs (Hi, Lo) fetched from the arithmetization.
	ctMax                 ifaces.Column
	hashFirst, hashSecond ifaces.Column
	// L2L1 logs: inter is a shifter version of hashSecond, necessary due to how the MiMC constraints operate
	inter ifaces.Column
}

// NewLogHasher returns a new LogHasher with initialized columns that are not constrained.
func NewLogHasher(comp *wizard.CompiledIOP, size int, name string) LogHasher {
	res := LogHasher{
		counter:    publicInput.CreateCol(name, "COUNTER", size, comp),
		ctMax:      publicInput.CreateCol(name, "COUNTER_MAX", size, comp),
		hashFirst:  publicInput.CreateCol(name, "HASH_FIRST", size, comp),
		hashSecond: publicInput.CreateCol(name, "HASH_SECOND", size, comp),
		inter:      publicInput.CreateCol(name, "INTER", size, comp),
	}
	return res
}

// DefineHasher specifies the constraints of the LogHasher with respect to the ExtractedData fetched from the arithmetization
func DefineHasher(comp *wizard.CompiledIOP, hasher LogHasher, name string, fetched ExtractedData) {

	// MiMC constraints
	comp.InsertMiMC(0, ifaces.QueryIDf("%s_%s", name, "MIMC_CONSTRAINT"), fetched.Hi, hasher.inter, hasher.hashFirst)
	comp.InsertMiMC(0, ifaces.QueryIDf("%s_%s", name, "MIMC_CONSTRAINT_SECOND"), fetched.Lo, hasher.hashFirst, hasher.hashSecond)

	// intermediary state integrity
	comp.InsertGlobal(0, ifaces.QueryIDf("%s_%s", name, "CONSISTENCY_INTER_AND_HASH_LAST"), // LAST is either hashSecond or hashThird
		sym.Sub(hasher.hashSecond,
			column.Shift(hasher.inter, 1),
		),
	)

	// Counter constraints
	// First, the counter starts from 0
	comp.InsertLocal(0, ifaces.QueryIDf("%s_%s", name, "COUNTER_LOCAL"), ifaces.ColumnAsVariable(hasher.counter))

	// Secondly, the counter increases by 1 every time.
	comp.InsertGlobal(0, ifaces.QueryIDf("%s_%s", name, "COUNTER_GLOBAL"),
		sym.Sub(hasher.counter,
			column.Shift(hasher.counter, -1),
			1,
		),
	)

	// active is already constrained in the fetcher, no need to constrain it again
	// two cases: Case 1: isActive is not completely full, then  ctMax is equal to the counter at the last cell where isActive is 1
	isActive := fetched.filterFetched
	comp.InsertGlobal(0, ifaces.QueryIDf("%s_%s", name, "COUNTER_MAX"),
		sym.Mul(
			isActive,
			sym.Sub(1,
				column.Shift(isActive, 1),
			),
			sym.Sub(
				hasher.counter,
				hasher.ctMax,
			),
		),
	)

	// Case 2: isActive is completely full, in which case we ask that isActive[size]*(counter[size]-ctMax[size]) = 0
	// i.e. at the last row, counter is equal to ctMax
	comp.InsertLocal(0, ifaces.QueryIDf("%s_%s", name, "COUNTER_MAX_LOCAL"),
		sym.Mul(
			column.Shift(isActive, -1),
			sym.Sub(
				column.Shift(hasher.counter, -1),
				column.Shift(hasher.ctMax, -1),
			),
		),
	)
}

// AssignHasher assigns the data in the LogHasher using the ExtractedData fetched from the arithmetization
func AssignHasher(run *wizard.ProverRuntime, hasher LogHasher, fetched ExtractedData) {
	size := fetched.Hi.Size()
	hashFirst := make([]field.Element, size)
	hashSecond := make([]field.Element, size)
	inter := make([]field.Element, size)
	counter := make([]field.Element, size)

	var (
		maxIndex field.Element
	)

	state := field.Zero() // the initial state is zero
	for i := 0; i < len(hashFirst); i++ {
		// first, hash the HI part of the fetched log message
		state = mimc.BlockCompression(state, fetched.Hi.GetColAssignmentAt(run, i))
		hashFirst[i].Set(&state)

		// secondly, hash the Lo part of the fetched log message
		state = mimc.BlockCompression(state, fetched.Lo.GetColAssignmentAt(run, i))
		hashSecond[i].Set(&state)

		// set the counter
		counter[i].SetInt64(int64(i))

		// the data in hashSecond is used to initialize the next initial state, stored in the inter column
		if i+1 < len(hashFirst) {
			inter[i+1] = hashSecond[i]
		}

		isActive := fetched.filterFetched.GetColAssignmentAt(run, i)
		if isActive.IsOne() {
			maxIndex.SetInt64(int64(i))
			// keep track of the maximal number of active rows
		}
	}

	// assign the hasher columns
	run.AssignColumn(hasher.hashFirst.GetColID(), smartvectors.NewRegular(hashFirst))
	run.AssignColumn(hasher.hashSecond.GetColID(), smartvectors.NewRegular(hashSecond))
	run.AssignColumn(hasher.inter.GetColID(), smartvectors.NewRegular(inter))
	run.AssignColumn(hasher.counter.GetColID(), smartvectors.NewRegular(counter))
	run.AssignColumn(hasher.ctMax.GetColID(), smartvectors.NewConstant(maxIndex, size))
}
