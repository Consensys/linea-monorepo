//go:build cuda

package plonk2

import (
	"fmt"
	"hash"
	"math/big"

	bls12377 "github.com/consensys/gnark-crypto/ecc/bls12-377"
	blsfr "github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	blsfft "github.com/consensys/gnark-crypto/ecc/bls12-377/fr/fft"
	blshtf "github.com/consensys/gnark-crypto/ecc/bls12-377/fr/hash_to_field"
	bn254 "github.com/consensys/gnark-crypto/ecc/bn254"
	bnfr "github.com/consensys/gnark-crypto/ecc/bn254/fr"
	bnfft "github.com/consensys/gnark-crypto/ecc/bn254/fr/fft"
	bnhtf "github.com/consensys/gnark-crypto/ecc/bn254/fr/hash_to_field"
	bw6761 "github.com/consensys/gnark-crypto/ecc/bw6-761"
	bwfr "github.com/consensys/gnark-crypto/ecc/bw6-761/fr"
	bwfft "github.com/consensys/gnark-crypto/ecc/bw6-761/fr/fft"
	bwhtf "github.com/consensys/gnark-crypto/ecc/bw6-761/fr/hash_to_field"
	fiatshamir "github.com/consensys/gnark-crypto/fiat-shamir"
	"github.com/consensys/gnark/backend"
	gnarkplonk "github.com/consensys/gnark/backend/plonk"
	blsplonk "github.com/consensys/gnark/backend/plonk/bls12-377"
	bnplonk "github.com/consensys/gnark/backend/plonk/bn254"
	bwplonk "github.com/consensys/gnark/backend/plonk/bw6-761"
	"github.com/consensys/gnark/backend/witness"
	"github.com/consensys/gnark/constraint"
	csbls12377 "github.com/consensys/gnark/constraint/bls12-377"
	csbn254 "github.com/consensys/gnark/constraint/bn254"
	csbw6761 "github.com/consensys/gnark/constraint/bw6-761"
	"github.com/consensys/gnark/constraint/solver"
	fcs "github.com/consensys/gnark/frontend/cs"
)

type genericProveArtifacts struct {
	scratch        *genericProofScratch
	proof          gnarkplonk.Proof
	lroRaw         [3][]uint64
	lroCanonical   [3][]uint64
	lroBlinded     [3][]uint64
	qkCanonical    []uint64
	betaRaw        []uint64
	gammaRaw       []uint64
	alphaRaw       []uint64
	zetaRaw        []uint64
	zRaw           []uint64
	zCanonical     []uint64
	zBlinded       []uint64
	bsb22Raw       [][]uint64
	bsb22Canonical [][]uint64
	commitmentVals []uint64
	hRaw           [3][]uint64
}

const (
	genericLROBlindingOrder = 1
	genericZBlindingOrder   = 2
)

func proveGenericGPUBackend(
	b *genericGPUBackend,
	fullWitness witness.Witness,
	opts ...backend.ProverOption,
) (gnarkplonk.Proof, error) {
	proverConfig, err := backend.NewProverConfig(opts...)
	if err != nil {
		return nil, err
	}
	if proverConfig.StatisticalZK {
		return nil, fmt.Errorf("plonk2: generic GPU backend does not support statistical zero-knowledge yet")
	}
	if b.state.scratch != nil {
		b.state.scratch.mu.Lock()
		defer b.state.scratch.mu.Unlock()
	}
	artifacts, err := b.buildArtifacts(fullWitness, &proverConfig)
	if err != nil {
		return nil, err
	}
	if err := b.finalizeProof(artifacts, &proverConfig); err != nil {
		return nil, err
	}
	return artifacts.proof, nil
}

func (b *genericGPUBackend) buildArtifacts(
	fullWitness witness.Witness,
	proverConfig *backend.ProverConfig,
) (*genericProveArtifacts, error) {
	ops, err := newCurveProofOps(b.state.curve)
	if err != nil {
		return nil, err
	}
	proof := ops.newProof(plonkCommitmentCountForCCS(b.ccs))
	artifacts := &genericProveArtifacts{proof: proof, scratch: b.state.scratch}

	switch spr := b.ccs.(type) {
	case *csbn254.SparseR1CS:
		if err := b.solveBN254(artifacts, ops, spr, fullWitness, proverConfig); err != nil {
			return nil, err
		}
	case *csbls12377.SparseR1CS:
		if err := b.solveBLS12377(artifacts, ops, spr, fullWitness, proverConfig); err != nil {
			return nil, err
		}
	case *csbw6761.SparseR1CS:
		if err := b.solveBW6761(artifacts, ops, spr, fullWitness, proverConfig); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("plonk2: unsupported constraint system type %T", b.ccs)
	}

	if err := b.commitSolvedLRO(artifacts, ops); err != nil {
		return nil, err
	}
	if err := b.buildAndCommitZ(artifacts, ops, fullWitness, proverConfig); err != nil {
		return nil, err
	}
	if err := b.computeAndCommitQuotient(artifacts, ops, fullWitness, proverConfig); err != nil {
		return nil, err
	}
	return artifacts, nil
}

func (b *genericGPUBackend) finalizeProof(
	artifacts *genericProveArtifacts,
	proverConfig *backend.ProverConfig,
) error {
	switch b.state.curve {
	case CurveBN254:
		return b.finalizeBN254(artifacts, proverConfig)
	case CurveBLS12377:
		return b.finalizeBLS12377(artifacts, proverConfig)
	case CurveBW6761:
		return b.finalizeBW6761(artifacts, proverConfig)
	default:
		return fmt.Errorf("plonk2: unsupported curve %s", b.state.curve)
	}
}

