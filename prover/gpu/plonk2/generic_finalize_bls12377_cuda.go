//go:build cuda

package plonk2

import (
	"fmt"
	"hash"
	"math/big"
	"runtime"

	bls12377 "github.com/consensys/gnark-crypto/ecc/bls12-377"
	blsfr "github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	blskzg "github.com/consensys/gnark-crypto/ecc/bls12-377/kzg"
	fiatshamir "github.com/consensys/gnark-crypto/fiat-shamir"
	"github.com/consensys/gnark/backend"
	blsplonk "github.com/consensys/gnark/backend/plonk/bls12-377"

	"github.com/consensys/linea-monorepo/prover/utils/parallel"
)

func (b *genericGPUBackend) finalizeBLS12377(
	artifacts *genericProveArtifacts,
	proverConfig *backend.ProverConfig,
) error {
	proof := artifacts.proof.(*blsplonk.Proof)
	pk := b.pk.(*blsplonk.ProvingKey)
	scratch := artifacts.scratch
	if scratch == nil || scratch.bls12377 == nil {
		return b.finalizeBLS12377Allocating(artifacts, proverConfig)
	}
	blsScratch := scratch.bls12377

	lBlinded := blsScratch.lBlinded[:len(artifacts.lroBlinded[0])/blsfr.Limbs]
	if err := bls12377RawToFieldInto(lBlinded, artifacts.lroBlinded[0]); err != nil {
		return err
	}
	rBlinded := blsScratch.rBlinded[:len(artifacts.lroBlinded[1])/blsfr.Limbs]
	if err := bls12377RawToFieldInto(rBlinded, artifacts.lroBlinded[1]); err != nil {
		return err
	}
	oBlinded := blsScratch.oBlinded[:len(artifacts.lroBlinded[2])/blsfr.Limbs]
	if err := bls12377RawToFieldInto(oBlinded, artifacts.lroBlinded[2]); err != nil {
		return err
	}
	zBlinded := blsScratch.zBlinded[:len(artifacts.zBlinded)/blsfr.Limbs]
	if err := bls12377RawToFieldInto(zBlinded, artifacts.zBlinded); err != nil {
		return err
	}
	var h [3][]blsfr.Element
	if err := bls12377HRawToFieldInto(&h, blsScratch.h, artifacts.hRaw); err != nil {
		return err
	}
	pi2 := blsScratch.pi2[:len(artifacts.bsb22Canonical)]
	if err := bls12377RawMatrixToFieldInto(pi2, artifacts.bsb22Canonical); err != nil {
		return err
	}
	fixed, err := b.bls12377FixedCanonical(blsScratch)
	if err != nil {
		return err
	}

	zeta := bls12377ScalarFromRaw(artifacts.zetaRaw)
	alpha := bls12377ScalarFromRaw(artifacts.alphaRaw)
	beta := bls12377ScalarFromRaw(artifacts.betaRaw)
	gamma := bls12377ScalarFromRaw(artifacts.gammaRaw)

	var shiftedZeta blsfr.Element
	shiftedZeta.Mul(&zeta, &pk.Vk.Generator)
	if err := b.openBLS12377ShiftedZ(proof, zBlinded, shiftedZeta, blsScratch.openZ); err != nil {
		return err
	}

	lZeta := evaluateBLS12377(lBlinded, zeta)
	rZeta := evaluateBLS12377(rBlinded, zeta)
	oZeta := evaluateBLS12377(oBlinded, zeta)
	s1Zeta := evaluateBLS12377(fixed.s1, zeta)
	s2Zeta := evaluateBLS12377(fixed.s2, zeta)
	qcpZeta := make([]blsfr.Element, len(fixed.qcp))
	for i := range fixed.qcp {
		qcpZeta[i] = evaluateBLS12377(fixed.qcp[i], zeta)
	}

	linPol := computeBLS12377LinearizedPolynomial(
		uint64(b.state.n),
		pk.Vk.CosetShift,
		pk.Vk.SizeInv,
		lZeta,
		rZeta,
		oZeta,
		alpha,
		beta,
		gamma,
		zeta,
		proof.ZShiftedOpening.ClaimedValue,
		s1Zeta,
		s2Zeta,
		qcpZeta,
		copyBLS12377Elements(blsScratch.linPol[:len(zBlinded)], zBlinded),
		blsScratch.linPol[:len(zBlinded)],
		pi2,
		h,
		fixed,
	)
	linDigestRaw, err := b.state.kzg.CommitRaw(genericRawBLS12377Fr(linPol))
	if err != nil {
		return fmt.Errorf("commit linearized polynomial: %w", err)
	}
	linDigestAny, err := (bls12377ProofOps{}).rawCommitmentToDigest(linDigestRaw)
	if err != nil {
		return err
	}
	linDigest := linDigestAny.(blskzg.Digest)

	polys := make([][]blsfr.Element, 6+len(fixed.qcp))
	polys[0] = linPol
	polys[1] = lBlinded
	polys[2] = rBlinded
	polys[3] = oBlinded
	polys[4] = fixed.s1
	polys[5] = fixed.s2
	copy(polys[6:], fixed.qcp)

	digests := make([]bls12377.G1Affine, len(polys))
	digests[0] = linDigest
	digests[1] = proof.LRO[0]
	digests[2] = proof.LRO[1]
	digests[3] = proof.LRO[2]
	digests[4] = pk.Vk.S[0]
	digests[5] = pk.Vk.S[1]
	copy(digests[6:], pk.Vk.Qcp)

	if err := b.batchOpenBLS12377(
		proof,
		polys,
		digests,
		zeta,
		proverConfig.KZGFoldingHash,
		proof.ZShiftedOpening.ClaimedValue.Marshal(),
		blsScratch.folded,
	); err != nil {
		return err
	}
	return nil
}

