package logderivativesum

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/coin"
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

	comp := wizard.Compile(define, CompileLookups, dummy.Compile)
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

	comp := wizard.Compile(define, CompileLookups, dummy.Compile)
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
		cola3 := b.RegisterCommit("S3", sizeA)
		colb := b.RegisterCommit("T", sizeB)
		b.Inclusion("LOOKUP", []ifaces.Column{colb}, []ifaces.Column{cola})
		b.Inclusion("LOOKUP2", []ifaces.Column{colb}, []ifaces.Column{cola2})
		b.Inclusion("LOOKUP3", []ifaces.Column{colb}, []ifaces.Column{cola3})
	}

	prover := func(run *wizard.ProverRuntime) {
		runtime = run
		// assign a and b
		cola := smartvectors.ForTest(1, 1, 1, 2, 3, 0, 0, 1, 1, 1, 1, 2, 3, 0, 0, 1)
		cola2 := smartvectors.ForTest(2, 2, 2, 1, 0, 3, 3, 2, 2, 2, 2, 1, 0, 3, 3, 2)
		cola3 := smartvectors.ForTest(2, 2, 2, 1, 0, 3, 3, 2, 2, 2, 2, 1, 0, 3, 3, 3)
		colb := smartvectors.ForTest(0, 1, 2, 3)
		// m expected to be
		run.AssignColumn("S", cola)
		run.AssignColumn("S2", cola2)
		run.AssignColumn("S3", cola3)
		run.AssignColumn("T", colb)
	}

	comp := wizard.Compile(define, CompileLookups, dummy.Compile)
	proof := wizard.Prove(comp, prover)

	// m should be
	expectedM := smartvectors.ForTest(8, 12, 17, 11)
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

	comp := wizard.Compile(define, CompileLookups, dummy.Compile)
	proof := wizard.Prove(comp, prover)

	// m should be
	expectedM := smartvectors.ForTest(0, 0, 0, 1, 1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0)
	t.Logf("all columns = %v", runtime.Columns.ListAllKeys())
	actualM := runtime.GetColumn("TABLE_XOR_TABLE_X,XOR_TABLE_XXORY,XOR_TABLE_Y_0_LOGDERIVATIVE_M")

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

	comp := wizard.Compile(define, CompileLookups, dummy.Compile)
	proof := wizard.Prove(comp, prover)

	// m should be
	expectedM := smartvectors.ForTest(1, 1, 1, 2, 2, 1, 1, 1, 0, 1, 0, 0, 0, 0, 1, 0)
	t.Logf("all column names = %v", runtime.Columns.ListAllKeys())
	actualM := runtime.GetColumn("TABLE_XOR_TABLE_X,XOR_TABLE_XXORY,XOR_TABLE_Y_0_LOGDERIVATIVE_M")

	assert.Equal(t, expectedM.Pretty(), actualM.Pretty(), "m does not match the expected value")

	err := wizard.Verify(comp, proof)
	require.NoError(t, err)
}

func TestLogDerivativeLookupRandomLinComb(t *testing.T) {

	var sizeA, sizeB int = 16, 8
	var col1, col2 ifaces.Column
	define := func(b *wizard.Builder) {
		col1 = b.RegisterPrecomputed("P1", smartvectors.ForTest(1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1))
		col2 = b.RegisterPrecomputed("P2", smartvectors.ForTest(12, 6, 8, 0, 3, 12, 13, 23, 17, 9, 8, 7, 6, 5, 4, 3))
		colI := b.RegisterPrecomputed("I", smartvectors.ForTest(0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15))

		_ = b.RegisterRandomCoin("COIN", coin.Field)

		uCol := b.InsertProof(1, "LC", sizeA)

		_ = b.RegisterRandomCoin("COIN1", coin.Field)

		colQ := b.RegisterCommit("Q", sizeB)
		uChosen := b.RegisterCommit("UChosen", sizeB)

		// multi-col query
		b.Inclusion("LOOKUP", []ifaces.Column{colI, uCol}, []ifaces.Column{colQ, uChosen})
	}

	prover := func(run *wizard.ProverRuntime) {
		// assign a and b

		coin := run.GetRandomCoinField("COIN")

		a := col1.GetColAssignment(run)
		b := col2.GetColAssignment(run)
		lc := smartvectors.PolyEval([]smartvectors.SmartVector{a, b}, coin)

		run.AssignColumn("LC", lc)

		run.GetRandomCoinField("COIN1")

		colQ := smartvectors.ForTest(0, 1, 2, 3, 4, 5, 6, 7)
		run.AssignColumn("Q", colQ)

		colQFr := colQ.IntoRegVecSaveAlloc()
		var t []field.Element
		for _, q := range colQFr {
			t = append(t, lc.Get(int(q.Uint64())))
		}
		run.AssignColumn("UChosen", smartvectors.NewRegular(t))
	}

	comp := wizard.Compile(define, CompileLookups, dummy.Compile)
	proof := wizard.Prove(comp, prover)
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

		comp := wizard.Compile(define, CompileLookups, dummy.Compile)
		proof := wizard.Prove(comp, prover)

		// m should be
		err := wizard.Verify(comp, proof)
		require.NoError(b, err)
	}
}

