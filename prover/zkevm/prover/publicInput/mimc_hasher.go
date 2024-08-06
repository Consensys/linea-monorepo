package publicInput

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

type MIMCHasher struct {
	// a typical isActive binary column, provided as an input to the module
	isActive ifaces.Column
	// the data to be hashed, this column is provided as an input to the module
	data ifaces.Column
	// this column stores the MiMC hashes
	hash ifaces.Column
	// a constant column that stores the last relevant value of the hash
	hashFinal ifaces.Column
	// inter is an intermediary column used to enforce the MiMC constraints
	inter ifaces.Column
}

func NewMIMCHasher(comp *wizard.CompiledIOP, data, isActive ifaces.Column, name string) *MIMCHasher {
	size := data.Size()
	res := &MIMCHasher{
		data:      data,
		isActive:  isActive,
		hash:      publicInput.CreateCol(name, "HASH", size, comp),
		hashFinal: publicInput.CreateCol(name, "HASH_FINAL", size, comp),
		inter:     publicInput.CreateCol(name, "INTER", size, comp),
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

	// below, constrain the hashFinal column
	// two cases: Case 1: isActive is not completely full, then  ctMax is equal to the counter at the last cell where isActive is 1
	comp.InsertGlobal(0, ifaces.QueryIDf("%s_%s", name, "HASH_FINAL_GLOBAL_CONSTRAINT"),
		sym.Mul(
			hasher.isActive,
			sym.Sub(1,
				column.Shift(hasher.isActive, 1),
			),
			sym.Sub(
				hasher.hash,
				hasher.hashFinal,
			),
		),
	)

	// Case 2: isActive is completely full, in which case we ask that isActive[size]*(counter[size]-ctMax[size]) = 0
	// i.e. at the last row, counter is equal to ctMax
	comp.InsertLocal(0, ifaces.QueryIDf("%s_%s", name, "HASH_FINAL_LOCAL_CONSTRAINT"),
		sym.Mul(
			column.Shift(hasher.isActive, -1),
			sym.Sub(
				column.Shift(hasher.hash, -1),
				column.Shift(hasher.hashFinal, -1),
			),
		),
	)

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
	run.AssignColumn(hasher.hashFinal.GetColID(), smartvectors.NewConstant(finalHash, size))
	run.AssignColumn(hasher.inter.GetColID(), smartvectors.NewRegular(inter))
}
