package gkrverifier

import (
	"math/rand/v2"
	"testing"

	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/gnark-crypto/field/koalabear/poseidon2"
)

const compressDriver = "../../src/test/zkc/gkr/compress.zkc"

func words(es []koalabear.Element) zkcInput {
	w := make(zkcInput, len(es))
	for i := range es {
		w[i] = uint32(es[i].Uint64())
	}
	return w
}

func TestCompress(t *testing.T) {
	compressor := poseidon2.NewCompressor(poseidon2.NewDefaultPermutation())
	rng := rand.New(rand.NewPCG(1, 2))

	randomBlock := func() []koalabear.Element {
		b := make([]koalabear.Element, 8)
		for i := range b {
			b[i].SetUint64(uint64(rng.Uint32()))
		}
		return b
	}

	run := func(left, right, expected zkcInput) bool {
		return runZkc(t, map[string]zkcInput{"left": left, "right": right, "expected": expected}, compressDriver)
	}

	pairs := [][2][]koalabear.Element{{make([]koalabear.Element, 8), make([]koalabear.Element, 8)}}
	for range 4 {
		pairs = append(pairs, [2][]koalabear.Element{randomBlock(), randomBlock()})
	}

	for i, p := range pairs {
		out := make([]koalabear.Element, 8)
		compressor.CompressField(out, p[0], p[1])
		if !run(words(p[0]), words(p[1]), words(out)) {
			t.Errorf("pair %d: correct output rejected", i)
		}

		var one koalabear.Element
		one.SetOne()
		out[i%8].Add(&out[i%8], &one)
		if run(words(p[0]), words(p[1]), words(out)) {
			t.Errorf("pair %d: corrupted output accepted", i)
		}
	}
}
