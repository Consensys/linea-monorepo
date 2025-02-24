package logderivativesum

import (
	"fmt"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/column/verifiercol"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/protocol/wizardutils"
	"github.com/consensys/linea-monorepo/prover/symbolic"
)

// table is an alias for a list of column. We use it in the scope of the lookup
// compiler as a shorthand to make the code more eye-parseable.
type table = []ifaces.Column

// LookupIntoLogDerivativeSum compiles  all the inclusion queries to a single
// LogDerivativeSum query without compiling it immediately. This compiler
// is used by the distributed wizard protocol feature as it allows to
// prepare the lookup queries to be split across several wizard-IOP
// in such a way that we can recombine them later.
func LookupIntoLogDerivativeSum(comp *wizard.CompiledIOP) {

	var (
		mainLookupCtx = CaptureLookupTables(comp)
		lastRound     = comp.NumRounds() - 1
		// zCatalog stores a mapping (round, size) into query.LogDerivativeSumInput and helps finding
		// which Z context should be used to handle a part of a given inclusion
		// query.
		zCatalog    = map[int]*query.LogDerivativeSumInput{}
		proverTasks = make([]proverTaskAtRound, lastRound+1)
	)

	// Skip the compilation phase if no lookup constraint is being used. Otherwise
	// it will register a verifier action that is not required and will be bugged.
	if len(mainLookupCtx.LookupTables) == 0 {
		return
	}

	// Step 1. construct the "per table" contexts and pack the Sigma's into
	// zCatalog.
	for _, lookupTable := range mainLookupCtx.LookupTables {

		var (
			// get checkedTables, rounds, Filters by lookupTableName
			lookupTableName = NameTable(lookupTable)
			checkTable      = mainLookupCtx.CheckedTables[lookupTableName]
			round           = mainLookupCtx.Rounds[lookupTableName]
			includedFilters = mainLookupCtx.IncludedFilters[lookupTableName]
			// collapse multiColumns to single Columns and commit to M.
			tableCtx = CompileLookupTable(comp, round, lookupTable, checkTable, includedFilters)
		)

		// push single-columns into zCatalog
		pushToZCatalog(tableCtx, zCatalog)

		// assign the multiplicity column
		proverTasks[round].pushMAssignment(MAssignmentTask{
			M:       tableCtx.M,
			S:       checkTable,
			T:       lookupTable,
			SFilter: includedFilters,
		})
	}

	for round, task := range proverTasks {
		if task.numTasks() > 0 {
			comp.RegisterProverAction(round, task)
		}
	}

	// insert a single LogDerivativeSum query for the global zCatalog.
	q := comp.InsertLogDerivativeSum(lastRound+1, "GlobalLogDerivativeSum", zCatalog)

	// assign parameters of LogDerivativeSum, it is just to prevent the panic attack in the prover
	comp.SubProvers.AppendToInner(lastRound+1, func(run *wizard.ProverRuntime) {
		run.AssignLogDerivSum("GlobalLogDerivativeSum", field.Zero())
	})

	// the verifier checks that the log-derivative sum result is zeroo
	comp.RegisterVerifierAction(lastRound+1, &CheckLogDerivativeSumMustBeZero{
		Q: q,
	})

}

// pushToZCatalog constructs the numerators and denominators for the collapsed S and T
// into zCatalog, for their corresponding rounds and size.
func pushToZCatalog(stc SingleTableCtx, zCatalog map[int]*query.LogDerivativeSumInput) {

	// tableCtx push to -> zCtx
	// Process the T columns
	for frag := range stc.T {
		size := stc.M[frag].Size()

		key := size
		if zCatalog[key] == nil {
			zCatalog[key] = &query.LogDerivativeSumInput{
				Size: size,
			}
		}

		zCtxEntry := zCatalog[key]
		zCtxEntry.Numerator = append(zCtxEntry.Numerator, symbolic.Neg(stc.M[frag])) // no functions for num, denom here
		zCtxEntry.Denominator = append(zCtxEntry.Denominator, symbolic.Add(stc.Gamma, stc.T[frag]))
	}

	// Process the S columns
	for table := range stc.S {
		var (
			_, _, size = wizardutils.AsExpr(stc.S[table])
			sFilter    = symbolic.NewConstant(1)
		)

		if stc.SFilters[table] != nil {
			sFilter = symbolic.NewVariable(stc.SFilters[table])
		}

		key := size
		if zCatalog[key] == nil {
			zCatalog[key] = &query.LogDerivativeSumInput{
				Size: size,
			}
		}

		zCtxEntry := zCatalog[key]
		zCtxEntry.Numerator = append(zCtxEntry.Numerator, sFilter)
		zCtxEntry.Denominator = append(zCtxEntry.Denominator, symbolic.Add(stc.Gamma, stc.S[table]))
	}
}

