package discoverer

import (
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/distributed"
	"github.com/consensys/linea-monorepo/prover/protocol/distributed/lpp"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/variables"
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
	// all the  columns involved in the LPP queries, including the verifier columns
	LPPColumns collection.Mapping[ModuleName, []ifaces.Column]
	// all the  columns involved in the GL queries, including the verifier columns
	GLColumns collection.Mapping[ModuleName, []ifaces.Column]
	// all the periodicSamples involved in the GL queries
	PeriodicSamplingGL collection.Mapping[ModuleName, []variables.PeriodicSample]
}

// Analyze first analyzes the simpleDiscoverer and then adds extra analyses based on the queries
// , to get the the verifier, GL , LPP columns and also PeriodicSamplings
func (d QueryBasedDiscoverer) Analyze(comp *wizard.CompiledIOP) {

	d.SimpleDiscoverer.Analyze(comp)

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

					// it analyzes de expression and update d.GLColumns and d.PeriodicSamplingGL
					d.analyzeExprGL(local.Expression, moduleName)

				}
			}

			// get the LPP columns from all the LPP queries, and check if they are in the module
			// this does not contain the new columns from preparation phase like the multiplicity columns.
			lppCols := lpp.GetLPPColumns(comp)
			for _, col := range lppCols {
				if d.SimpleDiscoverer.ColumnIsInModule(col, moduleName) {
					// update d content
					d.LPPColumns.AppendNew(moduleName, []ifaces.Column{col})
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

				d.GLColumns.AppendNew(moduleName, []ifaces.Column{shifted.Parent})

			} else {
				d.GLColumns.AppendNew(moduleName, []ifaces.Column{t})
			}

		case variables.PeriodicSample:

			d.PeriodicSamplingGL.AppendNew(moduleName, []variables.PeriodicSample{t})

		}
	}
}

func (p *QueryBasedDiscoverer) NewSizeOf(ifaces.Column) int {
	panic("unimplemented")
}
