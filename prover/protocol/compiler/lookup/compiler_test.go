package lookup

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogDerivativeLookupSimple(t *testing.T) {

	var sizeA, sizeB int = 16, 8

	define := func(b *wizard.Builder) {
		cola := b.RegisterCommit("A", sizeA)
		colb := b.RegisterCommit("B", sizeB)
		b.Inclusion("LOOKUP", []ifaces.Column{cola}, []ifaces.Column{colb})
	}

	prover := func(run *wizard.ProverRuntime) {
		// assign a and b
		cola := smartvectors.ForTest(0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15)
		colb := smartvectors.ForTest(0, 1, 2, 3, 4, 5, 6, 7)
		run.AssignColumn("A", cola)
		run.AssignColumn("B", colb)
	}

	comp := wizard.Compile(define, CompileLogDerivative, dummy.Compile)
	proof := wizard.Prove(comp, prover)
	err := wizard.Verify(comp, proof)
	require.NoError(t, err)
}

func TestLogDerivativeLookupSimple2(t *testing.T) {

	var sizeA, sizeB int = 16, 4
	var runtime *wizard.ProverRuntime

	define := func(b *wizard.Builder) {
		cola := b.RegisterCommit("S", sizeA)
		colb := b.RegisterCommit("T", sizeB)
		b.Inclusion("LOOKUP", []ifaces.Column{colb}, []ifaces.Column{cola})
	}

	prover := func(run *wizard.ProverRuntime) {
		runtime = run
		// assign a and b
		cola := smartvectors.ForTest(1, 1, 1, 2, 3, 0, 0, 1, 1, 1, 1, 2, 3, 0, 0, 1)
		colb := smartvectors.ForTest(0, 1, 2, 3)
		// m expected to be
		run.AssignColumn("S", cola)
		run.AssignColumn("T", colb)
	}

	comp := wizard.Compile(define, CompileLogDerivative, dummy.Compile)
	proof := wizard.Prove(comp, prover)

	// m should be
	expectedM := smartvectors.ForTest(4, 8, 2, 2)
	t.Logf("all columns = %v", runtime.Columns.ListAllKeys())
	actualM := runtime.GetColumn("TABLE_T_0_LOGDERIVATIVE_M")

	assert.Equal(t, expectedM.Pretty(), actualM.Pretty(), "m does not match the expected value")

	err := wizard.Verify(comp, proof)
	require.NoError(t, err)
}

func TestLogDerivativeLookupManyChecksOneTable(t *testing.T) {

	var sizeA, sizeB int = 16, 4
	var runtime *wizard.ProverRuntime

	define := func(b *wizard.Builder) {
		cola := b.RegisterCommit("S", sizeA)
		cola2 := b.RegisterCommit("S2", sizeA)
		colb := b.RegisterCommit("T", sizeB)
		b.Inclusion("LOOKUP", []ifaces.Column{colb}, []ifaces.Column{cola})
		b.Inclusion("LOOKUP2", []ifaces.Column{colb}, []ifaces.Column{cola2})
	}

	prover := func(run *wizard.ProverRuntime) {
		runtime = run
		// assign a and b
		cola := smartvectors.ForTest(1, 1, 1, 2, 3, 0, 0, 1, 1, 1, 1, 2, 3, 0, 0, 1)
		cola2 := smartvectors.ForTest(2, 2, 2, 1, 0, 3, 3, 2, 2, 2, 2, 1, 0, 3, 3, 2)
		colb := smartvectors.ForTest(0, 1, 2, 3)
		// m expected to be
		run.AssignColumn("S", cola)
		run.AssignColumn("S2", cola2)
		run.AssignColumn("T", colb)
	}

	comp := wizard.Compile(define, CompileLogDerivative, dummy.Compile)
	proof := wizard.Prove(comp, prover)

	// m should be
	expectedM := smartvectors.ForTest(6, 10, 10, 6)
	t.Logf("all columns = %v", runtime.Columns.ListAllKeys())
	actualM := runtime.GetColumn("TABLE_T_0_LOGDERIVATIVE_M")

	assert.Equal(t, expectedM.Pretty(), actualM.Pretty(), "m does not match the expected value")

	err := wizard.Verify(comp, proof)
	require.NoError(t, err)
}