func (b *genericGPUBackend) solveBN254(
	artifacts *genericProveArtifacts,
	ops curveProofOps,
	spr *csbn254.SparseR1CS,
	fullWitness witness.Witness,
	proverConfig *backend.ProverConfig,
) error {
	commitmentInfo := spr.CommitmentInfo.(constraint.PlonkCommitments)
	commitmentVals := make([]bnfr.Element, len(commitmentInfo))
	if artifacts.scratch != nil {
		artifacts.bsb22Raw = artifacts.scratch.bsb22Raw[:len(commitmentInfo)]
	} else {
		artifacts.bsb22Raw = make([][]uint64, len(commitmentInfo))
	}
	opts := append([]solver.Option(nil), proverConfig.SolverOpts...)
	if len(commitmentInfo) > 0 {
		htfFunc := proverConfig.HashToFieldFn
		if htfFunc == nil {
			htfFunc = bnhtf.New([]byte("BSB22-Plonk"))
		}
		opts = append(opts, solver.OverrideHint(
			solver.GetHintID(fcs.Bsb22CommitmentComputePlaceholder),
			func(_ *big.Int, ins, outs []*big.Int) error {
				commDepth, payload, err := splitCommitmentHintInputs(ins, len(commitmentInfo))
				if err != nil {
					return err
				}
				ci := commitmentInfo[commDepth]
				committedValues := make([]bnfr.Element, b.state.n)
				offset := spr.GetNbPublicVariables()
				for i := range payload {
					committedValues[offset+ci.Committed[i]].SetBigInt(payload[i])
				}
				if _, err = committedValues[offset+ci.CommitmentIndex].SetRandom(); err != nil {
					return err
				}
				if _, err = committedValues[offset+spr.GetNbConstraints()-1].SetRandom(); err != nil {
					return err
				}
				raw := genericRawBN254Fr(committedValues)
				if artifacts.scratch != nil {
					raw = artifacts.scratch.bsb22Raw[commDepth]
					copyBN254FrToRaw(raw, committedValues)
				}
				artifacts.bsb22Raw[commDepth] = raw
				digest, err := b.commitBSB22Raw(ops, artifacts.proof, commDepth, raw)
				if err != nil {
					return err
				}
				if err := bindCommitmentHintOutput(htfFunc, bnfr.Bytes, digest, ops, outs, func(hashBytes []byte, out *big.Int) {
					commitmentVals[commDepth].SetBytes(hashBytes)
					commitmentVals[commDepth].BigInt(out)
				}); err != nil {
					return err
				}
				return nil
			},
		))
	}

	solution, err := spr.Solve(fullWitness, opts...)
	if err != nil {
		return fmt.Errorf("solve: %w", err)
	}
	typed := solution.(*csbn254.SparseR1CSSolution)
	if artifacts.scratch != nil {
		artifacts.lroRaw[0] = artifacts.scratch.lroRaw[0]
		artifacts.lroRaw[1] = artifacts.scratch.lroRaw[1]
		artifacts.lroRaw[2] = artifacts.scratch.lroRaw[2]
		copyBN254FrToRaw(artifacts.lroRaw[0], []bnfr.Element(typed.L))
		copyBN254FrToRaw(artifacts.lroRaw[1], []bnfr.Element(typed.R))
		copyBN254FrToRaw(artifacts.lroRaw[2], []bnfr.Element(typed.O))
		artifacts.commitmentVals = artifacts.scratch.commitments[:len(commitmentVals)*bnfr.Limbs]
		copyBN254FrToRaw(artifacts.commitmentVals, commitmentVals)
	} else {
		artifacts.lroRaw[0] = genericRawBN254Fr([]bnfr.Element(typed.L))
		artifacts.lroRaw[1] = genericRawBN254Fr([]bnfr.Element(typed.R))
		artifacts.lroRaw[2] = genericRawBN254Fr([]bnfr.Element(typed.O))
		artifacts.commitmentVals = genericRawBN254Fr(commitmentVals)
	}
	return nil
}

func (b *genericGPUBackend) solveBLS12377(
	artifacts *genericProveArtifacts,
	ops curveProofOps,
	spr *csbls12377.SparseR1CS,
	fullWitness witness.Witness,
	proverConfig *backend.ProverConfig,
) error {
	commitmentInfo := spr.CommitmentInfo.(constraint.PlonkCommitments)
	commitmentVals := make([]blsfr.Element, len(commitmentInfo))
	if artifacts.scratch != nil {
		artifacts.bsb22Raw = artifacts.scratch.bsb22Raw[:len(commitmentInfo)]
	} else {
		artifacts.bsb22Raw = make([][]uint64, len(commitmentInfo))
	}
	opts := append([]solver.Option(nil), proverConfig.SolverOpts...)
	if len(commitmentInfo) > 0 {
		htfFunc := proverConfig.HashToFieldFn
		if htfFunc == nil {
			htfFunc = blshtf.New([]byte("BSB22-Plonk"))
		}
		opts = append(opts, solver.OverrideHint(
			solver.GetHintID(fcs.Bsb22CommitmentComputePlaceholder),
			func(_ *big.Int, ins, outs []*big.Int) error {
				commDepth, payload, err := splitCommitmentHintInputs(ins, len(commitmentInfo))
				if err != nil {
					return err
				}
				ci := commitmentInfo[commDepth]
				var committedValues []blsfr.Element
				if artifacts.scratch != nil && artifacts.scratch.bls12377 != nil {
					committedValues = artifacts.scratch.bls12377.bsb22Committed[commDepth]
					clearBLS12377Vector(committedValues)
				} else {
					committedValues = make([]blsfr.Element, b.state.n)
				}
				offset := spr.GetNbPublicVariables()
				for i := range payload {
					committedValues[offset+ci.Committed[i]].SetBigInt(payload[i])
				}
				if _, err = committedValues[offset+ci.CommitmentIndex].SetRandom(); err != nil {
					return err
				}
				if _, err = committedValues[offset+spr.GetNbConstraints()-1].SetRandom(); err != nil {
					return err
				}
				raw := genericRawBLS12377Fr(committedValues)
				if artifacts.scratch != nil {
					raw = artifacts.scratch.bsb22Raw[commDepth]
					copyBLS12377FrToRaw(raw, committedValues)
				}
				artifacts.bsb22Raw[commDepth] = raw
				digest, err := b.commitBSB22Raw(ops, artifacts.proof, commDepth, raw)
				if err != nil {
					return err
				}
				if err := bindCommitmentHintOutput(htfFunc, blsfr.Bytes, digest, ops, outs, func(hashBytes []byte, out *big.Int) {
					commitmentVals[commDepth].SetBytes(hashBytes)
					commitmentVals[commDepth].BigInt(out)
				}); err != nil {
					return err
				}
				return nil
			},
		))
	}

	solution, err := spr.Solve(fullWitness, opts...)
	if err != nil {
		return fmt.Errorf("solve: %w", err)
	}
	typed := solution.(*csbls12377.SparseR1CSSolution)
	if artifacts.scratch != nil {
		artifacts.lroRaw[0] = artifacts.scratch.lroRaw[0]
		artifacts.lroRaw[1] = artifacts.scratch.lroRaw[1]
		artifacts.lroRaw[2] = artifacts.scratch.lroRaw[2]
		copyBLS12377FrToRaw(artifacts.lroRaw[0], []blsfr.Element(typed.L))
		copyBLS12377FrToRaw(artifacts.lroRaw[1], []blsfr.Element(typed.R))
		copyBLS12377FrToRaw(artifacts.lroRaw[2], []blsfr.Element(typed.O))
		artifacts.commitmentVals = artifacts.scratch.commitments[:len(commitmentVals)*blsfr.Limbs]
		copyBLS12377FrToRaw(artifacts.commitmentVals, commitmentVals)
	} else {
		artifacts.lroRaw[0] = genericRawBLS12377Fr([]blsfr.Element(typed.L))
		artifacts.lroRaw[1] = genericRawBLS12377Fr([]blsfr.Element(typed.R))
		artifacts.lroRaw[2] = genericRawBLS12377Fr([]blsfr.Element(typed.O))
		artifacts.commitmentVals = genericRawBLS12377Fr(commitmentVals)
	}
	return nil
}

