//go:build cuda

package symbolic_test

import (
	"testing"

	"github.com/consensys/gnark-crypto/field/koalabear"
	fext "github.com/consensys/gnark-crypto/field/koalabear/extensions"
	"github.com/stretchr/testify/require"

	"github.com/consensys/linea-monorepo/prover/gpu"
	"github.com/consensys/linea-monorepo/prover/gpu/symbolic"
	"github.com/consensys/linea-monorepo/prover/gpu/vortex"
)

// ─── Compiler unit tests (pure Go) ──────────────────────────────────────────

func TestCompileGPU_Simple(t *testing.T) {
	// Expression: a + 2*b  (LinComb with 2 children)
	//
	// DAG:
	//   node 0: Input(a)
	//   node 1: Input(b)
	//   node 2: LinComb([0, 1], [1, 2])
	nodes := []symbolic.NodeOp{
		{Kind: symbolic.OpInput},
		{Kind: symbolic.OpInput},
		{Kind: symbolic.OpLinComb, Children: []int{0, 1}, Coeffs: []int{1, 2}},
	}
	pgm := symbolic.CompileGPU(nodes)

	require.Equal(t, 2, pgm.NumInputs)
	require.Equal(t, 0, len(pgm.Constants))
	require.True(t, pgm.NumSlots >= 2 && pgm.NumSlots <= 3)
	require.True(t, len(pgm.Bytecode) > 0)
}

func TestCompileGPU_Constant(t *testing.T) {
	// Expression: 42 (constant)
	var c koalabear.Element
	c.SetUint64(42)
	nodes := []symbolic.NodeOp{
		{Kind: symbolic.OpConst, ConstVal: [4]uint32{uint32(c[0]), 0, 0, 0}},
	}
	pgm := symbolic.CompileGPU(nodes)

	require.Equal(t, 0, pgm.NumInputs)
	require.Equal(t, 4, len(pgm.Constants))
	require.Equal(t, 1, pgm.NumSlots)
}

func TestCompileGPU_Product(t *testing.T) {
	// Expression: a * b^2
	nodes := []symbolic.NodeOp{
		{Kind: symbolic.OpInput},
		{Kind: symbolic.OpInput},
		{Kind: symbolic.OpProduct, Children: []int{0, 1}, Coeffs: []int{1, 2}},
	}
	pgm := symbolic.CompileGPU(nodes)
	require.Equal(t, 2, pgm.NumInputs)
}

func TestCompileGPU_PolyEval(t *testing.T) {
	// P(x) = c₀ + c₁·x where x=const(2), c₀=input(a), c₁=input(b)
	var two koalabear.Element
	two.SetUint64(2)
	nodes := []symbolic.NodeOp{
		{Kind: symbolic.OpConst, ConstVal: [4]uint32{uint32(two[0]), 0, 0, 0}}, // x=2
		{Kind: symbolic.OpInput}, // c₀ = a
		{Kind: symbolic.OpInput}, // c₁ = b
		{Kind: symbolic.OpPolyEval, Children: []int{0, 1, 2}},
	}
	pgm := symbolic.CompileGPU(nodes)
	require.Equal(t, 2, pgm.NumInputs)
	require.Equal(t, 4, len(pgm.Constants)) // one E4 constant
}

// ─── GPU evaluation tests ────────────────────────────────────────────────────

func TestGPUSymEval_LinComb(t *testing.T) {
	// f(a, b) = a + 2·b,  evaluate at a=3, b=5 → expect 13
	dev, err := gpu.New()
	require.NoError(t, err)
	defer dev.Close()

	nodes := []symbolic.NodeOp{
		{Kind: symbolic.OpInput},
		{Kind: symbolic.OpInput},
		{Kind: symbolic.OpLinComb, Children: []int{0, 1}, Coeffs: []int{1, 2}},
	}
	pgm := symbolic.CompileGPU(nodes)

	gpuPgm, err := symbolic.CompileSymGPU(dev, pgm)
	require.NoError(t, err)
	defer gpuPgm.Free()

	n := 1024
	aVec, _ := vortex.NewKBVector(dev, n)
	bVec, _ := vortex.NewKBVector(dev, n)
	defer aVec.Free()
	defer bVec.Free()

	// Fill a=3, b=5 (Montgomery form)
	var three, five koalabear.Element
	three.SetUint64(3)
	five.SetUint64(5)
	aHost := make([]koalabear.Element, n)
	bHost := make([]koalabear.Element, n)
	for i := range aHost {
		aHost[i] = three
		bHost[i] = five
	}
	aVec.CopyFromHost(aHost)
	bVec.CopyFromHost(bHost)

	inputs := []symbolic.SymInput{
		symbolic.SymInputFromVec(aVec),
		symbolic.SymInputFromVec(bVec),
	}

	result := symbolic.EvalSymGPU(dev, gpuPgm, inputs, n)

	// Expected: 3 + 2*5 = 13
	var expected fext.E4
	var thirteen koalabear.Element
	thirteen.SetUint64(13)
	expected.B0.A0 = thirteen
	for i := 0; i < n; i++ {
		require.Equal(t, expected, result[i], "mismatch at i=%d", i)
	}
}