func TestLogDerivativeLookupOneXor(t *testing.T) {

	var sizeTable, sizeCheckeds int = 16, 4
	var runtime *wizard.ProverRuntime

	// The test uses a lookup over a xor table

	define := func(b *wizard.Builder) {

		xorX := b.RegisterCommit("XOR_TABLE_X", sizeTable)
		xorY := b.RegisterCommit("XOR_TABLE_Y", sizeTable)
		xorXY := b.RegisterCommit("XOR_TABLE_XXORY", sizeTable)

		wX := b.RegisterCommit("WITNESS_X", sizeCheckeds)
		wY := b.RegisterCommit("WITNESS_Y", sizeCheckeds)
		wXY := b.RegisterCommit("WITNESS_XXORY", sizeCheckeds)

		b.Inclusion("LOOKUP", []ifaces.Column{xorX, xorY, xorXY}, []ifaces.Column{wX, wY, wXY})
	}

	prover := func(run *wizard.ProverRuntime) {
		runtime = run
		// assign a and b
		xorX := smartvectors.ForTest(0b00, 0b01, 0b10, 0b11, 0b00, 0b01, 0b10, 0b11, 0b00, 0b01, 0b10, 0b11, 0b00, 0b01, 0b10, 0b11)
		xorY := smartvectors.ForTest(0b00, 0b00, 0b00, 0b00, 0b01, 0b01, 0b01, 0b01, 0b10, 0b10, 0b10, 0b10, 0b11, 0b11, 0b11, 0b11)
		xorXY := smartvectors.ForTest(0b00, 0b01, 0b10, 0b11, 0b01, 0b00, 0b11, 0b10, 0b10, 0b11, 0b00, 0b01, 0b11, 0b10, 0b01, 0b00)

		wX := smartvectors.ForTest(0b00, 0b11, 0b10, 0b01)
		wY := smartvectors.ForTest(0b01, 0b00, 0b11, 0b10)
		wXY := smartvectors.ForTest(0b01, 0b11, 0b01, 0b11)

		run.AssignColumn("XOR_TABLE_X", xorX)
		run.AssignColumn("XOR_TABLE_Y", xorY)
		run.AssignColumn("XOR_TABLE_XXORY", xorXY)

		run.AssignColumn("WITNESS_X", wX)
		run.AssignColumn("WITNESS_Y", wY)
		run.AssignColumn("WITNESS_XXORY", wXY)
	}

	comp := wizard.Compile(define, CompileLogDerivative, dummy.Compile)
	proof := wizard.Prove(comp, prover)

	// m should be
	expectedM := smartvectors.ForTest(0, 0, 0, 1, 1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0)
	t.Logf("all columns = %v", runtime.Columns.ListAllKeys())
	actualM := runtime.GetColumn("TABLE_XOR_TABLE_X_XOR_TABLE_XXORY_XOR_TABLE_Y_0_LOGDERIVATIVE_M")

	assert.Equal(t, expectedM.Pretty(), actualM.Pretty(), "m does not match the expected value")

	err := wizard.Verify(comp, proof)
	require.NoError(t, err)
}

func TestLogDerivativeLookupMultiXor(t *testing.T) {

	var sizeTable, sizeCheckeds, sizeCheckedLarger int = 16, 4, 8
	var runtime *wizard.ProverRuntime

	// The test uses a lookup over a xor table

	define := func(b *wizard.Builder) {

		xorX := b.RegisterCommit("XOR_TABLE_X", sizeTable)
		xorY := b.RegisterCommit("XOR_TABLE_Y", sizeTable)
		xorXY := b.RegisterCommit("XOR_TABLE_XXORY", sizeTable)

		wX := b.RegisterCommit("WITNESS_X", sizeCheckeds)
		wY := b.RegisterCommit("WITNESS_Y", sizeCheckeds)
		wXY := b.RegisterCommit("WITNESS_XXORY", sizeCheckeds)

		w2X := b.RegisterCommit("W2_X", sizeCheckedLarger)
		w2Y := b.RegisterCommit("W2_Y", sizeCheckedLarger)
		w2XY := b.RegisterCommit("W2_XXORY", sizeCheckedLarger)

		b.Inclusion("LOOKUP", []ifaces.Column{xorX, xorY, xorXY}, []ifaces.Column{wX, wY, wXY})
		b.Inclusion("LOOKUP2", []ifaces.Column{xorX, xorY, xorXY}, []ifaces.Column{w2X, w2Y, w2XY})
	}

	prover := func(run *wizard.ProverRuntime) {
		runtime = run
		// assign a and b
		xorX := smartvectors.ForTest(0b00, 0b01, 0b10, 0b11, 0b00, 0b01, 0b10, 0b11, 0b00, 0b01, 0b10, 0b11, 0b00, 0b01, 0b10, 0b11)
		xorY := smartvectors.ForTest(0b00, 0b00, 0b00, 0b00, 0b01, 0b01, 0b01, 0b01, 0b10, 0b10, 0b10, 0b10, 0b11, 0b11, 0b11, 0b11)
		xorXY := smartvectors.ForTest(0b00, 0b01, 0b10, 0b11, 0b01, 0b00, 0b11, 0b10, 0b10, 0b11, 0b00, 0b01, 0b11, 0b10, 0b01, 0b00)

		wX := smartvectors.ForTest(0b00, 0b11, 0b10, 0b01)
		wY := smartvectors.ForTest(0b01, 0b00, 0b11, 0b10)
		wXY := smartvectors.ForTest(0b01, 0b11, 0b01, 0b11)

		w2X := smartvectors.ForTest(0b00, 0b01, 0b10, 0b11, 0b00, 0b01, 0b10, 0b11)
		w2Y := smartvectors.ForTest(0b00, 0b00, 0b00, 0b00, 0b01, 0b01, 0b01, 0b01)
		w2XY := smartvectors.ForTest(0b00, 0b01, 0b10, 0b11, 0b01, 0b00, 0b11, 0b10)

		run.AssignColumn("XOR_TABLE_X", xorX)
		run.AssignColumn("XOR_TABLE_Y", xorY)
		run.AssignColumn("XOR_TABLE_XXORY", xorXY)

		run.AssignColumn("WITNESS_X", wX)
		run.AssignColumn("WITNESS_Y", wY)
		run.AssignColumn("WITNESS_XXORY", wXY)

		run.AssignColumn("W2_X", w2X)
		run.AssignColumn("W2_Y", w2Y)
		run.AssignColumn("W2_XXORY", w2XY)
	}

	comp := wizard.Compile(define, CompileLogDerivative, dummy.Compile)
	proof := wizard.Prove(comp, prover)

	// m should be
	expectedM := smartvectors.ForTest(1, 1, 1, 2, 2, 1, 1, 1, 0, 1, 0, 0, 0, 0, 1, 0)
	t.Logf("all column names = %v", runtime.Columns.ListAllKeys())
	actualM := runtime.GetColumn("TABLE_XOR_TABLE_X_XOR_TABLE_XXORY_XOR_TABLE_Y_0_LOGDERIVATIVE_M")

	assert.Equal(t, expectedM.Pretty(), actualM.Pretty(), "m does not match the expected value")

	err := wizard.Verify(comp, proof)
	require.NoError(t, err)
}

