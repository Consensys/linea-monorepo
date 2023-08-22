package specialqueries

import (
	"fmt"
	"strings"

	sv "github.com/consensys/accelerated-crypto-monorepo/maths/common/smartvectors"
	"github.com/consensys/accelerated-crypto-monorepo/maths/common/vector"
	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/accessors"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/coin"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/column"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/dedicated/expr_handle"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/ifaces"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/query"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/wizard"
	"github.com/consensys/accelerated-crypto-monorepo/symbolic"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
	"github.com/consensys/accelerated-crypto-monorepo/utils/collection"
	"github.com/consensys/gnark/frontend"
)

const (
	/*
		Prefix to indicate an identifier is related to LOGDERIVATIVE
	*/
	LOGDERIVATIVE string = "LOGDERIVATIVE"

	/*
		Names for the intermediates commitments generated in the
		log derivative lookup compiler
	*/
	LOGDERIVATIVE_S_NAME_SUFFIX       string = "S"
	LOGDERIVATIVE_T_NAME_SUFFIX       string = "T"
	LOGDERIVATIVE_M_NAME_SUFFIX       string = "M"
	LOGDERIVATIVE_SIGMA_S_NAME_SUFFIX string = "SIGMA_S"
	LOGDERIVATIVE_SIGMA_T_NAME_SUFFIX string = "SIGMA_T"

	/*
		Suffixes for the coins
	*/
	LOGDERIVATIVE_ALPHA_SUFFIX string = "ALPHA"
	LOGDERIVATIVE_GAMMA_SUFFIX string = "GAMMA"
)

type T []ifaces.Column
type S []ifaces.Column

type logDerivCtx struct {
	// names
	name ifaces.ColID
	//items
	M            ifaces.Column
	alpha, gamma coin.Info
	S            []ifaces.Column
	T            ifaces.Column
	sigmaS       []ifaces.Column
	sigmaT       ifaces.Column
	// queries
	// SIGMA_S_OPENING_QUERY []ifaces.QueryID
	// SIGMA_T_OPENING_QUERY ifaces.QueryID
	S_LOCAL_OPENING []query.LocalOpening
	T_LOCAL_OPENING query.LocalOpening
}

