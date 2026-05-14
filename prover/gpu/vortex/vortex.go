// Package vortex implements GPU-accelerated Vortex polynomial commitment over KoalaBear.
//
// Vortex encodes a matrix of field elements using Reed-Solomon codes,
// hashes columns (SIS + Poseidon2), builds a Merkle tree, and commits to the root.
// Opening proofs reveal a random linear combination (UAlpha) and selected columns
// with Merkle inclusion proofs.
//
// This package is API-compatible with linea-monorepo/prover/crypto/vortex/vortex_koalabear.
// When built with CGO + CUDA, hot paths (RS encoding, Merkle tree, linear combination)
// run on GPU. Without CGO, it falls back to gnark-crypto's CPU implementation.
//
// Protocol overview (cf. https://eprint.iacr.org/2024/185):
//
//	                                   Prover                        Verifier
//	┌───────────────────────────────────────────────────────────────────────────┐
//	│  M[nRows × nCols]                                                        │
//	│       │                                                                   │
//	│  RS encode rows ──▶ Encoded[nRows × nCols·ρ]                              │
//	│       │                                                                   │
//	│  SIS hash columns ──▶ Poseidon2 ──▶ Merkle tree ──▶ root ──────▶ commit  │
//	│                                                                           │
//	│       ◀────────── α, x, S (random challenges) ◀──────────────────────     │
//	│                                                                           │
//	│  UAlpha = Σ αⁱ · row[i]      (E4 vector, length nCols·ρ)                 │
//	│  open columns S, Merkle proofs  ──────────────────────▶ verify            │
//	│                                                         │                 │
//	│                                              eval(UAlpha, x) = Σ yᵢ·αⁱ ? │
//	│                                              UAlpha ∈ RS code?            │
//	│                                              col(α) = UAlpha[col_idx]?    │
//	│                                              Merkle proofs valid?          │
//	└───────────────────────────────────────────────────────────────────────────┘
package vortex

import (
	"fmt"
	"hash"

	"github.com/consensys/gnark-crypto/field/koalabear"
	fext "github.com/consensys/gnark-crypto/field/koalabear/extensions"
	"github.com/consensys/gnark-crypto/field/koalabear/fft"
	"github.com/consensys/gnark-crypto/field/koalabear/sis"
	"github.com/consensys/gnark-crypto/field/koalabear/vortex"
	"github.com/consensys/gnark-crypto/utils"
)

// ─────────────────────────────────────────────────────────────────────────────
// Shared types (used by both CPU and GPU paths)
// ─────────────────────────────────────────────────────────────────────────────

// Hash is a Poseidon2 digest: 8 KoalaBear elements (32 bytes).
type Hash = [8]koalabear.Element

// HashConstructor mirrors gnark-crypto's vortex hash constructor.
type HashConstructor = vortex.HashConstructor

// Option configures the prover behavior (hash functions, etc).
type Option = vortex.Option

var (
	// ErrWrongSizeHash is returned when a custom hash is not 32 bytes.
	ErrWrongSizeHash = vortex.ErrWrongSizeHash
)

// MerkleProof is a sequence of sibling hashes from leaf to root.
type MerkleProof = vortex.MerkleProof

// Params holds the public parameters of the Vortex commitment scheme.
type Params struct {
	inner *vortex.Params
}

// Proof is a Vortex opening proof.
type Proof struct {
	// UAlpha is the random linear combination Σ αⁱ · encoded_row[i] ∈ E4^(nCols·ρ).
	UAlpha []fext.E4
	// Columns are the opened columns from the encoded matrix.
	Columns [][]koalabear.Element
	// MerkleProofs are the Merkle inclusion proofs for each opened column.
	MerkleProofs []MerkleProof
}

// ─────────────────────────────────────────────────────────────────────────────
// Parameters
// ─────────────────────────────────────────────────────────────────────────────

func WithMerkleHash(h hash.Hash) Option {
	return vortex.WithMerkleHash(h)
}

func WithColumnHash(h hash.Hash) Option {
	return vortex.WithColumnHash(h)
}

