//go:build cuda

package plonk2

import (
	"fmt"
	"hash"
	"math/big"

	bn254 "github.com/consensys/gnark-crypto/ecc/bn254"
	bnfr "github.com/consensys/gnark-crypto/ecc/bn254/fr"
	bnkzg "github.com/consensys/gnark-crypto/ecc/bn254/kzg"
	fiatshamir "github.com/consensys/gnark-crypto/fiat-shamir"
	"github.com/consensys/gnark/backend"
	bnplonk "github.com/consensys/gnark/backend/plonk/bn254"
)

func (b *genericGPUBackend) finalizeBN254(
	artifacts *genericProveArtifacts,
	proverConfig *backend.ProverConfig,
) error {
	proof := artifacts.proof.(*bnplonk.Proof)
	pk := b.pk.(*bnplonk.ProvingKey)

	lBlinded, err := bn254RawToField(artifacts.lroBlinded[0])
	if err != nil {
		return err
	}
	rBlinded, err := bn254RawToField(artifacts.lroBlinded[1])
	if err != nil {
		return err
	}
	oBlinded, err := bn254RawToField(artifacts.lroBlinded[2])
	if err != nil {
		return err
	}
	zBlinded, err := bn254RawToField(artifacts.zBlinded)
	if err != nil {
		return err
	}
	h, err := bn254HRawToField(artifacts.hRaw)
	if err != nil {
		return err
	}
	pi2, err := bn254RawMatrixToField(artifacts.bsb22Canonical)
	if err != nil {
		return err
	}
	fixed, err := b.bn254FixedCanonical()
	if err != nil {
		return err
	}

	zeta := bn254ScalarFromRaw(artifacts.zetaRaw)
	alpha := bn254ScalarFromRaw(artifacts.alphaRaw)
	beta := bn254ScalarFromRaw(artifacts.betaRaw)
	gamma := bn254ScalarFromRaw(artifacts.gammaRaw)

	var shiftedZeta bnfr.Element
	shiftedZeta.Mul(&zeta, &pk.Vk.Generator)
	if err := b.openBN254ShiftedZ(proof, zBlinded, shiftedZeta); err != nil {
		return err
	}

	lZeta := evaluateBN254(lBlinded, zeta)
	rZeta := evaluateBN254(rBlinded, zeta)
	oZeta := evaluateBN254(oBlinded, zeta)
	s1Zeta := evaluateBN254(fixed.s1, zeta)
	s2Zeta := evaluateBN254(fixed.s2, zeta)
	qcpZeta := make([]bnfr.Element, len(fixed.qcp))
	for i := range fixed.qcp {
		qcpZeta[i] = evaluateBN254(fixed.qcp[i], zeta)
	}

	linPol := computeBN254LinearizedPolynomial(
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
		append([]bnfr.Element(nil), zBlinded...),
		pi2,
		h,
		fixed,
	)
	linDigestRaw, err := b.state.kzg.CommitRaw(genericRawBN254Fr(linPol))
	if err != nil {
		return fmt.Errorf("commit linearized polynomial: %w", err)
	}
	linDigestAny, err := (bn254ProofOps{}).rawCommitmentToDigest(linDigestRaw)
	if err != nil {
		return err
	}
	linDigest := linDigestAny.(bnkzg.Digest)

	polys := make([][]bnfr.Element, 6+len(fixed.qcp))
	polys[0] = linPol
	polys[1] = lBlinded
	polys[2] = rBlinded
	polys[3] = oBlinded
	polys[4] = fixed.s1
	polys[5] = fixed.s2
	copy(polys[6:], fixed.qcp)

	digests := make([]bn254.G1Affine, len(polys))
	digests[0] = linDigest
	digests[1] = proof.LRO[0]
	digests[2] = proof.LRO[1]
	digests[3] = proof.LRO[2]
	digests[4] = pk.Vk.S[0]
	digests[5] = pk.Vk.S[1]
	copy(digests[6:], pk.Vk.Qcp)

	if err := b.batchOpenBN254(
		proof,
		polys,
		digests,
		zeta,
		proverConfig.KZGFoldingHash,
		proof.ZShiftedOpening.ClaimedValue.Marshal(),
	); err != nil {
		return err
	}
	return nil
}

