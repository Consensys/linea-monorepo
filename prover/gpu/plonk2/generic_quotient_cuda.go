//go:build cuda

package plonk2

import (
	"fmt"
	"math/big"

	blsfr "github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	blsfft "github.com/consensys/gnark-crypto/ecc/bls12-377/fr/fft"
	bnfr "github.com/consensys/gnark-crypto/ecc/bn254/fr"
	bnfft "github.com/consensys/gnark-crypto/ecc/bn254/fr/fft"
	bwfr "github.com/consensys/gnark-crypto/ecc/bw6-761/fr"
	bwfft "github.com/consensys/gnark-crypto/ecc/bw6-761/fr/fft"
	"github.com/consensys/gnark/backend"
	"github.com/consensys/gnark/backend/witness"
)

type genericQuotientScalars struct {
	cosets       [4]genericCosetScalars
	omega4Inv    []uint64
	quarter      []uint64
	uInvN        [3][]uint64
	cosetShift   []uint64
	cosetShiftSq []uint64
}

type genericCosetScalars struct {
	gen    []uint64
	genInv []uint64
	powN   []uint64
	zhInv  []uint64
	l1     []uint64
}

func (b *genericGPUBackend) computeAndCommitQuotient(
	artifacts *genericProveArtifacts,
	ops curveProofOps,
	fullWitness witness.Witness,
	proverConfig *backend.ProverConfig,
) error {
	if b.state.n < 6 {
		return errGPUProverNotWired
	}
	if err := b.state.releaseCanonicalCommitmentWave(); err != nil {
		return fmt.Errorf("release canonical MSM scratch before quotient: %w", err)
	}
	if err := b.state.releaseLagrangeCommitmentWave(); err != nil {
		return fmt.Errorf("release lagrange MSM scratch before quotient: %w", err)
	}

	hRaw, err := b.computeQuotientRaw(artifacts)
	if err != nil {
		return err
	}
	artifacts.hRaw = hRaw

	if err := b.state.pinCanonicalCommitmentWave(); err != nil {
		return err
	}
	defer b.state.releaseCanonicalCommitmentWave()

	commits, err := b.state.commitCanonicalWaveRaw([][]uint64{hRaw[0], hRaw[1], hRaw[2]})
	if err != nil {
		return fmt.Errorf("commit quotient H: %w", err)
	}
	for i := range commits {
		digest, err := ops.rawCommitmentToDigest(commits[i])
		if err != nil {
			return err
		}
		if err := ops.setH(artifacts.proof, i, digest); err != nil {
			return err
		}
	}
	zetaRaw, err := b.deriveZetaRaw(artifacts, fullWitness, proverConfig)
	if err != nil {
		return err
	}
	artifacts.zetaRaw = zetaRaw
	return nil
}