func TestGPUSymEval_Product(t *testing.T) {
	// f(a, b) = a · b², evaluate at a=3, b=5 → expect 75
	dev, err := gpu.New()
	require.NoError(t, err)
	defer dev.Close()

	nodes := []symbolic.NodeOp{
		{Kind: symbolic.OpInput},
		{Kind: symbolic.OpInput},
		{Kind: symbolic.OpProduct, Children: []int{0, 1}, Coeffs: []int{1, 2}},
	}
	pgm := symbolic.CompileGPU(nodes)

	gpuPgm, err := symbolic.CompileSymGPU(dev, pgm)
	require.NoError(t, err)
	defer gpuPgm.Free()

	n := 512
	aVec, _ := vortex.NewKBVector(dev, n)
	bVec, _ := vortex.NewKBVector(dev, n)
	defer aVec.Free()
	defer bVec.Free()

	var three, five koalabear.Element
	three.SetUint64(3)
	five.SetUint64(5)
	aHost := make([]koalabear.Element, n)
	bHost := make([]koalabear.Element, n)
	for i := range aHost {
		aHost[i] = three
		bHost[i] = five
	}
	aVec.CopyFromHost(aHost)
	bVec.CopyFromHost(bHost)

	result := symbolic.EvalSymGPU(dev, gpuPgm, inputs(aVec, bVec), n)

	// 3 * 5² = 75
	var expected fext.E4
	var seventyfive koalabear.Element
	seventyfive.SetUint64(75)
	expected.B0.A0 = seventyfive
	for i := 0; i < n; i++ {
		require.Equal(t, expected, result[i], "mismatch at i=%d", i)
	}
}

func TestGPUSymEval_ConstantExpr(t *testing.T) {
	// Expression: const(7) + const(3) = 10
	dev, err := gpu.New()
	require.NoError(t, err)
	defer dev.Close()

	var seven, three koalabear.Element
	seven.SetUint64(7)
	three.SetUint64(3)

	nodes := []symbolic.NodeOp{
		{Kind: symbolic.OpConst, ConstVal: [4]uint32{uint32(seven[0]), 0, 0, 0}},
		{Kind: symbolic.OpConst, ConstVal: [4]uint32{uint32(three[0]), 0, 0, 0}},
		{Kind: symbolic.OpLinComb, Children: []int{0, 1}, Coeffs: []int{1, 1}},
	}
	pgm := symbolic.CompileGPU(nodes)

	gpuPgm, err := symbolic.CompileSymGPU(dev, pgm)
	require.NoError(t, err)
	defer gpuPgm.Free()

	n := 256
	result := symbolic.EvalSymGPU(dev, gpuPgm, nil, n)

	var expected fext.E4
	var ten koalabear.Element
	ten.SetUint64(10)
	expected.B0.A0 = ten
	for i := 0; i < n; i++ {
		require.Equal(t, expected, result[i])
	}
}

func TestGPUSymEval_PolyEval(t *testing.T) {
	// P(x) = c₀ + c₁·x, with x=2, c₀=3, c₁=5 → P(2) = 3 + 5*2 = 13
	dev, err := gpu.New()
	require.NoError(t, err)
	defer dev.Close()

	var two, three_, five koalabear.Element
	two.SetUint64(2)
	three_.SetUint64(3)
	five.SetUint64(5)

	nodes := []symbolic.NodeOp{
		{Kind: symbolic.OpConst, ConstVal: [4]uint32{uint32(two[0]), 0, 0, 0}}, // x=2
		{Kind: symbolic.OpConst, ConstVal: [4]uint32{uint32(three_[0]), 0, 0, 0}}, // c₀=3
		{Kind: symbolic.OpConst, ConstVal: [4]uint32{uint32(five[0]), 0, 0, 0}},   // c₁=5
		{Kind: symbolic.OpPolyEval, Children: []int{0, 1, 2}}, // P(x) = c₀ + c₁·x
	}
	pgm := symbolic.CompileGPU(nodes)

	gpuPgm, err := symbolic.CompileSymGPU(dev, pgm)
	require.NoError(t, err)
	defer gpuPgm.Free()

	n := 256
	result := symbolic.EvalSymGPU(dev, gpuPgm, nil, n)

	var expected fext.E4
	var thirteen koalabear.Element
	thirteen.SetUint64(13)
	expected.B0.A0 = thirteen
	for i := 0; i < n; i++ {
		require.Equal(t, expected, result[i], "mismatch at i=%d", i)
	}
}

