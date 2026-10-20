package gkrverifier

import (
	"math/rand/v2"
	"os"
	"testing"

	"github.com/consensys/gnark-crypto/field/koalabear/extensions"
	"github.com/consensys/gnark-crypto/field/koalabear/extensions/polynomial"
	"github.com/consensys/gnark-crypto/field/koalabear/poseidon2"
	"github.com/consensys/gnark/constraint"
	"github.com/consensys/gnark/gkr"
	gkrposeidon2 "github.com/consensys/gnark/gkr/permutation/gkr-poseidon2/koalabear"
)

const roundDriver = "../../src/test/zkc/gkr/round.zkc"

const (
	modeEval = iota
	modeSumcheck
	modeSingleSource
	modeEq
)

// roundScript builds the inputs of the round driver: the cases, and the data they consume with the
// position in it of every expected element.
type roundScript struct {
	rng        *rand.Rand
	transcript *referenceTranscript
	cases      zkcInput
	data       zkcInput
	expectedAt []int
}

func newRoundScript(rng *rand.Rand) *roundScript {
	return &roundScript{rng: rng, transcript: newReferenceTranscript()}
}

func (s *roundScript) begin(mode, d, n int) {
	s.cases = append(s.cases, uint32(mode), uint32(d), uint32(n))
}

// random appends n random elements to the data and returns them.
func (s *roundScript) random(n int) []extensions.E6 {
	res := make([]extensions.E6, n)
	for i := range res {
		res[i] = randomE6(s.rng)
		s.data = append(s.data, e6Words(&res[i])...)
	}
	return res
}

func (s *roundScript) expect(e extensions.E6) {
	s.expectedAt = append(s.expectedAt, len(s.data))
	s.data = append(s.data, e6Words(&e)...)
}

func (s *roundScript) inputs() map[string]zkcInput {
	return map[string]zkcInput{"nb_cases": {uint32(len(s.cases) / 3)}, "cases": s.cases, "data": s.data}
}

// eval adds the evaluation of a polynomial of degree d given by its values on 0, …, d.
func (s *roundScript) eval(d int) {
	s.begin(modeEval, d, 0)
	values := s.random(d + 1)
	r := s.random(1)[0]
	s.expect(evalOnRange(values, &r))
}

// evalOnRange evaluates at r the polynomial whose values on 0, 1, … are the given ones.
func evalOnRange(values []extensions.E6, r *extensions.E6) extensions.E6 {
	coeffs := polynomial.InterpolateOnRangeE6(values)
	return coeffs.Eval(r)
}

// sumcheck adds n rounds of degree d.
func (s *roundScript) sumcheck(d, n int) {
	s.begin(modeSumcheck, d, n)
	claim := s.random(1)[0]
	polys := s.random(n * d)
	for j := range n {
		g := polys[j*d : (j+1)*d]
		values := append([]extensions.E6{{}}, g...)
		values[0].Sub(&claim, &g[0])

		s.transcript.absorb(g...)
		r := s.transcript.squeeze()
		claim = evalOnRange(values, &r)
		s.expect(claim)
	}
}

// singleSource adds n single-source rounds of degree d, with the division that the ZkC avoids.
func (s *roundScript) singleSource(d, n int) {
	s.begin(modeSingleSource, d, n)
	claim := s.random(1)[0]
	q := s.random(n)
	polys := s.random(n * d)
	for j := range n {
		g := polys[j*d : (j+1)*d]
		var oneMinusQ, qg1 extensions.E6
		oneMinusQ.SetOne()
		oneMinusQ.Sub(&oneMinusQ, &q[j])
		qg1.Mul(&q[j], &g[0])

		values := append([]extensions.E6{{}}, g...)
		values[0].Sub(&claim, &qg1)
		values[0].Div(&values[0], &oneMinusQ)

		s.transcript.absorb(g...)
		r := s.transcript.squeeze()
		claim = evalOnRange(values, &r)
		s.expect(claim)
	}
}

// eq adds eq(p, r) for random points of n elements.
func (s *roundScript) eq(n int) {
	s.begin(modeEq, 0, n)
	p := s.random(n)
	r := s.random(n)
	s.expect(polynomial.EvalEqE6(p, r))
}

func TestRound(t *testing.T) {
	rng := rand.New(rand.NewPCG(11, 12))

	compressor, err := gkrposeidon2.NewCompressor(poseidon2.GetDefaultParameters())
	if err != nil {
		t.Fatal(err)
	}
	if maxDegree := maxLevelDegree(compressor.Circuit, compressor.Schedule); maxDegree < 4 {
		t.Fatalf("the compressor's levels have at most %d values per round polynomial, the test covers degrees up to 4", maxDegree)
	}
	tables := t.TempDir() + "/lagrange.zkc"
	writeLagrange(t, tables, compressor.Circuit, compressor.Schedule)

	scripts := []struct {
		name string
		add  func(s *roundScript)
	}{
		{"evaluation", func(s *roundScript) {
			for d := 1; d <= 4; d++ {
				s.eval(d)
			}
		}},
		{"sumcheck rounds", func(s *roundScript) {
			for d := 1; d <= 4; d++ {
				for _, n := range []int{1, 2, 5} {
					s.sumcheck(d, n)
				}
			}
		}},
		{"single-source rounds", func(s *roundScript) {
			for d := 1; d <= 4; d++ {
				for _, n := range []int{1, 2, 5} {
					s.singleSource(d, n)
				}
			}
		}},
		{"eq", func(s *roundScript) {
			for _, n := range []int{1, 2, 3, 5, 8} {
				s.eq(n)
			}
		}},
	}

	for _, sc := range scripts {
		t.Run(sc.name, func(t *testing.T) {
			s := newRoundScript(rng)
			sc.add(s)
			if !runZkc(t, s.inputs(), roundDriver, tables) {
				t.Fatal("correct values rejected")
			}
			for _, i := range []int{0, len(s.expectedAt) / 2, len(s.expectedAt) - 1} {
				bad := s.inputs()
				bad["data"] = append(zkcInput(nil), bad["data"]...)
				w := &bad["data"][s.expectedAt[i]+5]
				*w = (*w + 1) % modulusU32
				if runZkc(t, bad, roundDriver, tables) {
					t.Errorf("corrupted expected value %d accepted", i)
				}
			}
		})
	}
}

func writeLagrange(t *testing.T, path string, circuit gkr.SerializableCircuit, schedule constraint.GkrProvingSchedule) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := WriteLagrangeTables(f, circuit, schedule); err != nil {
		t.Fatal(err)
	}
}