func (b *genericGPUBackend) computeQuotientRaw(
	artifacts *genericProveArtifacts,
) ([3][]uint64, error) {
	var out [3][]uint64
	n := b.state.n
	limbs := scalarLimbs(b.state.curve)
	scalars, err := newGenericQuotientScalars(b.state.curve, n)
	if err != nil {
		return out, err
	}

	lSrc, err := newFrVectorFromRaw(b.state, rawElementSlice(artifacts.lroBlinded[0], limbs, 0, n))
	if err != nil {
		return out, err
	}
	defer lSrc.Free()
	rSrc, err := newFrVectorFromRaw(b.state, rawElementSlice(artifacts.lroBlinded[1], limbs, 0, n))
	if err != nil {
		return out, err
	}
	defer rSrc.Free()
	oSrc, err := newFrVectorFromRaw(b.state, rawElementSlice(artifacts.lroBlinded[2], limbs, 0, n))
	if err != nil {
		return out, err
	}
	defer oSrc.Free()
	zSrc, err := newFrVectorFromRaw(b.state, rawElementSlice(artifacts.zBlinded, limbs, 0, n))
	if err != nil {
		return out, err
	}
	defer zSrc.Free()
	qkSrc, err := newFrVectorFromRaw(b.state, artifacts.qkCanonical)
	if err != nil {
		return out, err
	}
	defer qkSrc.Free()

	pi2Src := make([]*FrVector, len(artifacts.bsb22Canonical))
	for i := range artifacts.bsb22Canonical {
		if len(artifacts.bsb22Canonical[i]) == 0 {
			continue
		}
		pi2Src[i], err = newFrVectorFromRaw(b.state, artifacts.bsb22Canonical[i])
		if err != nil {
			for _, v := range pi2Src {
				if v != nil {
					v.Free()
				}
			}
			return out, fmt.Errorf("upload BSB22 canonical[%d]: %w", i, err)
		}
	}
	defer func() {
		for _, v := range pi2Src {
			if v != nil {
				v.Free()
			}
		}
	}()

	l, r, o, z, s1, s2, s3, work, result, err := newQuotientVectors(b.state)
	if err != nil {
		return out, err
	}
	defer func() {
		for _, v := range []*FrVector{l, r, o, z, s1, s2, s3, work, result} {
			if v != nil {
				v.Free()
			}
		}
	}()

	hFull := make([]uint64, 4*n*limbs)
	if artifacts.scratch != nil {
		hFull = artifacts.scratch.hFull
	}
	var cosetBlocks [3]*FrVector
	cosetResultsOnDevice := true
	for i := range cosetBlocks {
		block, allocErr := NewFrVector(b.state.dev, b.state.curve, b.state.n)
		if allocErr != nil {
			cosetResultsOnDevice = false
			for _, v := range cosetBlocks {
				if v != nil {
					v.Free()
				}
			}
			cosetBlocks = [3]*FrVector{}
			break
		}
		cosetBlocks[i] = block
	}
	defer func() {
		for _, v := range cosetBlocks {
			if v != nil {
				v.Free()
			}
		}
	}()

	for k := range scalars.cosets {
		coset := scalars.cosets[k]
		if err := ReduceBlindedCoset(l, lSrc, rawElementTail(artifacts.lroBlinded[0], limbs, n), coset.powN); err != nil {
			return out, fmt.Errorf("reduce L coset %d: %w", k, err)
		}
		if err := ReduceBlindedCoset(r, rSrc, rawElementTail(artifacts.lroBlinded[1], limbs, n), coset.powN); err != nil {
			return out, fmt.Errorf("reduce R coset %d: %w", k, err)
		}
		if err := ReduceBlindedCoset(o, oSrc, rawElementTail(artifacts.lroBlinded[2], limbs, n), coset.powN); err != nil {
			return out, fmt.Errorf("reduce O coset %d: %w", k, err)
		}
		if err := ReduceBlindedCoset(z, zSrc, rawElementTail(artifacts.zBlinded, limbs, n), coset.powN); err != nil {
			return out, fmt.Errorf("reduce Z coset %d: %w", k, err)
		}
		for _, v := range []*FrVector{l, r, o, z} {
			if err := b.state.fft.CosetFFT(v, coset.gen); err != nil {
				return out, fmt.Errorf("wire coset FFT %d: %w", k, err)
			}
		}

		if err := copyAndCosetFFT(b.state, s1, b.state.fixed.s1, coset.gen); err != nil {
			return out, fmt.Errorf("S1 coset %d: %w", k, err)
		}
		if err := copyAndCosetFFT(b.state, s2, b.state.fixed.s2, coset.gen); err != nil {
			return out, fmt.Errorf("S2 coset %d: %w", k, err)
		}
		if err := copyAndCosetFFT(b.state, s3, b.state.fixed.s3, coset.gen); err != nil {
			return out, fmt.Errorf("S3 coset %d: %w", k, err)
		}
		if err := ComputeL1Den(work, b.state.fft, coset.gen); err != nil {
			return out, fmt.Errorf("L1 denominator coset %d: %w", k, err)
		}
		if err := work.BatchInvert(result); err != nil {
			return out, fmt.Errorf("L1 denominator invert coset %d: %w", k, err)
		}
		if err := PlonkPermBoundary(
			result, l, r, o, z, s1, s2, s3, work, b.state.fft,
			artifacts.alphaRaw, artifacts.betaRaw, artifacts.gammaRaw,
			coset.l1, scalars.cosetShift, scalars.cosetShiftSq, coset.gen,
		); err != nil {
			return out, fmt.Errorf("permutation boundary coset %d: %w", k, err)
		}

		if err := copyAndCosetFFT(b.state, z, b.state.fixed.ql, coset.gen); err != nil {
			return out, fmt.Errorf("Ql coset %d: %w", k, err)
		}
		if err := copyAndCosetFFT(b.state, s1, b.state.fixed.qr, coset.gen); err != nil {
			return out, fmt.Errorf("Qr coset %d: %w", k, err)
		}
		if err := copyAndCosetFFT(b.state, s2, b.state.fixed.qm, coset.gen); err != nil {
			return out, fmt.Errorf("Qm coset %d: %w", k, err)
		}
		if err := copyAndCosetFFT(b.state, s3, b.state.fixed.qo, coset.gen); err != nil {
			return out, fmt.Errorf("Qo coset %d: %w", k, err)
		}
		if err := copyAndCosetFFT(b.state, work, qkSrc, coset.gen); err != nil {
			return out, fmt.Errorf("Qk coset %d: %w", k, err)
		}
		if err := PlonkGateAccum(result, z, s1, s2, s3, work, l, r, o, coset.zhInv); err != nil {
			return out, fmt.Errorf("gate accumulation coset %d: %w", k, err)
		}

		for j := range pi2Src {
			if pi2Src[j] == nil {
				continue
			}
			if err := copyAndCosetFFT(b.state, z, b.state.qcp[j], coset.gen); err != nil {
				return out, fmt.Errorf("Qcp[%d] coset %d: %w", j, k, err)
			}
			if err := copyAndCosetFFT(b.state, work, pi2Src[j], coset.gen); err != nil {
				return out, fmt.Errorf("BSB22[%d] coset %d: %w", j, k, err)
			}
			if err := z.Mul(z, work); err != nil {
				return out, fmt.Errorf("BSB22 product[%d] coset %d: %w", j, k, err)
			}
			if err := result.AddScalarMulRaw(z, coset.zhInv); err != nil {
				return out, fmt.Errorf("BSB22 accumulation[%d] coset %d: %w", j, k, err)
			}
		}

		if cosetResultsOnDevice {
			if k < len(cosetBlocks) {
				if err := cosetBlocks[k].CopyFromDevice(result); err != nil {
					return out, fmt.Errorf("store quotient coset %d on device: %w", k, err)
				}
			}
		} else {
			chunk := rawElementSlice(hFull, limbs, k*n, (k+1)*n)
			if err := result.CopyToHostRaw(chunk); err != nil {
				return out, fmt.Errorf("copy quotient coset %d: %w", k, err)
			}
		}
	}

	blocks := [4]*FrVector{l, r, o, z}
	if cosetResultsOnDevice {
		blocks = [4]*FrVector{cosetBlocks[0], cosetBlocks[1], cosetBlocks[2], result}
	}
	for k := range blocks {
		if !cosetResultsOnDevice {
			chunk := rawElementSlice(hFull, limbs, k*n, (k+1)*n)
			if err := blocks[k].CopyFromHostRaw(chunk); err != nil {
				return out, fmt.Errorf("upload quotient block %d: %w", k, err)
			}
		}
		if err := b.state.fft.CosetFFTInverse(blocks[k], scalars.cosets[k].genInv); err != nil {
			return out, fmt.Errorf("inverse quotient coset %d: %w", k, err)
		}
	}
	if err := Butterfly4Inverse(blocks[0], blocks[1], blocks[2], blocks[3], scalars.omega4Inv, scalars.quarter); err != nil {
		return out, fmt.Errorf("quotient butterfly: %w", err)
	}
	for k := 1; k < len(blocks); k++ {
		if err := blocks[k].ScalarMulRaw(scalars.uInvN[k-1]); err != nil {
			return out, fmt.Errorf("scale quotient block %d: %w", k, err)
		}
	}
	for k := range blocks {
		chunk := rawElementSlice(hFull, limbs, k*n, (k+1)*n)
		if err := blocks[k].CopyToHostRaw(chunk); err != nil {
			return out, fmt.Errorf("copy quotient block %d: %w", k, err)
		}
	}

	shard := n + 2
	if artifacts.scratch != nil {
		out[0] = artifacts.scratch.hShardRaw(0)
		out[1] = artifacts.scratch.hShardRaw(1)
		out[2] = artifacts.scratch.hShardRaw(2)
	} else {
		out[0] = append([]uint64(nil), rawElementSlice(hFull, limbs, 0, shard)...)
		out[1] = append([]uint64(nil), rawElementSlice(hFull, limbs, shard, 2*shard)...)
		out[2] = append([]uint64(nil), rawElementSlice(hFull, limbs, 2*shard, 3*shard)...)
	}
	return out, nil
}

