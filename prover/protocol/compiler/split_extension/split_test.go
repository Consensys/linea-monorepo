package splitextension

import (
	"fmt"
	"testing"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/stretchr/testify/assert"
)

type testCase struct {
	nbColumnsToSplit int
	size             int

	// original vector that we want to split defined over E4
	toSplit []smartvectors.SmartVector

	challenge fext.Element
}

// createMultiIPTestCase generates a testCase for multiple inner products (linear combination).
func createSplittingTestCase(
	splittedName string,
	toSplitName string,
	sizeColumns int,
	nbColumnsToSplit int,
) testCase {

	toSplit := make([][]fext.Element, nbColumnsToSplit)
	for i := 0; i < nbColumnsToSplit; i++ {
		toSplit[i] = make([]fext.Element, sizeColumns)
		for k := 0; k < sizeColumns; k++ {
			toSplit[i][k].SetRandom()
		}
	}

	smToSplit := make([]smartvectors.SmartVector, nbColumnsToSplit)
	for i := 0; i < nbColumnsToSplit; i++ {
		smToSplit[i] = smartvectors.NewRegularExt(toSplit[i])
	}

	toSplitnames := make([]ifaces.ColID, nbColumnsToSplit)
	splittedNames := make([]ifaces.ColID, 4*nbColumnsToSplit)
	for i := 0; i < nbColumnsToSplit; i++ {
		toSplitnames[i] = ifaces.ColID(fmt.Sprintf("%s_%d", toSplitName, i))
		for j := 0; j < 4; j++ {
			splittedNames[4*i+j] = ifaces.ColID(fmt.Sprintf("%s_%d", splittedName, 4*i+j))
		}
	}

	// TODO use PRNG
	var x fext.Element
	x.B0.A0.SetUint64(2)

	return testCase{

		nbColumnsToSplit: nbColumnsToSplit,
		size:             sizeColumns,

		// original vector that we want to split defined over E4
		toSplit: smToSplit,

		challenge: x,
	}

}

func TestSplitextension(t *testing.T) {

	// size of the columns
	sizeColumns := 64
	nbColumnsToSplit := 2
	testCase := createSplittingTestCase("splitted_cols", "to_split", sizeColumns, nbColumnsToSplit)

	define := func(b *wizard.Builder) {
		toSplit := make([]ifaces.Column, nbColumnsToSplit)
		for i := 0; i < nbColumnsToSplit; i++ {
			curName := fmt.Sprintf("%s_%d", baseNameToSplit, i)
			toSplit[i] = b.RegisterCommit(ifaces.ColID(curName), sizeColumns)
		}
		b.InsertUnivariate(0, fextQuery, toSplit)

	}

	witness := func(run *wizard.ProverRuntime) {
		y := make([]fext.Element, testCase.nbColumnsToSplit)
		for i := 0; i < nbColumnsToSplit; i++ {
			curName := fmt.Sprintf("%s_%d", baseNameToSplit, i)
			run.AssignColumn(ifaces.ColID(curName), testCase.toSplit[i])
			y[i] = smartvectors.EvaluateLagrangeFullFext(testCase.toSplit[i], testCase.challenge)
		}
		run.AssignUnivariate(fextQuery, testCase.challenge, y...)
	}

	comp := wizard.Compile(define, CompileSplitExtToBase, dummy.Compile)

	proof := wizard.Prove(comp, witness)
	assert.NoErrorf(t, wizard.Verify(comp, proof), "invalid proof")

}
