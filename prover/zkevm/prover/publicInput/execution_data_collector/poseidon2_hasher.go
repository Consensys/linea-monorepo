package execution_data_collector

import (
	"math"

	"github.com/consensys/linea-monorepo/prover/crypto/poseidon2_koalabear"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/dedicated"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	sym "github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/common"
	commonconstraints "github.com/consensys/linea-monorepo/prover/zkevm/prover/common/common_constraints"
	util "github.com/consensys/linea-monorepo/prover/zkevm/prover/publicInput/utilities"
)

type MIMCHasher struct {
	// a typical IsActive binary column, provided as an input to the module
	IsActive ifaces.Column
	// the data to be hashed, this column is provided as an input to the module
	InputData      ifaces.Column
	InputIsActive  ifaces.Column
	Data           [common.NbElemPerHash]ifaces.Column
	IsData         ifaces.Column //isActive * canBeData
	IsDataFirstRow *dedicated.HeartBeatColumn
	IsDataOddRows  *dedicated.HeartBeatColumn
	// this column stores the MiMC hashes
	Hash [common.NbElemPerHash]ifaces.Column
	// a constant column that stores the last relevant value of the hash
	HashFinal [common.NbElemPerHash]ifaces.Column
	// state is an intermediary column used to enforce the MiMC constraints
	State [common.NbElemPerHash]ifaces.Column
}

func NewMIMCHasher(comp *wizard.CompiledIOP, inputData, inputIsActive ifaces.Column, name string) *MIMCHasher {
	size := 2 * inputData.Size()
	res := &MIMCHasher{
		InputData:     inputData,
		InputIsActive: inputIsActive,
		IsActive:      util.CreateCol(name, "ACTIVE", size, comp),
		IsData:        util.CreateCol(name, "IS_DATA", size, comp),
	}
	// Passing a very large size will ensure that the heartbeat produces the
	// same result even if we change the length of the column.
	res.IsDataFirstRow = dedicated.CreateHeartBeat(comp, 0, math.MaxInt, 0, res.IsActive)
	res.IsDataOddRows = dedicated.CreateHeartBeat(comp, 0, 2, 1, res.IsActive)
	return res
}

func DefineHashFilterConstraints(comp *wizard.CompiledIOP, hasher *MIMCHasher, name string) {

	// we require that isActive is binary in DefineIndicatorsMustBeBinary
	// require that the isActive filter only contains 1s followed by 0s
	comp.InsertGlobal(
		0,
		ifaces.QueryIDf("%s_IS_ACTIVE_CONSTRAINT_NO_0_TO_1", name),
		sym.Sub(
			hasher.IsActive,
			sym.Mul(
				column.Shift(hasher.IsActive, -1),
				hasher.IsActive,
			),
		),
	)
	util.MustBeBinary(comp, hasher.IsActive)

	comp.InsertGlobal(
		0,
		ifaces.QueryIDf("%s_IS_DATA", name),
		sym.Sub(
			hasher.IsData,
			hasher.IsDataFirstRow.Natural,
			hasher.IsDataOddRows.Natural,
		),
	)
	util.MustBeBinary(comp, hasher.IsData)
}

// DefineHasher defines the constraints of the MIMCHasher.
// Its isActive and data columns are assumed to be already constrained in another module, no need to constrain them again.
func (hasher *MIMCHasher) DefineHasher(comp *wizard.CompiledIOP, name string) {

	comp.InsertPoseidon2(0, ifaces.QueryIDf("%s_%s", name, "MIMC_CONSTRAINT"), hasher.Data, hasher.State, hasher.Hash, nil)

	for i := range hasher.Hash {
		// intermediary state integrity
		comp.InsertGlobal(0, ifaces.QueryIDf("%s_CONSISTENCY_STATE_AND_HASH_LAST_%d", name, i), // LAST is either hashSecond
			sym.Add(
				sym.Mul(
					hasher.IsData,
					sym.Sub(hasher.State[i],
						column.Shift(hasher.Hash[i], -1),
					),
				),
				sym.Mul(
					sym.Sub(1, hasher.IsData),
					sym.Sub(hasher.State[i],
						0,
					),
				),
			),
		)

		// LAST is either hashSecond
		comp.InsertGlobal(0, ifaces.QueryIDf("%s_CONSISTENCY_STATE_AND_HASH_LAST_2_%d", name, i),
			sym.Mul(
				hasher.IsActive,
				sym.Sub(1, hasher.IsData),
				sym.Sub(hasher.Data[i],
					column.Shift(hasher.Hash[i], -1),
				),
			),
		)

		// state, the current state column, is initially zero
		comp.InsertLocal(0, ifaces.QueryIDf("%s_INTER_LOCAL_%d", name, i),
			ifaces.ColumnAsVariable(hasher.State[i]),
		)

		// constrain HashFinal
		commonconstraints.MustBeConstant(comp, hasher.HashFinal[i])
		util.CheckLastELemConsistency(comp, hasher.IsActive, hasher.Hash[i], hasher.HashFinal[i], name)
	}

	// constraint isActive
	DefineHashFilterConstraints(comp, hasher, name)

	comp.InsertProjection(
		ifaces.QueryIDf("%s_%s", name, "PROJECTION_DATA"),
		query.ProjectionInput{
			ColumnA: []ifaces.Column{hasher.Data[common.NbElemPerHash-1]}, // input data is the last limb of the data column
			ColumnB: []ifaces.Column{hasher.InputData},
			FilterA: hasher.IsData,
			FilterB: hasher.InputIsActive,
		},
	)

	// Check that the data column is zero for all but the last limb
	for i := 0; i < common.NbElemPerHash-1; i++ {
		comp.InsertGlobal(0, ifaces.QueryIDf("%s_DATA_ZERO_LIMBS_AT_INPUT_%d", name, i),
			sym.Mul(hasher.IsData, hasher.Data[i]),
		)
	}
}