func (b *genericGPUBackend) finalizeBLS12377Allocating(
	artifacts *genericProveArtifacts,
	proverConfig *backend.ProverConfig,
) error {
	proof := artifacts.proof.(*blsplonk.Proof)
	pk := b.pk.(*blsplonk.ProvingKey)

	lBlinded, err := bls12377RawToField(artifacts.lroBlinded[0])
	if err != nil {
		return err
	}
	rBlinded, err := bls12377RawToField(artifacts.lroBlinded[1])
	if err != nil {
		return err
	}
	oBlinded, err := bls12377RawToField(artifacts.lroBlinded[2])
	if err != nil {
		return err
	}
	zBlinded, err := bls12377RawToField(artifacts.zBlinded)
	if err != nil {
		return err
	}
	h, err := bls12377HRawToField(artifacts.hRaw)
	if err != nil {
		return err
	}
	pi2, err := bls12377RawMatrixToField(artifacts.bsb22Canonical)
	if err != nil {
		return err
	}
	fixed, err := b.bls12377FixedCanonical(nil)
	if err != nil {
		return err
	}

	zeta := bls12377ScalarFromRaw(artifacts.zetaRaw)
	alpha := bls12377ScalarFromRaw(artifacts.alphaRaw)
	beta := bls12377ScalarFromRaw(artifacts.betaRaw)
	gamma := bls12377ScalarFromRaw(artifacts.gammaRaw)

	var shiftedZeta blsfr.Element
	shiftedZeta.Mul(&zeta, &pk.Vk.Generator)
	if err := b.openBLS12377ShiftedZ(proof, zBlinded, shiftedZeta, nil); err != nil {
		return err
	}

	lZeta := evaluateBLS12377(lBlinded, zeta)
	rZeta := evaluateBLS12377(rBlinded, zeta)
	oZeta := evaluateBLS12377(oBlinded, zeta)
	s1Zeta := evaluateBLS12377(fixed.s1, zeta)
	s2Zeta := evaluateBLS12377(fixed.s2, zeta)
	qcpZeta := make([]blsfr.Element, len(fixed.qcp))
	for i := range fixed.qcp {
		qcpZeta[i] = evaluateBLS12377(fixed.qcp[i], zeta)
	}

	linPol := computeBLS12377LinearizedPolynomial(
		uint64(b.state.n),
		pk.Vk.CosetShift,
		pk.Vk.SizeInv,
		lZeta,
		rZeta,
		oZeta,
		alpha,
		beta,
		gamma,
		zeta,
		proof.ZShiftedOpening.ClaimedValue,
		s1Zeta,
		s2Zeta,
		qcpZeta,
		append([]blsfr.Element(nil), zBlinded...),
		pi2,
		h,
		fixed,
	)
	linDigestRaw, err := b.state.kzg.CommitRaw(genericRawBLS12377Fr(linPol))
	if err != nil {
		return fmt.Errorf("commit linearized polynomial: %w", err)
	}
	linDigestAny, err := (bls12377ProofOps{}).rawCommitmentToDigest(linDigestRaw)
	if err != nil {
		return err
	}
	linDigest := linDigestAny.(blskzg.Digest)

	polys := make([][]blsfr.Element, 6+len(fixed.qcp))
	polys[0] = linPol
	polys[1] = lBlinded
	polys[2] = rBlinded
	polys[3] = oBlinded
	polys[4] = fixed.s1
	polys[5] = fixed.s2
	copy(polys[6:], fixed.qcp)

	digests := make([]bls12377.G1Affine, len(polys))
	digests[0] = linDigest
	digests[1] = proof.LRO[0]
	digests[2] = proof.LRO[1]
	digests[3] = proof.LRO[2]
	digests[4] = pk.Vk.S[0]
	digests[5] = pk.Vk.S[1]
	copy(digests[6:], pk.Vk.Qcp)

	if err := b.batchOpenBLS12377(
		proof,
		polys,
		digests,
		zeta,
		proverConfig.KZGFoldingHash,
		proof.ZShiftedOpening.ClaimedValue.Marshal(),
		nil,
	); err != nil {
		return err
	}
	return nil
}

