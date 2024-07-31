package keccak

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/consensys/zkevm-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/zkevm-monorepo/prover/protocol/ifaces"
	"github.com/consensys/zkevm-monorepo/prover/utils"

	"github.com/consensys/zkevm-monorepo/prover/maths/field"
	"github.com/consensys/zkevm-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
	"github.com/stretchr/testify/assert"
)

func TestHash(t *testing.T) {
	const maxNbKeccakF = 400
	var m module
	t.Log("compiling")
	compiled := wizard.Compile(func(b *wizard.Builder) {
		m.DefineCustomizedKeccak(b.CompiledIOP, maxNbKeccakF)

		lane := b.Columns.GetHandle("Lane")
		hashHi := b.Columns.GetHandle("Hash_Hi")
		b.InsertCommit(0, "ExpectedLane", lane.Size())
		b.InsertCommit(0, "ExpectedIsActive", lane.Size())
		b.InsertCommit(0, "ExpectedIsFirstLane", lane.Size())
		b.InsertCommit(0, "ExpectedHashHi", hashHi.Size())
		b.InsertCommit(0, "ExpectedHashLo", hashHi.Size())

		assertEqual := func(u, v ifaces.ColID) {
			b.GlobalConstraint(ifaces.QueryID(u+"="+"v"), ifaces.ColumnAsVariable(b.Columns.GetHandle(u)).Sub(ifaces.ColumnAsVariable(b.Columns.GetHandle(v))))
		}

		assertEqual("Lane", "ExpectedLane")
		assertEqual("IsLaneActive", "ExpectedIsActive")
		assertEqual("IsFirstLaneOfNewHash", "ExpectedIsFirstLane")
		assertEqual("Hash_Hi", "ExpectedHashHi")
		assertEqual("Hash_Lo", "ExpectedHashLo")

	}, dummy.Compile)
	t.Log("proving")

	for i, c := range getTestCases(t) {
		if i != 3 {
			continue
		}
		m.SliceProviders = c.in
		proof := wizard.Prove(compiled, func(r *wizard.ProverRuntime) {
			m.AssignCustomizedKeccak(r)

			lanes, isLaneActive, isFirstLaneOfHash, hashHi, hashLo := AssignColumns(c.in, utils.NextPowerOfTwo(maxNbKeccakF*lanesPerBlock))

			// pad hashHiLo to the right length; TODO length is nextPowerOfTwo(maxNbKeccakF); move to assign?
			hashPaddingLen := m.DataTransfer.HashOutput.MaxNumRows - len(hashHi)
			hashHi = append(hashHi, make([]field.Element, hashPaddingLen)...)
			hashLo = append(hashLo, make([]field.Element, hashPaddingLen)...)

			r.AssignColumn("ExpectedLane", toVectorUint64(lanes))
			r.AssignColumn("ExpectedIsActive", smartvectors.ForTest(isLaneActive...)) // It's "for test" in the sense that it expects ints. No problem using it in a non-test context.
			r.AssignColumn("ExpectedIsFirstLane", smartvectors.ForTest(isFirstLaneOfHash...))
			r.AssignColumn("ExpectedHashHi", smartvectors.NewRegular(hashHi))
			r.AssignColumn("ExpectedHashLo", smartvectors.NewRegular(hashLo))
		})

		t.Log("verifying")
		assert.NoError(t, wizard.Verify(compiled, proof))
	}
}

func TestPureGoAssign(t *testing.T) {
	for i, c := range getTestCases(t) {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			assert.Equal(t, len(c.in), len(c.hash))
			lanes, active, isFirstLane, hashHi, hashLo := AssignColumns(c.in, len(c.lanes))
			for j := range c.hash {
				hi := hashHi[j].Bytes()
				lo := hashLo[j].Bytes()
				var buf [32]byte
				assert.True(t, bytes.Equal(hi[:16], buf[:16]))
				assert.True(t, bytes.Equal(lo[:16], buf[:16]))

				copy(buf[:], hi[16:])
				copy(buf[16:], lo[16:])

				assert.Equal(t, hex.EncodeToString(c.hash[j][:]), hex.EncodeToString(buf[:]))
			}

			for j := range c.lanes {
				var b [8]byte
				binary.LittleEndian.PutUint64(b[:], lanes[j])
				assert.Equal(t, hex.EncodeToString(c.lanes[j][:]), hex.EncodeToString(b[:]))
				assert.Equal(t, c.isFirstLaneOfHash[j], isFirstLane[j])
				assert.Equal(t, c.isLaneActive[j], active[j])
			}
		})
	}
}

func toVectorUint64(in []uint64) smartvectors.SmartVector {
	e := make([]field.Element, len(in))
	for i := range in {
		e[i].SetUint64(in[i])
	}
	return smartvectors.NewRegular(e)
}
