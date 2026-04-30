//go:build cuda

package plonk2

import (
	"fmt"
	"unsafe"

	bls12377 "github.com/consensys/gnark-crypto/ecc/bls12-377"
	blsfp "github.com/consensys/gnark-crypto/ecc/bls12-377/fp"
	blsfr "github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	blskzg "github.com/consensys/gnark-crypto/ecc/bls12-377/kzg"
	bn254 "github.com/consensys/gnark-crypto/ecc/bn254"
	bnfp "github.com/consensys/gnark-crypto/ecc/bn254/fp"
	bnfr "github.com/consensys/gnark-crypto/ecc/bn254/fr"
	bnkzg "github.com/consensys/gnark-crypto/ecc/bn254/kzg"
	bw6761 "github.com/consensys/gnark-crypto/ecc/bw6-761"
	bwfp "github.com/consensys/gnark-crypto/ecc/bw6-761/fp"
	bwfr "github.com/consensys/gnark-crypto/ecc/bw6-761/fr"
	bwkzg "github.com/consensys/gnark-crypto/ecc/bw6-761/kzg"
	gnarkplonk "github.com/consensys/gnark/backend/plonk"
	blsplonk "github.com/consensys/gnark/backend/plonk/bls12-377"
	bnplonk "github.com/consensys/gnark/backend/plonk/bn254"
	bwplonk "github.com/consensys/gnark/backend/plonk/bw6-761"
)

type curveProofOps interface {
	curve() Curve
	newProof(qcpLen int) gnarkplonk.Proof
	fieldSliceToRaw(values any) ([]uint64, error)
	rawToFieldSlice(raw []uint64) (any, error)
	rawCommitmentToDigest(raw []uint64) (any, error)
	setLRO(proof gnarkplonk.Proof, index int, digest any) error
	setZ(proof gnarkplonk.Proof, digest any) error
	setH(proof gnarkplonk.Proof, index int, digest any) error
	setBsb22Commitment(proof gnarkplonk.Proof, index int, digest any) error
	digestMarshal(digest any) ([]byte, error)
}

func newCurveProofOps(curve Curve) (curveProofOps, error) {
	switch curve {
	case CurveBN254:
		return bn254ProofOps{}, nil
	case CurveBLS12377:
		return bls12377ProofOps{}, nil
	case CurveBW6761:
		return bw6761ProofOps{}, nil
	default:
		return nil, fmt.Errorf("plonk2: unsupported curve %s", curve)
	}
}

type bn254ProofOps struct{}

func (bn254ProofOps) curve() Curve { return CurveBN254 }

func (bn254ProofOps) newProof(qcpLen int) gnarkplonk.Proof {
	return &bnplonk.Proof{Bsb22Commitments: make([]bnkzg.Digest, qcpLen)}
}

func (bn254ProofOps) fieldSliceToRaw(values any) ([]uint64, error) {
	typed, ok := values.([]bnfr.Element)
	if !ok {
		return nil, fmt.Errorf("plonk2: expected []bn254/fr.Element, got %T", values)
	}
	return genericRawBN254Fr(typed), nil
}

func (bn254ProofOps) rawToFieldSlice(raw []uint64) (any, error) {
	if len(raw)%bnfr.Limbs != 0 {
		return nil, fmt.Errorf("plonk2: BN254 raw field length %d is not divisible by %d", len(raw), bnfr.Limbs)
	}
	out := make([]bnfr.Element, len(raw)/bnfr.Limbs)
	copy(rawBN254Fr(out), raw)
	return out, nil
}

func (bn254ProofOps) rawCommitmentToDigest(raw []uint64) (any, error) {
	if len(raw) != 3*bnfp.Limbs {
		return nil, fmt.Errorf("plonk2: BN254 raw projective length %d, want %d", len(raw), 3*bnfp.Limbs)
	}
	var jac bn254.G1Jac
	copy(unsafe.Slice((*uint64)(unsafe.Pointer(&jac)), 3*bnfp.Limbs), raw)
	var digest bnkzg.Digest
	digest.FromJacobian(&jac)
	return digest, nil
}

