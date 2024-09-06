package plonk

import (
	"os"
	"testing"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/zkevm-monorepo/prover/maths/field"
	"github.com/consensys/zkevm-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
	"github.com/consensys/zkevm-monorepo/prover/utils/csvtraces"
)

const nbInputInstances = 3
const nbCircuitInstances = 2

func TestAlignment(t *testing.T) {
	f, err := os.Open("testdata/alignment.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	ct, err := csvtraces.NewCsvTrace(f)
	if err != nil {
		t.Fatal(err)
	}
	var toAlign *CircuitAlignmentInput
	var alignment *Alignment
	// plonk in wizard is deferred, so we need to capture the runtime to be able
	// to verify concatenated PI column assignment
	var runLeaked *wizard.ProverRuntime
	cmp := wizard.Compile(func(build *wizard.Builder) {
		toAlign = &CircuitAlignmentInput{
			Name:               "ALIGNMENT_TEST",
			Circuit:            &DummyAlignmentCircuit{Instances: make([]DummyAlignmentCircuitInstance, nbInputInstances)},
			DataToCircuit:      ct.GetCommit(build, "DATA"),
			DataToCircuitMask:  ct.GetCommit(build, "DATA_MASK"),
			NbCircuitInstances: nbCircuitInstances,
			InputFiller:        func(circuitInstance, inputIndex int) field.Element { return field.NewElement(uint64(inputIndex + 1)) },
		}
		alignment = DefineAlignment(build.CompiledIOP, toAlign)
	}, dummy.Compile)
	proof := wizard.Prove(cmp, func(run *wizard.ProverRuntime) {
		runLeaked = run
		ct.Assign(run, "DATA", "DATA_MASK")
		alignment.Assign(run)
	})

	ct.CheckAssignmentColumn(runLeaked, "IS_ACTIVE", alignment.IsActive)
	ct.CheckAssignmentColumn(runLeaked, "CIRCUIT_INPUT", alignment.CircuitInput)
	ct.CheckAssignmentColumn(runLeaked, "FULL_CIRCUIT_INPUT_MASK", alignment.FullCircuitInputMask)
	ct.CheckAssignmentColumn(runLeaked, "ACTUAL_CIRCUIT_INPUT_MASK", alignment.ActualCircuitInputMask)
	if err := wizard.Verify(cmp, proof); err != nil {
		t.Fatal("proof failed", err)
	}
	t.Log("proof succeeded")
}

// DummyAlignmentCircuit is a dummy circuit for testing alignment. It doesn't do
// anything except check that the inputs are in order.
type DummyAlignmentCircuit struct {
	Instances []DummyAlignmentCircuitInstance `gnark:",public"`
}

type DummyAlignmentCircuitInstance struct {
	Vars [6]frontend.Variable `gnark:",public"` // non-power of two to test padding. public input for asserting inputs only
}

func (c *DummyAlignmentCircuit) Define(api frontend.API) error {
	var counter frontend.Variable = 1
	for i := range c.Instances {
		for j := range c.Instances[i].Vars {
			api.AssertIsEqual(c.Instances[i].Vars[j], counter)
			counter = api.Add(counter, 1)
		}
	}
	return nil
}
