package conglomeration

import (
	"fmt"

	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/selfrecursion"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/vortex"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
)

// recursionCtx holds compilation context informations about the wizard
type recursionCtx struct {
	// A pointer to the compiled-IOP over which the compilation step has run
	Translator *compTranslator
	Tmpl       *wizard.CompiledIOP
	// The Vortex compilation context
	PcsCtx                      *vortex.Ctx
	PublicInputs                []wizard.PublicInput
	NonEmptyMerkleRootPositions []int
	FirstRound, LastRound       int
	Columns                     [][]ifaces.Column
	// The columns ignored are the one that are compiled by the vortex context.
	// They are added in the target 'comp' but are assigned to zero. Although,
	// they do not directly play a role in the protocol anymore, they are still
	// referenced by the self-recursion compiler.
	ColumnsIgnored  [][]ifaces.Column
	QueryParams     [][]ifaces.Query
	VerifierActions [][]wizard.VerifierAction
	Coins           [][]coin.Info
	FsHooks         [][]wizard.VerifierAction
	LocalOpenings   []query.LocalOpening
}

// ConglomerateDefineFunc returns a function that defines a conglomerate
// comp and a placeholder pointer for the recursion context. On return
// of the function, the pointer points to an empty slice and is populated
// once [wizard.Compile] has been called with def.
func ConglomerateDefineFunc(tmpl *wizard.CompiledIOP, maxNumSegment int) (def func(*wizard.Builder), ctxsPlaceHolder *[]*recursionCtx) {

	var ctxs []*recursionCtx
	def = func(b *wizard.Builder) {

		comp := b.CompiledIOP

		for id := 0; id < maxNumSegment; id++ {
			prefix := fmt.Sprintf("verifier-%v", id)
			ctx := initRecursionCtx(prefix, comp, tmpl)
			ctx.captureCompPreVortex(tmpl)
			ctx.captureVortexCtx(tmpl)
			ctxs = append(ctxs, ctx)
		}

		// This FS hook has to be defined before we add the pre-vortex verifier
		// hooks to ensure that the FS state is properly initialize the verifier
		// runtime.
		comp.FiatShamirHooks.AppendToInner(0, &SubFsInitialize{Ctxs: ctxs})

		for round := 0; round <= ctxs[0].LastRound; round++ {

			var (
				hasCoin    = len(ctxs[0].Coins[round]) > 0
				hasVAction = len(ctxs[0].VerifierActions[round]) > 0
				hasFsHook  = len(ctxs[0].FsHooks[round]) > 0
				hasColumn  = len(ctxs[0].Columns[round]) > 0
				hasQParams = len(ctxs[0].QueryParams[round]) > 0
			)

			if hasCoin || hasVAction || hasFsHook {
				// The way the verifier runtime is that it will generate all the random coins at once and
				// then, it runs all the verifier actions in parallel. What this action from the verifier
				// is trying to do is to prepare a ctx-local FS state that can be later used in a join to
				// derive a sound global FS state. Thus, we need it to run along side the "main" fs random
				// coin generation. This is why this is declared as an FS hook and not as a VerifierAction.
				comp.FiatShamirHooks.AppendToInner(round, &PreVortexVerifierStep{Ctxs: ctxs, Round: round})
			}

			if hasColumn || hasQParams {
				comp.RegisterProverAction(round, &PreVortexProverStep{Ctxs: ctxs, Round: round})
			}
		}

		comp.FiatShamirHooks.AppendToInner(ctxs[0].LastRound, &FsJoinHook{Ctxs: ctxs})
		comp.RegisterProverAction(ctxs[0].LastRound, &FsJoinProverStep{Ctxs: ctxs})
		comp.RegisterProverAction(ctxs[0].LastRound, &AssignVortexQuery{Ctxs: ctxs})
		comp.RegisterProverAction(ctxs[0].LastRound+1, &AssignVortexUAlpha{Ctxs: ctxs})
		comp.RegisterProverAction(ctxs[0].LastRound+2, &AssignVortexOpenedCols{Ctxs: ctxs})

		// Importantly, the recursion compilation should happen after we added the vortex
		// columns as they depends on the later.
		for _, ctx := range ctxs {
			selfrecursion.RecurseOverCustomCtx(comp, ctx.PcsCtx, ctx.Translator.Prefix)
		}
	}

	return def, &ctxs
}