func (bn254ProofOps) setLRO(proof gnarkplonk.Proof, index int, digest any) error {
	typedProof, ok := proof.(*bnplonk.Proof)
	if !ok {
		return fmt.Errorf("plonk2: expected BN254 proof, got %T", proof)
	}
	typedDigest, ok := digest.(bnkzg.Digest)
	if !ok {
		return fmt.Errorf("plonk2: expected BN254 digest, got %T", digest)
	}
	if index < 0 || index >= len(typedProof.LRO) {
		return fmt.Errorf("plonk2: BN254 LRO index %d out of range", index)
	}
	typedProof.LRO[index] = typedDigest
	return nil
}

func (bn254ProofOps) setBsb22Commitment(proof gnarkplonk.Proof, index int, digest any) error {
	typedProof, ok := proof.(*bnplonk.Proof)
	if !ok {
		return fmt.Errorf("plonk2: expected BN254 proof, got %T", proof)
	}
	typedDigest, ok := digest.(bnkzg.Digest)
	if !ok {
		return fmt.Errorf("plonk2: expected BN254 digest, got %T", digest)
	}
	if index < 0 || index >= len(typedProof.Bsb22Commitments) {
		return fmt.Errorf("plonk2: BN254 BSB22 index %d out of range", index)
	}
	typedProof.Bsb22Commitments[index] = typedDigest
	return nil
}

func (bn254ProofOps) setZ(proof gnarkplonk.Proof, digest any) error {
	typedProof, ok := proof.(*bnplonk.Proof)
	if !ok {
		return fmt.Errorf("plonk2: expected BN254 proof, got %T", proof)
	}
	typedDigest, ok := digest.(bnkzg.Digest)
	if !ok {
		return fmt.Errorf("plonk2: expected BN254 digest, got %T", digest)
	}
	typedProof.Z = typedDigest
	return nil
}

func (bn254ProofOps) setH(proof gnarkplonk.Proof, index int, digest any) error {
	typedProof, ok := proof.(*bnplonk.Proof)
	if !ok {
		return fmt.Errorf("plonk2: expected BN254 proof, got %T", proof)
	}
	typedDigest, ok := digest.(bnkzg.Digest)
	if !ok {
		return fmt.Errorf("plonk2: expected BN254 digest, got %T", digest)
	}
	if index < 0 || index >= len(typedProof.H) {
		return fmt.Errorf("plonk2: BN254 H index %d out of range", index)
	}
	typedProof.H[index] = typedDigest
	return nil
}

func (bn254ProofOps) digestMarshal(digest any) ([]byte, error) {
	typedDigest, ok := digest.(bnkzg.Digest)
	if !ok {
		return nil, fmt.Errorf("plonk2: expected BN254 digest, got %T", digest)
	}
	return typedDigest.Marshal(), nil
}

type bls12377ProofOps struct{}

func (bls12377ProofOps) curve() Curve { return CurveBLS12377 }

func (bls12377ProofOps) newProof(qcpLen int) gnarkplonk.Proof {
	return &blsplonk.Proof{Bsb22Commitments: make([]blskzg.Digest, qcpLen)}
}

func (bls12377ProofOps) fieldSliceToRaw(values any) ([]uint64, error) {
	typed, ok := values.([]blsfr.Element)
	if !ok {
		return nil, fmt.Errorf("plonk2: expected []bls12-377/fr.Element, got %T", values)
	}
	return genericRawBLS12377Fr(typed), nil
}

func (bls12377ProofOps) rawToFieldSlice(raw []uint64) (any, error) {
	if len(raw)%blsfr.Limbs != 0 {
		return nil, fmt.Errorf("plonk2: BLS12-377 raw field length %d is not divisible by %d", len(raw), blsfr.Limbs)
	}
	out := make([]blsfr.Element, len(raw)/blsfr.Limbs)
	copy(rawBLS12377Fr(out), raw)
	return out, nil
}

func (bls12377ProofOps) rawCommitmentToDigest(raw []uint64) (any, error) {
	if len(raw) != 3*blsfp.Limbs {
		return nil, fmt.Errorf("plonk2: BLS12-377 raw projective length %d, want %d", len(raw), 3*blsfp.Limbs)
	}
	var jac bls12377.G1Jac
	copy(unsafe.Slice((*uint64)(unsafe.Pointer(&jac)), 3*blsfp.Limbs), raw)
	var digest blskzg.Digest
	digest.FromJacobian(&jac)
	return digest, nil
}

