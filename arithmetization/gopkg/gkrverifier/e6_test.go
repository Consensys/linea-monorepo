package gkrverifier

import (
	"math/rand/v2"
	"testing"

	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/gnark-crypto/field/koalabear/extensions"
)

const (
	e6Driver   = "../../src/test/zkc/gkr/e6.zkc"
	modulusU32 = 0x7f000001
)

func e6Words(x *extensions.E6) zkcInput {
	return words(x.AppendBase(nil))
}

func TestE6(t *testing.T) {
	rng := rand.New(rand.NewPCG(3, 4))

	randomE6 := func() (x extensions.E6) {
		var coords [6]koalabear.Element
		for i := range coords {
			coords[i].SetUint64(uint64(rng.Uint32()))
		}
		return *x.SetBase(coords[:])
	}

	var minusOne koalabear.Element
	minusOne.SetInt64(-1)
	var maximal extensions.E6
	maximal.SetBase([]koalabear.Element{minusOne, minusOne, minusOne, minusOne, minusOne, minusOne})

	type testCase struct {
		a, b extensions.E6
		k    koalabear.Element
	}
	cases := []testCase{{a: maximal, b: maximal, k: minusOne}, {}}
	for range 5 {
		var k koalabear.Element
		k.SetUint64(uint64(rng.Uint32()))
		cases = append(cases, testCase{randomE6(), randomE6(), k})
	}

	inputsOf := func(c testCase) map[string]zkcInput {
		var sum, diff, prod, scaled, square extensions.E6
		sum.Add(&c.a, &c.b)
		diff.Sub(&c.a, &c.b)
		prod.Mul(&c.a, &c.b)
		scaled.MulByElement(&c.a, &c.k)
		square.Square(&c.a)
		return map[string]zkcInput{
			"a": e6Words(&c.a), "b": e6Words(&c.b), "k": words([]koalabear.Element{c.k}),
			"sum": e6Words(&sum), "diff": e6Words(&diff), "prod": e6Words(&prod),
			"scaled": e6Words(&scaled), "square": e6Words(&square),
		}
	}

	for i, c := range cases {
		if !runZkc(t, inputsOf(c), e6Driver) {
			t.Errorf("case %d: correct values rejected", i)
		}
	}

	c := cases[2]
	for _, name := range []string{"sum", "diff", "prod", "scaled", "square"} {
		in := inputsOf(c)
		in[name][5] = (in[name][5] + 1) % modulusU32
		if runZkc(t, in, e6Driver) {
			t.Errorf("corrupted %q accepted", name)
		}
	}

	// a word >= p, even one that is congruent to the right value, must be rejected
	for _, name := range []string{"a", "b", "k", "sum"} {
		in := inputsOf(c)
		in[name][0] += modulusU32
		if runZkc(t, in, e6Driver) {
			t.Errorf("non-canonical word in %q accepted", name)
		}
	}
}
