package execution_data_collector

import (
	"github.com/consensys/linea-monorepo/prover/crypto/mimc"
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
	// a typical isActive binary column, provided as an input to the module
	isActive ifaces.Column
	// the data to be hashed, this column is provided as an input to the module
	inputData      ifaces.Column
	inputIsActive  ifaces.Column
	data           ifaces.Column
	isData         ifaces.Column //isActive * canBeData
	isDataFirstRow *dedicated.HeartBeatColumn
	isDataOddRows  *dedicated.HeartBeatColumn
	// this column stores the MiMC hashes
	hash ifaces.Column
	// a constant column that stores the last relevant value of the hash
	HashFinal ifaces.Column
	// state is an intermediary column used to enforce the MiMC constraints
	state ifaces.Column
}

func NewMIMCHasher(comp *wizard.CompiledIOP, inputData, inputIsActive ifaces.Column, name string) *MIMCHasher {
	size := 2 * inputData.Size()
	res := &MIMCHasher{
		inputData:     inputData,
		inputIsActive: inputIsActive,
		data:          util.CreateCol(name, "DATA", size, comp),
		isActive:      util.CreateCol(name, "ACTIVE", size, comp),
		hash:          util.CreateCol(name, "HASH", size, comp),
		HashFinal:     util.CreateCol(name, "HASH_FINAL", size, comp),
		state:         util.CreateCol(name, "STATE", size, comp),
		isData:        util.CreateCol(name, "IS_DATA", size, comp),
	}
	res.isDataFirstRow = dedicated.CreateHeartBeat(comp, 0, size, 0, res.isActive)
	res.isDataOddRows = dedicated.CreateHeartBeat(comp, 0, 2, 1, res.isActive)
	return res
}

func DefineHashFilterConstraints(comp *wizard.CompiledIOP, hasher *MIMCHasher, name string) {

	// we require that isActive is binary in DefineIndicatorsMustBeBinary
	// require that the isActive filter only contains 1s followed by 0s
	comp.InsertGlobal(
		0,
		ifaces.QueryIDf("%s_IS_ACTIVE_CONSTRAINT_NO_0_TO_1", name),
		sym.Sub(
			hasher.isActive,
			sym.Mul(
				column.Shift(hasher.isActive, -1),
				hasher.isActive,
			),
		),
	)
	util.MustBeBinary(comp, hasher.isActive)

	comp.InsertGlobal(
		0,
		ifaces.QueryIDf("%s_IS_DATA", name),
		sym.Sub(
			hasher.isData,
			hasher.isDataFirstRow.Natural,
			hasher.isDataOddRows.Natural,
		),
	)
	util.MustBeBinary(comp, hasher.isData)
}

// DefineHasher defines the constraints of the MIMCHasher.
// Its isActive and data columns are assumed to be already constrained in another module, no need to constrain them again.
func (hasher *MIMCHasher) DefineHasher(comp *wizard.CompiledIOP, name string) {

	// MiMC constraints
	comp.InsertMiMC(0, ifaces.QueryIDf("%s_%s", name, "MIMC_CONSTRAINT"), hasher.data, hasher.state, hasher.hash, nil)

	// intermediary state integrity
	comp.InsertGlobal(0, ifaces.QueryIDf("%s_%s", name, "CONSISTENCY_STATE_AND_HASH_LAST"), // LAST is either hashSecond
		sym.Add(
			sym.Mul(
				hasher.isData,
				sym.Sub(hasher.state,
					column.Shift(hasher.hash, -1),
				),
			),
			sym.Mul(
				sym.Sub(1, hasher.isData),
				sym.Sub(hasher.state,
					0,
				),
			),
		),
	)

	comp.InsertGlobal(0, ifaces.QueryIDf("%s_%s", name, "CONSISTENCY_STATE_AND_HASH_LAST_2"), // LAST is either hashSecond
		sym.Mul(
			hasher.isActive,
			sym.Sub(1, hasher.isData),
			sym.Sub(hasher.data,
				column.Shift(hasher.hash, -1),
			),
		),
	)

	// state, the current state column, is initially zero
	comp.InsertLocal(0, ifaces.QueryIDf("%s_%s", name, "INTER_LOCAL"), ifaces.ColumnAsVariable(hasher.state))

	// constrain HashFinal
	commonconstraints.MustBeConstant(comp, hasher.HashFinal)
	util.CheckLastELemConsistency(comp, hasher.isActive, hasher.hash, hasher.HashFinal, name)

	// constraint isActive
	DefineHashFilterConstraints(comp, hasher, name)

	comp.InsertProjection(
		ifaces.QueryIDf("%s_%s", name, "PROJECTION_DATA"),
		query.ProjectionInput{
			ColumnA: []ifaces.Column{hasher.data},
			ColumnB: []ifaces.Column{hasher.inputData},
			FilterA: hasher.isData,
			FilterB: hasher.inputIsActive,
		},
	)

}

// AssignHasher assigns the data in the MIMCHasher. The data and isActive columns are fetched from another module.
func (hasher *MIMCHasher) AssignHasher(run *wizard.ProverRuntime) {

	var (
		inputSize = hasher.inputData.Size()
		isData    = common.NewVectorBuilder(hasher.isData)
		isActive  = common.NewVectorBuilder(hasher.isActive)
		state     = common.NewVectorBuilder(hasher.state)
		data      = common.NewVectorBuilder(hasher.data)
		hash      = common.NewVectorBuilder(hasher.hash)
		finalHash field.Element
	)

	// Writing the first row
	isData.PushOne()
	isActive.PushOne()
	state.PushZero()
	data.PushField(hasher.inputData.GetColAssignmentAt(run, 0))
	hash.PushField(mimc.BlockCompression(state.Last(), data.Last()))

	// Writing the second row
	isData.PushOne()
	isActive.PushOne()
	state.PushField(hash.Last())
	data.PushField(hasher.inputData.GetColAssignmentAt(run, 1))
	hash.PushField(mimc.BlockCompression(state.Last(), data.Last()))

	for j := 2; j < inputSize; j++ {

		inputIsActive := hasher.inputIsActive.GetColAssignmentAt(run, j)
		if !inputIsActive.IsOne() {
			finHash := hash.Last()
			finalHash.Set(&finHash)
			break
		}

		// Odds rows, we start a new hash from zero, using the previous hash
		// as data.
		isData.PushZero()
		isActive.PushOne()
		state.PushZero()
		data.PushField(hash.Last())
		hash.PushField(mimc.BlockCompression(state.Last(), data.Last()))

		// Evens rows, we continue the hash by adding the input data corresponding
		// to "j".
		isData.PushOne()
		isActive.PushOne()
		state.PushField(hash.Last())
		data.PushField(hasher.inputData.GetColAssignmentAt(run, j))
		hash.PushField(mimc.BlockCompression(state.Last(), data.Last()))
	}

	// assign the hasher columns
	isData.PadAndAssign(run, field.Zero())
	isActive.PadAndAssign(run, field.Zero())
	state.PadAndAssign(run, field.Zero())
	data.PadAndAssign(run, field.Zero())
	hash.PadAndAssign(run, mimc.BlockCompression(field.Zero(), field.Zero()))
	run.AssignColumn(hasher.HashFinal.GetColID(), smartvectors.NewConstant(finalHash, hasher.HashFinal.Size()))

	hasher.isDataFirstRow.Assign(run)
	hasher.isDataOddRows.Assign(run)
}
