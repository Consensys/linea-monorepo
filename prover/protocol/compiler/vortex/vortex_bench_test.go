package vortex_test

import (
	"math/rand/v2"
	"testing"
	"time"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/vortex"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

// BenchmarkVortexUAlpha measures prover and verifier wall-clock time
// for a single-round KoalaBear Vortex.
//
// Parameters chosen to be representative of an intermediate self-recursion
// round: polSize=1<<16, nPols=32, RS=4, numOpenedCols=64.
//
// Run with:
//
//	go test -run=^$ -bench=BenchmarkVortexUAlpha -benchtime=3x -v ./protocol/compiler/vortex/
func BenchmarkVortexUAlpha(b *testing.B) {
	const (
		polSize      = 1 << 16
		nPols        = 32
		rsRate       = 4
		numOpenedCol = 64
	)

	var cols []ifaces.Column

	define := func(bld *wizard.Builder) {
		cols = make([]ifaces.Column, nPols)
		for i := range cols {
			cols[i] = bld.RegisterCommit(ifaces.ColIDf("P_%v", i), polSize)
		}
		bld.UnivariateEval("EVAL", cols...)
	}

	prove := func(pr *wizard.ProverRuntime) {
		rng := rand.New(rand.NewPCG(0, 0)) // #nosec G404 -- bench only
		ys := make([]fext.Element, nPols)
		x := fext.PseudoRand(rng)
		for i, col := range cols {
			p := smartvectors.PseudoRand(rng, polSize)
			ys[i] = smartvectors.EvaluateBasePolyLagrange(p, x)
			pr.AssignColumn(col.GetColID(), p)
		}
		pr.AssignUnivariateExt("EVAL", x, ys...)
	}

	opts := []vortex.VortexOp{
		vortex.WithOptionalSISHashingThreshold(1 << 20), // disable SIS for isolation
		vortex.ForceNumOpenedColumns(numOpenedCol),
	}
	compiled := wizard.Compile(define, vortex.Compile(rsRate, false, opts...))

	for i := 0; i < b.N; i++ {
		t0 := time.Now()
		proof := wizard.Prove(compiled, prove)
		proveMs := time.Since(t0).Milliseconds()

		t1 := time.Now()
		err := wizard.Verify(compiled, proof)
		verifyMs := time.Since(t1).Milliseconds()

		if err != nil {
			b.Fatalf("verify failed: %v", err)
		}

		b.ReportMetric(float64(proveMs), "prove_ms")
		b.ReportMetric(float64(verifyMs), "verify_ms")
	}
}
