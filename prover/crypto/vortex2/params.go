package vortex2

import (
	"hash"

	"github.com/consensys/accelerated-crypto-monorepo/crypto/ringsis"
	"github.com/consensys/accelerated-crypto-monorepo/maths/fft"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
)

// Contains the public parameters of the commitment scheme
type Params struct {
	// Key of the ringsis instance in use
	Key ringsis.Key
	// BlowUpFactor⁻¹, rate of the RS code ( > 1)
	BlowUpFactor int
	// Domains[1] used for the Reed Solomon encoding
	Domains [2]*fft.Domain
	// NbColumns number of columns of the matrix storing the polynomials. The total size of
	// the polynomials which are committed is NbColumns x NbRows.
	// The Number of columns is a power of 2, it corresponds to the original size of the codewords
	// of the Reed Solomon code.
	NbColumns int
	// MaxNbRows number of rows of the matrix storing the polynomials. If a polynomial p is appended
	// whose size if not 0 mod MaxNbRows, it is padded as p' so that len(p')=0 mod MaxNbRows.
	MaxNbRows int
	// HashFunc is an optional function that returns a `hash.Hash` it is used when vortex is used
	// in "Merkle-tree" mode. In this case, the hash function is mandatory.
	HashFunc func() hash.Hash
	// NoSisHashFunc is an optional hash function that is used in place of the SIS. If it is set,
	NoSisHashFunc func() hash.Hash
}

// NewTensorCommitment retunrs a new TensorCommitment
// * blowUp factor : inverse-rate of the code ( > 1)
// * size size of the polynomial to be committed. The size of the commitment is
// then ρ * √(m) where m² = size
func NewParams(blowUpFactor, nbColumns, maxNbRows int, sisParams ringsis.Params) *Params {
	var res Params

	if !utils.IsPowerOfTwo(nbColumns) {
		utils.Panic("the number of columns has to be a power of two")
	}

	// domain[0]: domain to perform the FFT^-1, of size capacity * sqrt
	// domain[1]: domain to perform FFT, of size rho * capacity * sqrt

	res.Domains[0] = fft.NewDomain(utils.NextPowerOfTwo(nbColumns))
	res.Domains[1] = fft.NewDomain(blowUpFactor * nbColumns)

	// size of the matrix
	res.NbColumns = nbColumns
	res.MaxNbRows = maxNbRows

	// rate
	res.BlowUpFactor = blowUpFactor

	// Hash function
	res.Key = sisParams.GenerateKey(maxNbRows)
	return &res
}

// EncodedNumCols
func (p *Params) NumEncodedCols() int {
	return utils.NextPowerOfTwo(p.NbColumns) * p.BlowUpFactor
}

// NewParamsMerkle returns a new Params object that can return a Merkle-tree
func (p *Params) WithMerkleMode(hashFunc func() hash.Hash) *Params {
	p.HashFunc = hashFunc
	return p
}

// IsMerkleMode returns true if the Params have been amended with the MerkleMode
func (p *Params) IsMerkleMode() bool {
	return p.HashFunc != nil
}

// RemoveSis forces vortex to use another hash function than SIS
func (p *Params) RemoveSis(h func() hash.Hash) {
	p.NoSisHashFunc = h
	p.Key = ringsis.Key{} // and remove the key
}

// Returns true if the parameters are set to not use SIS
func (p *Params) HasSisReplacement() bool {
	return p.NoSisHashFunc != nil
}