func newQuotientVectors(state *genericProverState) (*FrVector, *FrVector, *FrVector, *FrVector, *FrVector, *FrVector, *FrVector, *FrVector, *FrVector, error) {
	alloc := func() (*FrVector, error) {
		return NewFrVector(state.dev, state.curve, state.n)
	}
	l, err := alloc()
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, nil, nil, err
	}
	r, err := alloc()
	if err != nil {
		l.Free()
		return nil, nil, nil, nil, nil, nil, nil, nil, nil, err
	}
	o, err := alloc()
	if err != nil {
		l.Free()
		r.Free()
		return nil, nil, nil, nil, nil, nil, nil, nil, nil, err
	}
	z, err := alloc()
	if err != nil {
		l.Free()
		r.Free()
		o.Free()
		return nil, nil, nil, nil, nil, nil, nil, nil, nil, err
	}
	s1, err := alloc()
	if err != nil {
		l.Free()
		r.Free()
		o.Free()
		z.Free()
		return nil, nil, nil, nil, nil, nil, nil, nil, nil, err
	}
	s2, err := alloc()
	if err != nil {
		l.Free()
		r.Free()
		o.Free()
		z.Free()
		s1.Free()
		return nil, nil, nil, nil, nil, nil, nil, nil, nil, err
	}
	s3, err := alloc()
	if err != nil {
		l.Free()
		r.Free()
		o.Free()
		z.Free()
		s1.Free()
		s2.Free()
		return nil, nil, nil, nil, nil, nil, nil, nil, nil, err
	}
	work, err := alloc()
	if err != nil {
		l.Free()
		r.Free()
		o.Free()
		z.Free()
		s1.Free()
		s2.Free()
		s3.Free()
		return nil, nil, nil, nil, nil, nil, nil, nil, nil, err
	}
	result, err := alloc()
	if err != nil {
		l.Free()
		r.Free()
		o.Free()
		z.Free()
		s1.Free()
		s2.Free()
		s3.Free()
		work.Free()
		return nil, nil, nil, nil, nil, nil, nil, nil, nil, err
	}
	return l, r, o, z, s1, s2, s3, work, result, nil
}