func (b *genericGPUBackend) openBN254ShiftedZ(
	proof *bnplonk.Proof,
	zBlinded []bnfr.Element,
	point bnfr.Element,
) error {
	quotient := append([]bnfr.Element(nil), zBlinded...)
	hornerQuotientBN254(quotient, point)
	proof.ZShiftedOpening.ClaimedValue = quotient[0]
	digest, err := b.commitBN254Polynomial(quotient[1:])
	if err != nil {
		return fmt.Errorf("open shifted Z: %w", err)
	}
	proof.ZShiftedOpening.H = digest
	return nil
}

func (b *genericGPUBackend) batchOpenBN254(
	proof *bnplonk.Proof,
	polys [][]bnfr.Element,
	digests []bn254.G1Affine,
	point bnfr.Element,
	foldingHash hash.Hash,
	dataTranscript []byte,
) error {
	claimed := make([]bnfr.Element, len(polys))
	for i := range polys {
		claimed[i] = evaluateBN254(polys[i], point)
	}
	proof.BatchedProof.ClaimedValues = claimed

	gamma, err := deriveBN254BatchGamma(point, digests, claimed, foldingHash, dataTranscript)
	if err != nil {
		return fmt.Errorf("derive batch opening challenge: %w", err)
	}
	largest := 0
	for i := range polys {
		if len(polys[i]) > largest {
			largest = len(polys[i])
		}
	}
	folded := make([]bnfr.Element, largest)
	gammaPower := bnfr.One()
	for i := range polys {
		for j := range polys[i] {
			var term bnfr.Element
			term.Mul(&polys[i][j], &gammaPower)
			folded[j].Add(&folded[j], &term)
		}
		gammaPower.Mul(&gammaPower, &gamma)
	}
	var foldedEval bnfr.Element
	for i := len(claimed) - 1; i >= 0; i-- {
		foldedEval.Mul(&foldedEval, &gamma).Add(&foldedEval, &claimed[i])
	}
	folded[0].Sub(&folded[0], &foldedEval)
	hornerQuotientBN254(folded, point)
	digest, err := b.commitBN254Polynomial(folded[1:])
	if err != nil {
		return fmt.Errorf("batch opening: %w", err)
	}
	proof.BatchedProof.H = digest
	return nil
}

func (b *genericGPUBackend) commitBN254Polynomial(poly []bnfr.Element) (bnkzg.Digest, error) {
	raw, err := b.state.kzg.CommitRaw(genericRawBN254Fr(poly))
	if err != nil {
		return bnkzg.Digest{}, err
	}
	digestAny, err := (bn254ProofOps{}).rawCommitmentToDigest(raw)
	if err != nil {
		return bnkzg.Digest{}, err
	}
	return digestAny.(bnkzg.Digest), nil
}