// CheckLogDerivativeSumMustBeZero is an implementation of the [wizard.VerifierAction] interface.
// It checks that the log-derivative sum result is zero.
type CheckLogDerivativeSumMustBeZero struct {
	Q       query.LogDerivativeSum
	skipped bool
}

func (c *CheckLogDerivativeSumMustBeZero) Run(run wizard.Runtime) error {
	y := run.GetLogDerivSumParams(c.Q.ID).Sum
	if !y.IsZero() {
		return fmt.Errorf("log-derivate sum; the final evaluation check failed for %v\n"+
			"expected '0' but calculated %v,",
			c.Q.ID, y.String())
	}
	return nil
}

func (c *CheckLogDerivativeSumMustBeZero) RunGnark(api frontend.API, run wizard.GnarkRuntime) {
	y := run.GetLogDerivSumParams(c.Q.ID).Sum
	api.AssertIsEqual(y, 0)
}

func (c *CheckLogDerivativeSumMustBeZero) Skip() {
	c.skipped = true
}

func (c *CheckLogDerivativeSumMustBeZero) IsSkipped() bool {
	return c.skipped
}

// CaptureLookupTables inspects comp and look for Inclusion queries that are not
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
func CaptureLookupTables(comp *wizard.CompiledIOP) mainLookupCtx {

	ctx := mainLookupCtx{
		LookupTables:    [][]table{},
		CheckedTables:   map[string][]table{},
		IncludedFilters: map[string][]ifaces.Column{},
		Rounds:          map[string]int{},
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
			checkedTable, lookupTable = GetTableCanonicalOrder(lookup)
			tableName                 = NameTable(lookupTable)
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

			tableName = NameTable(lookupTable)
		}

		if lookup.IsFilteredOnIncluded() {
			includedFilter = lookup.IncludedFilter
		}

		// In case this is the first iteration where we encounter the lookupTable
		// we need to add entries in the registering maps.
		if _, ok := ctx.CheckedTables[tableName]; !ok {
			ctx.IncludedFilters[tableName] = []ifaces.Column{}
			ctx.CheckedTables[tableName] = []table{}
			ctx.LookupTables = append(ctx.LookupTables, lookupTable)
			ctx.Rounds[tableName] = 0
		}

		ctx.IncludedFilters[tableName] = append(ctx.IncludedFilters[tableName], includedFilter)
		ctx.CheckedTables[tableName] = append(ctx.CheckedTables[tableName], checkedTable)
		ctx.Rounds[tableName] = max(ctx.Rounds[tableName], comp.QueriesNoParams.Round(lookup.ID))

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

// here we are looking up set of columns S in a single column T
func CompileLookupTable(
	comp *wizard.CompiledIOP,
	round int,
	lookupTable []table,
	checkedTables []table,
	includedFilters []ifaces.Column,
) (ctx SingleTableCtx) {

	ctx = SingleTableCtx{
		TableName: NameTable(lookupTable),
		S:         make([]*symbolic.Expression, len(checkedTables)),
		SFilters:  includedFilters,
		T:         make([]*symbolic.Expression, len(lookupTable)),
		M:         make([]ifaces.Column, len(lookupTable)),
	}

	var (
		// isMultiColumn indicates whether the lookup table (and thus the
		// checked tables) have the same number of
		isMultiColumn = len(lookupTable[0]) > 1
	)

	if !isMultiColumn {
		for frag := range ctx.T {
			ctx.T[frag] = symbolic.NewVariable(lookupTable[frag][0])
			ctx.M[frag] = comp.InsertCommit(
				round,
				DeriveTableNameWithIndex[ifaces.ColID](LogDerivativePrefix, lookupTable, frag, "M"),
				lookupTable[frag][0].Size(),
			)

		}

		for i := range ctx.S {
			ctx.S[i] = symbolic.NewVariable(checkedTables[i][0])
		}
	}

	if isMultiColumn {

		// alpha is the coin used to compute the linear combination of the
		// columns of T and S when they are (both) multi-columns.
		alpha := comp.InsertCoin(
			round+1,
			DeriveTableName[coin.Name](LogDerivativePrefix, lookupTable, "ALPHA"),
			coin.Field,
		)

		for frag := range ctx.T {
			ctx.T[frag] = wizardutils.RandLinCombColSymbolic(alpha, lookupTable[frag])
			ctx.M[frag] = comp.InsertCommit(
				round,
				DeriveTableNameWithIndex[ifaces.ColID](LogDerivativePrefix, lookupTable, frag, "M"),
				lookupTable[frag][0].Size(),
			)
		}

		for i := range ctx.S {
			ctx.S[i] = wizardutils.RandLinCombColSymbolic(alpha, checkedTables[i])
		}
	}

	ctx.Gamma = comp.InsertCoin(
		round+1,
		DeriveTableName[coin.Name](LogDerivativePrefix, lookupTable, "GAMMA"),
		coin.Field,
	)

	return ctx
}