func (bls12377ProofOps) setLRO(proof gnarkplonk.Proof, index int, digest any) error {
	typedProof, ok := proof.(*blsplonk.Proof)
	if !ok {
		return fmt.Errorf("plonk2: expected BLS12-377 proof, got %T", proof)
	}
	typedDigest, ok := digest.(blskzg.Digest)
	if !ok {
		return fmt.Errorf("plonk2: expected BLS12-377 digest, got %T", digest)
	}
	if index < 0 || index >= len(typedProof.LRO) {
		return fmt.Errorf("plonk2: BLS12-377 LRO index %d out of range", index)
	}
	typedProof.LRO[index] = typedDigest
	return nil
}

func (bls12377ProofOps) setBsb22Commitment(proof gnarkplonk.Proof, index int, digest any) error {
	typedProof, ok := proof.(*blsplonk.Proof)
	if !ok {
		return fmt.Errorf("plonk2: expected BLS12-377 proof, got %T", proof)
	}
	typedDigest, ok := digest.(blskzg.Digest)
	if !ok {
		return fmt.Errorf("plonk2: expected BLS12-377 digest, got %T", digest)
	}
	if index < 0 || index >= len(typedProof.Bsb22Commitments) {
		return fmt.Errorf("plonk2: BLS12-377 BSB22 index %d out of range", index)
	}
	typedProof.Bsb22Commitments[index] = typedDigest
	return nil
}

func (bls12377ProofOps) setZ(proof gnarkplonk.Proof, digest any) error {
	typedProof, ok := proof.(*blsplonk.Proof)
	if !ok {
		return fmt.Errorf("plonk2: expected BLS12-377 proof, got %T", proof)
	}
	typedDigest, ok := digest.(blskzg.Digest)
	if !ok {
		return fmt.Errorf("plonk2: expected BLS12-377 digest, got %T", digest)
	}
	typedProof.Z = typedDigest
	return nil
}

func (bls12377ProofOps) setH(proof gnarkplonk.Proof, index int, digest any) error {
	typedProof, ok := proof.(*blsplonk.Proof)
	if !ok {
		return fmt.Errorf("plonk2: expected BLS12-377 proof, got %T", proof)
	}
	typedDigest, ok := digest.(blskzg.Digest)
	if !ok {
		return fmt.Errorf("plonk2: expected BLS12-377 digest, got %T", digest)
	}
	if index < 0 || index >= len(typedProof.H) {
		return fmt.Errorf("plonk2: BLS12-377 H index %d out of range", index)
	}
	typedProof.H[index] = typedDigest
	return nil
}

func (bls12377ProofOps) digestMarshal(digest any) ([]byte, error) {
	typedDigest, ok := digest.(blskzg.Digest)
	if !ok {
		return nil, fmt.Errorf("plonk2: expected BLS12-377 digest, got %T", digest)
	}
	return typedDigest.Marshal(), nil
}

type bw6761ProofOps struct{}

func (bw6761ProofOps) curve() Curve { return CurveBW6761 }

func (bw6761ProofOps) newProof(qcpLen int) gnarkplonk.Proof {
	return &bwplonk.Proof{Bsb22Commitments: make([]bwkzg.Digest, qcpLen)}
}

func (bw6761ProofOps) fieldSliceToRaw(values any) ([]uint64, error) {
	typed, ok := values.([]bwfr.Element)
	if !ok {
		return nil, fmt.Errorf("plonk2: expected []bw6-761/fr.Element, got %T", values)
	}
	return genericRawBW6761Fr(typed), nil
}

func (bw6761ProofOps) rawToFieldSlice(raw []uint64) (any, error) {
	if len(raw)%bwfr.Limbs != 0 {
		return nil, fmt.Errorf("plonk2: BW6-761 raw field length %d is not divisible by %d", len(raw), bwfr.Limbs)
	}
	out := make([]bwfr.Element, len(raw)/bwfr.Limbs)
	copy(rawBW6761Fr(out), raw)
	return out, nil
}

