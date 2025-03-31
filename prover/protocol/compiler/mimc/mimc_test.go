package mimc_test

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/crypto/mimc"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	mimcComp "github.com/consensys/linea-monorepo/prover/protocol/compiler/mimc"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/stretchr/testify/require"
)

func TestMiMCCompilerSingleQuery(t *testing.T) {

	size := 16

	var block, old, new ifaces.Column

	define := func(b *wizard.Builder) {
		block = b.RegisterCommit("BLOCK", size)
		old = b.RegisterCommit("OLD", size)
		new = b.RegisterCommit("NEW", size)
		b.InsertMiMC(0, "MIMC", block, old, new, nil)
	}

	prove := func(run *wizard.ProverRuntime) {
		bl := make([]field.Element, size)
		ol := make([]field.Element, size)
		ne := make([]field.Element, size)

		for i := 0; i < size; i++ {
			bl[i] = field.NewElement(uint64(i))
			ol[i] = field.NewElement(uint64(i + size))
			ne[i] = mimc.BlockCompression(ol[i], bl[i])
		}

		run.AssignColumn(block.GetColID(), smartvectors.NewRegular(bl))
		run.AssignColumn(old.GetColID(), smartvectors.NewRegular(ol))
		run.AssignColumn(new.GetColID(), smartvectors.NewRegular(ne))
	}

	comp := wizard.Compile(define, mimcComp.CompileMiMC, dummy.Compile)
	proof := wizard.Prove(comp, prove)
	require.NoError(t, wizard.Verify(comp, proof))
}

func TestMiMCCompilerTwoQuery(t *testing.T) {

	size1 := 16
	size2 := 8

	var block1, old1, new1, block2, old2, new2 ifaces.Column

	define := func(b *wizard.Builder) {
		block1 = b.RegisterCommit("BLOCK1", size1)
		old1 = b.RegisterCommit("OLD1", size1)
		new1 = b.RegisterCommit("NEW1", size1)
		b.InsertMiMC(0, "MIMC1", block1, old1, new1, nil)

		block2 = b.RegisterCommit("BLOCK2", size2)
		old2 = b.RegisterCommit("OLD2", size2)
		new2 = b.RegisterCommit("NEW2", size2)
		b.InsertMiMC(0, "MIMC2", block2, old2, new2, nil)
	}

	prove := func(run *wizard.ProverRuntime) {
		bl1 := make([]field.Element, size1)
		ol1 := make([]field.Element, size1)
		ne1 := make([]field.Element, size1)

		for i := 0; i < size1; i++ {
			bl1[i] = field.NewElement(uint64(i))
			ol1[i] = field.NewElement(uint64(i + size1))
			ne1[i] = mimc.BlockCompression(ol1[i], bl1[i])
		}

		run.AssignColumn(block1.GetColID(), smartvectors.NewRegular(bl1))
		run.AssignColumn(old1.GetColID(), smartvectors.NewRegular(ol1))
		run.AssignColumn(new1.GetColID(), smartvectors.NewRegular(ne1))

		bl2 := make([]field.Element, size2)
		ol2 := make([]field.Element, size2)
		ne2 := make([]field.Element, size2)

		for i := 0; i < size2; i++ {
			bl2[i] = field.NewElement(uint64(i))
			ol2[i] = field.NewElement(uint64(i + size2))
			ne2[i] = mimc.BlockCompression(ol2[i], bl2[i])
		}

		run.AssignColumn(block2.GetColID(), smartvectors.NewRegular(bl2))
		run.AssignColumn(old2.GetColID(), smartvectors.NewRegular(ol2))
		run.AssignColumn(new2.GetColID(), smartvectors.NewRegular(ne2))
	}

	comp := wizard.Compile(define, mimcComp.CompileMiMC, dummy.Compile)
	proof := wizard.Prove(comp, prove)
	require.NoError(t, wizard.Verify(comp, proof))
}
