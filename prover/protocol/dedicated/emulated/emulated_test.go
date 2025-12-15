package emulated

import (
	"crypto/rand"
	"crypto/sha3"
	"errors"
	"math/big"
	"testing"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/stretchr/testify/require"
)

func TestEmulatedMultiplication(t *testing.T) {
	const nbEntries = (1 << 2) + 1 // test non power power of two as well
	const nbBits = 384
	const round_nr = 0
	const nbBitsPerLimb = 128
	const nbLimbs = (nbBits + nbBitsPerLimb - 1) / nbBitsPerLimb
	var pa, pa2 *EmulatedMultiplicationModule
	define := func(b *wizard.Builder) {
		P := NewLimbs(b.CompiledIOP, round_nr, "P", nbLimbs, nbEntries)
		A := NewLimbs(b.CompiledIOP, round_nr, "A", nbLimbs, nbEntries)
		B := NewLimbs(b.CompiledIOP, round_nr, "B", nbLimbs, nbEntries)
		expected := NewLimbs(b.CompiledIOP, 0, "EXPECTED", nbLimbs, nbEntries)
		pa = EmulatedMultiplication(b.CompiledIOP, "TEST", A, B, P, nbBitsPerLimb)
		for i := range pa.Result.Columns {
			b.CompiledIOP.InsertGlobal(
				round_nr, ifaces.QueryID(ifaces.QueryIDf("EMULATED_RESULT_CORRECTNESS_%d", i)),
				symbolic.Sub(pa.Result.Columns[i], expected.Columns[i]),
			)
		}
		pa2 = EmulatedMultiplication(b.CompiledIOP, "TEST2", A, B, P, nbBitsPerLimb)
		for i := range pa2.Result.Columns {
			b.CompiledIOP.InsertGlobal(
				round_nr, ifaces.QueryID(ifaces.QueryIDf("EMULATED_RESULT2_CORRECTNESS_%d", i)),
				symbolic.Sub(pa2.Result.Columns[i], expected.Columns[i]),
			)
		}
	}

	assignmentA := make([]*big.Int, nbEntries)
	assignmentB := make([]*big.Int, nbEntries)
	assignmentP := make([]*big.Int, nbEntries)
	assignmentExpected := make([]*big.Int, nbEntries)
	bound := new(big.Int).Lsh(big.NewInt(1), nbBits)
	var err error
	reader := sha3.NewSHAKE256()
	for i := range nbEntries {
		assignmentP[i], err = rand.Int(reader, bound)
		require.NoError(t, err)
		assignmentA[i], err = rand.Int(reader, assignmentP[i])
		require.NoError(t, err)
		assignmentB[i], err = rand.Int(reader, assignmentP[i])
		require.NoError(t, err)
		assignmentExpected[i] = new(big.Int).Mul(assignmentA[i], assignmentB[i])
		assignmentExpected[i].Mod(assignmentExpected[i], assignmentP[i])
	}

	prover := func(run *wizard.ProverRuntime) {
		assignEmulated(run, "A", assignmentA, nbBitsPerLimb, nbLimbs)
		assignEmulated(run, "B", assignmentB, nbBitsPerLimb, nbLimbs)
		assignEmulated(run, "P", assignmentP, nbBitsPerLimb, nbLimbs)
		assignEmulated(run, "EXPECTED", assignmentExpected, nbBitsPerLimb, nbLimbs)
	}

	comp := wizard.Compile(define, dummy.Compile)
	proof := wizard.Prove(comp, prover)
	err = wizard.Verify(comp, proof)
	require.NoError(t, err)
}

