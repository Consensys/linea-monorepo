package logs

import (
	"fmt"

	"github.com/consensys/gnark-crypto/field/koalabear/vortex"
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
// the data in L2L1 logs must be hashed as follows: msg1Hash, msg2Hash, and so on.
// The final value of the chained hash can be retrieved as hash[ctMax[any index]]
type LogHasher struct {
	// the HashFirst value after each step, as explained in the description of LogHasher
	HashFirst  [common.NbLimbU256 / 2]ifaces.Column
	HashSecond [common.NbLimbU256 / 2]ifaces.Column
	// L2L1 logs: Inter is a shifted version of hashSecond, necessary due to how the MiMC constraints operate
	Inter [common.NbLimbU256 / 2]ifaces.Column
	// the relevant value of the hash (the last value when isActive ends)
	HashFinal [common.NbLimbU256 / 2]ifaces.Column
}

// NewLogHasher returns a new LogHasher with initialized columns that are not constrained.
func NewLogHasher(comp *wizard.CompiledIOP, size int, name string) LogHasher {
	var res LogHasher

	for i := range res.HashFirst {
		res.HashFirst[i] = util.CreateCol(name, fmt.Sprintf("HASH_FIRST_%d", i), size, comp)
		res.Inter[i] = util.CreateCol(name, fmt.Sprintf("INTER_%d", i), size, comp)
		res.HashSecond[i] = util.CreateCol(name, fmt.Sprintf("HASH_SECOND_%d", i), size, comp)
		res.HashFinal[i] = util.CreateCol(name, fmt.Sprintf("HASH_FINAL_%d", i), size, comp)
	}

	return res
}

// DefineHasher specifies the constraints of the LogHasher with respect to the ExtractedData fetched from the arithmetization
func DefineHasher(comp *wizard.CompiledIOP, hasher LogHasher, name string, fetched ExtractedData) {

	// Needed for the limitless prover to understand that the columns are
	// not just empty with just padding and suboptimal representation.
	for i := range hasher.HashFirst {
		pragmas.MarkRightPadded(hasher.HashFirst[i])
		pragmas.MarkRightPadded(hasher.HashSecond[i])
		pragmas.MarkRightPadded(hasher.HashFinal[i])
		pragmas.MarkRightPadded(hasher.Inter[i])
	}

	// Inter[0:8] on the first row is the initial state, which is zero
	// we hash the first 8 limbs and place the data into the Inter[8:16] columns
	comp.InsertPoseidon2(0, ifaces.QueryIDf("%s_%s", name, "POSEIDON2_QUERY_FIRST"),
		[8]ifaces.Column(fetched.Data[0:8]),
		hasher.Inter,
		hasher.HashFirst,
		fetched.FilterFetched,
	)

	comp.InsertPoseidon2(0, ifaces.QueryIDf("%s_%s", name, "POSEIDON2_QUERY_SECOND"),
		[8]ifaces.Column(fetched.Data[8:16]),
		hasher.HashFirst,
		hasher.HashSecond,
		fetched.FilterFetched,
	)

	for i := range hasher.HashSecond {
		// intermediary state integrity
		comp.InsertGlobal(0, ifaces.QueryIDf("%s_%s_%d", name, "CONSISTENCY_INTER_AND_HASH_SECOND", i), // LAST is either hashSecond or hashThird
			sym.Mul(
				sym.Sub(
					column.Shift(hasher.HashSecond[i], -1),
					hasher.Inter[i],
				),
				fetched.FilterFetched,
			),
		)

		// inter, the old state column, is initially zero
		comp.InsertLocal(
			0, ifaces.QueryIDf("%s_INTER_LOCAL_%d", name, i),
			ifaces.ColumnAsVariable(hasher.Inter[i]),
		)

		// constrain HashFinal
		commonconstraints.MustBeConstant(comp, hasher.HashFinal[i])
		util.CheckLastELemConsistency(comp, fetched.FilterFetched, hasher.HashSecond[i], hasher.HashFinal[i], name)
	}
}

// AssignHasher assigns the data in the LogHasher using the ExtractedData fetched from the arithmetization
func AssignHasher(run *wizard.ProverRuntime, hasher LogHasher, fetched ExtractedData) {
	size := fetched.Data[0].Size()
	var hashFirst, hashSecond, inter [8]*common.VectorBuilder

	for index := range hasher.HashFirst {
		hashFirst[index] = common.NewVectorBuilder(hasher.HashFirst[index])
		hashSecond[index] = common.NewVectorBuilder(hasher.HashSecond[index])
		inter[index] = common.NewVectorBuilder(hasher.Inter[index])
	}

	var (
		fetchedData [common.NbLimbU256]field.Element
		hashFinal   [common.NbLimbU256 / 2]field.Element
	)

	zeroedState := make([]field.Element, common.NbLimbU256/2)
	for i := range zeroedState {
		zeroedState[i].SetZero()
	}
	state := vortex.Hash(zeroedState) // the initial state is zero

	for i := 0; i < size; i++ {

		isActive := fetched.FilterFetched.GetColAssignmentAt(run, i)
		if isActive.IsZero() {
			break
		}

		for j := range fetchedData {
			fetchedData[j] = fetched.Data[j].GetColAssignmentAt(run, i)
		}

		// Inter is either the initial state so (zero) for row i=0 or the last
		// state from row i-1
		for j := range hasher.Inter {
			inter[j].PushField(state[j])
		}

		// first, hash the first 8 bytes part of the fetched log message
		state = vortex.CompressPoseidon2(vortex.Hash(state), vortex.Hash(fetchedData[0:8]))
		for j := range hashFirst {
			hashFirst[j].PushField(state[j])
		}

		// secondly, hash the last 8 bytes of the fetched log message
		state = vortex.CompressPoseidon2(vortex.Hash(state), vortex.Hash(fetchedData[8:16]))
		for j := range hashSecond {
			hashSecond[j].PushField(state[j])
		}

		// continuously update HashFinal. AthFinal will be the last value of
		// hashSecond where isActive ends
		for j := range hashFinal {
			hashFinal[j].Set(&state[j])
		}

	}

	// assign the hasher columns
	for index := range hasher.HashFirst {
		inter[index].PadAndAssign(run, field.Zero())
		hashSecond[index].PadAndAssign(run, field.Zero())
		hashFirst[index].PadAndAssign(run, field.Zero())
		run.AssignColumn(hasher.HashFinal[index].GetColID(), smartvectors.NewConstant(hashFinal[index], size))
	}
}
