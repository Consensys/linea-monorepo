// Command ray-zig-codegen builds the placeholder wiop.System and writes a Zig
// verifier stub to stdout.
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
