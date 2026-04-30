//go:build cuda

package plonk2

import (
	"fmt"
	"sync"
	"unsafe"

	blsfr "github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	bnfr "github.com/consensys/gnark-crypto/ecc/bn254/fr"
	bwfr "github.com/consensys/gnark-crypto/ecc/bw6-761/fr"
)

type genericProofScratch struct {
	mu     sync.Mutex
	curve  Curve
	n      int
	limbs  int
	qcpLen int

	lroRaw       [3][]uint64
	lroCanonical [3][]uint64
	lroBlinded   [3][]uint64
	qkLagrange   []uint64
	qkCanonical  []uint64
	zRaw         []uint64
	zCanonical   []uint64
	zBlinded     []uint64
	bsb22Raw     [][]uint64
	bsb22Canon   [][]uint64
	commitments  []uint64
	hFull        []uint64

	bls12377 *bls12377ProofScratch
}

type bls12377ProofScratch struct {
	lBlinded, rBlinded, oBlinded []blsfr.Element
	zBlinded                     []blsfr.Element
	h                            [3][]blsfr.Element
	pi2                          [][]blsfr.Element
	fixed                        bls12377FixedCanonical
	fixedRaw                     []uint64
	openZ                        []blsfr.Element
	linPol                       []blsfr.Element
	folded                       []blsfr.Element
	bsb22Committed               [][]blsfr.Element
}

func newGenericProofScratch(curve Curve, n, qcpLen int) (*genericProofScratch, error) {
	if n <= 0 {
		return nil, fmt.Errorf("plonk2: scratch domain size must be positive")
	}
	limbs := scalarLimbs(curve)
	elementWords := n * limbs
	s := &genericProofScratch{
		curve:       curve,
		n:           n,
		limbs:       limbs,
		qcpLen:      qcpLen,
		qkLagrange:  make([]uint64, elementWords),
		qkCanonical: make([]uint64, elementWords),
		zRaw:        make([]uint64, elementWords),
		zCanonical:  make([]uint64, elementWords),
		zBlinded:    make([]uint64, (n+genericZBlindingOrder+1)*limbs),
		bsb22Raw:    make([][]uint64, qcpLen),
		bsb22Canon:  make([][]uint64, qcpLen),
		commitments: make([]uint64, qcpLen*limbs),
		hFull:       make([]uint64, 4*n*limbs),
	}
	for i := range s.lroRaw {
		s.lroRaw[i] = make([]uint64, elementWords)
		s.lroCanonical[i] = make([]uint64, elementWords)
		s.lroBlinded[i] = make([]uint64, (n+genericLROBlindingOrder+1)*limbs)
	}
	for i := 0; i < qcpLen; i++ {
		s.bsb22Raw[i] = make([]uint64, elementWords)
		s.bsb22Canon[i] = make([]uint64, elementWords)
	}
	if curve == CurveBLS12377 {
		s.bls12377 = newBLS12377ProofScratch(n, qcpLen)
	}
	return s, nil
}

func newBLS12377ProofScratch(n, qcpLen int) *bls12377ProofScratch {
	out := &bls12377ProofScratch{
		lBlinded:       make([]blsfr.Element, n+genericLROBlindingOrder+1),
		rBlinded:       make([]blsfr.Element, n+genericLROBlindingOrder+1),
		oBlinded:       make([]blsfr.Element, n+genericLROBlindingOrder+1),
		zBlinded:       make([]blsfr.Element, n+genericZBlindingOrder+1),
		fixedRaw:       make([]uint64, n*blsfr.Limbs),
		openZ:          make([]blsfr.Element, n+genericZBlindingOrder+1),
		linPol:         make([]blsfr.Element, n+genericZBlindingOrder+1),
		folded:         make([]blsfr.Element, n+genericZBlindingOrder+1),
		pi2:            make([][]blsfr.Element, qcpLen),
		bsb22Committed: make([][]blsfr.Element, qcpLen),
	}
	out.h[0] = make([]blsfr.Element, n+2)
	out.h[1] = make([]blsfr.Element, n+2)
	out.h[2] = make([]blsfr.Element, n+2)
	out.fixed = bls12377FixedCanonical{
		ql:  make([]blsfr.Element, n),
		qr:  make([]blsfr.Element, n),
		qm:  make([]blsfr.Element, n),
		qo:  make([]blsfr.Element, n),
		qk:  make([]blsfr.Element, n),
		s1:  make([]blsfr.Element, n),
		s2:  make([]blsfr.Element, n),
		s3:  make([]blsfr.Element, n),
		qcp: make([][]blsfr.Element, qcpLen),
	}
	for i := 0; i < qcpLen; i++ {
		out.pi2[i] = make([]blsfr.Element, n)
		out.fixed.qcp[i] = make([]blsfr.Element, n)
		out.bsb22Committed[i] = make([]blsfr.Element, n)
	}
	return out
}

func (s *genericProofScratch) hShardRaw(index int) []uint64 {
	shard := s.n + 2
	return rawElementSlice(s.hFull, s.limbs, index*shard, (index+1)*shard)
}

func copyBN254FrToRaw(dst []uint64, values []bnfr.Element) {
	copy(dst, rawBN254Fr(values))
}

func copyBLS12377FrToRaw(dst []uint64, values []blsfr.Element) {
	copy(dst, rawBLS12377Fr(values))
}

func clearBLS12377Vector(values []blsfr.Element) {
	var zero blsfr.Element
	for i := range values {
		values[i] = zero
	}
}

func copyBW6761FrToRaw(dst []uint64, values []bwfr.Element) {
	copy(dst, rawBW6761Fr(values))
}

func writeBN254RawElement(dst []uint64, value *bnfr.Element) {
	copy(dst, rawBN254Fr(unsafe.Slice(value, 1)))
}

func writeBLS12377RawElement(dst []uint64, value *blsfr.Element) {
	copy(dst, rawBLS12377Fr(unsafe.Slice(value, 1)))
}

func writeBW6761RawElement(dst []uint64, value *bwfr.Element) {
	copy(dst, rawBW6761Fr(unsafe.Slice(value, 1)))
}
