package test_cases_test

import (
	"testing"

	"github.com/consensys/accelerated-crypto-monorepo/crypto/mimc/gkrmimc"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/coin"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/compiler/arithmetics"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/compiler/dummy"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/compiler/specialqueries"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/compiler/splitter"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/compiler/univariates"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/compiler/vortex"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/ifaces"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/wizard"
	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

/*
Common identifiers, that can be reused across tests
*/
var (
	P1                ifaces.ColID   = "P1"
	P2                ifaces.ColID   = "P2"
	P3                ifaces.ColID   = "P3"
	P4                ifaces.ColID   = "P4"
	COIN1             coin.Name      = "C1"
	COIN2             coin.Name      = "C2"
	UNIV1             ifaces.QueryID = "UNIV1"
	UNIV2             ifaces.QueryID = "UNIV2"
	GLOBAL1           ifaces.QueryID = "GLOBAL1"
	GLOBAL2           ifaces.QueryID = "GLOBAL2"
	LOCAL1            ifaces.QueryID = "LOCAL1"
	LOOKUP1           ifaces.QueryID = "LOOKUP1"
	LOOKUP2           ifaces.QueryID = "LOOKUP2"
	PERMUTATION1      ifaces.QueryID = "PERMUTATION1"
	PERMUTATION2      ifaces.QueryID = "PERMUTATION2"
	FIXEDPERMUTATION1 ifaces.QueryID = "FIXEDPERMUTATION1"
	FIXEDPERMUTATION2 ifaces.QueryID = "FIXEDPERMUTATION2"
	RANGE1            ifaces.QueryID = "RANGE1"
	RANGE2            ifaces.QueryID = "RANGE2"
	RANGE3            ifaces.QueryID = "RANGE3"
	RANGE4            ifaces.QueryID = "RANGE4"
)

/*
Represents a list of compilation steps
*/
type compilationSuite []func(*wizard.CompiledIOP)

/*
Various compilations relevants for the compilation steps
*/
var (
	ALL_SPECIALS = compilationSuite{
		specialqueries.RangeProof,
		specialqueries.CompileFixedPermutations,
		specialqueries.LogDerivativeLookupCompiler,
		specialqueries.CompilePermutations,
		specialqueries.CompileInnerProduct,
	}
	ARITHMETICS = compilationSuite{
		splitter.SplitColumns(8),
		arithmetics.CompileLocal,
		arithmetics.CompileGlobal,
	}
	UNIVARIATES = compilationSuite{
		univariates.CompileLocalOpening,
		univariates.Naturalize,
		univariates.MultiPointToSinglePoint(8),
	}
	DUMMY       = compilationSuite{dummy.Compile}
	TENSOR      = compilationSuite{vortex.Compile(2, vortex.WithDryThreshold(1))} // dummy unsafe sis instance
	ALL_BUT_ILC = join(ALL_SPECIALS, ARITHMETICS, UNIVARIATES, DUMMY)
	WITH_TENSOR = join(ALL_SPECIALS, ARITHMETICS, UNIVARIATES, TENSOR)
)

func join(suites ...compilationSuite) compilationSuite {
	res := compilationSuite{}
	for _, s := range suites {
		res = append(res, s...)
	}
	return res
}

/*
Wraps the wizard verification gnark into a circuit
*/
type SimpleTestGnarkCircuit struct {
	C wizard.WizardVerifierCircuit
}

/*
Just verify the wizard-IOP equation, also verifies that
that the "x" is correctly set.
*/
func (c *SimpleTestGnarkCircuit) Define(api frontend.API) error {
	c.C.Verify(api)
	return nil
}

/*
Returns an assignment from a wizard proof
*/
func GetAssignment(comp *wizard.CompiledIOP, proof wizard.Proof) *SimpleTestGnarkCircuit {
	return &SimpleTestGnarkCircuit{
		C: *wizard.GetWizardVerifierCircuitAssignment(comp, proof),
	}
}

/*
The test verifies that the test pass through all layers of compilation
*/
func checkSolved(
	t *testing.T,
	define wizard.DefineFunc,
	prove wizard.ProverStep,
	suite compilationSuite,
	testCircuit bool,
	expectFail ...bool,
) {

	// Activate the logging in trace mode by default
	logrus.SetLevel(logrus.TraceLevel)

	// As this relies on the dummy compile, this does not
	compiled := wizard.Compile(define, suite...)

	proof := wizard.Prove(compiled, prove)
	err := wizard.Verify(compiled, proof)

	if !testCircuit {
		return
	}

	if err != nil {
		if len(expectFail) > 0 {
			// expected a failure
			return
		}
		t.Fatalf(err.Error())
	}

	/*
		Allocate the circuit
	*/
	circ := SimpleTestGnarkCircuit{}
	{
		c, err := wizard.AllocateWizardCircuit(compiled)
		if err != nil {
			// The only error case acknowledged here is that the returned circuit
			// is empty. In that case, there is simply no point to run the verification.
			return
		}

		circ.C = *c
	}

	scs, err := frontend.Compile(
		ecc.BN254.ScalarField(),
		scs.NewBuilder,
		&circ,
		frontend.IgnoreUnconstrainedInputs(),
	)

	if err != nil {
		// When the error string is too large `require.NoError` does not print
		// the error.
		t.Logf("circuit construction failed : %v\n", err)
		t.FailNow()
	}

	assignment := GetAssignment(compiled, proof)
	witness, err := frontend.NewWitness(assignment, ecc.BN254.ScalarField())
	require.NoError(t, err)

	err = scs.IsSolved(witness, gkrmimc.SolverOpts(scs)...)

	if err != nil {
		// When the error string is too large `require.NoError` does not print
		// the error.
		t.Logf("circuit solving failed : %v\n", err)
		t.FailNow()
	}
}