func TestGPUSymEval_RotatedInput(t *testing.T) {
	// f(a) = a + rot(a, 1),  where rot shifts by +1 cyclically
	dev, err := gpu.New()
	require.NoError(t, err)
	defer dev.Close()

	nodes := []symbolic.NodeOp{
		{Kind: symbolic.OpInput},
		{Kind: symbolic.OpInput}, // will bind to rotated version
		{Kind: symbolic.OpLinComb, Children: []int{0, 1}, Coeffs: []int{1, 1}},
	}
	pgm := symbolic.CompileGPU(nodes)

	gpuPgm, err := symbolic.CompileSymGPU(dev, pgm)
	require.NoError(t, err)
	defer gpuPgm.Free()

	n := 8
	aVec, _ := vortex.NewKBVector(dev, n)
	defer aVec.Free()

	// a = [0, 1, 2, 3, 4, 5, 6, 7]  (in Montgomery)
	aHost := make([]koalabear.Element, n)
	for i := 0; i < n; i++ {
		aHost[i].SetUint64(uint64(i))
	}
	aVec.CopyFromHost(aHost)

	// inputs[0] = a (regular), inputs[1] = rot(a, 1)
	syminputs := []symbolic.SymInput{
		symbolic.SymInputFromVec(aVec),
		symbolic.SymInputFromRotatedVec(aVec, 1),
	}

	result := symbolic.EvalSymGPU(dev, gpuPgm, syminputs, n)

	// expected[i] = a[i] + a[(i+1)%8]
	for i := 0; i < n; i++ {
		var expected koalabear.Element
		expected.SetUint64(uint64(i + (i+1)%n))
		require.Equal(t, expected, result[i].B0.A0, "mismatch at i=%d", i)
	}
}

func TestGPUSymEval_NegCoeff(t *testing.T) {
	// f(a, b) = a - b  (LinComb with coeffs [1, -1])
	dev, err := gpu.New()
	require.NoError(t, err)
	defer dev.Close()

	nodes := []symbolic.NodeOp{
		{Kind: symbolic.OpInput},
		{Kind: symbolic.OpInput},
		{Kind: symbolic.OpLinComb, Children: []int{0, 1}, Coeffs: []int{1, -1}},
	}
	pgm := symbolic.CompileGPU(nodes)

	gpuPgm, err := symbolic.CompileSymGPU(dev, pgm)
	require.NoError(t, err)
	defer gpuPgm.Free()

	n := 256
	aVec, _ := vortex.NewKBVector(dev, n)
	bVec, _ := vortex.NewKBVector(dev, n)
	defer aVec.Free()
	defer bVec.Free()

	var seven, three koalabear.Element
	seven.SetUint64(7)
	three.SetUint64(3)
	aHost := make([]koalabear.Element, n)
	bHost := make([]koalabear.Element, n)
	for i := range aHost {
		aHost[i] = seven
		bHost[i] = three
	}
	aVec.CopyFromHost(aHost)
	bVec.CopyFromHost(bHost)

	result := symbolic.EvalSymGPU(dev, gpuPgm, inputs(aVec, bVec), n)

	// 7 - 3 = 4
	var expected fext.E4
	var four koalabear.Element
	four.SetUint64(4)
	expected.B0.A0 = four
	for i := 0; i < n; i++ {
		require.Equal(t, expected, result[i])
	}
}

// inputs is a helper to construct a SymInput slice from KBVectors.
func inputs(vecs ...*vortex.KBVector) []symbolic.SymInput {
	out := make([]symbolic.SymInput, len(vecs))
	for i, v := range vecs {
		out[i] = symbolic.SymInputFromVec(v)
	}
	return out
}