func copyAndCosetFFT(state *genericProverState, dst, src *FrVector, gen []uint64) error {
	if err := dst.CopyFromDevice(src); err != nil {
		return err
	}
	return state.fft.CosetFFT(dst, gen)
}

func newFrVectorFromRaw(state *genericProverState, raw []uint64) (*FrVector, error) {
	v, err := NewFrVector(state.dev, state.curve, state.n)
	if err != nil {
		return nil, err
	}
	if err := v.CopyFromHostRaw(raw); err != nil {
		v.Free()
		return nil, err
	}
	return v, nil
}

func copyFrVectorRaw(v *FrVector) ([]uint64, error) {
	out := make([]uint64, v.RawWords())
	if err := v.CopyToHostRaw(out); err != nil {
		return nil, err
	}
	return out, nil
}

func rawElementSlice(raw []uint64, limbs, start, end int) []uint64 {
	return raw[start*limbs : end*limbs]
}

func rawElementTail(raw []uint64, limbs, start int) []uint64 {
	return raw[start*limbs:]
}

func scalarLimbs(curve Curve) int {
	switch curve {
	case CurveBN254, CurveBLS12377:
		return 4
	case CurveBW6761:
		return 6
	default:
		panic(fmt.Sprintf("unsupported curve %s", curve))
	}
}

