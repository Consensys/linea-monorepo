package multilinearmpts

import (
	"math/rand/v2"
	"testing"

	"github.com/consensys/linea-monorepo/prover-ray/crypto/koalabear/sumcheck"
	"github.com/consensys/linea-monorepo/prover-ray/maths/koalabear/field"
)

func newRng() *rand.Rand { return rand.New(rand.NewPCG(0xdeadbeef, 0)) }

// TestEvalGateDegree checks the declared degree.
func TestEvalGateDegree(t *testing.T) {
	g := &EvalGate{Lambdas: []field.Ext{{}, {}}}
	if d := g.Degree(); d != 2 {
		t.Fatalf("Degree: got %d want 2", d)
	}
}

// TestEvalGateImplementsInterface asserts at compile time that *EvalGate
// satisfies sumcheck.Gate.
func TestEvalGateImplementsInterface(t *testing.T) {
	var _ sumcheck.Gate = (*EvalGate)(nil)
}

// TestEvalGateEvalBatch verifies the gate formula against a hand-computed
// reference: res[j] = Σᵢ Lambdas[i] · inputs[2i][j] · inputs[2i+1][j].
func TestEvalGateEvalBatch(t *testing.T) {
	rng := newRng()
	const m = 3
	const n = 64

	lambda := field.PseudoRandExt(rng)
	gate := NewEvalGate(lambda, m)
	lambdas := gate.Lambdas

	// 2m random input tables.
	inputs := make([][]field.Ext, 2*m)
	for k := range inputs {
		inputs[k] = make([]field.Ext, n)
		for j := range inputs[k] {
			inputs[k][j] = field.PseudoRandExt(rng)
		}
	}

	res := make([]field.Ext, n)
	gate.EvalBatch(res, inputs...)

	for j := range n {
		var want field.Ext
		for i, λ := range lambdas {
			var prod, term field.Ext
			prod.Mul(&inputs[2*i][j], &inputs[2*i+1][j])
			term.Mul(&prod, &λ)
			want.Add(&want, &term)
		}
		if !res[j].Equal(&want) {
			t.Fatalf("j=%d: got %v want %v", j, res[j], want)
		}
	}
}
