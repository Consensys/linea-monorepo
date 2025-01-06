package distributed

import (
	sv "github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/common/vector"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/innerproduct"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/lookup"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/mimc"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/specialqueries"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/protocol/wizardutils"
	"github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
)

// prepare reduces any query to LPP or GL.
// For Inclusion query, it push the compilation one step further:
// all the inclusion queries are compiled into a LogDarivativeSum query,
// This is required due to the challenge that table M depends on whole the witness.
func prepare(comp *wizard.CompiledIOP) {
	mimc.CompileMiMC(comp)
	specialqueries.RangeProof(comp)
	specialqueries.CompileFixedPermutations(comp)
	innerproduct.Compile(comp)

	IntoLogDerivativeSum(comp)
}

// IIntoLogDerivativeSum compiles  all the inclusion queries to a single LogDerivativeSum query that is ready for the split.
// This step is necessary for inclusion,
// as the M table depends on the whole witness and so can not be handled modules-wise without changing the API of WizardIOP.
func IntoLogDerivativeSum(comp *wizard.CompiledIOP) {

	var (
		mainLookupCtx = lookup.CaptureLookupTables(comp)
		lastRound     = comp.NumRounds() - 1
		// zCatalog stores a mapping (round, size) into query.LogDerivativeSumInput and helps finding
		// which Z context should be used to handle a part of a given inclusion
		// query.
		zCatalog = map[[2]int]*query.LogDerivativeSumInput{}
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
			lookupTableName = lookup.NameTable(lookupTable)
			checkTable      = mainLookupCtx.CheckedTables[lookupTableName]
			round           = mainLookupCtx.Rounds[lookupTableName]
			includedFilters = mainLookupCtx.IncludedFilters[lookupTableName]
			// collapse multiColumns to single Columns and commit to M.
			tableCtx = lookup.CompileLookupTable(comp, round, lookupTable, checkTable, includedFilters)
		)

		// push single-columns into zCatalog
		PushToZCatalog(tableCtx, zCatalog)

		a := lookup.MAssignmentTask{
			M:       tableCtx.M,
			S:       checkTable,
			T:       lookupTable,
			SFilter: includedFilters,
		}
		comp.SubProvers.AppendToInner(round, a.Run)

	}

	// insert a single LogDerivativeSum query for the global zCatalog.
	comp.InsertLogDerivativeSum(lastRound, "GlobalLogDerivativeSum", zCatalog)
}

