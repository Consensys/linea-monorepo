package logdata

import (
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/sirupsen/logrus"
)

func Log(msg string) func(comp *wizard.CompiledIOP) {

	return func(comp *wizard.CompiledIOP) {

		var (
			numCommitment     = map[column.Status]int{}
			numCells          = map[column.Status]int{}
			encounteredStatus = []column.Status{}
		)

		// Count the total number of commitment into play and the number of cells
		for _, name := range comp.Columns.AllKeys() {
			status := comp.Columns.Status(name)
			size := comp.Columns.GetHandle(name).Size()

			// Don't log data relative to ignored status
			if status == column.Ignored {
				continue
			}

			if _, ok := numCommitment[status]; !ok {
				numCommitment[status] = 0
				numCells[status] = 0
				encounteredStatus = append(encounteredStatus, status)
			}

			numCommitment[status]++
			numCells[status] += size
		}

		for _, status := range encounteredStatus {
			logrus.Infof(
				"LOG METADATA : msg \"%v\"- %v - total numcomms %v - numcells %v\n",
				msg, status.String(), numCommitment[status], numCells[status],
			)
		}

		// And also calls the column dimensions break down analysis
		columnDims(msg)(comp)
	}
}

func columnDims(msg string) func(comp *wizard.CompiledIOP) {

	return func(comp *wizard.CompiledIOP) {

		columnsDims := map[int]int{}

		// Count the total number of commitment into play and the number of cells
		for _, name := range comp.Columns.AllKeys() {
			status := comp.Columns.Status(name)
			size := comp.Columns.GetHandle(name).Size()

			// Only care about committed data
			if status != column.Committed {
				continue
			}

			if _, ok := columnsDims[size]; !ok {
				columnsDims[size] = 0
			}

			columnsDims[size]++

		}

		numPrinted := 0
		cumulativeCells := 0

		// print the stats founds sorted in ascending column sizes
		for size := 1; numPrinted < len(columnsDims); size *= 2 {
			// found no columns with that size
			if _, ok := columnsDims[size]; !ok {
				continue
			}

			numCol := columnsDims[size]
			numCells := numCol * size
			cumulativeCells += numCells
			numPrinted += 1

			logrus.Infof("[%v] COLUMN DIMENSION PROFILE: size=%d #colums=%d #cells=%d cumCells=%d", msg, size, numCol, numCells, cumulativeCells)
		}
	}
}