func newGenericQuotientScalars(curve Curve, n int) (genericQuotientScalars, error) {
	switch curve {
	case CurveBN254:
		return newBN254QuotientScalars(n), nil
	case CurveBLS12377:
		return newBLS12377QuotientScalars(n), nil
	case CurveBW6761:
		return newBW6761QuotientScalars(n), nil
	default:
		return genericQuotientScalars{}, fmt.Errorf("plonk2: unsupported curve %s", curve)
	}
}

func newBN254QuotientScalars(n int) genericQuotientScalars {
	var out genericQuotientScalars
	domain0 := bnfft.NewDomain(uint64(n), bnfft.WithoutPrecompute())
	domain1 := bnfft.NewDomain(uint64(4*n), bnfft.WithoutPrecompute())
	bn := big.NewInt(int64(n))
	one := bnfr.One()
	u, g := domain1.FrMultiplicativeGen, domain1.Generator
	coset := u
	for i := range out.cosets {
		if i > 0 {
			coset.Mul(&coset, &g)
		}
		var pow, zh, inv, l1, genInv bnfr.Element
		pow.Exp(coset, bn)
		zh.Sub(&pow, &one)
		inv.Inverse(&zh)
		l1.Mul(&zh, &domain0.CardinalityInv)
		genInv.Inverse(&coset)
		out.cosets[i] = genericCosetScalars{
			gen:    genericRawBN254Fr([]bnfr.Element{coset}),
			genInv: genericRawBN254Fr([]bnfr.Element{genInv}),
			powN:   genericRawBN254Fr([]bnfr.Element{pow}),
			zhInv:  genericRawBN254Fr([]bnfr.Element{inv}),
			l1:     genericRawBN254Fr([]bnfr.Element{l1}),
		}
	}
	var omega4, omega4Inv, quarter, uN, uInvN, uInv2N, uInv3N, shiftSq bnfr.Element
	omega4.Exp(g, bn)
	omega4Inv.Inverse(&omega4)
	quarter.SetUint64(4)
	quarter.Inverse(&quarter)
	uN.Exp(u, bn)
	uInvN.Inverse(&uN)
	uInv2N.Mul(&uInvN, &uInvN)
	uInv3N.Mul(&uInv2N, &uInvN)
	shiftSq.Mul(&domain0.FrMultiplicativeGen, &domain0.FrMultiplicativeGen)
	out.omega4Inv = genericRawBN254Fr([]bnfr.Element{omega4Inv})
	out.quarter = genericRawBN254Fr([]bnfr.Element{quarter})
	out.uInvN = [3][]uint64{
		genericRawBN254Fr([]bnfr.Element{uInvN}),
		genericRawBN254Fr([]bnfr.Element{uInv2N}),
		genericRawBN254Fr([]bnfr.Element{uInv3N}),
	}
	out.cosetShift = genericRawBN254Fr([]bnfr.Element{domain0.FrMultiplicativeGen})
	out.cosetShiftSq = genericRawBN254Fr([]bnfr.Element{shiftSq})
	return out
}