// PushToZCatalog constructs the numerators and denominators for the collapsed S and T
// into zCatalog, for their corresponding rounds and size.
func PushToZCatalog(stc lookup.SingleTableCtx, zCatalog map[[2]int]*query.LogDerivativeSumInput) {

	var (
		round = stc.Gamma.Round
	)

	// tableCtx push to -> zCtx
	// Process the T columns
	for frag := range stc.T {
		size := stc.M[frag].Size()

		key := [2]int{round, size}
		if zCatalog[key] == nil {
			zCatalog[key] = &query.LogDerivativeSumInput{
				Size:  size,
				Round: round,
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

		key := [2]int{round, size}
		if zCatalog[key] == nil {
			zCatalog[key] = &query.LogDerivativeSumInput{
				Size:  size,
				Round: round,
			}
		}

		zCtxEntry := zCatalog[key]
		zCtxEntry.Numerator = append(zCtxEntry.Numerator, sFilter)
		zCtxEntry.Denominator = append(zCtxEntry.Denominator, symbolic.Add(stc.Gamma, stc.S[table]))
	}
}

// PrepareMAssignment outputs a list mapping the ID of a M column to its witness.
func PrepareMAssignment(a lookup.MAssignmentTask, run ifaces.Runtime) map[ifaces.ColID]*sv.Regular {

	var (
		// isMultiColumn flags whether the table have multiple column and
		// whether the "collapsing" trick is needed.
		isMultiColumn = len(a.T[0]) > 1

		// tCollapsed contains either the assignment of T if the table is a
		// single column (e.g. isMultiColumn=false) or its collapsed version
		// otherwise.
		tCollapsed = make([]sv.SmartVector, len(a.T))

		// sCollapsed contains either the assignment of the Ss if the table is a
		// single column (e.g. isMultiColumn=false) or their collapsed version
		// otherwise.
		sCollapsed = make([]sv.SmartVector, len(a.S))

		// fragmentUnionSize contains the total number of rows contained in all
		// the fragments of T combined.
		fragmentUnionSize int
	)

	if !isMultiColumn {
		for frag := range a.T {
			tCollapsed[frag] = a.T[frag][0].GetColAssignment(run)
			fragmentUnionSize += a.T[frag][0].Size()
		}

		for i := range a.S {
			sCollapsed[i] = a.S[i][0].GetColAssignment(run)
		}
	}

	if isMultiColumn {
		// collapsingRandomness is the randomness used in the collapsing trick.
		// It is sampled via `crypto/rand` internally to ensure it cannot be
		// predicted ahead of time by an adversary.
		var collapsingRandomness field.Element
		if _, err := collapsingRandomness.SetRandom(); err != nil {
			utils.Panic("could not sample the collapsing randomness: %v", err.Error())
		}

		for frag := range a.T {
			tCollapsed[frag] = column.RandLinCombColAssignment(run, collapsingRandomness, a.T[frag])
		}

		for i := range a.S {
			sCollapsed[i] = column.RandLinCombColAssignment(run, collapsingRandomness, a.S[i])
		}
	}

	var (
		// m  is associated with tCollapsed
		// m stores the assignment to the column M as we build it.
		m = make([][]field.Element, len(a.T))

		// mapm collects the entries in the inclusion set to their positions
		// in tCollapsed. If T contains duplicates, the last position is the
		// one that is kept in mapM.
		//
		// It is used to let us know where an entry of S appears in T. The stored
		// 2-uple of integers indicate [fragment, row]
		mapM = make(map[field.Element][2]int, fragmentUnionSize)

		// one stores a reference to the field element equals to 1 for
		// convenience so that we can use pointer on it directly.
		one = field.One()
	)

	// This loops initializes mapM so that it tracks to the positions of the
	// entries of T. It also preinitializes the values of ms
	for frag := range a.T {
		m[frag] = make([]field.Element, tCollapsed[frag].Len())
		for k := 0; k < tCollapsed[frag].Len(); k++ {
			mapM[tCollapsed[frag].Get(k)] = [2]int{frag, k}
		}
	}

	// This loops counts all the occurences of the rows of T within S and store
	// them into S.
	for i := range sCollapsed {

		var (
			hasFilter = a.SFilter[i] != nil
			filter    []field.Element
		)

		if hasFilter {
			filter = a.SFilter[i].GetColAssignment(run).IntoRegVecSaveAlloc()
		}

		for k := 0; k < sCollapsed[i].Len(); k++ {

			if hasFilter && filter[k].IsZero() {
				continue
			}

			if hasFilter && !filter[k].IsOne() {
				utils.Panic(
					"the filter column `%v` has a non-binary value at position `%v`: (%v)",
					a.SFilter[i].GetColID(),
					k,
					filter[k].String(),
				)
			}

			var (
				// v stores the entry of S that we are examining and looking for
				// in the look up table.
				v = sCollapsed[i].Get(k)

				// posInM stores the position of `v` in the look-up table
				posInM, ok = mapM[v]
			)

			if !ok {
				tableRow := make([]field.Element, len(a.S[i]))
				for j := range tableRow {
					tableRow[j] = a.S[i][j].GetColAssignmentAt(run, k)
				}
				utils.Panic(
					"entry %v of the table %v is not included in the table. tableRow=%v",
					k, lookup.NameTable([][]ifaces.Column{a.S[i]}), vector.Prettify(tableRow),
				)
			}

			mFrag, posInFragM := posInM[0], posInM[1]
			m[mFrag][posInFragM].Add(&m[mFrag][posInFragM], &one)
		}

	}

	for frag := range m {
		mList[a.M[frag].GetColID()] = sv.NewRegular(m[frag])
	}
	return mList
	/*
		 a more professional way:
			Pass all the prover message columns as part of the proof

			messages := collection.NewMapping[ifaces.ColID, ifaces.ColAssignment]()

			for _, name := range runtime.Spec.Columns.AllKeysProof() {
				messageValue := runtime.Columns.MustGet(name)
				messages.InsertNew(name, messageValue)
		}
	*/
}

var mList map[ifaces.ColID]*sv.Regular
