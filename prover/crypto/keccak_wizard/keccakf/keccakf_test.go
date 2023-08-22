//go:build !race

package keccak

import (
	"crypto/rand"
	"encoding/binary"
	"testing"

	"github.com/consensys/accelerated-crypto-monorepo/protocol/compiler/dummy"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/wizard"
	"github.com/consensys/accelerated-crypto-monorepo/utils/parallel"
	"github.com/stretchr/testify/require"
)

func TestKeccakF(t *testing.T) {
	var ctx KeccakFModule
	// give the number of permutations in a batch
	ctx.NP = 4
	define := func(b *wizard.Builder) {
		ctx.DefineKeccakF(b.CompiledIOP, 0)
	}
	// profiling.ProfileTrace("test-example", true, true, func() {
	compiled := wizard.Compile(
		define,
		/*
			specialqueries.RangeProof,
			specialqueries.CompileFixedPermutations,
			specialqueries.CompileInclusionPermutations,
			splitter.SplitColumns(1<<7),
			arithmetics.CompileLocal,
			arithmetics.CompileGlobal,
			univariates.CompileLocalOpening,
			univariates.Naturalize,
			univariates.MultiPointToSinglePoint,
			//vortex.Compile(2, 7),

			dummy.LazyCommit,
		*/
		dummy.Compile,
	)
	var a [5][5]uint64
	/* it receives the public input that is [5][5] elements of uint64 and convert it to bit-base-first field.Element
	Namely, bits are coefficients of powers of first */
	parallel.Execute(ctx.NP, func(start, stop int) {

		for l := start; l < stop; l++ {
			for i := 0; i < 5; i++ {
				for j := 0; j < 5; j++ {
					b := make([]byte, 8)
					if _, err := rand.Reader.Read(b); err != nil {
						panic(err)
					}
					a[i][j] = binary.LittleEndian.Uint64(b)
				}
			}
			in := a

			ctx.InputPI[l] = ConvertState(in, First)
			// the original keccakf
			ctx.OutputPI[l] = ConvertState(KeccakF1600Original(in), First)

		}
	})
	proof := wizard.Prove(compiled, ctx.ProverAssign)
	err := wizard.Verify(compiled, proof)
	require.NoError(t, err)
	// })

}
