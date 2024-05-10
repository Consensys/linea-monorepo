package lookup

import (
	"github.com/consensys/zkevm-monorepo/prover/maths/field"
	"github.com/consensys/zkevm-monorepo/prover/protocol/coin"
	"github.com/consensys/zkevm-monorepo/prover/protocol/column"
	"github.com/consensys/zkevm-monorepo/prover/protocol/column/verifiercol"
	"github.com/consensys/zkevm-monorepo/prover/protocol/ifaces"
	"github.com/consensys/zkevm-monorepo/prover/protocol/query"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizardutils"
	"github.com/consensys/zkevm-monorepo/prover/symbolic"
)

// table is an alias for a list of column. We use it in the scope of the lookup
// compiler as a shorthand to make the code more eye-parseable.
type table = []ifaces.Column

// CompileLogDerivative scans `comp“, looking for Inclusion queries and compile
// them using the LogDerivativeLookup technique. The compiler attempt to group
// queries relating to the same table. This allows saving in commitment because
// when grouping is possible, then we only need to commit to a single
// extract the table and the checked from the lookup query and ensures that the
// table are in canonical order. That is because we want to group lookups into
// the same columns but not in the same order.
func CompileLogDerivative(comp *wizard.CompiledIOP) {

	var (
		mainLookupCtx = captureLookupTables(comp)
		proverActions = make([]proverTaskAtRound, comp.NumRounds()+1)
	)

	for _, lookupTable := range mainLookupCtx.lookupTables {

		var (
			lookupTableName = nameTable(lookupTable)
			checkTable      = mainLookupCtx.checkedTables[lookupTableName]
			round           = mainLookupCtx.rounds[lookupTableName]
			includedFilters = mainLookupCtx.includedFilters[lookupTableName]
			tableCtx        = compileLookupTable(comp, round, lookupTable, checkTable, includedFilters)
		)

		proverActions[round].pushMAssignment(
			mAssignmentTask{
				M:       tableCtx.M,
				S:       checkTable,
				T:       lookupTable,
				SFilter: includedFilters,
			},
		)

		proverActions[round+1].pushSigmaAssignment(sigmaAssignmentTask(tableCtx))
	}

	for round := range proverActions {
		// It would not be a bugged to include a proverAction that does nothing
		// but this pollutes the performance analysis of the prover and logs.
		if proverActions[round].numTasks() > 0 {
			comp.RegisterProverAction(round, proverActions[round])
		}
	}
}

// captureLookupTables inspects comp and look for Inclusion queries that are not
// marked as ignored yet. All the queries matched queries are grouped by look-up
// table (e.g. all the queries that use the same lookup table). All the matched
// queries are marked as ignored. The function returns the thereby-initialized
// [mainLookupCtx].
//
// This step does not directly mutate the wizard (apart from marking the queries
// as ignored) and it prepares the next table-by-table compilation step.
//
// The function also implictly reduces the conditionals over the Including table
// be appending a "one" column on the included side and the filter on the
// including side.
func captureLookupTables(comp *wizard.CompiledIOP) mainLookupCtx {

	ctx := mainLookupCtx{
		lookupTables:    [][]table{},
		checkedTables:   map[string][]table{},
		includedFilters: map[string][]ifaces.Column{},
		rounds:          map[string]int{},
	}

	// Collect all the lookup queries into "lookups"
	for _, qName := range comp.QueriesNoParams.AllUnignoredKeys() {

		// Filter out non lookup queries
		lookup, ok := comp.QueriesNoParams.Data(qName).(query.Inclusion)
		if !ok {
			continue
		}

		// This ensures that the lookup query is not used again in the
		// compilation process. We know that the query was already ignored at
		// the beginning because we are iterating over the unignored keys.
		comp.QueriesNoParams.MarkAsIgnored(qName)

		var (
			// checkedTable corresponds to the "included" table and lookupTable
			// corresponds to the including table.
			checkedTable, lookupTable = getTableCanonicalOrder(lookup)
			tableName                 = nameTable(lookupTable)
			// includedFilters stores the query.IncludedFilter parameter. If the
			// query has no includedFilters on the Included side. Then this is
			// left as nil.
			includedFilter ifaces.Column
		)

		if lookup.IsFilteredOnIncluding() {
			var (
				checkedLen = checkedTable[0].Size()
				ones       = verifiercol.NewConstantCol(field.One(), checkedLen)
			)

			checkedTable = append([]ifaces.Column{ones}, checkedTable...)
			for frag := range lookupTable {
				lookupTable[frag] = append([]ifaces.Column{lookup.IncludingFilter[frag]}, lookupTable[frag]...)
			}

			tableName = nameTable(lookupTable)
		}

		if lookup.IsFilteredOnIncluded() {
			includedFilter = lookup.IncludedFilter
		}

		// In case this is the first iteration where we encounter the lookupTable
		// we need to add entries in the registering maps.
		if _, ok := ctx.checkedTables[tableName]; !ok {
			ctx.includedFilters[tableName] = []ifaces.Column{}
			ctx.checkedTables[tableName] = []table{}
			ctx.lookupTables = append(ctx.lookupTables, lookupTable)
			ctx.rounds[tableName] = 0
		}

		ctx.includedFilters[tableName] = append(ctx.includedFilters[tableName], includedFilter)
		ctx.checkedTables[tableName] = append(ctx.checkedTables[tableName], checkedTable)
		ctx.rounds[tableName] = max(ctx.rounds[tableName], comp.QueriesNoParams.Round(lookup.ID))

	}

	return ctx
}