func (bw6761ProofOps) rawCommitmentToDigest(raw []uint64) (any, error) {
	if len(raw) != 3*bwfp.Limbs {
		return nil, fmt.Errorf("plonk2: BW6-761 raw projective length %d, want %d", len(raw), 3*bwfp.Limbs)
	}
	var jac bw6761.G1Jac
	copy(unsafe.Slice((*uint64)(unsafe.Pointer(&jac)), 3*bwfp.Limbs), raw)
	var digest bwkzg.Digest
	digest.FromJacobian(&jac)
	return digest, nil
}

func (bw6761ProofOps) setLRO(proof gnarkplonk.Proof, index int, digest any) error {
	typedProof, ok := proof.(*bwplonk.Proof)
	if !ok {
		return fmt.Errorf("plonk2: expected BW6-761 proof, got %T", proof)
	}
	typedDigest, ok := digest.(bwkzg.Digest)
	if !ok {
		return fmt.Errorf("plonk2: expected BW6-761 digest, got %T", digest)
	}
	if index < 0 || index >= len(typedProof.LRO) {
		return fmt.Errorf("plonk2: BW6-761 LRO index %d out of range", index)
	}
	typedProof.LRO[index] = typedDigest
	return nil
}

func (bw6761ProofOps) setBsb22Commitment(proof gnarkplonk.Proof, index int, digest any) error {
	typedProof, ok := proof.(*bwplonk.Proof)
	if !ok {
		return fmt.Errorf("plonk2: expected BW6-761 proof, got %T", proof)
	}
	typedDigest, ok := digest.(bwkzg.Digest)
	if !ok {
		return fmt.Errorf("plonk2: expected BW6-761 digest, got %T", digest)
	}
	if index < 0 || index >= len(typedProof.Bsb22Commitments) {
		return fmt.Errorf("plonk2: BW6-761 BSB22 index %d out of range", index)
	}
	typedProof.Bsb22Commitments[index] = typedDigest
	return nil
}

func (bw6761ProofOps) setZ(proof gnarkplonk.Proof, digest any) error {
	typedProof, ok := proof.(*bwplonk.Proof)
	if !ok {
		return fmt.Errorf("plonk2: expected BW6-761 proof, got %T", proof)
	}
	typedDigest, ok := digest.(bwkzg.Digest)
	if !ok {
		return fmt.Errorf("plonk2: expected BW6-761 digest, got %T", digest)
	}
	typedProof.Z = typedDigest
	return nil
}

func (bw6761ProofOps) setH(proof gnarkplonk.Proof, index int, digest any) error {
	typedProof, ok := proof.(*bwplonk.Proof)
	if !ok {
		return fmt.Errorf("plonk2: expected BW6-761 proof, got %T", proof)
	}
	typedDigest, ok := digest.(bwkzg.Digest)
	if !ok {
		return fmt.Errorf("plonk2: expected BW6-761 digest, got %T", digest)
	}
	if index < 0 || index >= len(typedProof.H) {
		return fmt.Errorf("plonk2: BW6-761 H index %d out of range", index)
	}
	typedProof.H[index] = typedDigest
	return nil
}

func (bw6761ProofOps) digestMarshal(digest any) ([]byte, error) {
	typedDigest, ok := digest.(bwkzg.Digest)
	if !ok {
		return nil, fmt.Errorf("plonk2: expected BW6-761 digest, got %T", digest)
	}
	return typedDigest.Marshal(), nil
}

func rawBN254Fr(values []bnfr.Element) []uint64 {
	if len(values) == 0 {
		return nil
	}
	return unsafe.Slice((*uint64)(unsafe.Pointer(&values[0])), len(values)*bnfr.Limbs)
}

func rawBLS12377Fr(values []blsfr.Element) []uint64 {
	if len(values) == 0 {
		return nil
	}
	return unsafe.Slice((*uint64)(unsafe.Pointer(&values[0])), len(values)*blsfr.Limbs)
}

func rawBW6761Fr(values []bwfr.Element) []uint64 {
	if len(values) == 0 {
		return nil
	}
	return unsafe.Slice((*uint64)(unsafe.Pointer(&values[0])), len(values)*bwfr.Limbs)
}