// NewParams constructs Vortex parameters.
//
// Unlike gnark-crypto's vortex.NewParams (which restricts rate to 2, 4, 8),
// this constructor accepts any power-of-two rate as long as the total domain
// size (nbColumns × rate) does not exceed the KoalaBear two-adicity limit (2^24).
func NewParams(nbColumns, maxNbRows int, sisParams *sis.RSis,
	rate, numSelectedColumns int, opts ...Option) (*Params, error) {

	if nbColumns < 1 || nbColumns&(nbColumns-1) != 0 {
		return nil, fmt.Errorf("vortex: number of columns must be a power of two, got %d", nbColumns)
	}
	if rate < 2 || rate&(rate-1) != 0 {
		return nil, fmt.Errorf("vortex: rate must be a power of two >= 2, got %d", rate)
	}

	shift, err := koalabear.Generator(uint64(nbColumns * rate))
	if err != nil {
		return nil, fmt.Errorf("vortex: generator for domain size %d: %w", nbColumns*rate, err)
	}

	smallDomain := fft.NewDomain(uint64(nbColumns), fft.WithShift(shift))
	cosetTable, err := smallDomain.CosetTable()
	if err != nil {
		return nil, fmt.Errorf("vortex: coset table: %w", err)
	}
	cosetTableBitReverse := make(koalabear.Vector, len(cosetTable))
	copy(cosetTableBitReverse, cosetTable)
	utils.BitReverse(cosetTableBitReverse)

	bigDomain := fft.NewDomain(uint64(nbColumns * rate))

	p := &vortex.Params{
		Key:                  sisParams,
		Domains:              [2]*fft.Domain{smallDomain, bigDomain},
		ReedSolomonInvRate:   rate,
		NbColumns:            nbColumns,
		MaxNbRows:            maxNbRows,
		NumSelectedColumns:   numSelectedColumns,
		CosetTableBitReverse: cosetTableBitReverse,
	}
	return &Params{inner: p}, nil
}

// SizeCodeWord returns the number of columns after RS encoding (nbColumns × rate).
func (p *Params) SizeCodeWord() int {
	return p.inner.SizeCodeWord()
}

// ─────────────────────────────────────────────────────────────────────────────
// Verify (same for CPU and GPU)
// ─────────────────────────────────────────────────────────────────────────────

type VerifierInput = vortex.VerifierInput

// VerifyInput matches gnark-crypto's verifier API.
func (p *Params) VerifyInput(input VerifierInput) error {
	return p.inner.Verify(input)
}

// Verify checks a Vortex opening proof.
func (p *Params) Verify(root Hash, proof *Proof, claimedValues []fext.E4,
	x, alpha fext.E4, selectedCols []int) error {

	return p.inner.Verify(vortex.VerifierInput{
		MerkleRoot:      root,
		ClaimedValues:   claimedValues,
		EvaluationPoint: x,
		Alpha:           alpha,
		SelectedColumns: selectedCols,
		Proof: &vortex.Proof{
			UAlpha:                   proof.UAlpha,
			OpenedColumns:            proof.Columns,
			MerkleProofOpenedColumns: proof.MerkleProofs,
		},
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// Standalone helpers (exported for GPU kernel validation in tests)
// ─────────────────────────────────────────────────────────────────────────────

func (p *Params) EncodeReedSolomon(input, res []koalabear.Element) {
	p.inner.EncodeReedSolomon(input, res)
}

func CompressPoseidon2(a, b Hash) Hash {
	return vortex.CompressPoseidon2(a, b)
}

func CompressPoseidon2x16(matrix []koalabear.Element, colSize int, result []Hash) {
	vortex.CompressPoseidon2x16(matrix, colSize, result)
}

func HashPoseidon2(x []koalabear.Element) Hash {
	return vortex.HashPoseidon2(x)
}

func HashPoseidon2x16(sisHashes []koalabear.Element, merkleLeaves []Hash, sisKeySize int) {
	vortex.HashPoseidon2x16(sisHashes, merkleLeaves, sisKeySize)
}

func EvalBasePolyLagrange(poly []koalabear.Element, x fext.E4) (fext.E4, error) {
	return vortex.EvalBasePolyLagrange(poly, x)
}

func EvalFextPolyLagrange(poly []fext.E4, x fext.E4) (fext.E4, error) {
	return vortex.EvalFextPolyLagrange(poly, x)
}

func EvalBasePolyHorner(poly []koalabear.Element, x fext.E4) fext.E4 {
	return vortex.EvalBasePolyHorner(poly, x)
}

func EvalFextPolyHorner(poly []fext.E4, x fext.E4) fext.E4 {
	return vortex.EvalFextPolyHorner(poly, x)
}

func BatchEvalBasePolyLagrange(polys [][]koalabear.Element, x fext.E4, oncoset ...bool) ([]fext.E4, error) {
	return vortex.BatchEvalBasePolyLagrange(polys, x, oncoset...)
}

func BatchEvalFextPolyLagrange(polys [][]fext.E4, x fext.E4, oncoset ...bool) ([]fext.E4, error) {
	return vortex.BatchEvalFextPolyLagrange(polys, x, oncoset...)
}
