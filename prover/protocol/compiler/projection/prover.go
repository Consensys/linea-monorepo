package projection

import (
	"fmt"
	"strings"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	sym "github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/sirupsen/logrus"
)

// projectionProverAction is a compilation artefact generated during the
// execution of the [InsertProjection] and which implements the
// [wizard.ProverAction]. It is meant to compute to assign the "Horner" columns
// and their respective local opening queries.
type projectionProverAction struct {
	Name               ifaces.QueryID
	FilterA, FilterB   ifaces.Column
	ColA, ColB         []ifaces.Column
	ABoard, BBoard     sym.ExpressionBoard
	EvalCoin           coin.Info
	HornerA, HornerB   ifaces.Column
	HornerA0, HornerB0 query.LocalOpening
}

// Run implements the [wizard.ProverAction] interface.
func (pa projectionProverAction) Run(run *wizard.ProverRuntime) {

	var (
		a       = column.EvalExprColumn(run, pa.ABoard).IntoRegVecSaveAlloc()
		b       = column.EvalExprColumn(run, pa.BBoard).IntoRegVecSaveAlloc()
		fA      = pa.FilterA.GetColAssignment(run).IntoRegVecSaveAlloc()
		fB      = pa.FilterB.GetColAssignment(run).IntoRegVecSaveAlloc()
		x       = run.GetRandomCoinField(pa.EvalCoin.Name)
		hornerA = cmptHorner(a, fA, x)
		hornerB = cmptHorner(b, fB, x)
	)

	run.AssignColumn(pa.HornerA.GetColID(), smartvectors.NewRegular(hornerA))
	run.AssignColumn(pa.HornerB.GetColID(), smartvectors.NewRegular(hornerB))
	run.AssignLocalPoint(pa.HornerA0.ID, hornerA[0])
	run.AssignLocalPoint(pa.HornerB0.ID, hornerB[0])

	if hornerA[0] != hornerB[0] {

		var (
			colA  = make([][]field.Element, len(pa.ColA))
			colB  = make([][]field.Element, len(pa.ColB))
			cntA  = 0
			cntB  = 0
			rowsA = [][]string{}
			rowsB = [][]string{}
		)

		for c := range pa.ColA {
			colA[c] = pa.ColA[c].GetColAssignment(run).IntoRegVecSaveAlloc()
			colB[c] = pa.ColB[c].GetColAssignment(run).IntoRegVecSaveAlloc()
		}

		for i := range fA {

			if fA[i].IsZero() {
				continue
			}

			row := make([]string, len(pa.ColA))

			for c := range pa.ColA {
				fString := colA[c][i].Text(16)
				if colA[c][i].IsUint64() && colA[c][i].Uint64() < 1000000 {
					fString = colA[c][i].String()
				}
				row[c] = fmt.Sprintf("%v=%v", pa.ColA[c].GetColID(), fString)
			}

			rowsA = append(rowsA, row)
			cntA++
		}

		for i := range fB {

			if fB[i].IsZero() {
				continue
			}

			row := make([]string, len(pa.ColB))

			for c := range pa.ColB {
				fString := colB[c][i].Text(16)
				if colB[c][i].IsUint64() && colB[c][i].Uint64() < 1000000 {
					fString = colB[c][i].String()
				}
				row[c] = fmt.Sprintf("%v=%v", pa.ColB[c].GetColID(), fString)
			}

			rowsB = append(rowsB, row)
			cntB++
		}

		larger := max(len(rowsA), len(rowsB))

		for i := 0; i < larger; i++ {

			var (
				fa = "* * * * * *"
				fb = "* * * * * *"
			)

			if i < len(rowsA) {
				fa = strings.Join(rowsA[i], " ")
			}

			if i < len(rowsB) {
				fb = strings.Join(rowsB[i], " ")
			}

			fmt.Printf("row=%v %v %v\n", i, fa, fb)
		}

		logrus.Errorf("projection query %v failed", pa.Name)
	}
}
