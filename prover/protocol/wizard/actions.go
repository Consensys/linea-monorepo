package wizard

import (
	"fmt"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/zk"
)

// ProverAction represents an action to be performed by the prover.
// They have to be registered in the [CompiledIOP] via the
// [CompiledIOP.RegisterProverAction]
type ProverAction[T zk.Element] interface {
	// Run executes the ProverAction over a [ProverRuntime]
	Run(*ProverRuntime[T])
}

// mainProverStepWrapper adapts a  MainProverStep to the ProverAction interface.
type mainProverStepWrapper[T zk.Element] struct {
	step func(*ProverRuntime[T])
}

// Run implements the ProverAction interface for MainProverStep.
func (w mainProverStepWrapper[T]) Run(run *ProverRuntime[T]) {
	w.step(run)
}

// VerifierAction represents an action to be performed by the verifier of the
// protocol. Usually, this is used to represent verifier checks. They can be
// registered via [CompiledIOP.RegisterVerifierAction].
type VerifierAction[T zk.Element] interface {
	// Run executes the VerifierAction over a [VerifierRuntime] it returns an
	// error.
	Run(Runtime) error
	// RunGnark is as Run but in a gnark circuit. Instead, of the returning an
	// error the function enforces the passing of the verifier's checks.
	RunGnark(frontend.API, GnarkRuntime[T])
}

// PrintingProverAction is a ProverAction printing a column content. And an
// additional message.
type PrintingProverAction[T zk.Element] struct {
	Column          ifaces.Column[T]
	Message         string
	NameReplacement string
}

// PrintingVerifierAction is as PrintingProverAction but for the verifier.
type PrintingVerifierAction[T zk.Element] struct {
	Column          ifaces.Column[T]
	Message         string
	NameReplacement string
}

// Run implements the ProverAction interface for PrintingProverAction.
func (p PrintingProverAction[T]) Run(run *ProverRuntime[T]) {

	name := p.Column.GetColID()
	if len(p.NameReplacement) > 0 {
		name = ifaces.ColID(p.NameReplacement)
	}

	c := p.Column.GetColAssignment(run)
	fmt.Printf("name=%v message=%v value=%v\n", name, p.Message, c.Pretty())
}

// Run implements the VerifierAction interface for PrintingVerifierAction.
func (p PrintingVerifierAction[T]) Run(run Runtime) error {

	name := p.Column.GetColID()
	if len(p.NameReplacement) > 0 {
		name = ifaces.ColID(p.NameReplacement)
	}

	c := p.Column.GetColAssignment(run)
	fmt.Printf("name=%v message=%v value=%v\n", name, p.Message, c.Pretty())
	return nil
}

// RunGnark implements the VerifierAction interface for PrintingVerifierAction.
func (p PrintingVerifierAction[T]) RunGnark(api frontend.API, run GnarkRuntime[T]) {

	name := p.Column.GetColID()
	if len(p.NameReplacement) > 0 {
		name = ifaces.ColID(p.NameReplacement)
	}

	c := p.Column.GetColAssignmentGnark(run)
	names := fmt.Sprintf("name=%v message=%v value=\n", name, p.Message)
	// api.Println(append([]T{names}, c...)...)
	api.Println(names)
	apiGen, err := zk.NewApi[T](api)
	if err != nil {
		panic(err)
	}
	for i := 0; i < len(c); i++ {
		apiGen.Println(&c[i])
	}
}