func (b *genericGPUBackend) solveBW6761(
	artifacts *genericProveArtifacts,
	ops curveProofOps,
	spr *csbw6761.SparseR1CS,
	fullWitness witness.Witness,
	proverConfig *backend.ProverConfig,
) error {
	commitmentInfo := spr.CommitmentInfo.(constraint.PlonkCommitments)
	commitmentVals := make([]bwfr.Element, len(commitmentInfo))
	if artifacts.scratch != nil {
		artifacts.bsb22Raw = artifacts.scratch.bsb22Raw[:len(commitmentInfo)]
	} else {
		artifacts.bsb22Raw = make([][]uint64, len(commitmentInfo))
	}
	opts := append([]solver.Option(nil), proverConfig.SolverOpts...)
	if len(commitmentInfo) > 0 {
		htfFunc := proverConfig.HashToFieldFn
		if htfFunc == nil {
			htfFunc = bwhtf.New([]byte("BSB22-Plonk"))
		}
		opts = append(opts, solver.OverrideHint(
			solver.GetHintID(fcs.Bsb22CommitmentComputePlaceholder),
			func(_ *big.Int, ins, outs []*big.Int) error {
				commDepth, payload, err := splitCommitmentHintInputs(ins, len(commitmentInfo))
				if err != nil {
					return err
				}
				ci := commitmentInfo[commDepth]
				committedValues := make([]bwfr.Element, b.state.n)
				offset := spr.GetNbPublicVariables()
				for i := range payload {
					committedValues[offset+ci.Committed[i]].SetBigInt(payload[i])
				}
				if _, err = committedValues[offset+ci.CommitmentIndex].SetRandom(); err != nil {
					return err
				}
				if _, err = committedValues[offset+spr.GetNbConstraints()-1].SetRandom(); err != nil {
					return err
				}
				raw := genericRawBW6761Fr(committedValues)
				if artifacts.scratch != nil {
					raw = artifacts.scratch.bsb22Raw[commDepth]
					copyBW6761FrToRaw(raw, committedValues)
				}
				artifacts.bsb22Raw[commDepth] = raw
				digest, err := b.commitBSB22Raw(ops, artifacts.proof, commDepth, raw)
				if err != nil {
					return err
				}
				if err := bindCommitmentHintOutput(htfFunc, bwfr.Bytes, digest, ops, outs, func(hashBytes []byte, out *big.Int) {
					commitmentVals[commDepth].SetBytes(hashBytes)
					commitmentVals[commDepth].BigInt(out)
				}); err != nil {
					return err
				}
				return nil
			},
		))
	}

	solution, err := spr.Solve(fullWitness, opts...)
	if err != nil {
		return fmt.Errorf("solve: %w", err)
	}
	typed := solution.(*csbw6761.SparseR1CSSolution)
	if artifacts.scratch != nil {
		artifacts.lroRaw[0] = artifacts.scratch.lroRaw[0]
		artifacts.lroRaw[1] = artifacts.scratch.lroRaw[1]
		artifacts.lroRaw[2] = artifacts.scratch.lroRaw[2]
		copyBW6761FrToRaw(artifacts.lroRaw[0], []bwfr.Element(typed.L))
		copyBW6761FrToRaw(artifacts.lroRaw[1], []bwfr.Element(typed.R))
		copyBW6761FrToRaw(artifacts.lroRaw[2], []bwfr.Element(typed.O))
		artifacts.commitmentVals = artifacts.scratch.commitments[:len(commitmentVals)*bwfr.Limbs]
		copyBW6761FrToRaw(artifacts.commitmentVals, commitmentVals)
	} else {
		artifacts.lroRaw[0] = genericRawBW6761Fr([]bwfr.Element(typed.L))
		artifacts.lroRaw[1] = genericRawBW6761Fr([]bwfr.Element(typed.R))
		artifacts.lroRaw[2] = genericRawBW6761Fr([]bwfr.Element(typed.O))
		artifacts.commitmentVals = genericRawBW6761Fr(commitmentVals)
	}
	return nil
}

func (b *genericGPUBackend) commitBSB22Raw(
	ops curveProofOps,
	proof gnarkplonk.Proof,
	commDepth int,
	raw []uint64,
) (any, error) {
	commitRaw, err := b.state.commitLagrangeRaw(raw)
	if err != nil {
		return nil, fmt.Errorf("commit BSB22[%d]: %w", commDepth, err)
	}
	digest, err := ops.rawCommitmentToDigest(commitRaw)
	if err != nil {
		return nil, err
	}
	if err := ops.setBsb22Commitment(proof, commDepth, digest); err != nil {
		return nil, err
	}
	return digest, nil
}

func (b *genericGPUBackend) commitSolvedLRO(artifacts *genericProveArtifacts, ops curveProofOps) error {
	wave := make([][]uint64, len(artifacts.lroRaw))
	for i := range artifacts.lroRaw {
		var canonical []uint64
		var err error
		if artifacts.scratch != nil {
			canonical = artifacts.scratch.lroCanonical[i]
			if err := b.lagrangeRawToCanonicalRawInto(canonical, artifacts.lroRaw[i]); err != nil {
				return fmt.Errorf("canonicalize LRO[%d]: %w", i, err)
			}
		} else {
			canonical, err = b.lagrangeRawToCanonicalRaw(artifacts.lroRaw[i])
			if err != nil {
				return fmt.Errorf("canonicalize LRO[%d]: %w", i, err)
			}
		}
		artifacts.lroCanonical[i] = canonical
		var blinded []uint64
		if artifacts.scratch != nil {
			blinded = artifacts.scratch.lroBlinded[i]
			if err := blindCanonicalRawInto(blinded, b.state.curve, canonical, genericLROBlindingOrder); err != nil {
				return fmt.Errorf("blind LRO[%d]: %w", i, err)
			}
		} else {
			blinded, err = blindCanonicalRaw(b.state.curve, canonical, genericLROBlindingOrder)
			if err != nil {
				return fmt.Errorf("blind LRO[%d]: %w", i, err)
			}
		}
		artifacts.lroBlinded[i] = blinded
		wave[i] = blinded
	}

	if err := b.state.pinCanonicalCommitmentWave(); err != nil {
		return err
	}
	defer b.state.releaseCanonicalCommitmentWave()
	commits, err := b.state.commitCanonicalWaveRaw(wave)
	if err != nil {
		return err
	}
	for i := range commits {
		digest, err := ops.rawCommitmentToDigest(commits[i])
		if err != nil {
			return err
		}
		if err := ops.setLRO(artifacts.proof, i, digest); err != nil {
			return err
		}
	}
	return nil
}

func (b *genericGPUBackend) buildAndCommitZ(
	artifacts *genericProveArtifacts,
	ops curveProofOps,
	fullWitness witness.Witness,
	proverConfig *backend.ProverConfig,
) error {
	if err := b.completeQkAndCommitmentPolynomials(artifacts, fullWitness); err != nil {
		return err
	}
	betaRaw, gammaRaw, cosetShiftRaw, cosetShiftSqRaw, err := b.deriveGammaBetaRaw(
		artifacts,
		fullWitness,
		proverConfig,
	)
	if err != nil {
		return err
	}
	artifacts.betaRaw = betaRaw
	artifacts.gammaRaw = gammaRaw

	var zRaw []uint64
	if artifacts.scratch != nil {
		zRaw = artifacts.scratch.zRaw
		if err := b.buildZRawInto(zRaw, artifacts.lroRaw, betaRaw, gammaRaw, cosetShiftRaw, cosetShiftSqRaw); err != nil {
			return err
		}
	} else {
		zRaw, err = b.buildZRaw(artifacts.lroRaw, betaRaw, gammaRaw, cosetShiftRaw, cosetShiftSqRaw)
		if err != nil {
			return err
		}
	}
	artifacts.zRaw = zRaw
	var zCanonical []uint64
	if artifacts.scratch != nil {
		zCanonical = artifacts.scratch.zCanonical
		if err := b.lagrangeRawToCanonicalRawInto(zCanonical, zRaw); err != nil {
			return fmt.Errorf("canonicalize Z: %w", err)
		}
	} else {
		zCanonical, err = b.lagrangeRawToCanonicalRaw(zRaw)
		if err != nil {
			return fmt.Errorf("canonicalize Z: %w", err)
		}
	}
	artifacts.zCanonical = zCanonical
	var zBlinded []uint64
	if artifacts.scratch != nil {
		zBlinded = artifacts.scratch.zBlinded
		if err := blindCanonicalRawInto(zBlinded, b.state.curve, zCanonical, genericZBlindingOrder); err != nil {
			return fmt.Errorf("blind Z: %w", err)
		}
	} else {
		zBlinded, err = blindCanonicalRaw(b.state.curve, zCanonical, genericZBlindingOrder)
		if err != nil {
			return fmt.Errorf("blind Z: %w", err)
		}
	}
	artifacts.zBlinded = zBlinded
	commitRaw, err := b.state.kzg.CommitRaw(zBlinded)
	if err != nil {
		return fmt.Errorf("commit Z: %w", err)
	}
	digest, err := ops.rawCommitmentToDigest(commitRaw)
	if err != nil {
		return err
	}
	if err := ops.setZ(artifacts.proof, digest); err != nil {
		return err
	}
	alphaRaw, err := b.deriveAlphaRaw(artifacts, fullWitness, proverConfig)
	if err != nil {
		return err
	}
	artifacts.alphaRaw = alphaRaw
	return nil
}