func newBLS12377QuotientScalars(n int) genericQuotientScalars {
	var out genericQuotientScalars
	domain0 := blsfft.NewDomain(uint64(n), blsfft.WithoutPrecompute())
	domain1 := blsfft.NewDomain(uint64(4*n), blsfft.WithoutPrecompute())
	bn := big.NewInt(int64(n))
	one := blsfr.One()
	u, g := domain1.FrMultiplicativeGen, domain1.Generator
	coset := u
	for i := range out.cosets {
		if i > 0 {
			coset.Mul(&coset, &g)
		}
		var pow, zh, inv, l1, genInv blsfr.Element
		pow.Exp(coset, bn)
		zh.Sub(&pow, &one)
		inv.Inverse(&zh)
		l1.Mul(&zh, &domain0.CardinalityInv)
		genInv.Inverse(&coset)
		out.cosets[i] = genericCosetScalars{
			gen:    genericRawBLS12377Fr([]blsfr.Element{coset}),
			genInv: genericRawBLS12377Fr([]blsfr.Element{genInv}),
			powN:   genericRawBLS12377Fr([]blsfr.Element{pow}),
			zhInv:  genericRawBLS12377Fr([]blsfr.Element{inv}),
			l1:     genericRawBLS12377Fr([]blsfr.Element{l1}),
		}
	}
	var omega4, omega4Inv, quarter, uN, uInvN, uInv2N, uInv3N, shiftSq blsfr.Element
	omega4.Exp(g, bn)
	omega4Inv.Inverse(&omega4)
	quarter.SetUint64(4)
	quarter.Inverse(&quarter)
	uN.Exp(u, bn)
	uInvN.Inverse(&uN)
	uInv2N.Mul(&uInvN, &uInvN)
	uInv3N.Mul(&uInv2N, &uInvN)
	shiftSq.Mul(&domain0.FrMultiplicativeGen, &domain0.FrMultiplicativeGen)
	out.omega4Inv = genericRawBLS12377Fr([]blsfr.Element{omega4Inv})
	out.quarter = genericRawBLS12377Fr([]blsfr.Element{quarter})
	out.uInvN = [3][]uint64{
		genericRawBLS12377Fr([]blsfr.Element{uInvN}),
		genericRawBLS12377Fr([]blsfr.Element{uInv2N}),
		genericRawBLS12377Fr([]blsfr.Element{uInv3N}),
	}
	out.cosetShift = genericRawBLS12377Fr([]blsfr.Element{domain0.FrMultiplicativeGen})
	out.cosetShiftSq = genericRawBLS12377Fr([]blsfr.Element{shiftSq})
	return out
}

func newBW6761QuotientScalars(n int) genericQuotientScalars {
	var out genericQuotientScalars
	domain0 := bwfft.NewDomain(uint64(n), bwfft.WithoutPrecompute())
	domain1 := bwfft.NewDomain(uint64(4*n), bwfft.WithoutPrecompute())
	bn := big.NewInt(int64(n))
	one := bwfr.One()
	u, g := domain1.FrMultiplicativeGen, domain1.Generator
	coset := u
	for i := range out.cosets {
		if i > 0 {
			coset.Mul(&coset, &g)
		}
		var pow, zh, inv, l1, genInv bwfr.Element
		pow.Exp(coset, bn)
		zh.Sub(&pow, &one)
		inv.Inverse(&zh)
		l1.Mul(&zh, &domain0.CardinalityInv)
		genInv.Inverse(&coset)
		out.cosets[i] = genericCosetScalars{
			gen:    genericRawBW6761Fr([]bwfr.Element{coset}),
			genInv: genericRawBW6761Fr([]bwfr.Element{genInv}),
			powN:   genericRawBW6761Fr([]bwfr.Element{pow}),
			zhInv:  genericRawBW6761Fr([]bwfr.Element{inv}),
			l1:     genericRawBW6761Fr([]bwfr.Element{l1}),
		}
	}
	var omega4, omega4Inv, quarter, uN, uInvN, uInv2N, uInv3N, shiftSq bwfr.Element
	omega4.Exp(g, bn)
	omega4Inv.Inverse(&omega4)
	quarter.SetUint64(4)
	quarter.Inverse(&quarter)
	uN.Exp(u, bn)
	uInvN.Inverse(&uN)
	uInv2N.Mul(&uInvN, &uInvN)
	uInv3N.Mul(&uInv2N, &uInvN)
	shiftSq.Mul(&domain0.FrMultiplicativeGen, &domain0.FrMultiplicativeGen)
	out.omega4Inv = genericRawBW6761Fr([]bwfr.Element{omega4Inv})
	out.quarter = genericRawBW6761Fr([]bwfr.Element{quarter})
	out.uInvN = [3][]uint64{
		genericRawBW6761Fr([]bwfr.Element{uInvN}),
		genericRawBW6761Fr([]bwfr.Element{uInv2N}),
		genericRawBW6761Fr([]bwfr.Element{uInv3N}),
	}
	out.cosetShift = genericRawBW6761Fr([]bwfr.Element{domain0.FrMultiplicativeGen})
	out.cosetShiftSq = genericRawBW6761Fr([]bwfr.Element{shiftSq})
	return out
}