func (b *genericGPUBackend) openBLS12377ShiftedZ(
	proof *blsplonk.Proof,
	zBlinded []blsfr.Element,
	point blsfr.Element,
) error {
	quotient := append([]blsfr.Element(nil), zBlinded...)
	hornerQuotientBLS12377(quotient, point)
	proof.ZShiftedOpening.ClaimedValue = quotient[0]
	digest, err := b.commitBLS12377Polynomial(quotient[1:])
	if err != nil {
		return fmt.Errorf("open shifted Z: %w", err)
	}
	proof.ZShiftedOpening.H = digest
	return nil
}

func (b *genericGPUBackend) batchOpenBLS12377(
	proof *blsplonk.Proof,
	polys [][]blsfr.Element,
	digests []bls12377.G1Affine,
	point blsfr.Element,
	foldingHash hash.Hash,
	dataTranscript []byte,
) error {
	claimed := make([]blsfr.Element, len(polys))
	for i := range polys {
		claimed[i] = evaluateBLS12377(polys[i], point)
	}
	proof.BatchedProof.ClaimedValues = claimed

	gamma, err := deriveBLS12377BatchGamma(point, digests, claimed, foldingHash, dataTranscript)
	if err != nil {
		return fmt.Errorf("derive batch opening challenge: %w", err)
	}
	largest := 0
	for i := range polys {
		if len(polys[i]) > largest {
			largest = len(polys[i])
		}
	}
	folded := make([]blsfr.Element, largest)
	gammaPowers := make([]blsfr.Element, len(polys))
	gammaPowers[0].SetOne()
	for i := 1; i < len(gammaPowers); i++ {
		gammaPowers[i].Mul(&gammaPowers[i-1], &gamma)
	}
	parallel.Execute(largest, func(start, stop int) {
		for j := start; j < stop; j++ {
			var acc, term blsfr.Element
			for i := range polys {
				if j >= len(polys[i]) {
					continue
				}
				term.Mul(&polys[i][j], &gammaPowers[i])
				acc.Add(&acc, &term)
			}
			folded[j] = acc
		}
	})
	if largest == 0 {
		return fmt.Errorf("batch opening: no polynomial coefficients")
	}
	var foldedEval blsfr.Element
	for i := len(claimed) - 1; i >= 0; i-- {
		foldedEval.Mul(&foldedEval, &gamma).Add(&foldedEval, &claimed[i])
	}
	folded[0].Sub(&folded[0], &foldedEval)
	hornerQuotientBLS12377(folded, point)
	digest, err := b.commitBLS12377Polynomial(folded[1:])
	if err != nil {
		return fmt.Errorf("batch opening: %w", err)
	}
	proof.BatchedProof.H = digest
	return nil
}

