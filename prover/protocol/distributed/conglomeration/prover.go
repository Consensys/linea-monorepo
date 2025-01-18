package conglomeration

import (
	"strings"

	"github.com/consensys/linea-monorepo/prover/crypto/state-management/smt"
	vCom "github.com/consensys/linea-monorepo/prover/crypto/vortex"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/vortex"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
)

const (
	subProofInStatePrefixStr = ".subProof"
)

// Witness is a collection of inputs corresponding to a segment proof to provide
// to the main prover of a conglomerate comp.
type Witness struct {
	Proof             wizard.Proof
	CommittedMatrices []vCom.EncodedMatrix
	SisHashes         [][]field.Element
	Trees             []*smt.Tree
}

// PreVortexProverStep is a step replicating the prover of the tmpl at round
// `Round` before the Vortex compilation step. It works by adding columns from
// a wizard proof stored in the prover runtime. The proof is fetched from the
// runtime state. That means the prover step should be run after the proof has
// been attached to the runtime.
type PreVortexProverStep struct {
	Ctxs  []*recursionCtx
	Round int
}

// AssignVortexUAlpha assigns the UAlpha column for all the subproofs. As
// for [PreVortexVerifierStep], this step should be run after the corresponding
// proofs have been added to the runtime states.
type AssignVortexUAlpha struct {
	Ctxs  []*recursionCtx
	Round int
}

// AssignVortexOpenedCols assigns the OpenedCols for all the subproofs. As
// for [PreVortexVerifierStep], this step should be run after the corresponding
// proofs have been added to the runtime states.
type AssignVortexOpenedCols struct {
	Ctxs  []*recursionCtx
	Round int
}

// ProveConglomeration returns the main prover step of the conglomeration wizard.
// It takes a list of [Witness] as input and complete the list with the last
// value.
func ProveConglomeration(ctxs []*recursionCtx, witnesses []Witness) wizard.ProverStep {

	if len(witnesses) > len(ctxs) {
		utils.Panic("More witnesses than ctxs, numWitnesses: %v, numCtxs: %v", len(witnesses), len(ctxs))
	}

	return func(run *wizard.ProverRuntime) {

		for i, witness := range witnesses {
			storeWitnessInState(run, ctxs[i], witness)
		}

		for i := len(witnesses); i < len(ctxs); i++ {
			storeWitnessInState(run, ctxs[i], witnesses[len(witnesses)-1])
		}
	}
}

func storeWitnessInState(run *wizard.ProverRuntime, ctx *recursionCtx, witness Witness) {

	var (
		prefix    = ctx.Translator.Prefix
		lastRound = ctx.LastRound
	)

	run.State.InsertNew(prefix+subProofInStatePrefixStr, witness.Proof)

	for round := 0; round <= lastRound; round++ {

		if witness.CommittedMatrices[round] != nil {
			run.State.InsertNew(ctx.PcsCtx.VortexProverStateName(round), witness.CommittedMatrices[round])
		}

		if witness.SisHashes[round] != nil {
			run.State.InsertNew(ctx.PcsCtx.SisHashName(round), witness.SisHashes[round])
		}

		if witness.Trees[round] != nil {
			run.State.InsertNew(ctx.PcsCtx.MerkleTreeName(round), witness.Trees[round])
		}
	}
}

// ExtractWitness extracts a [Witness] from a prover runtime toward being conglomerated.
func ExtractWitness(run *wizard.ProverRuntime) Witness {

	var (
		pcs               = run.Spec.PcsCtxs.(*vortex.Ctx)
		committedMatrices []vCom.EncodedMatrix
		sisHashes         [][]field.Element
		trees             []*smt.Tree
		lastRound         = run.Spec.QueriesParams.Round(pcs.Query.QueryID)
	)

	for round := 0; round <= lastRound; round++ {

		var (
			committedMatrix, _ = run.State.TryGet(pcs.VortexProverStateName(round))
			sisHash, _         = run.State.TryGet(pcs.SisHashName(round))
			tree, _            = run.State.TryGet(pcs.MerkleTreeName(round))
		)

		if committedMatrix != nil {
			committedMatrices = append(committedMatrices, committedMatrix.(vCom.EncodedMatrix))
		}

		if sisHash != nil {
			sisHashes = append(sisHashes, sisHash.([]field.Element))
		}

		if tree != nil {
			trees = append(trees, tree.(*smt.Tree))
		}
	}

	return Witness{
		Proof:             run.ExtractProof(),
		CommittedMatrices: committedMatrices,
		SisHashes:         sisHashes,
		Trees:             trees,
	}
}

func (pa PreVortexProverStep) Run(run *wizard.ProverRuntime) {
	for _, ctx := range pa.Ctxs {

		var (
			prefix        = ctx.Translator.Prefix
			proof         = run.State.MustGet(prefix + subProofInStatePrefixStr).(wizard.Proof)
			queriesParams = ctx.QueryParams[pa.Round]
			colums        = ctx.Columns[pa.Round]
		)

		for _, col := range colums {
			name := unprefix(prefix, col.GetColID())
			run.AssignColumn(col.GetColID(), proof.Messages.MustGet(name))
		}

		for _, param := range queriesParams {
			name := unprefix(prefix, param.Name())
			run.QueriesParams.InsertNew(param.Name(), proof.QueriesParams.MustGet(name))
		}
	}
}

func (pa AssignVortexUAlpha) Run(run *wizard.ProverRuntime) {
	for _, ctx := range pa.Ctxs {
		// Since all the context of the pcs is translated, this does not
		// need to run over a translated prover runtime.
		ctx.PcsCtx.ComputeLinearComb(run)
	}
}

func (pa AssignVortexOpenedCols) Run(run *wizard.ProverRuntime) {
	for _, ctx := range pa.Ctxs {
		// Since all the context of the pcs is translated, this does not
		// need to run over a translated prover runtime.
		ctx.PcsCtx.OpenSelectedColumns(run)
	}
}

func unprefix[T ~string](prefix string, name T) T {
	p, n := string(prefix), string(name)
	r := strings.TrimPrefix(n, p)
	return T(r)
}
