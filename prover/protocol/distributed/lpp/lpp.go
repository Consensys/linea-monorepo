package lpp

import (
	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

// It applies the Compilation steps concerning the LPP queries over comp.
// It generates a LPP-CompiledIOP object internally, that is used for seed generation.
func CompileLPPAndGetSeed(comp *wizard.CompiledIOP, lppCompilers ...func(*wizard.CompiledIOP)) *wizard.CompiledIOP {

	var (
		lppComp    = wizard.NewCompiledIOP()
		oldColumns = []ifaces.Column{}
		lppCols    = []ifaces.Column{}
	)

	// get the LPP columns from comp.
	lppCols = append(lppCols, getLPPColumns(comp)...)

	for _, col := range comp.Columns.AllHandlesAtRound(0) {
		oldColumns = append(oldColumns, col)
	}

	// applies lppCompiler; this  would add a new round and probably new columns to the current round
	//  but no new column to the new round.
	for _, lppCompiler := range lppCompilers {
		lppCompiler(comp)

		if comp.NumRounds() != 2 || comp.Columns.NumRounds() != 1 {
			panic("we expect to have new round while no column is yet registered for the new round")
		}

		numRounds := comp.NumRounds()
		comp.EqualizeRounds(numRounds)
	}

	// filter the new lpp columns.
	for _, col := range comp.Columns.AllHandlesAtRound(0) {
		isOld := false
		for _, oldCol := range oldColumns {
			if col.GetColID() == oldCol.GetColID() {
				isOld = true
				break
			}
		}
		if !isOld {
			// if it is not in the oldColumns it is a new lpp column.
			lppCols = append(lppCols, col)
		}
	}

	// add the LPP columns to the lppComp.
	for _, col := range lppCols {
		lppComp.InsertCommit(0, col.GetColID(), col.Size())
	}

	// register the seed, generated from LPP, in comp
	// for the sake of the assignment it also should be registered in lppComp
	lppComp.InsertCoin(1, "SEED", coin.Field)
	comp.InsertCoin(1, "SEED", coin.Field)

	// prepare and register prover actions.
	lppProver := &lppProver{
		cols: lppCols,
	}

	lppComp.RegisterProverAction(1, lppProver)

	return lppComp

}

type lppProver struct {
	cols []ifaces.Column
}

func (p *lppProver) Run(run *wizard.ProverRuntime) {

	for _, col := range p.cols {
		colWitness := run.ParentRuntime.GetColumn(col.GetColID())
		run.AssignColumn(col.GetColID(), colWitness, col.Round())
	}

	// generate the seed based on LPP run time.
	seed := run.GetRandomCoinField("SEED")

	// pass the seed to the parent run time.
	// note that the parent of LPP is also parent to all compiledIOP of segment-Modules.
	// thus, this gives access to the seed for all segment-module-compiledIOPs.
	run.ParentRuntime.Coins.InsertNew("SEED", seed)
}

// GetLPPComp take the and old CompiledIOP object.
// It creates a fresh CompiledIOP object holding only the LPP columns.
// old CompiledIOP includes the LPP queries and new LPP Columns includes the new columns generated at round 0,
// due to the application of a compilation step (i.e., during the preparation).
// for example : multiplicity columns, for inclusion query, are retrieved from new LPP columns.
func GetLPPComp(oldComp *wizard.CompiledIOP, newLPPCols []ifaces.Column) *wizard.CompiledIOP {

	var (
		// initialize LPPComp
		lppComp = wizard.NewCompiledIOP()
		lppCols = []ifaces.Column{}
	)

	// get the LPP columns
	lppCols = append(lppCols, getLPPColumns(oldComp)...)
	lppCols = append(lppCols, newLPPCols...)

	for _, col := range lppCols {
		lppComp.InsertCommit(0, col.GetColID(), col.Size())
	}
	return lppComp
}

// it extract LPP columns from the context of each LPP query.
func getLPPColumns(c *wizard.CompiledIOP) []ifaces.Column {

	var (
		lppColumns = []ifaces.Column{}
	)

	for _, qName := range c.QueriesNoParams.AllKeysAt(0) {
		q := c.QueriesNoParams.Data(qName)
		switch v := q.(type) {
		case query.Inclusion:

			for i := range v.Including {
				lppColumns = append(lppColumns, v.Including[i]...)
			}

			lppColumns = append(lppColumns, v.Included...)

			if v.IncludingFilter != nil {
				lppColumns = append(lppColumns, v.IncludingFilter...)
			}

			if v.IncludedFilter != nil {
				lppColumns = append(lppColumns, v.IncludedFilter)
			}

		case query.Permutation:
			for i := range v.A {
				lppColumns = append(lppColumns, v.A[i]...)
				lppColumns = append(lppColumns, v.B[i]...)
			}
		case query.Projection:
			lppColumns = append(lppColumns, v.ColumnsA...)
			lppColumns = append(lppColumns, v.ColumnsB...)
			lppColumns = append(lppColumns, v.FilterA, v.FilterB)

		default:
			//do noting
		}

	}

	return lppColumns
}
