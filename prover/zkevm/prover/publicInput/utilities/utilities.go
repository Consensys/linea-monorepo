package utilities

import (
	"github.com/consensys/zkevm-monorepo/prover/maths/field"
	"github.com/consensys/zkevm-monorepo/prover/protocol/ifaces"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
	sym "github.com/consensys/zkevm-monorepo/prover/symbolic"
	"github.com/consensys/zkevm-monorepo/prover/utils/csvtraces"
	"github.com/ethereum/go-ethereum/core/vm"
	"os"
	"testing"
)

// CreateCol is a utility function to quickly register columns
func CreateCol(name, subName string, size int, comp *wizard.CompiledIOP) ifaces.Column {
	return comp.InsertCommit(
		0,
		ifaces.ColIDf("%s_%s", name, subName),
		size,
	)
}

// Ternary is a small utility to construct ternaries is constraints
func Ternary(cond, if1, if0 any) *sym.Expression {
	return sym.Add(
		sym.Mul(sym.Sub(1, cond), if0),
		sym.Mul(cond, if1),
	)
}

// GetTimestampField returns a field element that contains the hardcoded INST value for a timestamp
func GetTimestampField() field.Element {
	var timestampField field.Element
	stampCode := byte(vm.TIMESTAMP)
	hardcoded := [...]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, stampCode}
	timestampField.SetBytes(hardcoded[:])
	return timestampField
}

// InitializeCsv is used to initialize a CsvTrace based on a path
func InitializeCsv(csvPath string, t *testing.T) *csvtraces.CsvTrace {
	f, err := os.Open(csvPath)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	ct, err := csvtraces.NewCsvTrace(f)
	if err != nil {
		t.Fatal(err)
	}
	return ct
}

// MustBeBinary constrains the current column to be binary.
func MustBeBinary(comp *wizard.CompiledIOP, c ifaces.Column) {
	comp.InsertGlobal(
		0,
		ifaces.QueryIDf("%v_MUST_BE_BINARY", c.GetColID()),
		sym.Mul(c, sym.Sub(c, 1)),
	)
}