func (b *genericGPUBackend) completeQkAndCommitmentPolynomials(
	artifacts *genericProveArtifacts,
	fullWitness witness.Witness,
) error {
	qkLagrange, err := b.completeQkLagrangeRaw(artifacts, fullWitness)
	if err != nil {
		return err
	}
	var qkCanonical []uint64
	if artifacts.scratch != nil {
		qkCanonical = artifacts.scratch.qkCanonical
		if err := b.lagrangeRawToCanonicalRawInto(qkCanonical, qkLagrange); err != nil {
			return fmt.Errorf("canonicalize Qk: %w", err)
		}
	} else {
		qkCanonical, err = b.lagrangeRawToCanonicalRaw(qkLagrange)
		if err != nil {
			return fmt.Errorf("canonicalize Qk: %w", err)
		}
	}
	artifacts.qkCanonical = qkCanonical

	if artifacts.scratch != nil {
		artifacts.bsb22Canonical = artifacts.scratch.bsb22Canon[:len(artifacts.bsb22Raw)]
	} else {
		artifacts.bsb22Canonical = make([][]uint64, len(artifacts.bsb22Raw))
	}
	for i := range artifacts.bsb22Raw {
		if len(artifacts.bsb22Raw[i]) == 0 {
			continue
		}
		var canonical []uint64
		if artifacts.scratch != nil {
			canonical = artifacts.scratch.bsb22Canon[i]
			if err := b.lagrangeRawToCanonicalRawInto(canonical, artifacts.bsb22Raw[i]); err != nil {
				return fmt.Errorf("canonicalize BSB22[%d]: %w", i, err)
			}
		} else {
			canonical, err = b.lagrangeRawToCanonicalRaw(artifacts.bsb22Raw[i])
			if err != nil {
				return fmt.Errorf("canonicalize BSB22[%d]: %w", i, err)
			}
		}
		artifacts.bsb22Canonical[i] = canonical
	}
	return nil
}

func (b *genericGPUBackend) completeQkLagrangeRaw(
	artifacts *genericProveArtifacts,
	fullWitness witness.Witness,
) ([]uint64, error) {
	switch spr := b.ccs.(type) {
	case *csbn254.SparseR1CS:
		w, ok := fullWitness.Vector().(bnfr.Vector)
		if !ok {
			return nil, witness.ErrInvalidWitness
		}
		if artifacts.scratch != nil {
			qkRaw := artifacts.scratch.qkLagrange
			copy(qkRaw, b.state.qkLagrangeTemplate)
			for i := range spr.Public {
				writeBN254RawElement(rawElementSlice(qkRaw, bnfr.Limbs, i, i+1), &w[i])
			}
			commitments := spr.CommitmentInfo.(constraint.PlonkCommitments)
			for i := range commitments {
				value := rawElementSlice(artifacts.commitmentVals, bnfr.Limbs, i, i+1)
				copy(rawElementSlice(qkRaw, bnfr.Limbs, spr.GetNbPublicVariables()+commitments[i].CommitmentIndex, spr.GetNbPublicVariables()+commitments[i].CommitmentIndex+1), value)
			}
			return qkRaw, nil
		}
		pk := b.pk.(*bnplonk.ProvingKey)
		trace := bnplonk.NewTrace(spr, bnfft.NewDomain(pk.Vk.Size, bnfft.WithoutPrecompute()))
		qk := append([]bnfr.Element(nil), trace.Qk.Coefficients()...)
		copy(qk, w[:len(spr.Public)])
		commitments := spr.CommitmentInfo.(constraint.PlonkCommitments)
		commitmentVals, err := bn254RawToField(artifacts.commitmentVals)
		if err != nil {
			return nil, err
		}
		for i := range commitments {
			qk[spr.GetNbPublicVariables()+commitments[i].CommitmentIndex] = commitmentVals[i]
		}
		return genericRawBN254Fr(qk), nil
	case *csbls12377.SparseR1CS:
		w, ok := fullWitness.Vector().(blsfr.Vector)
		if !ok {
			return nil, witness.ErrInvalidWitness
		}
		if artifacts.scratch != nil {
			qkRaw := artifacts.scratch.qkLagrange
			copy(qkRaw, b.state.qkLagrangeTemplate)
			for i := range spr.Public {
				writeBLS12377RawElement(rawElementSlice(qkRaw, blsfr.Limbs, i, i+1), &w[i])
			}
			commitments := spr.CommitmentInfo.(constraint.PlonkCommitments)
			for i := range commitments {
				value := rawElementSlice(artifacts.commitmentVals, blsfr.Limbs, i, i+1)
				copy(rawElementSlice(qkRaw, blsfr.Limbs, spr.GetNbPublicVariables()+commitments[i].CommitmentIndex, spr.GetNbPublicVariables()+commitments[i].CommitmentIndex+1), value)
			}
			return qkRaw, nil
		}
		pk := b.pk.(*blsplonk.ProvingKey)
		trace := blsplonk.NewTrace(spr, blsfft.NewDomain(pk.Vk.Size, blsfft.WithoutPrecompute()))
		qk := append([]blsfr.Element(nil), trace.Qk.Coefficients()...)
		copy(qk, w[:len(spr.Public)])
		commitments := spr.CommitmentInfo.(constraint.PlonkCommitments)
		commitmentVals, err := bls12377RawToField(artifacts.commitmentVals)
		if err != nil {
			return nil, err
		}
		for i := range commitments {
			qk[spr.GetNbPublicVariables()+commitments[i].CommitmentIndex] = commitmentVals[i]
		}
		return genericRawBLS12377Fr(qk), nil
	case *csbw6761.SparseR1CS:
		w, ok := fullWitness.Vector().(bwfr.Vector)
		if !ok {
			return nil, witness.ErrInvalidWitness
		}
		if artifacts.scratch != nil {
			qkRaw := artifacts.scratch.qkLagrange
			copy(qkRaw, b.state.qkLagrangeTemplate)
			for i := range spr.Public {
				writeBW6761RawElement(rawElementSlice(qkRaw, bwfr.Limbs, i, i+1), &w[i])
			}
			commitments := spr.CommitmentInfo.(constraint.PlonkCommitments)
			for i := range commitments {
				value := rawElementSlice(artifacts.commitmentVals, bwfr.Limbs, i, i+1)
				copy(rawElementSlice(qkRaw, bwfr.Limbs, spr.GetNbPublicVariables()+commitments[i].CommitmentIndex, spr.GetNbPublicVariables()+commitments[i].CommitmentIndex+1), value)
			}
			return qkRaw, nil
		}
		pk := b.pk.(*bwplonk.ProvingKey)
		trace := bwplonk.NewTrace(spr, bwfft.NewDomain(pk.Vk.Size, bwfft.WithoutPrecompute()))
		qk := append([]bwfr.Element(nil), trace.Qk.Coefficients()...)
		copy(qk, w[:len(spr.Public)])
		commitments := spr.CommitmentInfo.(constraint.PlonkCommitments)
		commitmentVals, err := bw6761RawToField(artifacts.commitmentVals)
		if err != nil {
			return nil, err
		}
		for i := range commitments {
			qk[spr.GetNbPublicVariables()+commitments[i].CommitmentIndex] = commitmentVals[i]
		}
		return genericRawBW6761Fr(qk), nil
	default:
		return nil, fmt.Errorf("plonk2: unsupported constraint system type %T", b.ccs)
	}
}

