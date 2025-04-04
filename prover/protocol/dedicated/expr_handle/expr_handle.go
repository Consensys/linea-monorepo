package expr_handle

import (
	"fmt"
	"reflect"

	sv "github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/fft"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/variables"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/protocol/wizardutils"
	"github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/sirupsen/logrus"
)

// Create a handle from an expression. The name is
// optional, if not set a generic name will be derived
// from the ESH of the expression.
func ExprHandle(comp *wizard.CompiledIOP, expr *symbolic.Expression, name ...string) ifaces.Column {

	var (
		boarded    = expr.Board()
		maxRound   = wizardutils.LastRoundToEval(expr)
		length     = wizardutils.ExprIsOnSameLengthHandles(&boarded)
		handleName = fmt.Sprintf("SYMBOLIC_%v", expr.ESHash.String())
	)

	if len(name) > 0 {
		handleName = name[0]
	}

	res := comp.InsertCommit(maxRound, ifaces.ColID(handleName), length)
	cs := comp.InsertGlobal(maxRound, ifaces.QueryID(handleName), expr.Sub(ifaces.ColumnAsVariable(res)))

	comp.RegisterProverAction(maxRound, &exprHandleProverAction{
		handleName: handleName,
		maxRound:   maxRound,
		boarded:    boarded,
		cs:         cs,
		length:     length,
	})

	return res
}

// exprHandleProverAction is the action to assign the expression handle columns.
// It implements the [wizard.ProverAction] interface.
type exprHandleProverAction struct {
	handleName string
	maxRound   int
	boarded    symbolic.ExpressionBoard
	cs         query.GlobalConstraint
	length     int
}

// Run executes the exprHandleProverAction over a [ProverRuntime]
func (action *exprHandleProverAction) Run(run *wizard.ProverRuntime) {

	logrus.Tracef("running the expr handle assignment for %v, (round %v)", action.handleName, action.maxRound)

	metadatas := action.boarded.ListVariableMetadata()

	/*
		Sanity-check : All witnesses should have the same size as the expression
	*/
	for _, metadataInterface := range metadatas {
		if handle, ok := metadataInterface.(ifaces.Column); ok {
			witness := handle.GetColAssignment(run)
			if witness.Len() != action.cs.DomainSize {
				utils.Panic(
					"Query %v - Witness of %v has size %v  which is below %v",
					action.cs.ID, handle.String(), witness.Len(), action.cs.DomainSize,
				)
			}
		}
	}

	/*
		Collects the relevant datas into a slice for the evaluation
	*/
	evalInputs := make([]sv.SmartVector, len(metadatas))

	/*
		Omega is a root of unity which generates the domain of evaluation
		of the constraint. Its size coincide with the size of the domain
		of evaluation. For each value of `i`, X will evaluate to omega^i.
	*/
	omega := fft.GetOmega(action.cs.DomainSize)
	omegaI := field.One()

	// precomputations of the powers of omega, can be optimized if useful
	omegas := make([]field.Element, action.cs.DomainSize)
	for i := 0; i < action.cs.DomainSize; i++ {
		omegas[i] = omegaI
		omegaI.Mul(&omegaI, &omega)
	}

	/*
		Collect the relevants inputs for evaluating the constraint
	*/
	for k, metadataInterface := range metadatas {
		switch meta := metadataInterface.(type) {
		case ifaces.Column:
			w := meta.GetColAssignment(run)
			evalInputs[k] = w
		case coin.Info:
			// Implicitly, the coin has to be a field element in the expression
			// It will panic if not
			x := run.GetRandomCoinField(meta.Name)
			evalInputs[k] = sv.NewConstant(x, action.length)
		case variables.X:
			evalInputs[k] = meta.EvalCoset(action.length, 0, 1, false)
		case variables.PeriodicSample:
			evalInputs[k] = meta.EvalCoset(action.length, 0, 1, false)
		case ifaces.Accessor:
			evalInputs[k] = sv.NewConstant(meta.GetVal(run), action.length)
		default:
			utils.Panic("Not a variable type %v in query %v", reflect.TypeOf(metadataInterface), action.cs.ID)
		}
	}

	// This panics if the global constraints doesn't use any commitment
	resWitness := action.boarded.Evaluate(evalInputs)

	run.AssignColumn(ifaces.ColID(action.handleName), resWitness)
}