func (b *genericGPUBackend) commitBLS12377Polynomial(poly []blsfr.Element) (blskzg.Digest, error) {
	raw, err := b.state.kzg.CommitRaw(genericRawBLS12377Fr(poly))
	if err != nil {
		return blskzg.Digest{}, err
	}
	digestAny, err := (bls12377ProofOps{}).rawCommitmentToDigest(raw)
	if err != nil {
		return blskzg.Digest{}, err
	}
	return digestAny.(blskzg.Digest), nil
}

func deriveBLS12377BatchGamma(
	point blsfr.Element,
	digests []bls12377.G1Affine,
	claimed []blsfr.Element,
	foldingHash hash.Hash,
	dataTranscript []byte,
) (blsfr.Element, error) {
	fs := fiatshamir.NewTranscript(foldingHash, "gamma")
	var gamma blsfr.Element
	if err := fs.Bind("gamma", point.Marshal()); err != nil {
		return gamma, err
	}
	for i := range digests {
		if err := fs.Bind("gamma", digests[i].Marshal()); err != nil {
			return gamma, err
		}
	}
	for i := range claimed {
		if err := fs.Bind("gamma", claimed[i].Marshal()); err != nil {
			return gamma, err
		}
	}
	if len(dataTranscript) > 0 {
		if err := fs.Bind("gamma", dataTranscript); err != nil {
			return gamma, err
		}
	}
	gammaBytes, err := fs.ComputeChallenge("gamma")
	if err != nil {
		return gamma, err
	}
	gamma.SetBytes(gammaBytes)
	return gamma, nil
}

func hornerQuotientBLS12377(poly []blsfr.Element, z blsfr.Element) {
	n := len(poly)
	nCPU := runtime.GOMAXPROCS(0)
	if n < 4096 || nCPU < 2 {
		for i := n - 2; i >= 0; i-- {
			var tmp blsfr.Element
			tmp.Mul(&poly[i+1], &z)
			poly[i].Add(&poly[i], &tmp)
		}
		return
	}

	chunkSize := (n + nCPU - 1) / nCPU
	numChunks := (n + chunkSize - 1) / chunkSize

	parallel.Execute(numChunks, func(start, stop int) {
		for c := start; c < stop; c++ {
			lo := c * chunkSize
			hi := lo + chunkSize
			if hi > n {
				hi = n
			}
			for i := hi - 2; i >= lo; i-- {
				var tmp blsfr.Element
				tmp.Mul(&poly[i+1], &z)
				poly[i].Add(&poly[i], &tmp)
			}
		}
	})

	zk := expBLS12377(z, chunkSize)
	carries := make([]blsfr.Element, numChunks)
	for c := numChunks - 2; c >= 0; c-- {
		nextLo := (c + 1) * chunkSize
		nextLen := chunkSize
		if nextLo+nextLen > n {
			nextLen = n - nextLo
		}
		zkc := zk
		if nextLen != chunkSize {
			zkc = expBLS12377(z, nextLen)
		}
		var tmp blsfr.Element
		tmp.Mul(&carries[c+1], &zkc)
		carries[c].Add(&poly[nextLo], &tmp)
	}

	parallel.Execute(numChunks, func(start, stop int) {
		for c := start; c < stop; c++ {
			if carries[c].IsZero() {
				continue
			}
			lo := c * chunkSize
			hi := lo + chunkSize
			if hi > n {
				hi = n
			}
			var zPow blsfr.Element
			zPow.Set(&z)
			for i := hi - 1; i >= lo; i-- {
				var corr blsfr.Element
				corr.Mul(&zPow, &carries[c])
				poly[i].Add(&poly[i], &corr)
				zPow.Mul(&zPow, &z)
			}
		}
	})
}

type bls12377FixedCanonical struct {
	ql, qr, qm, qo, qk []blsfr.Element
	s1, s2, s3         []blsfr.Element
	qcp                [][]blsfr.Element
}