func (b *genericGPUBackend) buildZRaw(
	lro [3][]uint64,
	betaRaw, gammaRaw, cosetShiftRaw, cosetShiftSqRaw []uint64,
) ([]uint64, error) {
	zRaw := make([]uint64, b.state.n*scalarLimbs(b.state.curve))
	if err := b.buildZRawInto(zRaw, lro, betaRaw, gammaRaw, cosetShiftRaw, cosetShiftSqRaw); err != nil {
		return nil, err
	}
	return zRaw, nil
}

func (b *genericGPUBackend) buildZRawInto(
	zRaw []uint64,
	lro [3][]uint64,
	betaRaw, gammaRaw, cosetShiftRaw, cosetShiftSqRaw []uint64,
) error {
	l, err := NewFrVector(b.state.dev, b.state.curve, b.state.n)
	if err != nil {
		return err
	}
	defer l.Free()
	r, err := NewFrVector(b.state.dev, b.state.curve, b.state.n)
	if err != nil {
		return err
	}
	defer r.Free()
	o, err := NewFrVector(b.state.dev, b.state.curve, b.state.n)
	if err != nil {
		return err
	}
	defer o.Free()
	temp, err := NewFrVector(b.state.dev, b.state.curve, b.state.n)
	if err != nil {
		return err
	}
	defer temp.Free()
	z, err := NewFrVector(b.state.dev, b.state.curve, b.state.n)
	if err != nil {
		return err
	}
	defer z.Free()

	if err := l.CopyFromHostRaw(lro[0]); err != nil {
		return err
	}
	if err := r.CopyFromHostRaw(lro[1]); err != nil {
		return err
	}
	if err := o.CopyFromHostRaw(lro[2]); err != nil {
		return err
	}
	if err := PlonkZComputeFactors(
		l,
		r,
		o,
		b.state.perm,
		b.state.fft,
		betaRaw,
		gammaRaw,
		cosetShiftRaw,
		cosetShiftSqRaw,
		b.state.log2n,
	); err != nil {
		return err
	}
	if err := r.BatchInvert(temp); err != nil {
		return err
	}
	if err := l.Mul(l, r); err != nil {
		return err
	}
	if err := ZPrefixProduct(z, l, temp); err != nil {
		return err
	}
	if err := z.CopyToHostRaw(zRaw); err != nil {
		return err
	}
	return nil
}

func (b *genericGPUBackend) deriveGammaBetaRaw(
	artifacts *genericProveArtifacts,
	fullWitness witness.Witness,
	proverConfig *backend.ProverConfig,
) (betaRaw, gammaRaw, cosetShiftRaw, cosetShiftSqRaw []uint64, err error) {
	switch pk := b.pk.(type) {
	case *bnplonk.ProvingKey:
		proof := artifacts.proof.(*bnplonk.Proof)
		w, ok := fullWitness.Vector().(bnfr.Vector)
		if !ok {
			return nil, nil, nil, nil, witness.ErrInvalidWitness
		}
		fs := newGenericProverTranscript(proverConfig)
		if err := bindPublicDataBN254(fs, "gamma", pk.Vk, w[:pk.Vk.NbPublicWitness()]); err != nil {
			return nil, nil, nil, nil, err
		}
		gamma, err := deriveRandomnessBN254(fs, "gamma", &proof.LRO[0], &proof.LRO[1], &proof.LRO[2])
		if err != nil {
			return nil, nil, nil, nil, err
		}
		betaBytes, err := fs.ComputeChallenge("beta")
		if err != nil {
			return nil, nil, nil, nil, err
		}
		var beta, cosetShiftSq bnfr.Element
		beta.SetBytes(betaBytes)
		cosetShiftSq.Mul(&pk.Vk.CosetShift, &pk.Vk.CosetShift)
		return genericRawBN254Fr([]bnfr.Element{beta}),
			genericRawBN254Fr([]bnfr.Element{gamma}),
			genericRawBN254Fr([]bnfr.Element{pk.Vk.CosetShift}),
			genericRawBN254Fr([]bnfr.Element{cosetShiftSq}),
			nil
	case *blsplonk.ProvingKey:
		proof := artifacts.proof.(*blsplonk.Proof)
		w, ok := fullWitness.Vector().(blsfr.Vector)
		if !ok {
			return nil, nil, nil, nil, witness.ErrInvalidWitness
		}
		fs := newGenericProverTranscript(proverConfig)
		if err := bindPublicDataBLS12377(fs, "gamma", pk.Vk, w[:pk.Vk.NbPublicWitness()]); err != nil {
			return nil, nil, nil, nil, err
		}
		gamma, err := deriveRandomnessBLS12377(fs, "gamma", &proof.LRO[0], &proof.LRO[1], &proof.LRO[2])
		if err != nil {
			return nil, nil, nil, nil, err
		}
		betaBytes, err := fs.ComputeChallenge("beta")
		if err != nil {
			return nil, nil, nil, nil, err
		}
		var beta, cosetShiftSq blsfr.Element
		beta.SetBytes(betaBytes)
		cosetShiftSq.Mul(&pk.Vk.CosetShift, &pk.Vk.CosetShift)
		return genericRawBLS12377Fr([]blsfr.Element{beta}),
			genericRawBLS12377Fr([]blsfr.Element{gamma}),
			genericRawBLS12377Fr([]blsfr.Element{pk.Vk.CosetShift}),
			genericRawBLS12377Fr([]blsfr.Element{cosetShiftSq}),
			nil
	case *bwplonk.ProvingKey:
		proof := artifacts.proof.(*bwplonk.Proof)
		w, ok := fullWitness.Vector().(bwfr.Vector)
		if !ok {
			return nil, nil, nil, nil, witness.ErrInvalidWitness
		}
		fs := newGenericProverTranscript(proverConfig)
		if err := bindPublicDataBW6761(fs, "gamma", pk.Vk, w[:pk.Vk.NbPublicWitness()]); err != nil {
			return nil, nil, nil, nil, err
		}
		gamma, err := deriveRandomnessBW6761(fs, "gamma", &proof.LRO[0], &proof.LRO[1], &proof.LRO[2])
		if err != nil {
			return nil, nil, nil, nil, err
		}
		betaBytes, err := fs.ComputeChallenge("beta")
		if err != nil {
			return nil, nil, nil, nil, err
		}
		var beta, cosetShiftSq bwfr.Element
		beta.SetBytes(betaBytes)
		cosetShiftSq.Mul(&pk.Vk.CosetShift, &pk.Vk.CosetShift)
		return genericRawBW6761Fr([]bwfr.Element{beta}),
			genericRawBW6761Fr([]bwfr.Element{gamma}),
			genericRawBW6761Fr([]bwfr.Element{pk.Vk.CosetShift}),
			genericRawBW6761Fr([]bwfr.Element{cosetShiftSq}),
			nil
	default:
		return nil, nil, nil, nil, fmt.Errorf("plonk2: unsupported proving key type %T", b.pk)
	}
}