// Derives the identifier's name
func createLogDerivCtx() logDerivCtx {

	res := logDerivCtx{
		S:      []ifaces.Column{},
		sigmaS: []ifaces.Column{},
		// SIGMA_S_OPENING_QUERY: []ifaces.QueryID{},
	}

	return res
}
func LogDerivativeLookupCompiler(comp *wizard.CompiledIOP) {

	// used to iterate in deterministic order over the map
	tables := []T{}
	lookups := map[ifaces.ColID][]S{} // map[T][]S{} with nameTable(T) as key
	rounds := map[ifaces.ColID]int{}  // map[T]int{}  with nameTable(T) as key

	// collect all the lookup queries into "lookups"
	for _, qName := range comp.QueriesNoParams.AllUnignoredKeys() { // qName = lookup

		// filter out non lookup queries
		maybeLookup := comp.QueriesNoParams.Data(qName)
		lookup, ok := maybeLookup.(query.Inclusion)
		if !ok {
			// skip, it's not a lookup query
			continue
		}

		//should never touch this query again
		comp.QueriesNoParams.MarkAsIgnored(qName)

		// extract the table and the checked from the lookup query
		// and ensures that the table are in canonical order. That
		// is because we want to group lookups into the same columns
		// but not in the same order.

		// checked = S = sorted lookup columns included
		// table = T = sorted lookup columns including
		checked, table := getTableCanonicalOrder(lookup)

		// register the table
		isNewTable := appendToLookups(lookups, checked, table) // add table to lookups
		if isNewTable {
			tables = append(tables, table)
		}
		rounds[nameTable(table)] = utils.Max(rounds[nameTable(table)], comp.QueriesNoParams.Round(lookup.ID))
	}

	// meta is a collection of prover steps that can be grouped per rounds and
	// then run in parallel.
	numRounds := comp.NumRounds()
	meta := &metaCtx{provers: collection.NewVecVec[wizard.ProverStep](numRounds)}

	// Then unroll, the batched version of the lookup protocol over each table
	for i := range tables {

		table := tables[i]
		checkeds := lookups[nameTable(table)]
		round := rounds[nameTable(table)]

		ctx := createLogDerivCtx()
		ctx.name = nameTable(table)

		isMultiColumn := len(table) > 1

		if isMultiColumn {

			// declare $\alpha$ responsible for computing the linear combination
			// of the lookup only if it is a multi-column lookup.
			ctx.alpha = comp.InsertCoin(
				round+1,
				deriveTableName[coin.Name](LOGDERIVATIVE, table, LOGDERIVATIVE_ALPHA_SUFFIX),
				coin.Field,
			)

			// declare S_i and T
			for colID := range checkeds {
				s := checkeds[colID]
				ctx.S = append(ctx.S, expr_handle.RandLinCombCol(
					comp,
					accessors.AccessorFromCoin(ctx.alpha),
					checkeds[colID],
					deriveTableNameWithIndex[string](LOGDERIVATIVE, s, colID, LOGDERIVATIVE_S_NAME_SUFFIX),
				))
			}

			ctx.T = expr_handle.RandLinCombCol(
				comp,
				accessors.AccessorFromCoin(ctx.alpha),
				table,
				deriveTableName[string](LOGDERIVATIVE, table, LOGDERIVATIVE_T_NAME_SUFFIX),
			)

		} else {
			ctx.T = table[0]
			for colID := range checkeds {
				ctx.S = append(ctx.S, checkeds[colID][0])
			}
		}

		// declare M
		ctx.M = comp.InsertCommit(
			round,
			deriveTableName[ifaces.ColID](LOGDERIVATIVE, table, LOGDERIVATIVE_M_NAME_SUFFIX),
			table[0].Size())

		// Declare the coin gamma, giving a placeholder name for gamma
		ctx.gamma = comp.InsertCoin(
			round+1,
			deriveTableName[coin.Name](LOGDERIVATIVE, table, LOGDERIVATIVE_GAMMA_SUFFIX),
			coin.Field,
		)

		// declare SIGMA_S and SIGMA_T, giving a placeholder name for sigmas
		for colID := range checkeds {
			s := checkeds[colID]
			ctx.sigmaS = append(ctx.sigmaS, comp.InsertCommit(
				round+1,
				deriveTableNameWithIndex[ifaces.ColID](LOGDERIVATIVE, s, colID, LOGDERIVATIVE_SIGMA_S_NAME_SUFFIX),
				s[0].Size(),
			))
		}

		ctx.sigmaT = comp.InsertCommit(
			round+1,
			deriveTableName[ifaces.ColID](LOGDERIVATIVE, table, LOGDERIVATIVE_SIGMA_T_NAME_SUFFIX),
			table[0].Size())

		/* Queries for log_derivative_lookup (comments from the issue #607)
		* (1) The verifier queries that $\sum_{k=0\ldots n-1} (\Sigma_{S,k})[|S_k| - 1] == (\Sigma_T)[|T| - 1]$. Namely, the sum of the last entry of all $\Sigma_{S,k}$ equals the last entry of $\Sigma_T$
		* (2) **(For all k)** the verifier makes a `Local` query : $(\Sigma_{S,k})[0] = \frac{1}{S_{k,0} + \gamma}$
		* (3) The verifier makes a `Local` query : $(\Sigma_T)[0] = \frac{M_0}{T_0 + \gamma}$
		* (4) **(For all k)** The verifier makes a `Global` query : $\left((\Sigma_{S,k})[i] - (\Sigma_{S,k})[i-1]\right)(S_{k,i} + \gamma) = 1$
		* (5) The verier makes a `Global` query : $\left((\Sigma_T)[i] - (\Sigma_T)[i-1]\right)(T_i + \gamma) = M_i$
		 */

		// Query (4)
		for i := range checkeds {
			sigmaSI := ifaces.ColumnAsVariable(ctx.sigmaS[i])
			sigmaSImin1 := ifaces.ColumnAsVariable(column.Shift(ctx.sigmaS[i], -1))
			sigmaSIminSImin1 := sigmaSI.Sub(sigmaSImin1)
			sKI := ifaces.ColumnAsVariable(ctx.S[i])
			glob4 := sigmaSIminSImin1.Mul(sKI.Add(ctx.gamma.AsVariable())).Sub(symbolic.NewConstant(1))
			comp.InsertGlobal(
				round+1,
				deriveTableNameWithIndex[ifaces.QueryID](LOGDERIVATIVE, ctx.sigmaS, i, "LOGDERIVATIVE_GLOB4_NAME_SUFFIX"),
				glob4,
			)

		}

		// Query (2)
		// Query (1) prep
		ctx.S_LOCAL_OPENING = make([]query.LocalOpening, len(checkeds))
		for i := range checkeds {
			sigmaSI := ifaces.ColumnAsVariable(ctx.sigmaS[i])
			sKI := ifaces.ColumnAsVariable(ctx.S[i])
			loc2 := sigmaSI.Mul(sKI.Add(ctx.gamma.AsVariable())).Sub(symbolic.NewConstant(1))
			comp.InsertLocal(
				round+1,
				deriveTableNameWithIndex[ifaces.QueryID](LOGDERIVATIVE, ctx.sigmaS, i, "LOGDERIVATIVE_LOC2_NAME_SUFFIX"),
				loc2,
			)

			// Query (1) prep
			ctx.S_LOCAL_OPENING[i] = comp.InsertLocalOpening(round+1, ifaces.QueryIDf("%v__S#%v_%v", LOGDERIVATIVE, i, ctx.S[i].GetColID()), column.Shift(ctx.sigmaS[i], -1))
		}

		// Query (3)
		sigmaT := ifaces.ColumnAsVariable(ctx.sigmaT)
		t := ifaces.ColumnAsVariable(ctx.T)
		m := ifaces.ColumnAsVariable(ctx.M)
		loc3 := sigmaT.Mul(t.Add(ctx.gamma.AsVariable())).Sub(m)
		comp.InsertLocal(round+1,
			deriveTableName[ifaces.QueryID](LOGDERIVATIVE, table, "LOGDERIVATIVE_LOC3_NAME_SUFFIX"), loc3)

		// Query (1) prep
		ctx.T_LOCAL_OPENING = comp.InsertLocalOpening(round+1, ifaces.QueryIDf("%v_T_%v", LOGDERIVATIVE, ctx.T.GetColID()), column.Shift(ctx.sigmaT, -1))

		// Query (5)
		sigmaTmin1 := ifaces.ColumnAsVariable(column.Shift(ctx.sigmaT, -1))
		sigmaTminSigmaTmin1 := sigmaT.Sub(sigmaTmin1)
		glob5 := sigmaTminSigmaTmin1.Mul(t.Add(ctx.gamma.AsVariable())).Sub(m)
		comp.InsertGlobal(
			round+1,
			deriveTableName[ifaces.QueryID](LOGDERIVATIVE, table, "LOGDERIVATIVE_GLOB5_NAME_SUFFIX"),
			glob5,
		)

		// Step responsible for assigning "M"
		meta.provers.AppendToInner(round, func(run *wizard.ProverRuntime) {

			var tWitnessCollapsed sv.SmartVector
			sWitnessesCollapsed := make([]sv.SmartVector, len(checkeds))

			// compute M ... we define a map with key as table values and value as their number of occurrences in all Sks
			if !isMultiColumn {
				// single column case - take S and T directly to compute M
				tWitnessCollapsed = ctx.T.GetColAssignment(run)
				for i := range checkeds {
					sWitnessesCollapsed[i] = ctx.S[i].GetColAssignment(run)
				}
			} else {
				// multi-column case : S and T are not defined (it will be in the next round)
				var collapseRandomness field.Element
				collapseRandomness.SetRandom()

				tWitnessCollapsed = sv.NewConstant(field.Zero(), ctx.T.Size())
				x := field.One()

				for tableCol := range table {
					colTableWit := table[tableCol].GetColAssignment(run)
					tWitnessCollapsed = sv.Add(tWitnessCollapsed, sv.Mul(colTableWit, sv.NewConstant(x, ctx.T.Size())))
					x.Mul(&x, &collapseRandomness)
				}

				for sID := range sWitnessesCollapsed {
					sWitnessesCollapsed[sID] = sv.NewConstant(field.Zero(), ctx.S[sID].Size())
					x := field.One()

					for tableCol := range checkeds[sID] {
						colTableWit := checkeds[sID][tableCol].GetColAssignment(run)
						sWitnessesCollapsed[sID] = sv.Add(sWitnessesCollapsed[sID], sv.Mul(colTableWit, sv.NewConstant(x, ctx.S[sID].Size())))
						x.Mul(&x, &collapseRandomness)
					}
				}
			}

			tWitnessCollapsedArr := sv.IntoRegVec(tWitnessCollapsed)
			M := make([]field.Element, len(tWitnessCollapsedArr))

			// First collect the entries in the inclusion set
			mapm := make(map[field.Element]int, len(tWitnessCollapsedArr))
			for i, v := range tWitnessCollapsedArr {
				// Set of acceptable values
				mapm[v] = i
			}

			// Then map each value of S_i to a position in M to increment
			for sID := range checkeds {
				sWitness := sWitnessesCollapsed[sID]
				sWitnessArr := sv.IntoRegVec(sWitness)
				oneValue := field.Element(field.One())
				for k, v := range sWitnessArr {
					posInM, ok := mapm[v]
					if !ok {
						tableRow := make([]field.Element, len(checkeds[sID]))
						for i := range tableRow {
							tableRow[i] = run.GetColumnAt(checkeds[sID][i].GetColID(), k)
						}
						utils.Panic(
							"entry %v of the table %v is not included in the table. tableRow=%v",
							k, nameTable(checkeds[sID]), vector.Prettify(tableRow),
						)
					}
					M[posInM].Add(&M[posInM], &oneValue)
				}
			}

			// And assign it in the prover runtime to notify the framework
			// the M is now available and can be used in other routines.
			run.AssignColumn(ctx.M.GetColID(), sv.NewRegular(M))
		})

		// Step responsible for assigning "SIGMA_{S,k}, SIGMA_T and the LOCAL_OPENING"
		meta.provers.AppendToInner(round+1, func(run *wizard.ProverRuntime) {

			// As above, we can get the assignment of a column committed previously
			// with the instruction below. You will need it for recovering
			// assignments to S_k , T, and M.
			gamma := run.GetRandomCoinField(ctx.gamma.Name)
			// Compute and assign SIGMA_{S_k}
			for k := range checkeds {
				sK := ctx.S[k].GetColAssignment(run)
				sKArr := sv.IntoRegVec(sK)
				sIPlusGammmInv := make([]field.Element, len(sKArr))
				for i := 0; i < len(sKArr); i++ {
					sIPlusGammmInv[i].Add(&sKArr[i], &gamma)
				}
				sIPlusGammmInv = field.BatchInvert(sIPlusGammmInv)
				sigmaSKWitArr := make([]field.Element, len(sKArr))
				sigmaSKWitArr[0] = sIPlusGammmInv[0]
				for i := 1; i < len(sKArr); i++ {
					sigmaSKWitArr[i].Add(&sigmaSKWitArr[i-1], &sIPlusGammmInv[i])
				}
				sigmaSKWit := sv.NewRegular(sigmaSKWitArr)
				run.AssignColumn(ctx.sigmaS[k].GetColID(), sigmaSKWit)
				// Prep for Query (1)
				run.AssignLocalPoint(ctx.S_LOCAL_OPENING[k].ID, sigmaSKWitArr[len(sigmaSKWitArr)-1])
			}

			// Compute and assign SIGMA_T
			tWit := ctx.T.GetColAssignment(run)
			tWitArr := sv.IntoRegVec(tWit)
			mWit := ctx.M.GetColAssignment(run)
			mWitArr := sv.IntoRegVec(mWit)
			tIPlusGammmInv := make([]field.Element, len(tWitArr))
			for i := 0; i < len(tWitArr); i++ {
				tIPlusGammmInv[i].Add(&tWitArr[i], &gamma)
			}
			tIPlusGammmInv = field.BatchInvert(tIPlusGammmInv)
			sigmaTWitArr := make([]field.Element, len(tWitArr))
			sigmaTWitArr[0].Mul(&mWitArr[0], &tIPlusGammmInv[0])
			for i := 1; i < len(tWitArr); i++ {
				sigmaTWitArr[i].Mul(&mWitArr[i], &tIPlusGammmInv[i]).Add(&sigmaTWitArr[i], &sigmaTWitArr[i-1])
			}
			sigmaTWit := sv.NewRegular(sigmaTWitArr)
			run.AssignColumn(ctx.sigmaT.GetColID(), sigmaTWit)

			// Prep for Query (1)
			run.AssignLocalPoint(ctx.T_LOCAL_OPENING.ID, sigmaTWitArr[len(sigmaTWitArr)-1])
		})

		// Verifier step responsible for checking the consistency between the
		// last entries of SIGMA_S,k and SIGMA_T
		comp.InsertVerifier(round+1, func(run *wizard.VerifierRuntime) error {
			// Get the alleged LocalOpening query value
			sigmaSKSum := field.Zero()
			for k := range checkeds {
				temp := run.GetLocalPointEvalParams(ctx.S_LOCAL_OPENING[k].ID).Y
				sigmaSKSum.Add(&sigmaSKSum, &temp)
			}
			sigmaTSum := run.GetLocalPointEvalParams(ctx.T_LOCAL_OPENING.ID).Y

			// And check that the \sum_k SIGMA_S,k == SIGMA_T
			// If it fails, notify the caller with an error message
			var fail bool = (sigmaSKSum != sigmaTSum)
			if fail {
				return fmt.Errorf("LOG DERIVATIVE LOOKUPS : - verifier check failed")
			}
			return nil
		}, func(api frontend.API, run *wizard.WizardVerifierCircuit) {
			// Get the alleged LocalOpening query value
			var sigmaSKSum frontend.Variable = field.Zero()
			for k := range checkeds {
				temp := run.GetLocalPointEvalParams(ctx.S_LOCAL_OPENING[k].ID).Y
				sigmaSKSum = api.Add(sigmaSKSum, temp)
			}
			sigmaTSum := run.GetLocalPointEvalParams(ctx.T_LOCAL_OPENING.ID).Y

			// And check that the \sum_k SIGMA_S,k == SIGMA_T
			// If it fails, notify the caller with an error message
			api.AssertIsEqual(sigmaSKSum, sigmaTSum)
		})
	}

	/*
		And registers all the previously grouped lookups into the
	*/
	for i := 0; i < meta.provers.Len(); i++ {
		comp.SubProvers.AppendToInner(i, meta.Prover(i))
	}
}