func BenchmarkLogDeriveLookupMultiXor(b *testing.B) {
	for i := 0; i < b.N; i++ {

		var sizeTable, sizeCheckeds, sizeCheckedLarger int = 16, 4, 8

		// The test uses a lookup over a xor table

		define := func(b *wizard.Builder) {

			xorX := b.RegisterCommit("XOR_TABLE_X", sizeTable)
			xorY := b.RegisterCommit("XOR_TABLE_Y", sizeTable)
			xorXY := b.RegisterCommit("XOR_TABLE_XXORY", sizeTable)

			wX := b.RegisterCommit("WITNESS_X", sizeCheckeds)
			wY := b.RegisterCommit("WITNESS_Y", sizeCheckeds)
			wXY := b.RegisterCommit("WITNESS_XXORY", sizeCheckeds)

			w2X := b.RegisterCommit("W2_X", sizeCheckedLarger)
			w2Y := b.RegisterCommit("W2_Y", sizeCheckedLarger)
			w2XY := b.RegisterCommit("W2_XXORY", sizeCheckedLarger)

			b.Inclusion("LOOKUP", []ifaces.Column{xorX, xorY, xorXY}, []ifaces.Column{wX, wY, wXY})
			b.Inclusion("LOOKUP2", []ifaces.Column{xorX, xorY, xorXY}, []ifaces.Column{w2X, w2Y, w2XY})
		}

		prover := func(run *wizard.ProverRuntime) {

			// assign a and b
			xorX := smartvectors.ForTest(0b00, 0b01, 0b10, 0b11, 0b00, 0b01, 0b10, 0b11, 0b00, 0b01, 0b10, 0b11, 0b00, 0b01, 0b10, 0b11)
			xorY := smartvectors.ForTest(0b00, 0b00, 0b00, 0b00, 0b01, 0b01, 0b01, 0b01, 0b10, 0b10, 0b10, 0b10, 0b11, 0b11, 0b11, 0b11)
			xorXY := smartvectors.ForTest(0b00, 0b01, 0b10, 0b11, 0b01, 0b00, 0b11, 0b10, 0b10, 0b11, 0b00, 0b01, 0b11, 0b10, 0b01, 0b00)

			wX := smartvectors.ForTest(0b00, 0b11, 0b10, 0b01)
			wY := smartvectors.ForTest(0b01, 0b00, 0b11, 0b10)
			wXY := smartvectors.ForTest(0b01, 0b11, 0b01, 0b11)

			w2X := smartvectors.ForTest(0b00, 0b01, 0b10, 0b11, 0b00, 0b01, 0b10, 0b11)
			w2Y := smartvectors.ForTest(0b00, 0b00, 0b00, 0b00, 0b01, 0b01, 0b01, 0b01)
			w2XY := smartvectors.ForTest(0b00, 0b01, 0b10, 0b11, 0b01, 0b00, 0b11, 0b10)

			run.AssignColumn("XOR_TABLE_X", xorX)
			run.AssignColumn("XOR_TABLE_Y", xorY)
			run.AssignColumn("XOR_TABLE_XXORY", xorXY)

			run.AssignColumn("WITNESS_X", wX)
			run.AssignColumn("WITNESS_Y", wY)
			run.AssignColumn("WITNESS_XXORY", wXY)

			run.AssignColumn("W2_X", w2X)
			run.AssignColumn("W2_Y", w2Y)
			run.AssignColumn("W2_XXORY", w2XY)
		}

		comp := wizard.Compile(define, CompileLogDerivative, dummy.Compile)
		proof := wizard.Prove(comp, prover)

		// m should be
		err := wizard.Verify(comp, proof)
		require.NoError(b, err)
	}
}