func (b *genericGPUBackend) deriveAlphaRaw(
	artifacts *genericProveArtifacts,
	fullWitness witness.Witness,
	proverConfig *backend.ProverConfig,
) ([]uint64, error) {
	switch pk := b.pk.(type) {
	case *bnplonk.ProvingKey:
		proof := artifacts.proof.(*bnplonk.Proof)
		w, ok := fullWitness.Vector().(bnfr.Vector)
		if !ok {
			return nil, witness.ErrInvalidWitness
		}
		fs := newGenericProverTranscript(proverConfig)
		if err := bindPublicDataBN254(fs, "gamma", pk.Vk, w[:pk.Vk.NbPublicWitness()]); err != nil {
			return nil, err
		}
		if _, err := deriveRandomnessBN254(fs, "gamma", &proof.LRO[0], &proof.LRO[1], &proof.LRO[2]); err != nil {
			return nil, err
		}
		if _, err := fs.ComputeChallenge("beta"); err != nil {
			return nil, err
		}
		alphaDeps := make([]*bn254.G1Affine, len(proof.Bsb22Commitments)+1)
		for i := range proof.Bsb22Commitments {
			alphaDeps[i] = &proof.Bsb22Commitments[i]
		}
		alphaDeps[len(alphaDeps)-1] = &proof.Z
		alpha, err := deriveRandomnessBN254(fs, "alpha", alphaDeps...)
		if err != nil {
			return nil, err
		}
		return genericRawBN254Fr([]bnfr.Element{alpha}), nil
	case *blsplonk.ProvingKey:
		proof := artifacts.proof.(*blsplonk.Proof)
		w, ok := fullWitness.Vector().(blsfr.Vector)
		if !ok {
			return nil, witness.ErrInvalidWitness
		}
		fs := newGenericProverTranscript(proverConfig)
		if err := bindPublicDataBLS12377(fs, "gamma", pk.Vk, w[:pk.Vk.NbPublicWitness()]); err != nil {
			return nil, err
		}
		if _, err := deriveRandomnessBLS12377(fs, "gamma", &proof.LRO[0], &proof.LRO[1], &proof.LRO[2]); err != nil {
			return nil, err
		}
		if _, err := fs.ComputeChallenge("beta"); err != nil {
			return nil, err
		}
		alphaDeps := make([]*bls12377.G1Affine, len(proof.Bsb22Commitments)+1)
		for i := range proof.Bsb22Commitments {
			alphaDeps[i] = &proof.Bsb22Commitments[i]
		}
		alphaDeps[len(alphaDeps)-1] = &proof.Z
		alpha, err := deriveRandomnessBLS12377(fs, "alpha", alphaDeps...)
		if err != nil {
			return nil, err
		}
		return genericRawBLS12377Fr([]blsfr.Element{alpha}), nil
	case *bwplonk.ProvingKey:
		proof := artifacts.proof.(*bwplonk.Proof)
		w, ok := fullWitness.Vector().(bwfr.Vector)
		if !ok {
			return nil, witness.ErrInvalidWitness
		}
		fs := newGenericProverTranscript(proverConfig)
		if err := bindPublicDataBW6761(fs, "gamma", pk.Vk, w[:pk.Vk.NbPublicWitness()]); err != nil {
			return nil, err
		}
		if _, err := deriveRandomnessBW6761(fs, "gamma", &proof.LRO[0], &proof.LRO[1], &proof.LRO[2]); err != nil {
			return nil, err
		}
		if _, err := fs.ComputeChallenge("beta"); err != nil {
			return nil, err
		}
		alphaDeps := make([]*bw6761.G1Affine, len(proof.Bsb22Commitments)+1)
		for i := range proof.Bsb22Commitments {
			alphaDeps[i] = &proof.Bsb22Commitments[i]
		}
		alphaDeps[len(alphaDeps)-1] = &proof.Z
		alpha, err := deriveRandomnessBW6761(fs, "alpha", alphaDeps...)
		if err != nil {
			return nil, err
		}
		return genericRawBW6761Fr([]bwfr.Element{alpha}), nil
	default:
		return nil, fmt.Errorf("plonk2: unsupported proving key type %T", b.pk)
	}
}

func (b *genericGPUBackend) deriveZetaRaw(
	artifacts *genericProveArtifacts,
	fullWitness witness.Witness,
	proverConfig *backend.ProverConfig,
) ([]uint64, error) {
	switch pk := b.pk.(type) {
	case *bnplonk.ProvingKey:
		proof := artifacts.proof.(*bnplonk.Proof)
		w, ok := fullWitness.Vector().(bnfr.Vector)
		if !ok {
			return nil, witness.ErrInvalidWitness
		}
		fs := newGenericProverTranscript(proverConfig)
		if err := bindPublicDataBN254(fs, "gamma", pk.Vk, w[:pk.Vk.NbPublicWitness()]); err != nil {
			return nil, err
		}
		if _, err := deriveRandomnessBN254(fs, "gamma", &proof.LRO[0], &proof.LRO[1], &proof.LRO[2]); err != nil {
			return nil, err
		}
		if _, err := fs.ComputeChallenge("beta"); err != nil {
			return nil, err
		}
		alphaDeps := make([]*bn254.G1Affine, len(proof.Bsb22Commitments)+1)
		for i := range proof.Bsb22Commitments {
			alphaDeps[i] = &proof.Bsb22Commitments[i]
		}
		alphaDeps[len(alphaDeps)-1] = &proof.Z
		if _, err := deriveRandomnessBN254(fs, "alpha", alphaDeps...); err != nil {
			return nil, err
		}
		zeta, err := deriveRandomnessBN254(fs, "zeta", &proof.H[0], &proof.H[1], &proof.H[2])
		if err != nil {
			return nil, err
		}
		return genericRawBN254Fr([]bnfr.Element{zeta}), nil
	case *blsplonk.ProvingKey:
		proof := artifacts.proof.(*blsplonk.Proof)
		w, ok := fullWitness.Vector().(blsfr.Vector)
		if !ok {
			return nil, witness.ErrInvalidWitness
		}
		fs := newGenericProverTranscript(proverConfig)
		if err := bindPublicDataBLS12377(fs, "gamma", pk.Vk, w[:pk.Vk.NbPublicWitness()]); err != nil {
			return nil, err
		}
		if _, err := deriveRandomnessBLS12377(fs, "gamma", &proof.LRO[0], &proof.LRO[1], &proof.LRO[2]); err != nil {
			return nil, err
		}
		if _, err := fs.ComputeChallenge("beta"); err != nil {
			return nil, err
		}
		alphaDeps := make([]*bls12377.G1Affine, len(proof.Bsb22Commitments)+1)
		for i := range proof.Bsb22Commitments {
			alphaDeps[i] = &proof.Bsb22Commitments[i]
		}
		alphaDeps[len(alphaDeps)-1] = &proof.Z
		if _, err := deriveRandomnessBLS12377(fs, "alpha", alphaDeps...); err != nil {
			return nil, err
		}
		zeta, err := deriveRandomnessBLS12377(fs, "zeta", &proof.H[0], &proof.H[1], &proof.H[2])
		if err != nil {
			return nil, err
		}
		return genericRawBLS12377Fr([]blsfr.Element{zeta}), nil
	case *bwplonk.ProvingKey:
		proof := artifacts.proof.(*bwplonk.Proof)
		w, ok := fullWitness.Vector().(bwfr.Vector)
		if !ok {
			return nil, witness.ErrInvalidWitness
		}
		fs := newGenericProverTranscript(proverConfig)
		if err := bindPublicDataBW6761(fs, "gamma", pk.Vk, w[:pk.Vk.NbPublicWitness()]); err != nil {
			return nil, err
		}
		if _, err := deriveRandomnessBW6761(fs, "gamma", &proof.LRO[0], &proof.LRO[1], &proof.LRO[2]); err != nil {
			return nil, err
		}
		if _, err := fs.ComputeChallenge("beta"); err != nil {
			return nil, err
		}
		alphaDeps := make([]*bw6761.G1Affine, len(proof.Bsb22Commitments)+1)
		for i := range proof.Bsb22Commitments {
			alphaDeps[i] = &proof.Bsb22Commitments[i]
		}
		alphaDeps[len(alphaDeps)-1] = &proof.Z
		if _, err := deriveRandomnessBW6761(fs, "alpha", alphaDeps...); err != nil {
			return nil, err
		}
		zeta, err := deriveRandomnessBW6761(fs, "zeta", &proof.H[0], &proof.H[1], &proof.H[2])
		if err != nil {
			return nil, err
		}
		return genericRawBW6761Fr([]bwfr.Element{zeta}), nil
	default:
		return nil, fmt.Errorf("plonk2: unsupported proving key type %T", b.pk)
	}
}

