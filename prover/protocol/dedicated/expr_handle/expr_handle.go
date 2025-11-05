spackage expr_handle

import (
	"reflect"

	"github.com/consensys/gnark-crypto/field/koalabear/fft"
	sv "github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/variables"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/protocol/wizardutils"
	"github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/sirupsen/logrus"
)

type ExprHandleProverAction struct {
	Expr       *symbolic.Expression
	HandleName ifaces.ColID
	MaxRound   int
}

func (a *ExprHandleProverAction) DomainSize() int {
	_, _, n := wizardutils.AsExpr(a.Expr)
	return n
}

func (a *ExprHandleProverAction) Run(run *wizard.ProverRuntime) {

	boarded := a.Expr.Board()
	logrus.Tracef("running the expr handle assignment for %v, (round %v)", a.HandleName, a.MaxRound)
	metadatas := boarded.ListVariableMetadata()

	domainSize := a.DomainSize()

	evalInputs := make([]sv.SmartVector, len(metadatas))
	omega, err := fft.Generator(uint64(a.DomainSize()))
	if err != nil {
		// should not happen unless we have a very very large domain size
		utils.Panic("failed to generate omega for %v, size=%v", a.HandleName, a.DomainSize())
	}
	omegaI := field.One()
	omegas := make([]field.Element, a.DomainSize())
	for i := 0; i < a.DomainSize(); i++ {
		omegas[i] = omegaI
		omegaI.Mul(&omegaI, &omega)
	}

	for k, metadataInterface := range metadatas {
		switch meta := metadataInterface.(type) {
		case ifaces.Column:
			w := meta.GetColAssignment(run)
			if w.Len() != domainSize {
				utils.Panic("Query %v - Witness of %v has size %v which is below %v",
					a.HandleName, meta.String(), w.Len(), domainSize)
			}
			evalInputs[k] = w
		case coin.Info:
			if meta.IsBase() {
				utils.Panic("unsupported, coins are always over field extensions")

			} else {
				x := run.GetRandomCoinFieldExt(meta.Name)
				evalInputs[k] = sv.NewConstantExt(x, a.DomainSize())
			}
		case variables.X:
			evalInputs[k] = meta.EvalCoset(domainSize, 0, 1, false)
		case variables.PeriodicSample:
			evalInputs[k] = meta.EvalCoset(domainSize, 0, 1, false)
		case ifaces.Accessor:
			if metadataInterface.IsBase() {
				elem, errFetch := meta.GetValBase(run)
				if errFetch != nil {
					utils.Panic("failed to fetch base accessor %v for query %v: %v", meta.String(), a.HandleName, errFetch)
				}
				evalInputs[k] = sv.NewConstant(elem, a.DomainSize())
			} else {
				evalInputs[k] = sv.NewConstantExt(meta.GetValExt(run), a.DomainSize())
			}

		default:
			utils.Panic("Not a variable type %v in query %v", reflect.TypeOf(metadataInterface), a.HandleName)
		}
	}

	resWitness := boarded.Evaluate(evalInputs)
	run.AssignColumn(a.HandleName, resWitness)
}

// Create a handle from an expression.
func ExprHandle(comp *wizard.CompiledIOP, expr *symbolic.Expression, handleName string) ifaces.Column {

	var (
		boarded  = expr.Board()
		maxRound = wizardutils.LastRoundToEval(expr)
		length   = column.ExprIsOnSameLengthHandles(&boarded)
	)

	res := comp.InsertCommit(maxRound, ifaces.ColID(handleName), length, expr.IsBase)
	comp.InsertGlobal(maxRound, ifaces.QueryID(handleName), expr.Sub(ifaces.ColumnAsVariable(res)))

	comp.RegisterProverAction(maxRound, &ExprHandleProverAction{
		Expr:       expr,
		HandleName: ifaces.ColID(handleName),
		MaxRound:   maxRound,
	})
	return res
}