func TestEmulatedEvaluation(t *testing.T) {
	const nbEntries = (1 << 6) + 1 // to ensure non power of two sizes are handled
	const nbBits = 384
	const round_nr = 0
	const nbBitsPerLimb = 128
	const nbLimbs = (nbBits + nbBitsPerLimb - 1) / nbBitsPerLimb
	var pa *EmulatedEvaluationModule
	define := func(b *wizard.Builder) {
		P := NewLimbs(b.CompiledIOP, round_nr, "P", nbLimbs, nbEntries)
		T0 := NewLimbs(b.CompiledIOP, round_nr, "T0", nbLimbs, nbEntries)
		T1 := NewLimbs(b.CompiledIOP, round_nr, "T1", nbLimbs, nbEntries)
		T2 := NewLimbs(b.CompiledIOP, round_nr, "T2", nbLimbs, nbEntries)
		T3 := NewLimbs(b.CompiledIOP, round_nr, "T3", nbLimbs, nbEntries)
		pa = EmulatedEvaluation(b.CompiledIOP, "TEST", nbBitsPerLimb, P, [][]Limbs{
			{T0, T1}, {T0, T1, T2}, {T3}, // T0*T1 + T0*T1*T2 + T3 == 0
		})
	}
	_ = pa

	assignmentT0 := make([]*big.Int, nbEntries)
	assignmentT1 := make([]*big.Int, nbEntries)
	assignmentT2 := make([]*big.Int, nbEntries)
	assignmentT3 := make([]*big.Int, nbEntries)
	assignmentP := make([]*big.Int, nbEntries)
	bound := new(big.Int).Lsh(big.NewInt(1), nbBits)
	var err error
	reader := sha3.NewSHAKE256()
	tmp := new(big.Int)
	for i := range nbEntries {
		assignmentP[i], err = rand.Int(reader, bound)
		require.NoError(t, err)
		assignmentT0[i], err = rand.Int(reader, assignmentP[i])
		require.NoError(t, err)
		assignmentT1[i], err = rand.Int(reader, assignmentP[i])
		require.NoError(t, err)
		assignmentT2[i], err = rand.Int(reader, assignmentP[i])
		require.NoError(t, err)
		assignmentT3[i], err = rand.Int(reader, assignmentP[i])
		require.NoError(t, err)
		// set T3 = - (T0*T1 + T0*T1*T2) mod P
		tmp.Mul(assignmentT0[i], assignmentT1[i])
		tmp.Mod(tmp, assignmentP[i])
		assignmentT3[i] = new(big.Int).Set(tmp)
		tmp.Mul(tmp, assignmentT2[i])
		tmp.Mod(tmp, assignmentP[i])
		assignmentT3[i].Add(assignmentT3[i], tmp)
		assignmentT3[i].Mod(assignmentT3[i], assignmentP[i])
		assignmentT3[i].Sub(assignmentP[i], assignmentT3[i])
		assignmentT3[i].Mod(assignmentT3[i], assignmentP[i])
	}

	prover := func(run *wizard.ProverRuntime) {
		assignEmulated(run, "P", assignmentP, nbBitsPerLimb, nbLimbs)
		assignEmulated(run, "T0", assignmentT0, nbBitsPerLimb, nbLimbs)
		assignEmulated(run, "T1", assignmentT1, nbBitsPerLimb, nbLimbs)
		assignEmulated(run, "T2", assignmentT2, nbBitsPerLimb, nbLimbs)
		assignEmulated(run, "T3", assignmentT3, nbBitsPerLimb, nbLimbs)
	}

	comp := wizard.Compile(define, dummy.Compile)
	proof := wizard.Prove(comp, prover)
	err = wizard.Verify(comp, proof)
	require.NoError(t, err)
}

type assignable interface {
	field.Element | *big.Int | uint64 | uint32 | string
}

func assignEmulated[E assignable, S []E](run *wizard.ProverRuntime, name string, limbs S, nbBitsPerLimb int, nbLimbs int) error {
	vlimbs := make([][]field.Element, nbLimbs)
	for i := range nbLimbs {
		vlimbs[i] = make([]field.Element, len(limbs))
	}
	vBi := new(big.Int)
	mask := new(big.Int).Lsh(big.NewInt(1), uint(nbBitsPerLimb))
	mask.Sub(mask, big.NewInt(1))
	tmp := new(big.Int)
	for i := range limbs {
		switch val := any(limbs[i]).(type) {
		case field.Element:
			val.BigInt(vBi)
		case *big.Int:
			vBi.Set(val)
		case uint64:
			vBi.SetUint64(val)
		case uint32:
			vBi.SetUint64(uint64(val))
		case string:
			_, ok := vBi.SetString(val, 0)
			if !ok {
				return errors.New("failed to parse string input")
			}
		default:
			panic("unsupported type")
		}
		for j := range nbLimbs {
			tmp.And(vBi, mask)
			vlimbs[j][i].SetBigInt(tmp)
			vBi.Rsh(vBi, uint(nbBitsPerLimb))
		}
	}
	for j := range nbLimbs {
		sv := smartvectors.RightPadded(vlimbs[j], field.NewElement(0), utils.NextPowerOfTwo(len(vlimbs[j])))
		run.AssignColumn(ifaces.ColIDf("%s_LIMB_%d", name, j), sv)
	}
	return nil
}