func TestLogDerivativeLookupMultiXorMixed(t *testing.T) {

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
		xorY := smartvectors.ForTestFromPairs(0b00, 0, 0b00, 0, 0b00, 0, 0b00, 0, 0b01, 0, 0b01, 0, 0b01, 0, 0b01, 0, 0b10, 0, 0b10, 0, 0b10, 0, 0b10, 0, 0b11, 0, 0b11, 0, 0b11, 0, 0b11, 0)
		xorXY := smartvectors.ForTest(0b00, 0b01, 0b10, 0b11, 0b01, 0b00, 0b11, 0b10, 0b10, 0b11, 0b00, 0b01, 0b11, 0b10, 0b01, 0b00)

		wX := smartvectors.ForTest(0b00, 0b11, 0b10, 0b01)
		wY := smartvectors.ForTestFromPairs(0b01, 0, 0b00, 0, 0b11, 0, 0b10, 0)
		wXY := smartvectors.ForTest(0b01, 0b11, 0b01, 0b11)

		w2X := smartvectors.ForTest(0b00, 0b01, 0b10, 0b11, 0b00, 0b01, 0b10, 0b11)
		w2Y := smartvectors.ForTestFromPairs(0b00, 0, 0b00, 0, 0b00, 0, 0b00, 0, 0b01, 0, 0b01, 0, 0b01, 0, 0b01, 0)
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

	comp := wizard.Compile(define, CompileLookups, dummy.Compile)
	proof := wizard.Prove(comp, prover)

	// m should be
	expectedM := smartvectors.ForTest(1, 1, 1, 2, 2, 1, 1, 1, 0, 1, 0, 0, 0, 0, 1, 0)
	t.Logf("all column names = %v", runtime.Columns.ListAllKeys())
	actualM := runtime.GetColumn("TABLE_XOR_TABLE_X,XOR_TABLE_XXORY,XOR_TABLE_Y_0_LOGDERIVATIVE_M")

	assert.Equal(t, expectedM.Pretty(), actualM.Pretty(), "m does not match the expected value")

	err := wizard.Verify(comp, proof)
	require.NoError(t, err)
}

func TestLogDerivativeLookupMultiMixed(t *testing.T) {

	var sizeTable, sizeCheckeds, sizeCheckedLarger int = 8, 4, 4
	var runtime *wizard.ProverRuntime

	// The test uses a lookup over a xor table

	define := func(b *wizard.Builder) {

		tableA := b.RegisterCommit("TABLE_A", sizeTable)
		tableB := b.RegisterCommit("TABLE_B", sizeTable)
		tableC := b.RegisterCommit("TABLE_C", sizeTable)

		includedA := b.RegisterCommit("INCLUDED_A", sizeCheckeds)
		includedB := b.RegisterCommit("INCLUDED_B", sizeCheckeds)
		includedC := b.RegisterCommit("INCLUDED_C", sizeCheckeds)

		includedA2 := b.RegisterCommit("INCLUDED_A2", sizeCheckedLarger)
		includedB2 := b.RegisterCommit("INCLUDED_B2", sizeCheckedLarger)
		includedC2 := b.RegisterCommit("INCLUDED_C2", sizeCheckedLarger)

		b.Inclusion("LOOKUP", []ifaces.Column{tableA, tableB, tableC}, []ifaces.Column{includedA, includedB, includedC})
		b.Inclusion("LOOKUP2", []ifaces.Column{tableA, tableB, tableC}, []ifaces.Column{includedA2, includedB2, includedC2})
	}

	prover := func(run *wizard.ProverRuntime) {
		runtime = run
		// assign a and b
		// tableA contains base field integers
		tableA := smartvectors.ForTest(2816, 2817, 2816, 3001, 3002, 5008, 2816, 9000)
		// table B contains extension field elements computed as (first two digits, last two digits)
		tableB := smartvectors.ForTestFromPairs(28, 16, 28, 17, 28, 16, 30, 01, 30, 02, 50, 8, 28, 16, 90, 0)
		// table B contains base field elements (first three digits, last two digits)
		tableC := smartvectors.ForTest(281, 281, 281, 300, 300, 500, 281, 900)

		includedA := smartvectors.ForTest(2816, 9000, 3001, 9000)
		includedB := smartvectors.ForTestFromPairs(28, 16, 90, 00, 30, 01, 90, 0)
		includedC := smartvectors.ForTest(281, 900, 300, 900)

		includedA2 := smartvectors.ForTest(3002, 2817, 3001, 9000)
		includedB2 := smartvectors.ForTestFromPairs(30, 02, 28, 17, 30, 01, 90, 0)
		includedC2 := smartvectors.ForTest(300, 281, 300, 900)

		run.AssignColumn("TABLE_A", tableA)
		run.AssignColumn("TABLE_B", tableB)
		run.AssignColumn("TABLE_C", tableC)

		run.AssignColumn("INCLUDED_A", includedA)
		run.AssignColumn("INCLUDED_B", includedB)
		run.AssignColumn("INCLUDED_C", includedC)

		run.AssignColumn("INCLUDED_A2", includedA2)
		run.AssignColumn("INCLUDED_B2", includedB2)
		run.AssignColumn("INCLUDED_C2", includedC2)
	}

	comp := wizard.Compile(define, CompileLookups, dummy.Compile)
	proof := wizard.Prove(comp, prover)

	// m should be
	expectedM := smartvectors.ForTest(0, 1, 0, 2, 1, 0, 1, 3)
	t.Logf("all column names = %v", runtime.Columns.ListAllKeys())
	actualM := runtime.GetColumn("TABLE_TABLE_A,TABLE_B,TABLE_C_0_LOGDERIVATIVE_M")

	assert.Equal(t, expectedM.Pretty(), actualM.Pretty(), "m does not match the expected value")

	err := wizard.Verify(comp, proof)
	require.NoError(t, err)
}

