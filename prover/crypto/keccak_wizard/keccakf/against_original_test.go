// this test our KeccakF1600 against the original KeccakF to make sure they have the same output/input
package keccak

import (
	"crypto/rand"
	"encoding/binary"
	"testing"

	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
)

func TestKeccak1600(_ *testing.T) {
	var a [5][5]uint64
	var ctx KeccakFModule
	ctx.NP = 1
	ctx.PartialKecccakFCtx()
	ctx.BuildColumns()
	for i := 0; i < 5; i++ {
		for j := 0; j < 5; j++ {
			b := make([]byte, 8)
			if _, err := rand.Reader.Read(b); err != nil {
				panic(err)
			}
			a[i][j] = binary.LittleEndian.Uint64(b)
		}
	}
	in := ConvertState(a, First)
	out := ctx.KeccakF1600(in, 0)
	//sanity check
	var v, z [5][5]uint64
	for i := range in {
		for j := range in[i] {

			u := fieldToByte(in[i][j])
			v[i][j] = binary.LittleEndian.Uint64(u)

			if v[i][j] != a[i][j] {
				utils.Panic("conversion from field to byte is not correct at positions %v,%v", i, j)
			}

			W := fieldToByte(out[i][j])
			z[i][j] = binary.LittleEndian.Uint64(W)

		}
	}
	want := KeccakF1600Original(v)
	if want != z {
		panic("the permutation result is not correct")
	}

}

// it convert bit-base-first Field Element to bytes
func fieldToByte(a field.Element) (s []byte) {

	r := Convertbase(a, First, 64)
	rr := Compose(r, 2)
	//sanity check
	if !rr.IsUint64() {
		panic("expected a uint64 representation")
	}
	s = make([]byte, 8)
	binary.LittleEndian.PutUint64(s, rr.Uint64())

	//sanity check
	v := binary.LittleEndian.Uint64(s)
	if v != rr.Uint64() {
		panic("conversion between byte and Uint64 is not correct")
	}

	//sanity check
	t := rr.Uint64()
	for k := 0; k < 64; k++ {
		bit := t >> k & 1
		if field.NewElement(bit) != r[k] {
			panic("conversion of bytes to uint64 is not correct")
		}
	}

	return s
}