func newGenericProverTranscript(proverConfig *backend.ProverConfig) *fiatshamir.Transcript {
	return fiatshamir.NewTranscript(proverConfig.ChallengeHash, "gamma", "beta", "alpha", "zeta")
}

func bindPublicDataBN254(
	fs *fiatshamir.Transcript,
	challenge string,
	vk *bnplonk.VerifyingKey,
	publicInputs []bnfr.Element,
) error {
	for _, digest := range []bn254.G1Affine{vk.S[0], vk.S[1], vk.S[2], vk.Ql, vk.Qr, vk.Qm, vk.Qo, vk.Qk} {
		if err := fs.Bind(challenge, digest.Marshal()); err != nil {
			return err
		}
	}
	for i := range vk.Qcp {
		if err := fs.Bind(challenge, vk.Qcp[i].Marshal()); err != nil {
			return err
		}
	}
	for i := range publicInputs {
		if err := fs.Bind(challenge, publicInputs[i].Marshal()); err != nil {
			return err
		}
	}
	return nil
}

func bindPublicDataBLS12377(
	fs *fiatshamir.Transcript,
	challenge string,
	vk *blsplonk.VerifyingKey,
	publicInputs []blsfr.Element,
) error {
	for _, digest := range []bls12377.G1Affine{vk.S[0], vk.S[1], vk.S[2], vk.Ql, vk.Qr, vk.Qm, vk.Qo, vk.Qk} {
		if err := fs.Bind(challenge, digest.Marshal()); err != nil {
			return err
		}
	}
	for i := range vk.Qcp {
		if err := fs.Bind(challenge, vk.Qcp[i].Marshal()); err != nil {
			return err
		}
	}
	for i := range publicInputs {
		if err := fs.Bind(challenge, publicInputs[i].Marshal()); err != nil {
			return err
		}
	}
	return nil
}

func bindPublicDataBW6761(
	fs *fiatshamir.Transcript,
	challenge string,
	vk *bwplonk.VerifyingKey,
	publicInputs []bwfr.Element,
) error {
	for _, digest := range []bw6761.G1Affine{vk.S[0], vk.S[1], vk.S[2], vk.Ql, vk.Qr, vk.Qm, vk.Qo, vk.Qk} {
		if err := fs.Bind(challenge, digest.Marshal()); err != nil {
			return err
		}
	}
	for i := range vk.Qcp {
		if err := fs.Bind(challenge, vk.Qcp[i].Marshal()); err != nil {
			return err
		}
	}
	for i := range publicInputs {
		if err := fs.Bind(challenge, publicInputs[i].Marshal()); err != nil {
			return err
		}
	}
	return nil
}

func deriveRandomnessBN254(
	fs *fiatshamir.Transcript,
	challenge string,
	points ...*bn254.G1Affine,
) (bnfr.Element, error) {
	var buf [bn254.SizeOfG1AffineUncompressed]byte
	var r bnfr.Element
	for _, p := range points {
		buf = p.RawBytes()
		if err := fs.Bind(challenge, buf[:]); err != nil {
			return r, err
		}
	}
	b, err := fs.ComputeChallenge(challenge)
	if err != nil {
		return r, err
	}
	r.SetBytes(b)
	return r, nil
}

func deriveRandomnessBLS12377(
	fs *fiatshamir.Transcript,
	challenge string,
	points ...*bls12377.G1Affine,
) (blsfr.Element, error) {
	var buf [bls12377.SizeOfG1AffineUncompressed]byte
	var r blsfr.Element
	for _, p := range points {
		buf = p.RawBytes()
		if err := fs.Bind(challenge, buf[:]); err != nil {
			return r, err
		}
	}
	b, err := fs.ComputeChallenge(challenge)
	if err != nil {
		return r, err
	}
	r.SetBytes(b)
	return r, nil
}

func deriveRandomnessBW6761(
	fs *fiatshamir.Transcript,
	challenge string,
	points ...*bw6761.G1Affine,
) (bwfr.Element, error) {
	var buf [bw6761.SizeOfG1AffineUncompressed]byte
	var r bwfr.Element
	for _, p := range points {
		buf = p.RawBytes()
		if err := fs.Bind(challenge, buf[:]); err != nil {
			return r, err
		}
	}
	b, err := fs.ComputeChallenge(challenge)
	if err != nil {
		return r, err
	}
	r.SetBytes(b)
	return r, nil
}

func (b *genericGPUBackend) lagrangeRawToCanonicalRaw(lagrangeRaw []uint64) ([]uint64, error) {
	canonical := make([]uint64, b.state.n*scalarLimbs(b.state.curve))
	if err := b.lagrangeRawToCanonicalRawInto(canonical, lagrangeRaw); err != nil {
		return nil, err
	}
	return canonical, nil
}

func (b *genericGPUBackend) lagrangeRawToCanonicalRawInto(canonical, lagrangeRaw []uint64) error {
	v, err := b.state.uploadCanonical(lagrangeRaw)
	if err != nil {
		return err
	}
	defer v.Free()
	if err := v.CopyToHostRaw(canonical); err != nil {
		return err
	}
	return nil
}

func blindCanonicalRaw(curve Curve, canonicalRaw []uint64, blindingOrder int) ([]uint64, error) {
	limbs := scalarLimbs(curve)
	if len(canonicalRaw)%limbs != 0 {
		return nil, fmt.Errorf("plonk2: raw canonical length %d is not divisible by %d", len(canonicalRaw), limbs)
	}
	out := make([]uint64, len(canonicalRaw)+(blindingOrder+1)*limbs)
	if err := blindCanonicalRawInto(out, curve, canonicalRaw, blindingOrder); err != nil {
		return nil, err
	}
	return out, nil
}