func deriveBN254BatchGamma(
	point bnfr.Element,
	digests []bn254.G1Affine,
	claimed []bnfr.Element,
	foldingHash hash.Hash,
	dataTranscript []byte,
) (bnfr.Element, error) {
	fs := fiatshamir.NewTranscript(foldingHash, "gamma")
	var gamma bnfr.Element
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

func hornerQuotientBN254(poly []bnfr.Element, z bnfr.Element) {
	for i := len(poly) - 2; i >= 0; i-- {
		var tmp bnfr.Element
		tmp.Mul(&poly[i+1], &z)
		poly[i].Add(&poly[i], &tmp)
	}
}

type bn254FixedCanonical struct {
	ql, qr, qm, qo, qk []bnfr.Element
	s1, s2, s3         []bnfr.Element
	qcp                [][]bnfr.Element
}

func (b *genericGPUBackend) bn254FixedCanonical() (bn254FixedCanonical, error) {
	convert := func(v *FrVector) ([]bnfr.Element, error) {
		raw, err := copyFrVectorRaw(v)
		if err != nil {
			return nil, err
		}
		return bn254RawToField(raw)
	}
	ql, err := convert(b.state.fixed.ql)
	if err != nil {
		return bn254FixedCanonical{}, err
	}
	qr, err := convert(b.state.fixed.qr)
	if err != nil {
		return bn254FixedCanonical{}, err
	}
	qm, err := convert(b.state.fixed.qm)
	if err != nil {
		return bn254FixedCanonical{}, err
	}
	qo, err := convert(b.state.fixed.qo)
	if err != nil {
		return bn254FixedCanonical{}, err
	}
	qk, err := convert(b.state.fixed.qk)
	if err != nil {
		return bn254FixedCanonical{}, err
	}
	s1, err := convert(b.state.fixed.s1)
	if err != nil {
		return bn254FixedCanonical{}, err
	}
	s2, err := convert(b.state.fixed.s2)
	if err != nil {
		return bn254FixedCanonical{}, err
	}
	s3, err := convert(b.state.fixed.s3)
	if err != nil {
		return bn254FixedCanonical{}, err
	}
	qcp := make([][]bnfr.Element, len(b.state.qcp))
	for i := range b.state.qcp {
		qcp[i], err = convert(b.state.qcp[i])
		if err != nil {
			return bn254FixedCanonical{}, err
		}
	}
	return bn254FixedCanonical{
		ql: ql, qr: qr, qm: qm, qo: qo, qk: qk,
		s1: s1, s2: s2, s3: s3,
		qcp: qcp,
	}, nil
}

func computeBN254LinearizedPolynomial(
	n uint64,
	cosetShift, sizeInv bnfr.Element,
	lZeta, rZeta, oZeta, alpha, beta, gamma, zeta, zu bnfr.Element,
	s1Zeta, s2Zeta bnfr.Element,
	qcpZeta []bnfr.Element,
	zBlinded []bnfr.Element,
	pi2 [][]bnfr.Element,
	h [3][]bnfr.Element,
	fixed bn254FixedCanonical,
) []bnfr.Element {
	var rl bnfr.Element
	rl.Mul(&rZeta, &lZeta)

	var s1, tmp bnfr.Element
	s1.Mul(&s1Zeta, &beta).Add(&s1, &lZeta).Add(&s1, &gamma)
	tmp.Mul(&s2Zeta, &beta).Add(&tmp, &rZeta).Add(&tmp, &gamma)
	s1.Mul(&s1, &tmp).Mul(&s1, &zu).Mul(&s1, &beta).Mul(&s1, &alpha)

	var s2, uzeta, uuzeta bnfr.Element
	uzeta.Mul(&zeta, &cosetShift)
	uuzeta.Mul(&uzeta, &cosetShift)
	s2.Mul(&beta, &zeta).Add(&s2, &lZeta).Add(&s2, &gamma)
	tmp.Mul(&beta, &uzeta).Add(&tmp, &rZeta).Add(&tmp, &gamma)
	s2.Mul(&s2, &tmp)
	tmp.Mul(&beta, &uuzeta).Add(&tmp, &oZeta).Add(&tmp, &gamma)
	s2.Mul(&s2, &tmp).Neg(&s2).Mul(&s2, &alpha)

	var zhZeta, zetaNPlusTwo, alphaSquareLagrangeZero, den bnfr.Element
	alphaSquareLagrangeZero.Set(&zeta).Exp(alphaSquareLagrangeZero, big.NewInt(int64(n)))
	zetaNPlusTwo.Mul(&alphaSquareLagrangeZero, &zeta).Mul(&zetaNPlusTwo, &zeta)
	one := bnfr.One()
	alphaSquareLagrangeZero.Sub(&alphaSquareLagrangeZero, &one)
	zhZeta.Set(&alphaSquareLagrangeZero)
	den.Sub(&zeta, &one).Inverse(&den)
	alphaSquareLagrangeZero.Mul(&alphaSquareLagrangeZero, &den).
		Mul(&alphaSquareLagrangeZero, &alpha).
		Mul(&alphaSquareLagrangeZero, &alpha).
		Mul(&alphaSquareLagrangeZero, &sizeInv)

	var t, t0, t1 bnfr.Element
	for i := range zBlinded {
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
	return zBlinded
}

func bn254HRawToField(raw [3][]uint64) ([3][]bnfr.Element, error) {
	var out [3][]bnfr.Element
	for i := range raw {
		converted, err := bn254RawToField(raw[i])
		if err != nil {
			return out, err
		}
		out[i] = converted
	}
	return out, nil
}

func bn254RawMatrixToField(raw [][]uint64) ([][]bnfr.Element, error) {
	out := make([][]bnfr.Element, len(raw))
	for i := range raw {
		if len(raw[i]) == 0 {
			continue
		}
		converted, err := bn254RawToField(raw[i])
		if err != nil {
			return nil, err
		}
		out[i] = converted
	}
	return out, nil
}

func bn254ScalarFromRaw(raw []uint64) bnfr.Element {
	values, err := bn254RawToField(raw)
	if err != nil {
		panic(err)
	}
	return values[0]
}

func evaluateBN254(canonical []bnfr.Element, z bnfr.Element) bnfr.Element {
	var r bnfr.Element
	for i := len(canonical) - 1; i >= 0; i-- {
		r.Mul(&r, &z).Add(&r, &canonical[i])
	}
	return r
}
