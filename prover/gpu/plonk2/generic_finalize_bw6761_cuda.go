//go:build cuda

package plonk2

import (
	"fmt"
	"hash"
	"math/big"

	bw6761 "github.com/consensys/gnark-crypto/ecc/bw6-761"
	bwfr "github.com/consensys/gnark-crypto/ecc/bw6-761/fr"
	bwkzg "github.com/consensys/gnark-crypto/ecc/bw6-761/kzg"
	fiatshamir "github.com/consensys/gnark-crypto/fiat-shamir"
	"github.com/consensys/gnark/backend"
	bwplonk "github.com/consensys/gnark/backend/plonk/bw6-761"
)

func (b *genericGPUBackend) finalizeBW6761(
	artifacts *genericProveArtifacts,
	proverConfig *backend.ProverConfig,
) error {
	proof := artifacts.proof.(*bwplonk.Proof)
	pk := b.pk.(*bwplonk.ProvingKey)

	lBlinded, err := bw6761RawToField(artifacts.lroBlinded[0])
	if err != nil {
		return err
	}
	rBlinded, err := bw6761RawToField(artifacts.lroBlinded[1])
	if err != nil {
		return err
	}
	oBlinded, err := bw6761RawToField(artifacts.lroBlinded[2])
	if err != nil {
		return err
	}
	zBlinded, err := bw6761RawToField(artifacts.zBlinded)
	if err != nil {
		return err
	}
	h, err := bw6761HRawToField(artifacts.hRaw)
	if err != nil {
		return err
	}
	pi2, err := bw6761RawMatrixToField(artifacts.bsb22Canonical)
	if err != nil {
		return err
	}
	fixed, err := b.bw6761FixedCanonical()
	if err != nil {
		return err
	}

	zeta := bw6761ScalarFromRaw(artifacts.zetaRaw)
	alpha := bw6761ScalarFromRaw(artifacts.alphaRaw)
	beta := bw6761ScalarFromRaw(artifacts.betaRaw)
	gamma := bw6761ScalarFromRaw(artifacts.gammaRaw)

	var shiftedZeta bwfr.Element
	shiftedZeta.Mul(&zeta, &pk.Vk.Generator)
	if err := b.openBW6761ShiftedZ(proof, zBlinded, shiftedZeta); err != nil {
		return err
	}

	lZeta := evaluateBW6761(lBlinded, zeta)
	rZeta := evaluateBW6761(rBlinded, zeta)
	oZeta := evaluateBW6761(oBlinded, zeta)
	s1Zeta := evaluateBW6761(fixed.s1, zeta)
	s2Zeta := evaluateBW6761(fixed.s2, zeta)
	qcpZeta := make([]bwfr.Element, len(fixed.qcp))
	for i := range fixed.qcp {
		qcpZeta[i] = evaluateBW6761(fixed.qcp[i], zeta)
	}

	linPol := computeBW6761LinearizedPolynomial(
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
		append([]bwfr.Element(nil), zBlinded...),
		pi2,
		h,
		fixed,
	)
	linDigestRaw, err := b.state.kzg.CommitRaw(genericRawBW6761Fr(linPol))
	if err != nil {
		return fmt.Errorf("commit linearized polynomial: %w", err)
	}
	linDigestAny, err := (bw6761ProofOps{}).rawCommitmentToDigest(linDigestRaw)
	if err != nil {
		return err
	}
	linDigest := linDigestAny.(bwkzg.Digest)

	polys := make([][]bwfr.Element, 6+len(fixed.qcp))
	polys[0] = linPol
	polys[1] = lBlinded
	polys[2] = rBlinded
	polys[3] = oBlinded
	polys[4] = fixed.s1
	polys[5] = fixed.s2
	copy(polys[6:], fixed.qcp)

	digests := make([]bw6761.G1Affine, len(polys))
	digests[0] = linDigest
	digests[1] = proof.LRO[0]
	digests[2] = proof.LRO[1]
	digests[3] = proof.LRO[2]
	digests[4] = pk.Vk.S[0]
	digests[5] = pk.Vk.S[1]
	copy(digests[6:], pk.Vk.Qcp)

	if err := b.batchOpenBW6761(
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

func (b *genericGPUBackend) openBW6761ShiftedZ(
	proof *bwplonk.Proof,
	zBlinded []bwfr.Element,
	point bwfr.Element,
) error {
	quotient := append([]bwfr.Element(nil), zBlinded...)
	hornerQuotientBW6761(quotient, point)
	proof.ZShiftedOpening.ClaimedValue = quotient[0]
	digest, err := b.commitBW6761Polynomial(quotient[1:])
	if err != nil {
		return fmt.Errorf("open shifted Z: %w", err)
	}
	proof.ZShiftedOpening.H = digest
	return nil
}

func (b *genericGPUBackend) batchOpenBW6761(
	proof *bwplonk.Proof,
	polys [][]bwfr.Element,
	digests []bw6761.G1Affine,
	point bwfr.Element,
	foldingHash hash.Hash,
	dataTranscript []byte,
) error {
	claimed := make([]bwfr.Element, len(polys))
	for i := range polys {
		claimed[i] = evaluateBW6761(polys[i], point)
	}
	proof.BatchedProof.ClaimedValues = claimed

	gamma, err := deriveBW6761BatchGamma(point, digests, claimed, foldingHash, dataTranscript)
	if err != nil {
		return fmt.Errorf("derive batch opening challenge: %w", err)
	}
	largest := 0
	for i := range polys {
		if len(polys[i]) > largest {
			largest = len(polys[i])
		}
	}
	folded := make([]bwfr.Element, largest)
	gammaPower := bwfr.One()
	for i := range polys {
		for j := range polys[i] {
			var term bwfr.Element
			term.Mul(&polys[i][j], &gammaPower)
			folded[j].Add(&folded[j], &term)
		}
		gammaPower.Mul(&gammaPower, &gamma)
	}
	var foldedEval bwfr.Element
	for i := len(claimed) - 1; i >= 0; i-- {
		foldedEval.Mul(&foldedEval, &gamma).Add(&foldedEval, &claimed[i])
	}
	folded[0].Sub(&folded[0], &foldedEval)
	hornerQuotientBW6761(folded, point)
	digest, err := b.commitBW6761Polynomial(folded[1:])
	if err != nil {
		return fmt.Errorf("batch opening: %w", err)
	}
	proof.BatchedProof.H = digest
	return nil
}

func (b *genericGPUBackend) commitBW6761Polynomial(poly []bwfr.Element) (bwkzg.Digest, error) {
	raw, err := b.state.kzg.CommitRaw(genericRawBW6761Fr(poly))
	if err != nil {
		return bwkzg.Digest{}, err
	}
	digestAny, err := (bw6761ProofOps{}).rawCommitmentToDigest(raw)
	if err != nil {
		return bwkzg.Digest{}, err
	}
	return digestAny.(bwkzg.Digest), nil
}

func deriveBW6761BatchGamma(
	point bwfr.Element,
	digests []bw6761.G1Affine,
	claimed []bwfr.Element,
	foldingHash hash.Hash,
	dataTranscript []byte,
) (bwfr.Element, error) {
	fs := fiatshamir.NewTranscript(foldingHash, "gamma")
	var gamma bwfr.Element
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

func hornerQuotientBW6761(poly []bwfr.Element, z bwfr.Element) {
	for i := len(poly) - 2; i >= 0; i-- {
		var tmp bwfr.Element
		tmp.Mul(&poly[i+1], &z)
		poly[i].Add(&poly[i], &tmp)
	}
}

type bw6761FixedCanonical struct {
	ql, qr, qm, qo, qk []bwfr.Element
	s1, s2, s3         []bwfr.Element
	qcp                [][]bwfr.Element
}

func (b *genericGPUBackend) bw6761FixedCanonical() (bw6761FixedCanonical, error) {
	convert := func(v *FrVector) ([]bwfr.Element, error) {
		raw, err := copyFrVectorRaw(v)
		if err != nil {
			return nil, err
		}
		return bw6761RawToField(raw)
	}
	ql, err := convert(b.state.fixed.ql)
	if err != nil {
		return bw6761FixedCanonical{}, err
	}
	qr, err := convert(b.state.fixed.qr)
	if err != nil {
		return bw6761FixedCanonical{}, err
	}
	qm, err := convert(b.state.fixed.qm)
	if err != nil {
		return bw6761FixedCanonical{}, err
	}
	qo, err := convert(b.state.fixed.qo)
	if err != nil {
		return bw6761FixedCanonical{}, err
	}
	qk, err := convert(b.state.fixed.qk)
	if err != nil {
		return bw6761FixedCanonical{}, err
	}
	s1, err := convert(b.state.fixed.s1)
	if err != nil {
		return bw6761FixedCanonical{}, err
	}
	s2, err := convert(b.state.fixed.s2)
	if err != nil {
		return bw6761FixedCanonical{}, err
	}
	s3, err := convert(b.state.fixed.s3)
	if err != nil {
		return bw6761FixedCanonical{}, err
	}
	qcp := make([][]bwfr.Element, len(b.state.qcp))
	for i := range b.state.qcp {
		qcp[i], err = convert(b.state.qcp[i])
		if err != nil {
			return bw6761FixedCanonical{}, err
		}
	}
	return bw6761FixedCanonical{
		ql: ql, qr: qr, qm: qm, qo: qo, qk: qk,
		s1: s1, s2: s2, s3: s3,
		qcp: qcp,
	}, nil
}

func computeBW6761LinearizedPolynomial(
	n uint64,
	cosetShift, sizeInv bwfr.Element,
	lZeta, rZeta, oZeta, alpha, beta, gamma, zeta, zu bwfr.Element,
	s1Zeta, s2Zeta bwfr.Element,
	qcpZeta []bwfr.Element,
	zBlinded []bwfr.Element,
	pi2 [][]bwfr.Element,
	h [3][]bwfr.Element,
	fixed bw6761FixedCanonical,
) []bwfr.Element {
	var rl bwfr.Element
	rl.Mul(&rZeta, &lZeta)

	var s1, tmp bwfr.Element
	s1.Mul(&s1Zeta, &beta).Add(&s1, &lZeta).Add(&s1, &gamma)
	tmp.Mul(&s2Zeta, &beta).Add(&tmp, &rZeta).Add(&tmp, &gamma)
	s1.Mul(&s1, &tmp).Mul(&s1, &zu).Mul(&s1, &beta).Mul(&s1, &alpha)

	var s2, uzeta, uuzeta bwfr.Element
	uzeta.Mul(&zeta, &cosetShift)
	uuzeta.Mul(&uzeta, &cosetShift)
	s2.Mul(&beta, &zeta).Add(&s2, &lZeta).Add(&s2, &gamma)
	tmp.Mul(&beta, &uzeta).Add(&tmp, &rZeta).Add(&tmp, &gamma)
	s2.Mul(&s2, &tmp)
	tmp.Mul(&beta, &uuzeta).Add(&tmp, &oZeta).Add(&tmp, &gamma)
	s2.Mul(&s2, &tmp).Neg(&s2).Mul(&s2, &alpha)

	var zhZeta, zetaNPlusTwo, alphaSquareLagrangeZero, den bwfr.Element
	alphaSquareLagrangeZero.Set(&zeta).Exp(alphaSquareLagrangeZero, big.NewInt(int64(n)))
	zetaNPlusTwo.Mul(&alphaSquareLagrangeZero, &zeta).Mul(&zetaNPlusTwo, &zeta)
	one := bwfr.One()
	alphaSquareLagrangeZero.Sub(&alphaSquareLagrangeZero, &one)
	zhZeta.Set(&alphaSquareLagrangeZero)
	den.Sub(&zeta, &one).Inverse(&den)
	alphaSquareLagrangeZero.Mul(&alphaSquareLagrangeZero, &den).
		Mul(&alphaSquareLagrangeZero, &alpha).
		Mul(&alphaSquareLagrangeZero, &alpha).
		Mul(&alphaSquareLagrangeZero, &sizeInv)

	var t, t0, t1 bwfr.Element
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

func bw6761HRawToField(raw [3][]uint64) ([3][]bwfr.Element, error) {
	var out [3][]bwfr.Element
	for i := range raw {
		converted, err := bw6761RawToField(raw[i])
		if err != nil {
			return out, err
		}
		out[i] = converted
	}
	return out, nil
}

func bw6761RawMatrixToField(raw [][]uint64) ([][]bwfr.Element, error) {
	out := make([][]bwfr.Element, len(raw))
	for i := range raw {
		if len(raw[i]) == 0 {
			continue
		}
		converted, err := bw6761RawToField(raw[i])
		if err != nil {
			return nil, err
		}
		out[i] = converted
	}
	return out, nil
}

func bw6761ScalarFromRaw(raw []uint64) bwfr.Element {
	values, err := bw6761RawToField(raw)
	if err != nil {
		panic(err)
	}
	return values[0]
}

func evaluateBW6761(canonical []bwfr.Element, z bwfr.Element) bwfr.Element {
	var r bwfr.Element
	for i := len(canonical) - 1; i >= 0; i-- {
		r.Mul(&r, &z).Add(&r, &canonical[i])
	}
	return r
}