// initRecursionCtx initializes a new context
func initRecursionCtx(id string, target *wizard.CompiledIOP, tmpl *wizard.CompiledIOP) *recursionCtx {
	return &recursionCtx{
		Translator: &compTranslator{Prefix: id, Target: target},
		Tmpl:       tmpl,
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

		// Importantly, the coins are added before. Otherwise the 'assertConsistentRound'
		// clause would not accept inserting columns or queries.
		for _, cName := range tmpl.Coins.AllKeysAt(round) {

			if tmpl.Coins.IsSkippedFromVerifierTranscript(cName) {
				continue
			}

			coinInfo := tmpl.Coins.Data(cName)
			coinInfo = ctx.Translator.InsertCoin(coinInfo)
			ctx.Coins[round] = append(ctx.Coins[round], coinInfo)
			ctx.Translator.Target.Coins.MarkAsSkippedFromVerifierTranscript(coinInfo.Name)
		}

		for _, colName := range tmpl.Columns.AllKeysAt(round) {

			// filter the columns by status
			var (
				col    = tmpl.Columns.GetHandle(colName).(column.Natural)
				status = col.Status()
			)

			if !status.IsPublic() {
				// the column is not public so it is not part of the proof
				continue
			}

			newCol := ctx.Translator.InsertColumn(col)
			ctx.Columns[round] = append(ctx.Columns[round], newCol)
			ctx.Translator.Target.Columns.ExcludeFromProverFS(newCol.GetColID())
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

			// Note that we do not filter the already compiled queries
			qInfo := tmpl.QueriesParams.Data(qName)
			qInfo = ctx.Translator.InsertQueryParams(round, qInfo)
			ctx.QueryParams[round] = append(ctx.QueryParams[round], qInfo)
			ctx.Translator.Target.QueriesParams.MarkAsSkippedFromProverTranscript(qInfo.Name())
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
		comsByRound  = srcVortexCtx.CommitmentsByRounds.Inner()
	)

	for _, coms := range comsByRound {
		ctx.ColumnsIgnored = append(ctx.ColumnsIgnored, nil)
		for _, comID := range coms {
			com := tmpl.Columns.GetHandle(comID)
			com = ctx.Translator.InsertColumn(com.(column.Natural))
			ctx.ColumnsIgnored[len(ctx.ColumnsIgnored)-1] = append(ctx.ColumnsIgnored[len(ctx.ColumnsIgnored)-1], com)
		}
	}

	if !srcVortexCtx.IsSelfrecursed || srcVortexCtx.ReplaceSisByMimc {
		utils.Panic("the input vortex ctx is expected to be selfrecursed or having SIS replaced by MiMC. Please sure the input comp has been last compiled by Vortex with the option [vortex.MarkAsSelfRecursed]")
	}

	dstVortexCtx := &vortex.Ctx{
		RunStateNamePrefix: ctx.Translator.Prefix,
		BlowUpFactor:       srcVortexCtx.BlowUpFactor,
		DryTreshold:        srcVortexCtx.DryTreshold,
		CommittedRowsCount: srcVortexCtx.CommittedRowsCount,
		NumCols:            srcVortexCtx.NumCols,
		MaxCommittedRound:  srcVortexCtx.MaxCommittedRound,
		NumOpenedCol:       srcVortexCtx.NumOpenedCol,
		VortexParams:       srcVortexCtx.VortexParams,
		SisParams:          srcVortexCtx.SisParams,
		// Although the srcVor
		IsSelfrecursed:               true,
		CommitmentsByRounds:          ctx.Translator.TranslateColumnVecVec(srcVortexCtx.CommitmentsByRounds),
		DriedByRounds:                ctx.Translator.TranslateColumnVecVec(srcVortexCtx.DriedByRounds),
		PolynomialsTouchedByTheQuery: ctx.Translator.TranslateColumnSet(srcVortexCtx.PolynomialsTouchedByTheQuery),
		ShadowCols:                   ctx.Translator.TranslateColumnSet(srcVortexCtx.ShadowCols),
		Query:                        ctx.Translator.TranslateUniEval(ctx.LastRound, srcVortexCtx.Query),
	}

	if srcVortexCtx.ReplaceSisByMimc {
		panic("it should not replace by MiMC")
	}

	ctx.Translator.Target.QueriesParams.MarkAsIgnored(dstVortexCtx.Query.QueryID)

	if srcVortexCtx.IsCommitToPrecomputed() {
		dstVortexCtx.Items.Precomputeds.PrecomputedColums = ctx.Translator.TranslateColumnList(srcVortexCtx.Items.Precomputeds.PrecomputedColums)
		dstVortexCtx.Items.Precomputeds.MerkleRoot = ctx.Translator.GetColumn(srcVortexCtx.Items.Precomputeds.MerkleRoot.GetColID())
		dstVortexCtx.Items.Precomputeds.CommittedMatrix = srcVortexCtx.Items.Precomputeds.CommittedMatrix
		dstVortexCtx.Items.Precomputeds.DhWithMerkle = srcVortexCtx.Items.Precomputeds.DhWithMerkle
	}

	dstVortexCtx.Items.Alpha = ctx.Translator.InsertCoin(srcVortexCtx.Items.Alpha)
	dstVortexCtx.Items.Ualpha = ctx.Translator.InsertColumn(srcVortexCtx.Items.Ualpha.(column.Natural))
	dstVortexCtx.Items.Q = ctx.Translator.InsertCoin(srcVortexCtx.Items.Q)
	dstVortexCtx.Items.OpenedColumns = ctx.Translator.InsertColumns(srcVortexCtx.Items.OpenedColumns)
	dstVortexCtx.Items.MerkleProofs = ctx.Translator.InsertColumn(srcVortexCtx.Items.MerkleProofs.(column.Natural))
	dstVortexCtx.Items.MerkleRoots = ctx.Translator.TranslateColumnList(srcVortexCtx.Items.MerkleRoots)

	ctx.PcsCtx = dstVortexCtx
}
