package distributedprojection

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
)

type distributedProjectionVerifierAction struct {
	Name               ifaces.QueryID
	HornerA0, HornerB0 []query.LocalOpening
	isA, isB           []bool
	skipped            bool
}

// Run implements the [wizard.VerifierAction]
func (va *distributedProjectionVerifierAction) Run(run *wizard.VerifierRuntime) error {
	var (
		actualParam = field.Zero()
	)
	for index := range va.HornerA0 {
		var (
			elemParam = field.Zero()
		)
		if va.isA[index] && va.isB[index] {
			elemParam = run.GetLocalPointEvalParams(va.HornerB0[index].ID).Y
			elemParam.Neg(&elemParam)
			temp := run.GetLocalPointEvalParams(va.HornerA0[index].ID).Y
			elemParam.Add(&elemParam, &temp)
		} else if va.isA[index] && !va.isB[index] {
			elemParam = run.GetLocalPointEvalParams(va.HornerA0[index].ID).Y
		} else if !va.isA[index] && va.isB[index] {
			elemParam = run.GetLocalPointEvalParams(va.HornerB0[index].ID).Y
			elemParam.Neg(&elemParam)
		} else {
			utils.Panic("Unsupported verifier action registered for %v", va.Name)
		}
		actualParam.Add(&actualParam, &elemParam)
	}
	queryParam := run.GetDistributedProjectionParams(va.Name).HornerVal
	if actualParam != queryParam {
		utils.Panic("The distributed projection query %v did not pass, query param %v and actual param %v", va.Name, queryParam, actualParam)
	}
	return nil
}

// RunGnark implements the [wizard.VerifierAction] interface.
func (va *distributedProjectionVerifierAction) RunGnark(api frontend.API, run *wizard.WizardVerifierCircuit) {

	var (
		actualParam = frontend.Variable(0)
	)
	for index := range va.HornerA0 {
		var (
			elemParam = frontend.Variable(0)
		)
		if va.isA[index] && va.isB[index] {
			var (
				a, b frontend.Variable
			)
			a = run.GetLocalPointEvalParams(va.HornerA0[index].ID).Y
			b = run.GetLocalPointEvalParams(va.HornerB0[index].ID).Y
			elemParam = api.Sub(a, b)
		} else if va.isA[index] && !va.isB[index] {
			a := run.GetLocalPointEvalParams(va.HornerA0[index].ID).Y
			elemParam = api.Add(elemParam, a)
		} else if !va.isA[index] && va.isB[index] {
			b := run.GetLocalPointEvalParams(va.HornerB0[index].ID).Y
			elemParam = api.Sub(elemParam, b)
		} else {
			utils.Panic("Unsupported verifier action registered for %v", va.Name)
		}
		actualParam = api.Add(actualParam, elemParam)
	}
	queryParam := run.GetDistributedProjectionParams(va.Name).Sum

	api.AssertIsEqual(actualParam, queryParam)
}

// Skip implements the [wizard.VerifierAction]
func (va *distributedProjectionVerifierAction) Skip() {
	va.skipped = true
}

// IsSkipped implements the [wizard.VerifierAction]
func (va *distributedProjectionVerifierAction) IsSkipped() bool {
	return va.skipped
}