// compileLookupTable applies the log-derivative lookup compilation context to
// the supplied table. round denotes the interaction round in which to start the
// compilation.
//
// It registers the following queries
//   - (1) The verifier queries that $\sum_{k=0\ldots n-1} (\Sigma_{S,k})[|S_k| - 1] == (\Sigma_T)[|T| - 1]$. Namely, the sum of the last entry of all $\Sigma_{S,k}$ equals the last entry of $\Sigma_T$
//   - (2) **(For all k)** the verifier makes a `Local` query : $(\Sigma_{S,k})[0] = \frac{1}{S_{k,0} + \gamma}$
//   - (3) The verifier makes a `Local` query : $(\Sigma_T)[0] = \frac{M_0}{T_0 + \gamma}$
//   - (4) **(For all k)** The verifier makes a `Global` query : $\left((\Sigma_{S,k})[i] - (\Sigma_{S,k})[i-1]\right)(S_{k,i} + \gamma) = 1$
//   - (5) The verier makes a `Global` query : $\left((\Sigma_T)[i] - (\Sigma_T)[i-1]\right)(T_i + \gamma) = M_i$
func compileLookupTable(
	comp *wizard.CompiledIOP,
	round int,
	lookupTable []table,
	checkedTables []table,
	includedFilters []ifaces.Column,
) (ctx singleTableCtx) {

	ctx = singleTableCtx{
		TableName:      nameTable(lookupTable),
		S:              make([]*symbolic.Expression, len(checkedTables)),
		SBoard:         make([]symbolic.ExpressionBoard, len(checkedTables)),
		SigmaS:         make([]ifaces.Column, len(checkedTables)),
		SigmaSOpenings: make([]query.LocalOpening, len(checkedTables)),
		SFilters:       includedFilters,
		T:              make([]*symbolic.Expression, len(lookupTable)),
		M:              make([]ifaces.Column, len(lookupTable)),
		TBoard:         make([]symbolic.ExpressionBoard, len(lookupTable)),
		SigmaT:         make([]ifaces.Column, len(lookupTable)),
		SigmaTOpening:  make([]query.LocalOpening, len(lookupTable)),
	}

	var (
		// isMultiColumn indicates whether the lookup table (and thus the
		// checked tables) have the same number of
		isMultiColumn = len(lookupTable) > 1
	)

	if !isMultiColumn {
		for frag := range ctx.T {
			ctx.T[frag] = symbolic.NewVariable(lookupTable[frag][0])
			ctx.TBoard[frag] = ctx.T[frag].Board()
			ctx.M[frag] = comp.InsertCommit(
				round,
				deriveTableNameWithIndex[ifaces.ColID](logDerivativePrefix, lookupTable, frag, "M"),
				lookupTable[frag][0].Size(),
			)

		}

		for i := range ctx.S {
			ctx.S[i] = symbolic.NewVariable(checkedTables[i][0])
			ctx.SBoard[i] = ctx.S[i].Board()
		}
	}

	if isMultiColumn {

		// alpha is the coin used to compute the linear combination of the
		// columns of T and S when they are (both) multi-columns.
		alpha := comp.InsertCoin(
			round+1,
			deriveTableName[coin.Name](logDerivativePrefix, lookupTable, "ALPHA"),
			coin.Field,
		)

		for frag := range ctx.T {
			ctx.T[frag] = wizardutils.RandLinCombColSymbolic(alpha, lookupTable[frag])
			ctx.TBoard[frag] = ctx.T[frag].Board()
			ctx.M[frag] = comp.InsertCommit(
				round,
				deriveTableNameWithIndex[ifaces.ColID](logDerivativePrefix, lookupTable, frag, "M"),
				lookupTable[frag][0].Size(),
			)
		}

		for i := range ctx.S {
			ctx.S[i] = wizardutils.RandLinCombColSymbolic(alpha, checkedTables[i])
			ctx.SBoard[i] = ctx.S[i].Board()
		}
	}

	ctx.Gamma = comp.InsertCoin(
		round+1,
		deriveTableName[coin.Name](logDerivativePrefix, lookupTable, "GAMMA"),
		coin.Field,
	)

	for frag := range ctx.SigmaT {
		ctx.SigmaT[frag] = comp.InsertCommit(
			round+1,
			deriveTableNameWithIndex[ifaces.ColID](logDerivativePrefix, lookupTable, frag, "SIGMA_T"),
			lookupTable[frag][0].Size(),
		)
	}

	for i := range checkedTables {
		ctx.SigmaS[i] = comp.InsertCommit(
			round+1,
			deriveTableNameWithIndex[ifaces.ColID](logDerivativePrefix, lookupTable, i, "SIGMA_S"),
			checkedTables[i][0].Size(),
		)
	}

	for i := range ctx.SigmaS {

		// includedFilter stores the includedFilter[i] or 1 if it is `nil`
		includedFilter := symbolic.NewConstant(1)
		if includedFilters[i] != nil {
			includedFilter = symbolic.NewVariable(includedFilters[i])
		}

		// This registers the query(ies) #4 which ensures the internal consistency
		// of each SigmaS with respect to Alpha and their respective S
		//
		// As a reminder, this enforces
		//
		// 		(SigmaS[k] - SigmaS[k-1])(S[k] + Gamma) == 1
		comp.InsertGlobal(
			round+1,
			deriveTableNameWithIndex[ifaces.QueryID](logDerivativePrefix, lookupTable, i, "SIGMA_S_INTERNAL_CONSISTENCY"),
			symbolic.Sub(
				includedFilter,
				symbolic.Mul(
					symbolic.Sub(ctx.SigmaS[i], column.Shift(ctx.SigmaS[i], -1)),
					symbolic.Add(ctx.S[i], ctx.Gamma),
				),
			),
		)

		// This registers the query #2 which enforces that the summation column
		// SigmaS starts with the correct value of SigmaS = 1 / (S + Gamma)
		comp.InsertLocal(
			round+1,
			deriveTableNameWithIndex[ifaces.QueryID](logDerivativePrefix, lookupTable, i, "SIGMA_S_INIT"),
			symbolic.Sub(
				includedFilter,
				symbolic.Mul(
					ctx.SigmaS[i],
					symbolic.Add(ctx.S[i], ctx.Gamma),
				),
			),
		)

		// This prepares the check #1 which queries the last position of SigmaS.
		// The registered query is then used in the finalEvaluationCheck.
		ctx.SigmaSOpenings[i] = comp.InsertLocalOpening(
			round+1,
			deriveTableNameWithIndex[ifaces.QueryID](logDerivativePrefix, lookupTable, i, "SIGMA_S_FINAL"),
			column.Shift(ctx.SigmaS[i], -1),
		)

	}

	for frag := range ctx.T {

		// This registers the query #5, as a reminder this enforces the internal
		// consistency of SigmaT. Namely, this enforces that:
		//
		// 		(SigmaT[k] - Sigma[k-1])(T[k] + Gamma) == M[k]
		comp.InsertGlobal(
			round+1,
			deriveTableNameWithIndex[ifaces.QueryID](logDerivativePrefix, lookupTable, frag, "SIGMA_T_INTERNAL_CONSISTENCY"),
			symbolic.Sub(
				ctx.M[frag],
				symbolic.Mul(
					symbolic.Sub(ctx.SigmaT[frag], column.Shift(ctx.SigmaT[frag], -1)),
					symbolic.Add(ctx.T[frag], ctx.Gamma),
				),
			),
		)

		// This registers the query #3 which enforces that the summation column
		// SigmaT starts with the correct value of SigmaT = 1 / (T + Gamma)
		comp.InsertLocal(
			round+1,
			deriveTableNameWithIndex[ifaces.QueryID](logDerivativePrefix, lookupTable, frag, "SIGMA_T_INIT"),
			symbolic.Sub(
				ctx.M[frag],
				symbolic.Mul(
					ctx.SigmaT[frag],
					symbolic.Add(ctx.T[frag], ctx.Gamma),
				),
			),
		)

		// This prepares the check #1 which queries the last position of SigmaT.
		// The registered query is then used in the finalEvaluationCheck.
		ctx.SigmaTOpening[frag] = comp.InsertLocalOpening(
			round+1,
			deriveTableNameWithIndex[ifaces.QueryID](logDerivativePrefix, lookupTable, frag, "SIGMA_T_FINAL"),
			column.Shift(ctx.SigmaT[frag], -1),
		)
	}

	// This registers the check #1
	comp.RegisterVerifierAction(round+1, &finalEvaluationCheck{
		SigmaSOpenings: ctx.SigmaSOpenings,
		SigmaTOpening:  ctx.SigmaTOpening,
	})

	return ctx
}
