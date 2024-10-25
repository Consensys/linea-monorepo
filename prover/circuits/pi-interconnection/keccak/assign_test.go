package keccak

import (
	"crypto/rand"
	"encoding/binary"
	"github.com/consensys/linea-monorepo/prover/circuits/internal"
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/test"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/stretchr/testify/assert"
)

func TestAssignSingleStrict(t *testing.T) {
	testAssign(t, []int{-1}, []int{32})
}

func TestAssignSingleFlexible(t *testing.T) {

}

// maxSize = -1 means strict
func testAssign(t *testing.T, maxSizes []int, actualSizes []int) {
	assert.Equal(t, len(maxSizes), len(actualSizes))
	compiler := NewStrictHasherCompiler(len(maxSizes))
	for i, l := range actualSizes {
		if maxSizes[i] == -1 {
			compiler.WithStrictHashLengths(l)
		} else {
			compiler.WithFlexibleHashLengths(l)
		}
	}
	compiled := compiler.Compile(dummy.Compile)

	var (
		buf [32]byte
		v   uint64
		err error
	)

	assignment := testAssignCircuit{
		Ins:   make([][][32]frontend.Variable, len(actualSizes)),
		NbIns: make([]frontend.Variable, len(actualSizes)),
		Outs:  make([][32]frontend.Variable, len(actualSizes)),
	}
	circuit := testAssignCircuit{
		Ins:        [][][32]frontend.Variable{make([][32]frontend.Variable, len(actualSizes))},
		strictSize: internal.MapSlice(func(i int) bool { return i == -1 }, maxSizes...),
		NbIns:      make([]frontend.Variable, len(actualSizes)),
		Outs:       make([][32]frontend.Variable, len(actualSizes)),
	}

	hsh := compiled.GetHasher()
	for i := range actualSizes {
		hsh.Reset()
		if maxSizes[i] == -1 {
			assignment.NbIns[i], err = rand.Read(buf[:2])
			require.NoError(t, err)
		} else {
			assignment.NbIns[i] = maxSizes[i]
		}
		assignment.Ins[i] = make([][32]frontend.Variable, actualSizes[i]/32)
		circuit.Ins[i] = make([][32]frontend.Variable, len(assignment.Ins[i]))
		for j := range assignment.Ins[i] {
			binary.LittleEndian.PutUint64(buf[:], v)
			v++

			_, err = hsh.Write(buf[:])
			require.NoError(t, err)

			utils.Copy(assignment.Ins[i][j][:], buf[:])
		}

		utils.Copy(assignment.Outs[i][:], hsh.Sum(nil))
	}

	circuit.H, err = compiled.GetCircuit()
	assert.NoError(t, err)

	assignment.H, err = hsh.Assign()
	assert.NoError(t, err)

	assert.NoError(t, test.IsSolved(&circuit, &assignment, ecc.BLS12_377.ScalarField()))
}

func (c *testAssignCircuit) Define(api frontend.API) error {
	hsh := c.H.NewHasher(api)
	for i := range c.Ins {
		var nbIn frontend.Variable
		if !c.strictSize[i] {
			nbIn = c.NbIns[i]
		}
		out := hsh.Sum(nbIn, c.Ins[i]...)
		internal.AssertSliceEquals(api, c.Outs[i][:], out[:])
	}
	return hsh.Finalize()
}

type testAssignCircuit struct {
	H          StrictHasherCircuit
	Ins        [][][32]frontend.Variable
	NbIns      []frontend.Variable
	strictSize []bool
	Outs       [][32]frontend.Variable
}
