package execution_data_collector

import (
	"github.com/consensys/linea-monorepo/prover/crypto/mimc"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/accessors"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	sym "github.com/consensys/linea-monorepo/prover/symbolic"
	util "github.com/consensys/linea-monorepo/prover/zkevm/prover/publicInput/utilities"
)

type MIMCHasher struct {
	// a typical isActive binary column, provided as an input to the module
	isActive ifaces.Column
	// the data to be hashed, this column is provided as an input to the module
	data ifaces.Column
	// this column stores the MiMC hashes
	hash ifaces.Column
	// a constant column that stores the last relevant value of the hash
	HashFinal ifaces.Column
	// inter is an intermediary column used to enforce the MiMC constraints
	inter ifaces.Column
}

func NewMIMCHasher(comp *wizard.CompiledIOP, data, isActive ifaces.Column, name string) *MIMCHasher {
	size := data.Size()
	res := &MIMCHasher{
		data:      data,
		isActive:  isActive,
		hash:      util.CreateCol(name, "HASH", size, comp),
		HashFinal: util.CreateCol(name, "HASH_FINAL", 1, comp),
		inter:     util.CreateCol(name, "INTER", size, comp),
	}
	return res
}

// DefineHasher defines the constraints of the MIMCHasher.
// Its isActive and data columns are assumed to be already constrained in another module, no need to constrain them again.
func (hasher *MIMCHasher) DefineHasher(comp *wizard.CompiledIOP, name string) {

	// MiMC constraints
	comp.InsertMiMC(0, ifaces.QueryIDf("%s_%s", name, "MIMC_CONSTRAINT"), hasher.data, hasher.inter, hasher.hash)

	// intermediary state integrity
	comp.InsertGlobal(0, ifaces.QueryIDf("%s_%s", name, "CONSISTENCY_INTER_AND_HASH_LAST"), // LAST is either hashSecond
		sym.Sub(hasher.hash,
			column.Shift(hasher.inter, 1),
		),
	)

	// inter, the old state column, is initially zero
	comp.InsertLocal(0, ifaces.QueryIDf("%s_%s", name, "INTER_LOCAL"), ifaces.ColumnAsVariable(hasher.inter))

	// prepare accessors for HashFinal
	comp.Columns.SetStatus(hasher.HashFinal.GetColID(), column.Proof)
	accHashFinal := accessors.NewFromPublicColumn(hasher.HashFinal, 0)
	// constrain HashFinal
	util.CheckLastELemConsistency(comp, hasher.isActive, hasher.hash, accHashFinal, name)

}

// AssignHasher assigns the data in the MIMCHasher. The data and isActive columns are fetched from another module.
func (hasher *MIMCHasher) AssignHasher(run *wizard.ProverRuntime) {
	size := hasher.data.Size()
	hash := make([]field.Element, size)
	inter := make([]field.Element, size)
	counter := make([]field.Element, size)

	var (
		maxIndex  field.Element
		finalHash field.Element
	)

	state := field.Zero() // the initial state is zero
	for i := 0; i < len(hash); i++ {
		// first, hash the HI part of the fetched log message
		mimcBlock := hasher.data.GetColAssignmentAt(run, i)
		// debugString.WriteString(mimcBlock.)
		state = mimc.BlockCompression(state, mimcBlock)
		hash[i].Set(&state)

		// set the counter
		counter[i].SetInt64(int64(i))

		// the data in hashSecond is used to initialize the next initial state, stored in the inter column
		if i+1 < len(hash) {
			inter[i+1] = hash[i]
		}

		isActive := hasher.isActive.GetColAssignmentAt(run, i)
		if isActive.IsOne() {
			finalHash.Set(&hash[i])
			maxIndex.SetInt64(int64(i))
			// keep track of the maximal number of active rows
		}
	}

	// assign the hasher columns
	run.AssignColumn(hasher.hash.GetColID(), smartvectors.NewRegular(hash))
	run.AssignColumn(hasher.HashFinal.GetColID(), smartvectors.NewRegular([]field.Element{finalHash}))
	run.AssignColumn(hasher.inter.GetColID(), smartvectors.NewRegular(inter))
}
