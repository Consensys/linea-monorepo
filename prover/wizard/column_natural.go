package wizard

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/zkevm-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/zkevm-monorepo/prover/maths/field"
	"github.com/consensys/zkevm-monorepo/prover/utils"
)

// ColNatural designate a column that is explicitly part of the AIR it serves
// as building  block for constructing more elaborate columns: expression columns
// shifted columns.
type ColNatural struct {
	size       int
	round      int
	visibility *Visibility
	metadata   *metadata
}

func (api *API) NewCommit(round, size int) *ColNatural {
	return api.newColumn(round, size, Committed)
}

func (api *API) NewColumn(round, size int, visibility Visibility) *ColNatural {
	return api.newColumn(round, size, visibility)
}

func (api *API) NewPrecomputed(sv smartvectors.SmartVector) *ColNatural {
	res := api.newColumn(0, sv.Len(), Precomputed)
	api.precomputeds.InsertNew(res.id(), sv)
	return res
}

func (api *API) newColumn(round, size int, visibility Visibility) *ColNatural {

	var (
		// make a copy of the argument to ensure uniqueness of the variable to
		// the column. Without it, we would have side effects on a column A
		// if we changed the visibility of column B.
		v   = visibility
		nat = ColNatural{
			round:      round,
			size:       size,
			visibility: &v,
			metadata:   api.newMetadata(),
		}
	)

	api.columns.addToRound(round, nat)
	return &nat
}

func (api *CompiledIOP) AllColumns() []ColNatural {
	return api.columns.all()
}

func (api *CompiledIOP) AllMatchingColumns(optRoundFilter *int, optVisibilityFilter *Visibility) []ColNatural {

	var (
		fullList = api.columns.all()
		res      = make([]ColNatural, 0, len(fullList))
	)

	for _, nat := range fullList {

		if optRoundFilter != nil && nat.round != *optRoundFilter {
			continue
		}

		if optVisibilityFilter != nil && *nat.visibility != *optVisibilityFilter {
			continue
		}

		res = append(res, nat)
	}

	return res
}

func (nat *ColNatural) ChangeVisibility(v Visibility) {
	*nat.visibility = v
}

func (nat *ColNatural) Visibility() Visibility {
	return *nat.visibility
}

func (nat *ColNatural) GetAssignment(run Runtime) smartvectors.SmartVector {
	n, ok := run.tryGetColumn(nat)
	if !ok {
		utils.Panic("assignment for column %v is missing. Explainer: \n%v", nat.String(), nat.Explain())
	}
	return n
}

func (nat *ColNatural) GetAssignmentGnark(_ frontend.API, run GnarkRuntime) []frontend.Variable {
	n, ok := run.tryGetColumn(nat)
	if !ok {
		utils.Panic("assignment for column %v is missing. Explainer: \n%v", nat.String(), nat.Explain())
	}
	return n
}

func (nat *ColNatural) Size() int {
	return nat.size
}

func (nat *ColNatural) Round() int {
	return nat.round
}

func (nat *ColNatural) Shift(n int) Column {

	if n == 0 {
		return nat
	}

	return &ColShifted{
		parent: nat,
		offset: n,
	}
}

func (nat *ColNatural) AssignConstant(run *RuntimeProver, x field.Element) {
	run.columns.InsertNew(nat.id(), smartvectors.NewConstant(x, nat.Size()))
}

func (nat *ColNatural) AssignLeftPadded(run *RuntimeProver, vec []field.Element, x field.Element) {
	run.columns.InsertNew(nat.id(), smartvectors.LeftPadded(vec, x, nat.Size()))
}

func (nat *ColNatural) AssignRightPadded(run *RuntimeProver, vec []field.Element, x field.Element) {
	run.columns.InsertNew(nat.id(), smartvectors.RightPadded(vec, x, nat.Size()))
}

func (nat *ColNatural) AssignLeftZeroPadded(run *RuntimeProver, vec []field.Element, x field.Element) {
	run.columns.InsertNew(nat.id(), smartvectors.LeftZeroPadded(vec, nat.Size()))
}

func (nat *ColNatural) AssignRightZeroPadded(run *RuntimeProver, vec []field.Element, x field.Element) {
	run.columns.InsertNew(nat.id(), smartvectors.RightZeroPadded(vec, nat.Size()))
}

func (nat *ColNatural) AssignSlice(run *RuntimeProver, vec []field.Element) {
	run.columns.InsertNew(nat.id(), smartvectors.NewRegular(vec))
}
