package zkevm_keccak

import (
	keccakf "github.com/consensys/accelerated-crypto-monorepo/crypto/keccak_wizard/keccakf"
	"github.com/consensys/accelerated-crypto-monorepo/maths/common/smartvectors"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/ifaces"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/wizard"
	"github.com/sirupsen/logrus"
)

// attributes of a zkevm module from JSON Trace
type ZkModuleAtt struct {
	//hash or transaction number
	HashNum ifaces.ColID
	// index for limbs of the same hash/TX
	INDEX ifaces.ColID
	// limbs of the hash/TX (short limbs are padded by zero)
	LIMB ifaces.ColID
	// real size of a limb
	Nbytes ifaces.ColID

	// special columns for txRlp, these are flags for unsigned transactions
	LX ifaces.ColID
	LC ifaces.ColID
}

var AttTXRLP = ZkModuleAtt{HashNum: "txRlp.ABS_TX_NUM", INDEX: "txRlp.INDEX_LX", LIMB: "txRlp.LIMB", Nbytes: "txRlp.nBYTES", LX: "txRlp.LX", LC: "txRlp.LIMB_CONSTRUCTED"}

// a Module with random general tables, used for testing
var AttRAND = ZkModuleAtt{HashNum: "RANT.TX_NUM", INDEX: "RAND.INDEX", LIMB: "RAND.LIMB", Nbytes: "RAND.nBYTES"}
var AttPhoneyRlp = ZkModuleAtt{HashNum: "phoneyRLP.TX_NUM", INDEX: "phoneyRLP.INDEX", LIMB: "phoneyRLP.LIMB", Nbytes: "phoneyRLP.nBYTES"}

// InfoTrace, a table of three columns
// for the current zkModules we done have  the InfoTrace, so is muted for now but easy to adjust
/*var (
	HashNumInfo ifaces.ColID
	HashHI      ifaces.ColID
	HashLO      ifaces.ColID
)*/

/*
numPerm : the number of permutation supported by KeccakFModule

	size of the KeccakFModule is numPerm*32
*/

// Registers the Keccak wizard over the zkevm arithmetization

func RegisterKeccak(comp *wizard.CompiledIOP, round, numPerm int) {
	// define the keccak module. The module defines itself as a collection of columns
	// and queries that together make up the module. numPerm is an indication of the
	// number of rows that we need and round indicates at which round of the protocol
	// the keccak module should be populated. The define function **must not** schedule
	// any assignment of the column it declares.

	// Hotfix, declare the phoneyRLP columns if they are not
	if !comp.Columns.Exists(AttPhoneyRlp.Nbytes) {
		comp.InsertCommit(round, AttPhoneyRlp.Nbytes, 1<<17)
		comp.InsertCommit(round, AttPhoneyRlp.HashNum, 1<<17)
		comp.InsertCommit(round, AttPhoneyRlp.INDEX, 1<<17)
		comp.InsertCommit(round, AttPhoneyRlp.LIMB, 1<<17)
	}

	keccakFModule := DefineKeccakF(comp, round, numPerm)

	// registers a prover step that will populate the module with data arising from the arithmetization
	comp.SubProvers.AppendToInner(round, func(run *wizard.ProverRuntime) {
		logrus.Infof("assigning keccak tables from the runtime")
		// Extract the table from the zkEVM arithmetization
		extractedTables := ExtractTableFromRuntime(run, AttPhoneyRlp, PhoneyRLP)
		logrus.Infof("successfully assigned the tables from the runtime")
		// And we assign the columns of the keccakF module from the `table`.
		// NB: the function MUST NOT mutate the `keccakFMod` object
		AssignColKeccakF(run, extractedTables, keccakFModule, numPerm)
		logrus.Infof("successfully assigned the keccak module")
	})

}

func DefineKeccakF(comp *wizard.CompiledIOP, round, numPerm int) (keccakFModule keccakf.KeccakFModule) {
	// it build the Module via the given size 'numPerm'
	keccakFModule.NP = numPerm
	keccakFModule.DefineKeccakF(comp, round)

	return keccakFModule
}

// get the assignment for run  and output a table
func ExtractTableFromRuntime(run *wizard.ProverRuntime, moduleAtt ZkModuleAtt, moduleName Module) (table Tables) {
	n := DetectRedundancy(run, moduleAtt)
	hashNum := ConvertToInt(run.GetColumn(moduleAtt.HashNum), n)
	index := ConvertToInt(run.GetColumn(moduleAtt.INDEX), n)
	limb := ConvertToByte16(run.GetColumn(moduleAtt.LIMB), n)
	nByte := ConvertToInt(run.GetColumn(moduleAtt.Nbytes), n)

	if moduleName == TXRLP {
		table.LX = ConvertToInt(run.GetColumn(moduleAtt.LX), n)
		table.LC = ConvertToInt(run.GetColumn(moduleAtt.LC), n)

	}
	if moduleName == TXRLP {
		table.HasInfoTrace = false
	} else {
		table.HasInfoTrace = true
	}
	in := DataTrace{HashNum: hashNum, Limb: limb, Nbytes: nByte, Index: index}
	return Tables{InputTable: in, HasInfoTrace: table.HasInfoTrace, LX: table.LX, LC: table.LC}
}
func AssignColKeccakF(run *wizard.ProverRuntime, table Tables, module keccakf.KeccakFModule, numPerm int) {
	// extracts input/output of multihash from the table
	multiHash := table.MultiHashFromTable(numPerm, TXRLP)
	// extract the permutations and assign the column for the keccakf Module
	multiHash.Prover(run, module)
}

func ConvertToInt(v smartvectors.SmartVector, n int) (u []int) {
	len := v.Len()
	u = make([]int, len-n)
	for i := n; i < len; i++ {
		a := v.Get(i)
		u[i-n] = int(a.Uint64())
	}

	return u
}

func ConvertToByte16(v smartvectors.SmartVector, n int) (u []Byte16) {
	len := v.Len()
	u = make([]Byte16, len-n)
	for i := n; i < len; i++ {
		a := v.Get(i)
		b := a.Bytes()
		copy(u[i-n][:], b[16:])
	}
	return u
}

/*
it removes the redundant zero rows (added by the wizard)

Note: in the trace the zkevm Module may have some zero rows at the beginning  (before even going to wizard),
apparently they are not important for hashing and so are dropped as well
*/
func DetectRedundancy(run *wizard.ProverRuntime, att ZkModuleAtt) int {
	nbyte := run.GetColumn(att.Nbytes)
	vec := smartvectors.IntoRegVec(nbyte)

	// the first entry is non-zero, so we return 0
	if !vec[0].IsZero() {
		return 0
	}

	// Look for the first non-zero entry
	for i := range vec {
		if !vec[i].IsZero() {
			return i
		}
	}

	// Edge-case all entries are zero
	return len(vec)
}