func appendToLookups(lookups map[ifaces.ColID][]S, checked S, table T) bool {
	// check if T already exists in lookups
	_, exist := lookups[nameTable(table)]
	if !exist {
		lookups[nameTable(table)] = make([]S, 0)
	}
	lookups[nameTable(table)] = append(lookups[nameTable(table)], checked)

	return !exist
}

// Function that derives a name for both S and T
func nameTable[T ~[]ifaces.Column](t T) ifaces.ColID {

	colNames := make([]string, len(t))

	for col := range colNames {
		colNames = append(colNames, string(t[col].GetColID()))
	}
	// Can we use TABLE_colNames[0].ColID as a table name
	return ifaces.ColIDf("TABLE%v", strings.Join(colNames, "_"))
}

/*
Derive a name for a a coin created during the compilation process
*/
func deriveTableName[R ~string](context string, t []ifaces.Column, name string) R {
	var res string
	if nameTable(t) == "" {
		res = fmt.Sprintf("%v_%v", context, name)
	} else {
		res = fmt.Sprintf("%v_%v_%v", nameTable(t), context, name)
	}
	return R(res)
}

func deriveTableNameWithIndex[R ~string](context string, t []ifaces.Column, index int, name string) R {
	var res string
	if nameTable(t) == "" {
		res = fmt.Sprintf("%v_%v", context, name)
	} else {
		res = fmt.Sprintf("%v_%v_%v_%v", nameTable(t), index, context, name)
	}
	return R(res)
}

