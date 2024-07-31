package logdata

import (
	"fmt"
	"io"

	"github.com/consensys/zkevm-monorepo/prover/protocol/column"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
)

// Dump the columns into a csv file
func GenCSV(w io.Writer) func(comp *wizard.CompiledIOP) {

	return func(comp *wizard.CompiledIOP) {

		columns := comp.Columns.AllKeys()

		io.WriteString(w, "name; size; status; round\n")

		for _, colID := range columns {
			col := comp.Columns.GetHandle(colID)
			status := comp.Columns.Status(colID)
			size := col.Size()
			round := col.Round()

			// Skip the ignored columns
			if status == column.Ignored {
				continue
			}

			fmtline := fmt.Sprintf("%v; %v; %v; %v\n", colID, size, status, round)
			io.WriteString(w, fmtline)
		}
	}

}
