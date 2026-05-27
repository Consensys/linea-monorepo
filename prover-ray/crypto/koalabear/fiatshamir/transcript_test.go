package fiatshamir

import (
	"reflect"
	"testing"

	"github.com/consensys/gnark-crypto/field/koalabear"
	ext "github.com/consensys/gnark-crypto/field/koalabear/extensions"
	"github.com/consensys/linea-monorepo/prover-ray/crypto/koalabear/hash"
)

type recordingHasher struct {
	current []koalabear.Element
	inputs  [][]koalabear.Element
}

func (h *recordingHasher) Reset() {
	h.current = h.current[:0]
}

func (h *recordingHasher) WriteElements(elements ...koalabear.Element) {
	h.current = append(h.current, elements...)
}

func (h *recordingHasher) WriteExt(elements ...ext.E4) {
	for _, e := range elements {
		h.current = append(h.current, e.B0.A0, e.B0.A1, e.B1.A0, e.B1.A1)
	}
}

func (h *recordingHasher) Sum() hash.Digest {
	input := append([]koalabear.Element(nil), h.current...)
	h.inputs = append(h.inputs, input)

	var out hash.Digest
	for i := range out {
		out[i].SetUint64(uint64(100*len(h.inputs) + i + 1))
	}
	return out
}

func TestComputeChallengeAbsorbsChallengeID(t *testing.T) {
	hasher := &recordingHasher{}
	transcript := NewTranscript(hasher, "zeta", "alpha_DEEP")

	zetaBinding := testElement(11)
	if err := transcript.Bind("zeta", []koalabear.Element{zetaBinding}); err != nil {
		t.Fatalf("bind zeta: %v", err)
	}
	zeta, err := transcript.ComputeChallenge("zeta")
	if err != nil {
		t.Fatalf("compute zeta: %v", err)
	}

	alphaBinding := testElement(22)
	if err := transcript.Bind("alpha_DEEP", []koalabear.Element{alphaBinding}); err != nil {
		t.Fatalf("bind alpha: %v", err)
	}
	if _, err := transcript.ComputeChallenge("alpha_DEEP"); err != nil {
		t.Fatalf("compute alpha: %v", err)
	}

	if len(hasher.inputs) != 2 {
		t.Fatalf("expected 2 hash inputs, got %d", len(hasher.inputs))
	}

	expectedZetaInput := append(hash.StringToElements(challengeIDDomainTag, "zeta"), zetaBinding)
	if !reflect.DeepEqual(hasher.inputs[0], expectedZetaInput) {
		t.Fatalf("unexpected zeta input:\nwant %#v\n got %#v", expectedZetaInput, hasher.inputs[0])
	}

	expectedAlphaInput := append(hash.StringToElements(challengeIDDomainTag, "alpha_DEEP"), zeta[:]...)
	expectedAlphaInput = append(expectedAlphaInput, alphaBinding)
	if !reflect.DeepEqual(hasher.inputs[1], expectedAlphaInput) {
		t.Fatalf("unexpected alpha input:\nwant %#v\n got %#v", expectedAlphaInput, hasher.inputs[1])
	}
}

func TestChallengeIDDomainSeparation(t *testing.T) {
	binding := testElement(7)

	hasherA := hash.NewPoseidon2SpongeHasher()
	transcriptA := NewTranscript(&hasherA, "a")
	if err := transcriptA.Bind("a", []koalabear.Element{binding}); err != nil {
		t.Fatalf("bind a: %v", err)
	}
	a, err := transcriptA.ComputeChallenge("a")
	if err != nil {
		t.Fatalf("compute a: %v", err)
	}

	hasherB := hash.NewPoseidon2SpongeHasher()
	transcriptB := NewTranscript(&hasherB, "b")
	if err := transcriptB.Bind("b", []koalabear.Element{binding}); err != nil {
		t.Fatalf("bind b: %v", err)
	}
	b, err := transcriptB.ComputeChallenge("b")
	if err != nil {
		t.Fatalf("compute b: %v", err)
	}

	if a == b {
		t.Fatalf("different challenge IDs produced the same challenge")
	}
}

func testElement(v uint64) koalabear.Element {
	var e koalabear.Element
	e.SetUint64(v)
	return e
}
