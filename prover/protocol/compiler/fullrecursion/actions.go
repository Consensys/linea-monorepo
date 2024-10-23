package fullrecursion

import (
	"fmt"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

type CircuitAssignment fullRecursionCtx

func (c CircuitAssignment) Run(run *wizard.ProverRuntime) {
	c.PlonkInWizard.ProverAction.Run(run, WitnessAssigner(c))
}

type ReplacementAssignment fullRecursionCtx

func (c ReplacementAssignment) Run(run *wizard.ProverRuntime) {
	params := run.GetUnivariateParams(c.PolyQuery.QueryID)
	run.AssignUnivariate(c.PolyQueryReplacement.QueryID, params.X, params.Ys...)

	oldRoots := c.PcsCtx.Items.MerkleRoots
	for i := range c.MerkleRootsReplacement {

		if c.PcsCtx.Items.MerkleRoots[i] == nil {
			continue
		}

		run.AssignColumn(
			c.MerkleRootsReplacement[i].GetColID(),
			oldRoots[i].GetColAssignment(run),
		)
	}
}

type LocalOpeningAssignment fullRecursionCtx

func (c LocalOpeningAssignment) Run(run *wizard.ProverRuntime) {
	for i := range c.LocalOpenings {
		run.AssignLocalPoint(
			c.LocalOpenings[i].ID,
			c.PlonkInWizard.PI.GetColAssignmentAt(run, i),
		)
	}
}

type ConsistencyCheck struct {
	fullRecursionCtx
	isSkipped bool
}

func (c *ConsistencyCheck) Run(run *wizard.VerifierRuntime) error {

	var (
		initialFsCirc = run.GetLocalPointEvalParams(c.LocalOpenings[0].ID).Y
		initialFsRt   = run.FiatShamirHistory[c.FirstRound][0][0]
		piCursor      = 2
	)

	if initialFsCirc != initialFsRt {
		return fmt.Errorf("full recursion: the initial FS do not match")
	}

	for i := range c.NonEmptyMerkleRootPositions {

		var (
			pos      = c.NonEmptyMerkleRootPositions[i]
			fromRt   = c.MerkleRootsReplacement[pos].GetColAssignmentAt(run, 0)
			fromCirc = run.GetLocalPointEvalParams(c.LocalOpenings[piCursor+i].ID).Y
		)

		if fromRt != fromCirc {
			return fmt.Errorf("full recursion: the commitment does not match (pos: %v)", i)
		}
	}

	piCursor += len(c.NonEmptyMerkleRootPositions)

	var (
		paramsRt = run.GetUnivariateParams(c.PolyQueryReplacement.QueryID)
		xRt      = paramsRt.X
		xCirc    = run.GetLocalPointEvalParams(c.LocalOpenings[piCursor].ID).Y
	)

	if xRt != xCirc {
		return fmt.Errorf("full recursion: the Ys does not match")
	}

	piCursor++

	for i := range paramsRt.Ys {

		var (
			fromRt   = paramsRt.Ys[i]
			fromCirc = run.GetLocalPointEvalParams(c.LocalOpenings[piCursor+i].ID).Y
		)

		if fromRt != fromCirc {
			return fmt.Errorf("full recursion: the Ys does not match (pos: %v)", i)
		}
	}

	piCursor += len(paramsRt.Ys)

	// The public inputs do not need to be checked because they are redefined in
	// term of the local openings directly. So checking it would amount to checking
	// that the local openings are equal to themselves.

	return nil
}

func (c *ConsistencyCheck) RunGnark(api frontend.API, run *wizard.WizardVerifierCircuit) {

	var (
		initialFsCirc = run.GetLocalPointEvalParams(c.LocalOpenings[0].ID).Y
		initialFsRt   = run.FiatShamirHistory[c.FirstRound][0][0]
		piCursor      = 2
	)

	api.AssertIsEqual(initialFsCirc, initialFsRt)

	for i := range c.NonEmptyMerkleRootPositions {

		var (
			pos      = c.NonEmptyMerkleRootPositions[i]
			fromRt   = c.MerkleRootsReplacement[pos].GetColAssignmentGnarkAt(run, 0)
			fromCirc = run.GetLocalPointEvalParams(c.LocalOpenings[piCursor+i].ID).Y
		)

		api.AssertIsEqual(fromRt, fromCirc)
	}

	piCursor += len(c.NonEmptyMerkleRootPositions)

	var (
		paramsRt = run.GetUnivariateParams(c.PolyQueryReplacement.QueryID)
		xRt      = paramsRt.X
		xCirc    = run.GetLocalPointEvalParams(c.LocalOpenings[piCursor].ID).Y
	)

	api.AssertIsEqual(xRt, xCirc)

	piCursor++

	for i := range paramsRt.Ys {

		var (
			fromRt   = paramsRt.Ys[i]
			fromCirc = run.GetLocalPointEvalParams(c.LocalOpenings[piCursor+i].ID).Y
		)

		api.AssertIsEqual(fromRt, fromCirc)
	}

	piCursor += len(paramsRt.Ys)

	// The public inputs do not need to be checked because they are redefined in
	// term of the local openings directly. So checking it would amount to checking
	// that the local openings are equal to themselves.
}

func (c *ConsistencyCheck) Skip() {
	c.isSkipped = true
}

func (c *ConsistencyCheck) IsSkipped() bool {
	return c.isSkipped
}

type ResetFsActions struct {
	fullRecursionCtx
	isSkipped bool
}

func (r *ResetFsActions) Run(run *wizard.VerifierRuntime) error {
	finalFsCirc := run.GetLocalPointEvalParams(r.LocalOpenings[1].ID).Y
	run.FS.SetState([]field.Element{finalFsCirc})
	return nil
}

func (r *ResetFsActions) RunGnark(api frontend.API, run *wizard.WizardVerifierCircuit) {
	finalFsCirc := run.GetLocalPointEvalParams(r.LocalOpenings[1].ID).Y
	run.FS.SetState([]frontend.Variable{finalFsCirc})
}

func (r *ResetFsActions) Skip() {
	r.isSkipped = true
}

func (r *ResetFsActions) IsSkipped() bool {
	return r.isSkipped
}
