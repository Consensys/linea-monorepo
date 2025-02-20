package discoverer

import (
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/distributed"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils/collection"
)

// struct implementing more complex analysis for [distributed.ModuleDiscoverer]
type QueryBasedDiscoverer struct {
	// simple module discoverer; it does simple analysis
	// it can neither capture the specific columns (like verifier columns or periodicSampling variables),
	// nor categorizes the columns GL and LPP.
	SimpleDiscoverer distributed.ModuleDiscoverer
	// all the  columns involved in the LPP queries,
	// including the verifier columns and new columns from Preparation phase
	LPPColumns collection.Mapping[ModuleName, []ifaces.Column]
	// all the  columns involved in the GL queries, including the verifier columns
	GLColumns collection.Mapping[ModuleName, []ifaces.Column]
}

// Analyze first analyzes the simpleDiscoverer and then adds extra analyses based on the queries
// , to get the the verifier, GL , LPP columns.
func (d *QueryBasedDiscoverer) Analyze(comp *wizard.CompiledIOP, otherComp ...*wizard.CompiledIOP) {

	// initialize discoverer
	d.SimpleDiscoverer.Analyze(comp)
	d.GLColumns = collection.NewMapping[ModuleName, []ifaces.Column]()
	d.LPPColumns = collection.NewMapping[ModuleName, []ifaces.Column]()

	// sanity check
	if len(comp.QueriesParams.AllKeysAt(0)) != 0 {
		panic("At this step, we do not expect a query with parameters")
	}

	// capture and analysis GL Queries
	for _, moduleName := range d.SimpleDiscoverer.ModuleList() {
		for _, q := range comp.QueriesNoParams.AllKeys() {

			if global, ok := comp.QueriesNoParams.Data(q).(query.GlobalConstraint); ok {

				if d.SimpleDiscoverer.ExpressionIsInModule(global.Expression, moduleName) {

					// it analyzes the expression and update d.GLColumns and d.PeriodicSamplingGL
					d.analyzeExprGL(global.Expression, moduleName)

				}
			}

			// the same for local queries.
			if local, ok := comp.QueriesNoParams.Data(q).(query.LocalConstraint); ok {

				if d.SimpleDiscoverer.ExpressionIsInModule(local.Expression, moduleName) {

					// it analyzes de expression and update d.GLColumns
					d.analyzeExprGL(local.Expression, moduleName)

				}
			}
		}

		// the rest is column based analysis, for LPP columns
		if len(otherComp) == 1 {
			// get the LPP columns
			otherCols := otherComp[0].Columns.AllHandlesAtRound(0)
			for _, col := range otherCols {
				if d.SimpleDiscoverer.ColumnIsInModule(col, moduleName) {
					// update d content
					AppendNew(&d.LPPColumns, moduleName, col)
				}
			}
		}

	}

}

// ModuleList returns the list of module names
func (d *QueryBasedDiscoverer) ModuleList() []ModuleName {

	return d.SimpleDiscoverer.ModuleList()
}

// FindModule finds the module name for a given column
func (d *QueryBasedDiscoverer) FindModule(col ifaces.Column) ModuleName {
	return d.SimpleDiscoverer.FindModule(col)
}

// QueryIsInModule checks if the given query is inside the given module
func (d *QueryBasedDiscoverer) QueryIsInModule(q ifaces.Query, moduleName ModuleName) bool {
	return d.SimpleDiscoverer.QueryIsInModule(q, moduleName)
}

// ColumnIsInModule checks that the given column is inside the given module.
func (d *QueryBasedDiscoverer) ColumnIsInModule(col ifaces.Column, name ModuleName) bool {
	return d.SimpleDiscoverer.ColumnIsInModule(col, name)
}

//	ExpressionIsInModule checks that all the columns  (except verifiercol) in the expression are from the given module.
//
// It does not check the presence of the coins and other metadata in the module.
// the restriction over verifier column comes from the fact that the discoverer Analyses compiledIOP and the verifier columns are not accessible there.
func (p *QueryBasedDiscoverer) ExpressionIsInModule(expr *symbolic.Expression, name ModuleName) bool {
	return p.SimpleDiscoverer.ExpressionIsInModule(expr, name)
}

// analyzeExpr analyzes the expression and update the content of d.
func (d *QueryBasedDiscoverer) analyzeExprGL(expr *symbolic.Expression, moduleName ModuleName) {

	var (
		board    = expr.Board()
		metadata = board.ListVariableMetadata()
	)

	for _, m := range metadata {

		switch t := m.(type) {
		case ifaces.Column:

			if shifted, ok := t.(column.Shifted); ok {
				AppendNew(&d.GLColumns, moduleName, shifted.Parent)

			} else {
				AppendNew(&d.GLColumns, moduleName, t)
			}
		}
	}
}

func AppendNew(myMap *collection.Mapping[ModuleName, []ifaces.Column], name ModuleName, myCol ifaces.Column) {
	if !myMap.Exists(name) {
		myMap.InsertNew(name, []ifaces.Column{myCol})
		return
	}

	allCols := myMap.MustGet(name)
	for _, col := range allCols {
		if col.GetColID() == myCol.GetColID() {
			return
		}
	}

	allCols = append(allCols, myCol)
	myMap.Update(name, allCols)

}