func (b *genericGPUBackend) bls12377FixedCanonical() (bls12377FixedCanonical, error) {
	convert := func(v *FrVector) ([]blsfr.Element, error) {
		raw, err := copyFrVectorRaw(v)
		if err != nil {
			return nil, err
		}
		return bls12377RawToField(raw)
	}
	ql, err := convert(b.state.fixed.ql)
	if err != nil {
		return bls12377FixedCanonical{}, err
	}
	qr, err := convert(b.state.fixed.qr)
	if err != nil {
		return bls12377FixedCanonical{}, err
	}
	qm, err := convert(b.state.fixed.qm)
	if err != nil {
		return bls12377FixedCanonical{}, err
	}
	qo, err := convert(b.state.fixed.qo)
	if err != nil {
		return bls12377FixedCanonical{}, err
	}
	qk, err := convert(b.state.fixed.qk)
	if err != nil {
		return bls12377FixedCanonical{}, err
	}
	s1, err := convert(b.state.fixed.s1)
	if err != nil {
		return bls12377FixedCanonical{}, err
	}
	s2, err := convert(b.state.fixed.s2)
	if err != nil {
		return bls12377FixedCanonical{}, err
	}
	s3, err := convert(b.state.fixed.s3)
	if err != nil {
		return bls12377FixedCanonical{}, err
	}
	qcp := make([][]blsfr.Element, len(b.state.qcp))
	for i := range b.state.qcp {
		qcp[i], err = convert(b.state.qcp[i])
		if err != nil {
			return bls12377FixedCanonical{}, err
		}
	}
	return bls12377FixedCanonical{
		ql: ql, qr: qr, qm: qm, qo: qo, qk: qk,
		s1: s1, s2: s2, s3: s3,
		qcp: qcp,
	}, nil
}

func computeBLS12377LinearizedPolynomial(
	n uint64,
	cosetShift, sizeInv blsfr.Element,
	lZeta, rZeta, oZeta, alpha, beta, gamma, zeta, zu blsfr.Element,
	s1Zeta, s2Zeta blsfr.Element,
	qcpZeta []blsfr.Element,
	zBlinded []blsfr.Element,
	pi2 [][]blsfr.Element,
	h [3][]blsfr.Element,
	fixed bls12377FixedCanonical,
) []blsfr.Element {
	var rl blsfr.Element
	rl.Mul(&rZeta, &lZeta)

	var s1, tmp blsfr.Element
	s1.Mul(&s1Zeta, &beta).Add(&s1, &lZeta).Add(&s1, &gamma)
	tmp.Mul(&s2Zeta, &beta).Add(&tmp, &rZeta).Add(&tmp, &gamma)
	s1.Mul(&s1, &tmp).Mul(&s1, &zu).Mul(&s1, &beta).Mul(&s1, &alpha)

	var s2, uzeta, uuzeta blsfr.Element
	uzeta.Mul(&zeta, &cosetShift)
	uuzeta.Mul(&uzeta, &cosetShift)
	s2.Mul(&beta, &zeta).Add(&s2, &lZeta).Add(&s2, &gamma)
	tmp.Mul(&beta, &uzeta).Add(&tmp, &rZeta).Add(&tmp, &gamma)
	s2.Mul(&s2, &tmp)
	tmp.Mul(&beta, &uuzeta).Add(&tmp, &oZeta).Add(&tmp, &gamma)
	s2.Mul(&s2, &tmp).Neg(&s2).Mul(&s2, &alpha)

	var zhZeta, zetaNPlusTwo, alphaSquareLagrangeZero, den blsfr.Element
	alphaSquareLagrangeZero.Set(&zeta).Exp(alphaSquareLagrangeZero, big.NewInt(int64(n)))
	zetaNPlusTwo.Mul(&alphaSquareLagrangeZero, &zeta).Mul(&zetaNPlusTwo, &zeta)
	one := blsfr.One()
	alphaSquareLagrangeZero.Sub(&alphaSquareLagrangeZero, &one)
	zhZeta.Set(&alphaSquareLagrangeZero)
	den.Sub(&zeta, &one).Inverse(&den)
	alphaSquareLagrangeZero.Mul(&alphaSquareLagrangeZero, &den).
		Mul(&alphaSquareLagrangeZero, &alpha).
		Mul(&alphaSquareLagrangeZero, &alpha).
		Mul(&alphaSquareLagrangeZero, &sizeInv)

	parallel.Execute(len(zBlinded), func(start, stop int) {
		var t, t0, t1 blsfr.Element
		for i := start; i < stop; i++ {
			t.Mul(&zBlinded[i], &s2)
			if i < len(fixed.s3) {
				t0.Mul(&fixed.s3[i], &s1)
				t.Add(&t, &t0)
			}
			if i < len(fixed.qm) {
				t1.Mul(&fixed.qm[i], &rl)
				t.Add(&t, &t1)
				t0.Mul(&fixed.ql[i], &lZeta)
				t.Add(&t, &t0)
				t0.Mul(&fixed.qr[i], &rZeta)
				t.Add(&t, &t0)
				t0.Mul(&fixed.qo[i], &oZeta)
				t.Add(&t, &t0)
				t.Add(&t, &fixed.qk[i])
				for j := range qcpZeta {
					t0.Mul(&pi2[j][i], &qcpZeta[j])
					t.Add(&t, &t0)
				}
			}
			t0.Mul(&zBlinded[i], &alphaSquareLagrangeZero)
			zBlinded[i].Add(&t, &t0)
			if i < len(h[2]) {
				t.Mul(&h[2][i], &zetaNPlusTwo).
					Add(&t, &h[1][i]).
					Mul(&t, &zetaNPlusTwo).
					Add(&t, &h[0][i]).
					Mul(&t, &zhZeta)
				zBlinded[i].Sub(&zBlinded[i], &t)
			}
		}
	})
	return zBlinded
}

