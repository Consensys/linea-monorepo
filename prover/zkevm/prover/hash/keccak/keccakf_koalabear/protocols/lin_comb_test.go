package protocols_test

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/common/vector"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/keccak/keccakf_koalabear/protocols"
	"github.com/stretchr/testify/assert"
)

func TestLinearComb(t *testing.T) {
	var (
		linComb = &protocols.LinearCombination{}
		cols    = make([]ifaces.Column, 8)
	)

	define := func(b *wizard.Builder) {

		for i := 0; i < 8; i++ {
			cols[i] = b.InsertCommit(0, ifaces.ColIDf("COL_%v", i), 16)
		}
		linComb = protocols.NewLinearCombination(b.CompiledIOP, "LinComb_Test", cols, 11)
	}

	prover := func(run *wizard.ProverRuntime) {
		// assign values to input columns
		col := [16]field.Element{}
		for i := 0; i < 8; i++ {

			vector.Fill(col[:], field.NewElement(uint64(i+1)%11))
			run.AssignColumn(cols[i].GetColID(), smartvectors.NewRegular(col[:]))
		}
		linComb.Run(run)
	}

	compiled := wizard.Compile(define, dummy.Compile)
	proof := wizard.Prove(compiled, prover)
	assert.NoErrorf(t, wizard.Verify(compiled, proof), "verifier failed")
}
