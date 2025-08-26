package poseidon2

import (
	"testing"

	cryptoposeidon2 "github.com/consensys/linea-monorepo/prover/crypto/poseidon2"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/stretchr/testify/require"
)

func TestPermutation(t *testing.T) {
	var old, block [8]field.Element
	old[0] = field.NewElement(uint64(703724752))
	old[1] = field.NewElement(uint64(280040542))
	old[2] = field.NewElement(uint64(1514240686))
	old[3] = field.NewElement(uint64(986917665))
	old[4] = field.NewElement(uint64(1972211392))
	old[5] = field.NewElement(uint64(832875602))
	old[6] = field.NewElement(uint64(2095324332))
	old[7] = field.NewElement(uint64(36857942))

	block[0] = field.NewElement(uint64(760417386))
	block[1] = field.NewElement(uint64(1333026101))
	block[2] = field.NewElement(uint64(835587083))
	block[3] = field.NewElement(uint64(1017667263))
	block[4] = field.NewElement(uint64(669624325))
	block[5] = field.NewElement(uint64(1903375813))
	block[6] = field.NewElement(uint64(1853215757))
	block[7] = field.NewElement(uint64(199352308))

	state := Poseidon2BlockCompression(old, block)

	newstate := cryptoposeidon2.Poseidon2BlockCompression(old, block)

	for i := 0; i < 8; i++ {
		require.Equal(t, newstate[i].String(), state[i].String())
	}

}