func blindCanonicalRawInto(out []uint64, curve Curve, canonicalRaw []uint64, blindingOrder int) error {
	limbs := scalarLimbs(curve)
	if len(canonicalRaw)%limbs != 0 {
		return fmt.Errorf("plonk2: raw canonical length %d is not divisible by %d", len(canonicalRaw), limbs)
	}
	want := len(canonicalRaw) + (blindingOrder+1)*limbs
	if len(out) != want {
		return fmt.Errorf("plonk2: blinded raw length %d, want %d", len(out), want)
	}
	blindingRaw, err := randomBlindingRaw(curve, blindingOrder+1)
	if err != nil {
		return err
	}
	copy(out, canonicalRaw)
	copy(out[len(canonicalRaw):], blindingRaw)
	for i := 0; i < blindingOrder+1; i++ {
		dst := rawElementSlice(out, limbs, i, i+1)
		sub := rawElementSlice(blindingRaw, limbs, i, i+1)
		if err := subRawScalarInPlace(curve, dst, sub); err != nil {
			return err
		}
	}
	return nil
}

func randomBlindingRaw(curve Curve, count int) ([]uint64, error) {
	switch curve {
	case CurveBN254:
		blinding := make([]bnfr.Element, count)
		for i := range blinding {
			if _, err := blinding[i].SetRandom(); err != nil {
				return nil, err
			}
		}
		return genericRawBN254Fr(blinding), nil
	case CurveBLS12377:
		blinding := make([]blsfr.Element, count)
		for i := range blinding {
			if _, err := blinding[i].SetRandom(); err != nil {
				return nil, err
			}
		}
		return genericRawBLS12377Fr(blinding), nil
	case CurveBW6761:
		blinding := make([]bwfr.Element, count)
		for i := range blinding {
			if _, err := blinding[i].SetRandom(); err != nil {
				return nil, err
			}
		}
		return genericRawBW6761Fr(blinding), nil
	default:
		return nil, fmt.Errorf("plonk2: unsupported curve %s", curve)
	}
}

func subRawScalarInPlace(curve Curve, dst, sub []uint64) error {
	switch curve {
	case CurveBN254:
		a, err := bn254RawToField(dst)
		if err != nil {
			return err
		}
		b, err := bn254RawToField(sub)
		if err != nil {
			return err
		}
		a[0].Sub(&a[0], &b[0])
		copy(dst, genericRawBN254Fr(a))
	case CurveBLS12377:
		a, err := bls12377RawToField(dst)
		if err != nil {
			return err
		}
		b, err := bls12377RawToField(sub)
		if err != nil {
			return err
		}
		a[0].Sub(&a[0], &b[0])
		copy(dst, genericRawBLS12377Fr(a))
	case CurveBW6761:
		a, err := bw6761RawToField(dst)
		if err != nil {
			return err
		}
		b, err := bw6761RawToField(sub)
		if err != nil {
			return err
		}
		a[0].Sub(&a[0], &b[0])
		copy(dst, genericRawBW6761Fr(a))
	default:
		return fmt.Errorf("plonk2: unsupported curve %s", curve)
	}
	return nil
}

func bn254RawToField(raw []uint64) ([]bnfr.Element, error) {
	if len(raw)%bnfr.Limbs != 0 {
		return nil, fmt.Errorf("BN254 raw field length %d is not divisible by %d", len(raw), bnfr.Limbs)
	}
	out := make([]bnfr.Element, len(raw)/bnfr.Limbs)
	copy(rawBN254Fr(out), raw)
	return out, nil
}

func bls12377RawToField(raw []uint64) ([]blsfr.Element, error) {
	if len(raw)%blsfr.Limbs != 0 {
		return nil, fmt.Errorf("BLS12-377 raw field length %d is not divisible by %d", len(raw), blsfr.Limbs)
	}
	out := make([]blsfr.Element, len(raw)/blsfr.Limbs)
	copy(rawBLS12377Fr(out), raw)
	return out, nil
}

func bw6761RawToField(raw []uint64) ([]bwfr.Element, error) {
	if len(raw)%bwfr.Limbs != 0 {
		return nil, fmt.Errorf("BW6-761 raw field length %d is not divisible by %d", len(raw), bwfr.Limbs)
	}
	out := make([]bwfr.Element, len(raw)/bwfr.Limbs)
	copy(rawBW6761Fr(out), raw)
	return out, nil
}

func blindBN254Canonical(canonical []bnfr.Element, blindingOrder int) ([]bnfr.Element, error) {
	blinding := make([]bnfr.Element, blindingOrder+1)
	for i := range blinding {
		if _, err := blinding[i].SetRandom(); err != nil {
			return nil, err
		}
	}
	out := make([]bnfr.Element, len(canonical)+len(blinding))
	copy(out, canonical)
	copy(out[len(canonical):], blinding)
	for i := range blinding {
		out[i].Sub(&out[i], &blinding[i])
	}
	return out, nil
}

func blindBLS12377Canonical(canonical []blsfr.Element, blindingOrder int) ([]blsfr.Element, error) {
	blinding := make([]blsfr.Element, blindingOrder+1)
	for i := range blinding {
		if _, err := blinding[i].SetRandom(); err != nil {
			return nil, err
		}
	}
	out := make([]blsfr.Element, len(canonical)+len(blinding))
	copy(out, canonical)
	copy(out[len(canonical):], blinding)
	for i := range blinding {
		out[i].Sub(&out[i], &blinding[i])
	}
	return out, nil
}

func blindBW6761Canonical(canonical []bwfr.Element, blindingOrder int) ([]bwfr.Element, error) {
	blinding := make([]bwfr.Element, blindingOrder+1)
	for i := range blinding {
		if _, err := blinding[i].SetRandom(); err != nil {
			return nil, err
		}
	}
	out := make([]bwfr.Element, len(canonical)+len(blinding))
	copy(out, canonical)
	copy(out[len(canonical):], blinding)
	for i := range blinding {
		out[i].Sub(&out[i], &blinding[i])
	}
	return out, nil
}

func splitCommitmentHintInputs(ins []*big.Int, commitmentCount int) (int, []*big.Int, error) {
	if len(ins) == 0 {
		return 0, nil, fmt.Errorf("plonk2: empty BSB22 hint input")
	}
	commDepth := int(ins[0].Int64())
	if commDepth < 0 || commDepth >= commitmentCount {
		return 0, nil, fmt.Errorf("plonk2: invalid BSB22 commitment depth %d", commDepth)
	}
	return commDepth, ins[1:], nil
}

func bindCommitmentHintOutput(
	htfFunc hash.Hash,
	fieldBytes int,
	digest any,
	ops curveProofOps,
	outs []*big.Int,
	bind func([]byte, *big.Int),
) error {
	if len(outs) != 1 {
		return fmt.Errorf("plonk2: BSB22 hint output count %d, want 1", len(outs))
	}
	marshaled, err := ops.digestMarshal(digest)
	if err != nil {
		return err
	}
	if _, err := htfFunc.Write(marshaled); err != nil {
		return err
	}
	hashBts := htfFunc.Sum(nil)
	htfFunc.Reset()
	nbBuf := fieldBytes
	if htfFunc.Size() < fieldBytes {
		nbBuf = htfFunc.Size()
	}
	bind(hashBts[:nbBuf], outs[0])
	return nil
}

func plonkCommitmentCountForCCS(ccs any) int {
	switch spr := ccs.(type) {
	case *csbn254.SparseR1CS:
		return len(spr.CommitmentInfo.(constraint.PlonkCommitments))
	case *csbls12377.SparseR1CS:
		return len(spr.CommitmentInfo.(constraint.PlonkCommitments))
	case *csbw6761.SparseR1CS:
		return len(spr.CommitmentInfo.(constraint.PlonkCommitments))
	default:
		return 0
	}
}
