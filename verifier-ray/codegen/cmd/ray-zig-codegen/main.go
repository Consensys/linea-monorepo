// Command ray-zig-codegen generates src/generated/stub.zig by running
// GenerateZig on a wiop.System and writing the result to stdout.
//
// The Makefile target generate-stub pipes this output into stub.zig atomically:
//
//	cd codegen && go run ./cmd/ray-zig-codegen/ > ../src/generated/stub.zig.tmp \
//	    && mv ../src/generated/stub.zig.tmp ../src/generated/stub.zig
//
// # Current state (placeholder)
//
// buildSystem constructs a minimal toy circuit so the full codegen →
// compile → test pipeline can be exercised end-to-end before the real
// prover-ray circuit is available here.
//
// # When the real circuit is wired in
//
// Replace buildSystem with a function that constructs and compiles the real
// prover-ray wiop.System — the same System object that the prover uses at
// runtime. Concretely:
//
//  1. Call the prover-ray top-level builder (e.g. zkcdriver.BuildSystem or
//     equivalent) to obtain a *wiop.System.
//  2. Run all compiler passes on it (global.Compile, wiop.Materialize, …).
//  3. Return that fully-compiled System to main.
//
// GenerateZig and the Zig verifier require no changes: the generated stub will
// automatically reflect the real round structure, column layout, and
// global-constraint checks of the actual protocol.
package main

import (
	"fmt"
	"os"

	"github.com/consensys/linea-monorepo/prover-ray/maths/koalabear/field"
	"github.com/consensys/linea-monorepo/prover-ray/wiop"
	"github.com/consensys/linea-monorepo/prover-ray/wiop/compilers/global"
	zigcodegen "github.com/consensys/linea-monorepo/verifier-ray/codegen"
)

func main() {
	sys := buildSystem()
	if err := zigcodegen.GenerateZig(sys, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "ray-zig-codegen: %v\n", err)
		os.Exit(1)
	}
}

// buildSystem constructs the wiop.System whose verifier will be generated.
//
// TODO: replace this placeholder with the real prover-ray circuit.
//
// To wire in the real circuit:
//  1. Replace the toy module/vanishing below with a call to the prover-ray
//     top-level builder that produces the actual protocol's *wiop.System.
//  2. Keep the global.Compile and wiop.Materialize calls — they must run on
//     any System before it is passed to GenerateZig.
//
// Everything below this function (GenerateZig, the Zig runtime, the Zig
// verifier) requires no changes when the real circuit is substituted.
//
// Placeholder circuit: one module of size 4, one vanishing A·B = 0.
func buildSystem() *wiop.System {
	sys := wiop.NewSystemf("verifier-ray")
	sys.NewRound()

	m := sys.NewSizedModule(sys.Context.Childf("main"), 4, wiop.PaddingDirectionRight)

	a := m.NewPrecomputedColumn(sys.Context.Childf("A"), wiop.VisibilityOracle, makeVec(0, 1, 0, 0))
	b := m.NewPrecomputedColumn(sys.Context.Childf("B"), wiop.VisibilityOracle, makeVec(1, 0, 0, 0))
	m.NewVanishing(sys.Context.Childf("AB"), wiop.Mul(a.View(), b.View()))

	global.Compile(sys)
	wiop.Materialize(sys)
	return sys
}

func makeVec(vals ...uint64) *wiop.ConcreteVector {
	elems := make([]field.Element, len(vals))
	for i, v := range vals {
		elems[i].SetUint64(v)
	}
	return &wiop.ConcreteVector{Plain: field.VecFromBase(elems)}
}