// AssignHasher assigns the data in the MIMCHasher. The data and isActive columns are fetched from another module.
func (hasher *MIMCHasher) AssignHasher(run *wizard.ProverRuntime) {

	var (
		state, hash, data [common.NbElemPerHash]*common.VectorBuilder

		inputSize = hasher.InputData.Size()
		isData    = common.NewVectorBuilder(hasher.IsData)
		isActive  = common.NewVectorBuilder(hasher.IsActive)
	)

	for i := range state {
		state[i] = common.NewVectorBuilder(hasher.State[i])
		hash[i] = common.NewVectorBuilder(hasher.Hash[i])
		data[i] = common.NewVectorBuilder(hasher.Data[i])

		// the initial state is zero
		state[i].PushZero()
	}

	// Helper function to perform BlockCompression and update hash
	var prevState, dataToHash [common.NbElemPerHash]field.Element
	performBlockCompression := func(isDataVal field.Element) {
		isData.PushField(isDataVal)
		isActive.PushOne()

		for i := range prevState {
			prevState[i] = state[i].Last()
			dataToHash[i] = data[i].Last()
		}

		dataHash := poseidon2_koalabear.Compress(prevState, dataToHash)
		for i := range hash {
			hash[i].PushField(dataHash[i])
		}
	}

	// Helper to push data with only the last limb set
	pushDataWithLastLimb := func(value field.Element) {
		for i := 0; i < common.NbElemPerHash-1; i++ {
			dataToHash[i].SetZero()
		}

		dataToHash[common.NbElemPerHash-1] = value
		for i := range data {
			data[i].PushField(dataToHash[i])
		}
	}

	// Writing the first row
	pushDataWithLastLimb(hasher.InputData.GetColAssignmentAt(run, 0))
	performBlockCompression(field.One())

	// Assign the state for the next hashing
	for i := range state {
		state[i].PushField(hash[i].Last())
	}

	// Writing the second row
	pushDataWithLastLimb(hasher.InputData.GetColAssignmentAt(run, 1))
	performBlockCompression(field.One())

	// Process remaining rows
	for j := 2; j < inputSize; j++ {
		inputIsActive := hasher.InputIsActive.GetColAssignmentAt(run, j)
		if inputIsActive.IsZero() {
			break
		}

		// Odd rows: start a new hash from zero, using the previous hash as data
		for i := range state {
			state[i].PushZero()
			data[i].PushField(hash[i].Last())
		}
		performBlockCompression(field.Zero())

		// Even rows: continue the hash by adding the input data
		for i := range state {
			state[i].PushField(hash[i].Last())
		}
		pushDataWithLastLimb(hasher.InputData.GetColAssignmentAt(run, j))
		performBlockCompression(field.One())
	}

	zeroLimbs := [common.NbElemPerHash]field.Element{}
	zeroHash := poseidon2_koalabear.Compress(zeroLimbs, zeroLimbs)

	// Assign the hasher columns
	isData.PadAndAssign(run, field.Zero())
	isActive.PadAndAssign(run, field.Zero())
	for i := range hash {
		// The order is important here, we assign the final hash, and then pad and assign the hash column
		run.AssignColumn(hasher.HashFinal[i].GetColID(), smartvectors.NewConstant(hash[i].Last(), hasher.HashFinal[i].Size()))

		state[i].PadAndAssign(run, field.Zero())
		data[i].PadAndAssign(run, field.Zero())
		hash[i].PadAndAssign(run, zeroHash[i])
	}

	hasher.IsDataFirstRow.Assign(run)
	hasher.IsDataOddRows.Assign(run)
}