func bls12377HRawToField(raw [3][]uint64) ([3][]blsfr.Element, error) {
	var out [3][]blsfr.Element
	for i := range raw {
		converted, err := bls12377RawToField(raw[i])
		if err != nil {
			return out, err
		}
		out[i] = converted
	}
	return out, nil
}

func bls12377RawMatrixToField(raw [][]uint64) ([][]blsfr.Element, error) {
	out := make([][]blsfr.Element, len(raw))
	for i := range raw {
		if len(raw[i]) == 0 {
			continue
		}
		converted, err := bls12377RawToField(raw[i])
		if err != nil {
			return nil, err
		}
		out[i] = converted
	}
	return out, nil
}

func bls12377ScalarFromRaw(raw []uint64) blsfr.Element {
	values, err := bls12377RawToField(raw)
	if err != nil {
		panic(err)
	}
	return values[0]
}

func evaluateBLS12377(canonical []blsfr.Element, z blsfr.Element) blsfr.Element {
	n := len(canonical)
	nCPU := runtime.GOMAXPROCS(0)
	if n < 4096 || nCPU < 2 {
		return evaluateBLS12377Serial(canonical, z)
	}

	chunkSize := (n + nCPU - 1) / nCPU
	partials := make([]blsfr.Element, nCPU)
	parallel.Execute(nCPU, func(start, stop int) {
		for c := start; c < stop; c++ {
			lo := c * chunkSize
			if lo >= n {
				continue
			}
			hi := lo + chunkSize
			if hi > n {
				hi = n
			}
			partials[c] = evaluateBLS12377Serial(canonical[lo:hi], z)
		}
	}, nCPU)

	zChunk := expBLS12377(z, chunkSize)
	var result, zPow blsfr.Element
	zPow.SetOne()
	for c := range nCPU {
		if c*chunkSize >= n {
			break
		}
		var t blsfr.Element
		t.Mul(&partials[c], &zPow)
		result.Add(&result, &t)
		zPow.Mul(&zPow, &zChunk)
	}
	return result
}

func evaluateBLS12377Serial(canonical []blsfr.Element, z blsfr.Element) blsfr.Element {
	var r blsfr.Element
	for i := len(canonical) - 1; i >= 0; i-- {
		r.Mul(&r, &z).Add(&r, &canonical[i])
	}
	return r
}

func expBLS12377(z blsfr.Element, exp int) blsfr.Element {
	var base, acc blsfr.Element
	base.Set(&z)
	acc.SetOne()
	for exp > 0 {
		if exp&1 != 0 {
			acc.Mul(&acc, &base)
		}
		base.Square(&base)
		exp >>= 1
	}
	return acc
}