func getTableCanonicalOrder(q query.Inclusion) (S, T) {

	checked := make(S, len(q.Included))
	table := make(T, len(q.Including))

	// we sort the included and the including columns of q into
	// checked and table. but we need to reorder the columns in
	// alphabetic order.
	colNamesT := make([]ifaces.ColID, len(checked))
	sortingMap := make([]int, len(table))

	for col := range colNamesT {
		colNamesT[col] = q.Including[col].GetColID()
		sortingMap[col] = col
	}

	swap := func(i, j int) {
		colNamesT[i], colNamesT[j] = colNamesT[j], colNamesT[i]
		sortingMap[i], sortingMap[j] = sortingMap[j], sortingMap[i]
	}

	less := func(i, j int) int {
		return strings.Compare(string(colNamesT[i]), string(colNamesT[j]))
	}

	// bubble sort
	for i := range colNamesT {
		swapped := false
		for j := 0; j < len(colNamesT)-i-1; j++ {
			if less(j+1, j) < 0 {
				swap(j, j+1)
				swapped = true
			}
		}

		if !swapped {
			break
		}
	}

	for newPos, oldPos := range sortingMap {
		checked[newPos] = q.Included[oldPos]
		table[newPos] = q.Including[oldPos]
	}

	return checked, table
}
