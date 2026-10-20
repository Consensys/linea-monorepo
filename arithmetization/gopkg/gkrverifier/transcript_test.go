package gkrverifier

import (
	"math/rand/v2"
	"testing"

	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/gnark-crypto/field/koalabear/extensions"
	extensionsHash "github.com/consensys/gnark-crypto/field/koalabear/extensions/hash"
	"github.com/consensys/gnark-crypto/field/koalabear/poseidon2"
)

const transcriptDriver = "../../src/test/zkc/gkr/transcript.zkc"

const (
	opSeed = iota
	opAbsorb
	opSqueeze
)

// step is an operation of a transcript script. n is the number of elements a seed write or an
// absorption takes; a squeeze takes none.
type step struct{ op, n int }

// referenceTranscript is gkrcore.HashTranscript, whose type is internal to gnark, over the hasher
// that the GKR prover for the Poseidon2 compression uses.
type referenceTranscript struct {
	h        *extensionsHash.BaseFieldHashAdapterE6
	absorbed bool
}

func newReferenceTranscript() *referenceTranscript {
	return &referenceTranscript{h: extensionsHash.NewBaseFieldHashAdapterE6(poseidon2.NewMerkleDamgardHasher())}
}

func (t *referenceTranscript) seed(e ...extensions.E6) { t.h.WriteField(e...) }

func (t *referenceTranscript) absorb(e ...extensions.E6) {
	if len(e) == 0 {
		return
	}
	t.h.WriteField(e...)
	t.absorbed = true
}

func (t *referenceTranscript) squeeze() extensions.E6 {
	if !t.absorbed {
		var zero extensions.E6
		t.h.WriteField(zero)
	}
	t.absorbed = false
	t.h.PadBlock()
	return t.h.SumField()
}

// transcriptInputs runs the script on the reference transcript, with random elements, and returns
// the inputs of the driver.
func transcriptInputs(rng *rand.Rand, script []step) map[string]zkcInput {
	var ops, elements, expected zkcInput
	ref := newReferenceTranscript()
	for _, s := range script {
		ops = append(ops, uint32(s.op), uint32(s.n))
		if s.op == opSqueeze {
			challenge := ref.squeeze()
			expected = append(expected, e6Words(&challenge)...)
			continue
		}
		batch := make([]extensions.E6, s.n)
		for i := range batch {
			var coords [6]koalabear.Element
			for j := range coords {
				coords[j].SetUint64(uint64(rng.Uint32()))
			}
			batch[i].SetBase(coords[:])
			elements = append(elements, e6Words(&batch[i])...)
		}
		if s.op == opSeed {
			ref.seed(batch...)
		} else {
			ref.absorb(batch...)
		}
	}
	return map[string]zkcInput{
		"nb_ops":   {uint32(len(script))},
		"ops":      ops,
		"elements": elements,
		"expected": expected,
	}
}

func TestTranscript(t *testing.T) {
	rng := rand.New(rand.NewPCG(5, 6))

	var allSizes []step
	for n := range 10 {
		allSizes = append(allSizes, step{opAbsorb, n}, step{opSqueeze, 0})
	}

	var randomScript []step
	for range 40 {
		switch op := rng.IntN(3); op {
		case opSqueeze:
			randomScript = append(randomScript, step{opSqueeze, 0})
		default:
			randomScript = append(randomScript, step{op, rng.IntN(10)})
		}
	}

	scripts := []struct {
		name   string
		script []step
	}{
		{"squeeze right after the seed", []step{{opSeed, 1}, {opSqueeze, 0}}},
		{"squeeze without any seed", []step{{opSqueeze, 0}}},
		{"consecutive squeezes", []step{{opSeed, 2}, {opAbsorb, 1}, {opSqueeze, 0}, {opSqueeze, 0}, {opSqueeze, 0}}},
		{"absorptions of every size", allSizes},
		{"empty absorption sets no flag", []step{{opSeed, 3}, {opAbsorb, 0}, {opSqueeze, 0}, {opAbsorb, 0}, {opSqueeze, 0}}},
		{"block boundary", []step{{opAbsorb, 4}, {opSqueeze, 0}, {opAbsorb, 8}, {opSqueeze, 0}, {opSeed, 4}, {opAbsorb, 4}, {opSqueeze, 0}}},
		{"block boundary across writes", []step{{opAbsorb, 2}, {opAbsorb, 2}, {opSqueeze, 0}, {opSeed, 1}, {opAbsorb, 3}, {opSqueeze, 0}}},
		{"seed split from absorption", []step{{opSeed, 5}, {opSeed, 2}, {opAbsorb, 6}, {opSqueeze, 0}, {opAbsorb, 1}, {opSqueeze, 0}}},
		{"random operations", randomScript},
	}

	for _, sc := range scripts {
		t.Run(sc.name, func(t *testing.T) {
			inputs := transcriptInputs(rng, sc.script)
			if !runZkc(t, inputs, transcriptDriver) {
				t.Fatal("correct challenges rejected")
			}

			last := len(inputs["expected"]) - 1
			bad := transcriptInputs(rng, sc.script)
			bad["expected"][last] = (bad["expected"][last] + 1) % modulusU32
			if runZkc(t, bad, transcriptDriver) {
				t.Error("corrupted challenge accepted")
			}
		})
	}

	t.Run("non-canonical word", func(t *testing.T) {
		inputs := transcriptInputs(rng, []step{{opAbsorb, 2}, {opSqueeze, 0}})
		inputs["elements"][3] += modulusU32
		if runZkc(t, inputs, transcriptDriver) {
			t.Error("non-canonical word accepted")
		}
	})

	t.Run("unknown operation", func(t *testing.T) {
		inputs := transcriptInputs(rng, []step{{opSqueeze, 0}})
		inputs["ops"][0] = 3
		if runZkc(t, inputs, transcriptDriver) {
			t.Error("unknown operation accepted")
		}
	})
}
