package conglomeration

import (
	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/vortex"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

// recursionCtx holds compilation context informations about the wizard
type recursionCtx struct {
	// A pointer to the compiled-IOP over which the compilation step has run
	Translator *compTranslator
	// The Vortex compilation context
	PcsCtx                      *vortex.Ctx
	PublicInputs                []wizard.PublicInput
	NonEmptyMerkleRootPositions []int
	FirstRound, LastRound       int
	Columns                     [][]ifaces.Column
	QueryParams                 [][]ifaces.Query
	VerifierActions             [][]wizard.VerifierAction
	Coins                       [][]coin.Info
	FsHooks                     [][]wizard.VerifierAction
	LocalOpenings               []query.LocalOpening
}

func Conglomerate(
	tmpl *wizard.CompiledIOP,
	maxNumSegment int,
) (comp *wizard.CompiledIOP) {
	return nil
}

func addVerifierToComp(
	comp *wizard.CompiledIOP,
	tmpl *wizard.CompiledIOP,
) {
}

// initCtx initializes a new context
func initRecursionCtx(
	id string,
	target *wizard.CompiledIOP,
) *recursionCtx {
	return &recursionCtx{
		Translator: &compTranslator{Prefix: id, Target: target},
	}
}

// captureCompPreVortex scans the content of tmpl to store the compilation infos of the
// CompiledIOP at the beginning of the compilation. The scanned wizard items are
// inserted into `comp` with a prefix `id` and recorded within the context.
func (ctx *recursionCtx) captureCompPreVortex(tmpl *wizard.CompiledIOP) {

	var (
		polyQuery = tmpl.PcsCtxs.(*vortex.Ctx).Query
		lastRound = tmpl.QueriesParams.Round(polyQuery.QueryID)
	)

	ctx.LastRound = lastRound

	for round := 0; round <= lastRound; round++ {

		ctx.Columns = append(ctx.Columns, []ifaces.Column{})
		ctx.QueryParams = append(ctx.QueryParams, []ifaces.Query{})
		ctx.VerifierActions = append(ctx.VerifierActions, []wizard.VerifierAction{})
		ctx.Coins = append(ctx.Coins, []coin.Info{})
		ctx.FsHooks = append(ctx.FsHooks, []wizard.VerifierAction{})

		for _, colName := range tmpl.Columns.AllKeysAt(round) {

			// filter the columns by status
			var (
				status = tmpl.Columns.Status(colName)
				size   = tmpl.Columns.GetSize(colName)
			)

			if !status.IsPublic() {
				// the column is not public so it is not part of the proof
				continue
			}

			newCol := ctx.Translator.InsertColumn(round, colName, size, status)
			ctx.Columns[round] = append(ctx.Columns[round], newCol)
			ctx.Translator.Target.Columns.IgnoreButKeepInProverTranscript(newCol.GetColID())
		}

		for _, qName := range tmpl.QueriesParams.AllKeysAt(round) {

			if tmpl.QueriesParams.IsSkippedFromVerifierTranscript(qName) {
				continue
			}

			// Importantly, the queries that we port should be already
			// compiled in the tmpl.
			if !tmpl.QueriesParams.IsIgnored(qName) {
				panic("the template is invalid, all its queries should be compiled")
			}

			// The uni-eval query is directly handled in a different section
			// of the compilation.
			if qName == polyQuery.QueryID {
				continue
			}

			// Not that we do not filter the already compiled queries
			qInfo := tmpl.QueriesParams.Data(qName)
			qInfo = ctx.Translator.InsertQueryParams(round, qInfo)
			ctx.QueryParams[round] = append(ctx.QueryParams[round], qInfo)
			ctx.Translator.Target.QueriesParams.MarkAsSkippedFromVerifierTranscript(qInfo.Name())
		}

		for _, cName := range tmpl.Coins.AllKeysAt(round) {

			if tmpl.Coins.IsSkippedFromVerifierTranscript(cName) {
				continue
			}

			coin := tmpl.Coins.Data(cName)
			coin = ctx.Translator.InsertCoin(round, cName, coin.Type, coin.Size)
			ctx.Coins[round] = append(ctx.Coins[round], coin)
			ctx.Translator.Target.Coins.MarkAsSkippedFromVerifierTranscript(coin.Name)
		}

		verifierActions := tmpl.SubVerifiers.Inner()

		for i := range verifierActions[round] {

			va := verifierActions[round][i]
			if va.IsSkipped() {
				continue
			}

			ctx.VerifierActions[round] = append(ctx.VerifierActions[round], va)
		}

		resetFs := tmpl.FiatShamirHooks.Inner()

		for _, fsHook := range resetFs[round] {

			if fsHook.IsSkipped() {
				continue
			}

			ctx.FsHooks[round] = append(ctx.VerifierActions[round], fsHook)
		}
	}
}

func (ctx *recursionCtx) captureVortexCtx(tmpl *wizard.CompiledIOP) {

	var (
		srcVortexCtx = tmpl.PcsCtxs.(*vortex.Ctx)
		dstVortexCtx = &vortex.Ctx{
			BlowUpFactor:                 srcVortexCtx.BlowUpFactor,
			DryTreshold:                  srcVortexCtx.DryTreshold,
			CommittedRowsCount:           srcVortexCtx.CommittedRowsCount,
			NumCols:                      srcVortexCtx.NumCols,
			MaxCommittedRound:            srcVortexCtx.MaxCommittedRound,
			NumOpenedCol:                 srcVortexCtx.NumOpenedCol,
			CommitmentsByRounds:          ctx.Translator.TranslateColumnVecVec(srcVortexCtx.CommitmentsByRounds),
			DriedByRounds:                ctx.Translator.TranslateColumnVecVec(srcVortexCtx.DriedByRounds),
			PolynomialsTouchedByTheQuery: ctx.Translator.TranslateColumnSet(srcVortexCtx.PolynomialsTouchedByTheQuery),
			ShadowCols:                   ctx.Translator.TranslateColumnSet(srcVortexCtx.ShadowCols),
			Query:                        ctx.Translator.TranslateUniEval(ctx.LastRound, srcVortexCtx.Query),
		}
	)

	if srcVortexCtx.ReplaceSisByMimc {
		panic("it should not replace by MiMC")
	}

	dstVortexCtx.Items.Precomputeds.PrecomputedColums = ctx.Translator.TranslateColumnList(srcVortexCtx.Items.Precomputeds.PrecomputedColums)
	dstVortexCtx.Items.Precomputeds.MerkleRoot = ctx.Translator.GetColumn(srcVortexCtx.Items.Precomputeds.MerkleRoot.GetColID())
	dstVortexCtx.Items.Precomputeds.Dh = ctx.Translator.GetColumn(srcVortexCtx.Items.Precomputeds.Dh.GetColID())
	dstVortexCtx.Items.Precomputeds.CommittedMatrix = srcVortexCtx.Items.Precomputeds.CommittedMatrix
	dstVortexCtx.Items.Precomputeds.Tree = srcVortexCtx.Items.Precomputeds.Tree
	dstVortexCtx.Items.Precomputeds.DhWithMerkle = srcVortexCtx.Items.Precomputeds.DhWithMerkle

	dstVortexCtx.Items.Dh = ctx.Translator.TranslateColumnList(srcVortexCtx.Items.Dh)
	dstVortexCtx.Items.Alpha = ctx.Translator.GetCoin(srcVortexCtx.Items.Alpha.Name)
	dstVortexCtx.Items.Ualpha = ctx.Translator.GetColumn(srcVortexCtx.Items.Ualpha.GetColID())
	dstVortexCtx.Items.Q = ctx.Translator.GetCoin(srcVortexCtx.Items.Q.Name)
	dstVortexCtx.Items.OpenedColumns = ctx.Translator.TranslateColumnList(srcVortexCtx.Items.OpenedColumns)
	dstVortexCtx.Items.MerkleProofs = ctx.Translator.GetColumn(srcVortexCtx.Items.MerkleProofs.GetColID())
	dstVortexCtx.Items.MerkleRoots = ctx.Translator.TranslateColumnList(srcVortexCtx.Items.MerkleRoots)
}

// TranslateUniEval returns a copied UnivariateEval query with the columns translated
// and the names translated. The returned query is registered in the translator comp.
func (comp *compTranslator) TranslateUniEval(round int, q query.UnivariateEval) query.UnivariateEval {
	var (
		res = query.NewUnivariateEval(q.QueryID, q.Pols...)
	)

	for i := range res.Pols {
		res.Pols[i] = comp.GetColumn(res.Pols[i].GetColID())
	}

	return comp.InsertQueryParams(round, res).(query.UnivariateEval)
}
