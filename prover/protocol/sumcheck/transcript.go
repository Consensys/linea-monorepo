// Package sumcheck implements a multilinear sumcheck protocol over the
// KoalaBear 4-extension field. It is wizard-agnostic: the transcript abstraction
// in this file lets it be driven by either a deterministic mock (tests) or by
// the wizard Fiat-Shamir state (compilation step).
package sumcheck

import (
	"crypto/sha256"
	"encoding/binary"

	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
)

// Transcript is a Fiat-Shamir transcript producing fext-valued challenges.
// Implementations are expected to absorb prover messages between calls to
// Challenge so that challenges are bound to the transcript.
type Transcript interface {
	Append(label string, xs ...fext.Element)
	Challenge(label string) fext.Element
}

// MockTranscript is a deterministic SHA-256-based transcript intended for tests.
// It is NOT secure for production use: there is no domain separation hardening
// and the squeeze layout is intentionally simple.
type MockTranscript struct {
	state [32]byte
}

func NewMockTranscript(seed string) *MockTranscript {
	h := sha256.Sum256([]byte("sumcheck-mock|" + seed))
	return &MockTranscript{state: h}
}

func (t *MockTranscript) Append(label string, xs ...fext.Element) {
	h := sha256.New()
	h.Write(t.state[:])
	h.Write([]byte(label))
	var buf [8]byte
	for i := range xs {
		x := &xs[i]
		binary.LittleEndian.PutUint32(buf[:4], x.B0.A0[0])
		h.Write(buf[:4])
		binary.LittleEndian.PutUint32(buf[:4], x.B0.A1[0])
		h.Write(buf[:4])
		binary.LittleEndian.PutUint32(buf[:4], x.B1.A0[0])
		h.Write(buf[:4])
		binary.LittleEndian.PutUint32(buf[:4], x.B1.A1[0])
		h.Write(buf[:4])
	}
	copy(t.state[:], h.Sum(nil))
}

func (t *MockTranscript) Challenge(label string) fext.Element {
	h := sha256.New()
	h.Write(t.state[:])
	h.Write([]byte("challenge|"))
	h.Write([]byte(label))
	digest := h.Sum(nil)
	copy(t.state[:], digest)

	var z fext.Element
	z.B0.A0.SetUint64(binary.LittleEndian.Uint64(digest[0:8]))
	z.B0.A1.SetUint64(binary.LittleEndian.Uint64(digest[8:16]))
	z.B1.A0.SetUint64(binary.LittleEndian.Uint64(digest[16:24]))
	z.B1.A1.SetUint64(binary.LittleEndian.Uint64(digest[24:32]))
	return z
}