func TestLogDerivativeLookupOnlyExt(t *testing.T) {

	var sizeTable, sizeCheckeds, sizeCheckedLarger int = 8, 4, 4
	var runtime *wizard.ProverRuntime

	// The test uses a lookup over a xor table

	define := func(b *wizard.Builder) {

		tableA := b.RegisterCommit("TABLE_A", sizeTable)
		tableB := b.RegisterCommit("TABLE_B", sizeTable)
		tableC := b.RegisterCommit("TABLE_C", sizeTable)

		includedA := b.RegisterCommit("INCLUDED_A", sizeCheckeds)
		includedB := b.RegisterCommit("INCLUDED_B", sizeCheckeds)
		includedC := b.RegisterCommit("INCLUDED_C", sizeCheckeds)

		includedA2 := b.RegisterCommit("INCLUDED_A2", sizeCheckedLarger)
		includedB2 := b.RegisterCommit("INCLUDED_B2", sizeCheckedLarger)
		includedC2 := b.RegisterCommit("INCLUDED_C2", sizeCheckedLarger)

		b.Inclusion("LOOKUP", []ifaces.Column{tableA, tableB, tableC}, []ifaces.Column{includedA, includedB, includedC})
		b.Inclusion("LOOKUP2", []ifaces.Column{tableA, tableB, tableC}, []ifaces.Column{includedA2, includedB2, includedC2})
	}

	prover := func(run *wizard.ProverRuntime) {
		runtime = run
		// assign a and b
		// tableA contains base field integers
		tableA := smartvectors.ForTestExt(2816, 2817, 2816, 3001, 3002, 5008, 2816, 9000)
		// table B contains extension field elements computed as (first two digits, last two digits)
		tableB := smartvectors.ForTestFromPairs(28, 16, 28, 17, 28, 16, 30, 01, 30, 02, 50, 8, 28, 16, 90, 0)
		// table B contains base field elements (first three digits, last two digits)
		tableC := smartvectors.ForTestExt(281, 281, 281, 300, 300, 500, 281, 900)

		includedA := smartvectors.ForTestExt(2816, 9000, 3001, 9000)
		includedB := smartvectors.ForTestFromPairs(28, 16, 90, 00, 30, 01, 90, 0)
		includedC := smartvectors.ForTestExt(281, 900, 300, 900)

		includedA2 := smartvectors.ForTestExt(3002, 2817, 3001, 9000)
		includedB2 := smartvectors.ForTestFromPairs(30, 02, 28, 17, 30, 01, 90, 0)
		includedC2 := smartvectors.ForTestExt(300, 281, 300, 900)

		run.AssignColumn("TABLE_A", tableA)
		run.AssignColumn("TABLE_B", tableB)
		run.AssignColumn("TABLE_C", tableC)

		run.AssignColumn("INCLUDED_A", includedA)
		run.AssignColumn("INCLUDED_B", includedB)
		run.AssignColumn("INCLUDED_C", includedC)

		run.AssignColumn("INCLUDED_A2", includedA2)
		run.AssignColumn("INCLUDED_B2", includedB2)
		run.AssignColumn("INCLUDED_C2", includedC2)
	}

	comp := wizard.Compile(define, CompileLookups, dummy.Compile)
	proof := wizard.Prove(comp, prover)

	// m should be
	expectedM := smartvectors.ForTest(0, 1, 0, 2, 1, 0, 1, 3)
	t.Logf("all column names = %v", runtime.Columns.ListAllKeys())
	actualM := runtime.GetColumn("TABLE_TABLE_A,TABLE_B,TABLE_C_0_LOGDERIVATIVE_M")

	assert.Equal(t, expectedM.Pretty(), actualM.Pretty(), "m does not match the expected value")

	err := wizard.Verify(comp, proof)
	require.NoError(t, err)
}
