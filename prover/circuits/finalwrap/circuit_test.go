package finalwrap

import (
	"context"
	"testing"

	frBn254 "github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/linea-monorepo/prover/circuits"
	"github.com/consensys/linea-monorepo/prover/crypto/poseidon2_koalabear"
	"github.com/consensys/linea-monorepo/prover/crypto/ringsis"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/globalcs"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/localcs"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/logderivativesum"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/mpts"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/splitextension"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/univariates"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/vortex"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/stretchr/testify/require"
)

func init() {
	if err := poseidon2_koalabear.RegisterGates(); err != nil {
		panic(err)
	}
}

// toyWizardSuite returns a minimal compilation suite for a toy wizard IOP.
// It compiles through Vortex with PremarkAsSelfRecursed so the compiled IOP
// is suitable for wrapping in a BN254 final wrap circuit.
func toyWizardSuite() []func(*wizard.CompiledIOP) {
	return []func(*wizard.CompiledIOP){
		logderivativesum.CompileLookups,
		localcs.Compile,
		globalcs.Compile,
		univariates.Naturalize,
		mpts.Compile(),
		splitextension.CompileSplitExtToBase,
		vortex.Compile(
			2, true, // IsLastRound=true since this is the final wizard proof
			vortex.ForceNumOpenedColumns(4),
			vortex.WithSISParams(&ringsis.StdParams),
			vortex.WithOptionalSISHashingThreshold(64),
		),
	}
}

// TestMakeCS verifies that the BN254 final wrap circuit can be compiled
// around a toy wizard IOP.
func TestMakeCS(t *testing.T) {
	// Create and compile a tiny wizard IOP
	define := func(bui *wizard.Builder) {
		a := bui.RegisterCommit(ifaces.ColID("FW_A"), 64)
		b := bui.RegisterCommit(ifaces.ColID("FW_B"), 64)
		bui.Inclusion(ifaces.QueryID("FW_Q"), []ifaces.Column{a}, []ifaces.Column{b})
	}

	comp := wizard.Compile(define, toyWizardSuite()...)
	require.NotNil(t, comp)
	t.Logf("Toy wizard IOP compiled: %d rounds", comp.NumRounds())

	// Compile BN254 final wrap around it
	ccs, err := MakeCS(comp)
	require.NoError(t, err, "BN254 final wrap circuit should compile")
	t.Logf("Final wrap BN254 circuit: %d constraints", ccs.GetNbConstraints())
}

// TestFinalWrapEndToEnd tests the full flow:
// wizard IOP compile → prove → verify → BN254 wrap compile → PLONK setup → prove → verify
func TestFinalWrapEndToEnd(t *testing.T) {

	// 1. Create and compile a tiny wizard IOP
	define := func(bui *wizard.Builder) {
		a := bui.RegisterCommit(ifaces.ColID("E2E_A"), 64)
		b := bui.RegisterCommit(ifaces.ColID("E2E_B"), 64)
		bui.Inclusion(ifaces.QueryID("E2E_Q"), []ifaces.Column{a}, []ifaces.Column{b})
	}

	comp := wizard.Compile(define, toyWizardSuite()...)
	require.NotNil(t, comp)
	t.Logf("Wizard IOP compiled: %d rounds", comp.NumRounds())

	// 2. Prove the wizard IOP (full proof, all rounds)
	proverFunc := func(run *wizard.ProverRuntime) {
		run.AssignColumn(ifaces.ColID("E2E_A"), smartvectors.NewConstant(field.Zero(), 64))
		run.AssignColumn(ifaces.ColID("E2E_B"), smartvectors.NewConstant(field.Zero(), 64))
	}
	wizardProof := wizard.Prove(comp, proverFunc)

	// 3. Verify the wizard proof (sanity check)
	err := wizard.Verify(comp, wizardProof)
	require.NoError(t, err, "wizard proof should verify")
	t.Log("Wizard proof verified OK")

	// 4. Compile the BN254 final wrap circuit
	ccs, err := MakeCS(comp)
	require.NoError(t, err, "final wrap circuit should compile")
	t.Logf("Final wrap circuit: %d constraints", ccs.GetNbConstraints())

	// 5. Create PLONK setup (in-memory, unsafe for testing)
	srsProvider := circuits.NewUnsafeSRSProvider()
	setup, err := circuits.MakeSetup(
		context.Background(),
		circuits.FinalWrapCircuitID,
		ccs,
		srsProvider,
		nil,
	)
	require.NoError(t, err, "PLONK setup should succeed")
	t.Log("PLONK setup complete")

	// 6. Generate BN254 proof (MakeProof also verifies internally via ProveCheck)
	// No public inputs in the toy wizard → computePIDigest returns 0 → publicInput must be zero
	var publicInput frBn254.Element
	proof, err := MakeProof(&setup, comp, wizardProof, publicInput)
	require.NoError(t, err, "BN254 final wrap proof should succeed")
	require.NotNil(t, proof)

	t.Log("BN254 final wrap proof generated and verified successfully")
}
